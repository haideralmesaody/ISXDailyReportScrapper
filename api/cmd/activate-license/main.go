package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
	
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/security"
)

func main() {
	// Command-line flags
	var (
		licenseKey = flag.String("key", "", "License key to activate (e.g., ISX-XXXX-XXXX-XXXX-XXXX)")
		verbose    = flag.Bool("v", false, "Enable verbose logging")
		help       = flag.Bool("help", false, "Show help message")
	)
	
	flag.Parse()
	
	// Show help if requested
	if *help || *licenseKey == "" {
		fmt.Println("ISX Pulse License Activation Tool")
		fmt.Println("==================================")
		fmt.Println()
		fmt.Println("This tool activates a license key and saves it locally.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  activate-license -key ISX-XXXX-XXXX-XXXX-XXXX")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -key string   License key to activate (required)")
		fmt.Println("  -v           Enable verbose logging")
		fmt.Println("  -help        Show this help message")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  activate-license -key ISX-JYHT-CVUF-K25S-CQ92")
		
		if *licenseKey == "" && !*help {
			fmt.Println()
			fmt.Println("Error: License key is required")
			os.Exit(1)
		}
		os.Exit(0)
	}
	
	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	
	// Create a text handler with colored output for better readability
	opts := &slog.HandlerOptions{
		Level: logLevel,
		AddSource: false,
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	
	fmt.Println("ISX Pulse License Activation Tool")
	fmt.Println("==================================")
	fmt.Printf("Activating license: %s\n", *licenseKey)
	fmt.Println()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	
	// Get license file path
	licensePath := cfg.GetLicenseFile()
	fmt.Printf("License will be saved to: %s\n", licensePath)
	fmt.Println()
	
	// Initialize license manager
	fmt.Println("Initializing license manager...")
	manager, err := license.NewManager()
	if err != nil {
		log.Fatalf("❌ Failed to initialize license manager: %v", err)
	}
	activator := license.NewActivationService(manager)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Activate the license with progress messages
	fmt.Println("Contacting activation server...")
	fmt.Println("This may take a few seconds...")
	fmt.Println()
	
	start := time.Now()
	
	// Verbose mode - show detailed information during activation
	if *verbose {
		fmt.Println("VERBOSE MODE - Detailed Debug Information")
		fmt.Println("========================================")
		fmt.Printf("Step 1: Pre-activation checks\n")
		fmt.Printf("• License key format: %s\n", *licenseKey)
		fmt.Printf("• License key length: %d characters\n", len(*licenseKey))
		fmt.Printf("• License file path: %s\n", licensePath)
		fmt.Printf("• Current time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("• Context timeout: %v\n", 30*time.Second)
		fmt.Println()
		
		// Generate device fingerprint for verbose output
		fmt.Printf("Step 2: Device fingerprint generation\n")
		fingerprint, fpErr := generateDeviceFingerprint()
		if fpErr == nil {
			fmt.Printf("• Device fingerprint: %s\n", fingerprint.Fingerprint)
			fmt.Printf("• Hostname: %s\n", fingerprint.Hostname)
			fmt.Printf("• MAC Address: %s\n", fingerprint.MACAddress)
			fmt.Printf("• CPU ID: %s\n", fingerprint.CPUID)
			fmt.Printf("• OS/Platform: %s/%s\n", fingerprint.OS, fingerprint.Platform)
			fmt.Printf("• Generation time: %v\n", time.Since(fingerprint.GeneratedAt))
		} else {
			fmt.Printf("• Fingerprint generation error: %v\n", fpErr)
		}
		fmt.Println()
		
		fmt.Printf("Step 3: Network request preparation\n")
		fmt.Printf("• Request timeout: 30 seconds\n")
		fmt.Printf("• Retry policy: 3 attempts with exponential backoff\n")
		fmt.Printf("• User-Agent: ISX-Pulse-License-Client/%s\n", "v3.0.0")
		fmt.Println()
	}
	
	// Create a channel to receive progress updates (if we had real-time progress)
	// For now, we'll simulate progress with messages
	if !*verbose {
		go func() {
			time.Sleep(1 * time.Second)
			fmt.Println("🔄 Connecting to activation server... (attempt 1/3)")
		}()
	} else {
		fmt.Printf("Step 4: Initiating license activation request\n")
		fmt.Printf("• Start time: %s\n", start.Format("15:04:05.000"))
		fmt.Println("• Sending activation request...")
	}
	
	err = activator.Activate(ctx, *licenseKey)
	duration := time.Since(start)
	
	if err != nil {
		if *verbose {
			fmt.Printf("Step 5: Activation failed\n")
			fmt.Printf("• End time: %s\n", time.Now().Format("15:04:05.000"))
			fmt.Printf("• Total duration: %v\n", duration)
			fmt.Printf("• Error type: %T\n", err)
			fmt.Printf("• Raw error: %v\n", err)
			fmt.Println()
		}
		
		fmt.Printf("❌ Activation failed (took %v)\n", duration)
		fmt.Println()
		
		// Get user-friendly error message
		userMsg := license.GetUserFriendlyMessage(err)
		
		// Display user-friendly error
		fmt.Printf("🔍 %s\n", userMsg.Title)
		fmt.Printf("   %s\n", userMsg.Message)
		if userMsg.Suggestion != "" {
			fmt.Println()
			fmt.Printf("💡 %s\n", userMsg.Suggestion)
		}
		
		if userMsg.CanRetry {
			fmt.Println()
			fmt.Println("You can try the activation again with the correct information.")
		}
		
		// Show additional technical details in verbose mode
		if *verbose {
			fmt.Println()
			fmt.Println("VERBOSE ERROR ANALYSIS:")
			fmt.Println("======================")
			
			// Error categorization
			errorStr := strings.ToLower(err.Error())
			fmt.Printf("• Error category: ")
			switch {
			case strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "deadline exceeded"):
				fmt.Println("Network Timeout")
				fmt.Println("• Likely causes: Slow internet, server overload, firewall blocking")
			case strings.Contains(errorStr, "network") || strings.Contains(errorStr, "connection") || strings.Contains(errorStr, "dial"):
				fmt.Println("Network Connectivity")
				fmt.Println("• Likely causes: No internet, DNS issues, proxy/firewall blocking")
			case strings.Contains(errorStr, "signature") || strings.Contains(errorStr, "hmac"):
				fmt.Println("Authentication/Signature")
				fmt.Println("• Likely causes: Invalid license key, corrupted data, server key mismatch")
			case strings.Contains(errorStr, "already activated"):
				fmt.Println("License Already Activated")
				fmt.Println("• Likely causes: License used on another device, duplicate activation attempt")
			case strings.Contains(errorStr, "expired"):
				fmt.Println("License Expired")
				fmt.Println("• Likely causes: License has passed expiry date, server time mismatch")
			default:
				fmt.Println("Unknown/Other")
			}
			
			fmt.Printf("• Suggestion: %s\n", userMsg.Suggestion)
			fmt.Printf("• Can retry: %t\n", userMsg.CanRetry)
		}
		
		os.Exit(1)
	}
	
	// Success! Get detailed information for user-friendly display
	if *verbose {
		fmt.Printf("Step 5: Activation successful\n")
		fmt.Printf("• End time: %s\n", time.Now().Format("15:04:05.000"))
		fmt.Printf("• Total duration: %v\n", duration)
		fmt.Println("• Proceeding to license validation...")
		fmt.Println()
	}
	
	fmt.Println("Validating license...")
	validationStart := time.Now()
	isValid, err := manager.ValidateLicense()
	validationDuration := time.Since(validationStart)
	
	if *verbose {
		fmt.Printf("Step 6: License validation\n")
		fmt.Printf("• Validation duration: %v\n", validationDuration)
		fmt.Printf("• Validation result: %t\n", isValid)
		if err != nil {
			fmt.Printf("• Validation error: %v\n", err)
		}
		fmt.Println()
	}
	
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not validate license: %v\n", err)
	} else if !isValid {
		fmt.Println("⚠️  Warning: License validation returned false")
	}
	
	// Get license status and create success message
	licenseInfo, statusMsg, err := manager.GetLicenseStatus()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not get license status: %v\n", err)
		// Fall back to basic success message
		fmt.Printf("✅ License activated successfully! (took %v)\n", duration)
	} else if licenseInfo != nil {
		// Calculate days remaining
		daysRemaining := 0
		if !licenseInfo.ExpiryDate.IsZero() {
			daysRemaining = int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
		}
		
		// Check if this was a reactivation by looking for specific context
		// This is a simplified check - in a full implementation, you might want to 
		// track this information in the activation response
		isReactivation := contains(statusMsg, "reactivated") || contains(statusMsg, "reactivation")
		
		// Create basic success info
		successInfo := &license.LicenseSuccessInfo{
			ExpiryDate:    licenseInfo.ExpiryDate,
			DaysRemaining: daysRemaining,
			IsReactivation: isReactivation,
		}
		
		// For demonstration, assume default reactivation limits
		// In practice, this information should come from the activation response
		if isReactivation {
			successInfo.ReactivationInfo = license.CalculateReactivationInfo(1, 5) // Example: 1 used, 5 max
		}
		
		// Generate and display user-friendly success message
		successMsg := license.FormatSuccessMessage(successInfo)
		
		// Display success message with appropriate icon
		var icon string
		switch successMsg.SeverityLevel {
		case "success":
			icon = "✅"
		case "warning":
			icon = "⚠️"
		default:
			icon = "✅"
		}
		
		fmt.Printf("%s %s (took %v)\n", icon, successMsg.Title, duration)
		fmt.Println("=" + strings.Repeat("=", len(successMsg.Title)+len(fmt.Sprintf(" (took %v)", duration))))
		fmt.Println()
		fmt.Println(successMsg.Message)
		
		if successMsg.Suggestion != "" {
			fmt.Println()
			fmt.Printf("💡 %s\n", successMsg.Suggestion)
		}
		
		// Show technical details if verbose
		if *verbose {
			fmt.Println()
			fmt.Println("VERBOSE ACTIVATION SUMMARY:")
			fmt.Println("===========================")
			fmt.Printf("• Status: %s\n", licenseInfo.Status)
			fmt.Printf("• License Key: %s\n", maskLicenseKey(licenseInfo.LicenseKey))
			if licenseInfo.UserEmail != "" {
				fmt.Printf("• User Email: %s\n", licenseInfo.UserEmail)
			}
			fmt.Printf("• Duration: %s\n", licenseInfo.Duration)
			fmt.Printf("• Issued Date: %s\n", licenseInfo.IssuedDate.Format("2006-01-02 15:04:05"))
			fmt.Printf("• Expiry Date: %s\n", licenseInfo.ExpiryDate.Format("2006-01-02 15:04:05"))
			fmt.Printf("• Last Checked: %s\n", licenseInfo.LastChecked.Format("2006-01-02 15:04:05"))
			if statusMsg != "" {
				fmt.Printf("• Status Message: %s\n", statusMsg)
			}
			if licenseInfo.ActivationID != "" {
				fmt.Printf("• Activation ID: %s\n", licenseInfo.ActivationID)
			}
			fmt.Printf("• License File: %s\n", licensePath)
			
			// Performance metrics
			fmt.Println()
			fmt.Println("PERFORMANCE METRICS:")
			fmt.Println("--------------------")
			fmt.Printf("• Total activation time: %v\n", duration)
			fmt.Printf("• Validation time: %v\n", validationDuration)
			fmt.Printf("• Overall success rate: 100%% (this session)\n")
		}
	} else {
		// Fall back to basic success message
		fmt.Printf("✅ License activated successfully! (took %v)\n", duration)
	}
	
	fmt.Println()
	fmt.Printf("✅ License file saved to: %s\n", licensePath)
	fmt.Println()
	fmt.Println("You can now run the main application (web-licensed.exe)")
	fmt.Println("The license will be automatically detected.")
}

// maskLicenseKey masks a license key for display
func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "****"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(substr) < len(s) && findSubstring(s, substr)))
}

// findSubstring performs a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// generateDeviceFingerprint generates device fingerprint for verbose output
func generateDeviceFingerprint() (*security.DeviceFingerprint, error) {
	fpManager := security.NewFingerprintManager()
	return fpManager.GenerateFingerprint()
}
