package license

import (
	"strings"
	"sync"
	"time"
)

// ValidationCache manages cached license validation results with concurrency protection.
type ValidationCache struct {
	mu         sync.RWMutex
	result     *ValidationResult
	lastUpdate time.Time
}

// NewValidationCache creates an empty validation cache.
func NewValidationCache() *ValidationCache {
	return &ValidationCache{}
}

// GetValidResult returns the cached result when it is still valid.
func (vc *ValidationCache) GetValidResult() (*ValidationResult, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.result == nil {
		return nil, false
	}

	if vc.result.CachedUntil.IsZero() || time.Now().After(vc.result.CachedUntil) {
		return nil, false
	}

	resultCopy := *vc.result
	return &resultCopy, true
}

// Snapshot returns the current cached result regardless of expiration.
func (vc *ValidationCache) Snapshot() (*ValidationResult, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.result == nil {
		return nil, false
	}

	resultCopy := *vc.result
	return &resultCopy, true
}

// StoreResult caches the validation outcome using the same TTL rules as the legacy implementation.
func (vc *ValidationCache) StoreResult(isValid bool, err error) *ValidationResult {
	result := &ValidationResult{
		IsValid: isValid,
		Error:   err,
	}

	now := time.Now()
	switch {
	case isValid:
		result.CachedUntil = now.Add(5 * time.Minute)
	case err != nil:
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "expired") {
			result.ErrorType = "expired"
			result.CachedUntil = now.Add(1 * time.Hour)
			result.RetryAfter = 5 * time.Minute
		} else {
			result.ErrorType = "network_error"
			result.CachedUntil = now.Add(2 * time.Minute)
			result.RetryAfter = 30 * time.Second
		}
	}

	vc.mu.Lock()
	vc.result = result
	vc.lastUpdate = now
	vc.mu.Unlock()
	return result
}

// Invalidate clears the cache and returns the previous result if present.
func (vc *ValidationCache) Invalidate() (*ValidationResult, bool) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.result == nil {
		return nil, false
	}

	previous := *vc.result
	vc.result = nil
	vc.lastUpdate = time.Time{}
	return &previous, true
}
