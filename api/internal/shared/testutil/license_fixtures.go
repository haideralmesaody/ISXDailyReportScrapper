package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/isxcli/isxcli/internal/license"
)

// LicenseTestFixtures provides test data and utilities for license testing
type LicenseTestFixtures struct {
	TestDataDir string
}

// NewLicenseTestFixtures creates a new fixtures manager
func NewLicenseTestFixtures(testDataDir string) *LicenseTestFixtures {
	return &LicenseTestFixtures{
		TestDataDir: testDataDir,
	}
}

// GetValidLicenseInfo returns a valid license info for testing
func (f *LicenseTestFixtures) GetValidLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-VALID-12345-TESTS-67890",
		UserEmail:   "test@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(30 * 24 * time.Hour), // 30 days from now
		Duration:    "monthly",
		IssuedDate:  time.Now().Add(-24 * time.Hour), // Issued yesterday
		Status:      "Active",
		LastChecked: time.Now().Add(-1 * time.Hour), // Checked 1 hour ago
	}
}

// GetExpiredLicenseInfo returns an expired license info for testing
func (f *LicenseTestFixtures) GetExpiredLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-EXPIR-12345-TESTS-67890",
		UserEmail:   "expired@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(-10 * 24 * time.Hour), // Expired 10 days ago
		Duration:    "monthly",
		IssuedDate:  time.Now().Add(-40 * 24 * time.Hour), // Issued 40 days ago
		Status:      "Expired",
		LastChecked: time.Now().Add(-2 * time.Hour), // Checked 2 hours ago
	}
}

// GetCriticalLicenseInfo returns a license that expires soon (critical state)
func (f *LicenseTestFixtures) GetCriticalLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-CRIT-12345-TESTS-67890",
		UserEmail:   "critical@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(3 * 24 * time.Hour), // Expires in 3 days
		Duration:    "monthly",
		IssuedDate:  time.Now().Add(-27 * 24 * time.Hour), // Issued 27 days ago
		Status:      "Critical",
		LastChecked: time.Now().Add(-30 * time.Minute), // Checked 30 minutes ago
	}
}

// GetWarningLicenseInfo returns a license in warning state (expires in 2 weeks)
func (f *LicenseTestFixtures) GetWarningLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-WARN-12345-TESTS-67890",
		UserEmail:   "warning@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(14 * 24 * time.Hour), // Expires in 2 weeks
		Duration:    "monthly",
		IssuedDate:  time.Now().Add(-16 * 24 * time.Hour), // Issued 16 days ago
		Status:      "Warning",
		LastChecked: time.Now().Add(-10 * time.Minute), // Checked 10 minutes ago
	}
}

// GetMachineMismatchLicenseInfo is deprecated - machine ID validation has been removed
func (f *LicenseTestFixtures) GetMachineMismatchLicenseInfo() license.LicenseInfo {
	// This function is kept for backward compatibility but machine ID is no longer relevant
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-MISM-12345-TESTS-67890",
		UserEmail:   "mismatch@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
		Duration:    "monthly",
		IssuedDate:  time.Now().Add(-5 * 24 * time.Hour),
		Status:      "Active",
		LastChecked: time.Now().Add(-15 * time.Minute),
	}
}

// GetLifetimeLicenseInfo returns a lifetime license for testing
func (f *LicenseTestFixtures) GetLifetimeLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-LIFE-12345-TESTS-67890",
		UserEmail:   "lifetime@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(10 * 365 * 24 * time.Hour), // 10 years from now
		Duration:    "lifetime",
		IssuedDate:  time.Now().Add(-100 * 24 * time.Hour), // Issued 100 days ago
		Status:      "Active",
		LastChecked: time.Now().Add(-5 * time.Minute),
	}
}

// GetYearlyLicenseInfo returns a yearly license for testing
func (f *LicenseTestFixtures) GetYearlyLicenseInfo() license.LicenseInfo {
	return license.LicenseInfo{
		LicenseKey:  "ISX1Y-YEAR-12345-TESTS-67890",
		UserEmail:   "yearly@iraqiinvestor.gov.iq",
		ExpiryDate:  time.Now().Add(300 * 24 * time.Hour), // ~10 months remaining
		Duration:    "yearly",
		IssuedDate:  time.Now().Add(-65 * 24 * time.Hour), // Issued ~2 months ago
		Status:      "Active",
		LastChecked: time.Now().Add(-20 * time.Minute),
	}
}

// GetTestLicenseKeys returns various test license keys for different scenarios
func (f *LicenseTestFixtures) GetTestLicenseKeys() map[string]string {
	return map[string]string{
		"live":           "ISX1M02LYE1F9QJHR9D7Z", // The actual live license key
		"valid_format":   "ISX1Y-ABCDE-12345-FGHIJ-67890",
		"valid_format2":  "ISX1Y-12345-ABCDE-67890-FGHIJ",
		"invalid_prefix": "WRONG-ABCDE-12345-FGHIJ-67890",
		"too_short":      "ISX1Y-ABC-123-FGH-678",
		"too_long":       "ISX1Y-ABCDEF-123456-FGHIJK-678901",
		"no_dashes":      "ISX1YABCDE12345FGHIJ67890",
		"lowercase":      "isx1y-abcde-12345-fghij-67890",
		"special_chars":  "ISX1Y-ABC@E-12345-FGH!J-67890",
		"empty":          "",
		"spaces":         "   ",
		"null_bytes":     "ISX1Y-ABCDE\x00-12345-FGHIJ-67890",
	}
}

// GetTestEmails returns various test email addresses
func (f *LicenseTestFixtures) GetTestEmails() map[string]string {
	return map[string]string{
		"valid_standard":   "test@example.com",
		"valid_subdomain":  "user@mail.example.com",
		"valid_gov":        "user@iraqiinvestor.gov.iq",
		"valid_unicode":    "üser@exämple.com",
		"invalid_no_at":    "testexample.com",
		"invalid_no_domain": "test@",
		"invalid_no_local": "@example.com",
		"invalid_no_dot":   "test@example",
		"invalid_end_dot":  "test@example.com.",
		"empty":            "",
		"spaces":           "   ",
		"xss_attempt":      "<script>alert('xss')</script>@test.com",
		"sql_injection":    "'; DROP TABLE users; --@test.com",
		"very_long":        fmt.Sprintf("%s@example.com", string(make([]byte, 1000))),
	}
}

// CreateTestLicenseFile creates a test license file with given license info
func (f *LicenseTestFixtures) CreateTestLicenseFile(filepath string, info license.LicenseInfo) error {
	// Ensure directory exists
	dir := filepath[:len(filepath)-len(filepath[len(filepath)-1:])]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serialize license info to JSON (in real implementation this would be encrypted)
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal license info: %w", err)
	}

	// Write to file
	err = os.WriteFile(filepath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write license file: %w", err)
	}

	return nil
}

// LoadTestLicenseFile loads a license info from a test file
func (f *LicenseTestFixtures) LoadTestLicenseFile(filepath string) (*license.LicenseInfo, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}

	var info license.LicenseInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal license info: %w", err)
	}

	return &info, nil
}

// CreateCorruptedLicenseFile creates various types of corrupted license files for testing
func (f *LicenseTestFixtures) CreateCorruptedLicenseFile(filepath, corruptionType string) error {
	dir := filepath[:len(filepath)-len(filepath[len(filepath)-1:])]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var data []byte
	
	switch corruptionType {
	case "empty":
		data = []byte{}
	case "invalid_json":
		data = []byte("{invalid json content}")
	case "wrong_structure":
		data = []byte(`{"wrong": "structure", "missing": "fields"}`)
	case "binary_data":
		data = make([]byte, 256)
		for i := range data {
			data[i] = byte(i % 256)
		}
	case "partial_json":
		data = []byte(`{"LicenseKey": "PARTIAL"`)
	case "huge_file":
		// Create a very large file
		data = make([]byte, 10*1024*1024) // 10MB
		for i := range data {
			data[i] = 'A'
		}
	case "null_bytes":
		data = []byte("{\x00\"LicenseKey\x00\": \"TEST\x00\"}")
	default:
		return fmt.Errorf("unknown corruption type: %s", corruptionType)
	}

	err := os.WriteFile(filepath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write corrupted file: %w", err)
	}

	return nil
}

// GetMockAPIResponses returns mock API responses for testing
func (f *LicenseTestFixtures) GetMockAPIResponses() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"activation_success": {
			"success":      true,
			"message":      "License activated successfully for Iraqi Investor",
			"activated_at": time.Now().Format(time.RFC3339),
			"trace_id":     "mock-trace-12345",
		},
		"activation_invalid_key": {
			"type":    "/errors/validation",
			"title":   "Validation Error",
			"status":  400,
			"detail":  "Invalid license key format",
			"trace_id": "mock-trace-12345",
			"success": false,
		},
		"activation_expired": {
			"type":    "/errors/license-expired",
			"title":   "License Expired", 
			"status":  410,
			"detail":  "The provided license key has expired",
			"trace_id": "mock-trace-12345",
			"success": false,
		},
		"activation_machine_mismatch": {
			"type":    "/errors/machine-mismatch",
			"title":   "Machine Mismatch",
			"status":  409,
			"detail":  "License is already activated on another machine",
			"trace_id": "mock-trace-12345",
			"success": false,
		},
		"activation_network_error": {
			"type":    "/errors/network",
			"title":   "Network Error",
			"status":  503,
			"detail":  "Unable to connect to license server",
			"trace_id": "mock-trace-12345",
			"success": false,
		},
		"status_not_activated": {
			"license_status": "not_activated",
			"message":        "No license activated. Please activate a license to access Iraqi Investor features.",
			"trace_id":       "mock-trace-12345",
			"timestamp":      time.Now().Format(time.RFC3339),
			"type":           "/license/not-activated",
			"title":          "License Not Activated",
			"status":         200,
			"branding_info": map[string]interface{}{
				"application_name": "ISX Daily Reports Scrapper",
				"brand_name":       "Iraqi Investor",
				"website_url":      "https://iraqiinvestor.gov.iq",
				"support_email":    "support@iraqiinvestor.gov.iq",
			},
		},
		"status_active": {
			"license_status": "active",
			"message":        "License is active with 30 days remaining",
			"trace_id":       "mock-trace-12345",
			"timestamp":      time.Now().Format(time.RFC3339),
			"days_left":      30,
			"license_info": map[string]interface{}{
				"license_key": "ISX1Y-****-****-****-****",
				"user_email":  "test@iraqiinvestor.gov.iq",
				"expiry_date": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
				"status":      "Active",
			},
		},
		"detailed_status": {
			"license_status":   "active",
			"machine_id":       "MACHINE12...",
			"validation_count": 100,
			"network_status":   "connected",
			"activation_date":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"last_validation":  time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"recommendations": []string{
				"License is healthy and functioning properly",
				"No action required at this time",
			},
			"trace_id": "mock-trace-12345",
		},
		"renewal_status_healthy": {
			"needs_renewal":    false,
			"is_expired":       false,
			"days_until_expiry": 180,
			"renewal_urgency":  "low",
			"renewal_message":  "License is active with 180 days remaining. No immediate action required.",
			"contact_info":     "For renewal assistance, contact support@iraqiinvestor.gov.iq",
			"trace_id":         "mock-trace-12345",
		},
		"renewal_status_critical": {
			"needs_renewal":    true,
			"is_expired":       false,
			"days_until_expiry": 3,
			"renewal_urgency":  "critical",
			"renewal_message":  "License expires in 3 days! Urgent renewal needed to avoid service interruption.",
			"contact_info":     "Contact support@iraqiinvestor.gov.iq immediately for license renewal",
			"trace_id":         "mock-trace-12345",
		},
	}
}

// GetTestScenarios returns predefined test scenarios
func (f *LicenseTestFixtures) GetTestScenarios() map[string]TestScenario {
	return map[string]TestScenario{
		"happy_path": {
			Name:        "Happy Path License Activation",
			Description: "Test successful license activation with valid key and email",
			LicenseKey:  "ISX1Y-ABCDE-12345-FGHIJ-67890",
			Email:       "test@iraqiinvestor.gov.iq",
			ExpectedHTTPStatus: 200,
			ExpectedSuccess:    true,
		},
		"invalid_license_format": {
			Name:        "Invalid License Format",
			Description: "Test activation with invalid license key format",
			LicenseKey:  "INVALID-FORMAT",
			Email:       "test@example.com",
			ExpectedHTTPStatus: 400,
			ExpectedSuccess:    false,
		},
		"invalid_email": {
			Name:        "Invalid Email Format",
			Description: "Test activation with invalid email address",
			LicenseKey:  "ISX1Y-ABCDE-12345-FGHIJ-67890",
			Email:       "invalid-email",
			ExpectedHTTPStatus: 400,
			ExpectedSuccess:    false,
		},
		"empty_license_key": {
			Name:        "Empty License Key",
			Description: "Test activation with empty license key",
			LicenseKey:  "",
			Email:       "test@example.com",
			ExpectedHTTPStatus: 400,
			ExpectedSuccess:    false,
		},
		"live_license_key": {
			Name:        "Live License Key Test",
			Description: "Test activation with the live license key",
			LicenseKey:  "ISX1M02LYE1F9QJHR9D7Z",
			Email:       "live.test@iraqiinvestor.gov.iq",
			ExpectedHTTPStatus: -1, // Variable based on actual license state
			ExpectedSuccess:    false, // May succeed or fail
		},
	}
}

// TestScenario represents a test scenario
type TestScenario struct {
	Name               string
	Description        string
	LicenseKey         string
	Email              string
	ExpectedHTTPStatus int
	ExpectedSuccess    bool
	ExpectedErrorType  string
	Setup              func() error
	Cleanup            func() error
}

// GenerateTestDataFiles creates test data files in the specified directory
func (f *LicenseTestFixtures) GenerateTestDataFiles() error {
	if err := os.MkdirAll(f.TestDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create test data directory: %w", err)
	}

	// Create various license files
	testCases := map[string]license.LicenseInfo{
		"valid_license.json":    f.GetValidLicenseInfo(),
		"expired_license.json":  f.GetExpiredLicenseInfo(),
		"critical_license.json": f.GetCriticalLicenseInfo(),
		"warning_license.json":  f.GetWarningLicenseInfo(),
		"lifetime_license.json": f.GetLifetimeLicenseInfo(),
		"yearly_license.json":   f.GetYearlyLicenseInfo(),
	}

	for filename, info := range testCases {
		path := filepath.Join(f.TestDataDir, filename)
		if err := f.CreateTestLicenseFile(path, info); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	// Create corrupted files
	corruptedFiles := []string{
		"empty_file.json",
		"invalid_json.json", 
		"wrong_structure.json",
		"binary_data.dat",
		"partial_json.json",
	}

	for i, filename := range corruptedFiles {
		path := filepath.Join(f.TestDataDir, filename)
		corruptionTypes := []string{"empty", "invalid_json", "wrong_structure", "binary_data", "partial_json"}
		if err := f.CreateCorruptedLicenseFile(path, corruptionTypes[i]); err != nil {
			return fmt.Errorf("failed to create corrupted file %s: %w", filename, err)
		}
	}

	// Create API response fixtures
	apiResponses := f.GetMockAPIResponses()
	for name, response := range apiResponses {
		filename := fmt.Sprintf("api_response_%s.json", name)
		path := filepath.Join(f.TestDataDir, filename)
		
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal API response %s: %w", name, err)
		}
		
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write API response file %s: %w", filename, err)
		}
	}

	return nil
}

// CleanupTestData removes all test data files
func (f *LicenseTestFixtures) CleanupTestData() error {
	return os.RemoveAll(f.TestDataDir)
}