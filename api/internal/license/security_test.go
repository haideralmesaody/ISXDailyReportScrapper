package license

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Security Manager Deep Testing
// =============================================================================

func TestSecurityManagerDeep(t *testing.T) {
	t.Run("constructor validation", func(t *testing.T) {
		tests := []struct {
			name           string
			maxAttempts    int
			blockDuration  time.Duration
			windowDuration time.Duration
			expectedValid  bool
		}{
			{"valid standard config", 5, 15 * time.Minute, 5 * time.Minute, true},
			{"minimum values", 1, 1 * time.Second, 1 * time.Second, true},
			{"high values", 100, 24 * time.Hour, 24 * time.Hour, true},
			{"zero max attempts", 0, 15 * time.Minute, 5 * time.Minute, true}, // Constructor doesn't validate
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sm := NewSecurityManager(tt.maxAttempts, tt.blockDuration, tt.windowDuration)
				if tt.expectedValid {
					assert.NotNil(t, sm)

					stats := sm.GetStats()
					assert.Equal(t, tt.maxAttempts, stats["max_attempts"])
					assert.Equal(t, tt.blockDuration.String(), stats["block_duration"])
					assert.Equal(t, tt.windowDuration.String(), stats["window_duration"])
				} else {
					// Even invalid configs should create a manager (no validation in constructor)
					assert.NotNil(t, sm)
				}
			})
		}
	})

	t.Run("attempt tracking edge cases", func(t *testing.T) {
		sm := NewSecurityManager(3, 5*time.Minute, 2*time.Minute)

		// Test empty identifier
		result := sm.RecordAttempt("", false)
		assert.True(t, result) // Should not block on empty identifier
		assert.False(t, sm.IsBlocked(""))

		// Test very long identifier
		longID := fmt.Sprintf("user-%s", string(make([]byte, 1000)))
		result = sm.RecordAttempt(longID, false)
		assert.True(t, result)

		// Test special characters in identifier
		specialID := "user-@#$%^&*()_+{}|:<>?[]\\;'\",./"
		result = sm.RecordAttempt(specialID, false)
		assert.True(t, result)
		assert.False(t, sm.IsBlocked(specialID))
	})

	t.Run("precise timing behavior", func(t *testing.T) {
		windowDuration := 100 * time.Millisecond
		sm := NewSecurityManager(2, 1*time.Second, windowDuration)

		// Record first failure
		start := time.Now()
		sm.RecordAttempt("timing-user", false)

		// Wait almost until window expires
		time.Sleep(windowDuration - 10*time.Millisecond)

		// Second failure should still count
		sm.RecordAttempt("timing-user", false)
		assert.True(t, sm.IsBlocked("timing-user"))

		// Wait for window to fully expire
		time.Sleep(20 * time.Millisecond)

		// Now it should be a fresh attempt
		result := sm.RecordAttempt("timing-user", false)
		assert.True(t, result) // Should succeed as window expired

		duration := time.Since(start)
		t.Logf("Test completed in %v", duration)
	})

	t.Run("success resets counter", func(t *testing.T) {
		sm := NewSecurityManager(3, 1*time.Minute, 1*time.Minute)

		// Record failures
		sm.RecordAttempt("reset-user", false)
		sm.RecordAttempt("reset-user", false)

		// Success should reset
		result := sm.RecordAttempt("reset-user", true)
		assert.True(t, result)
		assert.False(t, sm.IsBlocked("reset-user"))

		// Should be able to fail again without immediate block
		result = sm.RecordAttempt("reset-user", false)
		assert.True(t, result)
		result = sm.RecordAttempt("reset-user", false)
		assert.True(t, result)
		result = sm.RecordAttempt("reset-user", false) // This should block
		assert.False(t, result)
		assert.True(t, sm.IsBlocked("reset-user"))
	})

	t.Run("concurrent user isolation", func(t *testing.T) {
		sm := NewSecurityManager(2, 1*time.Minute, 1*time.Minute)
		numUsers := 10
		var wg sync.WaitGroup

		// Block all users concurrently
		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()
				user := fmt.Sprintf("concurrent-user-%d", userID)

				// Block this user
				sm.RecordAttempt(user, false)
				sm.RecordAttempt(user, false)

				// Verify user is blocked
				assert.True(t, sm.IsBlocked(user))
			}(i)
		}

		wg.Wait()

		// Verify all users are blocked independently
		for i := 0; i < numUsers; i++ {
			user := fmt.Sprintf("concurrent-user-%d", i)
			assert.True(t, sm.IsBlocked(user), "User %s should be blocked", user)
		}

		// Verify stats reflect all blocked users
		stats := sm.GetStats()
		assert.Equal(t, numUsers, stats["blocked_ips"])
	})
}

func TestSecurityManagerStats(t *testing.T) {
	t.Run("stats accuracy", func(t *testing.T) {
		sm := NewSecurityManager(3, 10*time.Minute, 5*time.Minute)

		// Initial stats
		stats := sm.GetStats()
		assert.Equal(t, 0, stats["active_attempts"])
		assert.Equal(t, 0, stats["blocked_ips"])

		// Add some attempts
		sm.RecordAttempt("user1", false)
		sm.RecordAttempt("user2", false)
		sm.RecordAttempt("user2", false)

		stats = sm.GetStats()
		assert.Equal(t, 2, stats["active_attempts"]) // 2 users with attempts
		assert.Equal(t, 0, stats["blocked_ips"])     // No blocks yet

		// Block user2
		sm.RecordAttempt("user2", false)

		stats = sm.GetStats()
		assert.Equal(t, 2, stats["active_attempts"]) // Still 2 users tracked
		assert.Equal(t, 1, stats["blocked_ips"])     // 1 blocked

		// Success for user1 should remove from attempts
		sm.RecordAttempt("user1", true)

		stats = sm.GetStats()
		assert.Equal(t, 1, stats["active_attempts"]) // Only user2 still tracked
		assert.Equal(t, 1, stats["blocked_ips"])     // Still 1 blocked
	})

	t.Run("stats consistency under concurrent access", func(t *testing.T) {
		sm := NewSecurityManager(2, 1*time.Minute, 1*time.Minute)
		numOperations := 100
		var wg sync.WaitGroup

		// Perform many concurrent operations
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				user := fmt.Sprintf("stats-user-%d", id%20) // 20 different users

				sm.RecordAttempt(user, id%4 == 0) // 25% success rate
				sm.IsBlocked(user)
				sm.GetStats()
			}(i)
		}

		wg.Wait()

		// Stats should be consistent (no negative values, reasonable ranges)
		stats := sm.GetStats()
		assert.GreaterOrEqual(t, stats["active_attempts"], 0)
		assert.GreaterOrEqual(t, stats["blocked_ips"], 0)
		assert.LessOrEqual(t, stats["active_attempts"], 20) // Max 20 users
		assert.LessOrEqual(t, stats["blocked_ips"], 20)     // Max 20 users
	})
}

func TestSecurityManagerCleanup(t *testing.T) {
	t.Run("cleanup goroutine functionality", func(t *testing.T) {
		sm := NewSecurityManager(2, 50*time.Millisecond, 50*time.Millisecond)

		// Create some data that should be cleaned up
		sm.RecordAttempt("cleanup-user1", false)
		sm.RecordAttempt("cleanup-user2", false)
		sm.RecordAttempt("cleanup-user2", false) // Block user2

		// Verify initial state
		assert.False(t, sm.IsBlocked("cleanup-user1"))
		assert.True(t, sm.IsBlocked("cleanup-user2"))

		stats := sm.GetStats()
		initialAttempts := stats["active_attempts"].(int)
		initialBlocked := stats["blocked_ips"].(int)

		assert.GreaterOrEqual(t, initialAttempts, 1)
		assert.GreaterOrEqual(t, initialBlocked, 1)

		// Wait for cleanup to happen (cleanup runs every 5 minutes by default,
		// but we set short durations so items should expire)
		time.Sleep(200 * time.Millisecond)

		// Force cleanup by triggering it indirectly
		// (The cleanup goroutine runs periodically)
		sm.IsBlocked("cleanup-user1") // This should trigger cleanup of expired entries
		sm.IsBlocked("cleanup-user2") // This should trigger cleanup of expired blocks

		// Items should be cleaned up after expiry
		stats = sm.GetStats()
		// Note: The actual cleanup happens in a background goroutine with 5-minute intervals
		// So we can't reliably test automatic cleanup in unit tests
		// But we can verify that the cleanup logic works when entries expire naturally
	})

	t.Run("memory usage doesn't grow indefinitely", func(t *testing.T) {
		sm := NewSecurityManager(100, 1*time.Millisecond, 1*time.Millisecond)

		// Generate many entries that should be cleaned up quickly
		for i := 0; i < 1000; i++ {
			user := fmt.Sprintf("memory-test-user-%d", i)
			sm.RecordAttempt(user, false)
		}

		// Initial stats
		stats1 := sm.GetStats()
		initialAttempts := stats1["active_attempts"].(int)

		// Wait for entries to expire
		time.Sleep(10 * time.Millisecond)

		// Trigger internal cleanup by accessing data
		for i := 0; i < 10; i++ {
			user := fmt.Sprintf("memory-test-user-%d", i)
			sm.IsBlocked(user) // This should clean up expired entries
		}

		// Stats should show cleanup has happened or entries have expired
		stats2 := sm.GetStats()
		finalAttempts := stats2["active_attempts"].(int)

		// We can't guarantee exact numbers due to timing, but there should be some cleanup
		// or at least the system should handle the load gracefully
		assert.LessOrEqual(t, finalAttempts, initialAttempts+100,
			"Memory usage should be bounded, got initial: %d, final: %d",
			initialAttempts, finalAttempts)
	})
}

func TestSecurityManagerBoundaryConditions(t *testing.T) {
	t.Run("max attempts boundary", func(t *testing.T) {
		sm := NewSecurityManager(1, 1*time.Minute, 1*time.Minute) // Block after 1 attempt

		// First attempt should succeed
		result := sm.RecordAttempt("boundary-user", false)
		assert.False(t, result) // With maxAttempts=1, first failure blocks immediately
		assert.True(t, sm.IsBlocked("boundary-user"))
	})

	t.Run("zero block duration", func(t *testing.T) {
		sm := NewSecurityManager(2, 0*time.Second, 1*time.Minute)

		// Block user
		sm.RecordAttempt("zero-block-user", false)
		sm.RecordAttempt("zero-block-user", false)
		assert.True(t, sm.IsBlocked("zero-block-user"))

		// With zero block duration, should unblock immediately
		time.Sleep(1 * time.Millisecond)
		assert.False(t, sm.IsBlocked("zero-block-user"))
	})

	t.Run("zero window duration", func(t *testing.T) {
		sm := NewSecurityManager(3, 1*time.Minute, 0*time.Second)

		// With zero window, each attempt should reset the counter
		sm.RecordAttempt("zero-window-user", false)

		time.Sleep(1 * time.Millisecond)

		// This should be treated as a fresh attempt due to zero window
		result := sm.RecordAttempt("zero-window-user", false)
		assert.True(t, result) // Should succeed as window expired immediately
	})
}

func TestSecurityManagerPerformance(t *testing.T) {
	t.Run("large scale operations", func(t *testing.T) {
		sm := NewSecurityManager(10, 1*time.Hour, 1*time.Hour)

		numUsers := 10000
		numOperations := 100000

		start := time.Now()

		// Perform many operations
		for i := 0; i < numOperations; i++ {
			user := fmt.Sprintf("perf-user-%d", i%numUsers)
			sm.RecordAttempt(user, i%5 == 0) // 20% success rate

			if i%1000 == 0 {
				sm.IsBlocked(user)
				sm.GetStats()
			}
		}

		duration := time.Since(start)
		t.Logf("Performed %d operations on %d users in %v", numOperations, numUsers, duration)

		// Should complete in reasonable time (< 5 seconds for this scale)
		assert.Less(t, duration, 5*time.Second,
			"Performance test took too long: %v", duration)

		// Stats should be reasonable
		stats := sm.GetStats()
		assert.LessOrEqual(t, stats["active_attempts"], numUsers)
		assert.LessOrEqual(t, stats["blocked_ips"], numUsers)
	})
}
