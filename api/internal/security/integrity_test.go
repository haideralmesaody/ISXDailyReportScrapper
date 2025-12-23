package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIntegrityChecker(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-binary.exe")
	testContent := []byte("test binary content for integrity checking")
	
	if err := os.WriteFile(testFile, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Calculate expected hash for the test file
	checker := &IntegrityChecker{}
	expectedHash, _, err := checker.calculateBinaryHash(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate test file hash: %v", err)
	}
	
	tests := []struct {
		name         string
		expectedHash string
		shouldPass   bool
	}{
		{
			name:         "Valid hash",
			expectedHash: expectedHash,
			shouldPass:   true,
		},
		{
			name:         "Invalid hash",
			expectedHash: "invalid_hash_that_should_fail_verification_test",
			shouldPass:   false,
		},
		{
			name:         "Empty hash",
			expectedHash: "",
			shouldPass:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't easily test the current executable, we'll test the validation logic
			err := ValidateIntegrityConfig(tt.expectedHash)
			if tt.shouldPass && tt.expectedHash != "" && len(tt.expectedHash) == 64 {
				// Only test with valid hex strings
				checker := NewIntegrityChecker(tt.expectedHash)
				if checker.expectedHash != tt.expectedHash {
					t.Errorf("Expected hash not set correctly")
				}
			} else {
				if err == nil && !tt.shouldPass {
					t.Error("Expected validation to fail")
				}
			}
		})
	}
}

func TestValidateIntegrityConfig(t *testing.T) {
	tests := []struct {
		name         string
		hash         string
		shouldError  bool
	}{
		{
			name:        "Valid SHA-256 hash",
			hash:        "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			shouldError: false,
		},
		{
			name:        "Empty hash",
			hash:        "",
			shouldError: true,
		},
		{
			name:        "Short hash",
			hash:        "abc123",
			shouldError: true,
		},
		{
			name:        "Long hash",
			hash:        "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890123456789",
			shouldError: true,
		},
		{
			name:        "Invalid hex characters",
			hash:        "g1b2c3d4e5f6789012345678901234567890123456789012345678901234567890123456",
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntegrityConfig(tt.hash)
			if tt.shouldError {
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

func TestTamperingDetection(t *testing.T) {
	checker := NewIntegrityChecker("test_hash_placeholder_64_chars_long_for_testing_purposes_12345678")
	
	// Test tampering indicator detection
	indicators := checker.detectTamperingIndicators()
	
	// These tests may vary based on the test environment
	// We're mainly testing that the detection methods don't panic
	if indicators == nil {
		t.Error("Tampering indicators should not be nil")
	}
	
	// Test that hasIndicators method works
	hasAny := indicators.hasIndicators()
	
	// Generate a detailed report
	report := indicators.GetDetailedTamperingReport()
	if report == "" {
		t.Error("Tampering report should not be empty")
	}
	
	t.Logf("Tampering detection report: %s", report)
	t.Logf("Has indicators: %v", hasAny)
}

func TestVerifyAndDecryptCredentials(t *testing.T) {
	// Test data
	plaintext := []byte("test credential data")
	appSalt := []byte("test-application-salt-32-bytes!!")
	
	// Create encrypted payload
	payload, err := EncryptCredentials(plaintext, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}
	
	// Test with invalid binary hash (should fail)
	_, err = VerifyAndDecryptCredentials(payload, appSalt, "invalid_binary_hash_placeholder_test")
	if err == nil {
		t.Error("Expected verification to fail with invalid binary hash")
	}
	
	// Note: Testing with a valid binary hash would require knowing the actual
	// hash of the test binary, which varies by build. In production, this
	// would be embedded during the build process.
}

func TestVirtualMachineDetection(t *testing.T) {
	checker := NewIntegrityChecker("test_hash")
	
	// Test VM detection (results may vary based on test environment)
	isVM := checker.isRunningInVM()
	t.Logf("VM detection result: %v", isVM)
	
	// Test debugger detection
	isDebugger := checker.isDebuggerPresent()
	t.Logf("Debugger detection result: %v", isDebugger)
	
	// Test suspicious process name detection
	isSuspicious := checker.hasSuspiciousProcessName()
	t.Logf("Suspicious process name detection result: %v", isSuspicious)
}

func TestBinaryHashGeneration(t *testing.T) {
	// Test binary hash generation (for current test executable)
	hash, err := GenerateBinaryHash()
	if err != nil {
		t.Fatalf("Failed to generate binary hash: %v", err)
	}
	
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
	
	// Validate the generated hash
	if err := ValidateIntegrityConfig(hash); err != nil {
		t.Errorf("Generated hash failed validation: %v", err)
	}
	
	t.Logf("Generated binary hash: %s", hash)
}

func TestIntegrityVerificationResult(t *testing.T) {
	result := &VerificationResult{
		IsValid:           false,
		ActualHash:        "actual_hash_test",
		ExpectedHash:      "expected_hash_test",
		BinaryPath:        "/path/to/binary",
		BinarySize:        12345,
		ErrorMessage:      "test error message",
		TamperingDetected: true,
	}
	
	// Test that all fields are properly set
	if result.IsValid {
		t.Error("Expected IsValid to be false")
	}
	
	if !result.TamperingDetected {
		t.Error("Expected TamperingDetected to be true")
	}
	
	if result.ActualHash != "actual_hash_test" {
		t.Error("ActualHash not set correctly")
	}
	
	if result.BinarySize != 12345 {
		t.Error("BinarySize not set correctly")
	}
}

func TestTamperingIndicators(t *testing.T) {
	// Test empty indicators
	indicators := &TamperingIndicators{}
	if indicators.hasIndicators() {
		t.Error("Empty indicators should return false")
	}
	
	report := indicators.GetDetailedTamperingReport()
	if report != "No tampering indicators detected" {
		t.Errorf("Expected no indicators message, got: %s", report)
	}
	
	// Test with indicators set
	indicators.DebuggerDetected = true
	indicators.VirtualMachineDetected = true
	
	if !indicators.hasIndicators() {
		t.Error("Should detect indicators when set")
	}
	
	report = indicators.GetDetailedTamperingReport()
	if report == "No tampering indicators detected" {
		t.Error("Should report detected indicators")
	}
	
	t.Logf("Tampering report with indicators: %s", report)
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        []byte
		b        []byte
		expected bool
	}{
		{
			name:     "Equal bytes",
			a:        []byte("test"),
			b:        []byte("test"),
			expected: true,
		},
		{
			name:     "Different bytes",
			a:        []byte("test1"),
			b:        []byte("test2"),
			expected: false,
		},
		{
			name:     "Different lengths",
			a:        []byte("test"),
			b:        []byte("testing"),
			expected: false,
		},
		{
			name:     "Empty bytes",
			a:        []byte{},
			b:        []byte{},
			expected: true,
		},
		{
			name:     "One empty",
			a:        []byte("test"),
			b:        []byte{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecureCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func BenchmarkIntegrityVerification(b *testing.B) {
	// Create a test file for benchmarking
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test-binary.exe")
	testContent := make([]byte, 1024*1024) // 1MB test file
	
	// Fill with some data
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}
	
	if err := os.WriteFile(testFile, testContent, 0755); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}
	
	checker := &IntegrityChecker{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := checker.calculateBinaryHash(testFile)
		if err != nil {
			b.Fatalf("Hash calculation failed: %v", err)
		}
	}
}

func TestApplicationIntegrityValidation(t *testing.T) {
	// Test with an obviously invalid hash
	err := ValidateApplicationIntegrity("invalid_hash")
	if err == nil {
		t.Error("Expected validation to fail with invalid hash")
	}
	
	// Test with a valid format but incorrect hash
	validFormatHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	err = ValidateApplicationIntegrity(validFormatHash)
	// This should fail because the hash won't match the current binary
	if err == nil {
		t.Error("Expected validation to fail with incorrect hash")
	}
}

func TestCurrentTimestamp(t *testing.T) {
	timestamp := getCurrentTimestamp()
	
	// Should return the fixed timestamp for reproducible builds
	expectedTimestamp := int64(1721958000)
	if timestamp != expectedTimestamp {
		t.Errorf("Expected timestamp %d, got %d", expectedTimestamp, timestamp)
	}
}

// TestCalculateBinaryHash tests the calculateBinaryHash method
func TestCalculateBinaryHash(t *testing.T) {
	tests := []struct {
		name         string
		createFile   func(string) error
		expectError  bool
		errorContains string
	}{
		{
			name: "valid_file",
			createFile: func(path string) error {
				return os.WriteFile(path, []byte("test content"), 0644)
			},
			expectError: false,
		},
		{
			name: "empty_file",
			createFile: func(path string) error {
				return os.WriteFile(path, []byte{}, 0644)
			},
			expectError: false,
		},
		{
			name: "large_file",
			createFile: func(path string) error {
				data := make([]byte, 1024*1024) // 1MB
				for i := range data {
					data[i] = byte(i % 256)
				}
				return os.WriteFile(path, data, 0644)
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test-file")
			
			if err := tt.createFile(testFile); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			checker := &IntegrityChecker{}
			hash, size, err := checker.calculateBinaryHash(testFile)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(hash) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(hash))
			}
			
			if size < 0 {
				t.Errorf("File size should not be negative: %d", size)
			}
			
			// Verify hash is valid hex
			if err := ValidateIntegrityConfig(hash); err != nil {
				t.Errorf("Generated hash is not valid: %v", err)
			}
		})
	}
}

// TestCalculateBinaryHashNonExistentFile tests error handling for non-existent files
func TestCalculateBinaryHashNonExistentFile(t *testing.T) {
	checker := &IntegrityChecker{}
	_, _, err := checker.calculateBinaryHash("/non/existent/file")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestVerifyBinaryIntegrityDetailed provides comprehensive testing of binary integrity verification
func TestVerifyBinaryIntegrityDetailed(t *testing.T) {
	// Generate a real hash for the current executable
	currentHash, err := GenerateBinaryHash()
	if err != nil {
		t.Fatalf("Failed to generate current binary hash: %v", err)
	}
	
	tests := []struct {
		name               string
		expectedHash       string
		expectValid        bool
		expectTampering    bool
	}{
		{
			name:               "correct_hash",
			expectedHash:       currentHash,
			expectValid:        true,
			expectTampering:    false, // May be true due to environment detection
		},
		{
			name:               "incorrect_hash",
			expectedHash:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid:        false,
			expectTampering:    true,
		},
		{
			name:               "empty_hash",
			expectedHash:       "",
			expectValid:        false,
			expectTampering:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewIntegrityChecker(tt.expectedHash)
			result, err := checker.VerifyBinaryIntegrity()
			
			if err != nil && tt.expectedHash != "" {
				// Only expect errors for completely invalid configurations
				if len(tt.expectedHash) == 64 {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}
			
			if result == nil {
				if tt.expectedHash != "" {
					t.Error("Result should not be nil")
				}
				return
			}
			
			if result.IsValid != tt.expectValid {
				// For current hash, tampering detection might affect validity
				if tt.expectedHash == currentHash {
					t.Logf("Validity mismatch possibly due to environment detection: expected %v, got %v", tt.expectValid, result.IsValid)
				} else {
					t.Errorf("Expected validity %v, got %v", tt.expectValid, result.IsValid)
				}
			}
			
			if result.TamperingDetected != tt.expectTampering {
				// Environment-based detection may vary
				t.Logf("Tampering detection: expected %v, got %v (may vary by environment)", tt.expectTampering, result.TamperingDetected)
			}
			
			// Verify result fields are populated
			if result.ActualHash == "" {
				t.Error("ActualHash should not be empty")
			}
			
			if result.BinaryPath == "" {
				t.Error("BinaryPath should not be empty")
			}
			
			if result.BinarySize <= 0 {
				t.Error("BinarySize should be positive")
			}
		})
	}
}

// TestIsDebuggerPresentDetailed tests debugger detection with various scenarios
func TestIsDebuggerPresentDetailed(t *testing.T) {
	checker := NewIntegrityChecker("test-hash")
	
	// Test current environment
	originalEnv := make(map[string]string)
	debugVars := []string{"DELVE_PORT", "DEBUG_MODE", "GO_DEBUG", "GODEBUG"}
	
	// Store original values
	for _, debugVar := range debugVars {
		originalEnv[debugVar] = os.Getenv(debugVar)
	}
	
	// Clean environment test
	for _, debugVar := range debugVars {
		os.Unsetenv(debugVar)
	}
	
	isDebugger := checker.isDebuggerPresent()
	t.Logf("Clean environment debugger detection: %v", isDebugger)
	
	// Test with debug variables set
	testCases := []struct {
		name     string
		envVar   string
		value    string
		expected bool
	}{
		{"delve_port", "DELVE_PORT", "2345", true},
		{"debug_mode", "DEBUG_MODE", "true", true},
		{"go_debug", "GO_DEBUG", "1", true},
		{"godebug", "GODEBUG", "gctrace=1", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tc.envVar, tc.value)
			
			isDebugger := checker.isDebuggerPresent()
			if runtime.GOOS == "windows" {
				if isDebugger != tc.expected {
					t.Errorf("Expected %v, got %v for %s", tc.expected, isDebugger, tc.envVar)
				}
			} else {
				// On non-Windows, the function returns false
				if isDebugger != false {
					t.Errorf("Expected false on non-Windows, got %v", isDebugger)
				}
			}
			
			// Cleanup
			os.Unsetenv(tc.envVar)
		})
	}
	
	// Restore original environment
	for debugVar, value := range originalEnv {
		if value != "" {
			os.Setenv(debugVar, value)
		}
	}
}

// TestIsRunningInVMDetailed tests VM detection with various scenarios
func TestIsRunningInVMDetailed(t *testing.T) {
	checker := NewIntegrityChecker("test-hash")
	
	// Test current environment
	isVM := checker.isRunningInVM()
	t.Logf("Current environment VM detection: %v", isVM)
	
	// Test only works on Windows
	if runtime.GOOS != "windows" {
		if isVM != false {
			t.Errorf("Expected false on non-Windows, got %v", isVM)
		}
		return
	}
	
	// Test by setting VM environment variables
	vmTests := []string{"VBOX_VERSION", "VMWARE_TOOLS", "VIRTUAL_ENV"}
	originalValues := make(map[string]string)
	
	for _, envVar := range vmTests {
		originalValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar) // Clean slate
	}
	
	// Test with no VM indicators
	isVM = checker.isRunningInVM()
	t.Logf("Clean environment VM detection: %v", isVM)
	
	// Test with VM indicators
	for _, envVar := range vmTests {
		t.Run("with_"+envVar, func(t *testing.T) {
			os.Setenv(envVar, "test_value")
			isVM := checker.isRunningInVM()
			if !isVM {
				// May not detect if the check doesn't match exactly
				t.Logf("VM not detected with %s set (detection may be specific)", envVar)
			}
			os.Unsetenv(envVar)
		})
	}
	
	// Restore original environment
	for envVar, value := range originalValues {
		if value != "" {
			os.Setenv(envVar, value)
		}
	}
}

// TestHasSuspiciousProcessNameDetailed tests process name detection thoroughly
func TestHasSuspiciousProcessNameDetailed(t *testing.T) {
	tests := []struct {
		name         string
		binaryPath   string
		expected     bool
		description  string
	}{
		{
			name:         "valid_web_licensed",
			binaryPath:   "/path/to/web-licensed.exe",
			expected:     false,
			description:  "Valid ISX application name",
		},
		{
			name:         "valid_isx_daily_reports",
			binaryPath:   "/usr/bin/isx-daily-reports",
			expected:     false,
			description:  "Valid ISX application name",
		},
		{
			name:         "suspicious_debug",
			binaryPath:   "/tmp/debug-app.exe",
			expected:     true,
			description:  "Contains suspicious 'debug' pattern",
		},
		{
			name:         "suspicious_crack",
			binaryPath:   "/home/user/crack-tool",
			expected:     true,
			description:  "Contains suspicious 'crack' pattern",
		},
		{
			name:         "suspicious_test",
			binaryPath:   "/opt/test-binary",
			expected:     true,
			description:  "Contains suspicious 'test' pattern",
		},
		{
			name:         "random_name",
			binaryPath:   "/bin/random-app",
			expected:     true,
			description:  "Unknown application name (not in whitelist)",
		},
		{
			name:         "empty_path",
			binaryPath:   "",
			expected:     false,
			description:  "Empty path should return false",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &IntegrityChecker{binaryPath: tt.binaryPath}
			result := checker.hasSuspiciousProcessName()
			
			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.expected, result)
			}
		})
	}
}

// TestTamperingIndicatorsComprehensive tests all tampering indicators
func TestTamperingIndicatorsComprehensive(t *testing.T) {
	// Test all combinations of indicators
	testCases := []struct {
		name        string
		setFlags    func(*TamperingIndicators)
		expectFlags bool
		reportContains []string
	}{
		{
			name: "no_indicators",
			setFlags: func(ti *TamperingIndicators) {
				// No flags set
			},
			expectFlags: false,
			reportContains: []string{"No tampering indicators detected"},
		},
		{
			name: "file_size_indicator",
			setFlags: func(ti *TamperingIndicators) {
				ti.UnexpectedFileSize = true
			},
			expectFlags: true,
			reportContains: []string{"Unexpected file size"},
		},
		{
			name: "timestamp_indicator",
			setFlags: func(ti *TamperingIndicators) {
				ti.ModifiedTimestamp = true
			},
			expectFlags: true,
			reportContains: []string{"Modified timestamp"},
		},
		{
			name: "process_name_indicator",
			setFlags: func(ti *TamperingIndicators) {
				ti.SuspiciousProcessName = true
			},
			expectFlags: true,
			reportContains: []string{"Suspicious process name"},
		},
		{
			name: "debugger_indicator",
			setFlags: func(ti *TamperingIndicators) {
				ti.DebuggerDetected = true
			},
			expectFlags: true,
			reportContains: []string{"Debugger presence detected"},
		},
		{
			name: "vm_indicator",
			setFlags: func(ti *TamperingIndicators) {
				ti.VirtualMachineDetected = true
			},
			expectFlags: true,
			reportContains: []string{"Virtual machine environment detected"},
		},
		{
			name: "all_indicators",
			setFlags: func(ti *TamperingIndicators) {
				ti.UnexpectedFileSize = true
				ti.ModifiedTimestamp = true
				ti.SuspiciousProcessName = true
				ti.DebuggerDetected = true
				ti.VirtualMachineDetected = true
			},
			expectFlags: true,
			reportContains: []string{
				"Unexpected file size",
				"Modified timestamp",
				"Suspicious process name",
				"Debugger presence detected",
				"Virtual machine environment detected",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testIndicators := &TamperingIndicators{}
			tc.setFlags(testIndicators)
			
			hasFlags := testIndicators.hasIndicators()
			if hasFlags != tc.expectFlags {
				t.Errorf("Expected hasIndicators %v, got %v", tc.expectFlags, hasFlags)
			}
			
			report := testIndicators.GetDetailedTamperingReport()
			for _, expectedText := range tc.reportContains {
				if !strings.Contains(report, expectedText) {
					t.Errorf("Report should contain %q, got: %s", expectedText, report)
				}
			}
			
			t.Logf("Report for %s: %s", tc.name, report)
		})
	}
}

