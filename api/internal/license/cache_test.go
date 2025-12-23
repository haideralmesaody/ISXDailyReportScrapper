package license

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// License Cache Deep Testing
// =============================================================================

func TestLicenseCacheDeep(t *testing.T) {
	t.Run("constructor validation", func(t *testing.T) {
		tests := []struct {
			name        string
			ttl         time.Duration
			maxSize     int
			description string
		}{
			{"standard config", 5 * time.Minute, 1000, "Standard production cache"},
			{"minimal config", 1 * time.Second, 1, "Minimal cache for testing"},
			{"large config", 24 * time.Hour, 100000, "Large cache for high traffic"},
			{"zero TTL", 0 * time.Second, 100, "Zero TTL cache"},
			{"zero size", 1 * time.Hour, 0, "Zero size cache"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cache := NewLicenseCache(tt.ttl, tt.maxSize)
				require.NotNil(t, cache, tt.description)

				stats := cache.GetStats()
				assert.Equal(t, tt.maxSize, stats["max_size"])
				assert.Equal(t, tt.ttl.Seconds(), stats["ttl_seconds"])
				assert.Equal(t, int64(0), stats["hit_count"])
				assert.Equal(t, int64(0), stats["miss_count"])
				assert.Equal(t, float64(0), stats["hit_ratio"])
			})
		}
	})

	t.Run("entry lifecycle", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 10)

		license := LicenseInfo{
			LicenseKey:  "LIFECYCLE-KEY",
			UserEmail:   "lifecycle@test.com",
			Status:      "Active",
			ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
			Duration:    "1m",
			IssuedDate:  time.Now().Add(-24 * time.Hour),
			LastChecked: time.Now(),
		}

		// Initial miss
		_, found := cache.Get("LIFECYCLE-KEY")
		assert.False(t, found)

		stats := cache.GetStats()
		assert.Equal(t, int64(1), stats["miss_count"])
		assert.Equal(t, int64(0), stats["hit_count"])

		// Set entry
		cache.Set("LIFECYCLE-KEY", license)

		// Hit
		retrieved, found := cache.Get("LIFECYCLE-KEY")
		assert.True(t, found)
		assert.Equal(t, license.LicenseKey, retrieved.LicenseKey)
		assert.Equal(t, license.UserEmail, retrieved.UserEmail)

		stats = cache.GetStats()
		assert.Equal(t, int64(1), stats["miss_count"])
		assert.Equal(t, int64(1), stats["hit_count"])
		assert.Equal(t, float64(0.5), stats["hit_ratio"]) // 1 hit out of 2 total

		// Multiple hits should increase hit count
		for i := 0; i < 5; i++ {
			_, found := cache.Get("LIFECYCLE-KEY")
			assert.True(t, found)
		}

		stats = cache.GetStats()
		assert.Equal(t, int64(6), stats["hit_count"]) // 1 + 5 additional hits
		assert.Equal(t, int64(1), stats["miss_count"])

		// Invalidate
		cache.Invalidate("LIFECYCLE-KEY")
		_, found = cache.Get("LIFECYCLE-KEY")
		assert.False(t, found)

		stats = cache.GetStats()
		assert.Equal(t, int64(6), stats["hit_count"])
		assert.Equal(t, int64(2), stats["miss_count"]) // 1 initial + 1 after invalidation
	})

	t.Run("TTL expiry precision", func(t *testing.T) {
		shortTTL := 100 * time.Millisecond
		cache := NewLicenseCache(shortTTL, 10)

		license := LicenseInfo{
			LicenseKey: "TTL-TEST",
			Status:     "Active",
		}

		// Set entry
		start := time.Now()
		cache.Set("TTL-TEST", license)

		// Should be available immediately
		_, found := cache.Get("TTL-TEST")
		assert.True(t, found)

		// Wait for TTL to expire
		time.Sleep(shortTTL + 10*time.Millisecond)

		// Should be expired
		_, found = cache.Get("TTL-TEST")
		assert.False(t, found)

		elapsed := time.Since(start)
		t.Logf("TTL test completed in %v", elapsed)
	})

	t.Run("eviction policy", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 3) // Small cache

		licenses := []LicenseInfo{
			{LicenseKey: "EVICT-1", Status: "Active"},
			{LicenseKey: "EVICT-2", Status: "Active"},
			{LicenseKey: "EVICT-3", Status: "Active"},
			{LicenseKey: "EVICT-4", Status: "Active"},
		}

		// Fill cache to capacity
		for i := 0; i < 3; i++ {
			cache.Set(licenses[i].LicenseKey, licenses[i])
			time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		}

		// All should be available
		for i := 0; i < 3; i++ {
			_, found := cache.Get(licenses[i].LicenseKey)
			assert.True(t, found, "License %d should be in cache", i)
		}

		// Add fourth item - should evict oldest (EVICT-1)
		cache.Set(licenses[3].LicenseKey, licenses[3])

		// EVICT-1 should be evicted
		_, found := cache.Get("EVICT-1")
		assert.False(t, found, "EVICT-1 should be evicted")

		// Others should remain
		for i := 1; i < 4; i++ {
			_, found := cache.Get(licenses[i].LicenseKey)
			assert.True(t, found, "License %d should still be in cache", i)
		}

		stats := cache.GetStats()
		assert.Equal(t, 3, stats["entries"]) // Should maintain max size
	})
}

func TestLicenseCacheEdgeCases(t *testing.T) {
	t.Run("empty key handling", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 10)
		license := LicenseInfo{LicenseKey: "", Status: "Active"}

		// Should handle empty keys gracefully
		cache.Set("", license)
		_, found := cache.Get("")
		assert.True(t, found) // Empty key should work

		cache.Invalidate("")
		_, found = cache.Get("")
		assert.False(t, found)
	})

	t.Run("very long keys", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 10)
		longKey := fmt.Sprintf("VERY-LONG-KEY-%s", string(make([]byte, 1000)))
		license := LicenseInfo{LicenseKey: longKey, Status: "Active"}

		cache.Set(longKey, license)
		retrieved, found := cache.Get(longKey)
		assert.True(t, found)
		assert.Equal(t, longKey, retrieved.LicenseKey)
	})

	t.Run("special characters in keys", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 10)
		specialKey := "KEY-@#$%^&*()_+{}|:<>?[]\\;'\",./"
		license := LicenseInfo{LicenseKey: specialKey, Status: "Active"}

		cache.Set(specialKey, license)
		retrieved, found := cache.Get(specialKey)
		assert.True(t, found)
		assert.Equal(t, specialKey, retrieved.LicenseKey)
	})

	t.Run("nil values handling", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 10)

		// Set with zero-value license
		var emptyLicense LicenseInfo
		cache.Set("EMPTY-LICENSE", emptyLicense)

		retrieved, found := cache.Get("EMPTY-LICENSE")
		assert.True(t, found)
		assert.Equal(t, "", retrieved.LicenseKey)
		assert.Equal(t, "", retrieved.Status)
	})

	t.Run("concurrent modification safety", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 100)
		license := LicenseInfo{LicenseKey: "CONCURRENT", Status: "Active"}

		var wg sync.WaitGroup
		numGoroutines := 50

		// Concurrent set operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				key := fmt.Sprintf("concurrent-%d", id)
				testLicense := license
				testLicense.LicenseKey = key
				cache.Set(key, testLicense)
			}(i)
		}

		// Concurrent get operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				key := fmt.Sprintf("concurrent-%d", id%20) // Some overlap
				_, _ = cache.Get(key)
			}(i)
		}

		// Concurrent invalidate operations
		for i := 0; i < numGoroutines/5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				key := fmt.Sprintf("concurrent-%d", id)
				cache.Invalidate(key)
			}(i)
		}

		// Concurrent stats operations
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = cache.GetStats()
			}()
		}

		wg.Wait()

		// Should complete without panics or deadlocks
		stats := cache.GetStats()
		assert.GreaterOrEqual(t, stats["entries"], 0)
		assert.LessOrEqual(t, stats["entries"], 100) // Should not exceed max size
	})
}

func TestLicenseCacheStatistics(t *testing.T) {
	t.Run("hit ratio calculations", func(t *testing.T) {
		_ = NewLicenseCache(1*time.Hour, 10) // This cache is not used in the test
		license := LicenseInfo{LicenseKey: "STATS", Status: "Active"}

		// Test various hit/miss patterns
		patterns := []struct {
			name          string
			hits          int
			misses        int
			expectedRatio float64
		}{
			{"all hits", 10, 0, 1.0},
			{"all misses", 0, 10, 0.0},
			{"half and half", 5, 5, 0.5},
			{"75% hits", 15, 5, 0.75},
			{"single operations", 1, 1, 0.5},
		}

		for _, pattern := range patterns {
			t.Run(pattern.name, func(t *testing.T) {
				cache := NewLicenseCache(1*time.Hour, 10) // Fresh cache
				cache.Set("HIT-KEY", license)

				// Generate misses
				for i := 0; i < pattern.misses; i++ {
					_, _ = cache.Get(fmt.Sprintf("MISS-KEY-%d", i))
				}

				// Generate hits
				for i := 0; i < pattern.hits; i++ {
					_, _ = cache.Get("HIT-KEY")
				}

				stats := cache.GetStats()
				assert.Equal(t, int64(pattern.hits), stats["hit_count"])
				assert.Equal(t, int64(pattern.misses), stats["miss_count"])
				assert.InDelta(t, pattern.expectedRatio, stats["hit_ratio"], 0.001)
			})
		}
	})

	t.Run("stats accuracy under load", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 50)

		// Pre-populate cache
		for i := 0; i < 25; i++ {
			license := LicenseInfo{
				LicenseKey: fmt.Sprintf("PRELOAD-%d", i),
				Status:     "Active",
			}
			cache.Set(license.LicenseKey, license)
		}
		_ = cache // Mark as used

		var wg sync.WaitGroup
		var totalHits, totalMisses int64
		var mu sync.Mutex

		// Simulate realistic usage patterns
		numGoroutines := 20
		operationsPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				localHits, localMisses := int64(0), int64(0)

				for j := 0; j < operationsPerGoroutine; j++ {
					// 70% chance of hitting existing keys
					if j%10 < 7 {
						key := fmt.Sprintf("PRELOAD-%d", j%25)
						_, found := cache.Get(key)
						if found {
							localHits++
						} else {
							localMisses++
						}
					} else {
						// Miss on non-existent key
						key := fmt.Sprintf("MISS-%d-%d", goroutineID, j)
						_, found := cache.Get(key)
						if found {
							localHits++
						} else {
							localMisses++
						}
					}
				}

				mu.Lock()
				totalHits += localHits
				totalMisses += localMisses
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		stats := cache.GetStats()
		reportedHits := stats["hit_count"].(int64)
		reportedMisses := stats["miss_count"].(int64)

		// Stats should match our tracking (accounting for initial misses during preload)
		assert.Equal(t, totalHits, reportedHits,
			"Reported hits (%d) should match tracked hits (%d)", reportedHits, totalHits)

		// Total misses should be at least our tracked misses (could be more due to preload misses)
		assert.GreaterOrEqual(t, reportedMisses, totalMisses,
			"Reported misses (%d) should be at least tracked misses (%d)", reportedMisses, totalMisses)

		// Hit ratio should be reasonable
		expectedRatio := float64(totalHits) / float64(totalHits+reportedMisses)
		actualRatio := stats["hit_ratio"].(float64)
		assert.InDelta(t, expectedRatio, actualRatio, 0.01,
			"Hit ratio mismatch: expected %.3f, actual %.3f", expectedRatio, actualRatio)
	})
}

func TestLicenseCacheCleanup(t *testing.T) {
	t.Run("automatic cleanup", func(t *testing.T) {
		shortTTL := 50 * time.Millisecond
		cache := NewLicenseCache(shortTTL, 100)

		// Add many entries
		numEntries := 50
		for i := 0; i < numEntries; i++ {
			license := LicenseInfo{
				LicenseKey: fmt.Sprintf("CLEANUP-%d", i),
				Status:     "Active",
			}
			cache.Set(license.LicenseKey, license)
		}

		stats := cache.GetStats()
		assert.Equal(t, numEntries, stats["entries"])

		// Wait for TTL to expire
		time.Sleep(shortTTL + 50*time.Millisecond)

		// Wait for cleanup cycle (cleanup runs every 5 minutes, but we can trigger it)
		// Add a new entry to potentially trigger cleanup
		cache.Set("TRIGGER-CLEANUP", LicenseInfo{LicenseKey: "TRIGGER", Status: "Active"})

		// Try to access expired entries - this should show they're expired
		expiredFound := 0
		for i := 0; i < numEntries; i++ {
			key := fmt.Sprintf("CLEANUP-%d", i)
			_, found := cache.Get(key)
			if found {
				expiredFound++
			}
		}

		// All expired entries should be gone or most should be gone
		assert.LessOrEqual(t, expiredFound, numEntries/10,
			"Most expired entries should be cleaned up, but found %d out of %d",
			expiredFound, numEntries)
	})

	t.Run("cleanup performance", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Millisecond, 10000) // Very short TTL, large capacity

		// Add many entries that will expire quickly
		numEntries := 1000
		for i := 0; i < numEntries; i++ {
			license := LicenseInfo{
				LicenseKey: fmt.Sprintf("PERF-CLEANUP-%d", i),
				Status:     "Active",
			}
			cache.Set(license.LicenseKey, license)
		}

		// Wait for entries to expire
		time.Sleep(10 * time.Millisecond)

		// Accessing cache should handle expired entries efficiently
		start := time.Now()

		// This should trigger cleanup of expired entries
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("PERF-CLEANUP-%d", i)
			_, _ = cache.Get(key)
		}

		duration := time.Since(start)

		// Should complete quickly even with many expired entries
		assert.Less(t, duration, 100*time.Millisecond,
			"Cleanup should be efficient, took %v", duration)
	})
}

func TestLicenseCacheMemoryBehavior(t *testing.T) {
	t.Run("memory bounds with eviction", func(t *testing.T) {
		maxSize := 10
		cache := NewLicenseCache(1*time.Hour, maxSize)

		// Add more entries than max size
		numEntries := maxSize * 3
		for i := 0; i < numEntries; i++ {
			license := LicenseInfo{
				LicenseKey: fmt.Sprintf("MEMORY-%d", i),
				Status:     "Active",
				UserEmail:  fmt.Sprintf("user%d@test.com", i), // Add some data
			}
			cache.Set(license.LicenseKey, license)
		}

		stats := cache.GetStats()
		assert.LessOrEqual(t, stats["entries"], maxSize,
			"Cache should not exceed max size, got %d entries", stats["entries"])

		// Only the most recent entries should remain
		recentFound := 0
		for i := numEntries - maxSize; i < numEntries; i++ {
			key := fmt.Sprintf("MEMORY-%d", i)
			_, found := cache.Get(key)
			if found {
				recentFound++
			}
		}

		// Most recent entries should be present
		assert.GreaterOrEqual(t, recentFound, maxSize/2,
			"Most recent entries should be in cache, found %d", recentFound)
	})

	t.Run("zero size cache behavior", func(t *testing.T) {
		cache := NewLicenseCache(1*time.Hour, 0) // Zero capacity

		license := LicenseInfo{LicenseKey: "ZERO-SIZE", Status: "Active"}

		// Should handle zero capacity gracefully
		cache.Set("ZERO-SIZE", license)

		stats := cache.GetStats()
		assert.Equal(t, 0, stats["entries"]) // Should not store anything

		_, found := cache.Get("ZERO-SIZE")
		assert.False(t, found) // Should not find anything

		_ = cache // Ensure cache is used
	})
}
