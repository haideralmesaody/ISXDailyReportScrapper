package security

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFingerprintManagerCreation tests creation and initialization
func TestFingerprintManagerCreation(t *testing.T) {
	manager := NewFingerprintManager()
	
	assert.NotNil(t, manager)
	assert.Equal(t, time.Hour, manager.cacheDuration)
	assert.Nil(t, manager.cache)
	assert.True(t, manager.cacheExpiry.IsZero())
}

// TestMACAddressRetrieval tests MAC address retrieval with various scenarios
func TestMACAddressRetrieval(t *testing.T) {
	manager := NewFingerprintManager()

	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "retrieve MAC address",
			expectError: false, // Should succeed on most systems
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macAddr, err := manager.GetMACAddress()
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				// MAC address retrieval might fail on some test environments
				if err != nil {
					t.Logf("MAC address retrieval failed (expected in some test environments): %v", err)
					return
				}
				
				require.NoError(t, err)
				assert.NotEmpty(t, macAddr)
				
				// Validate MAC address format (XX:XX:XX:XX:XX:XX)
				parts := strings.Split(macAddr, ":")
				if len(parts) == 6 {
					for _, part := range parts {
						assert.Len(t, part, 2, "Each MAC address part should be 2 characters")
					}
				}
				
				// Should not be all zeros
				assert.NotEqual(t, "00:00:00:00:00:00", macAddr)
			}
		})
	}
}

// TestHostnameRetrieval tests hostname retrieval and normalization
func TestHostnameRetrieval(t *testing.T) {
	manager := NewFingerprintManager()

	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "retrieve hostname",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname, err := manager.GetHostname()
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hostname)
				
				// Hostname should be lowercase and trimmed
				assert.Equal(t, strings.ToLower(strings.TrimSpace(hostname)), hostname)
				
				// Should not contain spaces at start/end
				assert.Equal(t, strings.TrimSpace(hostname), hostname)
			}
		})
	}
}

// TestCPUIDRetrieval tests CPU ID retrieval for different operating systems
func TestCPUIDRetrieval(t *testing.T) {
	manager := NewFingerprintManager()

	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "retrieve CPU ID",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpuID, err := manager.GetCPUID()
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, cpuID)
				
				// CPU ID should be a reasonable length
				assert.GreaterOrEqual(t, len(cpuID), 4)
				assert.LessOrEqual(t, len(cpuID), 64)
				
				// Should contain alphanumeric characters or hyphens
				for _, char := range cpuID {
					assert.True(t, 
						(char >= 'a' && char <= 'z') ||
						(char >= 'A' && char <= 'Z') ||
						(char >= '0' && char <= '9') ||
						char == '-' || char == '_',
						"CPU ID contains invalid character: %c", char)
				}
			}
		})
	}
}

// TestCPUIDPlatformSpecific tests platform-specific CPU ID generation
func TestCPUIDPlatformSpecific(t *testing.T) {
	manager := NewFingerprintManager()

	switch runtime.GOOS {
	case "windows":
		t.Run("windows CPU ID", func(t *testing.T) {
			cpuID, err := manager.getCPUIDWindows()
			require.NoError(t, err)
			assert.NotEmpty(t, cpuID)
			
			// Windows CPU ID should be hex encoded (from hash)
			assert.Len(t, cpuID, 16) // 8 bytes hex encoded = 16 characters
		})
		
	case "linux":
		t.Run("linux CPU ID", func(t *testing.T) {
			cpuID, err := manager.getCPUIDLinux()
			require.NoError(t, err)
			assert.NotEmpty(t, cpuID)
			
			// Linux CPU ID should be hex encoded (from hash)
			assert.Len(t, cpuID, 16) // 8 bytes hex encoded = 16 characters
		})
		
	case "darwin":
		t.Run("darwin CPU ID", func(t *testing.T) {
			cpuID, err := manager.getCPUIDDarwin()
			require.NoError(t, err)
			assert.NotEmpty(t, cpuID)
			
			// Darwin CPU ID should be hex encoded (from hash)
			assert.Len(t, cpuID, 16) // 8 bytes hex encoded = 16 characters
		})
	}
}

// TestFingerprintGeneration tests complete fingerprint generation
func TestFingerprintGeneration(t *testing.T) {
	manager := NewFingerprintManager()

	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "generate complete fingerprint",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fingerprint, err := manager.GenerateFingerprint()
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fingerprint)
				
				// Validate fingerprint structure
				assert.NotEmpty(t, fingerprint.Fingerprint)
				assert.Len(t, fingerprint.Fingerprint, 64) // SHA256 hex = 64 chars
				assert.NotEmpty(t, fingerprint.OS)
				assert.NotEmpty(t, fingerprint.Platform)
				assert.False(t, fingerprint.GeneratedAt.IsZero())
				
				// OS and Platform should match runtime
				assert.Equal(t, runtime.GOOS, fingerprint.OS)
				assert.Equal(t, runtime.GOARCH, fingerprint.Platform)
				
				// Fingerprint should be hex encoded SHA256
				for _, char := range fingerprint.Fingerprint {
					assert.True(t, 
						(char >= 'a' && char <= 'f') ||
						(char >= '0' && char <= '9'),
						"Fingerprint contains invalid hex character: %c", char)
				}
			}
		})
	}
}

// TestFingerprintConsistency tests that fingerprints are consistent across calls
func TestFingerprintConsistency(t *testing.T) {
	manager := NewFingerprintManager()

	// Generate first fingerprint
	fingerprint1, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	// Wait a small amount to ensure time-based differences would show up
	time.Sleep(time.Millisecond * 10)

	// Generate second fingerprint
	fingerprint2, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	// Fingerprints should be identical (except GeneratedAt due to caching)
	assert.Equal(t, fingerprint1.Fingerprint, fingerprint2.Fingerprint)
	assert.Equal(t, fingerprint1.Hostname, fingerprint2.Hostname)
	assert.Equal(t, fingerprint1.MACAddress, fingerprint2.MACAddress)
	assert.Equal(t, fingerprint1.CPUID, fingerprint2.CPUID)
	assert.Equal(t, fingerprint1.OS, fingerprint2.OS)
	assert.Equal(t, fingerprint1.Platform, fingerprint2.Platform)
}

// TestFingerprintCaching tests caching behavior
func TestFingerprintCaching(t *testing.T) {
	manager := NewFingerprintManager()

	// First call should generate and cache
	fingerprint1, err := manager.GenerateFingerprint()
	require.NoError(t, err)
	
	startTime := fingerprint1.GeneratedAt

	// Second call within cache duration should return cached result
	time.Sleep(time.Millisecond * 50)
	fingerprint2, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	// Should be the exact same object from cache
	assert.Equal(t, startTime, fingerprint2.GeneratedAt)
	assert.Equal(t, fingerprint1.Fingerprint, fingerprint2.Fingerprint)

	// Clear cache and generate again
	manager.ClearCache()
	
	time.Sleep(time.Millisecond * 10)
	fingerprint3, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	// Should be a new generation
	assert.True(t, fingerprint3.GeneratedAt.After(startTime))
	assert.Equal(t, fingerprint1.Fingerprint, fingerprint3.Fingerprint) // Same device, same fingerprint
}

// TestFingerprintValidation tests fingerprint validation functionality
func TestFingerprintValidation(t *testing.T) {
	manager := NewFingerprintManager()

	// Generate a fingerprint
	fingerprint, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	tests := []struct {
		name               string
		storedFingerprint  string
		expectValid        bool
		expectError        bool
	}{
		{
			name:              "valid matching fingerprint",
			storedFingerprint: fingerprint.Fingerprint,
			expectValid:       true,
			expectError:       false,
		},
		{
			name:              "invalid non-matching fingerprint",
			storedFingerprint: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expectValid:       false,
			expectError:       false,
		},
		{
			name:              "malformed fingerprint",
			storedFingerprint: "invalid-fingerprint",
			expectValid:       false,
			expectError:       false,
		},
		{
			name:              "empty fingerprint",
			storedFingerprint: "",
			expectValid:       false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, err := manager.ValidateFingerprint(tt.storedFingerprint)
			
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectValid, isValid)
			}
		})
	}
}

// TestFingerprintComponents tests individual component retrieval
func TestFingerprintComponents(t *testing.T) {
	manager := NewFingerprintManager()

	components, err := manager.GetFingerprintComponents()
	require.NoError(t, err)
	require.NotNil(t, components)

	// Check expected component keys exist
	expectedKeys := []string{"mac_address", "hostname", "cpu_id", "os", "platform"}
	for _, key := range expectedKeys {
		assert.Contains(t, components, key, "Missing expected component: %s", key)
	}

	// Verify OS and platform are correct
	assert.Equal(t, runtime.GOOS, components["os"])
	assert.Equal(t, runtime.GOARCH, components["platform"])

	// Verify components are not empty (allowing for fallbacks)
	for key, value := range components {
		if value == "" && (key == "mac_address" || key == "hostname" || key == "cpu_id") {
			t.Logf("Warning: Component %s is empty (may be expected in test environment)", key)
		}
	}
}

// TestFingerprintFallbackScenarios tests fallback behavior when hardware info unavailable
func TestFingerprintFallbackScenarios(t *testing.T) {
	manager := NewFingerprintManager()

	// Test MAC address fallback (can't easily simulate failure, just ensure it doesn't panic)
	t.Run("MAC address fallback", func(t *testing.T) {
		macAddr, err := manager.GetMACAddress()
		// Error is acceptable in test environments
		if err != nil {
			t.Logf("MAC address retrieval failed (expected in some test environments): %v", err)
		} else {
			assert.NotEmpty(t, macAddr)
		}
	})

	// Test hostname fallback
	t.Run("hostname retrieval", func(t *testing.T) {
		hostname, err := manager.GetHostname()
		require.NoError(t, err)
		assert.NotEmpty(t, hostname)
	})

	// Test CPU ID fallback
	t.Run("CPU ID fallback", func(t *testing.T) {
		cpuID, err := manager.GetCPUID()
		require.NoError(t, err)
		assert.NotEmpty(t, cpuID)
	})
}

// TestCacheExpiry tests cache expiration behavior
func TestCacheExpiry(t *testing.T) {
	// Create manager with very short cache duration for testing
	manager := &FingerprintManager{
		cacheDuration: time.Millisecond * 100,
	}

	// Generate initial fingerprint
	fingerprint1, err := manager.GenerateFingerprint()
	require.NoError(t, err)
	
	startTime := fingerprint1.GeneratedAt

	// Wait for cache to expire
	time.Sleep(time.Millisecond * 150)

	// Generate again - should create new fingerprint
	fingerprint2, err := manager.GenerateFingerprint()
	require.NoError(t, err)

	// Should be a new generation (different timestamp)
	assert.True(t, fingerprint2.GeneratedAt.After(startTime))
	
	// But fingerprint value should be the same (same device)
	assert.Equal(t, fingerprint1.Fingerprint, fingerprint2.Fingerprint)
}

// TestConcurrentFingerprintGeneration tests thread safety
func TestConcurrentFingerprintGeneration(t *testing.T) {
	manager := NewFingerprintManager()
	const goroutineCount = 10

	// Generate fingerprints concurrently
	fingerprints := make([]*DeviceFingerprint, goroutineCount)
	errors := make([]error, goroutineCount)

	var startSignal = make(chan struct{})
	var doneSignal = make(chan struct{}, goroutineCount)

	// Start all goroutines
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			<-startSignal // Wait for start signal
			fingerprints[index], errors[index] = manager.GenerateFingerprint()
			doneSignal <- struct{}{}
		}(i)
	}

	// Signal all goroutines to start
	close(startSignal)

	// Wait for all to complete
	for i := 0; i < goroutineCount; i++ {
		<-doneSignal
	}

	// Verify all succeeded and generated identical fingerprints
	var baseFingerprint string
	for i, fingerprint := range fingerprints {
		require.NoError(t, errors[i], "Goroutine %d failed", i)
		require.NotNil(t, fingerprint, "Goroutine %d returned nil fingerprint", i)
		
		if i == 0 {
			baseFingerprint = fingerprint.Fingerprint
		} else {
			assert.Equal(t, baseFingerprint, fingerprint.Fingerprint, 
				"Fingerprint %d differs from base", i)
		}
	}
}

// Benchmark tests for performance validation
func BenchmarkFingerprintGeneration(b *testing.B) {
	manager := NewFingerprintManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateFingerprint()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkFingerprintValidation(b *testing.B) {
	manager := NewFingerprintManager()
	
	// Generate a fingerprint to validate against
	fingerprint, err := manager.GenerateFingerprint()
	if err != nil {
		b.Fatalf("Failed to generate test fingerprint: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ValidateFingerprint(fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkMACAddressRetrieval(b *testing.B) {
	manager := NewFingerprintManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetMACAddress()
		// Allow errors in test environments
		if err != nil {
			b.Logf("MAC address retrieval failed: %v", err)
		}
	}
}

func BenchmarkCPUIDGeneration(b *testing.B) {
	manager := NewFingerprintManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetCPUID()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}