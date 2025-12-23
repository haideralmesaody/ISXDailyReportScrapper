package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/app"
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/security"
)

// Embedded Next.js frontend files (all built assets)
//go:embed frontend/**
var frontendFiles embed.FS

func main() {
	// Parse command-line flags
	licenseInfo := flag.Bool("license-info", false, "Display license information")
	checkLicense := flag.Bool("check-license", false, "Check license validity (returns exit codes)")
	reactivationInfo := flag.Bool("reactivation-info", false, "Display reactivation information and limits")
	testConnection := flag.Bool("test-connection", false, "Test connectivity to activation server")
	supportInfo := flag.Bool("support-info", false, "Generate support package for troubleshooting")
	daysRemaining := flag.Bool("days-remaining", false, "Return number of days until license expires (for scripting)")
	help := flag.Bool("help", false, "Show help information")
	flag.Parse()

	// Handle help flag
	if *help {
		showHelp()
		return
	}

	// Handle license commands
	if *licenseInfo {
		displayLicenseInfo()
		return
	}

	if *checkLicense {
		checkLicenseValidity()
		return
	}

	if *reactivationInfo {
		displayReactivationInfo()
		return
	}

	if *testConnection {
		testServerConnection()
		return
	}

	if *supportInfo {
		generateSupportInfo()
		return
	}

	if *daysRemaining {
		getDaysRemaining()
		return
	}

	// Normal application startup
	startApplication()
}

// startApplication handles the normal application startup flow
func startApplication() {
	// Create frontend filesystem from embedded files
	var frontendFS fs.FS
	if frontendSubFS, err := fs.Sub(frontendFiles, "frontend"); err == nil {
		frontendFS = frontendSubFS
		slog.Info("Frontend embedded successfully")
	} else {
		slog.Info("Warning: Frontend embedding failed", slog.String("error", err.Error()))
		frontendFS = nil
	}

	// Create application instance
	application, err := app.NewApplication(frontendFS)
	if err != nil {
		slog.Error("Failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Start application
	if err := application.Run(); err != nil {
		slog.Error("Application error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// showHelp displays help information for command-line usage
func showHelp() {
	fmt.Printf("%s - %s\n\n", app.Executable, app.AppName)
	fmt.Printf("Version: %s\n", app.VERSION)
	fmt.Printf("Build Time: %s\n\n", app.BuildTime)
	fmt.Println("Usage:")
	fmt.Printf("  %s [options]\n\n", app.Executable)
	fmt.Println("Options:")
	fmt.Println("  --license-info       Display detailed license information")
	fmt.Println("  --check-license      Check license validity (exit codes: 0=valid, 1=expired, 2=not found, 3=invalid)")
	fmt.Println("  --reactivation-info  Display reactivation count, limits, and reset dates")
	fmt.Println("  --test-connection    Test connectivity to the activation server")
	fmt.Println("  --support-info       Generate support package for troubleshooting")
	fmt.Println("  --days-remaining     Return number of days until license expires (for scripting)")
	fmt.Println("  --help              Show this help message")
	fmt.Println("\nExit Codes for --check-license:")
	fmt.Println("  0 - License is valid and active")
	fmt.Println("  1 - License is expired")
	fmt.Println("  2 - License not found or not activated")
	fmt.Println("  3 - License is invalid or corrupted")
	fmt.Println("\nIf no options are provided, the application starts normally.")
	fmt.Printf("\nFor more information, visit: %s\n", app.REPO_URL)
}

// displayLicenseInfo shows detailed license information
func displayLicenseInfo() {
	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		fmt.Printf("Error initializing license manager: %v\n", err)
		os.Exit(3)
	}
	
	// Get license status
	licenseInfo, status, err := licenseManager.GetLicenseStatus()
	if err != nil {
		fmt.Printf("Error getting license status: %v\n", err)
		os.Exit(3)
	}
	
	fmt.Printf("%s - License Information\n", app.AppName)
	fmt.Printf("Version: %s\n", app.VERSION)
	fmt.Printf("Build Time: %s\n", app.BuildTime)
	fmt.Println(strings.Repeat("=", 50))
	
	if licenseInfo == nil {
		fmt.Printf("License Status: %s\n", status)
		fmt.Println("License Key: Not activated")
		fmt.Println("Expiry Date: N/A")
		fmt.Println("Days Remaining: N/A")
		fmt.Println("Reactivation Count: N/A")
		fmt.Println("Reactivation Limit: N/A")
	} else {
		// Calculate days remaining
		daysRemaining := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		
		// Mask license key (show first 4 and last 4 characters)
		maskedKey := maskLicenseKey(licenseInfo.LicenseKey)
		
		fmt.Printf("License Status: %s\n", status)
		fmt.Printf("License Key: %s\n", maskedKey)
		fmt.Printf("User Email: %s\n", licenseInfo.UserEmail)
		fmt.Printf("Duration: %s\n", licenseInfo.Duration)
		fmt.Printf("Issued Date: %s\n", licenseInfo.IssuedDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("Expiry Date: %s\n", licenseInfo.ExpiryDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("Days Remaining: %d\n", daysRemaining)
		fmt.Printf("Last Checked: %s\n", licenseInfo.LastChecked.Format("2006-01-02 15:04:05"))
		
		if licenseInfo.ActivationID != "" {
			fmt.Printf("Activation ID: %s\n", licenseInfo.ActivationID)
		}
		
		// Check renewal information (available fields only)
		if renewalInfo, err := licenseManager.CheckRenewalStatus(); err == nil && renewalInfo != nil {
			fmt.Printf("Renewal Status: %s\n", renewalInfo.Status)
			if renewalInfo.NeedsRenewal {
				fmt.Printf("Needs Renewal: Yes\n")
			} else {
				fmt.Printf("Needs Renewal: No\n")
			}
			if renewalInfo.Message != "" {
				fmt.Printf("Renewal Message: %s\n", renewalInfo.Message)
			}
		} else {
			fmt.Println("Renewal Status: N/A")
		}
	}
	
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Check performed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

// checkLicenseValidity checks license and returns appropriate exit codes
func checkLicenseValidity() {
	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		fmt.Printf("License manager error: %v\n", err)
		os.Exit(3) // Invalid/corrupted
	}
	
	// Get license status
	licenseInfo, status, err := licenseManager.GetLicenseStatus()
	if err != nil {
		fmt.Printf("License check error: %v\n", err)
		os.Exit(3) // Invalid/corrupted
	}
	
	// Check if license exists
	if licenseInfo == nil {
		fmt.Printf("License not found or not activated\n")
		os.Exit(2) // Not found
	}
	
	// Check license validity based on status
	statusLower := strings.ToLower(status)
	switch {
	case strings.Contains(statusLower, "active"):
		// Check if it's actually expired by date
		if time.Now().After(licenseInfo.ExpiryDate) {
			fmt.Printf("License expired on %s\n", licenseInfo.ExpiryDate.Format("2006-01-02"))
			os.Exit(1) // Expired
		}
		fmt.Printf("License valid until %s\n", licenseInfo.ExpiryDate.Format("2006-01-02"))
		os.Exit(0) // Valid
		
	case strings.Contains(statusLower, "warning"):
		// Warning status means active but expiring soon
		if time.Now().After(licenseInfo.ExpiryDate) {
			fmt.Printf("License expired on %s\n", licenseInfo.ExpiryDate.Format("2006-01-02"))
			os.Exit(1) // Expired
		}
		fmt.Printf("License valid until %s (expiring soon)\n", licenseInfo.ExpiryDate.Format("2006-01-02"))
		os.Exit(0) // Valid but warning
		
	case strings.Contains(statusLower, "expired"):
		fmt.Printf("License expired on %s\n", licenseInfo.ExpiryDate.Format("2006-01-02"))
		os.Exit(1) // Expired
		
	case strings.Contains(statusLower, "not activated"):
		fmt.Printf("License not activated\n")
		os.Exit(2) // Not found
		
	default:
		fmt.Printf("License status unknown or invalid: %s\n", status)
		os.Exit(3) // Invalid/corrupted
	}
}

// maskLicenseKey masks a license key for display (ISX-XXXX-****-****-****)
func maskLicenseKey(key string) string {
	if len(key) < 8 {
		return "****-****-****-****-****"
	}
	
	// For ISX scratch card format: ISX-XXXX-XXXX-XXXX
	if strings.HasPrefix(key, "ISX-") && len(key) == 17 {
		return key[:8] + "****-****"
	}
	
	// For longer keys: show first 4 and last 4
	if len(key) > 8 {
		return key[:4] + "-****-****-****-" + key[len(key)-4:]
	}
	
	return "****-****-****-****-****"
}

// displayReactivationInfo shows detailed reactivation information
func displayReactivationInfo() {
	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		fmt.Printf("Error initializing license manager: %v\n", err)
		os.Exit(3)
	}
	
	// Get license status and reactivation info
	licenseInfo, status, err := licenseManager.GetLicenseStatus()
	if err != nil {
		fmt.Printf("Error getting license status: %v\n", err)
		os.Exit(3)
	}
	
	fmt.Printf("%s - Reactivation Information\n", app.AppName)
	fmt.Printf("Version: %s\n", app.VERSION)
	fmt.Printf("Build Time: %s\n", app.BuildTime)
	fmt.Println(strings.Repeat("=", 55))
	
	if licenseInfo == nil {
		fmt.Println("License Status: Not activated")
		fmt.Println("Reactivation Count: N/A")
		fmt.Println("Reactivation Limit: N/A")
		fmt.Println("Remaining Reactivations: N/A")
		fmt.Println("Reactivation Reset Date: N/A")
		fmt.Println("\n⚠️  No license has been activated on this system.")
		fmt.Println("💡 Please activate a license first using the activate-license tool.")
	} else {
		// Get reactivation information from the license manager
		reactivationInfo, err := licenseManager.GetReactivationInfo()
		if err != nil {
			// Fall back to default values if we can't get reactivation info
			reactivationInfo = license.CalculateReactivationInfo(0, 5)
		}
		
		// Extract values for display
		reactivationCount := reactivationInfo.Count
		maxReactivations := reactivationInfo.Limit
		remaining := reactivationInfo.Remaining
		
		// Show current license status
		fmt.Printf("License Status: %s\n", status)
		fmt.Printf("License Key: %s\n", maskLicenseKey(licenseInfo.LicenseKey))
		fmt.Println(strings.Repeat("-", 55))
		
		// Show reactivation information
		fmt.Printf("Reactivations Used: %d of %d\n", reactivationCount, maxReactivations)
		fmt.Printf("Remaining Reactivations: %d\n", remaining)
		fmt.Printf("Reactivation Limit Resets: %s\n", reactivationInfo.NextResetDate.Format("January 2, 2006"))
		
		// Show warning messages based on usage
		if remaining == 0 {
			fmt.Println("\n🚨 WARNING: You have used all reactivations for this month!")
			fmt.Printf("💡 Reactivations reset on %s\n", reactivationInfo.NextResetDate.Format("January 2, 2006"))
			fmt.Println("   Contact support for immediate assistance if needed.")
		} else if remaining <= 2 {
			fmt.Printf("\n⚠️  WARNING: Only %d reactivation(s) remaining this month!\n", remaining)
			fmt.Println("💡 Please manage your device activations carefully.")
			fmt.Printf("   Reactivations reset on %s\n", reactivationInfo.NextResetDate.Format("January 2, 2006"))
		} else {
			fmt.Printf("\n✅ You have %d reactivations remaining this month.\n", remaining)
		}
		
		// Show additional context
		fmt.Println("\nReactivation Information:")
		fmt.Println("• Each license allows up to 5 reactivations per 30-day period")
		fmt.Println("• Reactivations reset on the 1st day of each month")
		fmt.Println("• Use reactivations when transferring to a new device")
		fmt.Println("• Contact support if you need additional reactivations")
	}
	
	fmt.Println(strings.Repeat("=", 55))
	fmt.Printf("Check performed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

// getDaysRemaining returns the number of days remaining for scripting
func getDaysRemaining() {
	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		fmt.Printf("-1")
		os.Exit(1)
	}
	
	// Get license status
	licenseInfo, _, err := licenseManager.GetLicenseStatus()
	if err != nil || licenseInfo == nil {
		fmt.Printf("-1")
		os.Exit(1)
	}
	
	// Calculate days remaining
	daysRemaining := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}
	
	// Just output the number for scripting
	fmt.Printf("%d", daysRemaining)
}

// generateSupportInfo generates a comprehensive support package
func generateSupportInfo() {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("support_info_%s.txt", timestamp)
	
	fmt.Printf("%s - Support Information Package\n", app.AppName)
	fmt.Printf("Version: %s\n", app.VERSION)
	fmt.Printf("Build Time: %s\n", app.BuildTime)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Generating support package: %s\n\n", filename)
	
	var reportLines []string
	
	// Add header
	reportLines = append(reportLines, 
		fmt.Sprintf("%s - Support Information Package", app.AppName),
		fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Version: %s", app.VERSION),
		fmt.Sprintf("Build Time: %s", app.BuildTime),
		"",
		strings.Repeat("=", 60),
		"",
	)
	
	// System Information
	reportLines = append(reportLines, "SYSTEM INFORMATION:")
	reportLines = append(reportLines, strings.Repeat("-", 30))
	reportLines = append(reportLines, fmt.Sprintf("Operating System: %s", runtime.GOOS))
	reportLines = append(reportLines, fmt.Sprintf("Architecture: %s", runtime.GOARCH))
	reportLines = append(reportLines, fmt.Sprintf("Go Version: %s", runtime.Version()))
	reportLines = append(reportLines, fmt.Sprintf("NumCPU: %d", runtime.NumCPU()))
	
	if hostname, err := os.Hostname(); err == nil {
		reportLines = append(reportLines, fmt.Sprintf("Hostname: %s", hostname))
	}
	
	// Working directory
	if wd, err := os.Getwd(); err == nil {
		reportLines = append(reportLines, fmt.Sprintf("Working Directory: %s", wd))
	}
	
	reportLines = append(reportLines, "")
	
	// Device Fingerprint
	reportLines = append(reportLines, "DEVICE FINGERPRINT:")
	reportLines = append(reportLines, strings.Repeat("-", 30))
	
	// Import security package for fingerprint generation
	fpManager := security.NewFingerprintManager()
	fingerprint, err := fpManager.GenerateFingerprint()
	if err == nil {
		reportLines = append(reportLines, fmt.Sprintf("Fingerprint: %s", fingerprint.Fingerprint))
		reportLines = append(reportLines, fmt.Sprintf("Hostname: %s", fingerprint.Hostname))
		reportLines = append(reportLines, fmt.Sprintf("MAC Address: %s", fingerprint.MACAddress))
		reportLines = append(reportLines, fmt.Sprintf("CPU ID: %s", fingerprint.CPUID))
		reportLines = append(reportLines, fmt.Sprintf("OS: %s", fingerprint.OS))
		reportLines = append(reportLines, fmt.Sprintf("Platform: %s", fingerprint.Platform))
	} else {
		reportLines = append(reportLines, fmt.Sprintf("Error generating fingerprint: %v", err))
	}
	
	reportLines = append(reportLines, "")
	
	// License Information
	reportLines = append(reportLines, "LICENSE INFORMATION:")
	reportLines = append(reportLines, strings.Repeat("-", 30))
	
	cfg, err := config.Load()
	if err != nil {
		reportLines = append(reportLines, fmt.Sprintf("Configuration Error: %v", err))
	} else {
		licenseFile := cfg.GetLicenseFile()
		licenseManager, err := license.NewManager()
		if err != nil {
			reportLines = append(reportLines, fmt.Sprintf("License Manager Error: %v", err))
		} else {
			licenseInfo, status, err := licenseManager.GetLicenseStatus()
			if err != nil {
				reportLines = append(reportLines, fmt.Sprintf("License Status Error: %v", err))
			} else if licenseInfo == nil {
				reportLines = append(reportLines, "License Status: Not activated")
				reportLines = append(reportLines, "License Key: N/A")
				reportLines = append(reportLines, "Expiry Date: N/A")
				reportLines = append(reportLines, "Days Remaining: N/A")
			} else {
				daysRemaining := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
				if daysRemaining < 0 {
					daysRemaining = 0
				}
				
				reportLines = append(reportLines, fmt.Sprintf("License Status: %s", status))
				reportLines = append(reportLines, fmt.Sprintf("License Key: %s", maskLicenseKey(licenseInfo.LicenseKey)))
				if licenseInfo.UserEmail != "" {
					reportLines = append(reportLines, fmt.Sprintf("User Email: %s", licenseInfo.UserEmail))
				}
				reportLines = append(reportLines, fmt.Sprintf("Duration: %s", licenseInfo.Duration))
				reportLines = append(reportLines, fmt.Sprintf("Issued Date: %s", licenseInfo.IssuedDate.Format("2006-01-02 15:04:05")))
				reportLines = append(reportLines, fmt.Sprintf("Expiry Date: %s", licenseInfo.ExpiryDate.Format("2006-01-02 15:04:05")))
				reportLines = append(reportLines, fmt.Sprintf("Days Remaining: %d", daysRemaining))
				reportLines = append(reportLines, fmt.Sprintf("Last Checked: %s", licenseInfo.LastChecked.Format("2006-01-02 15:04:05")))
				
				if licenseInfo.ActivationID != "" {
					reportLines = append(reportLines, fmt.Sprintf("Activation ID: %s", licenseInfo.ActivationID))
				}
			}
		}
		
		// Path information
		reportLines = append(reportLines, "")
		reportLines = append(reportLines, "PATH INFORMATION:")
		reportLines = append(reportLines, strings.Repeat("-", 30))
		reportLines = append(reportLines, fmt.Sprintf("License File: %s", licenseFile))
		
		if paths, err := config.GetPaths(); err == nil {
			reportLines = append(reportLines, fmt.Sprintf("Data Directory: %s", paths.DataDir))
			reportLines = append(reportLines, fmt.Sprintf("Logs Directory: %s", paths.LogsDir))
			reportLines = append(reportLines, fmt.Sprintf("Downloads Directory: %s", paths.DownloadsDir))
			reportLines = append(reportLines, fmt.Sprintf("Reports Directory: %s", paths.ReportsDir))
		}
	}
	
	reportLines = append(reportLines, "")
	
	// Recent Error Logs (last 50 lines from error logs if available)
	reportLines = append(reportLines, "RECENT LOGS:")
	reportLines = append(reportLines, strings.Repeat("-", 30))
	
	if paths, err := config.GetPaths(); err == nil {
		logFiles := []string{
			filepath.Join(paths.LogsDir, "error.log"),
			filepath.Join(paths.LogsDir, "app.log"),
			filepath.Join(paths.LogsDir, "license.log"),
		}
		
		logFound := false
		for _, logFile := range logFiles {
			if data, err := os.ReadFile(logFile); err == nil {
				lines := strings.Split(string(data), "\n")
				// Get last 20 lines from each log file
				start := len(lines) - 20
				if start < 0 {
					start = 0
				}
				
				if len(lines) > start {
					reportLines = append(reportLines, fmt.Sprintf("From %s (last 20 lines):", filepath.Base(logFile)))
					for i := start; i < len(lines); i++ {
						if strings.TrimSpace(lines[i]) != "" {
							reportLines = append(reportLines, "  " + lines[i])
						}
					}
					reportLines = append(reportLines, "")
					logFound = true
				}
			}
		}
		
		if !logFound {
			reportLines = append(reportLines, "No recent log files found or accessible.")
		}
	}
	
	reportLines = append(reportLines, "")
	reportLines = append(reportLines, strings.Repeat("=", 60))
	reportLines = append(reportLines, "End of Support Information Package")
	reportLines = append(reportLines, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	
	// Write to file
	content := strings.Join(reportLines, "\n")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Error writing support file: %v\n", err)
		fmt.Println("Support information:")
		fmt.Println(content)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Support package generated successfully: %s\n", filename)
	fmt.Printf("📦 File size: %d bytes\n", len(content))
	fmt.Println()
	fmt.Println("This file contains:")
	fmt.Println("• System information (OS, architecture, hostname)")
	fmt.Println("• Device fingerprint for hardware identification")
	fmt.Println("• License status and configuration details")
	fmt.Println("• Application paths and directory structure")
	fmt.Println("• Recent error logs (if available)")
	fmt.Println()
	fmt.Println("📧 Please attach this file when contacting support.")
	fmt.Println("⚠️  This file may contain sensitive information - share only with trusted support personnel.")
}

// testServerConnection tests connectivity to the activation server
func testServerConnection() {
	fmt.Printf("%s - Server Connection Test\n", app.AppName)
	fmt.Printf("Version: %s\n", app.VERSION)
	fmt.Println(strings.Repeat("=", 50))

	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		fmt.Printf("❌ License manager error: %v\n", err)
		os.Exit(3)
	}
	activator := license.NewActivationService(licenseManager)
	
	fmt.Println("🔍 Testing connection to activation server...")
	fmt.Println()
	
	// Test connectivity by attempting to validate a test license
	// This will fail but should test network connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	start := time.Now()
	
	// Use a known invalid key to test connectivity without activating anything
	testKey := "TEST-CONN-TEST-CONN-TEST"
	err = activator.Activate(ctx, testKey)
	duration := time.Since(start)
	
	if err != nil {
		errorStr := strings.ToLower(err.Error())
		
		// Check for network-related errors
		if strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "deadline exceeded") {
			fmt.Printf("❌ Connection test failed: Timeout after %v\n", duration)
			fmt.Println("   The activation server is not responding.")
			fmt.Println("💡 Check your internet connection and firewall settings.")
			os.Exit(1)
		} else if strings.Contains(errorStr, "network") || strings.Contains(errorStr, "connection") || strings.Contains(errorStr, "dial") {
			fmt.Printf("❌ Connection test failed: Network error after %v\n", duration)
			fmt.Println("   Cannot connect to the activation server.")
			fmt.Println("💡 Check your internet connection and firewall settings.")
			os.Exit(1)
		} else {
			// Server responded (even with an error), so connection works
			fmt.Printf("✅ Connection test successful! (%v)\n", duration)
			fmt.Println("   The activation server is reachable.")
			fmt.Println("   Server responded with: License validation error (expected)")
		}
	} else {
		// This shouldn't happen with our test key, but if it does, connection works
		fmt.Printf("✅ Connection test successful! (%v)\n", duration)
		fmt.Println("   The activation server is reachable.")
	}
	
	fmt.Println()
	fmt.Println("Connection Details:")
	fmt.Printf("• Response time: %v\n", duration)
	fmt.Println("• Server: ISX Pulse License Server")
	fmt.Println("• Protocol: HTTPS")
	
	fmt.Println()
	fmt.Printf("Test completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
