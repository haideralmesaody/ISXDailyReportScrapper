package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// ServiceAccountConfig represents the configuration for service account authentication
type ServiceAccountConfig struct {
	SpreadsheetID   string `json:"spreadsheet_id"`
	SheetName       string `json:"sheet_name"`
	CredentialsFile string `json:"credentials_file"`
}

// LicenseGenerator manages license generation using service account
type LicenseGenerator struct {
	config        *ServiceAccountConfig
	sheetsService *sheets.Service
}

// License represents a generated license
type License struct {
	Key           string
	Duration      string
	ExpiryDate    string
	Status        string
	MachineID     string
	ActivatedDate string
}

// NewLicenseGenerator creates a new license generator with service account auth
func NewLicenseGenerator(configFile string) (*LicenseGenerator, error) {
	config, err := loadServiceAccountConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Create Sheets service with service account
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(config.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %v", err)
	}

	return &LicenseGenerator{
		config:        config,
		sheetsService: sheetsService,
	}, nil
}

// loadServiceAccountConfig loads configuration from JSON file
func loadServiceAccountConfig(filename string) (*ServiceAccountConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ServiceAccountConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// generateLicenseKey generates a random license key
func generateLicenseKey(duration string) (string, error) {
	var prefix string
	switch duration {
	case "1 Month":
		prefix = "ISX1M"
	case "3 Months":
		prefix = "ISX3M"
	case "6 Months":
		prefix = "ISX6M"
	case "1 Year":
		prefix = "ISX1Y"
	default:
		prefix = "ISX1M"
	}

	// Generate random string
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return prefix + string(b), nil
}

// calculateExpiryDate calculates expiry date based on duration
func calculateExpiryDate(duration string) string {
	now := time.Now()
	var expiryDate time.Time

	switch duration {
	case "1 Month":
		expiryDate = now.AddDate(0, 1, 0)
	case "3 Months":
		expiryDate = now.AddDate(0, 3, 0)
	case "6 Months":
		expiryDate = now.AddDate(0, 6, 0)
	case "1 Year":
		expiryDate = now.AddDate(1, 0, 0)
	default:
		expiryDate = now.AddDate(0, 1, 0)
	}

	return expiryDate.Format("2006-01-02")
}

// generateLicense creates a new license
func (lg *LicenseGenerator) generateLicense(duration string) (*License, error) {
	key, err := generateLicenseKey(duration)
	if err != nil {
		return nil, err
	}

	return &License{
		Key:           key,
		Duration:      duration,
		ExpiryDate:    calculateExpiryDate(duration),
		Status:        "Available",
		MachineID:     "",
		ActivatedDate: "",
	}, nil
}

// appendLicenseToSheet adds a license to the Google Sheet
func (lg *LicenseGenerator) appendLicenseToSheet(license *License) error {
	spreadsheetID := lg.config.SpreadsheetID
	sheetName := lg.config.SheetName

	// Calculate expire status for new licenses (Available since not activated yet)
	expireStatus := "Available"

	// Prepare the values to append
	// Format: LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate | LastConnected | ExpireStatus
	values := [][]interface{}{
		{
			license.Key,
			license.Duration,
			license.ExpiryDate,
			license.Status,
			license.MachineID,
			license.ActivatedDate,
			"",           // LastConnected - empty for new licenses
			expireStatus, // ExpireStatus - Available for new licenses
		},
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Append the data (extended to column H)
	_, err := lg.sheetsService.Spreadsheets.Values.Append(
		spreadsheetID,
		sheetName+"!A:H", // Extended to column H
		valueRange,
	).ValueInputOption("RAW").Do()

	return err
}

// generateLicenses generates the specified number of licenses
func (lg *LicenseGenerator) generateLicenses(total int) error {
	durations := []string{"1 Month", "3 Months", "6 Months", "1 Year"}

	fmt.Printf("🎫 ISX License Generator v3.0 (Service Account)\n")
	fmt.Printf("📊 Generating %d licenses...\n\n", total)

	for i := 0; i < total; i++ {
		// Distribute licenses across different durations
		duration := durations[i%len(durations)]

		license, err := lg.generateLicense(duration)
		if err != nil {
			return fmt.Errorf("failed to generate license %d: %v", i+1, err)
		}

		// Add to Google Sheet
		err = lg.appendLicenseToSheet(license)
		if err != nil {
			return fmt.Errorf("failed to add license %d to sheet: %v", i+1, err)
		}

		// Rate limiting - wait 1.2 seconds between requests to stay under 60/minute
		time.Sleep(1200 * time.Millisecond)

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == total {
			fmt.Printf("✅ Generated %d/%d licenses\n", i+1, total)
		}
	}

	fmt.Printf("\n🎉 License Generation Complete!\n")
	fmt.Printf("📄 Check your Google Sheet: https://docs.google.com/spreadsheets/d/%s\n", lg.config.SpreadsheetID)

	return nil
}

// exportLicensesToFile exports licenses to a text file
func (lg *LicenseGenerator) exportLicensesToFile(filename string, total int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("ISX License Keys\n")
	file.WriteString("================\n\n")

	// Read from Google Sheet
	spreadsheetID := lg.config.SpreadsheetID
	sheetName := lg.config.SheetName

	resp, err := lg.sheetsService.Spreadsheets.Values.Get(
		spreadsheetID,
		sheetName+"!A:F",
	).Do()
	if err != nil {
		return err
	}

	for i, row := range resp.Values {
		if i == 0 {
			continue // Skip header
		}
		if len(row) >= 2 {
			file.WriteString(fmt.Sprintf("License: %s | Duration: %s\n", row[0], row[1]))
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: license-generator-sa.exe -total=<number>")
		fmt.Println("Example: license-generator-sa.exe -total=100")
		return
	}

	var total int
	var err error

	// Parse command line arguments
	for _, arg := range os.Args[1:] {
		if arg[:7] == "-total=" {
			total, err = strconv.Atoi(arg[7:])
			if err != nil {
				log.Fatal("Invalid total number:", err)
			}
		}
	}

	if total <= 0 {
		log.Fatal("Total must be a positive number")
	}

	// Create license generator
	generator, err := NewLicenseGenerator("service-account-config.json")
	if err != nil {
		log.Fatal("Failed to create license generator:", err)
	}

	// Generate licenses
	err = generator.generateLicenses(total)
	if err != nil {
		log.Fatal("Failed to generate licenses:", err)
	}

	// Export to file
	filename := fmt.Sprintf("licenses_%d_%s.txt", total, time.Now().Format("20060102_150405"))
	err = generator.exportLicensesToFile(filename, total)
	if err != nil {
		log.Printf("Warning: Failed to export to file: %v", err)
	} else {
		fmt.Printf("📁 Exported to: %s\n", filename)
	}
}
