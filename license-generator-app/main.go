package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// LicenseInfo represents license data
type LicenseInfo struct {
	LicenseKey    string    `json:"license_key"`
	Duration      string    `json:"duration"`
	ExpiryDate    time.Time `json:"expiry_date"`
	Status        string    `json:"status"`
	MachineID     string    `json:"machine_id"`
	IssuedDate    time.Time `json:"issued_date"`
	ActivatedDate time.Time `json:"activated_date"`
}

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	SheetID      string                 `json:"sheet_id"`
	SheetName    string                 `json:"sheet_name"`
	ClientID     string                 `json:"client_id"`
	ClientSecret string                 `json:"client_secret"`
	RedirectURL  string                 `json:"redirect_url"`
	Credentials  map[string]interface{} `json:"credentials"`
}

// LicenseGenerator handles license generation
type LicenseGenerator struct {
	config        OAuth2Config
	sheetsService *sheets.Service
}

func main() {
	var (
		count1m    = flag.Int("1m", 0, "Number of 1-month licenses to generate")
		count3m    = flag.Int("3m", 0, "Number of 3-month licenses to generate")
		count6m    = flag.Int("6m", 0, "Number of 6-month licenses to generate")
		count1y    = flag.Int("1y", 0, "Number of 1-year licenses to generate")
		total      = flag.Int("total", 0, "Total number of random licenses to generate")
		output     = flag.String("output", "", "Output file to save license keys (optional)")
		configFile = flag.String("config", "oauth-config.json", "Configuration file path")
	)
	flag.Parse()

	fmt.Println("🎫 ISX License Generator v2.0 (OAuth2)")
	fmt.Println("═══════════════════════════════════════")

	// Initialize random seed
	mathrand.Seed(time.Now().UnixNano())

	// Load configuration
	generator, err := NewLicenseGenerator(*configFile)
	if err != nil {
		log.Fatal("❌ Failed to initialize license generator:", err)
	}

	var generatedLicenses []LicenseInfo
	totalGenerated := 0

	// Generate specific duration licenses
	durations := []struct {
		duration string
		count    int
	}{
		{"1m", *count1m},
		{"3m", *count3m},
		{"6m", *count6m},
		{"1y", *count1y},
	}

	for _, d := range durations {
		if d.count > 0 {
			fmt.Printf("\n🔄 Generating %d licenses for %s duration...\n", d.count, d.duration)
			for i := 0; i < d.count; i++ {
				license, err := generator.GenerateRechargeCardLicense(d.duration)
				if err != nil {
					log.Printf("❌ Error generating license %d of %d for %s: %v", i+1, d.count, d.duration, err)
					continue
				}
				generatedLicenses = append(generatedLicenses, license)
				totalGenerated++
				if (i+1)%10 == 0 {
					fmt.Printf("   ✅ Generated %d/%d licenses for %s\n", i+1, d.count, d.duration)
				}
				// Small delay to avoid overwhelming the API
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	// Generate random duration licenses
	if *total > 0 {
		fmt.Printf("\n🔄 Generating %d random duration licenses...\n", *total)
		randomDurations := []string{"1m", "3m", "6m", "1y"}
		for i := 0; i < *total; i++ {
			duration := randomDurations[mathrand.Intn(len(randomDurations))]
			license, err := generator.GenerateRechargeCardLicense(duration)
			if err != nil {
				log.Printf("❌ Error generating random license %d of %d: %v", i+1, *total, err)
				continue
			}
			generatedLicenses = append(generatedLicenses, license)
			totalGenerated++
			if (i+1)%10 == 0 {
				fmt.Printf("   ✅ Generated %d/%d random licenses\n", i+1, *total)
			}
			// Small delay to avoid overwhelming the API
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Save to file if requested
	if *output != "" && len(generatedLicenses) > 0 {
		err := saveLicensesToFile(*output, generatedLicenses)
		if err != nil {
			log.Printf("❌ Error saving licenses to file: %v", err)
		} else {
			fmt.Printf("💾 Saved %d licenses to %s\n", len(generatedLicenses), *output)
		}
	}

	// Display summary
	fmt.Printf("\n🎉 License Generation Complete!\n")
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("✅ Total licenses generated: %d\n", totalGenerated)
	fmt.Printf("🔗 Google Sheet: https://docs.google.com/spreadsheets/d/%s/edit\n", generator.config.SheetID)

	if len(generatedLicenses) > 0 {
		fmt.Printf("\n📋 Sample licenses:\n")
		for i, license := range generatedLicenses {
			if i < 5 { // Show first 5 licenses
				fmt.Printf("   • %s (%s)\n", license.LicenseKey, license.Duration)
			}
		}
		if len(generatedLicenses) > 5 {
			fmt.Printf("   ... and %d more\n", len(generatedLicenses)-5)
		}
	}
}

// NewLicenseGenerator creates a new license generator with OAuth2
func NewLicenseGenerator(configFile string) (*LicenseGenerator, error) {
	config, err := loadOAuth2Config(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Create OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{sheets.SpreadsheetsScope},
		Endpoint:     google.Endpoint,
	}

	// Get OAuth2 token
	ctx := context.Background()
	client, err := getOAuth2Client(ctx, oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 client: %v", err)
	}

	// Create Sheets service
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %v", err)
	}

	return &LicenseGenerator{
		config:        config,
		sheetsService: sheetsService,
	}, nil
}

// GenerateRechargeCardLicense creates a new recharge card license
func (lg *LicenseGenerator) GenerateRechargeCardLicense(duration string) (LicenseInfo, error) {
	// Generate unique license key
	licenseKey := lg.generateLicenseKey(duration)

	// Create license info (no expiry date until activated)
	license := LicenseInfo{
		LicenseKey: licenseKey,
		Duration:   duration,
		Status:     "Available", // Available for activation
		IssuedDate: time.Now(),
		// ExpiryDate will be set when activated
		// MachineID will be set when activated
		// ActivatedDate will be set when activated
	}

	// Save to Google Sheets
	err := lg.saveLicenseToSheets(license)
	if err != nil {
		return LicenseInfo{}, fmt.Errorf("failed to save license to sheets: %w", err)
	}

	return license, nil
}

// generateLicenseKey creates a unique license key
func (lg *LicenseGenerator) generateLicenseKey(duration string) string {
	// Generate random bytes
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based random
		bytes = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}

	// Encode to base64 and clean up
	licenseKey := base64.URLEncoding.EncodeToString(bytes)
	licenseKey = strings.ReplaceAll(licenseKey, "=", "")
	licenseKey = strings.ReplaceAll(licenseKey, "+", "")
	licenseKey = strings.ReplaceAll(licenseKey, "/", "")

	// Take first 12 characters
	if len(licenseKey) > 12 {
		licenseKey = licenseKey[:12]
	}

	// Add prefix based on duration
	prefix := "ISX"
	switch duration {
	case "1m":
		prefix = "ISX1M"
	case "3m":
		prefix = "ISX3M"
	case "6m":
		prefix = "ISX6M"
	case "1y":
		prefix = "ISX1Y"
	}

	// Add timestamp for uniqueness
	timestamp := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)

	return fmt.Sprintf("%s-%s-%s", prefix, licenseKey, timestamp)
}

// saveLicenseToSheets saves license to Google Sheets using OAuth2
func (lg *LicenseGenerator) saveLicenseToSheets(license LicenseInfo) error {
	// Prepare data for Google Sheets
	values := [][]interface{}{
		{
			license.LicenseKey,
			license.Duration,
			"", // ExpiryDate (empty until activated)
			license.Status,
			"", // MachineID (empty until activated)
			"", // ActivatedDate (empty until activated)
		},
	}

	// Create value range
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Append to sheet
	_, err := lg.sheetsService.Spreadsheets.Values.Append(
		lg.config.SheetID,
		lg.config.SheetName,
		valueRange,
	).ValueInputOption("RAW").Do()

	return err
}

// getOAuth2Client gets an OAuth2 client
func getOAuth2Client(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	// Try to load token from file
	token, err := tokenFromFile("token.json")
	if err != nil {
		// Get token from web
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %v", err)
		}
		// Save token for future use
		saveTokenToFile("token.json", token)
	}

	return config.Client(ctx, token), nil
}

// getTokenFromWeb gets a token from the web
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("🔐 OAuth2 Authentication Required\n")
	fmt.Printf("═══════════════════════════════════\n")
	fmt.Printf("📋 Please visit this URL to authorize the application:\n")
	fmt.Printf("🔗 %s\n\n", authURL)
	fmt.Printf("✏️  Enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("failed to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %v", err)
	}

	return tok, nil
}

// tokenFromFile loads token from file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveTokenToFile saves token to file
func saveTokenToFile(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// loadOAuth2Config loads OAuth2 configuration from file
func loadOAuth2Config(configFile string) (OAuth2Config, error) {
	var config OAuth2Config

	file, err := os.Open(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to decode config: %v", err)
	}

	return config, nil
}

// saveLicensesToFile saves licenses to a text file
func saveLicensesToFile(filename string, licenses []LicenseInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write header
	_, err = file.WriteString("ISX License Keys\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString("================\n\n")
	if err != nil {
		return err
	}

	// Write licenses
	for _, license := range licenses {
		_, err = file.WriteString(fmt.Sprintf("%s (%s)\n", license.LicenseKey, license.Duration))
		if err != nil {
			return err
		}
	}

	return nil
}
