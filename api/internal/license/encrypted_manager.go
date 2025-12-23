package license

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/security"
)

// EncryptedLicenseManager wraps the base manager with encryption capabilities
type EncryptedLicenseManager struct {
	*Manager
	encryptionConfig *security.EncryptionConfig
	appSalt          []byte
}

// NewEncryptedManager creates a new encrypted license manager
func NewEncryptedManager(baseManager *Manager) (*EncryptedLicenseManager, error) {
	// Get encryption configuration
	encryptionConfig := security.DefaultEncryptionConfig()

	// Generate or load application salt
	appSalt, err := generateApplicationSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate application salt: %w", err)
	}

	return &EncryptedLicenseManager{
		Manager:          baseManager,
		encryptionConfig: encryptionConfig,
		appSalt:          appSalt,
	}, nil
}

// generateApplicationSalt generates a persistent application-wide salt
func generateApplicationSalt() ([]byte, error) {
	// Use a fixed salt for the application (could be loaded from config)
	// This ensures consistency across restarts
	appSaltStr := "ISX-Pulse-2024-License-Encryption-Salt"
	return []byte(appSaltStr), nil
}

// generateSecureSalt generates a cryptographically secure salt for encryption
func (m *EncryptedLicenseManager) generateSecureSalt() ([]byte, error) {
	// OWASP recommended: minimum 32 bytes for salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate secure salt: %w", err)
	}
	return salt, nil
}

// saveLicenseLocalEncrypted saves an encrypted license
func (m *EncryptedLicenseManager) saveLicenseLocalEncrypted(license LicenseInfo) error {
	ctx := context.Background()

	// Log encryption attempt
	m.logInfo(ctx, "license_encryption", "Starting license encryption",
		slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
		slog.String("encryption_method", "AES-256-GCM"),
		slog.String("kdf", "scrypt"),
	)

	// Convert license to JSON
	plaintext, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal license: %w", err)
	}

	// Ensure app salt is at least 16 bytes (required by encryption)
	if len(m.appSalt) < 16 {
		// Pad the app salt if it's too short
		paddedSalt := make([]byte, 16)
		copy(paddedSalt, m.appSalt)
		m.appSalt = paddedSalt
		m.logWarn(ctx, "license_encryption", "App salt was padded to meet minimum requirements",
			slog.Int("original_size", len(m.appSalt)),
			slog.Int("padded_size", 16),
		)
	}

	// Encrypt with proper configuration
	encrypted, err := security.EncryptCredentials(plaintext, m.appSalt, m.encryptionConfig)
	if err != nil {
		return fmt.Errorf("failed to encrypt license: %w", err)
	}

	// Convert encrypted payload to JSON for storage
	encryptedJSON, err := json.Marshal(encrypted)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted data: %w", err)
	}

	// Save encrypted data
	if err := os.WriteFile(m.licenseFile, encryptedJSON, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted license: %w", err)
	}

	m.logInfo(ctx, "license_encryption", "License encrypted and saved successfully",
		slog.String("license_file", m.licenseFile),
		slog.Int("encrypted_size", len(encryptedJSON)),
		slog.Bool("encryption_enabled", true),
	)

	return nil
}

// loadLicenseLocalEncrypted loads and decrypts an encrypted license
func (m *EncryptedLicenseManager) loadLicenseLocalEncrypted() (LicenseInfo, error) {
	var license LicenseInfo
	ctx := context.Background()

	// Read encrypted file
	encryptedData, err := os.ReadFile(m.licenseFile)
	if err != nil {
		return license, fmt.Errorf("failed to read license file: %w", err)
	}

	// Try to parse as encrypted payload
	var payload security.EncryptedPayload
	if err := json.Unmarshal(encryptedData, &payload); err != nil {
		// Maybe it's a plain license (backward compatibility)
		if err := json.Unmarshal(encryptedData, &license); err == nil {
			m.logWarn(ctx, "license_load", "Loaded plain license file (not encrypted)",
				slog.String("license_file", m.licenseFile),
			)
			return license, nil
		}
		return license, fmt.Errorf("failed to parse license file: %w", err)
	}

	// Ensure app salt is at least 16 bytes
	if len(m.appSalt) < 16 {
		paddedSalt := make([]byte, 16)
		copy(paddedSalt, m.appSalt)
		m.appSalt = paddedSalt
	}

	// Decrypt the payload
	decrypted, err := security.DecryptCredentials(&payload, m.appSalt, m.encryptionConfig)
	if err != nil {
		return license, fmt.Errorf("failed to decrypt license: %w", err)
	}

	// Parse decrypted data
	if err := json.Unmarshal(decrypted.Data(), &license); err != nil {
		// Clean up secure credentials
		decrypted.Clear()
		return license, fmt.Errorf("failed to parse decrypted license: %w", err)
	}

	// Clean up secure credentials
	decrypted.Clear()

	m.logInfo(ctx, "license_load", "License decrypted and loaded successfully",
		slog.String("license_file", m.licenseFile),
		slog.Bool("encryption_enabled", true),
	)

	return license, nil
}

// SaveLicense saves a license with encryption if enabled
func (m *EncryptedLicenseManager) SaveLicense(license LicenseInfo) error {
	creds := config.GetCredentials()

	if creds.EnableEncryption {
		return m.saveLicenseLocalEncrypted(license)
	}

	// Fall back to plain save
	return m.saveLicenseLocal(license)
}

// LoadLicense loads a license with decryption if needed
func (m *EncryptedLicenseManager) LoadLicense() (LicenseInfo, error) {
	creds := config.GetCredentials()

	// First, check if file exists
	if _, err := os.Stat(m.licenseFile); os.IsNotExist(err) {
		return LicenseInfo{}, fmt.Errorf("license file not found: %s", m.licenseFile)
	}

	// Try to load with decryption if encryption is enabled
	if creds.EnableEncryption {
		return m.loadLicenseLocalEncrypted()
	}

	// Try to load as encrypted first (in case it was encrypted before)
	if license, err := m.loadLicenseLocalEncrypted(); err == nil {
		return license, nil
	}

	// Fall back to plain load
	return m.loadLicenseLocal()
}

// min function is already defined in manager.go
