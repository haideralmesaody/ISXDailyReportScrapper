package license

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// State Management Comprehensive Tests
// =============================================================================

func TestStateManagementComprehensive(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "state_comprehensive_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("state file lifecycle", func(t *testing.T) {
		stateFilePath := filepath.Join(tempDir, "lifecycle_test.json")

		// Initially no state file
		exists := func() bool {
			_, err := os.Stat(stateFilePath)
			return err == nil
		}

		assert.False(t, exists())

		// Create state file
		err := manager.CreateStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.True(t, exists())

		// Validate immediately after creation
		valid, err := manager.ValidateStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Cleanup
		err = CleanupStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.False(t, exists())

		// Cleanup non-existent file should not error
		err = CleanupStateFile(stateFilePath)
		assert.NoError(t, err)
	})

	t.Run("state file expiration timing", func(t *testing.T) {
		stateFilePath := filepath.Join(tempDir, "expiration_test.json")

		// Create state file
		err := manager.CreateStateFile(stateFilePath)
		require.NoError(t, err)

		// Should be valid immediately
		valid, err := manager.ValidateStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Read the state file to check timing
		data, err := os.ReadFile(stateFilePath)
		require.NoError(t, err)

		var state StateFile
		err = json.Unmarshal(data, &state)
		require.NoError(t, err)

		// Should expire in approximately 5 minutes
		expectedExpiry := time.Now().Add(5 * time.Minute)
		timeDiff := state.ValidUntil.Sub(expectedExpiry)
		assert.Less(t, timeDiff, 10*time.Second, "Expiry time should be close to expected")
		assert.Greater(t, timeDiff, -10*time.Second, "Expiry time should be close to expected")

		// Test edge case: file created in past should be invalid
		pastState := StateFile{
			ValidatedAt: time.Now().Add(-10 * time.Minute),
			ValidUntil:  time.Now().Add(-5 * time.Minute), // Expired 5 minutes ago
		}
		pastState.Signature = generateStateSignature(pastState)

		pastData, err := json.MarshalIndent(pastState, "", "  ")
		require.NoError(t, err)

		pastFilePath := filepath.Join(tempDir, "past_state.json")
		err = os.WriteFile(pastFilePath, pastData, 0600)
		require.NoError(t, err)

		// Should be invalid due to expiration
		valid, err = manager.ValidateStateFile(pastFilePath)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("concurrent state file operations", func(t *testing.T) {
		numOperations := 20
		g, _ := errgroup.WithContext(context.Background())

		stateFiles := make([]string, numOperations)
		for i := 0; i < numOperations; i++ {
			stateFiles[i] = filepath.Join(tempDir, fmt.Sprintf("concurrent_state_%d.json", i))
		}

		// Concurrent state file creation
		for i := 0; i < numOperations; i++ {
			filePath := stateFiles[i]
			g.Go(func() error {
				return manager.CreateStateFile(filePath)
			})
		}

		assert.NoError(t, g.Wait())

		// All files should exist and be valid
		g, _ = errgroup.WithContext(context.Background())
		results := make(chan bool, numOperations)

		for i := 0; i < numOperations; i++ {
			filePath := stateFiles[i]
			g.Go(func() error {
				valid, err := manager.ValidateStateFile(filePath)
				results <- valid && err == nil
				return nil
			})
		}

		assert.NoError(t, g.Wait())
		close(results)

		validCount := 0
		for valid := range results {
			if valid {
				validCount++
			}
		}

		assert.Equal(t, numOperations, validCount)

		// Cleanup all files
		for _, filePath := range stateFiles {
			CleanupStateFile(filePath)
		}
	})

	t.Run("state file corruption scenarios", func(t *testing.T) {
		corruptionTests := []struct {
			name    string
			content string
			valid   bool
		}{
			{
				name:    "invalid JSON",
				content: `{invalid json`,
				valid:   false,
			},
			{
				name:    "missing signature",
				content: `{"validated_at":"2024-08-01T12:00:00Z","valid_until":"2024-08-01T12:05:00Z"}`,
				valid:   false,
			},
			{
				name:    "missing timestamps",
				content: `{"signature":"test"}`,
				valid:   false,
			},
			{
				name:    "empty file",
				content: ``,
				valid:   false,
			},
			{
				name:    "binary data",
				content: string([]byte{0x00, 0x01, 0x02, 0x03}),
				valid:   false,
			},
		}

		for _, test := range corruptionTests {
			t.Run(test.name, func(t *testing.T) {
				corruptFilePath := filepath.Join(tempDir, fmt.Sprintf("corrupt_%s.json", test.name))

				err := os.WriteFile(corruptFilePath, []byte(test.content), 0600)
				require.NoError(t, err)

				valid, err := manager.ValidateStateFile(corruptFilePath)

				if test.valid {
					assert.NoError(t, err)
					assert.True(t, valid)
				} else {
					// Should either return false or error, but not crash
					if err != nil {
						assert.Error(t, err)
					} else {
						assert.False(t, valid)
					}
				}
			})
		}
	})

	t.Run("signature generation consistency", func(t *testing.T) {
		// Test that signature generation is consistent
		now := time.Now()
		state1 := StateFile{
			ValidatedAt: now,
			ValidUntil:  now.Add(5 * time.Minute),
		}

		state2 := StateFile{
			ValidatedAt: now,
			ValidUntil:  now.Add(5 * time.Minute),
		}

		sig1 := generateStateSignature(state1)
		sig2 := generateStateSignature(state2)

		assert.Equal(t, sig1, sig2, "Same state should produce same signature")
		assert.NotEmpty(t, sig1, "Signature should not be empty")
		assert.Len(t, sig1, 64, "SHA256 hex should be 64 characters")

		// Different state should produce different signature
		state3 := StateFile{
			ValidatedAt: now.Add(1 * time.Second),
			ValidUntil:  now.Add(5 * time.Minute),
		}

		sig3 := generateStateSignature(state3)
		assert.NotEqual(t, sig1, sig3, "Different state should produce different signature")
	})

	t.Run("machine ID deprecation validation", func(t *testing.T) {
		// Test that machine ID functionality is properly deprecated
		machineID := manager.GetMachineID()
		assert.Empty(t, machineID, "Machine ID should be empty (deprecated)")

		// State files should work without machine ID validation
		stateFilePath := filepath.Join(tempDir, "no_machine_id.json")

		err := manager.CreateStateFile(stateFilePath)
		assert.NoError(t, err)

		valid, err := manager.ValidateStateFile(stateFilePath)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Read and verify no machine ID in state file
		data, err := os.ReadFile(stateFilePath)
		require.NoError(t, err)

		var state StateFile
		err = json.Unmarshal(data, &state)
		require.NoError(t, err)

		// State file should only contain time fields and signature
		assert.False(t, state.ValidatedAt.IsZero())
		assert.False(t, state.ValidUntil.IsZero())
		assert.NotEmpty(t, state.Signature)
	})

	t.Run("state file permissions", func(t *testing.T) {
		if os.Getenv("CI") == "true" || runtime.GOOS == "windows" {
			t.Skip("Skipping permission tests in CI or Windows environment")
		}

		stateFilePath := filepath.Join(tempDir, "permissions_test.json")

		err := manager.CreateStateFile(stateFilePath)
		require.NoError(t, err)

		// Check file permissions (should be 0600)
		fileInfo, err := os.Stat(stateFilePath)
		require.NoError(t, err)

		mode := fileInfo.Mode()
		assert.Equal(t, os.FileMode(0600), mode.Perm(), "State file should have 0600 permissions")
	})

	t.Run("state file validation edge cases", func(t *testing.T) {
		edgeCases := []struct {
			name        string
			setupState  func() StateFile
			shouldError bool
			shouldValid bool
		}{
			{
				name: "future validation time",
				setupState: func() StateFile {
					return StateFile{
						ValidatedAt: time.Now().Add(1 * time.Hour), // Future
						ValidUntil:  time.Now().Add(2 * time.Hour),
					}
				},
				shouldError: false,
				shouldValid: false, // Should be invalid due to future validation time
			},
			{
				name: "validation after expiry",
				setupState: func() StateFile {
					return StateFile{
						ValidatedAt: time.Now().Add(-10 * time.Minute),
						ValidUntil:  time.Now().Add(-5 * time.Minute), // Already expired
					}
				},
				shouldError: false,
				shouldValid: false,
			},
			{
				name: "zero timestamps",
				setupState: func() StateFile {
					return StateFile{
						ValidatedAt: time.Time{},
						ValidUntil:  time.Time{},
					}
				},
				shouldError: false,
				shouldValid: false,
			},
			{
				name: "very short validity period",
				setupState: func() StateFile {
					now := time.Now()
					return StateFile{
						ValidatedAt: now,
						ValidUntil:  now.Add(1 * time.Millisecond),
					}
				},
				shouldError: false,
				shouldValid: false, // Will likely be expired by the time we check
			},
		}

		for _, test := range edgeCases {
			t.Run(test.name, func(t *testing.T) {
				state := test.setupState()
				state.Signature = generateStateSignature(state)

				data, err := json.MarshalIndent(state, "", "  ")
				require.NoError(t, err)

				edgeFilePath := filepath.Join(tempDir, fmt.Sprintf("edge_%s.json", test.name))
				err = os.WriteFile(edgeFilePath, data, 0600)
				require.NoError(t, err)

				valid, err := manager.ValidateStateFile(edgeFilePath)

				if test.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, test.shouldValid, valid)
			})
		}
	})

	t.Run("state file race conditions", func(t *testing.T) {
		// Test for race conditions in state file operations
		stateFilePath := filepath.Join(tempDir, "race_test.json")

		numOperations := 50
		var wg sync.WaitGroup

		// Concurrent create and validate operations
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()

				if iteration%2 == 0 {
					// Create state file
					err := manager.CreateStateFile(stateFilePath)
					assert.NoError(t, err)
				} else {
					// Validate state file (may or may not exist)
					valid, err := manager.ValidateStateFile(stateFilePath)
					// Don't assert on results since file may not exist
					_ = valid
					_ = err
				}
			}(i)
		}

		wg.Wait()

		// Final validation should work if file exists
		if _, err := os.Stat(stateFilePath); err == nil {
			valid, err := manager.ValidateStateFile(stateFilePath)
			assert.NoError(t, err)
			assert.True(t, valid)
		}
	})
}

// =============================================================================
// Performance and Benchmark Tests
// =============================================================================

func TestPerformanceMetrics(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "performance_test.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("performance tracking accuracy", func(t *testing.T) {
		// Create a test license for performance testing
		testLicense := LicenseInfo{
			LicenseKey:  "PERFORMANCE-TEST-KEY",
			UserEmail:   "performance@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(testLicense)
		require.NoError(t, err)

		ctx := context.Background()

		// Perform multiple operations to generate metrics
		numOperations := 10
		for i := 0; i < numOperations; i++ {
			valid, err := manager.ValidateLicenseWithContext(ctx)
			assert.NoError(t, err)
			assert.True(t, valid)
		}

		// Check performance metrics
		metrics := manager.GetPerformanceMetrics()
		assert.NotNil(t, metrics)

		// Should have validation metrics
		if validationMetric, exists := metrics["license_validation_complete"]; exists {
			assert.Greater(t, validationMetric.Count, int64(0))
			assert.Greater(t, validationMetric.SuccessCount, int64(0))
			assert.Equal(t, int64(0), validationMetric.ErrorCount)
			assert.Greater(t, validationMetric.TotalTime, time.Duration(0))
			assert.Greater(t, validationMetric.AverageTime, time.Duration(0))
			assert.GreaterOrEqual(t, validationMetric.MaxTime, validationMetric.MinTime)
		}
	})

	t.Run("system statistics collection", func(t *testing.T) {
		stats := manager.GetSystemStats()
		assert.NotNil(t, stats)

		// Should contain expected sections
		assert.Contains(t, stats, "performance")
		assert.Contains(t, stats, "timestamp")
		assert.Contains(t, stats, "version")

		// Performance section should be a map
		if perfStats, ok := stats["performance"].(map[string]*PerformanceMetrics); ok {
			assert.NotNil(t, perfStats)
		}

		// Timestamp should be recent
		if timestamp, ok := stats["timestamp"].(time.Time); ok {
			assert.True(t, time.Since(timestamp) < 1*time.Second)
		}

		// Version should be set
		if version, ok := stats["version"].(string); ok {
			assert.NotEmpty(t, version)
		}

		// Cache stats should be included if cache exists
		if manager.cache != nil {
			assert.Contains(t, stats, "cache")
		}

		// Security stats should be included if security manager exists
		if manager.security != nil {
			assert.Contains(t, stats, "security")
		}
	})

	t.Run("metric collection under load", func(t *testing.T) {
		// Create test license
		loadTestLicense := LicenseInfo{
			LicenseKey:  "LOAD-TEST-KEY",
			UserEmail:   "load@test.com",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			Status:      "Activated",
			LastChecked: time.Now(),
		}

		err := manager.saveLicenseLocal(loadTestLicense)
		require.NoError(t, err)

		ctx := context.Background()
		numOperations := 100

		start := time.Now()

		// Perform many operations concurrently
		g, ctx := errgroup.WithContext(ctx)
		for i := 0; i < numOperations; i++ {
			g.Go(func() error {
				// Mix of different operations
				_, err := manager.ValidateLicenseWithContext(ctx)
				if err != nil {
					return err
				}

				_, _, err = manager.GetLicenseStatus()
				if err != nil {
					return err
				}

				_ = manager.GetPerformanceMetrics()
				_ = manager.GetSystemStats()

				return nil
			})
		}

		assert.NoError(t, g.Wait())

		totalTime := time.Since(start)
		t.Logf("Completed %d operations in %v", numOperations*4, totalTime) // 4 ops per iteration

		// Verify metrics were collected properly
		finalMetrics := manager.GetPerformanceMetrics()
		assert.NotNil(t, finalMetrics)

		// Should have significant operation counts
		totalOperations := int64(0)
		for _, metric := range finalMetrics {
			totalOperations += metric.Count
		}

		// Allow generous slack because validation/status paths may short-circuit or hit cache
		assert.Greater(t, totalOperations, int64(numOperations/10))
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkLicenseOperations(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "benchmark.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	// Setup test license
	testLicense := LicenseInfo{
		LicenseKey:  "BENCHMARK-TEST-KEY",
		UserEmail:   "benchmark@test.com",
		ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
		Duration:    "1m",
		IssuedDate:  time.Now().Add(-24 * time.Hour),
		Status:      "Activated",
		LastChecked: time.Now(),
	}

	err = manager.saveLicenseLocal(testLicense)
	require.NoError(b, err)

	b.Run("ValidateLicense", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := manager.ValidateLicenseWithContext(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetLicenseStatus", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, err := manager.GetLicenseStatus()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LoadLicenseLocal", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := manager.loadLicenseLocal()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SaveLicenseLocal", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := manager.saveLicenseLocal(testLicense)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CacheOperations", func(b *testing.B) {
		if manager.cache == nil {
			b.Skip("Cache not available")
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("benchmark-key-%d", i%100) // Cycle through 100 keys

			// Set and get operations
			manager.cache.Set(key, testLicense)
			_, _ = manager.cache.Get(key)
		}
	})

	b.Run("StateFileOperations", func(b *testing.B) {
		stateDir := b.TempDir()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			stateFile := filepath.Join(stateDir, fmt.Sprintf("state_%d.json", i))

			err := manager.CreateStateFile(stateFile)
			if err != nil {
				b.Fatal(err)
			}

			_, err = manager.ValidateStateFile(stateFile)
			if err != nil {
				b.Fatal(err)
			}

			CleanupStateFile(stateFile)
		}
	})
}

func BenchmarkConcurrentOperations(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "concurrent_benchmark.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	// Setup test license
	testLicense := LicenseInfo{
		LicenseKey:  "CONCURRENT-BENCHMARK-KEY",
		UserEmail:   "concurrent@test.com",
		ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
		Duration:    "1m",
		IssuedDate:  time.Now().Add(-24 * time.Hour),
		Status:      "Activated",
		LastChecked: time.Now(),
	}

	err = manager.saveLicenseLocal(testLicense)
	require.NoError(b, err)

	b.Run("ConcurrentValidation", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := manager.ValidateLicenseWithContext(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("ConcurrentStatusCheck", func(b *testing.B) {
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _, err := manager.GetLicenseStatus()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("ConcurrentCacheAccess", func(b *testing.B) {
		if manager.cache == nil {
			b.Skip("Cache not available")
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("concurrent-key-%d", i%50)

				if i%2 == 0 {
					manager.cache.Set(key, testLicense)
				} else {
					_, _ = manager.cache.Get(key)
				}
				i++
			}
		})
	})
}
