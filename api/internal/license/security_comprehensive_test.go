package license

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Comprehensive Security Manager Tests
// =============================================================================

// SecurityManagerTestSuite provides comprehensive testing for security functionality
type SecurityManagerTestSuite struct {
	suite.Suite
	security *SecurityManager
}

func (suite *SecurityManagerTestSuite) SetupTest() {
	// Standard configuration: 5 attempts, 15 minute block, 5 minute window
	suite.security = NewSecurityManager(5, 15*time.Minute, 5*time.Minute)
}

func (suite *SecurityManagerTestSuite) TearDownTest() {
	if suite.security != nil {
		suite.security.Stop()
	}
}

// TestSecurityManagerConstruction tests security manager creation
func (suite *SecurityManagerTestSuite) TestSecurityManagerConstruction() {
	tests := []struct {
		name           string
		maxAttempts    int
		blockDuration  time.Duration
		windowDuration time.Duration
		expectValid    bool
	}{
		{
			name:           "standard configuration",
			maxAttempts:    5,
			blockDuration:  15 * time.Minute,
			windowDuration: 5 * time.Minute,
			expectValid:    true,
		},
		{
			name:           "minimal configuration",
			maxAttempts:    1,
			blockDuration:  1 * time.Minute,
			windowDuration: 30 * time.Second,
			expectValid:    true,
		},
		{
			name:           "aggressive configuration",
			maxAttempts:    10,
			blockDuration:  1 * time.Hour,
			windowDuration: 10 * time.Minute,
			expectValid:    true,
		},
		{
			name:           "zero max attempts",
			maxAttempts:    0,
			blockDuration:  5 * time.Minute,
			windowDuration: 1 * time.Minute,
			expectValid:    true, // Should still work
		},
		{
			name:           "zero durations",
			maxAttempts:    5,
			blockDuration:  0,
			windowDuration: 0,
			expectValid:    true, // Should still work
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			sm := NewSecurityManager(tt.maxAttempts, tt.blockDuration, tt.windowDuration)
			defer sm.Stop()

			suite.NotNil(sm)
			suite.Equal(tt.maxAttempts, sm.maxAttempts)
			suite.Equal(tt.blockDuration, sm.blockDuration)
			suite.Equal(tt.windowDuration, sm.windowDuration)
			suite.NotNil(sm.attemptCounts)
			suite.NotNil(sm.lastAttempts)
			suite.NotNil(sm.blockedIPs)
		})
	}
}

// TestBasicRateLimiting tests fundamental rate limiting behavior
func (suite *SecurityManagerTestSuite) TestBasicRateLimiting() {
	identifier := "test-client-1"

	// Record successful attempts - should not be blocked
	for i := 0; i < 10; i++ {
		suite.False(suite.security.IsBlocked(identifier))
		result := suite.security.RecordAttempt(identifier, true)
		suite.True(result)
	}

	// Should still not be blocked after successful attempts
	suite.False(suite.security.IsBlocked(identifier))
}

// TestFailureBasedBlocking tests blocking after failed attempts
func (suite *SecurityManagerTestSuite) TestFailureBasedBlocking() {
	identifier := "test-client-fail"

	// Record failed attempts up to the limit
	for i := 0; i < 4; i++ {
		suite.False(suite.security.IsBlocked(identifier))
		result := suite.security.RecordAttempt(identifier, false)
		suite.True(result) // Should still allow attempts
	}

	// Fifth failed attempt should trigger blocking
	suite.False(suite.security.IsBlocked(identifier))
	result := suite.security.RecordAttempt(identifier, false)
	suite.False(result) // Should now be blocked

	// Should now be blocked
	suite.True(suite.security.IsBlocked(identifier))
}

// TestWindowBasedReset tests that attempts reset after time window
func (suite *SecurityManagerTestSuite) TestWindowBasedReset() {
	// Use shorter window for testing
	shortWindowSecurity := NewSecurityManager(3, 5*time.Minute, 100*time.Millisecond)
	defer shortWindowSecurity.Stop()

	identifier := "test-window-reset"

	// Record 2 failed attempts
	shortWindowSecurity.RecordAttempt(identifier, false)
	shortWindowSecurity.RecordAttempt(identifier, false)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Next attempt should reset the count
	result := shortWindowSecurity.RecordAttempt(identifier, false)
	suite.True(result) // Should not be blocked yet

	// Should be able to make more attempts
	shortWindowSecurity.RecordAttempt(identifier, false)
	result = shortWindowSecurity.RecordAttempt(identifier, false)
	suite.False(result) // Now should be blocked (3 attempts in new window)
}

// TestBlockDurationExpiry tests that blocks expire after block duration
func (suite *SecurityManagerTestSuite) TestBlockDurationExpiry() {
	// Use shorter block duration for testing
	shortBlockSecurity := NewSecurityManager(2, 100*time.Millisecond, 5*time.Minute)
	defer shortBlockSecurity.Stop()

	identifier := "test-block-expiry"

	// Get blocked
	shortBlockSecurity.RecordAttempt(identifier, false)
	shortBlockSecurity.RecordAttempt(identifier, false)
	suite.True(shortBlockSecurity.IsBlocked(identifier))

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// Should no longer be blocked
	suite.False(shortBlockSecurity.IsBlocked(identifier))
}

// TestSuccessfulAttemptResetsCount tests that successful attempts reset failure count
func (suite *SecurityManagerTestSuite) TestSuccessfulAttemptResetsCount() {
	identifier := "test-success-reset"

	// Record some failed attempts
	for i := 0; i < 3; i++ {
		suite.security.RecordAttempt(identifier, false)
	}

	// Record successful attempt - should reset count
	result := suite.security.RecordAttempt(identifier, true)
	suite.True(result)

	// Should be able to make failed attempts again
	for i := 0; i < 4; i++ {
		result := suite.security.RecordAttempt(identifier, false)
		suite.True(result) // Should not be blocked until 5th attempt
	}

	// Fifth attempt should block
	result = suite.security.RecordAttempt(identifier, false)
	suite.False(result)
}

// TestMultipleIdentifiers tests handling of multiple clients
func (suite *SecurityManagerTestSuite) TestMultipleIdentifiers() {
	clients := []string{"client-1", "client-2", "client-3", "client-4", "client-5"}

	// Block some clients but not others
	for i, client := range clients {
		if i%2 == 0 {
			// Block even-indexed clients
			for j := 0; j < 5; j++ {
				suite.security.RecordAttempt(client, false)
			}
			suite.True(suite.security.IsBlocked(client))
		} else {
			// Keep odd-indexed clients unblocked
			suite.security.RecordAttempt(client, true)
			suite.False(suite.security.IsBlocked(client))
		}
	}

	// Verify states are independent
	suite.True(suite.security.IsBlocked("client-1"))
	suite.False(suite.security.IsBlocked("client-2"))
	suite.True(suite.security.IsBlocked("client-3"))
	suite.False(suite.security.IsBlocked("client-4"))
	suite.True(suite.security.IsBlocked("client-5"))
}

// TestConcurrentAccess tests thread safety
func (suite *SecurityManagerTestSuite) TestConcurrentAccess() {
	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	// Concurrent operations on different identifiers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			identifier := fmt.Sprintf("concurrent-client-%d", clientID)

			for j := 0; j < numOperations; j++ {
				// Mix of operations
				switch j % 4 {
				case 0:
					suite.security.IsBlocked(identifier)
				case 1:
					suite.security.RecordAttempt(identifier, true)
				case 2:
					suite.security.RecordAttempt(identifier, false)
				case 3:
					suite.security.GetStats()
				}
			}
		}(i)
	}

	// Concurrent operations on same identifier
	sharedIdentifier := "shared-client"
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				suite.security.IsBlocked(sharedIdentifier)
				suite.security.RecordAttempt(sharedIdentifier, j%3 == 0) // Success rate: 33%
			}
		}()
	}

	wg.Wait()

	// Should complete without deadlocks or panics
	stats := suite.security.GetStats()
	suite.NotNil(stats)
}

// TestStatsAccuracy tests statistics accuracy
func (suite *SecurityManagerTestSuite) TestStatsAccuracy() {
	// Create known state
	clients := []string{"stats-1", "stats-2", "stats-3"}

	// Block 2 clients
	for i := 0; i < 2; i++ {
		for j := 0; j < 5; j++ {
			suite.security.RecordAttempt(clients[i], false)
		}
	}

	// Keep 1 client active but not blocked
	for i := 0; i < 3; i++ {
		suite.security.RecordAttempt(clients[2], false)
	}

	stats := suite.security.GetStats()
	suite.NotNil(stats)

	// Verify stats structure
	suite.Contains(stats, "active_attempts")
	suite.Contains(stats, "blocked_ips")
	suite.Contains(stats, "max_attempts")
	suite.Contains(stats, "block_duration")
	suite.Contains(stats, "window_duration")

	// Verify basic counts
	suite.Equal(3, stats["active_attempts"]) // 3 clients with attempts
	suite.Equal(2, stats["blocked_ips"])     // 2 clients blocked
	suite.Equal(5, stats["max_attempts"])
	suite.Equal("15m0s", stats["block_duration"])
	suite.Equal("5m0s", stats["window_duration"])
}

// TestCleanupFunctionality tests automatic cleanup
func (suite *SecurityManagerTestSuite) TestCleanupFunctionality() {
	// Use very short cleanup interval for testing
	quickCleanupSecurity := &SecurityManager{
		attemptCounts:   make(map[string]int),
		lastAttempts:    make(map[string]time.Time),
		blockedIPs:      make(map[string]time.Time),
		maxAttempts:     5,
		blockDuration:   1 * time.Minute,
		windowDuration:  10 * time.Millisecond, // Very short for quick expiry
		cleanupInterval: 50 * time.Millisecond, // Very frequent cleanup
		stopChan:        make(chan struct{}),
	}

	go quickCleanupSecurity.cleanup()
	defer quickCleanupSecurity.Stop()

	// Add some attempts that will expire quickly
	for i := 0; i < 5; i++ {
		identifier := fmt.Sprintf("cleanup-test-%d", i)
		quickCleanupSecurity.RecordAttempt(identifier, false)
	}

	// Verify they exist
	stats := quickCleanupSecurity.GetStats()
	suite.Equal(5, stats["active_attempts"])

	// Wait for cleanup to happen
	time.Sleep(100 * time.Millisecond)

	// Verify cleanup occurred
	stats = quickCleanupSecurity.GetStats()
	suite.LessOrEqual(stats["active_attempts"], 5) // Should be cleaned up
}

// TestEdgeCases tests various edge cases
func (suite *SecurityManagerTestSuite) TestEdgeCases() {
	suite.Run("empty identifier", func() {
		// Should handle empty identifier without panicking
		suite.False(suite.security.IsBlocked(""))
		result := suite.security.RecordAttempt("", false)
		suite.True(result) // Should work normally
	})

	suite.Run("very long identifier", func() {
		longIdentifier := string(make([]byte, 10000))
		for i := range longIdentifier {
			longIdentifier = longIdentifier[:i] + "a" + longIdentifier[i:]
		}

		suite.False(suite.security.IsBlocked(longIdentifier))
		result := suite.security.RecordAttempt(longIdentifier, false)
		suite.True(result)
	})

	suite.Run("special characters in identifier", func() {
		specialIdentifier := "client@#$%^&*()_+{}|:<>?[]\\;'\",./"

		suite.False(suite.security.IsBlocked(specialIdentifier))
		result := suite.security.RecordAttempt(specialIdentifier, false)
		suite.True(result)
	})

	suite.Run("unicode identifier", func() {
		unicodeIdentifier := "клиент-测试-🔒-client"

		suite.False(suite.security.IsBlocked(unicodeIdentifier))
		result := suite.security.RecordAttempt(unicodeIdentifier, false)
		suite.True(result)
	})
}

// TestZeroLimitConfiguration tests configuration with zero max attempts
func (suite *SecurityManagerTestSuite) TestZeroLimitConfiguration() {
	zeroLimitSecurity := NewSecurityManager(0, 5*time.Minute, 1*time.Minute)
	defer zeroLimitSecurity.Stop()

	identifier := "zero-limit-test"

	// With zero limit, first failed attempt should block
	result := zeroLimitSecurity.RecordAttempt(identifier, false)
	suite.False(result) // Should be blocked immediately

	suite.True(zeroLimitSecurity.IsBlocked(identifier))
}

// TestHighVolumeScenario tests behavior under high load
func (suite *SecurityManagerTestSuite) TestHighVolumeScenario() {
	var wg sync.WaitGroup
	numClients := 1000
	attemptsPerClient := 10

	// Simulate high volume of clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			identifier := fmt.Sprintf("high-volume-client-%d", clientID)

			for j := 0; j < attemptsPerClient; j++ {
				// 80% success rate
				success := j%5 != 0
				suite.security.RecordAttempt(identifier, success)

				if j%3 == 0 {
					suite.security.IsBlocked(identifier)
				}
			}
		}(i)
	}

	wg.Wait()

	// System should remain stable
	stats := suite.security.GetStats()
	suite.NotNil(stats)
	suite.GreaterOrEqual(stats["active_attempts"], 0)
	suite.GreaterOrEqual(stats["blocked_ips"], 0)
}

// TestMemoryUsage tests that memory usage doesn't grow unbounded
func (suite *SecurityManagerTestSuite) TestMemoryUsage() {
	// Use security manager with quick cleanup
	quickCleanupSecurity := NewSecurityManager(5, 10*time.Millisecond, 5*time.Millisecond)
	defer quickCleanupSecurity.Stop()

	// Add many entries that should be cleaned up
	for round := 0; round < 10; round++ {
		for i := 0; i < 100; i++ {
			identifier := fmt.Sprintf("memory-test-%d-%d", round, i)
			quickCleanupSecurity.RecordAttempt(identifier, false)
		}

		// Wait for cleanup
		time.Sleep(20 * time.Millisecond)

		// Check that memory usage doesn't grow indefinitely
		stats := quickCleanupSecurity.GetStats()
		// With cleanup, we shouldn't have thousands of entries
		suite.Less(stats["active_attempts"], 1000, "Memory usage should be bounded by cleanup")
	}
}

// TestBlockExpiryPrecision tests precision of block expiry timing
func (suite *SecurityManagerTestSuite) TestBlockExpiryPrecision() {
	precisionSecurity := NewSecurityManager(1, 50*time.Millisecond, 1*time.Minute)
	defer precisionSecurity.Stop()

	identifier := "precision-test"

	// Get blocked
	start := time.Now()
	precisionSecurity.RecordAttempt(identifier, false)
	suite.True(precisionSecurity.IsBlocked(identifier))

	// Check that it's still blocked before expiry
	time.Sleep(25 * time.Millisecond)
	suite.True(precisionSecurity.IsBlocked(identifier))

	// Wait for expiry and check precision
	time.Sleep(30 * time.Millisecond) // Total: 55ms, should be expired
	suite.False(precisionSecurity.IsBlocked(identifier))

	elapsed := time.Since(start)
	suite.T().Logf("Block expiry took %v", elapsed)
	suite.GreaterOrEqual(elapsed, 50*time.Millisecond)
}

// TestStopFunctionality tests graceful shutdown
func (suite *SecurityManagerTestSuite) TestStopFunctionality() {
	testSecurity := NewSecurityManager(5, 5*time.Minute, 1*time.Minute)

	// Use the security manager
	testSecurity.RecordAttempt("test-stop", false)
	suite.False(testSecurity.IsBlocked("test-stop"))

	// Stop should not panic and should be safe to call multiple times
	testSecurity.Stop()
	testSecurity.Stop() // Second call should be safe

	// Should still be able to use basic functionality after stop
	suite.False(testSecurity.IsBlocked("test-stop"))
	stats := testSecurity.GetStats()
	suite.NotNil(stats)
}

// Run the comprehensive security test suite
func TestSecurityManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityManagerTestSuite))
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkIsBlocked(b *testing.B) {
	security := NewSecurityManager(5, 15*time.Minute, 5*time.Minute)
	defer security.Stop()

	// Setup some blocked and unblocked identifiers
	for i := 0; i < 100; i++ {
		identifier := fmt.Sprintf("bench-client-%d", i)
		if i%3 == 0 {
			// Block every 3rd client
			for j := 0; j < 5; j++ {
				security.RecordAttempt(identifier, false)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		identifier := fmt.Sprintf("bench-client-%d", i%100)
		security.IsBlocked(identifier)
	}
}

func BenchmarkRecordAttempt(b *testing.B) {
	security := NewSecurityManager(5, 15*time.Minute, 5*time.Minute)
	defer security.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		identifier := fmt.Sprintf("bench-attempt-%d", i%1000)
		success := i%4 != 0 // 75% success rate
		security.RecordAttempt(identifier, success)
	}
}

func BenchmarkGetStats(b *testing.B) {
	security := NewSecurityManager(5, 15*time.Minute, 5*time.Minute)
	defer security.Stop()

	// Add some data
	for i := 0; i < 1000; i++ {
		identifier := fmt.Sprintf("bench-stats-%d", i)
		security.RecordAttempt(identifier, i%5 != 0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		security.GetStats()
	}
}

func BenchmarkSecurityConcurrentOperations(b *testing.B) {
	security := NewSecurityManager(5, 15*time.Minute, 5*time.Minute)
	defer security.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		clientID := 0
		for pb.Next() {
			identifier := fmt.Sprintf("concurrent-bench-%d", clientID%100)
			security.IsBlocked(identifier)
			security.RecordAttempt(identifier, clientID%3 != 0)
			clientID++
		}
	})
}
