// +build example

package config

import (
	"fmt"
	"log"
	"time"
	"log/slog"
	"os"
)

// ExampleUsage demonstrates how to use the paths package throughout the application
func ExampleUsage() {
	// Always get paths from the centralized GetPaths() function
	paths, err := GetPaths()
	if err != nil {
		slog.Error("Failed to get paths", slog.String("error", err.Error())); os.Exit(1)
	}

	// Ensure all directories exist at startup
	if err := paths.EnsureDirectories(); err != nil {
		slog.Error("Failed to ensure directories", slog.String("error", err.Error())); os.Exit(1)
	}

	// Log all resolved paths for debugging
	paths.LogPathResolution()

	// Example 1: License Manager usage
	licensePath := paths.LicenseFile
	slog.Info("License will be saved/loaded from: %s\n", licensePath)

	// Example 2: Scraper downloading files
	today := time.Now()
	excelPath := paths.GetExcelPathForDate(today)
	slog.Info("Today's Excel file should be saved to: %s\n", excelPath)

	// Example 3: Data processor generating reports
	dailyCSV := paths.GetDailyCSVPath(today)
	slog.Info("Daily CSV will be generated at: %s\n", dailyCSV)
	
	tickerCSV := paths.GetTickerDailyCSVPath("BBOB")
	slog.Info("BBOB ticker CSV will be at: %s\n", tickerCSV)

	// Example 4: Index CSV processor
	indexCSV := paths.GetIndexCSVPath()
	slog.Info("Index CSV is at: %s\n", indexCSV)

	// Example 5: Configuration files
	credentialsPath := paths.GetCredentialsPath()
	sheetsConfigPath := paths.GetSheetsConfigPath()
	slog.Info("Google credentials: %s\n", credentialsPath)
	slog.Info("Sheets config: %s\n", sheetsConfigPath)

	// Example 6: Validate required files exist before starting
	if err := paths.ValidateRequiredFiles(); err != nil {
		slog.Info("Warning: %v", err)
		// Application might want to handle missing files gracefully
	}

	// Example 7: Using the license path helper (for backward compatibility)
	licensePath2, err := GetLicensePath()
	if err != nil {
		slog.Info("Failed to get license path: %v", err)
	}
	slog.Info("License path (via helper): %s\n", licensePath2)
}

// Migration Guide:
//
// OLD CODE (problematic):
//   licensePath := filepath.Join(os.Getwd(), "license.dat")
//   excelPath := "data/downloads/file.xlsx"
//
// NEW CODE (correct):
//   paths, _ := config.GetPaths()
//   licensePath := paths.LicenseFile
//   excelPath := paths.GetExcelPath("file.xlsx")
//
// Benefits:
// 1. All paths relative to executable, not working directory
// 2. Consistent across all components
// 3. Cross-platform path handling
// 4. Centralized logging and debugging
// 5. Easy to test and mock