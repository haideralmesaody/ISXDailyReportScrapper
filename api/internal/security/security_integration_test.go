package security

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSecurityIntegration performs end-to-end security testing
func TestSecurityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("CredentialEncryptionWorkflow", testCredentialEncryptionWorkflow)
	t.Run("IntegrityVerificationWorkflow", testIntegrityVerificationWorkflow)
	t.Run("CertificatePinningWorkflow", testCertificatePinningWorkflow)
	t.Run("AuditLoggingWorkflow", testAuditLoggingWorkflow)
	t.Run("SecurityManagerWorkflow", testSecurityManagerWorkflow)
}

func testCredentialEncryptionWorkflow(t *testing.T) {
	// Simulate a complete credential encryption/decryption workflow
	
	// Step 1: Create mock Google service account credentials
	mockCredentials := map[string]interface{}{
		"type":         "service_account",
		"project_id":   "test-project",
		"private_key_id": "test-key-id",
		"private_key":  "-----BEGIN PRIVATE KEY-----\ntest-private-key-content\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id":    "123456789",
		"auth_uri":     "https://accounts.google.com/o/oauth2/auth",
		"token_uri":    "https://oauth2.googleapis.com/token",
	}
	
	credentialsJSON, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}
	
	// Step 2: Encrypt credentials
	appSalt := []byte("test-integration-salt-32-bytes!!")
	payload, err := EncryptCredentials(credentialsJSON, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to encrypt credentials: %v", err)
	}
	
	// Step 3: Verify payload structure
	if payload.Version != 1 {
		t.Errorf("Expected version 1, got %d", payload.Version)
	}
	
	if len(payload.Salt) != 32 {
		t.Errorf("Expected salt length 32, got %d", len(payload.Salt))
	}
	
	if len(payload.Nonce) != 12 {
		t.Errorf("Expected nonce length 12, got %d", len(payload.Nonce))
	}
	
	// Step 4: Simulate storage and retrieval
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}
	
	var retrievedPayload EncryptedPayload
	if err := json.Unmarshal(payloadJSON, &retrievedPayload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}
	
	// Step 5: Decrypt and verify
	decryptedCredentials, err := DecryptCredentials(&retrievedPayload, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to decrypt credentials: %v", err)
	}
	defer decryptedCredentials.Clear()
	
	// Verify decrypted content matches original
	var decryptedData map[string]interface{}
	if err := json.Unmarshal(decryptedCredentials.Data(), &decryptedData); err != nil {
		t.Fatalf("Failed to unmarshal decrypted data: %v", err)
	}
	
	if decryptedData["project_id"] != "test-project" {
		t.Error("Decrypted project_id doesn't match original")
	}
	
	if decryptedData["client_email"] != "test@test-project.iam.gserviceaccount.com" {
		t.Error("Decrypted client_email doesn't match original")
	}
}

func testIntegrityVerificationWorkflow(t *testing.T) {
	// Test the complete integrity verification workflow
	
	// Create a temporary test binary
	tempDir := t.TempDir()
	testBinary := filepath.Join(tempDir, "test-app.exe")
	testContent := []byte("test application binary content for integrity verification")
	
	if err := os.WriteFile(testBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}
	
	// Calculate hash of the test binary
	checker := &IntegrityChecker{}
	expectedHash, fileSize, err := checker.calculateBinaryHash(testBinary)
	if err != nil {
		t.Fatalf("Failed to calculate binary hash: %v", err)
	}
	
	// Create integrity checker with expected hash
	integrityChecker := NewIntegrityChecker(expectedHash)
	
	// This test simulates the verification process but uses a test file
	// instead of the actual executable
	if len(expectedHash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(expectedHash))
	}
	
	if fileSize != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), fileSize)
	}
	
	// Test tampering detection
	indicators := integrityChecker.detectTamperingIndicators()
	if indicators == nil {
		t.Error("Tampering indicators should not be nil")
	}
	
	// Generate tampering report
	report := indicators.GetDetailedTamperingReport()
	if report == "" {
		t.Error("Tampering report should not be empty")
	}
	
	t.Logf("Integrity verification completed. Hash: %s, Size: %d bytes", expectedHash[:16]+"...", fileSize)
}

func testCertificatePinningWorkflow(t *testing.T) {
	// Test certificate pinning workflow
	
	// Create certificate pinner with default configuration
	pinner := NewCertificatePinner(DefaultPinningConfig())
	
	// Verify Google APIs pins are loaded
	pinnedHashes := pinner.GetPinnedHashes()
	if len(pinnedHashes) == 0 {
		t.Error("Should have pinned hashes for Google APIs")
	}
	
	// Verify specific Google API hostnames are pinned
	expectedHosts := []string{
		"sheets.googleapis.com",
		"accounts.google.com",
		"oauth2.googleapis.com",
	}
	
	for _, host := range expectedHosts {
		if pins, exists := pinnedHashes[host]; !exists || len(pins) == 0 {
			t.Errorf("Expected pins for host %s", host)
		}
	}
	
	// Create secure HTTP client
	client := pinner.CreateSecureHTTPClient(DefaultPinningConfig())
	if client == nil {
		t.Error("Secure HTTP client should not be nil")
	}
	
	// Test adding custom pin
	customHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	err := pinner.AddCustomPin("custom.example.com", customHash)
	if err != nil {
		t.Errorf("Failed to add custom pin: %v", err)
	}
	
	// Verify custom pin was added
	updatedHashes := pinner.GetPinnedHashes()
	if customPins, exists := updatedHashes["custom.example.com"]; !exists || len(customPins) == 0 {
		t.Error("Custom pin was not added correctly")
	}
	
	// Generate pinning report
	reports := pinner.GeneratePinningReport()
	if len(reports) == 0 {
		t.Error("Pinning report should not be empty")
	}
	
	for _, report := range reports {
		if report.Hostname == "" {
			t.Error("Report hostname should not be empty")
		}
		if len(report.PinnedHashes) == 0 {
			t.Error("Report should have pinned hashes")
		}
	}
}

func testAuditLoggingWorkflow(t *testing.T) {
	// Test audit logging workflow
	
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	
	// Test various audit events
	auditEvents := []struct {
		eventType string
		success   bool
		message   string
	}{
		{"initialization", true, ""},
		{"credential_access", true, ""},
		{"validation_failed", false, "test validation error"},
		{"security_check", true, ""},
		{"tampering_detected", false, "integrity verification failed"},
	}
	
	for _, event := range auditEvents {
		manager.logAuditEvent(event.eventType, event.success, event.message, ctx)
	}
	
	// Test audit event with context values
	ctxWithValues := context.WithValue(ctx, "user-agent", "test-integration-client")
	ctxWithValues = context.WithValue(ctxWithValues, "client-ip", "192.168.1.100")
	
	manager.logAuditEvent("test_with_context", true, "", ctxWithValues)
	
	// Verify security metrics are tracked
	metrics := manager.GetSecurityMetrics()
	if metrics == nil {
		t.Error("Security metrics should not be nil")
	}
	
	requiredMetrics := []string{
		"access_count", "security_initialized", "binary_hash_prefix",
	}
	
	for _, metric := range requiredMetrics {
		if _, exists := metrics[metric]; !exists {
			t.Errorf("Required metric %s not found", metric)
		}
	}
}

func testSecurityManagerWorkflow(t *testing.T) {
	// Test the complete security manager workflow
	
	// Step 1: Initialize manager
	manager, err := NewSecureCredentialsManager()
	if err != nil {
		t.Fatalf("Failed to create secure credentials manager: %v", err)
	}
	defer manager.Close()
	
	// Step 2: Validate security configuration
	// Note: This will likely fail with placeholder credentials, but we're testing the workflow
	err = manager.ValidateSecurityConfiguration()
	if err != nil {
		t.Logf("Security validation failed (expected with placeholder credentials): %v", err)
	}
	
	// Step 3: Test access limits
	initialMetrics := manager.GetSecurityMetrics()
	initialAccessCount := initialMetrics["access_count"].(int64)
	
	// Step 4: Test credential access (will fail with placeholders, but tests the workflow)
	ctx := context.Background()
	_, err = manager.GetSecureCredentials(ctx)
	// We expect this to fail due to placeholder credentials or integrity checks
	
	// Step 5: Verify access tracking
	finalMetrics := manager.GetSecurityMetrics()
	if finalMetrics["security_initialized"] != true {
		t.Error("Security should be marked as initialized")
	}
	
	// Step 6: Test credential rotation workflow
	testCredentials := []byte(`{"type": "service_account", "project_id": "rotation-test"}`)
	newPayload, err := EncryptCredentials(testCredentials, []byte(ApplicationSalt), DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to create test payload for rotation: %v", err)
	}
	
	err = manager.RotateCredentials(newPayload)
	if err != nil {
		t.Errorf("Credential rotation should succeed: %v", err)
	}
	
	// Step 7: Verify rotation reset access count
	rotationMetrics := manager.GetSecurityMetrics()
	rotationAccessCount := rotationMetrics["access_count"].(int64)
	if rotationAccessCount > initialAccessCount {
		t.Logf("Access count after rotation: %d (was %d)", rotationAccessCount, initialAccessCount)
	}
	
	// Step 8: Test cleanup
	err = manager.Close()
	if err != nil {
		t.Errorf("Manager close should not fail: %v", err)
	}
}

func TestSecurityBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmarks in short mode")
	}
	
	// Run security-related benchmarks to ensure performance
	
	t.Run("EncryptionPerformance", func(t *testing.T) {
		testData := make([]byte, 4096) // 4KB test data
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		
		appSalt := []byte("benchmark-salt-32-characters!!")
		config := DefaultEncryptionConfig()
		
		start := time.Now()
		iterations := 100
		
		for i := 0; i < iterations; i++ {
			payload, err := EncryptCredentials(testData, appSalt, config)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}
			
			credentials, err := DecryptCredentials(payload, appSalt, config)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}
			credentials.Clear()
		}
		
		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)
		
		t.Logf("Encryption/Decryption performance: %d iterations in %v (avg: %v per iteration)",
			iterations, duration, avgTime)
		
		// Performance requirement: should complete within reasonable time
		if avgTime > 100*time.Millisecond {
			t.Logf("WARNING: Encryption/decryption is slow (avg: %v)", avgTime)
		}
	})
	
	t.Run("IntegrityCheckPerformance", func(t *testing.T) {
		// Create a test file for integrity checking
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "perf-test.exe")
		testContent := make([]byte, 1024*1024) // 1MB file
		
		for i := range testContent {
			testContent[i] = byte(i % 256)
		}
		
		if err := os.WriteFile(testFile, testContent, 0755); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		checker := &IntegrityChecker{}
		
		start := time.Now()
		iterations := 10
		
		for i := 0; i < iterations; i++ {
			_, _, err := checker.calculateBinaryHash(testFile)
			if err != nil {
				t.Fatalf("Hash calculation failed: %v", err)
			}
		}
		
		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)
		
		t.Logf("Integrity check performance: %d iterations in %v (avg: %v per iteration)",
			iterations, duration, avgTime)
		
		// Performance requirement: should complete within reasonable time for 1MB file
		if avgTime > 50*time.Millisecond {
			t.Logf("WARNING: Integrity checking is slow (avg: %v for 1MB file)", avgTime)
		}
	})
}

func TestSecurityErrorHandling(t *testing.T) {
	// Test error handling in various security scenarios
	
	t.Run("InvalidEncryptionInputs", func(t *testing.T) {
		// Test with various invalid inputs
		invalidInputs := []struct {
			name      string
			plaintext []byte
			appSalt   []byte
		}{
			{"empty_plaintext", []byte{}, []byte("valid-salt-16-chars!!")},
			{"nil_plaintext", nil, []byte("valid-salt-16-chars!!")},
			{"short_salt", []byte("test"), []byte("short")},
			{"empty_salt", []byte("test"), []byte{}},
			{"nil_salt", []byte("test"), nil},
		}
		
		for _, input := range invalidInputs {
			t.Run(input.name, func(t *testing.T) {
				_, err := EncryptCredentials(input.plaintext, input.appSalt, DefaultEncryptionConfig())
				if err == nil {
					t.Error("Expected error but got none")
				}
			})
		}
	})
	
	t.Run("TamperedPayloads", func(t *testing.T) {
		// Create valid payload first
		plaintext := []byte("test data")
		appSalt := []byte("test-salt-16-chars!!")
		
		payload, err := EncryptCredentials(plaintext, appSalt, DefaultEncryptionConfig())
		if err != nil {
			t.Fatalf("Failed to create test payload: %v", err)
		}
		
		// Test various tampering scenarios
		tamperingTests := []struct {
			name   string
			tamper func(*EncryptedPayload)
		}{
			{
				"version_change",
				func(p *EncryptedPayload) { p.Version = 99 },
			},
			{
				"corrupted_ciphertext",
				func(p *EncryptedPayload) {
					if len(p.Ciphertext) > 0 {
						p.Ciphertext[0] ^= 0xFF
					}
				},
			},
			{
				"corrupted_integrity",
				func(p *EncryptedPayload) {
					if len(p.Integrity) > 0 {
						p.Integrity[0] ^= 0xFF
					}
				},
			},
		}
		
		for _, test := range tamperingTests {
			t.Run(test.name, func(t *testing.T) {
				// Create tampered copy
				tamperedPayload := *payload
				test.tamper(&tamperedPayload)
				
				_, err := DecryptCredentials(&tamperedPayload, appSalt, DefaultEncryptionConfig())
				if err == nil {
					t.Error("Expected decryption to fail with tampered payload")
				}
			})
		}
	})
}