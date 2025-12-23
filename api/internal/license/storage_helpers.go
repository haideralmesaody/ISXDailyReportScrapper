package license

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/security"
)

// writeAuditToFile appends audit entry to JSON file.
func (m *Manager) writeAuditToFile(audit LicenseAudit, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}

	data, err := json.Marshal(audit)
	if err != nil {
		return fmt.Errorf("marshal audit: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit: %w", err)
	}

	return nil
}

// saveLicenseLocal saves license to local file (encrypted if configured).
func (m *Manager) saveLicenseLocal(license LicenseInfo) error {
	ctx := context.Background()
	absPath, _ := filepath.Abs(m.licenseFile)
	dir := filepath.Dir(absPath)
	dirInfo, dirErr := os.Stat(dir)
	var dirDetails string
	if dirErr == nil {
		dirDetails = fmt.Sprintf("exists=%t, writable=%t", dirInfo.IsDir(), isWritable(dir))
	} else {
		dirDetails = fmt.Sprintf("error=%v", dirErr)
	}

	m.logInfo(ctx, "license_save_attempt", "Attempting to save license file",
		slog.String("configured_path", m.licenseFile),
		slog.String("absolute_path", absPath),
		slog.String("directory", dir),
		slog.String("directory_details", dirDetails),
		slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
		slog.String("working_dir", m.getWorkingDir()),
	)

	var saveErr error

	creds := config.GetCredentials()
	if creds != nil && creds.EnableEncryption {
		m.logInfo(ctx, "license_save", "Encryption is enabled, using encrypted save",
			slog.Bool("encryption_enabled", true),
		)

		encryptedMgr, err := NewEncryptedManager(m)
		if err != nil {
			m.logError(ctx, "license_save", "Failed to create encrypted manager",
				slog.String("error", err.Error()),
			)
			saveErr = m.saveLicenseLocalPlain(license)
		} else {
			saveErr = encryptedMgr.saveLicenseLocalEncrypted(license)
		}
	} else {
		saveErr = m.saveLicenseLocalPlain(license)
	}

	if saveErr != nil {
		return saveErr
	}

	m.invalidateValidationCache("license_saved")
	return nil
}

// saveLicenseLocalPlain saves license as plain JSON.
func (m *Manager) saveLicenseLocalPlain(license LicenseInfo) error {
	ctx := context.Background()

	data, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		m.logError(ctx, "license_save", "Failed to marshal license data",
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := os.WriteFile(m.licenseFile, data, 0600); err != nil {
		m.logError(ctx, "license_save", "Failed to write license file",
			slog.String("path", m.licenseFile),
			slog.String("error", err.Error()),
		)
		return err
	}

	m.logInfo(ctx, "license_save", "License saved successfully (plain)",
		slog.String("path", m.licenseFile),
		slog.Int("size_bytes", len(data)),
		slog.Bool("encrypted", false),
	)

	return nil
}

// loadLicenseLocal loads license from local file.
func (m *Manager) loadLicenseLocal() (LicenseInfo, error) {
	var license LicenseInfo
	ctx := context.Background()

	absPath, _ := filepath.Abs(m.licenseFile)
	fileInfo, statErr := os.Stat(m.licenseFile)
	var fileDetails string
	if statErr == nil {
		fileDetails = fmt.Sprintf("size=%d, mode=%s, modtime=%s",
			fileInfo.Size(), fileInfo.Mode(), fileInfo.ModTime().Format(time.RFC3339))
	} else {
		fileDetails = fmt.Sprintf("stat_error=%v", statErr)
	}

	m.logInfo(ctx, "license_load_attempt", "Attempting to load license file",
		slog.String("configured_path", m.licenseFile),
		slog.String("absolute_path", absPath),
		slog.String("working_dir", m.getWorkingDir()),
		slog.Bool("file_exists", config.FileExists(m.licenseFile)),
		slog.String("file_details", fileDetails),
	)

	data, err := os.ReadFile(m.licenseFile)
	if err != nil {
		m.logDebug(ctx, "license_load", "Failed to read license file",
			slog.String("path", m.licenseFile),
			slog.String("error", err.Error()),
		)
		if errors.Is(err, os.ErrNotExist) {
			return license, fmt.Errorf("no such file: %w", err)
		}
		return license, err
	}

	creds := config.GetCredentials()
	if creds != nil && creds.EnableEncryption {
		encryptedMgr, err := NewEncryptedManager(m)
		if err == nil {
			if encryptedLicense, err := encryptedMgr.LoadLicense(); err == nil {
				m.logDebug(ctx, "license_load", "License loaded successfully (encrypted)",
					slog.String("path", m.licenseFile),
					slog.String("license_key_prefix", encryptedLicense.LicenseKey[:min(8, len(encryptedLicense.LicenseKey))]),
					slog.String("status", encryptedLicense.Status),
					slog.Bool("encrypted", true),
				)
				return encryptedLicense, nil
			}
		}
	}

	var encryptedPayload security.EncryptedPayload
	if err := json.Unmarshal(data, &encryptedPayload); err == nil && encryptedPayload.Version > 0 {
		m.logWarn(ctx, "license_load", "Found encrypted license but encryption is disabled or failed",
			slog.String("path", m.licenseFile),
			slog.Bool("encryption_enabled", creds != nil && creds.EnableEncryption),
		)
		encryptedMgr, err := NewEncryptedManager(m)
		if err == nil {
			if encryptedLicense, err := encryptedMgr.LoadLicense(); err == nil {
				return encryptedLicense, nil
			}
		}
		return license, fmt.Errorf("license is encrypted but decryption failed")
	}

	if err := json.Unmarshal(data, &license); err != nil {
		m.logError(ctx, "license_load", "Failed to unmarshal license data",
			slog.String("path", m.licenseFile),
			slog.String("error", err.Error()),
		)
		return license, err
	}

	if license.LicenseKey == "" {
		m.logError(ctx, "license_load", "Corrupted license: missing license key",
			slog.String("path", m.licenseFile))
		return license, fmt.Errorf("corrupted license file: missing license key")
	}
	if license.ExpiryDate.IsZero() {
		m.logError(ctx, "license_load", "Corrupted license: invalid expiry date",
			slog.String("path", m.licenseFile))
		return license, fmt.Errorf("corrupted license file: invalid expiry date")
	}

	if license.ActivationID == "" {
		if id, err := generateShortID(10); err == nil {
			license.ActivationID = id
		}
	}

	m.logDebug(ctx, "license_load", "License loaded successfully",
		slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
		slog.String("status", license.Status),
		slog.Bool("encrypted", false),
	)

	return license, nil
}

func generateShortID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:n], nil
}
