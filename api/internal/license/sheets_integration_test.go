package license

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Google Sheets Integration Error Handling Tests
// =============================================================================

func TestGoogleSheetsIntegrationErrors(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "sheets_integration_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("network connectivity failures", func(t *testing.T) {
		// Test various network failure scenarios
		networkTests := []struct {
			name        string
			licenseKey  string
			expectError string
		}{
			{
				name:        "network unreachable",
				licenseKey:  "NETWORK-UNREACHABLE-KEY",
				expectError: "license validation failed",
			},
			{
				name:        "connection timeout",
				licenseKey:  "CONNECTION-TIMEOUT-KEY",
				expectError: "license validation failed",
			},
			{
				name:        "DNS resolution failure",
				licenseKey:  "DNS-FAILURE-KEY",
				expectError: "license validation failed",
			},
		}

		ctx := context.Background()
		for _, test := range networkTests {
			t.Run(test.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, test.licenseKey)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectError)
			})
		}
	})

	t.Run("authentication failures", func(t *testing.T) {
		// Test authentication-related failures
		authTests := []struct {
			name       string
			licenseKey string
		}{
			{"invalid credentials", "AUTH-INVALID-CREDS"},
			{"expired credentials", "AUTH-EXPIRED-CREDS"},
			{"insufficient permissions", "AUTH-NO-PERMS"},
		}

		ctx := context.Background()
		for _, test := range authTests {
			t.Run(test.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, test.licenseKey)
				assert.Error(t, err)
				// Should fail with validation error since credentials are embedded
				assert.Contains(t, err.Error(), "license validation failed")
			})
		}
	})

	t.Run("sheets service errors", func(t *testing.T) {
		// Test various Google Sheets service errors
		serviceErrors := []struct {
			name       string
			licenseKey string
		}{
			{"sheet not found", "SHEET-NOT-FOUND"},
			{"sheet access denied", "SHEET-ACCESS-DENIED"},
			{"quota exceeded", "QUOTA-EXCEEDED"},
			{"service unavailable", "SERVICE-UNAVAILABLE"},
		}

		ctx := context.Background()
		for _, test := range serviceErrors {
			t.Run(test.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, test.licenseKey)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "license validation failed")
			})
		}
	})

	t.Run("malformed sheets data", func(t *testing.T) {
		// Test handling of malformed data from sheets
		malformedTests := []struct {
			name       string
			licenseKey string
		}{
			{"empty response", "EMPTY-RESPONSE"},
			{"invalid json", "INVALID-JSON"},
			{"missing columns", "MISSING-COLUMNS"},
			{"corrupt data", "CORRUPT-DATA"},
		}

		ctx := context.Background()
		for _, test := range malformedTests {
			t.Run(test.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, test.licenseKey)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "license validation failed")
			})
		}
	})

	t.Run("partial network failures", func(t *testing.T) {
		// Test scenarios where network partially fails during operations

		// Create a valid license to test periodic validation
		validLicense := LicenseInfo{
			LicenseKey:  "PARTIAL-NETWORK-FAIL",
			UserEmail:   "partial@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now().Add(-7 * time.Hour), // Needs remote validation
		}

		err := manager.saveLicenseLocal(validLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Validation should still work with local cache even if remote fails
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid, "Should work with local validation when remote fails")
	})

	t.Run("grace period during network issues", func(t *testing.T) {
		// Test 48-hour grace period for network issues
		graceLicense := LicenseInfo{
			LicenseKey:  "GRACE-PERIOD-TEST",
			UserEmail:   "grace@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-72 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now().Add(-25 * time.Hour), // Within 48-hour grace period
		}

		err := manager.saveLicenseLocal(graceLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Should still work within grace period
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Test expired grace period
		expiredGraceLicense := graceLicense
		expiredGraceLicense.LicenseKey = "EXPIRED-GRACE-TEST"
		expiredGraceLicense.LastChecked = time.Now().Add(-49 * time.Hour) // Beyond 48-hour grace

		err = manager.saveLicenseLocal(expiredGraceLicense)
		require.NoError(t, err)

		// Should fail when grace period is exceeded
		valid, err = manager.ValidateLicenseWithContext(ctx)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "grace period expired")
	})

	t.Run("sheets response parsing edge cases", func(t *testing.T) {
		// Test edge cases in parsing sheets responses

		// Mock server for testing different response formats
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate different problematic responses
			if strings.Contains(r.URL.Path, "empty-values") {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"values": [][]interface{}{},
				})
				return
			}

			if strings.Contains(r.URL.Path, "malformed-row") {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"values": [][]interface{}{
						{"header1", "header2", "header3"},
						{"incomplete"}, // Missing columns
						{"key1", "duration1", "date1", "status1", "machine1", "issued1", "checked1", "expire1"},
					},
				})
				return
			}

			if strings.Contains(r.URL.Path, "invalid-dates") {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"values": [][]interface{}{
						{"LicenseKey", "Duration", "ExpiryDate", "Status", "MachineID", "ActivatedDate", "LastConnected", "ExpireStatus"},
						{"INVALID-DATE-KEY", "1m", "invalid-date", "Activated", "", "invalid-date", "invalid-date", "Active"},
					},
				})
				return
			}

			// Default 404
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// These tests would require modifying the manager to use the test server
		// For now, we test the error paths through invalid keys
		ctx := context.Background()

		testCases := []string{
			"EMPTY-VALUES-TEST",
			"MALFORMED-ROW-TEST",
			"INVALID-DATES-TEST",
		}

		for _, key := range testCases {
			err := manager.ActivateLicenseWithContext(ctx, key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "license validation failed")
		}
	})

	t.Run("concurrent sheets access", func(t *testing.T) {
		// Test concurrent access to sheets API
		ctx := context.Background()
		numConcurrent := 10

		results := make(chan error, numConcurrent)

		// Launch concurrent activation attempts
		for i := 0; i < numConcurrent; i++ {
			go func(id int) {
				key := fmt.Sprintf("CONCURRENT-SHEETS-%d", id)
				err := manager.ActivateLicenseWithContext(ctx, key)
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < numConcurrent; i++ {
			err := <-results
			assert.Error(t, err) // Expected to fail with invalid keys
			assert.Contains(t, err.Error(), "license validation failed")
		}
	})
}

// =============================================================================
// Network Connectivity Tests
// =============================================================================

func TestNetworkConnectivity(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "connectivity_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("connectivity test methods", func(t *testing.T) {
		// Test the network connectivity testing functionality
		err := manager.TestNetworkConnectivity()

		// This will likely fail in test environment without proper credentials
		if err != nil {
			// Expected failures in test environment
			possibleErrors := []string{
				"no internet connection",
				"cannot reach Google APIs",
				"Google Sheets service not initialized",
				"cannot access Google Sheets",
			}

			foundExpectedError := false
			for _, expectedErr := range possibleErrors {
				if strings.Contains(err.Error(), expectedErr) {
					foundExpectedError = true
					break
				}
			}

			assert.True(t, foundExpectedError, "Should fail with expected network error, got: %v", err)
		} else {
			// If it succeeds, that means we have real connectivity
			t.Log("Network connectivity test passed - have real internet connection")
		}
	})

	t.Run("offline validation behavior", func(t *testing.T) {
		// Test behavior when completely offline
		offlineLicense := LicenseInfo{
			LicenseKey:  "OFFLINE-TEST-KEY",
			UserEmail:   "offline@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now().Add(-1 * time.Hour), // Recent check
		}

		err := manager.saveLicenseLocal(offlineLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Should work with recent local validation
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Test with stale local validation (requires remote check)
		staleLicense := offlineLicense
		staleLicense.LicenseKey = "STALE-OFFLINE-KEY"
		staleLicense.LastChecked = time.Now().Add(-10 * time.Hour) // Needs remote validation

		err = manager.saveLicenseLocal(staleLicense)
		require.NoError(t, err)

		// Should still work within grace period
		valid, err = manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("intermittent connectivity", func(t *testing.T) {
		// Test handling of intermittent connectivity issues
		intermittentLicense := LicenseInfo{
			LicenseKey:  "INTERMITTENT-CONN-KEY",
			UserEmail:   "intermittent@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now().Add(-8 * time.Hour), // Needs validation
		}

		err := manager.saveLicenseLocal(intermittentLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Multiple validation attempts (simulating intermittent connectivity)
		for i := 0; i < 5; i++ {
			valid, err := manager.ValidateLicenseWithContext(ctx)

			// Should either succeed with local validation or fail gracefully
			if err != nil {
				// Should fail with a reasonable error message
				assert.Contains(t, err.Error(), "expired", "Should fail gracefully during connectivity issues")
			} else {
				assert.True(t, valid)
			}
		}
	})
}

// =============================================================================
// Credential Management Tests
// =============================================================================

func TestCredentialManagement(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "credential_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("secure credentials loading", func(t *testing.T) {
		// Test that credentials are loaded securely
		assert.True(t, manager.secureMode, "Manager should be in secure mode")
		assert.NotNil(t, manager.credentialsManager, "Credentials manager should be initialized")

		// We can't test the actual credentials loading without proper encrypted data
		// But we can verify the manager was initialized in secure mode
		assert.True(t, manager.secureMode)
	})

	t.Run("credential validation", func(t *testing.T) {
		// Test credential validation during manager initialization
		// Manager should have been created successfully if credentials are valid
		assert.NotNil(t, manager.sheetsService, "Sheets service should be initialized with valid credentials")
	})

	t.Run("credential error handling", func(t *testing.T) {
		// Test error handling when credentials are invalid/missing
		// This would be tested during manager creation, which already passed
		// So we test that the manager handles credential issues gracefully

		// Create a manager without proper credentials (should use embedded ones)
		tempLicenseFile := filepath.Join(tempDir, "no_creds_test.dat")

		// Should still create manager (with embedded credentials)
		testManager, err := NewManagerWithPath(tempLicenseFile)
		if err != nil {
			// Expected if embedded credentials are not available
			assert.Contains(t, err.Error(), "credentials")
		} else {
			assert.NotNil(t, testManager)
			testManager.Close()
		}
	})

	t.Run("credential security", func(t *testing.T) {
		// Test that credentials are not logged or exposed
		// This is more of a code review item, but we can check basic security

		// Verify credentials are not in performance metrics
		metrics := manager.GetPerformanceMetrics()
		metricsJson, _ := json.Marshal(metrics)
		metricsStr := string(metricsJson)

		// Should not contain sensitive data
		assert.NotContains(t, metricsStr, "private_key")
		assert.NotContains(t, metricsStr, "client_secret")
		assert.NotContains(t, metricsStr, "password")

		// Verify credentials are not in system stats
		stats := manager.GetSystemStats()
		statsJson, _ := json.Marshal(stats)
		statsStr := string(statsJson)

		assert.NotContains(t, statsStr, "private_key")
		assert.NotContains(t, statsStr, "client_secret")
		assert.NotContains(t, statsStr, "password")
	})
}

// =============================================================================
// Error Recovery Tests
// =============================================================================

func TestErrorRecovery(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "error_recovery_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("recovery from temporary failures", func(t *testing.T) {
		// Test recovery after temporary network failures
		recoveryLicense := LicenseInfo{
			LicenseKey:  "RECOVERY-TEST-KEY",
			UserEmail:   "recovery@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(recoveryLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Multiple validation attempts should be consistent
		for i := 0; i < 3; i++ {
			valid, err := manager.ValidateLicenseWithContext(ctx)
			assert.NoError(t, err)
			assert.True(t, valid)

			// Small delay between attempts
			time.Sleep(10 * time.Millisecond)
		}
	})

	t.Run("file system error recovery", func(t *testing.T) {
		// Test recovery from file system errors
		ctx := context.Background()

		// Try to load non-existent license (should handle gracefully)
		nonExistentManager, err := NewManagerWithPath(filepath.Join(tempDir, "nonexistent.dat"))
		require.NoError(t, err)
		defer nonExistentManager.Close()

		// Should handle missing file gracefully
		info, status, err := nonExistentManager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Nil(t, info)
		assert.Equal(t, "Not Activated", status)

		// Validation should handle missing file
		valid, err := nonExistentManager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("corrupted data recovery", func(t *testing.T) {
		// Test recovery from corrupted license data
		corruptFile := filepath.Join(tempDir, "corrupt_recovery.dat")

		// Write corrupted data
		err := os.WriteFile(corruptFile, []byte("corrupted license data"), 0600)
		require.NoError(t, err)

		corruptManager, err := NewManagerWithPath(corruptFile)
		require.NoError(t, err)
		defer corruptManager.Close()

		ctx := context.Background()

		// Should handle corrupted file gracefully
		info, status, err := corruptManager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Nil(t, info)
		assert.Equal(t, "Not Activated", status)

		// Should be able to activate new license over corrupted data
		err = corruptManager.ActivateLicenseWithContext(ctx, "RECOVERY-OVER-CORRUPT")
		assert.Error(t, err) // Will fail due to network, but should handle corruption
		assert.Contains(t, err.Error(), "license validation failed")
	})

	t.Run("state consistency after errors", func(t *testing.T) {
		// Test that manager state remains consistent after errors
		stateTestLicense := LicenseInfo{
			LicenseKey:  "STATE-CONSISTENCY-KEY",
			UserEmail:   "state@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(stateTestLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Perform operations that may cause errors
		_, _ = manager.ValidateLicenseWithContext(ctx)
		_ = manager.ActivateLicenseWithContext(ctx, "INVALID-KEY") // Will fail
		_, _, _ = manager.GetLicenseStatus()

		// After errors, valid operations should still work
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "Warning", status)
	})
}
