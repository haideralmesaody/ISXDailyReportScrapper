package license

import (
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
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// State File Management Tests
// =============================================================================

// StateFileTestSuite tests state file functionality
type StateFileTestSuite struct {
	suite.Suite
	tempDir     string
	licenseFile string
	stateFile   string
	manager     *Manager
}

func (suite *StateFileTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
	suite.licenseFile = filepath.Join(suite.tempDir, "test_license.dat")
	suite.stateFile = filepath.Join(suite.tempDir, "test_state.json")

	var err error
	suite.manager, err = NewManagerWithPath(suite.licenseFile)
	require.NoError(suite.T(), err)
}

func (suite *StateFileTestSuite) TearDownTest() {
	if suite.manager != nil {
		suite.manager.Close()
	}
}

// TestCreateStateFile tests state file creation
func (suite *StateFileTestSuite) TestCreateStateFile() {
	tests := []struct {
		name        string
		stateFile   string
		expectError bool
		setup       func()
		cleanup     func()
	}{
		{
			name:        "valid state file path",
			stateFile:   suite.stateFile,
			expectError: false,
			setup:       func() {},
			cleanup:     func() { os.Remove(suite.stateFile) },
		},
		{
			name:        "invalid directory path",
			stateFile:   "/nonexistent/directory/state.json",
			expectError: true,
			setup:       func() {},
			cleanup:     func() {},
		},
		{
			name:        "empty path",
			stateFile:   "",
			expectError: true,
			setup:       func() {},
			cleanup:     func() {},
		},
		{
			name:        "read-only directory",
			stateFile:   filepath.Join(suite.tempDir, "readonly", "state.json"),
			expectError: true,
			setup: func() {
				readOnlyDir := filepath.Join(suite.tempDir, "readonly")
				os.Mkdir(readOnlyDir, 0555) // Read-only directory
			},
			cleanup: func() {
				readOnlyDir := filepath.Join(suite.tempDir, "readonly")
				os.Chmod(readOnlyDir, 0755) // Make writable for cleanup
				os.RemoveAll(readOnlyDir)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup()
				}
			}()

			err := suite.manager.CreateStateFile(tt.stateFile)

			if tt.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)

				// Verify file was created
				_, err := os.Stat(tt.stateFile)
				suite.NoError(err)

				// Verify file content
				data, err := os.ReadFile(tt.stateFile)
				suite.NoError(err)

				var state StateFile
				err = json.Unmarshal(data, &state)
				suite.NoError(err)

				// Verify state file structure
				suite.False(state.ValidatedAt.IsZero())
				suite.False(state.ValidUntil.IsZero())
				suite.NotEmpty(state.Signature)
				suite.True(state.ValidUntil.After(state.ValidatedAt))

				// Should be valid for approximately 5 minutes
				duration := state.ValidUntil.Sub(state.ValidatedAt)
				suite.InDelta(5*time.Minute, duration, float64(time.Second))
			}
		})
	}
}

// TestValidateStateFile tests state file validation
func (suite *StateFileTestSuite) TestValidateStateFile() {
	tests := []struct {
		name        string
		setupState  func() string
		expectValid bool
		expectError bool
	}{
		{
			name: "valid state file",
			setupState: func() string {
				err := suite.manager.CreateStateFile(suite.stateFile)
				suite.NoError(err)
				return suite.stateFile
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "non-existent file",
			setupState: func() string {
				return filepath.Join(suite.tempDir, "nonexistent.json")
			},
			expectValid: false,
			expectError: false, // Not an error, just returns false
		},
		{
			name: "expired state file",
			setupState: func() string {
				// Create an expired state file
				expiredState := StateFile{
					ValidatedAt: time.Now().Add(-10 * time.Minute),
					ValidUntil:  time.Now().Add(-5 * time.Minute), // Expired 5 minutes ago
				}
				expiredState.Signature = generateStateSignature(expiredState)

				data, err := json.MarshalIndent(expiredState, "", "  ")
				suite.NoError(err)

				stateFile := filepath.Join(suite.tempDir, "expired_state.json")
				err = os.WriteFile(stateFile, data, 0600)
				suite.NoError(err)

				return stateFile
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "future validated time",
			setupState: func() string {
				// Create a state file with future validation time
				futureState := StateFile{
					ValidatedAt: time.Now().Add(10 * time.Minute), // Future time
					ValidUntil:  time.Now().Add(15 * time.Minute),
				}
				futureState.Signature = generateStateSignature(futureState)

				data, err := json.MarshalIndent(futureState, "", "  ")
				suite.NoError(err)

				stateFile := filepath.Join(suite.tempDir, "future_state.json")
				err = os.WriteFile(stateFile, data, 0600)
				suite.NoError(err)

				return stateFile
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid signature",
			setupState: func() string {
				// Create a state file with invalid signature
				invalidState := StateFile{
					ValidatedAt: time.Now(),
					ValidUntil:  time.Now().Add(5 * time.Minute),
					Signature:   "invalid-signature-12345",
				}

				data, err := json.MarshalIndent(invalidState, "", "  ")
				suite.NoError(err)

				stateFile := filepath.Join(suite.tempDir, "invalid_sig_state.json")
				err = os.WriteFile(stateFile, data, 0600)
				suite.NoError(err)

				return stateFile
			},
			expectValid: false,
			expectError: true, // Invalid signature is an error
		},
		{
			name: "corrupted JSON",
			setupState: func() string {
				stateFile := filepath.Join(suite.tempDir, "corrupted_state.json")
				err := os.WriteFile(stateFile, []byte("invalid json {"), 0600)
				suite.NoError(err)
				return stateFile
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "empty file",
			setupState: func() string {
				stateFile := filepath.Join(suite.tempDir, "empty_state.json")
				err := os.WriteFile(stateFile, []byte(""), 0600)
				suite.NoError(err)
				return stateFile
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "missing signature field",
			setupState: func() string {
				// Create state file without signature
				partialState := map[string]interface{}{
					"validated_at": time.Now().Format(time.RFC3339),
					"valid_until":  time.Now().Add(5 * time.Minute).Format(time.RFC3339),
					// No signature field
				}

				data, err := json.MarshalIndent(partialState, "", "  ")
				suite.NoError(err)

				stateFile := filepath.Join(suite.tempDir, "no_sig_state.json")
				err = os.WriteFile(stateFile, data, 0600)
				suite.NoError(err)

				return stateFile
			},
			expectValid: false,
			expectError: true, // Missing signature is an error
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			stateFile := tt.setupState()
			defer os.Remove(stateFile) // Clean up

			valid, err := suite.manager.ValidateStateFile(stateFile)

			suite.Equal(tt.expectValid, valid)
			if tt.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestGenerateStateSignature tests signature generation
func (suite *StateFileTestSuite) TestGenerateStateSignature() {
	now := time.Now()
	later := now.Add(5 * time.Minute)

	state1 := StateFile{
		ValidatedAt: now,
		ValidUntil:  later,
	}

	state2 := StateFile{
		ValidatedAt: now,
		ValidUntil:  later,
	}

	state3 := StateFile{
		ValidatedAt: now.Add(time.Second), // Different time
		ValidUntil:  later,
	}

	// Same state should generate same signature
	sig1 := generateStateSignature(state1)
	sig2 := generateStateSignature(state2)
	sig3 := generateStateSignature(state3)

	suite.Equal(sig1, sig2, "Same state should generate same signature")
	suite.NotEqual(sig1, sig3, "Different state should generate different signature")

	// Signature should be a hex string
	suite.Regexp("^[a-f0-9]+$", sig1)
	suite.Equal(64, len(sig1), "HMAC-SHA256 should produce 64 character hex string")
}

// TestStateFileEdgeCases tests edge cases and error conditions
func (suite *StateFileTestSuite) TestStateFileEdgeCases() {
	suite.Run("state file with zero times", func() {
		zeroState := StateFile{
			ValidatedAt: time.Time{}, // Zero time
			ValidUntil:  time.Time{}, // Zero time
		}
		zeroState.Signature = generateStateSignature(zeroState)

		data, err := json.MarshalIndent(zeroState, "", "  ")
		suite.NoError(err)

		stateFile := filepath.Join(suite.tempDir, "zero_time_state.json")
		err = os.WriteFile(stateFile, data, 0600)
		suite.NoError(err)
		defer os.Remove(stateFile)

		valid, err := suite.manager.ValidateStateFile(stateFile)
		suite.NoError(err)
		suite.False(valid) // Zero times should be invalid
	})

	suite.Run("state file with same validated and valid until times", func() {
		now := time.Now()
		sameTimeState := StateFile{
			ValidatedAt: now,
			ValidUntil:  now, // Same time
		}
		sameTimeState.Signature = generateStateSignature(sameTimeState)

		data, err := json.MarshalIndent(sameTimeState, "", "  ")
		suite.NoError(err)

		stateFile := filepath.Join(suite.tempDir, "same_time_state.json")
		err = os.WriteFile(stateFile, data, 0600)
		suite.NoError(err)
		defer os.Remove(stateFile)

		valid, err := suite.manager.ValidateStateFile(stateFile)
		suite.NoError(err)
		suite.False(valid) // Should be invalid since ValidUntil is not after ValidatedAt
	})

	suite.Run("state file with inverted times", func() {
		now := time.Now()
		invertedState := StateFile{
			ValidatedAt: now,
			ValidUntil:  now.Add(-5 * time.Minute), // Before validated time
		}
		invertedState.Signature = generateStateSignature(invertedState)

		data, err := json.MarshalIndent(invertedState, "", "  ")
		suite.NoError(err)

		stateFile := filepath.Join(suite.tempDir, "inverted_time_state.json")
		err = os.WriteFile(stateFile, data, 0600)
		suite.NoError(err)
		defer os.Remove(stateFile)

		valid, err := suite.manager.ValidateStateFile(stateFile)
		suite.NoError(err)
		suite.False(valid) // Should be invalid
	})
}

// TestStateFileConsistency tests consistency of state file operations
func (suite *StateFileTestSuite) TestStateFileConsistency() {
	// Create multiple state files and verify they all work consistently
	for i := 0; i < 10; i++ {
		stateFile := filepath.Join(suite.tempDir, fmt.Sprintf("consistency_test_%d.json", i))

		// Create state file
		err := suite.manager.CreateStateFile(stateFile)
		suite.NoError(err)

		// Validate immediately - should be valid
		valid, err := suite.manager.ValidateStateFile(stateFile)
		suite.NoError(err)
		suite.True(valid)

		// Read and verify structure
		data, err := os.ReadFile(stateFile)
		suite.NoError(err)

		var state StateFile
		err = json.Unmarshal(data, &state)
		suite.NoError(err)

		// Verify signature is correct
		expectedSig := generateStateSignature(state)
		suite.Equal(expectedSig, state.Signature)

		// Clean up
		os.Remove(stateFile)
	}
}

// TestStateFilePermissions tests file permission handling
func (suite *StateFileTestSuite) TestStateFilePermissions() {
	if os.Getenv("CI") == "true" {
		suite.T().Skip("Skipping permission tests in CI environment")
	}
	if runtime.GOOS == "windows" {
		suite.T().Skip("Skipping permission tests on Windows where chmod is not enforced")
	}

	err := suite.manager.CreateStateFile(suite.stateFile)
	suite.NoError(err)
	defer os.Remove(suite.stateFile)

	// Check file permissions
	info, err := os.Stat(suite.stateFile)
	suite.NoError(err)

	// Should have restrictive permissions (0600)
	mode := info.Mode()
	suite.Equal(os.FileMode(0600), mode&0777, "State file should have 0600 permissions")
}

// TestGetMachineID tests deprecated machine ID function
func (suite *StateFileTestSuite) TestGetMachineID() {
	// GetMachineID is deprecated and should return empty string
	machineID := suite.manager.GetMachineID()
	suite.Equal("", machineID, "GetMachineID should return empty string as it's deprecated")
}

// TestCleanupStateFile tests state file cleanup utility
func (suite *StateFileTestSuite) TestCleanupStateFile() {
	suite.Run("cleanup existing file", func() {
		// Create a state file
		err := suite.manager.CreateStateFile(suite.stateFile)
		suite.NoError(err)

		// Verify file exists
		_, err = os.Stat(suite.stateFile)
		suite.NoError(err)

		// Cleanup
		err = CleanupStateFile(suite.stateFile)
		suite.NoError(err)

		// Verify file is gone
		_, err = os.Stat(suite.stateFile)
		suite.True(os.IsNotExist(err))
	})

	suite.Run("cleanup non-existent file", func() {
		nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.json")

		// Should not error when cleaning up non-existent file
		err := CleanupStateFile(nonExistentFile)
		suite.NoError(err)
	})

	suite.Run("cleanup read-only file", func() {
		if os.Getenv("CI") == "true" {
			suite.T().Skip("Skipping read-only test in CI environment")
		}

		// Create a state file
		err := suite.manager.CreateStateFile(suite.stateFile)
		suite.NoError(err)

		// Make file read-only
		err = os.Chmod(suite.stateFile, 0444)
		suite.NoError(err)

		// Cleanup should still work (or at least not panic)
		err = CleanupStateFile(suite.stateFile)
		// May or may not succeed depending on system, but shouldn't panic
		suite.T().Logf("Cleanup read-only file result: %v", err)

		// Restore permissions for final cleanup
		os.Chmod(suite.stateFile, 0644)
		os.Remove(suite.stateFile)
	})
}

// TestStateFileRaceConditions tests concurrent state file operations
func (suite *StateFileTestSuite) TestStateFileRaceConditions() {
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent state file creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			stateFile := filepath.Join(suite.tempDir, fmt.Sprintf("race_test_%d.json", id))
			err := suite.manager.CreateStateFile(stateFile)
			suite.NoError(err)

			// Validate immediately
			valid, err := suite.manager.ValidateStateFile(stateFile)
			suite.NoError(err)
			suite.True(valid)

			// Cleanup
			os.Remove(stateFile)
		}(i)
	}

	wg.Wait()
}

// TestStateFileTimingAccuracy tests timing accuracy of state files
func (suite *StateFileTestSuite) TestStateFileTimingAccuracy() {
	start := time.Now()

	err := suite.manager.CreateStateFile(suite.stateFile)
	suite.NoError(err)
	defer os.Remove(suite.stateFile)

	end := time.Now()

	// Read the state file
	data, err := os.ReadFile(suite.stateFile)
	suite.NoError(err)

	var state StateFile
	err = json.Unmarshal(data, &state)
	suite.NoError(err)

	// ValidatedAt should be between start and end
	suite.True(state.ValidatedAt.After(start) || state.ValidatedAt.Equal(start))
	suite.True(state.ValidatedAt.Before(end) || state.ValidatedAt.Equal(end))

	// ValidUntil should be approximately 5 minutes after ValidatedAt
	expectedValidUntil := state.ValidatedAt.Add(5 * time.Minute)
	suite.WithinDuration(expectedValidUntil, state.ValidUntil, time.Second)
}

// Run the state file test suite
func TestStateFileTestSuite(t *testing.T) {
	suite.Run(t, new(StateFileTestSuite))
}

// =============================================================================
// Unit Tests for Signature Generation
// =============================================================================

func TestGenerateStateSignature(t *testing.T) {
	tests := []struct {
		name     string
		state    StateFile
		expected bool // Whether we expect a valid signature format
	}{
		{
			name: "normal state",
			state: StateFile{
				ValidatedAt: time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC),
				ValidUntil:  time.Date(2024, 8, 1, 12, 5, 0, 0, time.UTC),
				Signature:   "", // Will be ignored in signature generation
			},
			expected: true,
		},
		{
			name: "zero times",
			state: StateFile{
				ValidatedAt: time.Time{},
				ValidUntil:  time.Time{},
			},
			expected: true, // Should still generate valid signature format
		},
		{
			name: "future times",
			state: StateFile{
				ValidatedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
				ValidUntil:  time.Date(2030, 1, 1, 0, 5, 0, 0, time.UTC),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := generateStateSignature(tt.state)

			if tt.expected {
				assert.NotEmpty(t, signature)
				assert.Regexp(t, "^[a-f0-9]+$", signature, "Signature should be hex string")
				assert.Equal(t, 64, len(signature), "HMAC-SHA256 should produce 64 character hex")
			}

			// Test consistency - same input should produce same signature
			signature2 := generateStateSignature(tt.state)
			assert.Equal(t, signature, signature2, "Same state should produce same signature")
		})
	}
}

func TestStateSignatureConsistency(t *testing.T) {
	baseTime := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)

	state1 := StateFile{
		ValidatedAt: baseTime,
		ValidUntil:  baseTime.Add(5 * time.Minute),
	}

	state2 := StateFile{
		ValidatedAt: baseTime,
		ValidUntil:  baseTime.Add(5 * time.Minute),
		Signature:   "this-should-be-ignored",
	}

	state3 := StateFile{
		ValidatedAt: baseTime.Add(time.Nanosecond), // Tiny difference
		ValidUntil:  baseTime.Add(5 * time.Minute),
	}

	sig1 := generateStateSignature(state1)
	sig2 := generateStateSignature(state2)
	sig3 := generateStateSignature(state3)

	// Same times should produce same signature, regardless of existing signature field
	assert.Equal(t, sig1, sig2, "Signature field should be ignored in generation")

	// Different times should produce different signatures
	assert.NotEqual(t, sig1, sig3, "Different times should produce different signatures")
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkCreateStateFile(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "bench_license.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stateFile := filepath.Join(tempDir, fmt.Sprintf("bench_state_%d.json", i))
		err := manager.CreateStateFile(stateFile)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(stateFile) // Clean up
	}
}

func BenchmarkValidateStateFile(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "bench_license.dat")
	stateFile := filepath.Join(tempDir, "bench_state.json")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	// Create a valid state file
	err = manager.CreateStateFile(stateFile)
	require.NoError(b, err)
	defer os.Remove(stateFile)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		valid, err := manager.ValidateStateFile(stateFile)
		if err != nil {
			b.Fatal(err)
		}
		if !valid {
			b.Fatal("State file should be valid")
		}
	}
}

func BenchmarkGenerateStateSignature(b *testing.B) {
	state := StateFile{
		ValidatedAt: time.Now(),
		ValidUntil:  time.Now().Add(5 * time.Minute),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generateStateSignature(state)
	}
}
