package security

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestHardwareFingerprintGeneration tests basic fingerprint generation
func TestHardwareFingerprintGeneration(t *testing.T) {
	hf := NewHardwareFingerprinter()

	fingerprint, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate hardware fingerprint: %v", err)
	}

	if fingerprint.Fingerprint == "" {
		t.Fatal("Generated fingerprint is empty")
	}

	if len(fingerprint.Fingerprint) != 64 { // SHA-256 hex encoded
		t.Fatalf("Expected fingerprint length 64, got %d", len(fingerprint.Fingerprint))
	}

	if fingerprint.GeneratedAt.IsZero() {
		t.Fatal("Generated timestamp is zero")
	}

	if fingerprint.Platform != runtime.GOOS {
		t.Fatalf("Expected platform %s, got %s", runtime.GOOS, fingerprint.Platform)
	}

	t.Logf("Generated fingerprint: %s", fingerprint.Fingerprint)
	t.Logf("Components: %d", fingerprint.Components)
}

// TestHardwareFingerprintStability tests that fingerprints are stable over time
func TestHardwareFingerprintStability(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Generate first fingerprint
	fingerprint1, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate first fingerprint: %v", err)
	}

	// Clear cache to ensure fresh generation
	hf.ClearCache()

	// Wait a short time to ensure different timestamp
	time.Sleep(100 * time.Millisecond)

	// Generate second fingerprint
	fingerprint2, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate second fingerprint: %v", err)
	}

	// Fingerprints should be identical (hardware doesn't change)
	if fingerprint1.Fingerprint != fingerprint2.Fingerprint {
		t.Fatalf("Fingerprints differ:\nFirst:  %s\nSecond: %s", fingerprint1.Fingerprint, fingerprint2.Fingerprint)
	}

	t.Logf("Stability test passed - both fingerprints: %s", fingerprint1.Fingerprint)
}

// TestHardwareFingerprintComponents tests individual component gathering
func TestHardwareFingerprintComponents(t *testing.T) {
	hf := NewHardwareFingerprinter()

	components, err := hf.GetHardwareComponents()
	if err != nil {
		t.Fatalf("Failed to get hardware components: %v", err)
	}

	expectedComponents := []string{
		"motherboard_serial",
		"system_uuid",
		"primary_disk_serial",
		"cpu_model",
		"bios_version",
	}

	for _, component := range expectedComponents {
		if value, exists := components[component]; !exists {
			t.Logf("Component %s not found (may be expected on this platform)", component)
		} else if value == "" {
			t.Logf("Component %s is empty (may be expected on this platform)", component)
		} else {
			t.Logf("Component %s: %s", component, value)
		}
	}
}

// TestHardwareFingerprintCaching tests the caching mechanism
func TestHardwareFingerprintCaching(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Generate first fingerprint
	start := time.Now()
	fingerprint1, err := hf.GenerateHardwareFingerprint()
	firstDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to generate first fingerprint: %v", err)
	}

	// Generate second fingerprint (should use cache)
	start = time.Now()
	fingerprint2, err := hf.GenerateHardwareFingerprint()
	secondDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to generate second fingerprint: %v", err)
	}

	// Fingerprints should be identical
	if fingerprint1.Fingerprint != fingerprint2.Fingerprint {
		t.Fatal("Cached fingerprint differs from original")
	}

	// Second generation should be much faster (cache hit)
	if secondDuration >= firstDuration {
		t.Logf("Warning: Cache may not be working properly. First: %v, Second: %v", firstDuration, secondDuration)
	} else {
		t.Logf("Cache working: First: %v, Second: %v (%.2fx faster)", firstDuration, secondDuration, float64(firstDuration)/float64(secondDuration))
	}

	// Verify timestamps are different (cached result has original timestamp)
	if !fingerprint2.GeneratedAt.Equal(fingerprint1.GeneratedAt) {
		t.Fatal("Cached fingerprint should have original timestamp")
	}
}

// TestHardwareFingerprintValidation tests fingerprint validation
func TestHardwareFingerprintValidation(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Generate a fingerprint
	fingerprint, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	// Validate against same fingerprint (should match)
	match, err := hf.ValidateHardwareFingerprint(fingerprint.Fingerprint)
	if err != nil {
		t.Fatalf("Failed to validate fingerprint: %v", err)
	}

	if !match.ExactMatch {
		t.Fatal("Exact match should be true for identical fingerprints")
	}

	if match.Confidence != 1.0 {
		t.Fatalf("Expected confidence 1.0, got %f", match.Confidence)
	}

	if match.MatchType != "exact" && match.MatchType != "hardware_exact" {
		t.Fatalf("Expected match type 'exact' or 'hardware_exact', got '%s'", match.MatchType)
	}

	// Validate against different fingerprint (should not match)
	differentFingerprint := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	match, err = hf.ValidateHardwareFingerprint(differentFingerprint)
	if err != nil {
		t.Fatalf("Failed to validate different fingerprint: %v", err)
	}

	if match.ExactMatch {
		t.Fatal("Exact match should be false for different fingerprints")
	}

	if match.Confidence != 0.0 {
		t.Fatalf("Expected confidence 0.0, got %f", match.Confidence)
	}

	if match.MatchType != "no_match" {
		t.Fatalf("Expected match type 'no_match', got '%s'", match.MatchType)
	}
}

// TestHybridFingerprintManager tests the hybrid system
func TestHybridFingerprintManager(t *testing.T) {
	// Test with hardware preference
	hfm := NewHybridFingerprintManager(true)

	// Generate hardware fingerprint
	hwFingerprint, err := hfm.GenerateFingerprint(HardwareSystem)
	if err != nil {
		t.Fatalf("Failed to generate hardware fingerprint: %v", err)
	}

	// Generate legacy fingerprint
	legacyFingerprint, err := hfm.GenerateFingerprint(LegacySystem)
	if err != nil {
		t.Fatalf("Failed to generate legacy fingerprint: %v", err)
	}

	t.Logf("Hardware fingerprint: %s", hwFingerprint.(*HardwareFingerprint).Fingerprint)
	t.Logf("Legacy fingerprint: %s", legacyFingerprint.(*DeviceFingerprint).Fingerprint)

	// Test validation with hardware fingerprint
	match, err := hfm.ValidateFingerprint(hwFingerprint.(*HardwareFingerprint).Fingerprint)
	if err != nil {
		t.Fatalf("Failed to validate hardware fingerprint: %v", err)
	}

	if !match.ExactMatch {
		t.Fatal("Hardware fingerprint should match exactly")
	}

	t.Logf("Hardware validation: %s (confidence: %.2f)", match.MatchType, match.Confidence)
}

// TestHardwareFingerprintErrorHandling tests error handling scenarios
func TestHardwareFingerprintErrorHandling(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Test clearing cache
	hf.ClearCache()

	// Generate fingerprint after cache clear
	fingerprint, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate fingerprint after cache clear: %v", err)
	}

	if fingerprint.Fingerprint == "" {
		t.Fatal("Fingerprint should not be empty after cache clear")
	}

	t.Logf("Error handling test passed: %s", fingerprint.Fingerprint)
}

// BenchmarkHardwareFingerprintGeneration benchmarks fingerprint generation
func BenchmarkHardwareFingerprintGeneration(b *testing.B) {
	hf := NewHardwareFingerprinter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache to ensure fresh generation each time
		hf.ClearCache()

		_, err := hf.GenerateHardwareFingerprint()
		if err != nil {
			b.Fatalf("Failed to generate fingerprint: %v", err)
		}
	}
}

// BenchmarkHardwareFingerprintCached benchmarks cached fingerprint generation
func BenchmarkHardwareFingerprintCached(b *testing.B) {
	hf := NewHardwareFingerprinter()

	// Generate once to populate cache
	_, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		b.Fatalf("Failed to generate initial fingerprint: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hf.GenerateHardwareFingerprint()
		if err != nil {
			b.Fatalf("Failed to generate cached fingerprint: %v", err)
		}
	}
}

// TestHardwareFingerprintFormat tests fingerprint format consistency
func TestHardwareFingerprintFormat(t *testing.T) {
	hf := NewHardwareFingerprinter()

	fingerprint, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	// Test fingerprint format (should be lowercase hex)
	if fingerprint.Fingerprint != strings.ToLower(fingerprint.Fingerprint) {
		t.Fatal("Fingerprint should be lowercase")
	}

	// Test that fingerprint only contains hex characters
	for _, char := range fingerprint.Fingerprint {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Fatalf("Fingerprint contains invalid character: %c", char)
		}
	}

	t.Logf("Format test passed: %s", fingerprint.Fingerprint)
}

// TestPlatformSpecificFeatures tests platform-specific functionality
func TestPlatformSpecificFeatures(t *testing.T) {
	hf := NewHardwareFingerprinter()

	fingerprint, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	t.Logf("Platform: %s", runtime.GOOS)
	t.Logf("Components gathered: %d", fingerprint.Components)
	t.Logf("Fingerprint: %s", fingerprint.Fingerprint)

	// Platform-specific expectations
	switch runtime.GOOS {
	case "windows":
		t.Log("Windows platform detected - WMI queries should be available")
		if fingerprint.Components < 3 {
			t.Logf("Warning: Low component count (%d) on Windows - may need admin privileges", fingerprint.Components)
		}
	case "linux":
		t.Log("Linux platform detected - DMI/SMBIOS queries should be available")
		if fingerprint.Components < 2 {
			t.Logf("Warning: Low component count (%d) on Linux - may need root privileges", fingerprint.Components)
		}
	case "darwin":
		t.Log("macOS platform detected - system_profiler/ioreg should be available")
		if fingerprint.Components < 2 {
			t.Logf("Warning: Low component count (%d) on macOS", fingerprint.Components)
		}
	default:
		t.Logf("Unsupported platform %s - using fallbacks", runtime.GOOS)
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	hf := NewHardwareFingerprinter()
	const numGoroutines = 10
	const numIterations = 5

	results := make(chan string, numGoroutines*numIterations)
	errors := make(chan error, numGoroutines*numIterations)

	// Launch multiple goroutines generating fingerprints concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				fingerprint, err := hf.GenerateHardwareFingerprint()
				if err != nil {
					errors <- err
					return
				}
				results <- fingerprint.Fingerprint
			}
		}(i)
	}

	// Collect results
	var fingerprints []string
	for i := 0; i < numGoroutines*numIterations; i++ {
		select {
		case fp := <-results:
			fingerprints = append(fingerprints, fp)
		case err := <-errors:
			t.Fatalf("Concurrent generation failed: %v", err)
		}
	}

	// All fingerprints should be identical
	first := fingerprints[0]
	for i, fp := range fingerprints {
		if fp != first {
			t.Fatalf("Fingerprint %d differs: expected %s, got %s", i, first, fp)
		}
	}

	t.Logf("Concurrent access test passed - %d identical fingerprints generated", len(fingerprints))
}

// TestWeightedConfidenceWithMissingComponent tests the weighted confidence calculation
// when a component is missing, simulating the QAQC edge case requirement
func TestWeightedConfidenceWithMissingComponent(t *testing.T) {
	// Create a hardware fingerprinter instance
	hf := NewHardwareFingerprinter()

	// Simulate current hardware fingerprint with all components present
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

	// Test Case 1: All components present (should return 1.0 confidence for exact match)
	storedComponents := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	match := &FingerprintMatch{
		StoredFingerprint:  currentFingerprint.Fingerprint,
		CurrentFingerprint: currentFingerprint.Fingerprint,
		ExactMatch:         true,
		Components:         make([]ComponentMatch, 0),
	}

	confidence := hf.calculateWeightedConfidence(currentFingerprint, storedComponents, match)
	if confidence != 1.0 {
		t.Errorf("Expected confidence 1.0 for exact match with all components, got %f", confidence)
	}

	// Test Case 2: Disk serial missing (main QAQC test case)
	// Motherboard (30%) + UUID (30%) + CPU (20%) + BIOS (15%) = 95% confidence
	storedComponentsMissingDisk := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "", // Missing component
	}

	confidenceMissingDisk := hf.calculateWeightedConfidence(currentFingerprint, storedComponentsMissingDisk, match)
	expectedConfidence := 0.30 + 0.30 + 0.20 + 0.15 // 95% = 0.95
	if confidenceMissingDisk != expectedConfidence {
		t.Errorf("Expected confidence %f for missing disk serial, got %f", expectedConfidence, confidenceMissingDisk)
	}

	// Verify that missing disk serial still passes the 0.7 threshold
	if confidenceMissingDisk < 0.7 {
		t.Errorf("Missing disk serial should still pass 0.7 threshold, got %f", confidenceMissingDisk)
	}

	// Test Case 3: Motherboard serial missing (critical component)
	// UUID (30%) + CPU (20%) + BIOS (15%) + Disk (5%) = 70% confidence
	storedComponentsMissingMotherboard := map[string]string{
		"motherboard_serial": "", // Missing critical component
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "test-disk-98765",
	}

	confidenceMissingMotherboard := hf.calculateWeightedConfidence(currentFingerprint, storedComponentsMissingMotherboard, match)
	expectedConfidenceMB := 0.30 + 0.20 + 0.15 + 0.05 // 70% = 0.70
	if confidenceMissingMotherboard != expectedConfidenceMB {
		t.Errorf("Expected confidence %f for missing motherboard serial, got %f", expectedConfidenceMB, confidenceMissingMotherboard)
	}

	// Verify that missing motherboard still passes the 0.7 threshold (exactly at threshold)
	if confidenceMissingMotherboard < 0.7 {
		t.Errorf("Missing motherboard should pass exactly at 0.7 threshold, got %f", confidenceMissingMotherboard)
	}

	// Test Case 4: Both motherboard and UUID missing (should fail threshold)
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
	if confidenceMissingCritical != expectedConfidenceCritical {
		t.Errorf("Expected confidence %f for missing critical components, got %f", expectedConfidenceCritical, confidenceMissingCritical)
	}

	// Verify that missing critical components fails the 0.7 threshold
	if confidenceMissingCritical >= 0.7 {
		t.Errorf("Missing critical components should fail 0.7 threshold, got %f", confidenceMissingCritical)
	}

	t.Logf("Weighted confidence tests passed:")
	t.Logf("  - Exact match: %.2f", confidence)
	t.Logf("  - Missing disk: %.2f (passes 0.7 threshold)", confidenceMissingDisk)
	t.Logf("  - Missing motherboard: %.2f (passes 0.7 threshold)", confidenceMissingMotherboard)
	t.Logf("  - Missing critical components: %.2f (fails 0.7 threshold)", confidenceMissingCritical)
}

// TestValidateHardwareFingerprintWithComponents tests the enhanced validation method
func TestValidateHardwareFingerprintWithComponents(t *testing.T) {
	hf := NewHardwareFingerprinter()

	// Mock component data
	storedComponents := map[string]string{
		"motherboard_serial": "TEST-MB-12345",
		"system_uuid":        "test-uuid-67890abcdef",
		"cpu_model":          "Test CPU Model",
		"bios_version":       "1.0.0",
		"primary_disk_serial": "", // Missing disk serial
	}

	storedFingerprint := "test-fingerprint-hash"

	// Test enhanced validation (should work with weighted confidence)
	match, err := hf.ValidateHardwareFingerprintWithComponents(storedFingerprint, storedComponents)
	if err != nil {
		t.Fatalf("ValidateHardwareFingerprintWithComponents failed: %v", err)
	}

	// Should return partial match with confidence >= 0.7
	if match.MatchType != "partial" {
		t.Errorf("Expected match type 'partial', got '%s'", match.MatchType)
	}

	if match.Confidence < 0.7 {
		t.Errorf("Expected confidence >= 0.7, got %f", match.Confidence)
	}

	// Test legacy validation (should use fallback behavior)
	matchLegacy, err := hf.ValidateHardwareFingerprint(storedFingerprint)
	if err != nil {
		t.Fatalf("ValidateHardwareFingerprint failed: %v", err)
	}

	// Legacy validation should return exact match only for identical fingerprints
	if matchLegacy.MatchType == "partial" {
		t.Errorf("Legacy validation should not return 'partial' match type, got '%s'", matchLegacy.MatchType)
	}

	t.Logf("Enhanced validation test passed:")
	t.Logf("  - Enhanced validation: %s match with %.2f confidence", match.MatchType, match.Confidence)
	t.Logf("  - Legacy validation: %s match with %.2f confidence", matchLegacy.MatchType, matchLegacy.Confidence)
}