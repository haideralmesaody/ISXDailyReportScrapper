package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StateFile represents a temporary license validation state
type StateFile struct {
	MachineID    string    `json:"machine_id"`
	ValidatedAt  time.Time `json:"validated_at"`
	ValidUntil   time.Time `json:"valid_until"`
	Signature    string    `json:"signature"`
}

// stateFileSecret is used for HMAC signature generation
// In production, this should be generated dynamically or stored securely
const stateFileSecret = "ISX-State-File-Secret-2024-Do-Not-Share"

// CreateStateFile creates a new state file for license validation bypass
func (m *Manager) CreateStateFile(stateFilePath string) error {
	// Create state file data
	now := time.Now()
	state := StateFile{
		MachineID:   m.machineID,
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
	
	// Write to file with restricted permissions
	if err := os.WriteFile(stateFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %v", err)
	}
	
	// Log state file creation
	if m.logger != nil {
		m.logger.Log(LogEntry{
			Level:     LogLevelInfo,
			Action:    "state_file_created",
			Result:    "License validation state file created",
			MachineID: m.machineID[:min(8, len(m.machineID))],
			Details: map[string]interface{}{
				"valid_until": state.ValidUntil.Format(time.RFC3339),
				"path":        stateFilePath,
			},
		})
	}
	
	return nil
}

// ValidateStateFile checks if a state file is valid for the current machine
func (m *Manager) ValidateStateFile(stateFilePath string) (bool, error) {
	// Check if file exists
	if _, err := os.Stat(stateFilePath); os.IsNotExist(err) {
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
	
	// Check machine ID
	if state.MachineID != m.machineID {
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:     LogLevelWarn,
				Action:    "state_file_validation",
				Result:    "State file machine ID mismatch",
				MachineID: m.machineID[:min(8, len(m.machineID))],
				Details: map[string]interface{}{
					"expected": m.machineID[:min(8, len(m.machineID))],
					"found":    state.MachineID[:min(8, len(state.MachineID))],
				},
			})
		}
		return false, nil
	}
	
	// Check validity period
	now := time.Now()
	if now.Before(state.ValidatedAt) || now.After(state.ValidUntil) {
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:     LogLevelWarn,
				Action:    "state_file_validation",
				Result:    "State file expired",
				MachineID: m.machineID[:min(8, len(m.machineID))],
				Details: map[string]interface{}{
					"validated_at": state.ValidatedAt.Format(time.RFC3339),
					"valid_until":  state.ValidUntil.Format(time.RFC3339),
					"current_time": now.Format(time.RFC3339),
				},
			})
		}
		return false, nil
	}
	
	// Verify signature
	expectedSignature := generateStateSignature(state)
	if state.Signature != expectedSignature {
		if m.logger != nil {
			m.logger.Log(LogEntry{
				Level:  LogLevelError,
				Action: "state_file_validation",
				Result: "State file signature mismatch - possible tampering",
			})
		}
		return false, fmt.Errorf("invalid state file signature")
	}
	
	// State file is valid
	if m.logger != nil {
		m.logger.Log(LogEntry{
			Level:     LogLevelInfo,
			Action:    "state_file_validation",
			Result:    "State file validated successfully",
			MachineID: m.machineID[:min(8, len(m.machineID))],
			Details: map[string]interface{}{
				"remaining_validity": state.ValidUntil.Sub(now).String(),
			},
		})
	}
	
	return true, nil
}

// GetMachineID returns the machine ID for this manager instance
func (m *Manager) GetMachineID() string {
	return m.machineID
}

// generateStateSignature creates an HMAC-SHA256 signature for the state file
func generateStateSignature(state StateFile) string {
	// Create signature data without the signature field
	signatureData := fmt.Sprintf("%s|%s|%s",
		state.MachineID,
		state.ValidatedAt.Format(time.RFC3339),
		state.ValidUntil.Format(time.RFC3339))
	
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