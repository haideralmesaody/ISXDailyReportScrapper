package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/scrypt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SecureCredentialsManager handles encrypted credential management with audit logging
type SecureCredentialsManager struct {
	encryptedPayload *EncryptedPayload
	appSalt          []byte
	binaryHash       string
	certPinner       *CertificatePinner
	integrityChecker *IntegrityChecker
	mutex            sync.Mutex
	lastAccess       time.Time
	accessCount      int64
	auditLogger      *slog.Logger
}

// CredentialAccessEvent represents a credential access audit event
type CredentialAccessEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	EventType     string    `json:"event_type"`     // "decrypt", "access", "clear", "error"
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	BinaryHash    string    `json:"binary_hash"`
	ProcessID     int       `json:"process_id"`
	AccessCount   int64     `json:"access_count"`
	ClientIP      string    `json:"client_ip,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
}

// Embedded encrypted credentials (will be replaced during build process)
var embeddedCredentials = `{
  "version": 1,
  "salt": "c3ql9hH+MFZ0m8DbYVbctzfBmer/XpBDbkzMtcUQSnA=",
  "nonce": "1teiLQ7fG2CmjO8s",
  "ciphertext": "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",
  "auth_tag": "u6hw1EqZLQWtAq+C9bevTg==",
  "integrity": "z03V2qiceICTkMtv5q3xbEPkx7Fhbqq4Hci/UtLQ6Vo=",
  "timestamp": 1721958000
}`

// loadEncryptedCredentials loads encrypted credentials - first tries embedded, then external file
func loadEncryptedCredentials() (*EncryptedPayload, error) {
	// First, check if we should use an external file (override)
	if externalFile := os.Getenv("ISX_EXTERNAL_CREDENTIALS"); externalFile != "" {
		// Try to load from external file
		possiblePaths := []string{
			"encrypted_credentials.dat",
			"./encrypted_credentials.dat",
			"../encrypted_credentials.dat",
			os.Getenv("ISX_CREDENTIALS_FILE"),
		}

		for _, path := range possiblePaths {
			if path == "" {
				continue
			}
			if credentialsData, err := os.ReadFile(path); err == nil {
				var payload EncryptedPayload
				if err := json.Unmarshal(credentialsData, &payload); err == nil {
					return &payload, nil
				}
			}
		}
	}

	// Use embedded credentials by default
	var payload EncryptedPayload
	if err := json.Unmarshal([]byte(embeddedCredentials), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse embedded credentials: %v", err)
	}

	return &payload, nil
}

// Application constants for security
const (
	ApplicationSalt = "ISX-Daily-Reports-Scrapper-v2.0-Salt-2025"
	MaxAccessCount  = 1000 // Maximum credential accesses before requiring restart
	AccessTimeout   = 1 * time.Hour // Maximum time credentials can remain in memory
)

// Expected binary hash (will be replaced during build process)
var expectedBinaryHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// NewSecureCredentialsManager creates a new secure credentials manager
func NewSecureCredentialsManager() (*SecureCredentialsManager, error) {
	// Initialize audit logger
	auditLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	// Load credentials (embedded or external)
	payload, err := loadEncryptedCredentials()
	if err != nil {
		auditLogger.Error("Failed to load credentials",
			slog.String("error", err.Error()),
			slog.String("event_type", "credential_load_error"),
		)
		return nil, fmt.Errorf("failed to load credentials: %v", err)
	}

	// Initialize certificate pinner with Google APIs pins
	certPinner := NewCertificatePinner(DefaultPinningConfig())

	// Initialize integrity checker
	integrityChecker := NewIntegrityChecker(expectedBinaryHash)

	manager := &SecureCredentialsManager{
		encryptedPayload: payload,
		appSalt:          []byte(ApplicationSalt),
		binaryHash:       expectedBinaryHash,
		certPinner:       certPinner,
		integrityChecker: integrityChecker,
		auditLogger:      auditLogger,
	}

	// Log initialization
	manager.logAuditEvent("initialization", true, "", nil)

	return manager, nil
}

// GetSecureCredentials decrypts and returns credentials with full security validation
func (scm *SecureCredentialsManager) GetSecureCredentials(ctx context.Context) (*SecureCredentials, error) {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	// Check access limits
	if err := scm.checkAccessLimits(); err != nil {
		scm.logAuditEvent("access_limit_exceeded", false, err.Error(), ctx)
		return nil, err
	}

	// Verify binary integrity first
	integrityResult, err := scm.integrityChecker.VerifyBinaryIntegrity()
	if err != nil {
		scm.logAuditEvent("integrity_check_failed", false, err.Error(), ctx)
		return nil, fmt.Errorf("binary integrity verification failed: %v", err)
	}

	if !integrityResult.IsValid {
		scm.logAuditEvent("integrity_verification_failed", false, integrityResult.ErrorMessage, ctx)
		return nil, fmt.Errorf("binary integrity verification failed: %s", integrityResult.ErrorMessage)
	}

	// Decrypt credentials using the security package
	credentials, err := DecryptCredentials(scm.encryptedPayload, scm.appSalt, DefaultEncryptionConfig())
	if err != nil {
		scm.logAuditEvent("decryption_failed", false, err.Error(), ctx)
		return nil, fmt.Errorf("credential decryption failed: %v", err)
	}

	// Update access tracking
	scm.accessCount++
	scm.lastAccess = time.Now()

	// Log successful access
	scm.logAuditEvent("credentials_accessed", true, "", ctx)

	return credentials, nil
}

// CreateSecureSheetsService creates a Google Sheets service with encrypted credentials and certificate pinning
func (scm *SecureCredentialsManager) CreateSecureSheetsService(ctx context.Context) (*sheets.Service, func(), error) {
	// Get decrypted credentials
	credentials, err := scm.GetSecureCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get credentials: %v", err)
	}

	// Create cleanup function to clear credentials from memory
	cleanup := func() {
		credentials.Clear()
		scm.logAuditEvent("credentials_cleared", true, "", ctx)
	}

	// Create HTTP client with certificate pinning
	httpClient := scm.certPinner.CreateSecureHTTPClient(DefaultPinningConfig())

	// Create Google API credentials option
	credentialsOption := option.WithCredentialsJSON(credentials.Data())
	httpClientOption := option.WithHTTPClient(httpClient)

	// Initialize sheets service with secure options
	sheetsService, err := sheets.NewService(ctx, credentialsOption, httpClientOption)
	if err != nil {
		cleanup() // Clean up credentials on error
		scm.logAuditEvent("sheets_service_creation_failed", false, err.Error(), ctx)
		return nil, nil, fmt.Errorf("failed to create sheets service: %v", err)
	}

	scm.logAuditEvent("sheets_service_created", true, "", ctx)
	return sheetsService, cleanup, nil
}

// ValidateSecurityConfiguration performs comprehensive security validation
func (scm *SecureCredentialsManager) ValidateSecurityConfiguration() error {
	var errors []string

	// Validate encryption configuration
	config := DefaultEncryptionConfig()
	if err := ValidateEncryptionConfig(config); err != nil {
		errors = append(errors, fmt.Sprintf("encryption config invalid: %v", err))
	}

	// Validate integrity configuration
	if err := ValidateIntegrityConfig(scm.binaryHash); err != nil {
		errors = append(errors, fmt.Sprintf("integrity config invalid: %v", err))
	}

	// Validate certificate pinning
	if scm.certPinner == nil {
		errors = append(errors, "certificate pinner not initialized")
	}

	// Test Google APIs connectivity with certificate pinning
	if err := scm.certPinner.ValidateGoogleAPIsConnectivity(); err != nil {
		errors = append(errors, fmt.Sprintf("Google APIs connectivity failed: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("security validation failed: %v", errors)
	}

	scm.logAuditEvent("security_validation_passed", true, "", nil)
	return nil
}

// checkAccessLimits verifies access count and time limits
func (scm *SecureCredentialsManager) checkAccessLimits() error {
	// Check maximum access count
	if scm.accessCount >= MaxAccessCount {
		return fmt.Errorf("maximum credential access count exceeded (%d)", MaxAccessCount)
	}

	// Check access timeout
	if !scm.lastAccess.IsZero() && time.Since(scm.lastAccess) > AccessTimeout {
		return fmt.Errorf("credential access timeout exceeded")
	}

	return nil
}

// GetCredentials decrypts and returns the Google service account credentials
func (scm *SecureCredentialsManager) GetCredentials(ctx context.Context) ([]byte, error) {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	// Check access limits
	if err := scm.checkAccessLimits(); err != nil {
		scm.logAuditEvent("access_limit_exceeded", false, err.Error(), ctx)
		return nil, err
	}

	// Perform binary integrity check
	_, err := scm.integrityChecker.VerifyBinaryIntegrity()
	if err != nil {
		scm.logAuditEvent("integrity_check_failed", false, err.Error(), ctx)
		return nil, fmt.Errorf("binary integrity verification failed: %v", err)
	}

	// Decrypt credentials
	credentialsJSON, err := scm.decryptCredentials()
	if err != nil {
		scm.logAuditEvent("decryption_failed", false, err.Error(), ctx)
		return nil, fmt.Errorf("failed to decrypt credentials: %v", err)
	}

	// Update access tracking
	scm.accessCount++
	scm.lastAccess = time.Now()

	// Log successful access
	scm.logAuditEvent("credentials_accessed", true, "", ctx)

	return credentialsJSON, nil
}

// decryptCredentials performs the actual credential decryption
func (scm *SecureCredentialsManager) decryptCredentials() ([]byte, error) {
	// Decrypt the credentials using the encryption module
	secureCredentials, err := DecryptCredentials(scm.encryptedPayload, []byte(ApplicationSalt), nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	defer secureCredentials.Clear() // Ensure cleanup

	// Get the data and validate it's valid JSON
	decryptedData := secureCredentials.Data()
	var testJSON interface{}
	if err := json.Unmarshal(decryptedData, &testJSON); err != nil {
		return nil, fmt.Errorf("decrypted data is not valid JSON: %v", err)
	}

	// Return a copy of the data since secureCredentials will be cleared
	result := make([]byte, len(decryptedData))
	copy(result, decryptedData)
	return result, nil
}

// deriveKey derives an encryption key using SCRYPT from input data and salt
func deriveKey(inputData, salt []byte, keyLen int) ([]byte, error) {
	// Use secure SCRYPT parameters (OWASP recommended)
	return scrypt.Key(inputData, salt, 32768, 8, 1, keyLen)
}

// logAuditEvent logs credential access events for security auditing
func (scm *SecureCredentialsManager) logAuditEvent(eventType string, success bool, errorMessage string, ctx context.Context) {
	event := CredentialAccessEvent{
		Timestamp:    time.Now(),
		EventType:    eventType,
		Success:      success,
		ErrorMessage: errorMessage,
		BinaryHash:   scm.binaryHash[:16], // First 16 chars for audit
		ProcessID:    os.Getpid(),
		AccessCount:  scm.accessCount,
	}

	// Extract request context information if available
	if ctx != nil {
		if userAgent := ctx.Value("user-agent"); userAgent != nil {
			if ua, ok := userAgent.(string); ok {
				event.UserAgent = ua
			}
		}
		if clientIP := ctx.Value("client-ip"); clientIP != nil {
			if ip, ok := clientIP.(string); ok {
				event.ClientIP = ip
			}
		}
	}

	// Log event using structured logging
	logLevel := slog.LevelInfo
	if !success {
		logLevel = slog.LevelError
	}

	scm.auditLogger.Log(context.Background(), logLevel, "Credential access event",
		slog.String("event_type", event.EventType),
		slog.Bool("success", event.Success),
		slog.String("error_message", event.ErrorMessage),
		slog.String("binary_hash_prefix", event.BinaryHash),
		slog.Int("process_id", event.ProcessID),
		slog.Int64("access_count", event.AccessCount),
		slog.String("client_ip", event.ClientIP),
		slog.String("user_agent", event.UserAgent),
		slog.Time("timestamp", event.Timestamp),
	)
}

// GetSecurityMetrics returns security-related metrics for monitoring
func (scm *SecureCredentialsManager) GetSecurityMetrics() map[string]interface{} {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	return map[string]interface{}{
		"access_count":          scm.accessCount,
		"last_access":           scm.lastAccess,
		"max_access_count":      MaxAccessCount,
		"access_timeout":        AccessTimeout,
		"binary_hash_prefix":    scm.binaryHash[:16],
		"encryption_version":    scm.encryptedPayload.Version,
		"certificate_pins":      len(scm.certPinner.GetPinnedHashes()),
		"security_initialized":  true,
	}
}

// RotateCredentials updates embedded credentials (for future use)
func (scm *SecureCredentialsManager) RotateCredentials(newPayload *EncryptedPayload) error {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	// Validate new payload
	if newPayload == nil {
		return fmt.Errorf("new payload cannot be nil")
	}

	if newPayload.Version != 1 {
		return fmt.Errorf("unsupported payload version: %d", newPayload.Version)
	}

	// Test decryption with new payload
	testCredentials, err := DecryptCredentials(newPayload, scm.appSalt, DefaultEncryptionConfig())
	if err != nil {
		scm.logAuditEvent("rotation_test_failed", false, err.Error(), nil)
		return fmt.Errorf("credential rotation test failed: %v", err)
	}
	testCredentials.Clear()

	// Update payload
	scm.encryptedPayload = newPayload
	scm.accessCount = 0 // Reset access count after rotation
	scm.lastAccess = time.Time{}

	scm.logAuditEvent("credentials_rotated", true, "", nil)
	return nil
}

// Close performs cleanup and final audit logging
func (scm *SecureCredentialsManager) Close() error {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	scm.logAuditEvent("manager_shutdown", true, "", nil)

	// Clear sensitive data
	if scm.encryptedPayload != nil {
		scm.encryptedPayload = nil
	}

	if scm.appSalt != nil {
		for i := range scm.appSalt {
			scm.appSalt[i] = 0
		}
		scm.appSalt = nil
	}

	return nil
}

// GenerateApplicationHash generates a hash for the current application binary
func GenerateApplicationHash() (string, error) {
	return GenerateBinaryHash()
}

// ValidateApplicationIntegrity validates the current application against expected hash
func ValidateApplicationIntegrity(expectedHash string) error {
	checker := NewIntegrityChecker(expectedHash)
	result, err := checker.VerifyBinaryIntegrity()
	if err != nil {
		return fmt.Errorf("integrity verification failed: %v", err)
	}

	if !result.IsValid {
		return fmt.Errorf("binary integrity verification failed: %s", result.ErrorMessage)
	}

	return nil
}