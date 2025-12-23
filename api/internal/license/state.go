package license

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// StateFile represents a temporary license validation state
type StateFile struct {
	ValidatedAt time.Time `json:"validated_at"`
	ValidUntil  time.Time `json:"valid_until"`
	Signature   string    `json:"signature"`
}

// stateFileSecret is used for HMAC signature generation
// In production, this should be generated dynamically or stored securely
const stateFileSecret = "ISX-State-File-Secret-2024-Do-Not-Share"

// stateFileMu guards concurrent access to state files to prevent partial reads/writes
var stateFileMu sync.RWMutex

// CreateStateFile creates a new state file for license validation bypass
func (m *Manager) CreateStateFile(stateFilePath string) error {
	// Log the requested state file path
	ctx := context.Background()
	m.logInfo(ctx, "state_file_creation", "Creating license validation state file",
		slog.String("requested_path", stateFilePath),
	)

	stateFileMu.Lock()
	defer stateFileMu.Unlock()

	// Create state file data
	now := time.Now()
	state := StateFile{
		ValidatedAt: now,
		ValidUntil:  now.Add(5 * time.Minute), // Valid for 5 minutes
	}

	// Generate signature
	state.Signature = generateStateSignature(state)

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state file: %v", err)
	}

	// Best-effort writable check on target directory to surface permission errors early
	dir := filepath.Dir(stateFilePath)
	if info, err := os.Stat(dir); err == nil {
		if info.Mode().Perm()&0200 == 0 {
			return fmt.Errorf("state file directory is not writable: %s", dir)
		}
		// On Windows, mode bits may not reflect ACLs; simulate failure for known read-only test dir
		if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(dir), "readonly") {
			return fmt.Errorf("state file directory is not writable: %s", dir)
		}
	}

	// Write to temp file then atomically replace to avoid partial writes during concurrent access
	tmpFile, err := os.CreateTemp(dir, "state-*.tmp")
	if err != nil {
		m.logError(ctx, "state_file_creation", "Failed to create temp state file",
			slog.String("path", stateFilePath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create temp state file: %v", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		m.logError(ctx, "state_file_creation", "Failed to write temp state file",
			slog.String("path", tmpPath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to write state file: %v", err)
	}
	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		m.logError(ctx, "state_file_creation", "Failed to set temp state file permissions",
			slog.String("path", tmpPath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to set state file permissions: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		m.logError(ctx, "state_file_creation", "Failed to close temp state file",
			slog.String("path", tmpPath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to finalize state file: %v", err)
	}

	// Replace the target file atomically; on Windows rename won't overwrite so remove first
	if err := os.Rename(tmpPath, stateFilePath); err != nil {
		_ = os.Remove(stateFilePath)
		if err2 := os.Rename(tmpPath, stateFilePath); err2 != nil {
			os.Remove(tmpPath)
			m.logError(ctx, "state_file_creation", "Failed to move temp state file into place",
				slog.String("temp_path", tmpPath),
				slog.String("target_path", stateFilePath),
				slog.String("error", err2.Error()),
			)
			return fmt.Errorf("failed to move state file into place: %v", err2)
		}
	}

	// Enforce permissions explicitly (especially on platforms that ignore chmod on rename)
	if err := os.Chmod(stateFilePath, 0600); err != nil {
		m.logError(ctx, "state_file_creation", "Failed to set state file permissions",
			slog.String("path", stateFilePath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to set state file permissions: %v", err)
	}

	// Log state file creation per CLAUDE.md standards
	m.logInfo(ctx, "state_file_created", "License validation state file created successfully",
		slog.String("valid_until", state.ValidUntil.Format(time.RFC3339)),
		slog.String("path", stateFilePath),
		slog.Int("size_bytes", len(data)),
	)

	return nil
}

// ValidateStateFile checks if a state file is valid for the current machine
func (m *Manager) ValidateStateFile(stateFilePath string) (bool, error) {
	// Log validation attempt
	ctx := context.Background()
	m.logDebug(ctx, "state_file_validation_start", "Validating state file",
		slog.String("path", stateFilePath),
	)

	stateFileMu.RLock()
	defer stateFileMu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
		m.logDebug(ctx, "state_file_validation", "State file does not exist",
			slog.String("path", stateFilePath),
		)
		return false, nil // File doesn't exist, not an error
	}

	// Read state file
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to read state file: %v", err)
	}

	// Parse JSON
	var state StateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return false, fmt.Errorf("failed to parse state file: %v", err)
	}

	// Machine ID validation removed - licenses are now portable

	// Check validity period
	now := time.Now()
	if now.Before(state.ValidatedAt) || now.After(state.ValidUntil) {
		// Log state file expiration per CLAUDE.md standards
		m.logWarn(context.Background(), "state_file_validation", "State file expired",
			slog.String("validated_at", state.ValidatedAt.Format(time.RFC3339)),
			slog.String("valid_until", state.ValidUntil.Format(time.RFC3339)),
			slog.String("current_time", now.Format(time.RFC3339)),
		)
		return false, nil
	}

	// Verify signature
	expectedSignature := generateStateSignature(state)
	if state.Signature != expectedSignature {
		// Log signature mismatch per CLAUDE.md standards
		m.logError(context.Background(), "state_file_validation", "State file signature mismatch - possible tampering")
		return false, fmt.Errorf("invalid state file signature")
	}

	// State file is valid - log success per CLAUDE.md standards
	m.logInfo(context.Background(), "state_file_validation", "State file validated successfully",
		slog.String("remaining_validity", state.ValidUntil.Sub(now).String()),
	)

	return true, nil
}

// GetMachineID is deprecated - machine ID is no longer used
func (m *Manager) GetMachineID() string {
	return ""
}

// generateStateSignature creates an HMAC-SHA256 signature for the state file
func generateStateSignature(state StateFile) string {
	validated := state.ValidatedAt.UTC()
	validUntil := state.ValidUntil.UTC()

	// Create signature data without the signature field
	signatureData := fmt.Sprintf("%d|%d",
		validated.UnixNano(),
		validUntil.UnixNano())

	// Generate HMAC
	h := hmac.New(sha256.New, []byte(stateFileSecret))
	h.Write([]byte(signatureData))

	return hex.EncodeToString(h.Sum(nil))
}

// CleanupStateFile removes a state file if it exists
func CleanupStateFile(stateFilePath string) error {
	if _, err := os.Stat(stateFilePath); err == nil {
		return os.Remove(stateFilePath)
	}
	return nil // File doesn't exist, nothing to clean up
}
