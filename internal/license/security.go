package license

import (
	"sync"
	"time"
)

// SecurityManager handles rate limiting and security protection
type SecurityManager struct {
	attemptCounts   map[string]int
	lastAttempts    map[string]time.Time
	blockedIPs      map[string]time.Time
	mutex           sync.RWMutex
	maxAttempts     int
	blockDuration   time.Duration
	windowDuration  time.Duration
	cleanupInterval time.Duration
	logger          *Logger
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(maxAttempts int, blockDuration, windowDuration time.Duration, logger *Logger) *SecurityManager {
	sm := &SecurityManager{
		attemptCounts:   make(map[string]int),
		lastAttempts:    make(map[string]time.Time),
		blockedIPs:      make(map[string]time.Time),
		maxAttempts:     maxAttempts,
		blockDuration:   blockDuration,
		windowDuration:  windowDuration,
		cleanupInterval: 5 * time.Minute,
		logger:          logger,
	}

	go sm.cleanup()

	return sm
}

// IsBlocked checks if an identifier is currently blocked
func (s *SecurityManager) IsBlocked(identifier string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if blockTime, exists := s.blockedIPs[identifier]; exists {
		if time.Since(blockTime) < s.blockDuration {
			return true
		}
		delete(s.blockedIPs, identifier)
	}
	return false
}

// RecordAttempt records a license operation attempt
func (s *SecurityManager) RecordAttempt(identifier string, success bool) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()

	if success {
		delete(s.attemptCounts, identifier)
		delete(s.lastAttempts, identifier)
		return true
	}

	if lastAttempt, exists := s.lastAttempts[identifier]; exists {
		if now.Sub(lastAttempt) > s.windowDuration {
			s.attemptCounts[identifier] = 1
		} else {
			s.attemptCounts[identifier]++
		}
	} else {
		s.attemptCounts[identifier] = 1
	}

	s.lastAttempts[identifier] = now

	if s.attemptCounts[identifier] >= s.maxAttempts {
		s.blockedIPs[identifier] = now

		if s.logger != nil {
			s.logger.Log(LogEntry{
				Level:     LogLevelWarn,
				Action:    "security_violation",
				Result:    "IP blocked due to too many failed attempts",
				IPAddress: identifier,
				Details: map[string]interface{}{
					"attempt_count": s.attemptCounts[identifier],
					"max_attempts":  s.maxAttempts,
				},
			})
		}

		return false
	}

	return true
}

// GetStats returns security statistics
func (s *SecurityManager) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"active_attempts": len(s.attemptCounts),
		"blocked_ips":     len(s.blockedIPs),
		"max_attempts":    s.maxAttempts,
		"block_duration":  s.blockDuration.String(),
		"window_duration": s.windowDuration.String(),
	}
}

// cleanup periodically removes old entries
func (s *SecurityManager) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()

		for identifier, lastAttempt := range s.lastAttempts {
			if now.Sub(lastAttempt) > s.windowDuration {
				delete(s.attemptCounts, identifier)
				delete(s.lastAttempts, identifier)
			}
		}

		for identifier, blockTime := range s.blockedIPs {
			if now.Sub(blockTime) > s.blockDuration {
				delete(s.blockedIPs, identifier)
			}
		}

		s.mutex.Unlock()
	}
}
