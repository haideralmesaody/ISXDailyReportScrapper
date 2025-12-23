package license

import (
	"testing"
	"time"

	"github.com/isxcli/isxcli/internal/security"
)

func TestValidateScratchCardFormat(t *testing.T) {
	tests := []struct {
		name        string
		licenseKey  string
		expectError bool
	}{
		{
			name:        "valid scratch card with dashes",
			licenseKey:  "ISX-1M23-4567-890A",
			expectError: false,
		},
		{
			name:        "valid scratch card without dashes",
			licenseKey:  "ISX1M234567890A",
			expectError: false,
		},
		{
			name:        "invalid prefix",
			licenseKey:  "ABC-1M23-4567-890A",
			expectError: true,
		},
		{
			name:        "too short",
			licenseKey:  "ISX-123-456-789",
			expectError: true,
		},
		{
			name:        "too long",
			licenseKey:  "ISX-1M23-4567-890AB",
			expectError: true,
		},
		{
			name:        "invalid characters",
			licenseKey:  "ISX-1M23-456!-890A",
			expectError: true,
		},
		{
			name:        "lowercase",
			licenseKey:  "isx-1m23-4567-890a",
			expectError: true, // Should fail before normalization
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScratchCardFormat(tt.licenseKey)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for license key %s, but got none", tt.licenseKey)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for license key %s, but got: %v", tt.licenseKey, err)
			}
		})
	}
}

func TestNormalizeScratchCardKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with dashes",
			input:    "ISX-1M23-4567-890A",
			expected: "ISX1M234567890A",
		},
		{
			name:     "without dashes",
			input:    "ISX1M234567890A",
			expected: "ISX1M234567890A",
		},
		{
			name:     "with spaces",
			input:    "ISX 1M23 4567 890A",
			expected: "ISX1M234567890A",
		},
		{
			name:     "lowercase",
			input:    "isx-1m23-4567-890a",
			expected: "ISX1M234567890A",
		},
		{
			name:     "mixed case with spaces and dashes",
			input:    "isx-1M23 4567-890a",
			expected: "ISX1M234567890A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeScratchCardKey(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatScratchCardKeyWithDashes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normalized key",
			input:    "ISX1M234567890A",
			expected: "ISX-1M23-4567-890A",
		},
		{
			name:     "already formatted",
			input:    "ISX-1M23-4567-890A",
			expected: "ISX-1M23-4567-890A",
		},
		{
			name:     "invalid length",
			input:    "ISX123",
			expected: "ISX123", // Return as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatScratchCardKeyWithDashes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeviceFingerprintGeneration(t *testing.T) {
	manager := security.NewFingerprintManager()

	// Test fingerprint generation
	fingerprint1, err := manager.GenerateFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	if fingerprint1.Fingerprint == "" {
		t.Error("Generated fingerprint is empty")
	}

	if fingerprint1.OS == "" {
		t.Error("OS information is empty")
	}

	// Test that fingerprints are consistent
	fingerprint2, err := manager.GenerateFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate second fingerprint: %v", err)
	}

	if fingerprint1.Fingerprint != fingerprint2.Fingerprint {
		t.Error("Fingerprints should be consistent")
	}

	// Test fingerprint validation
	isValid, err := manager.ValidateFingerprint(fingerprint1.Fingerprint)
	if err != nil {
		t.Fatalf("Failed to validate fingerprint: %v", err)
	}

	if !isValid {
		t.Error("Fingerprint validation should pass for same device")
	}

	// Test with invalid fingerprint
	isValid, err = manager.ValidateFingerprint("invalid-fingerprint")
	if err != nil {
		t.Fatalf("Failed to validate invalid fingerprint: %v", err)
	}

	if isValid {
		t.Error("Invalid fingerprint should not validate")
	}
}

func TestDeviceFingerprintComponents(t *testing.T) {
	manager := security.NewFingerprintManager()

	components, err := manager.GetFingerprintComponents()
	if err != nil {
		t.Fatalf("Failed to get fingerprint components: %v", err)
	}

	// Check that we have expected components
	expectedKeys := []string{"mac_address", "hostname", "cpu_id", "os", "platform"}
	for _, key := range expectedKeys {
		if _, exists := components[key]; !exists {
			t.Errorf("Missing expected component: %s", key)
		}
	}

	// Check that components are not empty (allowing for fallbacks)
	if components["os"] == "" {
		t.Error("OS component should not be empty")
	}

	if components["platform"] == "" {
		t.Error("Platform component should not be empty")
	}
}

func TestLicenseInfoWithNewFields(t *testing.T) {
	// Test that LicenseInfo includes new fields
	license := LicenseInfo{
		LicenseKey:        "ISX1M234567890A",
		UserEmail:         "",
		ExpiryDate:        time.Now().Add(30 * 24 * time.Hour),
		Duration:          "1m",
		IssuedDate:        time.Now(),
		Status:            "Activated",
		LastChecked:       time.Now(),
		ActivationID:      "act_12345",
		DeviceFingerprint: "abc123def456",
	}

	if license.ActivationID == "" {
		t.Error("ActivationID field should be accessible")
	}

	if license.DeviceFingerprint == "" {
		t.Error("DeviceFingerprint field should be accessible")
	}

	// Test that scratch card format works
	if err := ValidateScratchCardFormat(license.LicenseKey); err != nil {
		t.Errorf("Valid scratch card license key should pass validation: %v", err)
	}
}

func TestCalculateExpiryDateFromDuration(t *testing.T) {
	// Create a minimal manager for testing
	manager := &Manager{}

	tests := []struct {
		name              string
		duration          string
		expectedMonthsAdd int
		expectedYearsAdd  int
	}{
		{"1 month", "1m", 1, 0},
		{"3 months", "3m", 3, 0},
		{"6 months", "6m", 6, 0},
		{"1 year", "1y", 0, 1},
		{"unknown default", "unknown", 1, 0}, // Should default to 1 month
	}

	baseTime := time.Now()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.calculateExpiryDateFromDuration(tt.duration)

			expected := baseTime.AddDate(tt.expectedYearsAdd, tt.expectedMonthsAdd, 1) // +1 day as per implementation
			expected = time.Date(expected.Year(), expected.Month(), expected.Day(), 0, 0, 0, 0, expected.Location())

			// Allow for small time differences (test might run across time boundaries)
			timeDiff := result.Sub(expected)
			if timeDiff > time.Minute || timeDiff < -time.Minute {
				t.Errorf("Expected expiry around %v, got %v (diff: %v)", expected, result, timeDiff)
			}
		})
	}
}
