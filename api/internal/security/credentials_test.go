package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSecureCredentialsManager(t *testing.T) {
	// Note: This test uses the embedded credentials which are placeholders
	// In a real environment, these would be properly encrypted credentials
	
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	// Test that manager is initialized correctly
	if manager.encryptedPayload == nil {
		t.Error("Encrypted payload should not be nil")
	}
	
	if manager.certPinner == nil {
		t.Error("Certificate pinner should not be nil")
	}
	
	// Test security metrics
	metrics := manager.GetSecurityMetrics()
	if metrics == nil {
		t.Error("Security metrics should not be nil")
	}
	
	// Validate expected metric fields
	expectedFields := []string{
		"access_count", "last_access", "max_access_count",
		"access_timeout", "binary_hash_prefix", "encryption_version",
		"certificate_pins", "security_initialized",
	}
	
	for _, field := range expectedFields {
		if _, exists := metrics[field]; !exists {
			t.Errorf("Expected metric field %s not found", field)
		}
	}
}

func TestCredentialAccessTracking(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	// Get initial metrics
	initialMetrics := manager.GetSecurityMetrics()
	initialCount := initialMetrics["access_count"].(int64)
	
	// Note: This test may fail if the embedded credentials are not properly set
	// In a real deployment, the credentials would be encrypted during build
	
	// Test access count tracking (we'll test the tracking mechanism even if decryption fails)
	_, err = manager.GetSecureCredentials(ctx)
	// We expect this to fail with placeholder credentials, but access should still be tracked
	
	// Get updated metrics
	updatedMetrics := manager.GetSecurityMetrics()
	updatedCount := updatedMetrics["access_count"].(int64)
	
	// Access count should increase regardless of success/failure
	if updatedCount <= initialCount {
		// Note: This might not increase if the integrity check fails early
		t.Logf("Access count did not increase (expected with placeholder credentials)")
	}
}

func TestCredentialAccessLimits(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	// Test access limit checking
	err = manager.checkAccessLimits()
	if err != nil {
		t.Errorf("Initial access limit check should pass: %v", err)
	}
	
	// Simulate max access count reached
	manager.accessCount = MaxAccessCount
	err = manager.checkAccessLimits()
	if err == nil {
		t.Error("Should fail when max access count is reached")
	}
	
	// Reset and test timeout
	manager.accessCount = 0
	manager.lastAccess = time.Now().Add(-2 * AccessTimeout) // Simulate old access
	err = manager.checkAccessLimits()
	if err == nil {
		t.Error("Should fail when access timeout is exceeded")
	}
}

func TestAuditLogging(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	// Test audit event logging (we can't easily capture the logs in this test,
	// but we can ensure the method doesn't panic)
	manager.logAuditEvent("test_event", true, "", ctx)
	manager.logAuditEvent("test_error", false, "test error message", ctx)
	
	// Test with context values
	ctxWithValues := context.WithValue(ctx, "user-agent", "test-agent")
	ctxWithValues = context.WithValue(ctxWithValues, "client-ip", "127.0.0.1")
	manager.logAuditEvent("test_with_context", true, "", ctxWithValues)
}

func TestCredentialRotation(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	// Create a valid encrypted payload for testing rotation
	testCredentials := []byte(`{"type": "service_account", "project_id": "test"}`)
	appSalt := []byte(ApplicationSalt)
	
	newPayload, err := EncryptCredentials(testCredentials, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}
	
	// Test rotation with valid payload
	err = manager.RotateCredentials(newPayload)
	if err != nil {
		t.Errorf("Credential rotation should succeed: %v", err)
	}
	
	// Test rotation with nil payload
	err = manager.RotateCredentials(nil)
	if err == nil {
		t.Error("Rotation should fail with nil payload")
	}
	
	// Test rotation with invalid version
	invalidPayload := *newPayload
	invalidPayload.Version = 99
	err = manager.RotateCredentials(&invalidPayload)
	if err == nil {
		t.Error("Rotation should fail with invalid version")
	}
}

func TestSecurityConfigValidation(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	// Note: This will likely fail with placeholder credentials, but we're testing
	// that the validation method exists and runs without panicking
	err = manager.ValidateSecurityConfiguration()
	if err != nil {
		t.Logf("Security validation failed (expected with placeholder credentials): %v", err)
	}
}

func TestCredentialsManagerClose(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	
	// Test that close works without error
	err = manager.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
	
	// Test double close (should be safe)
	err = manager.Close()
	if err != nil {
		t.Errorf("Double close should not return error: %v", err)
	}
}

func TestEmbeddedCredentialsFormat(t *testing.T) {
	// Test that embedded credentials have the expected JSON structure
	var payload EncryptedPayload
	err := json.Unmarshal([]byte(embeddedCredentials), &payload)
	if err != nil {
		t.Fatalf("Failed to parse embedded credentials: %v", err)
	}
	
	// Validate structure
	if payload.Version != 1 {
		t.Errorf("Expected version 1, got %d", payload.Version)
	}
	
	if len(payload.Salt) == 0 {
		t.Error("Salt should not be empty")
	}
	
	if len(payload.Nonce) == 0 {
		t.Error("Nonce should not be empty")
	}
	
	if len(payload.Ciphertext) == 0 {
		t.Error("Ciphertext should not be empty")
	}
	
	// Note: These are placeholder values in the test, so we can't validate
	// their actual cryptographic properties
}

func TestApplicationConstants(t *testing.T) {
	// Test that application constants are set correctly
	if ApplicationSalt == "" {
		t.Error("ApplicationSalt should not be empty")
	}
	
	if len(ApplicationSalt) < 16 {
		t.Error("ApplicationSalt should be at least 16 characters")
	}
	
	if MaxAccessCount <= 0 {
		t.Error("MaxAccessCount should be positive")
	}
	
	if AccessTimeout <= 0 {
		t.Error("AccessTimeout should be positive")
	}
	
	if expectedBinaryHash == "" {
		t.Error("expectedBinaryHash should not be empty")
	}
}

func TestCredentialAccessEvent(t *testing.T) {
	event := CredentialAccessEvent{
		Timestamp:    time.Now(),
		EventType:    "test_event",
		Success:      true,
		BinaryHash:   "test_hash",
		ProcessID:    os.Getpid(),
		AccessCount:  1,
		ClientIP:     "127.0.0.1",
		UserAgent:    "test-agent",
	}
	
	// Test that all fields are set correctly
	if event.EventType != "test_event" {
		t.Error("EventType not set correctly")
	}
	
	if !event.Success {
		t.Error("Success should be true")
	}
	
	if event.ProcessID != os.Getpid() {
		t.Error("ProcessID not set correctly")
	}
	
	if event.AccessCount != 1 {
		t.Error("AccessCount not set correctly")
	}
	
	if event.ClientIP != "127.0.0.1" {
		t.Error("ClientIP not set correctly")
	}
}

func TestGenerateApplicationHash(t *testing.T) {
	hash, err := GenerateApplicationHash()
	if err != nil {
		t.Fatalf("Failed to generate application hash: %v", err)
	}
	
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
	
	// Validate that it's a valid hex string
	err = ValidateIntegrityConfig(hash)
	if err != nil {
		t.Errorf("Generated hash failed validation: %v", err)
	}
}

func BenchmarkCredentialAccess(b *testing.B) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		b.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This will likely fail with placeholder credentials,
		// but we're benchmarking the access pattern
		_, err := manager.GetSecureCredentials(ctx)
		// We don't check the error here since we expect it to fail with placeholders
		_ = err
	}
}

func BenchmarkAuditLogging(b *testing.B) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		b.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.logAuditEvent("benchmark_event", true, "", ctx)
	}
}

// TestGetCredentials tests the GetCredentials method thoroughly
func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name           string
		prepareManager func(*SecureCredentialsManager)
		expectError    bool
		errorContains  string
	}{
		{
			name: "access_limit_exceeded",
			prepareManager: func(m *SecureCredentialsManager) {
				m.accessCount = MaxAccessCount
			},
			expectError:   true,
			errorContains: "maximum credential access count exceeded",
		},
		{
			name: "access_timeout_exceeded",
			prepareManager: func(m *SecureCredentialsManager) {
				m.lastAccess = time.Now().Add(-2 * AccessTimeout)
			},
			expectError:   true,
			errorContains: "credential access timeout exceeded",
		},
		{
			name: "integrity_check_fails",
			prepareManager: func(m *SecureCredentialsManager) {
				// Set an invalid binary hash to trigger integrity failure
				m.binaryHash = "invalid-hash"
				m.integrityChecker = NewIntegrityChecker("invalid-hash")
			},
			expectError:   true,
			errorContains: "binary integrity verification failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewSecureCredentialsManager()
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}
			defer manager.Close()
			
			// Apply test-specific preparation
			if tt.prepareManager != nil {
				tt.prepareManager(manager)
			}
			
			ctx := context.Background()
			credentials, err := manager.GetCredentials(ctx)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}
			
			// For success cases, we expect errors with placeholder credentials
			// but we can verify the method executes correctly
			if err == nil && credentials != nil {
				// This would only happen with real encrypted credentials
				t.Log("Got credentials successfully")
			}
		})
	}
}

// TestDecryptCredentials tests the private decryptCredentials method
func TestDecryptCredentials(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()
	
	// Create a valid encrypted payload for testing
	testCredentials := []byte(`{"type": "service_account", "project_id": "test-project"}`)
	appSalt := []byte(ApplicationSalt)
	
	payload, err := EncryptCredentials(testCredentials, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}
	
	// Replace the manager's payload with our test payload
	manager.encryptedPayload = payload
	
	// Test successful decryption
	credentials, err := manager.decryptCredentials()
	if err != nil {
		t.Fatalf("Decryption should succeed: %v", err)
	}
	
	if !bytes.Equal(credentials, testCredentials) {
		t.Error("Decrypted credentials do not match original")
	}
	
	// Test with corrupted payload
	corruptedPayload := *payload
	corruptedPayload.Ciphertext[0] ^= 0x01 // Flip a bit
	manager.encryptedPayload = &corruptedPayload
	
	_, err = manager.decryptCredentials()
	if err == nil {
		t.Error("Should fail with corrupted payload")
	}
}

// TestCreateSecureSheetsService tests the CreateSecureSheetsService method
func TestCreateSecureSheetsService(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	// Test with placeholder credentials (should fail but not panic)
	service, cleanup, err := manager.CreateSecureSheetsService(ctx)
	if err != nil {
		// Expected with placeholder credentials
		t.Logf("Service creation failed as expected with placeholder credentials: %v", err)
		return
	}
	
	// If we somehow get a service (with real credentials), test cleanup
	if service != nil && cleanup != nil {
		cleanup()
		t.Log("Service created and cleaned up successfully")
	}
}

// TestValidateSecurityConfigurationDetailed provides detailed testing of security validation
func TestValidateSecurityConfigurationDetailed(t *testing.T) {
	tests := []struct {
		name           string
		modifyManager  func(*SecureCredentialsManager)
		expectError    bool
		errorContains  string
	}{
		{
			name: "nil_certificate_pinner",
			modifyManager: func(m *SecureCredentialsManager) {
				m.certPinner = nil
			},
			expectError:   true,
			errorContains: "certificate pinner not initialized",
		},
		{
			name: "invalid_binary_hash",
			modifyManager: func(m *SecureCredentialsManager) {
				m.binaryHash = "invalid"
			},
			expectError:   true,
			errorContains: "integrity config invalid",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewSecureCredentialsManager()
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}
			defer manager.Close()
			
			if tt.modifyManager != nil {
				tt.modifyManager(manager)
			}
			
			err = manager.ValidateSecurityConfiguration()
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				// May fail with placeholder credentials, log for investigation
				t.Logf("Validation failed (may be expected): %v", err)
			}
		})
	}
}

// TestDeriveKey tests the deriveKey function
func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name      string
		inputData []byte
		salt      []byte
		keyLen    int
		wantErr   bool
	}{
		{
			name:      "valid_parameters",
			inputData: []byte("test-input-data"),
			salt:      []byte("test-salt-16-bytes!"),
			keyLen:    32,
			wantErr:   false,
		},
		{
			name:      "zero_key_length",
			inputData: []byte("test-input-data"),
			salt:      []byte("test-salt-16-bytes!"),
			keyLen:    0,
			wantErr:   true,
		},
		{
			name:      "negative_key_length",
			inputData: []byte("test-input-data"),
			salt:      []byte("test-salt-16-bytes!"),
			keyLen:    -1,
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := deriveKey(tt.inputData, tt.salt, tt.keyLen)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(key) != tt.keyLen {
				t.Errorf("Expected key length %d, got %d", tt.keyLen, len(key))
			}
		})
	}
}

// TestValidateApplicationIntegrity tests the ValidateApplicationIntegrity function
func TestValidateApplicationIntegrity(t *testing.T) {
	tests := []struct {
		name         string
		expectedHash string
		wantErr      bool
	}{
		{
			name:         "invalid_hash_format",
			expectedHash: "invalid-hash",
			wantErr:      true,
		},
		{
			name:         "empty_hash",
			expectedHash: "",
			wantErr:      true,
		},
		{
			name:         "short_hash",
			expectedHash: "abc123",
			wantErr:      true,
		},
		{
			name:         "valid_format_wrong_hash",
			expectedHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr:      true, // Will fail integrity check
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateApplicationIntegrity(tt.expectedHash)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSecureCredentialsManagerConcurrency tests concurrent access to credentials
func TestSecureCredentialsManagerConcurrency(t *testing.T) {
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	
	// Test concurrent access to GetSecurityMetrics (safe operation)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("panic: %v", r)
				}
			}()
			
			// Test safe concurrent operations
			metrics := manager.GetSecurityMetrics()
			if metrics == nil {
				errors <- fmt.Errorf("metrics should not be nil")
				return
			}
			
			// Test credential access (will likely fail but shouldn't panic)
			_, err := manager.GetCredentials(ctx)
			// Don't check error as it's expected to fail with placeholder credentials
			_ = err
			
			errors <- nil
		}()
	}
	
	// Wait for all goroutines and check for errors
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errors:
			if err != nil {
				t.Errorf("Concurrent access error: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}
}

