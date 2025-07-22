package license

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	ServiceAccountJSON string `json:"service_account_json"` // Embedded JSON credentials
}

// PerformanceMetrics tracks operation performance
type PerformanceMetrics struct {
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	MaxTime      time.Duration `json:"max_time"`
	MinTime      time.Duration `json:"min_time"`
	ErrorCount   int64         `json:"error_count"`
	SuccessCount int64         `json:"success_count"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// Manager handles license operations with enhanced logging, caching, and security
type Manager struct {
	config          GoogleSheetsConfig
	licenseFile     string
	machineID       string
	sheetsService   *sheets.Service
	logger          *Logger
	cache           *LicenseCache
	security        *SecurityManager
	performanceData map[string]*PerformanceMetrics
	perfMutex       sync.RWMutex
	// Add validation state tracking
	lastValidationResult *ValidationResult
	lastValidationTime   time.Time
	validationMutex      sync.RWMutex
}

// ValidationResult holds cached validation results
type ValidationResult struct {
	IsValid     bool
	Error       error
	ErrorType   string // "machine_mismatch", "expired", "network_error", etc.
	CachedUntil time.Time
	RetryAfter  time.Duration
}

// RenewalInfo contains information about license renewal requirements
type RenewalInfo struct {
	DaysLeft     int    `json:"days_left"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	NeedsRenewal bool   `json:"needs_renewal"`
	IsExpired    bool   `json:"is_expired"`
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// getBuiltInConfig returns the embedded Google Sheets configuration
// Credentials are compiled directly into the binary for self-contained deployment
func getBuiltInConfig() GoogleSheetsConfig {
	// Production credentials are loaded from environment or file
	// To use this package, create a service account credentials JSON file
	// and either:
	// 1. Set ISX_CREDENTIALS environment variable with the JSON content
	// 2. Place credentials.json in the same directory as the executable
	// 3. Replace this placeholder during build process
	serviceAccountJSON := os.Getenv("ISX_CREDENTIALS")
	if serviceAccountJSON == "" {
		// Try to load from file if environment variable not set
		if credData, err := os.ReadFile("credentials.json"); err == nil {
			serviceAccountJSON = string(credData)
		} else {
			// Use placeholder that will fail validation
			serviceAccountJSON = `{"type": "service_account", "project_id": "PLACEHOLDER"}`
		}
	}

	// Credentials loaded successfully

	// Validate JSON structure (without logging sensitive data)
	var testData map[string]interface{}
	if err := json.Unmarshal([]byte(serviceAccountJSON), &testData); err != nil {
		// JSON validation failed, but don't log the error details
		return GoogleSheetsConfig{}
	}
	
	// Sheet configuration embedded in binary
	sheetID := "1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc"
	sheetName := "Licenses"

	config := GoogleSheetsConfig{
		SheetID:            sheetID,
		SheetName:          sheetName,
		UseServiceAccount:  true,
		ServiceAccountJSON: serviceAccountJSON,
	}
	
	// Return configuration
	return config
}

// NewManager creates a new license manager with enhanced capabilities
func NewManager(licenseFile string) (*Manager, error) {
	// Use built-in configuration (self-contained)
	// Use built-in configuration (self-contained mode)
	config := getBuiltInConfig()

	machineID, err := generateMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine ID: %v", err)
	}

	// Initialize logger
	logger, err := NewLogger(LogLevelInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}

	// Initialize cache (30 minute TTL, max 1000 entries)
	cache := NewLicenseCache(30*time.Minute, 1000)

	// Initialize security manager (max 5 attempts, 15 minute block, 5 minute window)
	security := NewSecurityManager(5, 15*time.Minute, 5*time.Minute, logger)

	manager := &Manager{
		config:          config,
		licenseFile:     licenseFile,
		machineID:       machineID,
		logger:          logger,
		cache:           cache,
		security:        security,
		performanceData: make(map[string]*PerformanceMetrics),
	}

	// Log manager initialization
	logger.Log(LogEntry{
		Level:     LogLevelInfo,
		Action:    "manager_initialization",
		Result:    "License manager initialized successfully",
		MachineID: machineID[:min(8, len(machineID))],
		Details: map[string]interface{}{
			"cache_ttl":               "30m",
			"cache_max_size":          1000,
			"security_max_attempts":   5,
			"security_block_duration": "15m",
		},
	})

	// Initialize Google Sheets service with embedded credentials
	if config.UseServiceAccount && config.ServiceAccountJSON != "" {
		// Initialize Google Sheets service with embedded credentials
		
		ctx := context.Background()

		// Create temporary credentials from embedded JSON
		// Create credentials from embedded JSON
		credentialsOption := option.WithCredentialsJSON([]byte(config.ServiceAccountJSON))
		
		// Initialize sheets service
		sheetsService, err := sheets.NewService(ctx, credentialsOption)
		if err != nil {
			// Log error through proper logger, not console
			logger.Log(LogEntry{
				Level:  LogLevelError,
				Action: "sheets_initialization",
				Result: "Failed to initialize Google Sheets service",
				Error:  err.Error(),
			})
			return nil, fmt.Errorf("failed to create sheets service with embedded credentials: %v", err)
		}
		
		// Service initialized successfully
		manager.sheetsService = sheetsService

		logger.Log(LogEntry{
			Level:  LogLevelInfo,
			Action: "sheets_initialization",
			Result: "Google Sheets service initialized successfully",
		})
	} else {
		// Google Sheets initialization skipped
	}

	return manager, nil
}

// NewManagerWithConfig creates a new license manager with custom configuration (for backward compatibility)
func NewManagerWithConfig(configFile, licenseFile string) (*Manager, error) {
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
		var sheetsService *sheets.Service
		var err error

		if config.ServiceAccountJSON != "" {
			// Use embedded JSON credentials
			credentialsOption := option.WithCredentialsJSON([]byte(config.ServiceAccountJSON))
			sheetsService, err = sheets.NewService(ctx, credentialsOption)
		} else {
			// Use external credentials file - make path absolute
			credentialsPath := config.ServiceAccountFile
			if !filepath.IsAbs(credentialsPath) {
				// If path is relative, make it relative to current working directory
				if wd, err := os.Getwd(); err == nil {
					credentialsPath = filepath.Join(wd, credentialsPath)
				}
			}
			sheetsService, err = sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
		}

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

	// Calculate expiry date - expires at 12am next day after standard period
	var standardExpiry time.Time
	switch duration {
	case "1m":
		standardExpiry = time.Now().AddDate(0, 1, 0)
	case "3m":
		standardExpiry = time.Now().AddDate(0, 3, 0)
	case "6m":
		standardExpiry = time.Now().AddDate(0, 6, 0)
	case "1y":
		standardExpiry = time.Now().AddDate(1, 0, 0)
	default:
		standardExpiry = time.Now().AddDate(0, 1, 0)
	}
	// Set expiry to 12:00 AM next day after standard expiry
	expiryDate := time.Date(standardExpiry.Year(), standardExpiry.Month(), standardExpiry.Day()+1, 0, 0, 0, 0, standardExpiry.Location())

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

// ActivateLicense activates a license with enhanced tracking and security
func (m *Manager) ActivateLicense(licenseKey string) error {
	return m.TrackOperation("license_activation", func() error {
		return m.performActivation(licenseKey)
	})
}

// performActivation contains the actual license activation logic
func (m *Manager) performActivation(licenseKey string) error {
	// Validate input
	if licenseKey == "" {
		return fmt.Errorf("license key cannot be empty")
	}

	// Check rate limiting (use license key as identifier for now)
	identifier := licenseKey[:min(8, len(licenseKey))]
	if m.security != nil && m.security.IsBlocked(identifier) {
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelWarn,
				Action:     "license_activation",
				Result:     "Activation blocked due to rate limiting",
				LicenseKey: identifier,
			})
		}
		return fmt.Errorf("too many failed attempts - please try again later")
	}

	// Log activation attempt
	if m.logger != nil {
		m.logger.Log(LogEntry{
			Level:      LogLevelInfo,
			Action:     "license_activation",
			Result:     "Starting license activation",
			LicenseKey: identifier,
			MachineID:  m.machineID[:min(8, len(m.machineID))],
		})
	}

	// Try Google Sheets validation first (remove local validation fallback for now)

	// Test Google Sheets connectivity first
	if m.sheetsService == nil {
		return fmt.Errorf("Google Sheets service not initialized - network connectivity may be an issue")
	}

	// Try to validate the license from Google Sheets (with caching)
	licenseInfo, err := m.validateLicenseFromSheetsWithCache(licenseKey)
	if err != nil {
		// Record failed attempt
		if m.security != nil {
			m.security.RecordAttempt(identifier, false)
		}

		// Provide more specific error context
		if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("connection timeout while accessing license validation service - please check your internet connection")
		} else if strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "network is unreachable") {
			return fmt.Errorf("network connection error - please check your internet connection and firewall settings")
		} else if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "unauthorized") {
			return fmt.Errorf("license validation service access denied - please contact support")
		} else if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("invalid license key - license not found in our system")
		}
		return fmt.Errorf("license validation failed: %v", err)
	}

	// Check if license is already activated on a different machine
	if licenseInfo.MachineID != "" && licenseInfo.MachineID != m.machineID {
		if m.security != nil {
			m.security.RecordAttempt(identifier, false)
		}
		return fmt.Errorf("license is already activated on another machine (Machine ID: %s)", licenseInfo.MachineID[:8])
	}

	// Handle Available status licenses - calculate expiry date during activation
	if licenseInfo.Status == "Available" || licenseInfo.ExpiryDate.IsZero() {
		// Calculate expiry date - expires at 12am next day after standard period
		var standardExpiry time.Time
		switch licenseInfo.Duration {
		case "1m":
			standardExpiry = time.Now().AddDate(0, 1, 0)
		case "3m":
			standardExpiry = time.Now().AddDate(0, 3, 0)
		case "6m":
			standardExpiry = time.Now().AddDate(0, 6, 0)
		case "1y":
			standardExpiry = time.Now().AddDate(1, 0, 0)
		default:
			standardExpiry = time.Now().AddDate(0, 1, 0) // Default to 1 month
		}
		// Set expiry to 12:00 AM next day after standard expiry
		licenseInfo.ExpiryDate = time.Date(standardExpiry.Year(), standardExpiry.Month(), standardExpiry.Day()+1, 0, 0, 0, 0, standardExpiry.Location())
		licenseInfo.IssuedDate = time.Now()
	} else {
		// Check if already activated license has expired
		if time.Now().After(licenseInfo.ExpiryDate) {
			if m.security != nil {
				m.security.RecordAttempt(identifier, false)
			}
			return fmt.Errorf("license has expired on %s", licenseInfo.ExpiryDate.Format("2006-01-02"))
		}
	}

	// Update license with machine ID and activation info
	licenseInfo.MachineID = m.machineID
	licenseInfo.Status = "Activated"
	licenseInfo.LastChecked = time.Now()

	// Save license locally
	if err := m.saveLicenseLocal(licenseInfo); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	// Update license in Google Sheets
	if err := m.updateLicenseInSheets(licenseInfo); err != nil {
		// Don't fail activation if we can't update sheets, but log the warning
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelWarn,
				Action:     "license_activation",
				Result:     "Failed to update Google Sheets",
				LicenseKey: identifier,
				Error:      err.Error(),
			})
		}
	}

	// Invalidate cache to ensure fresh data on next validation
	if m.cache != nil {
		m.cache.Invalidate(licenseKey)
	}

	// Record successful attempt
	if m.security != nil {
		m.security.RecordAttempt(identifier, true)
	}

	// Log successful activation
	if m.logger != nil {
		daysLeft := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
		m.logger.Log(LogEntry{
			Level:      LogLevelInfo,
			Action:     "license_activation",
			Result:     "License activated successfully",
			LicenseKey: identifier,
			MachineID:  m.machineID[:min(8, len(m.machineID))],
			Details: map[string]interface{}{
				"expiry_date": licenseInfo.ExpiryDate.Format("2006-01-02"),
				"duration":    licenseInfo.Duration,
				"days_left":   daysLeft,
			},
		})
	}

	return nil
}

// ValidateLicense checks if current license is valid with enhanced tracking
func (m *Manager) ValidateLicense() (bool, error) {
	// Check if we have a recent cached result
	m.validationMutex.RLock()
	if m.lastValidationResult != nil && time.Now().Before(m.lastValidationResult.CachedUntil) {
		result := m.lastValidationResult
		m.validationMutex.RUnlock()
		return result.IsValid, result.Error
	}
	m.validationMutex.RUnlock()

	// Perform actual validation
	var valid bool
	var err error

	trackErr := m.TrackOperation("license_validation_complete", func() error {
		valid, err = m.performValidation()

		// Cache the result with appropriate duration
		m.cacheValidationResult(valid, err)

		if !valid {
			return err
		}
		return nil
	})

	if trackErr != nil {
		return false, trackErr
	}

	return valid, err
}

// cacheValidationResult caches validation results with appropriate durations
func (m *Manager) cacheValidationResult(isValid bool, err error) {
	m.validationMutex.Lock()
	defer m.validationMutex.Unlock()

	result := &ValidationResult{
		IsValid: isValid,
		Error:   err,
	}

	if isValid {
		// Cache successful validations for 30 minutes
		result.CachedUntil = time.Now().Add(30 * time.Minute)
	} else if err != nil {
		// Determine error type and cache duration
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "license_machine_mismatch") {
			result.ErrorType = "machine_mismatch"
			// Cache machine mismatch errors for 10 minutes to avoid spam
			result.CachedUntil = time.Now().Add(10 * time.Minute)
			result.RetryAfter = 1 * time.Minute
		} else if strings.Contains(errorMsg, "expired") {
			result.ErrorType = "expired"
			// Cache expiry errors for 1 hour
			result.CachedUntil = time.Now().Add(1 * time.Hour)
			result.RetryAfter = 5 * time.Minute
		} else {
			result.ErrorType = "network_error"
			// Cache network errors for 2 minutes
			result.CachedUntil = time.Now().Add(2 * time.Minute)
			result.RetryAfter = 30 * time.Second
		}
	}

	m.lastValidationResult = result
	m.lastValidationTime = time.Now()
}

// GetValidationState returns the current validation state for better user feedback
func (m *Manager) GetValidationState() (*ValidationResult, error) {
	m.validationMutex.RLock()
	defer m.validationMutex.RUnlock()

	if m.lastValidationResult == nil {
		return nil, fmt.Errorf("no validation performed yet")
	}

	// Return a copy to avoid concurrent access issues
	result := *m.lastValidationResult
	return &result, nil
}

// performValidation contains the actual validation logic
func (m *Manager) performValidation() (bool, error) {
	// Load local license
	license, err := m.loadLicenseLocal()
	if err != nil {
		return false, fmt.Errorf("no local license found: %v", err)
	}

	// Check expiry
	if time.Now().After(license.ExpiryDate) {
		license.Status = "expired"
		m.saveLicenseLocal(license)

		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelWarn,
				Action:     "license_validation",
				Result:     "License expired",
				LicenseKey: license.LicenseKey[:min(8, len(license.LicenseKey))],
				MachineID:  m.machineID[:min(8, len(m.machineID))],
				Details: map[string]interface{}{
					"expiry_date": license.ExpiryDate.Format("2006-01-02"),
				},
			})
		}

		return false, fmt.Errorf("license expired on %s", license.ExpiryDate.Format("2006-01-02"))
	}

	// Check machine ID
	if license.MachineID != m.machineID {
		// Only log machine mismatch errors once per hour to avoid spam
		shouldLog := false
		if m.logger != nil {
			// Check if we've logged this recently
			if m.lastValidationTime.IsZero() || time.Since(m.lastValidationTime) > 1*time.Hour {
				shouldLog = true
			}
		}

		if shouldLog && m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelError,
				Action:     "license_validation",
				Result:     "License not valid for this machine",
				LicenseKey: license.LicenseKey[:min(8, len(license.LicenseKey))],
				MachineID:  m.machineID[:min(8, len(m.machineID))],
				Details: map[string]interface{}{
					"expected_machine_id": license.MachineID[:min(8, len(license.MachineID))],
					"current_machine_id":  m.machineID[:min(8, len(m.machineID))],
					"user_action":         "Contact Iraqi Investor to get new license for this machine",
				},
			})
		}
		return false, fmt.Errorf("license_machine_mismatch")
	}

	// Periodic validation with Google Sheets (every 6 hours for better security)
	if time.Since(license.LastChecked) > 6*time.Hour {
		if err := m.validateWithSheets(license); err != nil {
			// For better user experience, don't fail immediately on network issues
			// Log the error but allow offline usage for up to 48 hours total
			if time.Since(license.LastChecked) > 48*time.Hour {
				if m.logger != nil {
					m.logger.Log(LogEntry{
						Level:      LogLevelError,
						Action:     "license_validation",
						Result:     "Remote validation failed and grace period expired",
						LicenseKey: license.LicenseKey[:min(8, len(license.LicenseKey))],
						Error:      err.Error(),
					})
				}
				return false, fmt.Errorf("remote validation failed and offline grace period expired: %v", err)
			}
			// Just log the warning but continue with local validation
			if m.logger != nil {
				m.logger.Log(LogEntry{
					Level:      LogLevelWarn,
					Action:     "license_validation",
					Result:     "Remote validation failed, using local cache",
					LicenseKey: license.LicenseKey[:min(8, len(license.LicenseKey))],
					Error:      err.Error(),
				})
			}
		}
	}

	return true, nil
}

// TransferLicense transfers a license with enhanced tracking
func (m *Manager) TransferLicense(licenseKey string, forceTransfer bool) error {
	return m.TrackOperation("license_transfer", func() error {
		return m.performTransfer(licenseKey, forceTransfer)
	})
}

// performTransfer contains the actual transfer logic
func (m *Manager) performTransfer(licenseKey string, forceTransfer bool) error {
	// Validate input
	if licenseKey == "" {
		return fmt.Errorf("license key cannot be empty")
	}

	identifier := licenseKey[:min(8, len(licenseKey))]

	// Check rate limiting
	if m.security != nil && m.security.IsBlocked(identifier) {
		return fmt.Errorf("too many failed attempts - please try again later")
	}

	// Log transfer attempt
	if m.logger != nil {
		m.logger.Log(LogEntry{
			Level:      LogLevelInfo,
			Action:     "license_transfer",
			Result:     "Starting license transfer",
			LicenseKey: identifier,
			MachineID:  m.machineID[:min(8, len(m.machineID))],
			Details: map[string]interface{}{
				"force_transfer": forceTransfer,
			},
		})
	}

	// Test Google Sheets connectivity first
	if m.sheetsService == nil {
		return fmt.Errorf("Google Sheets service not initialized - network connectivity may be an issue")
	}

	// Try to validate the license from Google Sheets (with caching)
	licenseInfo, err := m.validateLicenseFromSheetsWithCache(licenseKey)
	if err != nil {
		if m.security != nil {
			m.security.RecordAttempt(identifier, false)
		}
		return fmt.Errorf("license validation failed: %v", err)
	}

	// Check if license has expired
	if time.Now().After(licenseInfo.ExpiryDate) {
		if m.security != nil {
			m.security.RecordAttempt(identifier, false)
		}
		return fmt.Errorf("license has expired on %s", licenseInfo.ExpiryDate.Format("2006-01-02"))
	}

	// Check if license is already activated on a different machine
	if licenseInfo.MachineID != "" && licenseInfo.MachineID != m.machineID {
		if !forceTransfer {
			if m.security != nil {
				m.security.RecordAttempt(identifier, false)
			}
			return fmt.Errorf("license is already activated on another machine (Machine ID: %s). Use force transfer if this is intentional", licenseInfo.MachineID[:8])
		}

		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelWarn,
				Action:     "license_transfer",
				Result:     "Forcing license transfer from another machine",
				LicenseKey: identifier,
				Details: map[string]interface{}{
					"previous_machine_id": licenseInfo.MachineID[:min(8, len(licenseInfo.MachineID))],
					"new_machine_id":      m.machineID[:min(8, len(m.machineID))],
				},
			})
		}
	}

	// Update license with new machine ID
	licenseInfo.MachineID = m.machineID
	licenseInfo.Status = "Activated"
	licenseInfo.LastChecked = time.Now()

	// Save license locally
	if err := m.saveLicenseLocal(licenseInfo); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	// Update license in Google Sheets
	if err := m.updateLicenseInSheets(licenseInfo); err != nil {
		// Don't fail transfer if we can't update sheets, but log the warning
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelWarn,
				Action:     "license_transfer",
				Result:     "Failed to update Google Sheets",
				LicenseKey: identifier,
				Error:      err.Error(),
			})
		}
	}

	// Invalidate cache
	if m.cache != nil {
		m.cache.Invalidate(licenseKey)
	}

	// Record successful attempt
	if m.security != nil {
		m.security.RecordAttempt(identifier, true)
	}

	// Log successful transfer
	if m.logger != nil {
		daysLeft := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
		m.logger.Log(LogEntry{
			Level:      LogLevelInfo,
			Action:     "license_transfer",
			Result:     "License transferred successfully",
			LicenseKey: identifier,
			MachineID:  m.machineID[:min(8, len(m.machineID))],
			Details: map[string]interface{}{
				"expiry_date":    licenseInfo.ExpiryDate.Format("2006-01-02"),
				"days_left":      daysLeft,
				"force_transfer": forceTransfer,
			},
		})
	}

	return nil
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

// generateMachineID creates a unique machine identifier using hardware fingerprinting
func generateMachineID() (string, error) {
	var fingerprint strings.Builder

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	fingerprint.WriteString(hostname)

	// Get username from environment
	user := os.Getenv("USERNAME")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user != "" {
		fingerprint.WriteString(user)
	}

	// Get OS and architecture
	fingerprint.WriteString(runtime.GOOS)
	fingerprint.WriteString(runtime.GOARCH)

	// Try to get MAC address
	if macAddr := getMACAddress(); macAddr != "" {
		fingerprint.WriteString(macAddr)
	}

	// Try to get CPU info (Windows specific)
	if runtime.GOOS == "windows" {
		if cpuInfo := getWindowsCPUInfo(); cpuInfo != "" {
			fingerprint.WriteString(cpuInfo)
		}
	}

	// Try to get system UUID (Windows)
	if runtime.GOOS == "windows" {
		if systemUUID := getWindowsSystemUUID(); systemUUID != "" {
			fingerprint.WriteString(systemUUID)
		}
	}

	// Use SHA256 instead of MD5 for better security
	h := sha256.New()
	h.Write([]byte(fingerprint.String()))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// Return first 24 characters for better uniqueness (was 16)
	return hash[:24], nil
}

// getMACAddress gets the MAC address of the first network interface
func getMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			macAddr := iface.HardwareAddr.String()
			if macAddr != "" && macAddr != "00:00:00:00:00:00" {
				return macAddr
			}
		}
	}
	return ""
}

// getWindowsCPUInfo gets CPU information on Windows
func getWindowsCPUInfo() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("wmic", "cpu", "get", "ProcessorId", "/value")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ProcessorId=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "ProcessorId="))
		}
	}
	return ""
}

// getWindowsSystemUUID gets system UUID on Windows
func getWindowsSystemUUID() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("wmic", "csproduct", "get", "UUID", "/value")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "UUID=") {
			uuid := strings.TrimSpace(strings.TrimPrefix(line, "UUID="))
			// Filter out common invalid UUIDs
			if uuid != "" && uuid != "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF" {
				return uuid
			}
		}
	}
	return ""
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

	// Check if license status changed to revoked or invalid
	if sheetLicense.Status == "Revoked" {
		return fmt.Errorf("license has been revoked - please contact support")
	}

	if sheetLicense.Status != "Activated" && sheetLicense.Status != "Active" {
		return fmt.Errorf("license is no longer active - status: %s", sheetLicense.Status)
	}

	// Check if expiry date changed (e.g., license was extended)
	if !sheetLicense.ExpiryDate.IsZero() && !sheetLicense.ExpiryDate.Equal(license.ExpiryDate) {
		// License expiry date updated
		license.ExpiryDate = sheetLicense.ExpiryDate
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
		// Failed to update last connected time, but continue operation
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

// TestNetworkConnectivity tests connectivity to Google Sheets API
func (m *Manager) TestNetworkConnectivity() error {
	fmt.Printf("🔍 Testing network connectivity...\n")

	// Test basic internet connectivity
	fmt.Printf("   • Testing basic internet connectivity...")
	resp, err := http.Get("https://www.google.com")
	if err != nil {
		fmt.Printf(" ❌ FAILED\n")
		return fmt.Errorf("no internet connection: %v", err)
	}
	resp.Body.Close()
	fmt.Printf(" ✅ OK\n")

	// Test Google APIs connectivity
	fmt.Printf("   • Testing Google APIs connectivity...")
	resp, err = http.Get("https://sheets.googleapis.com")
	if err != nil {
		fmt.Printf(" ❌ FAILED\n")
		return fmt.Errorf("cannot reach Google APIs: %v", err)
	}
	resp.Body.Close()
	fmt.Printf(" ✅ OK\n")

	// Test Google Sheets service initialization
	fmt.Printf("   • Testing Google Sheets service...")
	if m.sheetsService == nil {
		fmt.Printf(" ❌ FAILED\n")
		return fmt.Errorf("Google Sheets service not initialized")
	}
	fmt.Printf(" ✅ OK\n")

	// Test actual Google Sheets access
	fmt.Printf("   • Testing Google Sheets access...")
	_, err = m.sheetsService.Spreadsheets.Get(m.config.SheetID).Do()
	if err != nil {
		fmt.Printf(" ❌ FAILED\n")
		return fmt.Errorf("cannot access Google Sheets: %v", err)
	}
	fmt.Printf(" ✅ OK\n")

	fmt.Printf("🎉 All connectivity tests passed!\n")
	return nil
}

// RevokeLicense revokes a license (admin operation)
func (m *Manager) RevokeLicense(licenseKey string) error {
	if licenseKey == "" {
		return fmt.Errorf("license key cannot be empty")
	}

	// Revoking license

	// Try to validate the license from Google Sheets
	licenseInfo, err := m.validateLicenseFromSheets(licenseKey)
	if err != nil {
		return fmt.Errorf("license validation failed: %v", err)
	}

	// Update license status to revoked
	licenseInfo.Status = "Revoked"
	licenseInfo.LastChecked = time.Now()

	// Update license in Google Sheets
	if err := m.updateLicenseInSheets(licenseInfo); err != nil {
		return fmt.Errorf("failed to revoke license in Google Sheets: %v", err)
	}

	// License revoked successfully
	return nil
}

// GetLicenseStatus returns detailed license status information
func (m *Manager) GetLicenseStatus() (*LicenseInfo, string, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return nil, "No License", fmt.Errorf("no local license found: %v", err)
	}

	// Calculate status based on expiry date
	daysLeft := int(time.Until(license.ExpiryDate).Hours() / 24)
	var status string

	if time.Now().After(license.ExpiryDate) {
		status = "Expired"
	} else if daysLeft <= 7 {
		status = "Critical" // 7 or fewer days
	} else if daysLeft <= 30 {
		status = "Warning" // 8-30 days
	} else {
		status = "Active" // More than 30 days
	}

	return &license, status, nil
}

// CheckRenewalStatus checks if license needs renewal and returns detailed info
func (m *Manager) CheckRenewalStatus() (*RenewalInfo, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return &RenewalInfo{
			Status:       "No License",
			Message:      "No license found. Please activate a license.",
			NeedsRenewal: true,
			IsExpired:    true,
		}, fmt.Errorf("no local license found: %v", err)
	}

	daysLeft := int(time.Until(license.ExpiryDate).Hours() / 24)
	renewalInfo := &RenewalInfo{DaysLeft: daysLeft}

	if time.Now().After(license.ExpiryDate) {
		renewalInfo.Status = "Expired"
		renewalInfo.Message = fmt.Sprintf("License expired %d days ago. Please renew immediately.", -daysLeft)
		renewalInfo.NeedsRenewal = true
		renewalInfo.IsExpired = true
	} else if daysLeft <= 7 {
		renewalInfo.Status = "Critical"
		renewalInfo.Message = fmt.Sprintf("License expires in %d days! Please renew soon to avoid interruption.", daysLeft)
		renewalInfo.NeedsRenewal = true
		renewalInfo.IsExpired = false
	} else if daysLeft <= 30 {
		renewalInfo.Status = "Warning"
		renewalInfo.Message = fmt.Sprintf("License expires in %d days. Consider renewing soon.", daysLeft)
		renewalInfo.NeedsRenewal = true
		renewalInfo.IsExpired = false
	} else {
		renewalInfo.Status = "Active"
		renewalInfo.Message = fmt.Sprintf("License is active with %d days remaining.", daysLeft)
		renewalInfo.NeedsRenewal = false
		renewalInfo.IsExpired = false
	}

	return renewalInfo, nil
}

// ShowRenewalNotification displays renewal notification if needed
func (m *Manager) ShowRenewalNotification() error {
	renewalInfo, err := m.CheckRenewalStatus()
	if err != nil {
		return err
	}

	if renewalInfo.NeedsRenewal {
		fmt.Printf("\n")
		fmt.Printf("┌─────────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│                    LICENSE RENEWAL NOTICE                      │\n")
		fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")

		switch renewalInfo.Status {
		case "Expired":
			fmt.Printf("│ ❌ STATUS: EXPIRED                                             │\n")
		case "Critical":
			fmt.Printf("│ 🚨 STATUS: CRITICAL - EXPIRES SOON                            │\n")
		case "Warning":
			fmt.Printf("│ ⚠️  STATUS: WARNING - RENEWAL RECOMMENDED                      │\n")
		}

		fmt.Printf("│                                                                 │\n")
		fmt.Printf("│ %s\n", fmt.Sprintf("%-63s", renewalInfo.Message)+"│")
		fmt.Printf("│                                                                 │\n")

		if renewalInfo.IsExpired {
			fmt.Printf("│ 🔒 Application functionality is limited until renewal.         │\n")
		} else {
			fmt.Printf("│ 📧 Contact support for license renewal options.               │\n")
		}

		fmt.Printf("│                                                                 │\n")
		fmt.Printf("│ 📞 Support: contact your license provider                      │\n")
		fmt.Printf("│ 🌐 Web: Use license activation interface for new licenses     │\n")
		fmt.Printf("└─────────────────────────────────────────────────────────────────┘\n")
		fmt.Printf("\n")
	}

	return nil
}

// ExtendLicense extends a license for additional time (admin operation)
func (m *Manager) ExtendLicense(licenseKey string, additionalDuration string) error {
	if licenseKey == "" {
		return fmt.Errorf("license key cannot be empty")
	}

	// Extending license

	// Try to validate the license from Google Sheets
	licenseInfo, err := m.validateLicenseFromSheets(licenseKey)
	if err != nil {
		return fmt.Errorf("license validation failed: %v", err)
	}

	// Calculate additional time
	var additionalTime time.Duration
	switch additionalDuration {
	case "1m":
		additionalTime = 30 * 24 * time.Hour // 30 days
	case "3m":
		additionalTime = 90 * 24 * time.Hour // 90 days
	case "6m":
		additionalTime = 180 * 24 * time.Hour // 180 days
	case "1y":
		additionalTime = 365 * 24 * time.Hour // 365 days
	default:
		return fmt.Errorf("invalid duration: %s (use 1m, 3m, 6m, or 1y)", additionalDuration)
	}

	// Extend the expiry date
	licenseInfo.ExpiryDate = licenseInfo.ExpiryDate.Add(additionalTime)
	licenseInfo.LastChecked = time.Now()

	// License extended successfully

	// Update license in Google Sheets
	if err := m.updateLicenseInSheets(licenseInfo); err != nil {
		return fmt.Errorf("failed to extend license in Google Sheets: %v", err)
	}

	// Extension updated in Google Sheets
	return nil
}

// ValidateWithRenewalCheck performs validation and checks for renewal needs
func (m *Manager) ValidateWithRenewalCheck() (bool, *RenewalInfo, error) {
	// First perform normal validation
	isValid, err := m.ValidateLicense()

	// Get renewal information regardless of validation result
	renewalInfo, renewalErr := m.CheckRenewalStatus()
	if renewalErr != nil {
		renewalInfo = &RenewalInfo{
			Status:       "No License",
			Message:      "No license found",
			NeedsRenewal: true,
			IsExpired:    true,
		}
	}

	// Show notification if renewal is needed
	if renewalInfo.NeedsRenewal {
		m.ShowRenewalNotification()
	}

	return isValid, renewalInfo, err
}

// TrackOperation wraps an operation with performance tracking and logging
func (m *Manager) TrackOperation(operation string, fn func() error) error {
	start := time.Now()

	// Log operation start
	if m.logger != nil {
		m.logger.Log(LogEntry{
			Level:     LogLevelDebug,
			Action:    operation + "_start",
			Result:    "Operation initiated",
			MachineID: m.machineID[:min(8, len(m.machineID))],
		})
	}

	err := fn()
	duration := time.Since(start)

	// Record performance metrics
	m.recordPerformanceMetric(operation, duration, err == nil)

	// Log operation completion
	if m.logger != nil {
		level := LogLevelInfo
		result := "Operation completed successfully"
		if err != nil {
			level = LogLevelError
			result = "Operation failed"
		}

		m.logger.Log(LogEntry{
			Level:    level,
			Action:   operation + "_complete",
			Result:   result,
			Duration: duration.Milliseconds(),
			Error: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
			MachineID: m.machineID[:min(8, len(m.machineID))],
		})
	}

	return err
}

// recordPerformanceMetric updates performance statistics
func (m *Manager) recordPerformanceMetric(operation string, duration time.Duration, success bool) {
	m.perfMutex.Lock()
	defer m.perfMutex.Unlock()

	if m.performanceData == nil {
		m.performanceData = make(map[string]*PerformanceMetrics)
	}

	metric, exists := m.performanceData[operation]
	if !exists {
		metric = &PerformanceMetrics{
			MinTime: duration,
			MaxTime: duration,
		}
		m.performanceData[operation] = metric
	}

	// Update metrics
	metric.Count++
	metric.TotalTime += duration
	metric.AverageTime = time.Duration(int64(metric.TotalTime) / metric.Count)
	metric.LastUpdated = time.Now()

	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}
	if duration < metric.MinTime {
		metric.MinTime = duration
	}

	if success {
		metric.SuccessCount++
	} else {
		metric.ErrorCount++
	}
}

// GetPerformanceMetrics returns performance statistics
func (m *Manager) GetPerformanceMetrics() map[string]*PerformanceMetrics {
	m.perfMutex.RLock()
	defer m.perfMutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*PerformanceMetrics)
	for k, v := range m.performanceData {
		result[k] = &PerformanceMetrics{
			Count:        v.Count,
			TotalTime:    v.TotalTime,
			AverageTime:  v.AverageTime,
			MaxTime:      v.MaxTime,
			MinTime:      v.MinTime,
			ErrorCount:   v.ErrorCount,
			SuccessCount: v.SuccessCount,
			LastUpdated:  v.LastUpdated,
		}
	}
	return result
}

// GetSystemStats returns comprehensive system statistics
func (m *Manager) GetSystemStats() map[string]interface{} {
	stats := map[string]interface{}{
		"performance": m.GetPerformanceMetrics(),
		"timestamp":   time.Now(),
		"machine_id":  m.machineID[:min(8, len(m.machineID))],
		"version":     "enhanced-v2.0.0",
	}

	if m.cache != nil {
		stats["cache"] = m.cache.GetStats()
	}

	if m.security != nil {
		stats["security"] = m.security.GetStats()
	}

	return stats
}

// Close properly shuts down the manager and its components
func (m *Manager) Close() error {
	var errors []string

	if m.logger != nil {
		if err := m.logger.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("logger: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to close manager: %s", strings.Join(errors, ", "))
	}

	return nil
}

// validateLicenseFromSheetsWithCache validates license with caching support
func (m *Manager) validateLicenseFromSheetsWithCache(licenseKey string) (LicenseInfo, error) {
	// Check cache first
	if m.cache != nil {
		if cachedInfo, found := m.cache.Get(licenseKey); found {
			if m.logger != nil {
				m.logger.Log(LogEntry{
					Level:      LogLevelDebug,
					Action:     "cache_hit",
					Result:     "License found in cache",
					LicenseKey: licenseKey[:min(8, len(licenseKey))],
				})
			}
			return *cachedInfo, nil
		}
	}

	// Cache miss - fetch from Google Sheets
	licenseInfo, err := m.validateLicenseFromSheets(licenseKey)
	if err != nil {
		return licenseInfo, err
	}

	// Store in cache
	if m.cache != nil {
		m.cache.Set(licenseKey, licenseInfo)
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:      LogLevelDebug,
				Action:     "cache_store",
				Result:     "License stored in cache",
				LicenseKey: licenseKey[:min(8, len(licenseKey))],
			})
		}
	}

	return licenseInfo, nil
}
