package config

import "time"

// Application constants - all hardcoded values for the ISX Pulse system
const (
	// Application Info
	AppName    = "ISX Pulse"
	AppVersion = "3.0.0"
	AppVendor  = "Iraqi Investor"
	
	// License System Constants
	LicenseFileName       = "license.dat"
	LicenseFileBackup     = "license_backup.dat"
	ScratchCardPrefix     = "ISX-"
	ScratchCardLength     = 23 // ISX-XXXX-XXXX-XXXX-XXXX
	StandardLicenseMinLen = 9  // ISX1MXXXX minimum
	
	// Scratch Card Format Validation
	ScratchCardPattern = "^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$"
	
	// Standard License Prefixes
	LicensePrefix1Month  = "ISX1M"
	LicensePrefix3Months = "ISX3M"
	LicensePrefix6Months = "ISX6M"
	LicensePrefix1Year   = "ISX1Y"
	
	// Security Constants
	MaxLoginAttempts     = 5
	LoginBlockDuration   = 15 * time.Minute
	SessionTimeout       = 24 * time.Hour
	TokenExpiryDuration  = 7 * 24 * time.Hour
	
	// Rate Limiting
	DefaultRateLimit      = 100 // requests per minute
	DefaultBurstSize      = 50
	LicenseCheckRateLimit = 10 // license checks per minute
	
	// Network Timeouts
	DefaultHTTPTimeout   = 30 * time.Second
	AppsScriptTimeout    = 45 * time.Second
	LicenseCheckTimeout  = 10 * time.Second
	WebSocketPingPeriod  = 30 * time.Second
	WebSocketPongWait    = 60 * time.Second
	
	// File Paths (relative to executable)
	DefaultDataDir      = "data"
	DefaultLogsDir      = "logs"
	DefaultWebDir       = "web"
	DefaultDownloadsDir = "data/downloads"
	DefaultReportsDir   = "data/reports"
	
	// Cache Settings
	LicenseCacheDuration = 5 * time.Minute
	DataCacheDuration    = 15 * time.Minute
	ReportCacheDuration  = 1 * time.Hour
	
	// Operation Timeouts
	DefaultOperationTimeout = 2 * time.Hour
	ScraperTimeout          = 30 * time.Minute
	ProcessorTimeout        = 1 * time.Hour
	ReportGenerationTimeout = 15 * time.Minute
	
	// WebSocket Buffer Sizes
	WebSocketReadBufferSize  = 1024
	WebSocketWriteBufferSize = 1024
	
	// Log Settings
	DefaultLogLevel      = "info"
	DefaultLogFormat     = "json"
	MaxLogFileSize       = 100 * 1024 * 1024 // 100MB
	MaxLogFileAge        = 30                // days
	MaxLogFileBackups    = 10
	
	// ISX Data Processing
	ISXDailyReportPattern   = "ISX_Daily_Price_.*\\.xlsx?"
	ISXWeeklyReportPattern  = "ISX_Weekly_.*\\.xlsx?"
	ISXMonthlyReportPattern = "ISX_Monthly_.*\\.xlsx?"
	
	// Error Messages
	ErrLicenseNotActivated = "License not activated. Please activate a license to access Iraqi Investor features."
	ErrLicenseExpired      = "License has expired. Please renew your license to continue."
	ErrLicenseInvalid      = "Invalid license key format. Expected: ISX-XXXX-XXXX-XXXX-XXXX or ISX1M/3M/6M/1Y followed by alphanumeric code."
	ErrDeviceMismatch      = "License is registered to a different device."
	ErrNetworkError        = "Network error. Please check your internet connection."
	
	// Success Messages
	MsgLicenseActivated = "License activated successfully. You can now access all Iraqi Investor features."
	MsgLicenseValid     = "License is valid and active."
	MsgOperationSuccess = "Operation completed successfully."
)

// Feature Flags - compile-time configuration
const (
	// Core Features
	FeatureScratchCardEnabled   = true
	FeatureDeviceFingerprintEnabled = true
	FeatureSecureModeEnabled    = true
	FeatureWebSocketEnabled      = true
	FeatureMetricsEnabled        = true
	FeatureHealthCheckEnabled    = true
	
	// Security Features
	FeatureHMACSigningEnabled   = true
	FeatureRateLimitingEnabled   = true
	FeatureCertPinningEnabled    = false // Disabled for local development
	FeatureAntiDebugEnabled      = false // Disabled for development
	
	// Development Features
	FeatureDebugLoggingEnabled  = false
	FeatureVerboseModeEnabled   = false
	FeatureMockDataEnabled      = false
)

// URLs and Endpoints (all embedded)
const (
	// Google Apps Script Endpoints
	GoogleAppsScriptBaseURL = "https://script.google.com"
	GoogleAppsScriptDomain  = "script.google.com"
	
	// ISX Data Sources
	ISXWebsiteURL   = "https://www.isx-iq.net"
	ISXDataEndpoint = "https://www.isx-iq.net/isxportal/portal/sectorProfileView.html"
	
	// API Endpoints (internal)
	APIBasePath        = "/api/v1"
	LicenseEndpoint    = "/api/v1/license"
	OperationsEndpoint = "/api/v1/operations"
	DataEndpoint       = "/api/v1/data"
	ReportsEndpoint    = "/api/v1/reports"
	HealthEndpoint     = "/health"
	MetricsEndpoint    = "/metrics"
	
	// WebSocket Endpoints
	WebSocketEndpoint = "/ws"
	WebSocketHub      = "/ws/hub"
)

// GetFeatureFlag returns the value of a feature flag
func GetFeatureFlag(flag string) bool {
	switch flag {
	case "scratch_card":
		return FeatureScratchCardEnabled
	case "device_fingerprint":
		return FeatureDeviceFingerprintEnabled
	case "secure_mode":
		return FeatureSecureModeEnabled
	case "websocket":
		return FeatureWebSocketEnabled
	case "metrics":
		return FeatureMetricsEnabled
	case "health_check":
		return FeatureHealthCheckEnabled
	case "hmac_signing":
		return FeatureHMACSigningEnabled
	case "rate_limiting":
		return FeatureRateLimitingEnabled
	case "cert_pinning":
		return FeatureCertPinningEnabled
	case "anti_debug":
		return FeatureAntiDebugEnabled
	case "debug_logging":
		return FeatureDebugLoggingEnabled
	case "verbose_mode":
		return FeatureVerboseModeEnabled
	case "mock_data":
		return FeatureMockDataEnabled
	default:
		return false
	}
}