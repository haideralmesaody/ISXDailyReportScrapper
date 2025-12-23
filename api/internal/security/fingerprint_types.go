package security

import (
	"time"
)

// FingerprintMatch represents the result of fingerprint comparison
type FingerprintMatch struct {
	StoredFingerprint   string           `json:"stored_fingerprint"`
	CurrentFingerprint  string           `json:"current_fingerprint"`
	MatchType           string           `json:"match_type"` // "exact", "partial", "fuzzy", "no_match"
	Confidence          float64          `json:"confidence"` // 0.0 to 1.0
	ExactMatch          bool             `json:"exact_match"`
	Components          []ComponentMatch `json:"components"`
	GeneratedAt         time.Time        `json:"generated_at"`
	ValidationDuration  time.Duration    `json:"validation_duration"`
}

// ComponentMatch represents individual component comparison
type ComponentMatch struct {
	Name         string `json:"name"`         // "motherboard_serial", "system_uuid", etc.
	StoredValue  string `json:"stored_value"`
	CurrentValue string `json:"current_value"`
	Matches      bool   `json:"matches"`
	Weight       float64 `json:"weight"` // Importance weight for this component
}

// FingerprintSystem represents different fingerprinting systems
type FingerprintSystem int

const (
	// LegacySystem uses the original 5-factor system (MAC, hostname, CPU ID, OS, platform)
	LegacySystem FingerprintSystem = iota

	// HardwareSystem uses stable hardware components only
	HardwareSystem
)

// String returns string representation of fingerprint system
func (fs FingerprintSystem) String() string {
	switch fs {
	case LegacySystem:
		return "legacy"
	case HardwareSystem:
		return "hardware"
	default:
		return "unknown"
	}
}

// HybridFingerprintManager manages both legacy and hardware fingerprint systems
type HybridFingerprintManager struct {
	legacyManager     *FingerprintManager
	hardwareManager   *HardwareFingerprinter
	preferHardware    bool // During migration, prefer hardware system
	migrationMode     bool // Support both systems during transition
}

// NewHybridFingerprintManager creates a manager that supports both fingerprint systems
func NewHybridFingerprintManager(preferHardware bool) *HybridFingerprintManager {
	return &HybridFingerprintManager{
		legacyManager:   NewFingerprintManager(),
		hardwareManager: NewHardwareFingerprinter(),
		preferHardware:  preferHardware,
		migrationMode:   true, // Enable migration mode by default
	}
}

// ValidateFingerprint validates using both systems during migration
func (hfm *HybridFingerprintManager) ValidateFingerprint(storedFingerprint string) (*FingerprintMatch, error) {
	start := time.Now()

	// Try hardware fingerprint first if preferred
	if hfm.preferHardware {
		if match, err := hfm.hardwareManager.ValidateHardwareFingerprint(storedFingerprint); err == nil {
			if match.ExactMatch {
				match.ValidationDuration = time.Since(start)
				match.MatchType = "hardware_exact"
				return match, nil
			}
		}
	}

	// Fall back to legacy system
	if matchBool, err := hfm.legacyManager.ValidateFingerprint(storedFingerprint); err == nil {
		match := &FingerprintMatch{
			StoredFingerprint:   storedFingerprint,
			CurrentFingerprint:  "legacy_fingerprint", // We don't have the current one easily
			MatchType:          "legacy_exact",
			Confidence:         1.0,
			ExactMatch:         matchBool,
			GeneratedAt:        time.Now(),
			ValidationDuration: time.Since(start),
		}
		if !matchBool {
			match.MatchType = "legacy_no_match"
			match.Confidence = 0.0
		}
		return match, nil
	}

	// Try hardware fingerprint if not preferred
	if !hfm.preferHardware {
		if match, err := hfm.hardwareManager.ValidateHardwareFingerprint(storedFingerprint); err == nil {
			match.ValidationDuration = time.Since(start)
			if match.ExactMatch {
				match.MatchType = "hardware_exact"
			} else {
				match.MatchType = "hardware_no_match"
			}
			return match, nil
		}
	}

	// No match found in any system
	return &FingerprintMatch{
		StoredFingerprint:  storedFingerprint,
		MatchType:          "no_match",
		Confidence:         0.0,
		ExactMatch:         false,
		GeneratedAt:        time.Now(),
		ValidationDuration: time.Since(start),
	}, nil
}

// GenerateFingerprint generates fingerprint using preferred system
func (hfm *HybridFingerprintManager) GenerateFingerprint(system FingerprintSystem) (interface{}, error) {
	switch system {
	case HardwareSystem:
		return hfm.hardwareManager.GenerateHardwareFingerprint()
	case LegacySystem:
		return hfm.legacyManager.GenerateFingerprint()
	default:
		if hfm.preferHardware {
			return hfm.hardwareManager.GenerateHardwareFingerprint()
		}
		return hfm.legacyManager.GenerateFingerprint()
	}
}

// GetFingerprintComponents returns components from both systems
func (hfm *HybridFingerprintManager) GetFingerprintComponents() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get legacy components
	if legacyComps, err := hfm.legacyManager.GetFingerprintComponents(); err == nil {
		result["legacy"] = legacyComps
	}

	// Get hardware components
	if hwComps, err := hfm.hardwareManager.GetHardwareComponents(); err == nil {
		result["hardware"] = hwComps
	}

	return result, nil
}

// ClearCache clears caches for both systems
func (hfm *HybridFingerprintManager) ClearCache() {
	hfm.legacyManager.ClearCache()
	hfm.hardwareManager.ClearCache()
}

// SetMigrationMode enables or disables migration mode
func (hfm *HybridFingerprintManager) SetMigrationMode(enabled bool) {
	hfm.migrationMode = enabled
}

// SetPreferredSystem sets which fingerprint system to prefer
func (hfm *HybridFingerprintManager) SetPreferredSystem(system FingerprintSystem) {
	hfm.preferHardware = (system == HardwareSystem)
}