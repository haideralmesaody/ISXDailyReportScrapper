package license

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// LicenseInfo represents license data
type LicenseInfo struct {
	LicenseKey  string    `json:"license_key"`
	UserEmail   string    `json:"user_email"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Duration    string    `json:"duration"`
	MachineID   string    `json:"machine_id"`
	IssuedDate  time.Time `json:"issued_date"`
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
}

// GoogleSheetsConfig represents Google Sheets configuration
type GoogleSheetsConfig struct {
	SheetID            string `json:"sheet_id"`
	APIKey             string `json:"api_key"`
	SheetName          string `json:"sheet_name"`
	UseServiceAccount  bool   `json:"use_service_account"`
	ServiceAccountFile string `json:"service_account_file"`
}

// Manager handles license operations
type Manager struct {
	config        GoogleSheetsConfig
	licenseFile   string
	machineID     string
	sheetsService *sheets.Service
}

// NewManager creates a new license manager
func NewManager(configFile, licenseFile string) (*Manager, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	machineID, err := generateMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine ID: %v", err)
	}

	manager := &Manager{
		config:      config,
		licenseFile: licenseFile,
		machineID:   machineID,
	}

	// Initialize Google Sheets service if using service account
	if config.UseServiceAccount {
		ctx := context.Background()
		sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(config.ServiceAccountFile))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service: %v", err)
		}
		manager.sheetsService = sheetsService
	}

	return manager, nil
}

// GenerateLicense creates a new license key
func (m *Manager) GenerateLicense(userEmail string, duration string) (string, error) {
	// Generate random license key
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	licenseKey := base64.URLEncoding.EncodeToString(bytes)
	licenseKey = strings.ReplaceAll(licenseKey, "=", "")

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

	licenseKey = fmt.Sprintf("%s-%s", prefix, licenseKey)

	// Calculate expiry date
	var expiryDate time.Time
	switch duration {
	case "1m":
		expiryDate = time.Now().AddDate(0, 1, 0)
	case "3m":
		expiryDate = time.Now().AddDate(0, 3, 0)
	case "6m":
		expiryDate = time.Now().AddDate(0, 6, 0)
	case "1y":
		expiryDate = time.Now().AddDate(1, 0, 0)
	default:
		expiryDate = time.Now().AddDate(0, 1, 0)
	}

	// Create license info
	license := LicenseInfo{
		LicenseKey:  licenseKey,
		UserEmail:   userEmail,
		ExpiryDate:  expiryDate,
		Duration:    duration,
		MachineID:   "", // Will be set when activated
		IssuedDate:  time.Now(),
		Status:      "issued",
		LastChecked: time.Now(),
	}

	// Save to Google Sheets
	if err := m.saveLicenseToSheets(license); err != nil {
		return "", fmt.Errorf("failed to save license: %v", err)
	}

	return licenseKey, nil
}

// ActivateLicense activates a license for the current machine
func (m *Manager) ActivateLicense(licenseKey string) error {
	// Validate license from Google Sheets
	license, err := m.validateLicenseFromSheets(licenseKey)
	if err != nil {
		return fmt.Errorf("license validation failed: %v", err)
	}

	// Check if license is available for activation
	if license.Status != "Available" {
		if license.Status == "Activated" {
			return fmt.Errorf("license already activated")
		}
		return fmt.Errorf("license status is '%s' - not available for activation", license.Status)
	}

	// Check if already activated on another machine
	if license.MachineID != "" && license.MachineID != m.machineID {
		return fmt.Errorf("license already activated on another machine")
	}

	// Calculate expiry date based on duration (for recharge cards)
	var expiryDate time.Time
	now := time.Now()
	switch license.Duration {
	case "1 Month":
		expiryDate = now.AddDate(0, 1, 0)
	case "3 Months":
		expiryDate = now.AddDate(0, 3, 0)
	case "6 Months":
		expiryDate = now.AddDate(0, 6, 0)
	case "1 Year":
		expiryDate = now.AddDate(1, 0, 0)
	default:
		expiryDate = now.AddDate(0, 1, 0) // Default to 1 month
	}

	// Activate license
	license.MachineID = m.machineID
	license.Status = "Activated"
	license.ExpiryDate = expiryDate
	license.IssuedDate = now
	license.LastChecked = now
	license.UserEmail = "" // Recharge cards don't have user emails initially

	// Save locally
	if err := m.saveLicenseLocal(license); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	// Update Google Sheets
	if err := m.updateLicenseInSheets(license); err != nil {
		return fmt.Errorf("failed to update license in sheets: %v", err)
	}

	return nil
}

// ValidateLicense checks if current license is valid
func (m *Manager) ValidateLicense() (bool, error) {
	// Load local license
	license, err := m.loadLicenseLocal()
	if err != nil {
		return false, fmt.Errorf("no local license found: %v", err)
	}

	// Check expiry
	if time.Now().After(license.ExpiryDate) {
		license.Status = "expired"
		m.saveLicenseLocal(license)
		return false, fmt.Errorf("license expired on %s", license.ExpiryDate.Format("2006-01-02"))
	}

	// Check machine ID
	if license.MachineID != m.machineID {
		return false, fmt.Errorf("license not valid for this machine")
	}

	// Periodic validation with Google Sheets (once per day)
	if time.Since(license.LastChecked) > 24*time.Hour {
		if err := m.validateWithSheets(license); err != nil {
			return false, fmt.Errorf("remote validation failed: %v", err)
		}
	}

	return true, nil
}

// UpdateLastConnected updates the last connected time in both local storage and Google Sheets
func (m *Manager) UpdateLastConnected() error {
	// Load current license
	license, err := m.loadLicenseLocal()
	if err != nil {
		return fmt.Errorf("no local license found: %v", err)
	}

	// Update last checked time
	license.LastChecked = time.Now()

	// Save locally
	if err := m.saveLicenseLocal(license); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	// Update Google Sheets with expire status
	if err := m.updateLicenseInSheets(license); err != nil {
		return fmt.Errorf("failed to update last connected time in sheets: %v", err)
	}

	return nil
}

// GetLicenseInfo returns current license information
func (m *Manager) GetLicenseInfo() (*LicenseInfo, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return nil, err
	}
	return &license, nil
}

// generateMachineID creates a unique machine identifier
func generateMachineID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	// Create a hash based on hostname and other machine-specific data
	h := md5.New()
	h.Write([]byte(hostname))

	// Add some additional entropy
	if user := os.Getenv("USERNAME"); user != "" {
		h.Write([]byte(user))
	}
	if user := os.Getenv("USER"); user != "" {
		h.Write([]byte(user))
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash[:16], nil
}

// loadConfig loads Google Sheets configuration
func loadConfig(configFile string) (GoogleSheetsConfig, error) {
	var config GoogleSheetsConfig

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}

// saveLicenseLocal saves license to local file
func (m *Manager) saveLicenseLocal(license LicenseInfo) error {
	data, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.licenseFile, data, 0600)
}

// loadLicenseLocal loads license from local file
func (m *Manager) loadLicenseLocal() (LicenseInfo, error) {
	var license LicenseInfo

	data, err := os.ReadFile(m.licenseFile)
	if err != nil {
		return license, err
	}

	err = json.Unmarshal(data, &license)
	return license, err
}

// saveLicenseToSheets saves license to Google Sheets
func (m *Manager) saveLicenseToSheets(license LicenseInfo) error {
	// Implementation for Google Sheets API
	// This would use the Google Sheets API to append a new row
	// Format: [LicenseKey, UserEmail, ExpiryDate, Duration, MachineID, IssuedDate, Status, LastChecked]

	values := []interface{}{
		license.LicenseKey,
		license.UserEmail,
		license.ExpiryDate.Format("2006-01-02 15:04:05"),
		license.Duration,
		license.MachineID,
		license.IssuedDate.Format("2006-01-02 15:04:05"),
		license.Status,
		license.LastChecked.Format("2006-01-02 15:04:05"),
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:append?valueInputOption=RAW&key=%s",
		m.config.SheetID, m.config.SheetName, m.config.APIKey)

	payload := map[string]interface{}{
		"values": [][]interface{}{values},
	}

	return m.makeSheetRequest("POST", url, payload)
}

// validateLicenseFromSheets validates license against Google Sheets
func (m *Manager) validateLicenseFromSheets(licenseKey string) (LicenseInfo, error) {
	var license LicenseInfo

	if m.config.UseServiceAccount && m.sheetsService != nil {
		// Use service account authentication
		resp, err := m.sheetsService.Spreadsheets.Values.Get(m.config.SheetID, m.config.SheetName).Do()
		if err != nil {
			return license, fmt.Errorf("failed to read from sheets: %v", err)
		}

		// Parse sheet data and find license
		for i, row := range resp.Values {
			if i == 0 {
				continue // Skip header row
			}
			if len(row) >= 1 && row[0].(string) == licenseKey {
				// Recharge card format: LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate | LastConnected
				license.LicenseKey = row[0].(string)

				// Duration (column B)
				if len(row) > 1 {
					license.Duration = row[1].(string)
				}

				// ExpiryDate (column C) - may be empty for Available licenses
				if len(row) > 2 && row[2].(string) != "" {
					if expiryDate, err := time.Parse("2006-01-02", row[2].(string)); err == nil {
						license.ExpiryDate = expiryDate
					}
				}

				// Status (column D)
				if len(row) > 3 {
					license.Status = row[3].(string)
				}

				// MachineID (column E)
				if len(row) > 4 {
					license.MachineID = row[4].(string)
				}

				// ActivatedDate (column F)
				if len(row) > 5 && row[5].(string) != "" {
					if activatedDate, err := time.Parse("2006-01-02", row[5].(string)); err == nil {
						license.IssuedDate = activatedDate
					}
				}

				// LastConnected (column G) - new field
				if len(row) > 6 && row[6].(string) != "" {
					if lastConnected, err := time.Parse("2006-01-02 15:04:05", row[6].(string)); err == nil {
						license.LastChecked = lastConnected
					}
				} else {
					// Set default if column doesn't exist yet
					license.LastChecked = time.Now()
				}

				// ExpireStatus (column H) - new field (optional, for future use)
				// This is automatically calculated, so we don't need to parse it here

				// Set defaults for recharge cards
				license.UserEmail = "" // Recharge cards don't have user emails

				return license, nil
			}
		}
	} else {
		// Fallback to API key method
		url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?key=%s",
			m.config.SheetID, m.config.SheetName, m.config.APIKey)

		resp, err := http.Get(url)
		if err != nil {
			return license, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return license, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return license, err
		}

		// Parse sheet data and find license
		if values, ok := result["values"].([]interface{}); ok {
			for i, row := range values {
				if i == 0 {
					continue // Skip header row
				}
				if rowData, ok := row.([]interface{}); ok && len(rowData) >= 4 {
					// Check if this is our license key
					if len(rowData) > 0 && rowData[0].(string) == licenseKey {
						// Recharge card format: LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate
						license.LicenseKey = rowData[0].(string)

						// Duration (column B)
						if len(rowData) > 1 {
							license.Duration = rowData[1].(string)
						}

						// ExpiryDate (column C) - may be empty for Available licenses
						if len(rowData) > 2 && rowData[2].(string) != "" {
							if expiryDate, err := time.Parse("2006-01-02", rowData[2].(string)); err == nil {
								license.ExpiryDate = expiryDate
							}
						}

						// Status (column D)
						if len(rowData) > 3 {
							license.Status = rowData[3].(string)
						}

						// MachineID (column E)
						if len(rowData) > 4 {
							license.MachineID = rowData[4].(string)
						}

						// ActivatedDate (column F)
						if len(rowData) > 5 && rowData[5].(string) != "" {
							if activatedDate, err := time.Parse("2006-01-02", rowData[5].(string)); err == nil {
								license.IssuedDate = activatedDate
							}
						}

						// LastConnected (column G) - new field
						if len(rowData) > 6 && rowData[6].(string) != "" {
							if lastConnected, err := time.Parse("2006-01-02 15:04:05", rowData[6].(string)); err == nil {
								license.LastChecked = lastConnected
							}
						} else {
							// Set default if column doesn't exist yet
							license.LastChecked = time.Now()
						}

						// ExpireStatus (column H) - new field (optional, for future use)
						// This is automatically calculated, so we don't need to parse it here

						// Set defaults for recharge cards
						license.UserEmail = "" // Recharge cards don't have user emails

						return license, nil
					}
				}
			}
		}
	}

	return license, fmt.Errorf("license not found")
}

// updateLicenseInSheets updates license in Google Sheets
func (m *Manager) updateLicenseInSheets(license LicenseInfo) error {
	if m.config.UseServiceAccount && m.sheetsService != nil {
		// Use service account authentication
		// First, find the row number for this license
		resp, err := m.sheetsService.Spreadsheets.Values.Get(m.config.SheetID, m.config.SheetName).Do()
		if err != nil {
			return fmt.Errorf("failed to read from sheets: %v", err)
		}

		var rowIndex int = -1
		for i, row := range resp.Values {
			if i == 0 {
				continue // Skip header row
			}
			if len(row) > 0 && row[0].(string) == license.LicenseKey {
				rowIndex = i + 1 // Google Sheets uses 1-based indexing
				break
			}
		}

		if rowIndex == -1 {
			return fmt.Errorf("license not found in sheet")
		}

		// Calculate expire status
		expireStatus := m.calculateExpireStatus(license.ExpiryDate)

		// Update the row with new license data
		// Format: LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate | LastConnected | ExpireStatus
		values := [][]interface{}{
			{
				license.LicenseKey,
				license.Duration,
				license.ExpiryDate.Format("2006-01-02"),
				license.Status,
				license.MachineID,
				license.IssuedDate.Format("2006-01-02"),
				license.LastChecked.Format("2006-01-02 15:04:05"), // Add LastConnected timestamp
				expireStatus,                                      // Add ExpireStatus
			},
		}

		rangeStr := fmt.Sprintf("%s!A%d:H%d", m.config.SheetName, rowIndex, rowIndex) // Extended to column H
		valueRange := &sheets.ValueRange{Values: values}

		_, err = m.sheetsService.Spreadsheets.Values.Update(
			m.config.SheetID,
			rangeStr,
			valueRange,
		).ValueInputOption("RAW").Do()

		return err
	} else {
		// Fallback to API key method
		// First, find the row number for this license
		url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?key=%s",
			m.config.SheetID, m.config.SheetName, m.config.APIKey)

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}

		var rowIndex int = -1
		if values, ok := result["values"].([]interface{}); ok {
			for i, row := range values {
				if i == 0 {
					continue // Skip header row
				}
				if rowData, ok := row.([]interface{}); ok && len(rowData) > 0 {
					if rowData[0].(string) == license.LicenseKey {
						rowIndex = i + 1 // Google Sheets uses 1-based indexing
						break
					}
				}
			}
		}

		if rowIndex == -1 {
			return fmt.Errorf("license not found in sheet")
		}

		// Calculate expire status
		expireStatus := m.calculateExpireStatus(license.ExpiryDate)

		// Update the row with new license data
		// Format: LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate | LastConnected | ExpireStatus
		values := [][]interface{}{
			{
				license.LicenseKey,
				license.Duration,
				license.ExpiryDate.Format("2006-01-02"),
				license.Status,
				license.MachineID,
				license.IssuedDate.Format("2006-01-02"),
				license.LastChecked.Format("2006-01-02 15:04:05"), // Add LastConnected timestamp
				expireStatus,                                      // Add ExpireStatus
			},
		}

		updateURL := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s!A%d:H%d?valueInputOption=RAW&key=%s",
			m.config.SheetID, m.config.SheetName, rowIndex, rowIndex, m.config.APIKey) // Extended to column H

		payload := map[string]interface{}{
			"values": values,
		}

		return m.makeSheetRequest("PUT", updateURL, payload)
	}
}

// validateWithSheets performs periodic validation with Google Sheets
func (m *Manager) validateWithSheets(license LicenseInfo) error {
	sheetLicense, err := m.validateLicenseFromSheets(license.LicenseKey)
	if err != nil {
		return err
	}

	// Check if license status changed
	if sheetLicense.Status != "Activated" {
		return fmt.Errorf("license is no longer active - status: %s", sheetLicense.Status)
	}

	// Update last checked time locally
	license.LastChecked = time.Now()

	// Save updated license locally
	if err := m.saveLicenseLocal(license); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	// Update Google Sheets with current timestamp to track "last connected"
	if err := m.updateLicenseInSheets(license); err != nil {
		// Don't fail if Google Sheets update fails, but log it
		// This prevents loss of local functionality if there are connectivity issues
		fmt.Printf("Warning: Failed to update last connected time in Google Sheets: %v\n", err)
	}

	return nil
}

// makeSheetRequest makes HTTP request to Google Sheets API
func (m *Manager) makeSheetRequest(method, url string, payload interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// calculateExpireStatus calculates the expire status based on days remaining
func (m *Manager) calculateExpireStatus(expiryDate time.Time) string {
	if expiryDate.IsZero() {
		return "Available" // For unactivated licenses
	}

	daysLeft := int(time.Until(expiryDate).Hours() / 24)

	if daysLeft <= 0 {
		return "Expired"
	} else if daysLeft <= 7 {
		return "Critical" // Red - 7 or fewer days
	} else if daysLeft <= 30 {
		return "Warning" // Yellow - 8-30 days
	} else {
		return "Active" // Green - more than 30 days
	}
}
