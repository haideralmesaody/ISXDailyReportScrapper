package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"
)

// HardwareFingerprint represents a stable hardware-based device fingerprint
type HardwareFingerprint struct {
	// Core stable hardware components
	MotherboardSerial string `json:"motherboard_serial"`
	SystemUUID        string `json:"system_uuid"`
	PrimaryDiskSerial string `json:"primary_disk_serial"`
	CPUModel          string `json:"cpu_model"`
	BIOSVersion       string `json:"bios_version"`

	// Computed fingerprint
	Fingerprint      string    `json:"fingerprint"`
	GeneratedAt      time.Time `json:"generated_at"`
	Platform         string    `json:"platform"`
	Components       int       `json:"components"` // Number of components successfully gathered
}

// HardwareFingerprinter handles stable hardware fingerprinting
type HardwareFingerprinter struct {
	cache          *HardwareFingerprint
	cacheMutex     sync.RWMutex
	cacheExpiry    time.Time
	cacheDuration  time.Duration
}

// NewHardwareFingerprinter creates a new hardware fingerprinter
func NewHardwareFingerprinter() *HardwareFingerprinter {
	return &HardwareFingerprinter{
		cacheDuration: 24 * time.Hour, // Cache for 24 hours - hardware doesn't change
	}
}

// GenerateHardwareFingerprint creates a fingerprint based on stable hardware components
func (hf *HardwareFingerprinter) GenerateHardwareFingerprint() (*HardwareFingerprint, error) {
	// Check cache first
	hf.cacheMutex.RLock()
	if hf.cache != nil && time.Now().Before(hf.cacheExpiry) {
		cachedFingerprint := *hf.cache
		hf.cacheMutex.RUnlock()

		slog.Debug("Using cached hardware fingerprint",
			slog.String("fingerprint", cachedFingerprint.Fingerprint),
			slog.Time("cached_at", cachedFingerprint.GeneratedAt),
		)
		return &cachedFingerprint, nil
	}
	hf.cacheMutex.RUnlock()

	start := time.Now()
	slog.Debug("Generating new hardware fingerprint")

	var fingerprint HardwareFingerprint
	var successfulComponents []string
	var componentCount int
	var detectedComponents []string

	// Get motherboard serial (most stable - weight: 30%)
	if serial, err := hf.getMotherboardSerial(); err == nil && serial != "" {
		fingerprint.MotherboardSerial = serial
		successfulComponents = append(successfulComponents, "mb:"+serial)
		detectedComponents = append(detectedComponents, fmt.Sprintf("motherboard_serial:%s", serial))
		componentCount++
		slog.Info("Successfully retrieved motherboard serial",
			slog.String("serial", serial),
			slog.Int("component_count", componentCount))
	} else {
		slog.Warn("Failed to get motherboard serial",
			slog.String("error", err.Error()),
			slog.String("component", "motherboard_serial"))
	}

	// Get system UUID (very stable - weight: 30%)
	if uuid, err := hf.getSystemUUID(); err == nil && uuid != "" {
		fingerprint.SystemUUID = uuid
		successfulComponents = append(successfulComponents, "uuid:"+uuid)
		detectedComponents = append(detectedComponents, fmt.Sprintf("system_uuid:%s", uuid))
		componentCount++
		slog.Info("Successfully retrieved system UUID",
			slog.String("uuid", uuid),
			slog.Int("component_count", componentCount))
	} else {
		slog.Warn("Failed to get system UUID",
			slog.String("error", err.Error()),
			slog.String("component", "system_uuid"))
	}

	// Get primary disk serial (very stable - weight: 20%)
	if serial, err := hf.getPrimaryDiskSerial(); err == nil && serial != "" {
		fingerprint.PrimaryDiskSerial = serial
		successfulComponents = append(successfulComponents, "disk:"+serial)
		detectedComponents = append(detectedComponents, fmt.Sprintf("primary_disk_serial:%s", serial))
		componentCount++
		slog.Info("Successfully retrieved primary disk serial",
			slog.String("serial", serial),
			slog.Int("component_count", componentCount))
	} else {
		slog.Warn("Failed to get primary disk serial",
			slog.String("error", err.Error()),
			slog.String("component", "primary_disk_serial"))
	}

	// Get CPU model (stable - weight: 10%)
	if model, err := hf.getCPUModel(); err == nil && model != "" {
		fingerprint.CPUModel = model
		successfulComponents = append(successfulComponents, "cpu:"+model)
		detectedComponents = append(detectedComponents, fmt.Sprintf("cpu_model:%s", model))
		componentCount++
		slog.Info("Successfully retrieved CPU model",
			slog.String("model", model),
			slog.Int("component_count", componentCount))
	} else {
		slog.Warn("Failed to get CPU model",
			slog.String("error", err.Error()),
			slog.String("component", "cpu_model"))
	}

	// Get BIOS version (mostly stable - weight: 10%)
	if version, err := hf.getBIOSVersion(); err == nil && version != "" {
		fingerprint.BIOSVersion = version
		successfulComponents = append(successfulComponents, "bios:"+version)
		detectedComponents = append(detectedComponents, fmt.Sprintf("bios_version:%s", version))
		componentCount++
		slog.Info("Successfully retrieved BIOS version",
			slog.String("version", version),
			slog.Int("component_count", componentCount))
	} else {
		slog.Warn("Failed to get BIOS version",
			slog.String("error", err.Error()),
			slog.String("component", "bios_version"))
	}

	// Require at least 2 stable components for a valid fingerprint
	if componentCount < 2 {
		return nil, fmt.Errorf("insufficient hardware components detected: %d (minimum 2 required). "+
			"Detected components: [%s]. This may indicate permission issues or hardware virtualization. "+
			"Try running as administrator or check if running in a virtual environment",
			componentCount, strings.Join(detectedComponents, ", "))
	}

	// Create fingerprint from successfully detected hardware components only
	combinedData := strings.Join(successfulComponents, "|")
	hash := sha256.Sum256([]byte(combinedData))
	fingerprint.Fingerprint = hex.EncodeToString(hash[:])
	fingerprint.GeneratedAt = time.Now()
	fingerprint.Platform = runtime.GOOS
	fingerprint.Components = componentCount

	// Cache the result
	hf.cacheMutex.Lock()
	hf.cache = &fingerprint
	hf.cacheExpiry = time.Now().Add(hf.cacheDuration)
	hf.cacheMutex.Unlock()

	duration := time.Since(start)
	slog.Info("Hardware fingerprint generated successfully",
		slog.String("fingerprint", fingerprint.Fingerprint),
		slog.String("platform", fingerprint.Platform),
		slog.Int("components", componentCount),
		slog.Duration("generation_time", duration),
		slog.String("motherboard_serial", fingerprint.MotherboardSerial),
		slog.String("system_uuid", fingerprint.SystemUUID),
		slog.String("primary_disk_serial", fingerprint.PrimaryDiskSerial),
		slog.String("cpu_model", fingerprint.CPUModel),
		slog.String("bios_version", fingerprint.BIOSVersion),
	)

	return &fingerprint, nil
}

// ValidateHardwareFingerprintWithComponents compares current hardware with stored component data
// This is the enhanced validation method that uses weighted confidence scoring
func (hf *HardwareFingerprinter) ValidateHardwareFingerprintWithComponents(storedFingerprint string, storedComponents map[string]string) (*FingerprintMatch, error) {
	current, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate current hardware fingerprint: %w", err)
	}

	match := &FingerprintMatch{
		StoredFingerprint:   storedFingerprint,
		CurrentFingerprint:  current.Fingerprint,
		ExactMatch:          current.Fingerprint == storedFingerprint,
		GeneratedAt:         current.GeneratedAt,
		Components:          make([]ComponentMatch, 0),
	}

	// Calculate weighted confidence based on component comparison
	match.Confidence = hf.calculateWeightedConfidence(current, storedComponents, match)

	if match.ExactMatch {
		match.MatchType = "exact"
	} else if match.Confidence >= 0.7 {
		match.MatchType = "partial"
	} else {
		match.MatchType = "no_match"
	}

	slog.Info("Enhanced hardware fingerprint validation completed",
		slog.String("stored", storedFingerprint),
		slog.String("current", current.Fingerprint),
		slog.String("match_type", match.MatchType),
		slog.Float64("confidence", match.Confidence),
		slog.Int("components", current.Components),
	)

	return match, nil
}

// ValidateHardwareFingerprint compares current hardware fingerprint with stored one (legacy method)
func (hf *HardwareFingerprinter) ValidateHardwareFingerprint(storedFingerprint string) (*FingerprintMatch, error) {
	current, err := hf.GenerateHardwareFingerprint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate current hardware fingerprint: %w", err)
	}

	match := &FingerprintMatch{
		StoredFingerprint:   storedFingerprint,
		CurrentFingerprint:  current.Fingerprint,
		ExactMatch:          current.Fingerprint == storedFingerprint,
		GeneratedAt:         current.GeneratedAt,
		Components:          make([]ComponentMatch, 0),
	}

	// For legacy validation without component data, fall back to exact match only
	if match.ExactMatch {
		match.Confidence = 1.0
		match.MatchType = "exact"
	} else {
		match.Confidence = 0.0
		match.MatchType = "no_match"
	}

	slog.Info("Hardware fingerprint validation completed",
		slog.String("stored", storedFingerprint),
		slog.String("current", current.Fingerprint),
		slog.String("match_type", match.MatchType),
		slog.Float64("confidence", match.Confidence),
		slog.Int("components", current.Components),
	)

	return match, nil
}

// GetHardwareComponents returns individual hardware components for debugging
func (hf *HardwareFingerprinter) GetHardwareComponents() (map[string]string, error) {
	components := make(map[string]string)

	if serial, err := hf.getMotherboardSerial(); err == nil {
		components["motherboard_serial"] = serial
	}

	if uuid, err := hf.getSystemUUID(); err == nil {
		components["system_uuid"] = uuid
	}

	if serial, err := hf.getPrimaryDiskSerial(); err == nil {
		components["primary_disk_serial"] = serial
	}

	if model, err := hf.getCPUModel(); err == nil {
		components["cpu_model"] = model
	}

	if version, err := hf.getBIOSVersion(); err == nil {
		components["bios_version"] = version
	}

	return components, nil
}

// ClearCache clears the cached hardware fingerprint
func (hf *HardwareFingerprinter) ClearCache() {
	hf.cacheMutex.Lock()
	defer hf.cacheMutex.Unlock()

	hf.cache = nil
	hf.cacheExpiry = time.Time{}

	slog.Debug("Hardware fingerprint cache cleared")
}

// calculateWeightedConfidence calculates confidence based on component matching and weights
func (hf *HardwareFingerprinter) calculateWeightedConfidence(current *HardwareFingerprint, storedComponents map[string]string, match *FingerprintMatch) float64 {
	// Define component weights based on stability and importance
	componentWeights := map[string]float64{
		"motherboard_serial": 0.30, // Most stable hardware identifier
		"system_uuid":        0.30, // Most stable system identifier
		"cpu_model":          0.20, // Stable but can change with upgrades
		"bios_version":       0.15, // Can change with updates but relatively stable
		"primary_disk_serial": 0.05, // Least stable - can change with disk replacements
	}

	// Extract stored components from the provided component map
	storedComponents = hf.extractStoredComponents(storedComponents)

	totalConfidence := 0.0
	totalWeight := 0.0

	// Compare each component and calculate weighted confidence
	components := []struct {
		name     string
		current  string
		stored   string
		weight   float64
	}{
		{"motherboard_serial", current.MotherboardSerial, storedComponents["motherboard_serial"], componentWeights["motherboard_serial"]},
		{"system_uuid", current.SystemUUID, storedComponents["system_uuid"], componentWeights["system_uuid"]},
		{"cpu_model", current.CPUModel, storedComponents["cpu_model"], componentWeights["cpu_model"]},
		{"bios_version", current.BIOSVersion, storedComponents["bios_version"], componentWeights["bios_version"]},
		{"primary_disk_serial", current.PrimaryDiskSerial, storedComponents["primary_disk_serial"], componentWeights["primary_disk_serial"]},
	}

	for _, comp := range components {
		componentMatch := ComponentMatch{
			Name:         comp.name,
			StoredValue:  comp.stored,
			CurrentValue: comp.current,
			Weight:       comp.weight,
		}

		// Determine if component matches
		if comp.current != "" && comp.stored != "" {
			componentMatch.Matches = comp.current == comp.stored
			if componentMatch.Matches {
				totalConfidence += comp.weight
			}
		} else if comp.current == "" && comp.stored == "" {
			// Both empty - consider as matched with lower confidence
			componentMatch.Matches = true
			totalConfidence += comp.weight * 0.5 // Reduced confidence for missing components
		} else {
			// One is empty, one is not - no match
			componentMatch.Matches = false
		}

		totalWeight += comp.weight
		match.Components = append(match.Components, componentMatch)
	}

	// Normalize confidence
	if totalWeight > 0 {
		return totalConfidence / totalWeight
	}

	return 0.0
}

// extractStoredComponents extracts components from stored component data
// For backward compatibility, handles both component data and legacy fingerprint-only storage
func (hf *HardwareFingerprinter) extractStoredComponents(storedComponents map[string]string) map[string]string {
	if storedComponents == nil {
		return make(map[string]string)
	}

	// Return a copy to prevent modification of the original map
	result := make(map[string]string)
	for key, value := range storedComponents {
		result[key] = value
	}

	return result
}

// Platform-specific implementation methods will be added in separate files
// to keep this file clean and maintainable