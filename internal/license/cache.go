package license

import (
	"sync"
	"time"
)

// CacheEntry represents a cached license validation result
type CacheEntry struct {
	LicenseInfo LicenseInfo `json:"license_info"`
	CachedAt    time.Time   `json:"cached_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	HitCount    int         `json:"hit_count"`
}

// LicenseCache provides intelligent caching for license validations
type LicenseCache struct {
	entries   map[string]CacheEntry
	mutex     sync.RWMutex
	ttl       time.Duration
	maxSize   int
	hitCount  int64
	missCount int64
}

// NewLicenseCache creates a new license cache
func NewLicenseCache(ttl time.Duration, maxSize int) *LicenseCache {
	cache := &LicenseCache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	go cache.cleanup()

	return cache
}

// Get retrieves a license from cache
func (c *LicenseCache) Get(licenseKey string) (*LicenseInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[licenseKey]
	if !exists || time.Now().After(entry.ExpiresAt) {
		c.missCount++
		return nil, false
	}

	entry.HitCount++
	c.entries[licenseKey] = entry
	c.hitCount++

	return &entry.LicenseInfo, true
}

// Set stores a license in cache
func (c *LicenseCache) Set(licenseKey string, info LicenseInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[licenseKey] = CacheEntry{
		LicenseInfo: info,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(c.ttl),
		HitCount:    0,
	}
}

// Invalidate removes a license from cache
func (c *LicenseCache) Invalidate(licenseKey string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.entries, licenseKey)
}

// GetStats returns cache statistics
func (c *LicenseCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	totalRequests := c.hitCount + c.missCount
	hitRatio := float64(0)
	if totalRequests > 0 {
		hitRatio = float64(c.hitCount) / float64(totalRequests)
	}

	return map[string]interface{}{
		"entries":     len(c.entries),
		"max_size":    c.maxSize,
		"hit_count":   c.hitCount,
		"miss_count":  c.missCount,
		"hit_ratio":   hitRatio,
		"ttl_seconds": c.ttl.Seconds(),
	}
}

func (c *LicenseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CachedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *LicenseCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mutex.Unlock()
	}
}
