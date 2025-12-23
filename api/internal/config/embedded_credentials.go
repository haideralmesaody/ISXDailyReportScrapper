package config

import (
	"os"
	"strconv"
)

// EmbeddedCredentials contains all configuration for the application
// Values are read from environment variables with sensible defaults
type EmbeddedCredentials struct {
	// Google Apps Script Configuration
	AppsScriptURL    string
	AppsScriptSecret string
	
	// Feature Flags
	EnableScratchCardMode   bool
	EnableDeviceFingerprint bool
	
	// Security Settings
	EnableSecureMode     bool
	RequireSignature     bool
	EnableEncryption     bool
	AllowInsecureForTest bool
	
	// License System Settings
	MaxLicenseAttempts   int
	LicenseBlockDuration int // in minutes
	LicenseWindowDuration int // in minutes
}

// getEnvBool returns a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return b
}

// getEnvInt returns an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return i
}

// getEnvString returns a string environment variable with a default value
func getEnvString(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetEmbeddedCredentials returns the configuration with embedded values
// Environment variables override embedded values for development flexibility
func GetEmbeddedCredentials() *EmbeddedCredentials {
	// Embedded Apps Script URL - Updated with working deployment
	const embeddedAppsScriptURL = "https://script.google.com/macros/s/AKfycby3E1ZTnEKYCRZzALL9HYKFaNbPkJgcFaVKNPMNVuwdwIfD5kQ7HQgxNzN_6chWGoMD3w/exec"
	const embeddedAppsScriptSecret = "ISX-Pulse-S3cur3-K3y-2024-@lm3s@0dy" // Production shared secret
	
	return &EmbeddedCredentials{
		// Google Apps Script Configuration
		// Environment variables override embedded values if set
		AppsScriptURL:    getEnvString("APPS_SCRIPT_URL", embeddedAppsScriptURL),
		AppsScriptSecret: getEnvString("APPS_SCRIPT_SECRET", embeddedAppsScriptSecret),
		
		// Feature Flags
		EnableScratchCardMode:   getEnvBool("ENABLE_SCRATCH_CARD_MODE", true),
		EnableDeviceFingerprint: getEnvBool("ENABLE_DEVICE_FINGERPRINT", true),
		
		// Security Settings
		EnableSecureMode:     getEnvBool("ENABLE_SECURE_MODE", true), // Production mode
		RequireSignature:     getEnvBool("REQUIRE_SIGNATURE", true), // Enable signature verification
		EnableEncryption:     getEnvBool("ENABLE_ENCRYPTION", false),
		AllowInsecureForTest: getEnvBool("ALLOW_INSECURE_FOR_TEST", false), // Disable test mode
		
		// License System Settings
		MaxLicenseAttempts:    getEnvInt("MAX_LICENSE_ATTEMPTS", 5),
		LicenseBlockDuration:  getEnvInt("LICENSE_BLOCK_DURATION", 15),
		LicenseWindowDuration: getEnvInt("LICENSE_WINDOW_DURATION", 60),
	}
}

// Singleton instance for easy access
var embeddedCreds *EmbeddedCredentials

// GetCredentials returns the singleton embedded credentials instance
func GetCredentials() *EmbeddedCredentials {
	if embeddedCreds == nil {
		embeddedCreds = GetEmbeddedCredentials()
	}
	return embeddedCreds
}

// For backward compatibility, these functions return values from environment
// DEPRECATED: Use GetCredentials() instead for better type safety

// GetAppsScriptURL returns the Apps Script URL from environment
func GetAppsScriptURL() string {
	return GetCredentials().AppsScriptURL
}

// GetAppsScriptSecret returns the Apps Script secret from environment
func GetAppsScriptSecret() string {
	return GetCredentials().AppsScriptSecret
}