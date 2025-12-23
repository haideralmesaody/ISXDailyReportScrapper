package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/security"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== ISX Pulse Fingerprint Analysis Tool ===")
	fmt.Printf("Platform: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	// Get paths
	paths, err := config.GetPaths()
	if err != nil {
		fmt.Printf("Error: Failed to get paths: %v\n", err)
		os.Exit(1)
	}

	// Check for license file
	licenseFile := paths.LicenseFile
	fmt.Printf("License file: %s\n", licenseFile)
	if _, err := os.Stat(licenseFile); err == nil {
		fmt.Println("✓ License file found")
		analyzeExistingLicense(licenseFile)
	} else {
		fmt.Println("✗ No license file found")
	}

	fmt.Println("\n=== Fingerprint Analysis ===")

	// Initialize fingerprint managers
	legacyManager := security.NewFingerprintManager()
	hardwareManager := security.NewHardwareFingerprinter()
	hybridManager := security.NewHybridFingerprintManager(false) // Don't prefer hardware for analysis

	// Generate legacy fingerprint
	fmt.Println("\n--- Legacy Fingerprint System ---")
	legacyFingerprint, err := legacyManager.GenerateFingerprint()
	if err != nil {
		fmt.Printf("✗ Failed to generate legacy fingerprint: %v\n", err)
	} else {
		fmt.Printf("✓ Legacy fingerprint: %s\n", legacyFingerprint.Fingerprint)
		fmt.Printf("  Hostname: %s\n", legacyFingerprint.Hostname)
		fmt.Printf("  MAC Address: %s\n", legacyFingerprint.MACAddress)
		fmt.Printf("  CPU ID: %s\n", legacyFingerprint.CPUID)
		fmt.Printf("  OS: %s\n", legacyFingerprint.OS)
		fmt.Printf("  Platform: %s\n", legacyFingerprint.Platform)
		fmt.Printf("  Generated: %s\n", legacyFingerprint.GeneratedAt.Format(time.RFC3339))
	}

	// Generate hardware fingerprint
	fmt.Println("\n--- Hardware Fingerprint System ---")
	hardwareFingerprint, err := hardwareManager.GenerateHardwareFingerprint()
	if err != nil {
		fmt.Printf("✗ Failed to generate hardware fingerprint: %v\n", err)
	} else {
		fmt.Printf("✓ Hardware fingerprint: %s\n", hardwareFingerprint.Fingerprint)
		fmt.Printf("  Motherboard Serial: %s\n", hardwareFingerprint.MotherboardSerial)
		fmt.Printf("  System UUID: %s\n", hardwareFingerprint.SystemUUID)
		fmt.Printf("  Primary Disk Serial: %s\n", hardwareFingerprint.PrimaryDiskSerial)
		fmt.Printf("  CPU Model: %s\n", hardwareFingerprint.CPUModel)
		fmt.Printf("  BIOS Version: %s\n", hardwareFingerprint.BIOSVersion)
		fmt.Printf("  Platform: %s\n", hardwareFingerprint.Platform)
		fmt.Printf("  Components: %d\n", hardwareFingerprint.Components)
		fmt.Printf("  Generated: %s\n", hardwareFingerprint.GeneratedAt.Format(time.RFC3339))
	}

	// Get detailed components for debugging
	fmt.Println("\n--- Detailed Component Analysis ---")

	// Legacy components
	if legacyComps, err := legacyManager.GetFingerprintComponents(); err == nil {
		fmt.Println("\nLegacy System Components:")
		for key, value := range legacyComps {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Hardware components
	if hwComps, err := hardwareManager.GetHardwareComponents(); err == nil {
		fmt.Println("\nHardware System Components:")
		for key, value := range hwComps {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Hybrid system comparison
	fmt.Println("\n--- Hybrid System Analysis ---")
	if hybridComps, err := hybridManager.GetFingerprintComponents(); err == nil {
		fmt.Println("Hybrid System Components:")
		componentsJSON, _ := json.MarshalIndent(hybridComps, "  ", "  ")
		fmt.Printf("  %s\n", string(componentsJSON))
	}

	// Compare fingerprints if both are available
	if legacyFingerprint != nil && hardwareFingerprint != nil {
		fmt.Println("\n--- Fingerprint Comparison ---")
		if legacyFingerprint.Fingerprint == hardwareFingerprint.Fingerprint {
			fmt.Println("✓ Both systems generated the same fingerprint")
		} else {
			fmt.Println("✗ Different fingerprints generated")
			fmt.Printf("  Legacy:  %s\n", legacyFingerprint.Fingerprint)
			fmt.Printf("  Hardware: %s\n", hardwareFingerprint.Fingerprint)
			fmt.Println("\nThis is normal and expected. The hardware system is more stable.")
		}
	}

	// Test validation scenarios
	if legacyFingerprint != nil && hardwareFingerprint != nil {
		fmt.Println("\n--- Validation Test Scenarios ---")

		// Test legacy validation with hardware fingerprint
		legacyMatch, err := legacyManager.ValidateFingerprint(hardwareFingerprint.Fingerprint)
		if err != nil {
			fmt.Printf("✗ Legacy validation with hardware fingerprint failed: %v\n", err)
		} else {
			fmt.Printf("Legacy validation with hardware fingerprint: %t\n", legacyMatch)
		}

		// Test hardware validation with legacy fingerprint
		hwMatch, err := hardwareManager.ValidateHardwareFingerprint(legacyFingerprint.Fingerprint)
		if err != nil {
			fmt.Printf("✗ Hardware validation with legacy fingerprint failed: %v\n", err)
		} else {
			fmt.Printf("Hardware validation with legacy fingerprint: %t (confidence: %.2f)\n", hwMatch.ExactMatch, hwMatch.Confidence)
		}

		// Test hybrid validation
		hybridMatch, err := hybridManager.ValidateFingerprint(legacyFingerprint.Fingerprint)
		if err != nil {
			fmt.Printf("✗ Hybrid validation failed: %v\n", err)
		} else {
			fmt.Printf("Hybrid validation: %s (confidence: %.2f)\n", hybridMatch.MatchType, hybridMatch.Confidence)
		}
	}

	fmt.Println("\n=== Recommendations ===")

	if hardwareFingerprint != nil && hardwareFingerprint.Components >= 3 {
		fmt.Println("✓ Hardware fingerprinting is working well")
		fmt.Println("  - All major hardware components detected")
		fmt.Println("  - System is ready for hardware-based licensing")
	} else if hardwareFingerprint != nil {
		fmt.Println("⚠ Hardware fingerprinting partially working")
		fmt.Printf("  - Only %d/%d components detected\n", hardwareFingerprint.Components, 5)
		fmt.Println("  - Consider checking system permissions")
	} else {
		fmt.Println("✗ Hardware fingerprinting failed")
		fmt.Println("  - System will fall back to legacy fingerprinting")
		fmt.Println("  - Consider running as administrator/sudo")
	}

	if legacyFingerprint != nil {
		fmt.Println("✓ Legacy fingerprinting is available as fallback")
		fmt.Println("  - System supports dual-mode operation")
	}

	fmt.Println("\n=== Analysis Complete ===")
}

// analyzeExistingLicense analyzes an existing license file
func analyzeExistingLicense(licenseFile string) {
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		fmt.Printf("✗ Failed to read license file: %v\n", err)
		return
	}

	// Try to parse as JSON
	var licenseData map[string]interface{}
	if err := json.Unmarshal(data, &licenseData); err != nil {
		fmt.Printf("✗ Failed to parse license file as JSON: %v\n", err)
		return
	}

	fmt.Println("License file contents:")
	if fingerprint, ok := licenseData["device_fingerprint"].(string); ok {
		fmt.Printf("  Stored fingerprint: %s\n", fingerprint)

		// Test validation with current systems
		legacyManager := security.NewFingerprintManager()
		hardwareManager := security.NewHardwareFingerprinter()

		// Test legacy validation
		if match, err := legacyManager.ValidateFingerprint(fingerprint); err == nil {
			fmt.Printf("  Legacy validation: %t\n", match)
		}

		// Test hardware validation
		if match, err := hardwareManager.ValidateHardwareFingerprint(fingerprint); err == nil {
			fmt.Printf("  Hardware validation: %t (confidence: %.2f)\n", match.ExactMatch, match.Confidence)
		}
	} else {
		fmt.Println("  No device fingerprint found in license")
	}

	if expiry, ok := licenseData["expiry_date"].(string); ok {
		fmt.Printf("  Expiry date: %s\n", expiry)
	}

	if status, ok := licenseData["status"].(string); ok {
		fmt.Printf("  Status: %s\n", status)
	}
}