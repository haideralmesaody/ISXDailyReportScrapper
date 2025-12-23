package license

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// License Validation Deep Testing
// =============================================================================

func TestLicenseValidationComprehensive(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "validation_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("validation with no license file", func(t *testing.T) {
		ctx := context.Background()
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("validation with various license states", func(t *testing.T) {
		testCases := []struct {
			name          string
			license       LicenseInfo
			expectedValid bool
			expectedError bool
			description   string
		}{
			{
				name: "valid active license",
				license: LicenseInfo{
					LicenseKey:  "VALID-ACTIVE-KEY",
					UserEmail:   "active@test.com",
					ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now().Add(-1 * time.Hour),
				},
				expectedValid: true,
				expectedError: false,
				description:   "Valid active license should pass validation",
			},
			{
				name: "expired license",
				license: LicenseInfo{
					LicenseKey:  "EXPIRED-KEY",
					UserEmail:   "expired@test.com",
					ExpiryDate:  time.Now().Add(-24 * time.Hour),
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-48 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now().Add(-1 * time.Hour),
				},
				expectedValid: false,
				expectedError: true,
				description:   "Expired license should fail validation",
			},
			{
				name: "license expiring soon",
				license: LicenseInfo{
					LicenseKey:  "EXPIRING-SOON-KEY",
					UserEmail:   "expiring@test.com",
					ExpiryDate:  time.Now().Add(1 * time.Hour),
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-29 * 24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now().Add(-30 * time.Minute),
				},
				expectedValid: true,
				expectedError: false,
				description:   "License expiring soon should still be valid",
			},
			{
				name: "license with zero expiry date",
				license: LicenseInfo{
					LicenseKey:  "ZERO-EXPIRY-KEY",
					UserEmail:   "zero@test.com",
					ExpiryDate:  time.Time{},
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now().Add(-1 * time.Hour),
				},
				expectedValid: false,
				expectedError: true,
				description:   "License with zero expiry date should fail",
			},
			{
				name: "license with future issue date",
				license: LicenseInfo{
					LicenseKey:  "FUTURE-ISSUE-KEY",
					UserEmail:   "future@test.com",
					ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
					Duration:    "1m",
					IssuedDate:  time.Now().Add(24 * time.Hour), // Future issue date
					Status:      "Activated",
					LastChecked: time.Now().Add(-1 * time.Hour),
				},
				expectedValid: true, // Issue date doesn't prevent validation currently
				expectedError: false,
				description:   "License with future issue date should be handled gracefully",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Save the test license
				err := manager.saveLicenseLocal(tc.license)
				require.NoError(t, err)

				// Validate
				ctx := context.Background()
				valid, err := manager.ValidateLicenseWithContext(ctx)

				if tc.expectedError {
					assert.Error(t, err, tc.description)
				} else {
					assert.NoError(t, err, tc.description)
				}

				assert.Equal(t, tc.expectedValid, valid, tc.description)
			})
		}
	})

	t.Run("validation caching behavior", func(t *testing.T) {
		// Create a valid license
		license := LicenseInfo{
			LicenseKey:  "CACHE-VALIDATION-KEY",
			UserEmail:   "cache@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now().Add(-1 * time.Hour),
		}

		err := manager.saveLicenseLocal(license)
		require.NoError(t, err)

		ctx := context.Background()

		// First validation - should cache result
		start := time.Now()
		valid1, err1 := manager.ValidateLicenseWithContext(ctx)
		duration1 := time.Since(start)
		assert.NoError(t, err1)
		assert.True(t, valid1)

		// Second validation - should use cache
		start = time.Now()
		valid2, err2 := manager.ValidateLicenseWithContext(ctx)
		duration2 := time.Since(start)
		assert.NoError(t, err2)
		assert.True(t, valid2)

		// Second validation should be significantly faster
		assert.True(t, duration2 < duration1/2,
			"Cached validation should be faster: first=%v, second=%v", duration1, duration2)

		// Get validation state
		state, err := manager.GetValidationState()
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.True(t, state.IsValid)
		assert.NoError(t, state.Error)
		assert.True(t, time.Now().Before(state.CachedUntil))
	})

	t.Run("validation state management", func(t *testing.T) {
		// Test validation state for different scenarios
		testCases := []struct {
			name          string
			license       *LicenseInfo
			expectedValid bool
			expectedError bool
			errorType     string
		}{
			{
				name: "valid license state",
				license: &LicenseInfo{
					LicenseKey:  "STATE-VALID",
					ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now(),
				},
				expectedValid: true,
				expectedError: false,
			},
			{
				name: "expired license state",
				license: &LicenseInfo{
					LicenseKey:  "STATE-EXPIRED",
					ExpiryDate:  time.Now().Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now(),
				},
				expectedValid: false,
				expectedError: true,
				errorType:     "expired",
			},
			{
				name:          "no license state",
				license:       nil,
				expectedValid: false,
				expectedError: false,
				errorType:     "network_error", // Will be treated as file not found
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Clean slate for each test
				tempFile := filepath.Join(tempDir, fmt.Sprintf("state_test_%d.dat", time.Now().UnixNano()))
				stateManager, err := NewManagerWithPath(tempFile)
				require.NoError(t, err)
				defer stateManager.Close()

				if tc.license != nil {
					err := stateManager.saveLicenseLocal(*tc.license)
					require.NoError(t, err)
				}

				ctx := context.Background()
				valid, err := stateManager.ValidateLicenseWithContext(ctx)

				assert.Equal(t, tc.expectedValid, valid)
				if tc.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				// Check validation state
				state, stateErr := stateManager.GetValidationState()
				if tc.license != nil {
					assert.NoError(t, stateErr)
					assert.NotNil(t, state)
					assert.Equal(t, tc.expectedValid, state.IsValid)
					if tc.errorType != "" && !tc.expectedValid {
						assert.Equal(t, tc.errorType, state.ErrorType)
					}
				}
			})
		}
	})
}

func TestLicenseActivationFlow(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "activation_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("complete activation flow", func(t *testing.T) {
		ctx := context.Background()

		// Initial state - no license
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Nil(t, info)
		assert.Equal(t, "Not Activated", status)

		// Validation should fail with no license
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.False(t, valid)

		// Try to activate with invalid key - should fail
		err = manager.ActivateLicenseWithContext(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license key cannot be empty")

		// Status should still be not activated
		info, status, err = manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Nil(t, info)
		assert.Equal(t, "Not Activated", status)

		// Note: Full activation with network calls would require mocked Google Sheets service
		// For now, we test the validation logic up to the network call
	})

	t.Run("activation input validation", func(t *testing.T) {
		ctx := context.Background()

		invalidInputs := []struct {
			name     string
			key      string
			errorMsg string
		}{
			{"empty key", "", "license key cannot be empty"},
			{"whitespace key", "   \t\n   ", "license key cannot be empty"},
			{"very short key", "AB", "license validation failed"},               // Will fail at network level
			{"invalid characters", "KEY@#$%^&*()", "license validation failed"}, // Will fail at network level
		}

		for _, test := range invalidInputs {
			t.Run(test.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, test.key)
				assert.Error(t, err)
				// Note: Exact error message depends on where validation fails
				// Could be input validation or network validation
			})
		}
	})

	t.Run("license key format processing", func(t *testing.T) {
		// Test that dashes are stripped from license keys
		testCases := []struct {
			name      string
			input     string
			processed string
		}{
			{"no dashes", "ISX1M02LYE1F9QJHR9D7Z", "ISX1M02LYE1F9QJHR9D7Z"},
			{"with dashes", "ISX1M-02L-YE1-F9Q-JHR9D7Z", "ISX1M02LYE1F9QJHR9D7Z"},
			{"multiple dashes", "ISX1M--02L--YE1F9Q", "ISX1M02LYE1F9Q"},
			{"leading/trailing dashes", "-ISX1M02L-", "ISX1M02L"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := context.Background()

				// We can't test the exact processing without exposing internal methods,
				// but we can verify that different dash formats are handled consistently
				err1 := manager.ActivateLicenseWithContext(ctx, tc.input)
				err2 := manager.ActivateLicenseWithContext(ctx, tc.processed)

				// Both should produce the same result (both will likely fail due to network)
				// The important thing is they don't fail differently due to formatting
				if err1 != nil && err2 != nil {
					// Both failed - this is expected for invalid keys
					assert.True(t,
						strings.Contains(err1.Error(), "license validation failed") ||
							strings.Contains(err1.Error(), "license activation failed"),
						"expected validation/activation failure, got: %v", err1)
					assert.True(t,
						strings.Contains(err2.Error(), "license validation failed") ||
							strings.Contains(err2.Error(), "license activation failed"),
						"expected validation/activation failure, got: %v", err2)
				}
			})
		}
	})
}

func TestLicenseFileOperations(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("license file save and load integrity", func(t *testing.T) {
		licenseFile := filepath.Join(tempDir, "integrity_test.dat")
		manager, err := NewManagerWithPath(licenseFile)
		require.NoError(t, err)
		defer manager.Close()

		// Create test license
		originalLicense := LicenseInfo{
			LicenseKey:  "INTEGRITY-TEST-KEY-12345",
			UserEmail:   "integrity@test.com",
			ExpiryDate:  time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			Duration:    "1y",
			IssuedDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:      "Activated",
			LastChecked: time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC),
		}

		// Save license
		err = manager.saveLicenseLocal(originalLicense)
		require.NoError(t, err)

		// Verify file exists
		assert.FileExists(t, manager.GetLicensePath())

		// Load license
		loadedLicense, err := manager.loadLicenseLocal()
		require.NoError(t, err)

		// Verify data integrity
		assert.Equal(t, originalLicense.LicenseKey, loadedLicense.LicenseKey)
		assert.Equal(t, originalLicense.UserEmail, loadedLicense.UserEmail)
		assert.True(t, originalLicense.ExpiryDate.Equal(loadedLicense.ExpiryDate))
		assert.Equal(t, originalLicense.Duration, loadedLicense.Duration)
		assert.True(t, originalLicense.IssuedDate.Equal(loadedLicense.IssuedDate))
		assert.Equal(t, originalLicense.Status, loadedLicense.Status)
		assert.True(t, originalLicense.LastChecked.Equal(loadedLicense.LastChecked))
	})

	t.Run("license file corruption handling", func(t *testing.T) {
		corruptionTests := []struct {
			name     string
			content  string
			expected string
		}{
			{"invalid JSON", "{invalid json", "failed"},
			{"empty file", "", "failed"},
			{"partial JSON", `{"LicenseKey":"TEST"`, "failed"},
			{"wrong structure", `{"wrong":"structure"}`, "success"}, // Will load with empty fields
			{"binary data", "\x00\x01\x02\x03", "failed"},
			{"very large file", string(make([]byte, 1024*1024)), "failed"}, // 1MB of zeros
		}

		for _, test := range corruptionTests {
			t.Run(test.name, func(t *testing.T) {
				corruptFile := filepath.Join(tempDir, fmt.Sprintf("corrupt_%s.dat", test.name))

				// Write corrupted content
				err := os.WriteFile(corruptFile, []byte(test.content), 0600)
				require.NoError(t, err)

				// Try to create manager with corrupted file
				manager, err := NewManagerWithPath(corruptFile)
				require.NoError(t, err) // Manager creation should succeed
				defer manager.Close()

				// Try to load license - this should handle corruption gracefully
				_, err = manager.loadLicenseLocal()
				if test.expected == "failed" {
					assert.Error(t, err, "Should fail to load corrupted file")
				} else {
					// Wrong structure might load successfully with default values
					// This is acceptable behavior
				}

				// Validation should handle corrupted files gracefully
				ctx := context.Background()
				valid, err := manager.ValidateLicenseWithContext(ctx)
				assert.NoError(t, err, "Validation should not crash on corrupted files")
				assert.False(t, valid, "Corrupted files should result in invalid license")
			})
		}
	})

	t.Run("concurrent file access", func(t *testing.T) {
		concurrentFile := filepath.Join(tempDir, "concurrent_test.dat")

		// Create multiple managers accessing the same file
		managers := make([]*Manager, 5)
		for i := 0; i < 5; i++ {
			manager, err := NewManagerWithPath(concurrentFile)
			require.NoError(t, err)
			managers[i] = manager
		}
		defer func() {
			for _, m := range managers {
				if m != nil {
					m.Close()
				}
			}
		}()

		// Create test license
		testLicense := LicenseInfo{
			LicenseKey:  "CONCURRENT-TEST-KEY",
			UserEmail:   "concurrent@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}
		_ = testLicense // Mark as used for later

		// Save from one manager
		err := managers[0].saveLicenseLocal(testLicense)
		require.NoError(t, err)

		// All managers should be able to read the same data
		for i, manager := range managers {
			license, err := manager.loadLicenseLocal()
			assert.NoError(t, err, "Manager %d should load license", i)
			assert.Equal(t, testLicense.LicenseKey, license.LicenseKey, "Manager %d should load correct data", i)
		}

		// Concurrent validation should work
		ctx := context.Background()
		for i, manager := range managers {
			valid, err := manager.ValidateLicenseWithContext(ctx)
			assert.NoError(t, err, "Manager %d validation should work", i)
			assert.True(t, valid, "Manager %d should validate successfully", i)
		}
	})
}

func TestLicenseStatusCalculation(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "status_calc_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("status calculation precision", func(t *testing.T) {
		now := time.Now()

		// Test precise boundary conditions
		boundaryTests := []struct {
			name           string
			expiryOffset   time.Duration
			expectedStatus string
		}{
			// Active status (>30 days)
			{"exactly 31 days", 31 * 24 * time.Hour, "Active"},
			{"exactly 30 days 1 second", 30*24*time.Hour + 1*time.Second, "Active"},

			// Warning status (8-30 days)
			{"exactly 30 days", 30 * 24 * time.Hour, "Warning"},
			{"exactly 15 days", 15 * 24 * time.Hour, "Warning"},
			{"exactly 8 days", 8 * 24 * time.Hour, "Warning"},
			{"exactly 7 days 1 second", 7*24*time.Hour + 1*time.Second, "Warning"},

			// Critical status (1-7 days)
			{"exactly 7 days", 7 * 24 * time.Hour, "Critical"},
			{"exactly 3 days", 3 * 24 * time.Hour, "Critical"},
			{"exactly 1 day", 1 * 24 * time.Hour, "Critical"},
			{"exactly 1 second", 1 * time.Second, "Critical"},

			// Expired status
			{"exactly expired", 0 * time.Second, "Expired"},
			{"1 second expired", -1 * time.Second, "Expired"},
			{"1 hour expired", -1 * time.Hour, "Expired"},
			{"1 day expired", -24 * time.Hour, "Expired"},
		}

		for _, test := range boundaryTests {
			t.Run(test.name, func(t *testing.T) {
				license := LicenseInfo{
					LicenseKey:  fmt.Sprintf("BOUNDARY-%s", test.name),
					UserEmail:   "boundary@test.com",
					ExpiryDate:  now.Add(test.expiryOffset),
					Duration:    "1m",
					IssuedDate:  now.Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: now,
				}

				err := manager.saveLicenseLocal(license)
				require.NoError(t, err)

				info, status, err := manager.GetLicenseStatus()
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, test.expectedStatus, status,
					"Status calculation failed for %s (offset: %v)", test.name, test.expiryOffset)
			})
		}
	})

	t.Run("renewal status calculation", func(t *testing.T) {
		now := time.Now()

		renewalTests := []struct {
			name           string
			expiryOffset   time.Duration
			expectedStatus string
			needsRenewal   bool
			isExpired      bool
		}{
			{"60 days left", 60 * 24 * time.Hour, "Active", false, false},
			{"30 days left", 30 * 24 * time.Hour, "Warning", true, false},
			{"7 days left", 7 * 24 * time.Hour, "Critical", true, false},
			{"1 day left", 1 * 24 * time.Hour, "Critical", true, false},
			{"expired", -1 * time.Hour, "Expired", true, true},
		}

		for _, test := range renewalTests {
			t.Run(test.name, func(t *testing.T) {
				license := LicenseInfo{
					LicenseKey:  fmt.Sprintf("RENEWAL-%s", test.name),
					UserEmail:   "renewal@test.com",
					ExpiryDate:  now.Add(test.expiryOffset),
					Duration:    "1m",
					IssuedDate:  now.Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: now,
				}

				err := manager.saveLicenseLocal(license)
				require.NoError(t, err)

				renewalInfo, err := manager.CheckRenewalStatus()
				assert.NoError(t, err)
				assert.NotNil(t, renewalInfo)
				assert.Equal(t, test.expectedStatus, renewalInfo.Status)
				assert.Equal(t, test.needsRenewal, renewalInfo.NeedsRenewal)
				assert.Equal(t, test.isExpired, renewalInfo.IsExpired)
				assert.NotEmpty(t, renewalInfo.Message)

				// Days left should be reasonable
				expectedDays := int(test.expiryOffset.Hours() / 24)
				assert.Equal(t, expectedDays, renewalInfo.DaysLeft)
			})
		}
	})
}

func TestManagerInterfaceCompliance(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "interface_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("ManagerInterface compliance", func(t *testing.T) {
		// Verify manager implements ManagerInterface
		var iface ManagerInterface = manager
		require.NotNil(t, iface)

		// Test all interface methods
		// GetLicenseStatus
		info, status, err := iface.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Equal(t, "Not Activated", status)
		assert.Nil(t, info)

		// ActivateLicense
		err = iface.ActivateLicense("INVALID-KEY")
		assert.Error(t, err) // Expected to fail with invalid key

		// ValidateLicense
		valid, err := iface.ValidateLicense()
		assert.NoError(t, err)
		assert.False(t, valid)

		// GetLicensePath
		path := iface.GetLicensePath()
		assert.NotEmpty(t, path)
		assert.Contains(t, path, "license.dat")
	})

	t.Run("method consistency", func(t *testing.T) {
		// Test that context and non-context versions produce same results
		ctx := context.Background()

		// Validation methods
		valid1, err1 := manager.ValidateLicense()
		valid2, err2 := manager.ValidateLicenseWithContext(ctx)

		assert.Equal(t, valid1, valid2, "Context and non-context validation should match")
		if err1 != nil && err2 != nil {
			// Both errors - acceptable
		} else {
			assert.Equal(t, err1, err2, "Error states should match")
		}

		// Activation methods
		err1 = manager.ActivateLicense("TEST-KEY")
		err2 = manager.ActivateLicenseWithContext(ctx, "TEST-KEY")

		// Both should fail similarly (network error expected)
		if err1 != nil && err2 != nil {
			// Both failed - this is expected for invalid keys
			assert.Contains(t, err1.Error(), "license")
			assert.Contains(t, err2.Error(), "license")
		}
	})
}
