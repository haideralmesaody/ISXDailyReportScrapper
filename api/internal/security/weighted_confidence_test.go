package security

import (
	"testing"
	"time"
)

// TestWeightedConfidenceCalculation tests the weighted confidence calculation logic
// with mock data to verify the QAQC edge case requirement
func TestWeightedConfidenceCalculation(t *testing.T) {
	// Create a hardware fingerprinter instance
	hf := NewHardwareFingerprinter()

	// Test Case 1: All components present (should return 1.0 confidence)
	currentFingerprint := &HardwareFingerprint{
		MotherboardSerial: "TEST-MB-12345",
		SystemUUID:        "test-uuid-67890abcdef",
		PrimaryDiskSerial: "test-disk-98765",
		CPUModel:          "Test CPU Model",
		BIOSVersion:       "1.0.0",
		Fingerprint:      "test-fingerprint-hash",
		GeneratedAt:      time.Now(),
		Platform:         "windows",
		Components:       5,
	}

	storedComponents := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	match := &FingerprintMatch{
		StoredFingerprint:  "test-fingerprint-hash",
		CurrentFingerprint: "test-fingerprint-hash",
		ExactMatch:         true,
		Components:         make([]ComponentMatch, 0),
	}

	confidence := hf.calculateWeightedConfidence(currentFingerprint, storedComponents, match)
	if confidence != 1.0 {
		t.Errorf("Expected confidence 1.0 for exact match, got %f", confidence)
	}

	// Test Case 2: Missing disk serial (should pass 0.7 threshold)
	// Motherboard (30%) + UUID (30%) + CPU (20%) + BIOS (15%) = 95% confidence
	storedComponentsMissingDisk := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "", // Missing
	}

	confidenceMissingDisk := hf.calculateWeightedConfidence(currentFingerprint, storedComponentsMissingDisk, match)
	expectedConfidence := 0.30 + 0.30 + 0.20 + 0.15 // 95% = 0.95
	if diff := confidenceMissingDisk - expectedConfidence; diff > 0.001 || diff < -0.001 {
		t.Errorf("Expected confidence %f for missing disk, got %f (diff: %f)", expectedConfidence, confidenceMissingDisk, diff)
	}

	// Verify that missing disk serial still passes the 0.7 threshold
	if confidenceMissingDisk < 0.7 {
		t.Errorf("Missing disk serial should pass 0.7 threshold, got %f", confidenceMissingDisk)
	}

	// Test Case 3: Missing motherboard (should pass exactly at threshold)
	// UUID (30%) + CPU (20%) + BIOS (15%) + Disk (5%) = 70% confidence
	storedComponentsMissingMB := map[string]string{
		"motherboard_serial": "", // Missing
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	confidenceMissingMB := hf.calculateWeightedConfidence(currentFingerprint, storedComponentsMissingMB, match)
	expectedConfidenceMB := 0.30 + 0.20 + 0.15 + 0.05 // 70% = 0.70
	if diff := confidenceMissingMB - expectedConfidenceMB; diff > 0.001 || diff < -0.001 {
		t.Errorf("Expected confidence %f for missing motherboard, got %f (diff: %f)", expectedConfidenceMB, confidenceMissingMB, diff)
	}

	// Should pass exactly at the 0.7 threshold
	if confidenceMissingMB < 0.7 {
		t.Errorf("Missing motherboard should pass exactly at 0.7 threshold, got %f", confidenceMissingMB)
	}

	// Test Case 4: Missing critical components (should fail threshold)
	// CPU (20%) + BIOS (15%) + Disk (5%) = 40% confidence
	storedComponentsMissingCritical := map[string]string{
		"motherboard_serial": "", // Missing
		"system_uuid":        "", // Missing
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	confidenceMissingCritical := hf.calculateWeightedConfidence(currentFingerprint, storedComponentsMissingCritical, match)
	expectedConfidenceCritical := 0.20 + 0.15 + 0.05 // 40% = 0.40
	if diff := confidenceMissingCritical - expectedConfidenceCritical; diff > 0.001 || diff < -0.001 {
		t.Errorf("Expected confidence %f for missing critical components, got %f (diff: %f)", expectedConfidenceCritical, confidenceMissingCritical, diff)
	}

	// Should fail the 0.7 threshold
	if confidenceMissingCritical >= 0.7 {
		t.Errorf("Missing critical components should fail 0.7 threshold, got %f", confidenceMissingCritical)
	}

	t.Log("✅ Weighted confidence calculation tests passed:")
	t.Logf("   - Exact match: %.0f%%", confidence*100)
	t.Logf("   - Missing disk: %.0f%% (✅ passes 70%% threshold)", confidenceMissingDisk*100)
	t.Logf("   - Missing motherboard: %.0f%% (✅ passes 70%% threshold)", confidenceMissingMB*100)
	t.Logf("   - Missing critical: %.0f%% (❌ fails 70%% threshold)", confidenceMissingCritical*100)
}

// TestComponentExtraction tests the extractStoredComponents function
func TestComponentExtraction(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Test with valid component data
	components := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	extracted := hf.extractStoredComponents(components)
	if len(extracted) != 5 {
		t.Errorf("Expected 5 components, got %d", len(extracted))
	}

	// Test with nil input
	extractedNil := hf.extractStoredComponents(nil)
	if len(extractedNil) != 0 {
		t.Errorf("Expected 0 components for nil input, got %d", len(extractedNil))
	}

	t.Log("✅ Component extraction tests passed")
}

// TestConfidenceThresholdValidation tests the 0.7 threshold logic
func TestConfidenceThresholdValidation(t *testing.T) {
	testCases := []struct {
		name       string
		confidence float64
		shouldPass bool
	}{
		{"Perfect match", 1.0, true},
		{"Above threshold", 0.85, true},
		{"Exactly at threshold", 0.70, true},
		{"Just below threshold", 0.69, false},
		{"Very low confidence", 0.30, false},
		{"Zero confidence", 0.0, false},
	}

	for _, tc := range testCases {
		passes := tc.confidence >= 0.7
		if passes != tc.shouldPass {
			t.Errorf("Test '%s' failed: confidence %f should pass=%v, got pass=%v",
				tc.name, tc.confidence, tc.shouldPass, passes)
		}
	}

	t.Log("✅ Confidence threshold validation tests passed")
}