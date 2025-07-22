package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"isxcli/internal/license"
)

func main() {
	var (
		userEmail = flag.String("email", "", "User email address")
		duration  = flag.String("duration", "1m", "License duration (1m, 3m, 6m, 1y)")
		config    = flag.String("config", "license-config.json", "Google Sheets config file")
	)
	flag.Parse()

	if *userEmail == "" {
		fmt.Println("Usage: license-generator -email=user@example.com -duration=1m")
		fmt.Println("Durations: 1m (1 month), 3m (3 months), 6m (6 months), 1y (1 year)")
		os.Exit(1)
	}

	// Get the directory of the executable
	exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Failed to get executable directory:", err)
	}

	configPath := filepath.Join(exeDir, *config)

	// Create license manager
	manager, err := license.NewManagerWithConfig(configPath, "")
	if err != nil {
		log.Fatal("Failed to create license manager:", err)
	}

	// Generate license
	licenseKey, err := manager.GenerateLicense(*userEmail, *duration)
	if err != nil {
		log.Fatal("Failed to generate license:", err)
	}

	fmt.Printf("\n🎫 LICENSE GENERATED SUCCESSFULLY!\n")
	fmt.Printf("═══════════════════════════════════\n")
	fmt.Printf("📧 Email:      %s\n", *userEmail)
	fmt.Printf("⏱️  Duration:   %s\n", *duration)
	fmt.Printf("🔑 License:    %s\n", licenseKey)
	fmt.Printf("═══════════════════════════════════\n")
	fmt.Printf("\n📋 Instructions for user:\n")
	fmt.Printf("1. Run the ISX scraper application\n")
	fmt.Printf("2. When prompted, enter this license key: %s\n", licenseKey)
	fmt.Printf("3. The application will be activated for %s\n", getDurationText(*duration))
	fmt.Printf("\n💾 License has been saved to Google Sheets for tracking.\n")
}

func getDurationText(duration string) string {
	switch duration {
	case "1m":
		return "1 month"
	case "3m":
		return "3 months"
	case "6m":
		return "6 months"
	case "1y":
		return "1 year"
	default:
		return duration
	}
}
