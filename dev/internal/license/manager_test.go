package license

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestManager tests the license manager functionality
func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		licenseFile string
		wantErr     bool
	}{
		{
			name:        "valid license file path",
			licenseFile: "test_license.dat",
			wantErr:     false,
		},
		{
			name:        "empty license file path",
			licenseFile: "",
			wantErr:     false, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.licenseFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if manager != nil {
				// Cleanup
				if manager.logger != nil {
					manager.logger.Close()
				}
			}
		})
	}
}

// TestGenerateMachineID tests machine ID generation
func TestGenerateMachineID(t *testing.T) {
	id1, err := generateMachineID()
	if err != nil {
		t.Fatalf("generateMachineID() error = %v", err)
	}

	if len(id1) != 24 {
		t.Errorf("generateMachineID() returned ID with length %d, want 24", len(id1))
	}

	// Generate another ID - should be the same on the same machine
	id2, err := generateMachineID()
	if err != nil {
		t.Fatalf("generateMachineID() error = %v", err)
	}

	if id1 != id2 {
		t.Errorf("generateMachineID() returned different IDs on same machine: %s != %s", id1, id2)
	}
}

// TestLicenseActivation tests the license activation flow
func TestLicenseActivation(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	// Create manager
	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Mock license data for testing
	mockLicense := LicenseInfo{
		LicenseKey:  "TEST-LICENSE-KEY-123",
		UserEmail:   "test@example.com",
		ExpiryDate:  time.Now().Add(30 * 24 * time.Hour), // 30 days from now
		Duration:    "1m",
		MachineID:   manager.machineID,
		IssuedDate:  time.Now(),
		Status:      "Active",
		LastChecked: time.Now(),
	}

	// Save mock license locally
	err = manager.saveLicenseLocal(mockLicense)
	if err != nil {
		t.Fatalf("Failed to save license: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(licenseFile); os.IsNotExist(err) {
		t.Errorf("License file was not created")
	}

	// Load and verify
	loaded, err := manager.loadLicenseLocal()
	if err != nil {
		t.Fatalf("Failed to load license: %v", err)
	}

	if loaded.LicenseKey != mockLicense.LicenseKey {
		t.Errorf("Loaded license key = %v, want %v", loaded.LicenseKey, mockLicense.LicenseKey)
	}

	if loaded.MachineID != mockLicense.MachineID {
		t.Errorf("Loaded machine ID = %v, want %v", loaded.MachineID, mockLicense.MachineID)
	}
}

// TestLicenseValidation tests license validation logic
func TestLicenseValidation(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	tests := []struct {
		name    string
		license LicenseInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid license",
			license: LicenseInfo{
				LicenseKey:  "VALID-KEY",
				ExpiryDate:  time.Now().Add(24 * time.Hour),
				MachineID:   manager.machineID,
				Status:      "Active",
				LastChecked: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "expired license",
			license: LicenseInfo{
				LicenseKey:  "EXPIRED-KEY",
				ExpiryDate:  time.Now().Add(-24 * time.Hour),
				MachineID:   manager.machineID,
				Status:      "Active",
				LastChecked: time.Now(),
			},
			wantErr: true,
			errMsg:  "expired",
		},
		{
			name: "wrong machine ID",
			license: LicenseInfo{
				LicenseKey:  "WRONG-MACHINE-KEY",
				ExpiryDate:  time.Now().Add(24 * time.Hour),
				MachineID:   "different-machine-id",
				Status:      "Active",
				LastChecked: time.Now(),
			},
			wantErr: true,
			errMsg:  "machine_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save test license
			err := manager.saveLicenseLocal(tt.license)
			if err != nil {
				t.Fatalf("Failed to save test license: %v", err)
			}

			// Perform validation
			valid, err := manager.performValidation()

			if tt.wantErr {
				if err == nil {
					t.Errorf("performValidation() expected error containing %q, got nil", tt.errMsg)
				} else if !errors.Is(err, errors.New(tt.errMsg)) && err.Error() != tt.errMsg {
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("performValidation() error = %v, want error containing %q", err, tt.errMsg)
					}
				}
				if valid {
					t.Errorf("performValidation() = true, want false for invalid license")
				}
			} else {
				if err != nil {
					t.Errorf("performValidation() unexpected error = %v", err)
				}
				if !valid {
					t.Errorf("performValidation() = false, want true for valid license")
				}
			}
		})
	}
}

// TestLicenseExpiry tests license expiry calculations
func TestLicenseExpiry(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name       string
		expiryDate time.Time
		wantStatus string
	}{
		{
			name:       "expired",
			expiryDate: time.Now().Add(-24 * time.Hour),
			wantStatus: "Expired",
		},
		{
			name:       "critical - 5 days left",
			expiryDate: time.Now().Add(5 * 24 * time.Hour),
			wantStatus: "Critical",
		},
		{
			name:       "warning - 20 days left",
			expiryDate: time.Now().Add(20 * 24 * time.Hour),
			wantStatus: "Warning",
		},
		{
			name:       "active - 60 days left",
			expiryDate: time.Now().Add(60 * 24 * time.Hour),
			wantStatus: "Active",
		},
		{
			name:       "available - no expiry set",
			expiryDate: time.Time{},
			wantStatus: "Available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := manager.calculateExpireStatus(tt.expiryDate)
			if status != tt.wantStatus {
				t.Errorf("calculateExpireStatus() = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestRenewalInfo tests renewal information generation
func TestRenewalInfo(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	tests := []struct {
		name         string
		license      *LicenseInfo
		wantStatus   string
		wantNeedsRenewal bool
		wantExpired  bool
	}{
		{
			name:         "no license",
			license:      nil,
			wantStatus:   "No License",
			wantNeedsRenewal: true,
			wantExpired:  true,
		},
		{
			name: "active license",
			license: &LicenseInfo{
				ExpiryDate: time.Now().Add(60 * 24 * time.Hour),
				MachineID:  manager.machineID,
			},
			wantStatus:   "Active",
			wantNeedsRenewal: false,
			wantExpired:  false,
		},
		{
			name: "expiring soon",
			license: &LicenseInfo{
				ExpiryDate: time.Now().Add(5 * 24 * time.Hour),
				MachineID:  manager.machineID,
			},
			wantStatus:   "Critical",
			wantNeedsRenewal: true,
			wantExpired:  false,
		},
		{
			name: "expired",
			license: &LicenseInfo{
				ExpiryDate: time.Now().Add(-5 * 24 * time.Hour),
				MachineID:  manager.machineID,
			},
			wantStatus:   "Expired",
			wantNeedsRenewal: true,
			wantExpired:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.license != nil {
				// Save test license
				err := manager.saveLicenseLocal(*tt.license)
				if err != nil {
					t.Fatalf("Failed to save test license: %v", err)
				}
			} else {
				// Remove any existing license file
				os.Remove(licenseFile)
			}

			info, err := manager.CheckRenewalStatus()
			if tt.license == nil && err == nil {
				// Should have error when no license
				if !info.NeedsRenewal {
					t.Errorf("CheckRenewalStatus() should indicate renewal needed when no license")
				}
			}

			if info.Status != tt.wantStatus {
				t.Errorf("CheckRenewalStatus() status = %v, want %v", info.Status, tt.wantStatus)
			}

			if info.NeedsRenewal != tt.wantNeedsRenewal {
				t.Errorf("CheckRenewalStatus() needsRenewal = %v, want %v", info.NeedsRenewal, tt.wantNeedsRenewal)
			}

			if info.IsExpired != tt.wantExpired {
				t.Errorf("CheckRenewalStatus() isExpired = %v, want %v", info.IsExpired, tt.wantExpired)
			}
		})
	}
}

// TestConcurrentAccess tests thread safety of license operations
func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a valid license
	license := LicenseInfo{
		LicenseKey:  "CONCURRENT-TEST",
		ExpiryDate:  time.Now().Add(24 * time.Hour),
		MachineID:   manager.machineID,
		Status:      "Active",
		LastChecked: time.Now(),
	}

	err = manager.saveLicenseLocal(license)
	if err != nil {
		t.Fatalf("Failed to save license: %v", err)
	}

	// Run concurrent validations
	done := make(chan bool)
	errorChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			valid, err := manager.ValidateLicense()
			if err != nil {
				errorChan <- err
			}
			if !valid {
				errorChan <- errors.New("validation failed")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errorChan)

	// Check for errors
	for err := range errorChan {
		t.Errorf("Concurrent validation error: %v", err)
	}
}

// TestLicenseFileCorruption tests handling of corrupted license files
func TestLicenseFileCorruption(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	tests := []struct {
		name     string
		content  []byte
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty file",
			content:  []byte{},
			wantErr:  true,
			errMsg:   "invalid",
		},
		{
			name:     "invalid JSON",
			content:  []byte("{invalid json}"),
			wantErr:  true,
			errMsg:   "invalid",
		},
		{
			name:     "valid JSON wrong structure",
			content:  []byte(`{"foo": "bar"}`),
			wantErr:  false, // Will load but with empty fields
		},
		{
			name: "valid license",
			content: func() []byte {
				license := LicenseInfo{
					LicenseKey:  "VALID",
					ExpiryDate:  time.Now().Add(24 * time.Hour),
					MachineID:   manager.machineID,
					Status:      "Active",
					LastChecked: time.Now(),
				}
				data, _ := json.MarshalIndent(license, "", "  ")
				return data
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test content
			err := os.WriteFile(licenseFile, tt.content, 0600)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Try to load
			_, err = manager.loadLicenseLocal()

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadLicenseLocal() expected error containing %q, got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("loadLicenseLocal() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("loadLicenseLocal() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestContextCancellation tests proper context handling
func TestContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManager(licenseFile)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a context that we'll cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create a mock validation that would take time
	done := make(chan bool, 1)
	var validationErr error

	go func() {
		// This should respect the cancelled context
		select {
		case <-ctx.Done():
			validationErr = ctx.Err()
		case <-time.After(5 * time.Second):
			validationErr = errors.New("context cancellation not respected")
		}
		done <- true
	}()

	select {
	case <-done:
		if validationErr != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", validationErr)
		}
	case <-time.After(1 * time.Second):
		t.Error("Context cancellation took too long")
	}
}

// Helper function to check if error message contains substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}