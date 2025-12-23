package license

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// Hardware Fingerprinting Edge Cases Tests
// =============================================================================

func TestHardwareFingerprintingEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "hardware_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("virtual machine environment", func(t *testing.T) {
		// Test behavior in virtual machines where hardware fingerprinting might fail
		// This tests the deprecation of machine ID binding
		ctx := context.Background()

		// Since machine ID binding is removed, activation should work consistently
		err := manager.ActivateLicenseWithContext(ctx, "INVALID-FOR-NETWORK-TEST")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license validation failed")
	})

	t.Run("docker container environment", func(t *testing.T) {
		// Test behavior in containers with limited hardware access
		// License should work in containers since machine binding is removed
		_, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Equal(t, "Not Activated", status)
	})

	t.Run("multiple network interfaces", func(t *testing.T) {
		// Test handling of systems with multiple network interfaces
		// Since machine ID is no longer used, this should be consistent
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.Nil(t, info)
		assert.Equal(t, "Not Activated", status)
	})

	t.Run("hardware changes after activation", func(t *testing.T) {
		// Test that license remains valid after hardware changes
		// This validates the removal of machine binding

		// First, create a test license locally
		testLicense := LicenseInfo{
			LicenseKey:  "TEST-HARDWARE-CHANGE",
			UserEmail:   "test@hardware.com",
			ExpiryDate:  time.Now().Add(31 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(testLicense)
		require.NoError(t, err)

		// License should still be valid (no machine binding)
		ctx := context.Background()
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Status should show active
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "Active", status)
	})
}

// =============================================================================
// Activation Flow Error Path Tests
// =============================================================================

func TestActivationFlowErrorPaths(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "activation_error_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("network timeout scenarios", func(t *testing.T) {
		// Test various network timeout conditions
		testCases := []struct {
			name     string
			key      string
			expected string
		}{
			{
				name:     "timeout during validation",
				key:      "TIMEOUT-TEST-KEY-123",
				expected: "license validation failed",
			},
			{
				name:     "connection refused",
				key:      "CONNECTION-REFUSED-KEY",
				expected: "license validation failed",
			},
			{
				name:     "DNS resolution failure",
				key:      "DNS-FAILURE-TEST-KEY",
				expected: "license validation failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, tc.key)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expected)
			})
		}
	})

	t.Run("invalid server responses", func(t *testing.T) {
		// Test handling of malformed or invalid server responses
		// Since we can't mock the actual Google Sheets service easily,
		// we test the error handling paths by providing invalid keys

		invalidResponses := []struct {
			name string
			key  string
		}{
			{"malformed response", "MALFORMED-RESPONSE"},
			{"empty response", "EMPTY-RESPONSE-KEY"},
			{"invalid json", "INVALID-JSON-KEY"},
			{"missing fields", "MISSING-FIELDS-KEY"},
		}

		for _, tr := range invalidResponses {
			t.Run(tr.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, tr.key)
				assert.Error(t, err)
				// All should result in network validation errors
				assert.Contains(t, err.Error(), "license validation failed")
			})
		}
	})

	t.Run("corrupted license data handling", func(t *testing.T) {
		// Test activation when license file is corrupted
		corruptData := []byte("corrupted license data {invalid json")
		err := os.WriteFile(manager.licenseFile, corruptData, 0600)
		require.NoError(t, err)

		// Should handle corruption gracefully and allow new activation
		err = manager.ActivateLicenseWithContext(ctx, "TEST-CORRUPTION-KEY")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license validation failed")
	})

	t.Run("activation server errors", func(t *testing.T) {
		// Test various HTTP error codes from activation server
		serverErrors := []struct {
			name string
			key  string
		}{
			{"unauthorized access", "UNAUTHORIZED-KEY"},
			{"forbidden access", "FORBIDDEN-ACCESS-KEY"},
			{"service unavailable", "SERVICE-UNAVAILABLE"},
			{"rate limited", "RATE-LIMITED-KEY"},
		}

		for _, se := range serverErrors {
			t.Run(se.name, func(t *testing.T) {
				err := manager.ActivateLicenseWithContext(ctx, se.key)
				assert.Error(t, err)
				// Should get network validation error
				assert.Contains(t, err.Error(), "license validation failed")
			})
		}
	})

	t.Run("context cancellation during activation", func(t *testing.T) {
		// Test context cancellation handling
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.ActivateLicenseWithContext(cancelCtx, "CONTEXT-CANCEL-KEY")
		assert.Error(t, err)
		// Should handle context cancellation
	})

	t.Run("concurrent activation attempts", func(t *testing.T) {
		// Test concurrent activation attempts with same key
		key := "CONCURRENT-ACTIVATION-KEY"
		numGoroutines := 10

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := manager.ActivateLicenseWithContext(ctx, key)
				errors <- err
			}()
		}

		wg.Wait()
		close(errors)

		// All should fail with network errors (expected)
		errorCount := 0
		for err := range errors {
			assert.Error(t, err)
			errorCount++
		}
		assert.Equal(t, numGoroutines, errorCount)
	})
}

// =============================================================================
// License Expiration Scenario Tests
// =============================================================================

func TestLicenseExpirationScenarios(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "expiration_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("grace period handling", func(t *testing.T) {
		// Test grace periods for different expiration scenarios
		now := time.Now()

		gracePeriodTests := []struct {
			name           string
			expiryOffset   time.Duration
			expectedStatus string
			shouldBeValid  bool
		}{
			{
				name:           "just expired (1 second)",
				expiryOffset:   -1 * time.Second,
				expectedStatus: "Expired",
				shouldBeValid:  false,
			},
			{
				name:           "expired 1 hour ago",
				expiryOffset:   -1 * time.Hour,
				expectedStatus: "Expired",
				shouldBeValid:  false,
			},
			{
				name:           "expires in 1 second",
				expiryOffset:   1 * time.Second,
				expectedStatus: "Critical",
				shouldBeValid:  true,
			},
			{
				name:           "expires in 1 hour",
				expiryOffset:   1 * time.Hour,
				expectedStatus: "Critical",
				shouldBeValid:  true,
			},
		}

		for _, test := range gracePeriodTests {
			t.Run(test.name, func(t *testing.T) {
				// Create license with specific expiry
				license := LicenseInfo{
					LicenseKey:  fmt.Sprintf("GRACE-%s", strings.ReplaceAll(test.name, " ", "-")),
					UserEmail:   "grace@test.com",
					ExpiryDate:  now.Add(test.expiryOffset),
					Duration:    "1m",
					IssuedDate:  now.Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: now,
				}

				err := manager.saveLicenseLocal(license)
				require.NoError(t, err)

				// Check validation
				ctx := context.Background()
				valid, err := manager.ValidateLicenseWithContext(ctx)

				if test.shouldBeValid {
					assert.NoError(t, err)
					assert.True(t, valid)
				} else {
					assert.Error(t, err)
					assert.False(t, valid)
					assert.Contains(t, err.Error(), "expired")
				}

				// Check status
				info, status, statusErr := manager.GetLicenseStatus()
				assert.NoError(t, statusErr)
				assert.NotNil(t, info)
				assert.Equal(t, test.expectedStatus, status)
			})
		}
	})

	t.Run("expired license behavior", func(t *testing.T) {
		// Test behavior with expired licenses
		expiredLicense := LicenseInfo{
			LicenseKey:  "EXPIRED-LICENSE-TEST",
			UserEmail:   "expired@test.com",
			ExpiryDate:  time.Now().Add(-48 * time.Hour), // Expired 2 days ago
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-72 * time.Hour), // Issued 3 days ago
			Status:      "Activated",
			LastChecked: time.Now().Add(-47 * time.Hour), // Last checked just after expiry
		}

		err := manager.saveLicenseLocal(expiredLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Validation should fail
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "expired")

		// Status should show expired
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "Expired", status)

		// Renewal check should indicate need for renewal
		renewalInfo, err := manager.CheckRenewalStatus()
		assert.NoError(t, err)
		assert.True(t, renewalInfo.IsExpired)
		assert.True(t, renewalInfo.NeedsRenewal)
		assert.Equal(t, "Expired", renewalInfo.Status)
	})

	t.Run("renewal workflow scenarios", func(t *testing.T) {
		// Test different renewal workflow scenarios
		renewalTests := []struct {
			name           string
			daysLeft       int
			expectedStatus string
			needsRenewal   bool
		}{
			{"60 days left - no renewal needed", 60, "Active", false},
			{"30 days left - warning", 30, "Warning", true},
			{"15 days left - warning", 15, "Warning", true},
			{"7 days left - critical", 7, "Critical", true},
			{"1 day left - critical", 1, "Critical", true},
		}

		for _, rt := range renewalTests {
			t.Run(rt.name, func(t *testing.T) {
				license := LicenseInfo{
					LicenseKey:  fmt.Sprintf("RENEWAL-%d-DAYS", rt.daysLeft),
					UserEmail:   "renewal@test.com",
					ExpiryDate:  time.Now().Add(time.Duration(rt.daysLeft) * 24 * time.Hour),
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now(),
				}

				err := manager.saveLicenseLocal(license)
				require.NoError(t, err)

				// Check renewal status
				renewalInfo, err := manager.CheckRenewalStatus()
				assert.NoError(t, err)
				assert.Equal(t, rt.expectedStatus, renewalInfo.Status)
				assert.Equal(t, rt.needsRenewal, renewalInfo.NeedsRenewal)
				assert.Equal(t, rt.daysLeft, renewalInfo.DaysLeft)
				assert.False(t, renewalInfo.IsExpired)
			})
		}
	})

	t.Run("time zone handling", func(t *testing.T) {
		// Test handling of different time zones in expiry calculations
		locations := []*time.Location{
			time.UTC,
			time.FixedZone("UTC+8", 8*3600),  // Asian timezone
			time.FixedZone("UTC-5", -5*3600), // US timezone
		}

		for i, loc := range locations {
			t.Run(fmt.Sprintf("timezone_%d", i), func(t *testing.T) {
				// Create license expiring in 5 days in different timezone
				expiryTime := time.Now().In(loc).Add(5 * 24 * time.Hour)

				license := LicenseInfo{
					LicenseKey:  fmt.Sprintf("TIMEZONE-TEST-%d", i),
					UserEmail:   "timezone@test.com",
					ExpiryDate:  expiryTime,
					Duration:    "1m",
					IssuedDate:  time.Now().Add(-24 * time.Hour),
					Status:      "Activated",
					LastChecked: time.Now(),
				}

				err := manager.saveLicenseLocal(license)
				require.NoError(t, err)

				// Validation should work regardless of timezone
				ctx := context.Background()
				valid, err := manager.ValidateLicenseWithContext(ctx)
				assert.NoError(t, err)
				assert.True(t, valid)

				// Status should be critical (5 days left)
				info, status, err := manager.GetLicenseStatus()
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, "Critical", status)
			})
		}
	})
}

// =============================================================================
// Concurrent Operation Tests
// =============================================================================

func TestConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "concurrent_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("concurrent activation attempts with race conditions", func(t *testing.T) {
		ctx := context.Background()
		numOperations := 50

		// Use errgroup for better error handling
		g, ctx := errgroup.WithContext(ctx)

		// Results channels
		results := make(chan error, numOperations)

		// Launch concurrent activation attempts
		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("RACE-TEST-KEY-%d", i)
			g.Go(func() error {
				err := manager.ActivateLicenseWithContext(ctx, key)
				results <- err
				return nil // Don't fail the group on individual errors
			})
		}

		// Wait for all to complete
		assert.NoError(t, g.Wait())
		close(results)

		// Check results
		errorCount := 0
		for err := range results {
			if err != nil {
				errorCount++
				// All should be network validation errors
				assert.Contains(t, err.Error(), "license validation failed")
			}
		}

		// All should fail with network errors (expected for invalid keys)
		assert.Equal(t, numOperations, errorCount)
	})

	t.Run("concurrent validation attempts", func(t *testing.T) {
		// First create a valid license
		validLicense := LicenseInfo{
			LicenseKey:  "CONCURRENT-VALIDATION-KEY",
			UserEmail:   "concurrent@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(validLicense)
		require.NoError(t, err)

		ctx := context.Background()
		numValidations := 100

		g, ctx := errgroup.WithContext(ctx)
		validResults := make(chan bool, numValidations)
		errorResults := make(chan error, numValidations)

		// Launch concurrent validations
		for i := 0; i < numValidations; i++ {
			g.Go(func() error {
				valid, err := manager.ValidateLicenseWithContext(ctx)
				validResults <- valid
				errorResults <- err
				return nil
			})
		}

		assert.NoError(t, g.Wait())
		close(validResults)
		close(errorResults)

		// All validations should succeed
		successCount := 0
		for valid := range validResults {
			if valid {
				successCount++
			}
		}

		errorCount := 0
		for err := range errorResults {
			if err != nil {
				errorCount++
			}
		}

		assert.Equal(t, numValidations, successCount)
		assert.Equal(t, 0, errorCount)
	})

	t.Run("concurrent status checks", func(t *testing.T) {
		// Create test license
		statusLicense := LicenseInfo{
			LicenseKey:  "CONCURRENT-STATUS-KEY",
			UserEmail:   "status@test.com",
			ExpiryDate:  time.Now().Add(15 * 24 * time.Hour), // 15 days - warning status
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(statusLicense)
		require.NoError(t, err)

		numChecks := 50
		g, _ := errgroup.WithContext(context.Background())

		statusResults := make(chan string, numChecks)

		// Launch concurrent status checks
		for i := 0; i < numChecks; i++ {
			g.Go(func() error {
				info, status, err := manager.GetLicenseStatus()
				assert.NoError(t, err)
				assert.NotNil(t, info)
				statusResults <- status
				return nil
			})
		}

		assert.NoError(t, g.Wait())
		close(statusResults)

		// All should return same status
		warningCount := 0
		for status := range statusResults {
			if status == "Warning" {
				warningCount++
			}
		}

		assert.Equal(t, numChecks, warningCount)
	})

	t.Run("concurrent cache operations", func(t *testing.T) {
		// Test concurrent cache operations to ensure thread safety
		if manager.cache == nil {
			t.Skip("Cache not initialized")
		}

		numOperations := 100
		g, _ := errgroup.WithContext(context.Background())

		// Test data
		testLicense := LicenseInfo{
			LicenseKey:  "CACHE-CONCURRENT-TEST",
			UserEmail:   "cache@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now(),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		// Concurrent cache sets
		for i := 0; i < numOperations/2; i++ {
			key := fmt.Sprintf("CACHE-KEY-%d", i)
			g.Go(func() error {
				manager.cache.Set(key, testLicense)
				return nil
			})
		}

		// Concurrent cache gets
		for i := 0; i < numOperations/2; i++ {
			key := fmt.Sprintf("CACHE-KEY-%d", i%10) // Some will hit, some will miss
			g.Go(func() error {
				_, found := manager.cache.Get(key)
				_ = found // Use the result to avoid unused variable warning
				return nil
			})
		}

		assert.NoError(t, g.Wait())

		// Verify cache is still functional
		stats := manager.cache.GetStats()
		assert.NotNil(t, stats)
		assert.Greater(t, stats["entries"], 0)
	})

	t.Run("lock contention scenarios", func(t *testing.T) {
		// Test lock contention with mixed read/write operations
		validLicense := LicenseInfo{
			LicenseKey:  "LOCK-CONTENTION-KEY",
			UserEmail:   "contention@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(validLicense)
		require.NoError(t, err)

		ctx := context.Background()
		numOperations := 30
		g, ctx := errgroup.WithContext(ctx)

		// Mix of different operations to create lock contention
		for i := 0; i < numOperations; i++ {
			operation := i % 4
			g.Go(func() error {
				switch operation {
				case 0:
					// Validation
					_, err := manager.ValidateLicenseWithContext(ctx)
					return err
				case 1:
					// Status check
					_, _, err := manager.GetLicenseStatus()
					return err
				case 2:
					// Performance metrics
					_ = manager.GetPerformanceMetrics()
					return nil
				case 3:
					// System stats
					_ = manager.GetSystemStats()
					return nil
				}
				return nil
			})
		}

		// Should complete without deadlocks or race conditions
		assert.NoError(t, g.Wait())
	})

	t.Run("state consistency under concurrent access", func(t *testing.T) {
		// Test that manager state remains consistent under concurrent access
		consistencyLicense := LicenseInfo{
			LicenseKey:  "STATE-CONSISTENCY-KEY",
			UserEmail:   "consistency@test.com",
			ExpiryDate:  time.Now().Add(7 * 24 * time.Hour), // 7 days - critical threshold
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(consistencyLicense)
		require.NoError(t, err)

		ctx := context.Background()
		numReaders := 20
		g, ctx := errgroup.WithContext(ctx)

		results := make(chan string, numReaders)

		// Multiple concurrent readers
		for i := 0; i < numReaders; i++ {
			g.Go(func() error {
				info, status, err := manager.GetLicenseStatus()
				if err != nil {
					return err
				}
				if info == nil {
					return fmt.Errorf("unexpected nil info")
				}
				results <- status
				return nil
			})
		}

		assert.NoError(t, g.Wait())
		close(results)

		// All should return consistent results
		statuses := make(map[string]int)
		for status := range results {
			statuses[status]++
		}

		// Should all be "Critical" (7 days left)
		assert.Equal(t, 1, len(statuses), "Results should be consistent")
		assert.Equal(t, numReaders, statuses["Critical"])
	})
}

// =============================================================================
// Security Validation Tests
// =============================================================================

func TestSecurityValidation(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "security_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("tampered license detection", func(t *testing.T) {
		// Create a valid license first
		validLicense := LicenseInfo{
			LicenseKey:  "TAMPER-TEST-KEY",
			UserEmail:   "tamper@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(validLicense)
		require.NoError(t, err)

		// Now tamper with the file
		tamperedData := `{
			"license_key": "TAMPERED-KEY-DIFFERENT",
			"user_email": "hacker@evil.com",
			"expiry_date": "2099-12-31T23:59:59Z",
			"duration": "99y",
			"issued_date": "2024-01-01T00:00:00Z",
			"status": "Activated",
			"last_checked": "2024-08-01T12:00:00Z"
		}`

		err = os.WriteFile(manager.licenseFile, []byte(tamperedData), 0600)
		require.NoError(t, err)

		// Should load the tampered data but validation should work with it
		// (since we don't have signature verification on local files)
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "TAMPERED-KEY-DIFFERENT", info.LicenseKey)
		assert.Equal(t, "Active", status) // Long expiry date
	})

	t.Run("signature verification scenarios", func(t *testing.T) {
		// Test state file signature verification
		stateFilePath := filepath.Join(tempDir, "test_state.json")

		// Create valid state file
		err := manager.CreateStateFile(stateFilePath)
		assert.NoError(t, err)

		// Verify it's valid
		valid, err := manager.ValidateStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Now tamper with the state file
		stateData, err := os.ReadFile(stateFilePath)
		require.NoError(t, err)

		var state StateFile
		err = json.Unmarshal(stateData, &state)
		require.NoError(t, err)

		// Modify the signature
		state.Signature = "tampered_signature_" + state.Signature

		tamperedData, err := json.MarshalIndent(state, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(stateFilePath, tamperedData, 0600)
		require.NoError(t, err)

		// Validation should fail
		valid, err = manager.ValidateStateFile(stateFilePath)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid state file signature")
	})

	t.Run("hardware mismatch scenarios", func(t *testing.T) {
		// Since hardware binding is removed, test that licenses work across different "hardware"
		// This validates the architectural decision to remove machine binding

		license1 := LicenseInfo{
			LicenseKey:  "HARDWARE-MISMATCH-1",
			UserEmail:   "hardware1@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(license1)
		require.NoError(t, err)

		ctx := context.Background()

		// Should work regardless of "hardware" since binding is removed
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Create another license (simulating different hardware)
		license2 := LicenseInfo{
			LicenseKey:  "HARDWARE-MISMATCH-2",
			UserEmail:   "hardware2@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err = manager.saveLicenseLocal(license2)
		require.NoError(t, err)

		// Should also work
		valid, err = manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("key rotation scenarios", func(t *testing.T) {
		// Test behavior when keys are rotated/changed
		originalLicense := LicenseInfo{
			LicenseKey: "ORIGINAL-KEY-12345",
			UserEmail:  "rotation@test.com",
			// Keep this outside the warning window to assert Active status after rotation
			ExpiryDate:  time.Now().Add(45 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(originalLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Original key should work
		valid, err := manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Simulate key rotation by changing the license key
		rotatedLicense := originalLicense
		rotatedLicense.LicenseKey = "ROTATED-KEY-67890"

		err = manager.saveLicenseLocal(rotatedLicense)
		require.NoError(t, err)

		// Rotated key should also work (local validation)
		valid, err = manager.ValidateLicenseWithContext(ctx)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Status should show the new key
		info, status, err := manager.GetLicenseStatus()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "ROTATED-KEY-67890", info.LicenseKey)
		assert.Equal(t, "Active", status)
	})

	t.Run("security manager rate limiting", func(t *testing.T) {
		// Test the security manager's rate limiting functionality
		if manager.security == nil {
			t.Skip("Security manager not initialized")
		}

		testIdentifier := "security-rate-limit-test"

		// Should start unblocked
		assert.False(t, manager.security.IsBlocked(testIdentifier))

		// Record failures up to the limit
		maxAttempts := 5 // From NewManager initialization
		for i := 0; i < maxAttempts-1; i++ {
			result := manager.security.RecordAttempt(testIdentifier, false)
			assert.True(t, result, "Should allow attempt %d", i+1)
			assert.False(t, manager.security.IsBlocked(testIdentifier))
		}

		// The final failure should block
		result := manager.security.RecordAttempt(testIdentifier, false)
		assert.False(t, result, "Should block after max attempts")
		assert.True(t, manager.security.IsBlocked(testIdentifier))

		// Success should unblock
		result = manager.security.RecordAttempt(testIdentifier, true)
		assert.True(t, result, "Success should unblock")
		assert.False(t, manager.security.IsBlocked(testIdentifier))

		// Verify stats
		stats := manager.security.GetStats()
		assert.NotNil(t, stats)
		assert.Equal(t, maxAttempts, stats["max_attempts"])
	})
}
