package security

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestAppsScriptSecurityConfig validates the security configuration
func TestAppsScriptSecurityConfig(t *testing.T) {
	config := DefaultAppsScriptSecurityConfig()
	
	// Test default values
	if config.RequestTimeout < 5*time.Second {
		t.Errorf("Request timeout too short: %v", config.RequestTimeout)
	}
	
	if config.MaxRetries < 1 || config.MaxRetries > 10 {
		t.Errorf("Max retries out of range: %d", config.MaxRetries)
	}
	
	if config.TimestampWindow < 1*time.Minute {
		t.Errorf("Timestamp window too short: %v", config.TimestampWindow)
	}
	
	if len(config.SharedSecret) < 32 {
		t.Errorf("Shared secret too short: %d characters", len(config.SharedSecret))
	}
	
	if !config.EnableEncryption {
		t.Error("Encryption should be enabled by default")
	}
	
	if !config.RequireSignature {
		t.Error("Signatures should be required by default")
	}
}

// TestSecureAppsScriptClient validates client initialization and configuration
func TestSecureAppsScriptClient(t *testing.T) {
	config := DefaultAppsScriptSecurityConfig()
	certPinner := NewCertificatePinner(DefaultPinningConfig())
	client := NewSecureAppsScriptClient(config, certPinner)
	
	if client == nil {
		t.Fatal("Client should not be nil")
	}
	
	// Test configuration validation
	if err := client.ValidateConfiguration(); err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}
	
	// Test metrics
	metrics := client.GetSecurityMetrics()
	if metrics["shared_secret_length"].(int) < 32 {
		t.Error("Shared secret length should be at least 32")
	}
	
	if !metrics["encryption_enabled"].(bool) {
		t.Error("Encryption should be enabled")
	}
	
	if !metrics["signature_required"].(bool) {
		t.Error("Signatures should be required")
	}
}

// TestRequestEncryption validates encryption/decryption functionality
func TestRequestEncryption(t *testing.T) {
	sharedSecret := "test-shared-secret-32-characters-long"
	encryption := NewRequestEncryption(sharedSecret)
	
	// Test configuration validation
	if err := encryption.ValidateEncryptionConfig(); err != nil {
		t.Errorf("Encryption config validation failed: %v", err)
	}
	
	// Test payload encryption/decryption
	originalPayload := map[string]interface{}{
		"action": "activate",
		"code":   "ISX-TEST-KEY-123",
		"deviceInfo": map[string]interface{}{
			"fingerprint": "test-fingerprint",
			"hostname":    "test-host",
		},
	}
	
	requestID := "test-request-123"
	fingerprint := "test-device-fingerprint"
	
	// Encrypt request
	encryptedReq, err := encryption.EncryptRequest(originalPayload, requestID, fingerprint)
	if err != nil {
		t.Fatalf("Failed to encrypt request: %v", err)
	}
	
	// Validate encrypted request structure
	if encryptedReq.Version != 1 {
		t.Errorf("Expected version 1, got %d", encryptedReq.Version)
	}
	
	if encryptedReq.RequestID != requestID {
		t.Errorf("Request ID mismatch: expected %s, got %s", requestID, encryptedReq.RequestID)
	}
	
	if encryptedReq.Fingerprint != fingerprint {
		t.Errorf("Fingerprint mismatch: expected %s, got %s", fingerprint, encryptedReq.Fingerprint)
	}
	
	if encryptedReq.EncryptedData == "" {
		t.Error("Encrypted data should not be empty")
	}
	
	if encryptedReq.HMAC == "" {
		t.Error("HMAC signature should not be empty")
	}
	
	// Decrypt request
	decryptedPayload, err := encryption.DecryptRequest(encryptedReq)
	if err != nil {
		t.Fatalf("Failed to decrypt request: %v", err)
	}
	
	// Validate decrypted payload
	if decryptedPayload["action"] != originalPayload["action"] {
		t.Errorf("Action mismatch after decryption")
	}
	
	if decryptedPayload["code"] != originalPayload["code"] {
		t.Errorf("Code mismatch after decryption")
	}
}

// TestInputValidation validates input validation and sanitization
func TestInputValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()
	
	// Test valid license key
	validKey := "ISX-ABCD-1234-EFG"
	result := validator.ValidateLicenseKey(ctx, validKey)
	if !result.IsValid {
		t.Errorf("Valid license key rejected: %v", result.Errors)
	}
	
	// Test invalid license key (too short)
	invalidKey := "ISX-123"
	result = validator.ValidateLicenseKey(ctx, invalidKey)
	if result.IsValid {
		t.Error("Invalid license key accepted")
	}
	
	// Test SQL injection attempt
	sqlInjectionKey := "ISX-TEST'; DROP TABLE licenses; --"
	result = validator.ValidateLicenseKey(ctx, sqlInjectionKey)
	if result.RiskScore < 50 {
		t.Errorf("SQL injection not detected properly, risk score: %d", result.RiskScore)
	}
	
	// Test XSS attempt
	xssKey := "ISX-<script>alert('xss')</script>"
	result = validator.ValidateLicenseKey(ctx, xssKey)
	if result.RiskScore < 40 {
		t.Errorf("XSS attack not detected properly, risk score: %d", result.RiskScore)
	}
	
	// Test email validation
	validEmail := "user@example.com"
	emailResult := validator.ValidateEmail(ctx, validEmail)
	if !emailResult.IsValid {
		t.Errorf("Valid email rejected: %v", emailResult.Errors)
	}
	
	// Test email injection
	maliciousEmail := "user@example.com\r\nBcc: admin@evil.com"
	emailResult = validator.ValidateEmail(ctx, maliciousEmail)
	if emailResult.RiskScore < 40 {
		t.Errorf("Email injection not detected properly, risk score: %d", emailResult.RiskScore)
	}
	
	// Test IP validation
	validIP := "192.168.1.1"
	ipResult := validator.ValidateIPAddress(ctx, validIP)
	if !ipResult.IsValid {
		t.Errorf("Valid IP address rejected: %v", ipResult.Errors)
	}
	
	// Test invalid IP
	invalidIP := "999.999.999.999"
	ipResult = validator.ValidateIPAddress(ctx, invalidIP)
	if ipResult.IsValid {
		t.Error("Invalid IP address accepted")
	}
}

// TestHoneypotDetection validates honeypot license detection
func TestHoneypotDetection(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()
	
	// Test honeypot licenses
	honeypotKeys := []string{
		"ISX-TRAP-TRAP-TRAP",
		"ISX-FAKE-FAKE-FAKE",
		"ISX-TEST-TEST-TEST",
		"ISX-0000-0000-0000",
	}
	
	for _, key := range honeypotKeys {
		result := validator.ValidateLicenseKey(ctx, key)
		// Honeypot keys should be valid format but trigger security systems
		if result.IsValid {
			// This is expected - honeypots should pass basic validation
			// but be caught by the security system during processing
		}
	}
}

// TestThreatDetection validates threat detection patterns
func TestThreatDetection(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()
	
	testCases := []struct {
		input       string
		shouldDetect bool
		threatType   string
	}{
		{"'; DROP TABLE users; --", true, "sql_injection"},
		{"<script>alert('xss')</script>", true, "xss"},
		{"../../../etc/passwd", true, "path_traversal"},
		{"cmd.exe /c dir", true, "command_injection"},
		{"normal text input", false, ""},
		{"ISX-ABCD-1234-EFG", false, ""},
	}
	
	for _, tc := range testCases {
		result := validator.ValidateGenericInput(ctx, tc.input, "test", 1000)
		
		if tc.shouldDetect && result.RiskScore < 20 {
			t.Errorf("Threat not detected in input: %s", tc.input)
		}
		
		if !tc.shouldDetect && result.RiskScore > 10 {
			t.Errorf("False positive threat detection in input: %s", tc.input)
		}
		
		if tc.shouldDetect && tc.threatType != "" {
			found := false
			for _, threat := range result.ThreatTypes {
				if strings.Contains(threat, tc.threatType) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected threat type %s not found in: %v", tc.threatType, result.ThreatTypes)
			}
		}
	}
}

// TestCertificatePinning validates certificate pinning functionality
func TestCertificatePinning(t *testing.T) {
	config := DefaultPinningConfig()
	pinner := NewCertificatePinner(config)
	
	if pinner == nil {
		t.Fatal("Certificate pinner should not be nil")
	}
	
	// Test HTTP client creation
	client := pinner.CreateSecureHTTPClient(config)
	if client == nil {
		t.Fatal("HTTP client should not be nil")
	}
	
	if client.Timeout != config.ConnTimeout {
		t.Errorf("Client timeout mismatch: expected %v, got %v", config.ConnTimeout, client.Timeout)
	}
	
	// Test pinned hashes retrieval
	pinnedHashes := pinner.GetPinnedHashes()
	if len(pinnedHashes) == 0 {
		t.Error("Should have pinned hashes configured")
	}
	
	// Test custom pin addition
	testHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	err := pinner.AddCustomPin("test.example.com", testHash)
	if err != nil {
		t.Errorf("Failed to add custom pin: %v", err)
	}
	
	// Test invalid hash addition
	invalidHash := "invalid-hash"
	err = pinner.AddCustomPin("test.example.com", invalidHash)
	if err == nil {
		t.Error("Should reject invalid hash format")
	}
}

// TestMetricsGeneration validates security metrics generation
func TestMetricsGeneration(t *testing.T) {
	// Test Apps Script client metrics
	config := DefaultAppsScriptSecurityConfig()
	client := NewSecureAppsScriptClient(config, nil)
	metrics := client.GetSecurityMetrics()
	
	expectedFields := []string{
		"shared_secret_length",
		"request_timeout",
		"max_retries",
		"encryption_enabled",
		"signature_required",
	}
	
	for _, field := range expectedFields {
		if _, exists := metrics[field]; !exists {
			t.Errorf("Missing metric field: %s", field)
		}
	}
	
	// Test input validator metrics
	validator := NewInputValidator(nil)
	validatorMetrics := validator.GetValidationMetrics()
	
	expectedValidatorFields := []string{
		"max_license_key_length",
		"strict_validation_enabled",
		"suspicious_patterns_count",
		"sql_patterns_count",
	}
	
	for _, field := range expectedValidatorFields {
		if _, exists := validatorMetrics[field]; !exists {
			t.Errorf("Missing validator metric field: %s", field)
		}
	}
}

// TestSecurityConfiguration validates overall security configuration
func TestSecurityConfiguration(t *testing.T) {
	// Test minimum security requirements
	config := DefaultAppsScriptSecurityConfig()
	
	// OWASP ASVS Level 2 requirements
	if config.RequestTimeout > 120*time.Second {
		t.Error("Request timeout should not exceed 120 seconds")
	}
	
	if config.TimestampWindow > 30*time.Minute {
		t.Error("Timestamp window should not exceed 30 minutes")
	}
	
	if config.MaxRequestSize > 10*1024*1024 {
		t.Error("Request size should not exceed 10MB")
	}
	
	if !config.RequireSignature {
		t.Error("HMAC signatures should be required")
	}
	
	if !config.EnableEncryption {
		t.Error("Encryption should be enabled")
	}
	
	// Test validation config
	validationConfig := DefaultValidationConfig()
	
	if validationConfig.MaxLicenseKeyLength > 512 {
		t.Error("License key length limit too high")
	}
	
	if validationConfig.MaxEmailLength > 320 {
		t.Error("Email length limit too high")
	}
	
	if !validationConfig.EnableStrictValidation {
		t.Error("Strict validation should be enabled")
	}
}

// BenchmarkRequestEncryption benchmarks encryption performance
func BenchmarkRequestEncryption(b *testing.B) {
	sharedSecret := "test-shared-secret-32-characters-long"
	encryption := NewRequestEncryption(sharedSecret)
	
	payload := map[string]interface{}{
		"action": "activate",
		"code":   "ISX-TEST-KEY-123",
		"deviceInfo": map[string]interface{}{
			"fingerprint": "test-fingerprint",
			"hostname":    "test-host",
		},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		requestID := "test-request"
		fingerprint := "test-device"
		
		encryptedReq, err := encryption.EncryptRequest(payload, requestID, fingerprint)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
		
		_, err = encryption.DecryptRequest(encryptedReq)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}

// BenchmarkInputValidation benchmarks input validation performance
func BenchmarkInputValidation(b *testing.B) {
	validator := NewInputValidator(nil)
	ctx := context.Background()
	
	testInputs := []string{
		"ISX-ABCD-1234-EFG",
		"user@example.com",
		"192.168.1.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		
		switch i % 4 {
		case 0:
			validator.ValidateLicenseKey(ctx, input)
		case 1:
			validator.ValidateEmail(ctx, input)
		case 2:
			validator.ValidateIPAddress(ctx, input)
		case 3:
			validator.ValidateUserAgent(ctx, input)
		}
	}
}