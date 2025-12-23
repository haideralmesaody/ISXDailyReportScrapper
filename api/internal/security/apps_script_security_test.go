package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultAppsScriptSecurityConfig tests default configuration creation
func TestDefaultAppsScriptSecurityConfig(t *testing.T) {
	config := DefaultAppsScriptSecurityConfig()
	
	require.NotNil(t, config)
	assert.NotEmpty(t, config.SharedSecret)
	assert.GreaterOrEqual(t, len(config.SharedSecret), 32, "Shared secret should be at least 32 characters")
	assert.Equal(t, 30*time.Second, config.RequestTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryBaseDelay)
	assert.Equal(t, 5*time.Minute, config.TimestampWindow)
	assert.True(t, config.EnableEncryption)
	assert.True(t, config.RequireSignature)
	assert.Equal(t, "ISX-Pulse-SecureClient/1.0", config.UserAgent)
	assert.Equal(t, int64(1024*1024), config.MaxRequestSize)
	assert.False(t, config.AllowInsecure)
}

// TestSecureAppsScriptClientCreation tests client creation with various configurations
func TestSecureAppsScriptClientCreation(t *testing.T) {
	tests := []struct {
		name   string
		config *AppsScriptSecurityConfig
	}{
		{
			name:   "default configuration",
			config: nil, // Should use default
		},
		{
			name: "custom configuration",
			config: &AppsScriptSecurityConfig{
				SharedSecret:     "custom_secret_that_is_very_long_and_secure_12345678",
				RequestTimeout:   45 * time.Second,
				MaxRetries:      5,
				RetryBaseDelay:  2 * time.Second,
				TimestampWindow: 10 * time.Minute,
				EnableEncryption: true,
				RequireSignature: true,
				UserAgent:       "Custom-Client/2.0",
				MaxRequestSize:  2 * 1024 * 1024,
				AllowInsecure:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSecureAppsScriptClient(tt.config, nil)
			
			require.NotNil(t, client)
			require.NotNil(t, client.config)
			require.NotNil(t, client.httpClient)
			
			if tt.config != nil {
				assert.Equal(t, tt.config.SharedSecret, client.config.SharedSecret)
				assert.Equal(t, tt.config.RequestTimeout, client.config.RequestTimeout)
				assert.Equal(t, tt.config.MaxRetries, client.config.MaxRetries)
			} else {
				// Should use defaults
				assert.NotEmpty(t, client.config.SharedSecret)
				assert.Equal(t, 30*time.Second, client.config.RequestTimeout)
				assert.Equal(t, 3, client.config.MaxRetries)
			}
		})
	}
}

// TestHMACSignatureGeneration tests HMAC signature creation and validation
func TestHMACSignatureGeneration(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "test_secret_key_for_hmac_validation_testing",
		RequireSignature: true,
	}
	
	client := NewSecureAppsScriptClient(config, nil)

	tests := []struct {
		name        string
		payload     map[string]interface{}
		fingerprint string
		requestID   string
	}{
		{
			name: "simple payload",
			payload: map[string]interface{}{
				"action":     "activateScratchCard",
				"licenseKey": "ISX-1234-5678-90AB",
			},
			fingerprint: "test_fingerprint_123",
			requestID:   "req_test_001",
		},
		{
			name: "complex payload",
			payload: map[string]interface{}{
				"action":     "batchActivation",
				"licenseKeys": []string{"ISX-1111-2222-3333", "ISX-4444-5555-6666"},
				"metadata": map[string]interface{}{
					"batchId":   "batch_001",
					"timestamp": time.Now().Unix(),
				},
			},
			fingerprint: "complex_test_fingerprint_456",
			requestID:   "req_test_complex_002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create signed request
			signedReq, err := client.createSignedRequest(tt.payload, tt.fingerprint, tt.requestID)
			require.NoError(t, err)
			require.NotNil(t, signedReq)

			// Verify signature is present
			assert.NotEmpty(t, signedReq.Signature)
			assert.Equal(t, tt.requestID, signedReq.RequestID)
			assert.Equal(t, tt.fingerprint, signedReq.Fingerprint)
			assert.NotEmpty(t, signedReq.Nonce)
			assert.Greater(t, signedReq.Timestamp, int64(0))

			// Verify signature is valid by re-creating it
			regeneratedSig, err := client.signRequest(signedReq)
			require.NoError(t, err)
			assert.Equal(t, signedReq.Signature, regeneratedSig)
		})
	}
}

// TestTimestampValidation tests timestamp validation in requests and responses
func TestTimestampValidation(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "test_secret_for_timestamp_validation",
		RequireSignature: true,
		TimestampWindow:  5 * time.Minute,
	}
	
	client := NewSecureAppsScriptClient(config, nil)

	tests := []struct {
		name             string
		timestampOffset  time.Duration
		expectValidation bool
		description      string
	}{
		{
			name:             "current timestamp",
			timestampOffset:  0,
			expectValidation: true,
			description:      "Current timestamp should be valid",
		},
		{
			name:             "recent past timestamp",
			timestampOffset:  -2 * time.Minute,
			expectValidation: true,
			description:      "Recent past timestamp within window should be valid",
		},
		{
			name:             "recent future timestamp",
			timestampOffset:  2 * time.Minute,
			expectValidation: true,
			description:      "Recent future timestamp within window should be valid",
		},
		{
			name:             "old timestamp outside window",
			timestampOffset:  -10 * time.Minute,
			expectValidation: false,
			description:      "Old timestamp outside window should be invalid",
		},
		{
			name:             "far future timestamp outside window",
			timestampOffset:  15 * time.Minute,
			expectValidation: false,
			description:      "Far future timestamp outside window should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			testTimestamp := now.Add(tt.timestampOffset)

			// Create a mock response with the test timestamp
			response := &SignedResponse{
				Timestamp: testTimestamp.Unix(),
				RequestID: "test_request_123",
				Success:   true,
				Data: map[string]interface{}{
					"activationId": "act_test_123",
				},
				Signature: "", // Will be set below
			}

			// Create signature for the response
			canonical := fmt.Sprintf("%d|%s|%t", response.Timestamp, response.RequestID, response.Success)
			dataJSON, _ := json.Marshal(response.Data)
			canonical += "|" + string(dataJSON)

			h := hmac.New(sha256.New, []byte(config.SharedSecret))
			h.Write([]byte(canonical))
			response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))

			// Test signature verification (which includes timestamp validation)
			err := client.verifyResponseSignature(response, "test_request_123")
			
			if tt.expectValidation {
				assert.NoError(t, err, tt.description)
			} else {
				assert.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), "timestamp", "Error should mention timestamp")
			}
		})
	}
}

// TestRequestSigning tests request signing with various payload types
func TestRequestSigning(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "signing_test_secret_key_with_sufficient_length",
		RequireSignature: true,
	}
	
	client := NewSecureAppsScriptClient(config, nil)

	tests := []struct {
		name     string
		payload  map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name: "standard license activation",
			payload: map[string]interface{}{
				"action":            "activateScratchCard",
				"licenseKey":        "ISX-1234-5678-90AB",
				"deviceFingerprint": "device_hash_123",
			},
			wantErr: false,
		},
		{
			name: "batch operation",
			payload: map[string]interface{}{
				"action": "batchOperation",
				"items": []interface{}{
					map[string]interface{}{"id": 1, "value": "test1"},
					map[string]interface{}{"id": 2, "value": "test2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty payload",
			payload: map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nil payload should handle gracefully",
			payload: nil,
			wantErr: true,
			errMsg:  "payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock signed request
			if tt.payload == nil {
				// Test error handling for nil payload
				signedReq := &SignedRequest{
					Timestamp:   time.Now().Unix(),
					Nonce:       "test_nonce",
					RequestID:   "test_request",
					Payload:     tt.payload,
					Fingerprint: "test_fingerprint",
				}

				_, err := client.signRequest(signedReq)
				if tt.wantErr {
					require.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
				return
			}

			// Normal flow for non-nil payloads
			signedReq, err := client.createSignedRequest(tt.payload, "test_fingerprint", "test_request_id")
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, signedReq)
				assert.NotEmpty(t, signedReq.Signature)
				
				// Verify signature can be regenerated consistently
				sig1, err1 := client.signRequest(signedReq)
				require.NoError(t, err1)
				
				sig2, err2 := client.signRequest(signedReq)
				require.NoError(t, err2)
				
				assert.Equal(t, sig1, sig2, "Signature should be deterministic")
				assert.Equal(t, signedReq.Signature, sig1, "Generated signature should match stored signature")
			}
		})
	}
}

// TestSignatureVerification tests response signature verification
func TestSignatureVerification(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "verification_test_secret_key_long_enough_for_security",
		RequireSignature: true,
	}
	
	client := NewSecureAppsScriptClient(config, nil)

	tests := []struct {
		name           string
		response       *SignedResponse
		expectedReqID  string
		tamperWith     func(*SignedResponse)
		expectValid    bool
		description    string
	}{
		{
			name: "valid response",
			response: &SignedResponse{
				Timestamp: time.Now().Unix(),
				RequestID: "valid_request_123",
				Success:   true,
				Data: map[string]interface{}{
					"activationId": "act_valid_123",
					"message":      "Success",
				},
			},
			expectedReqID: "valid_request_123",
			expectValid:   true,
			description:   "Valid response should pass verification",
		},
		{
			name: "response with error",
			response: &SignedResponse{
				Timestamp: time.Now().Unix(),
				RequestID: "error_request_456",
				Success:   false,
				Error:     "License not found",
			},
			expectedReqID: "error_request_456",
			expectValid:   true,
			description:   "Valid error response should pass verification",
		},
		{
			name: "tampered data",
			response: &SignedResponse{
				Timestamp: time.Now().Unix(),
				RequestID: "tampered_request_789",
				Success:   true,
				Data: map[string]interface{}{
					"activationId": "act_original_789",
				},
			},
			expectedReqID: "tampered_request_789",
			tamperWith: func(r *SignedResponse) {
				// Change data after signature is created
				r.Data["activationId"] = "act_tampered_789"
			},
			expectValid: false,
			description: "Response with tampered data should fail verification",
		},
		{
			name: "wrong request ID",
			response: &SignedResponse{
				Timestamp: time.Now().Unix(),
				RequestID: "wrong_id_request",
				Success:   true,
				Data: map[string]interface{}{
					"activationId": "act_wrong_id",
				},
			},
			expectedReqID: "expected_different_id",
			expectValid:   false,
			description:   "Response with wrong request ID should fail verification",
		},
		{
			name: "missing signature",
			response: &SignedResponse{
				Timestamp: time.Now().Unix(),
				RequestID: "missing_sig_request",
				Success:   true,
				Data: map[string]interface{}{
					"activationId": "act_missing_sig",
				},
				Signature: "", // Empty signature
			},
			expectedReqID: "missing_sig_request",
			expectValid:   false,
			description:   "Response with missing signature should fail verification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create proper signature for the response (unless it should be missing)
			if tt.response.Signature != "" || tt.name != "missing signature" {
				canonical := fmt.Sprintf("%d|%s|%t", tt.response.Timestamp, tt.response.RequestID, tt.response.Success)
				
				if tt.response.Data != nil {
					dataJSON, _ := json.Marshal(tt.response.Data)
					canonical += "|" + string(dataJSON)
				}
				
				if tt.response.Error != "" {
					canonical += "|" + tt.response.Error
				}

				h := hmac.New(sha256.New, []byte(config.SharedSecret))
				h.Write([]byte(canonical))
				tt.response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))
			}

			// Apply tampering if specified
			if tt.tamperWith != nil {
				tt.tamperWith(tt.response)
			}

			// Verify signature
			err := client.verifyResponseSignature(tt.response, tt.expectedReqID)
			
			if tt.expectValid {
				assert.NoError(t, err, tt.description)
			} else {
				assert.Error(t, err, tt.description)
			}
		})
	}
}

// TestSecureRequest tests the complete secure request flow
func TestSecureRequest(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "secure_request_test_secret_key_long_for_security",
		RequestTimeout:   10 * time.Second,
		MaxRetries:      2,
		RetryBaseDelay:  100 * time.Millisecond,
		RequireSignature: true,
		AllowInsecure:   true, // Allow HTTP for testing
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("User-Agent"), "ISX-Pulse-SecureClient")
		assert.NotEmpty(t, r.Header.Get("X-Request-ID"))
		assert.NotEmpty(t, r.Header.Get("X-Timestamp"))

		// Parse request body
		var requestBody SignedRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create response
		response := &SignedResponse{
			Timestamp: time.Now().Unix(),
			RequestID: requestBody.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"activationId": "act_mock_12345",
				"message":      "Mock activation successful",
			},
		}

		// Sign response
		canonical := fmt.Sprintf("%d|%s|%t", response.Timestamp, response.RequestID, response.Success)
		dataJSON, _ := json.Marshal(response.Data)
		canonical += "|" + string(dataJSON)

		h := hmac.New(sha256.New, []byte(config.SharedSecret))
		h.Write([]byte(canonical))
		response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewSecureAppsScriptClient(config, nil)
	ctx := context.Background()

	// Test successful request
	payload := map[string]interface{}{
		"action":            "activateScratchCard",
		"licenseKey":        "ISX-TEST-1234-5678",
		"deviceFingerprint": "test_device_fingerprint",
	}

	response, err := client.SecureRequest(ctx, server.URL, payload, "test_fingerprint")
	require.NoError(t, err)
	require.NotNil(t, response)
	
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data)
	assert.Equal(t, "act_mock_12345", response.Data["activationId"])
	assert.Equal(t, "Mock activation successful", response.Data["message"])
}

// TestRetryLogic tests retry logic for failed requests
func TestRetryLogic(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "retry_test_secret_key_long_enough_for_hmac",
		RequestTimeout:   5 * time.Second,
		MaxRetries:      3,
		RetryBaseDelay:  50 * time.Millisecond,
		RequireSignature: true,
		AllowInsecure:   true,
	}

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First two requests fail, third succeeds
		if requestCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server temporarily unavailable"))
			return
		}

		// Success response
		var requestBody SignedRequest
		json.NewDecoder(r.Body).Decode(&requestBody)

		response := &SignedResponse{
			Timestamp: time.Now().Unix(),
			RequestID: requestBody.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"message": "Success after retries",
			},
		}

		// Sign response
		canonical := fmt.Sprintf("%d|%s|%t", response.Timestamp, response.RequestID, response.Success)
		dataJSON, _ := json.Marshal(response.Data)
		canonical += "|" + string(dataJSON)

		h := hmac.New(sha256.New, []byte(config.SharedSecret))
		h.Write([]byte(canonical))
		response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewSecureAppsScriptClient(config, nil)
	ctx := context.Background()

	payload := map[string]interface{}{
		"action": "testRetry",
	}

	response, err := client.SecureRequest(ctx, server.URL, payload, "test_fingerprint")
	require.NoError(t, err)
	require.NotNil(t, response)
	
	assert.True(t, response.Success)
	assert.Equal(t, "Success after retries", response.Data["message"])
	assert.Equal(t, 3, requestCount, "Should have made exactly 3 requests")
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *AppsScriptSecurityConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid configuration",
			config:  DefaultAppsScriptSecurityConfig(),
			wantErr: false,
		},
		{
			name: "shared secret too short",
			config: &AppsScriptSecurityConfig{
				SharedSecret:    "short",
				RequestTimeout:  30 * time.Second,
				MaxRetries:     3,
				TimestampWindow: 5 * time.Minute,
				MaxRequestSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "shared secret",
		},
		{
			name: "request timeout too short",
			config: &AppsScriptSecurityConfig{
				SharedSecret:    "long_enough_secret_key_for_validation",
				RequestTimeout:  1 * time.Second,
				MaxRetries:     3,
				TimestampWindow: 5 * time.Minute,
				MaxRequestSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "too many retries",
			config: &AppsScriptSecurityConfig{
				SharedSecret:    "long_enough_secret_key_for_validation",
				RequestTimeout:  30 * time.Second,
				MaxRetries:     15,
				TimestampWindow: 5 * time.Minute,
				MaxRequestSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "retries",
		},
		{
			name: "timestamp window too short",
			config: &AppsScriptSecurityConfig{
				SharedSecret:    "long_enough_secret_key_for_validation",
				RequestTimeout:  30 * time.Second,
				MaxRetries:     3,
				TimestampWindow: 30 * time.Second,
				MaxRequestSize: 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "timestamp window",
		},
		{
			name: "max request size too small",
			config: &AppsScriptSecurityConfig{
				SharedSecret:    "long_enough_secret_key_for_validation",
				RequestTimeout:  30 * time.Second,
				MaxRetries:     3,
				TimestampWindow: 5 * time.Minute,
				MaxRequestSize: 512,
			},
			wantErr: true,
			errMsg:  "request size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSecureAppsScriptClient(tt.config, nil)
			err := client.ValidateConfiguration()
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestConcurrentRequests tests thread safety of the secure client
func TestConcurrentRequests(t *testing.T) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "concurrent_test_secret_key_for_thread_safety",
		RequestTimeout:   5 * time.Second,
		MaxRetries:      1,
		RequireSignature: true,
		AllowInsecure:   true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody SignedRequest
		json.NewDecoder(r.Body).Decode(&requestBody)

		response := &SignedResponse{
			Timestamp: time.Now().Unix(),
			RequestID: requestBody.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"message": "Concurrent request handled",
				"id":      requestBody.RequestID,
			},
		}

		// Sign response
		canonical := fmt.Sprintf("%d|%s|%t", response.Timestamp, response.RequestID, response.Success)
		dataJSON, _ := json.Marshal(response.Data)
		canonical += "|" + string(dataJSON)

		h := hmac.New(sha256.New, []byte(config.SharedSecret))
		h.Write([]byte(canonical))
		response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewSecureAppsScriptClient(config, nil)
	ctx := context.Background()

	const goroutineCount = 10
	var wg sync.WaitGroup
	results := make([]*SignedResponse, goroutineCount)
	errors := make([]error, goroutineCount)

	// Launch concurrent requests
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			payload := map[string]interface{}{
				"action": "concurrentTest",
				"index":  index,
			}
			
			results[index], errors[index] = client.SecureRequest(ctx, server.URL, payload, fmt.Sprintf("fingerprint_%d", index))
		}(i)
	}

	wg.Wait()

	// Verify all requests succeeded
	for i, err := range errors {
		require.NoError(t, err, "Request %d failed", i)
		require.NotNil(t, results[i], "Result %d is nil", i)
		assert.True(t, results[i].Success, "Request %d was not successful", i)
	}
}

// Benchmark tests for performance validation
func BenchmarkSignatureGeneration(b *testing.B) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "benchmark_secret_key_for_performance_testing",
		RequireSignature: true,
	}
	
	client := NewSecureAppsScriptClient(config, nil)
	
	payload := map[string]interface{}{
		"action":            "benchmarkTest",
		"licenseKey":        "ISX-BENCH-TEST-1234",
		"deviceFingerprint": "benchmark_device_fingerprint",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.createSignedRequest(payload, "fingerprint", fmt.Sprintf("req_%d", i))
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkSignatureVerification(b *testing.B) {
	config := &AppsScriptSecurityConfig{
		SharedSecret:     "benchmark_verification_secret_key_for_testing",
		RequireSignature: true,
	}
	
	client := NewSecureAppsScriptClient(config, nil)

	// Create a sample response to verify
	response := &SignedResponse{
		Timestamp: time.Now().Unix(),
		RequestID: "bench_request_123",
		Success:   true,
		Data: map[string]interface{}{
			"activationId": "act_bench_123",
			"message":      "Benchmark response",
		},
	}

	// Sign the response
	canonical := fmt.Sprintf("%d|%s|%t", response.Timestamp, response.RequestID, response.Success)
	dataJSON, _ := json.Marshal(response.Data)
	canonical += "|" + string(dataJSON)

	h := hmac.New(sha256.New, []byte(config.SharedSecret))
	h.Write([]byte(canonical))
	response.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := client.verifyResponseSignature(response, "bench_request_123")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}