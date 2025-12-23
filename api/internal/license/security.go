package license

import (
	"context"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/isxcli/isxcli/internal/infrastructure"
)

// SecurityManager handles advanced rate limiting and anti-abuse protection
type SecurityManager struct {
	// Legacy rate limiting (kept for compatibility)
	attemptCounts      map[string]int
	lastAttempts       map[string]time.Time
	blockedIPs         map[string]time.Time
	attemptWindowStart map[string]time.Time

	// Advanced sliding window rate limiting
	slidingWindows map[string]*SlidingWindow

	// Anti-abuse tracking
	deviceAttempts   map[string]*DeviceAttemptTracker
	suspiciousIPs    map[string]*SuspiciousActivity
	honeypotLicenses map[string]time.Time

	// Security configuration
	mutex           sync.RWMutex
	maxAttempts     int
	blockDuration   time.Duration
	windowDuration  time.Duration
	cleanupInterval time.Duration
	stopChan        chan struct{}
	stopOnce        sync.Once

	// Enhanced security settings
	maxAttemptsPerIP         int
	maxAttemptsPerDevice     int
	slidingWindowSize        time.Duration
	permanentBanThreshold    int
	honeypotDetectionEnabled bool
	patternDetectionEnabled  bool

	// Security patterns
	suspiciousPatterns []*regexp.Regexp

	// logger removed per CLAUDE.md - use infrastructure logger
}

// NewSecurityManager creates an enhanced security manager with advanced anti-abuse features
func NewSecurityManager(maxAttempts int, blockDuration, windowDuration time.Duration) *SecurityManager {
	cleanupInterval := 5 * time.Minute
	if windowDuration > 0 {
		cleanupInterval = windowDuration
	}

	sm := &SecurityManager{
		// Legacy fields
		attemptCounts:      make(map[string]int),
		lastAttempts:       make(map[string]time.Time),
		blockedIPs:         make(map[string]time.Time),
		attemptWindowStart: make(map[string]time.Time),
		maxAttempts:        maxAttempts,
		blockDuration:      blockDuration,
		windowDuration:     windowDuration,
		cleanupInterval:    cleanupInterval,
		stopChan:           make(chan struct{}),

		// Enhanced fields
		slidingWindows:   make(map[string]*SlidingWindow),
		deviceAttempts:   make(map[string]*DeviceAttemptTracker),
		suspiciousIPs:    make(map[string]*SuspiciousActivity),
		honeypotLicenses: make(map[string]time.Time),

		// Enhanced configuration
		maxAttemptsPerIP:         100, // 100 attempts per hour per IP
		maxAttemptsPerDevice:     10,  // 10 attempts per hour per device
		slidingWindowSize:        1 * time.Hour,
		permanentBanThreshold:    1000, // Permanent ban after 1000 failed attempts
		honeypotDetectionEnabled: true,
		patternDetectionEnabled:  true,

		// logger removed per CLAUDE.md - use infrastructure logger
	}

	// Initialize suspicious patterns
	sm.initializeSuspiciousPatterns()

	// Initialize honeypot licenses
	sm.initializeHoneypotLicenses()

	go sm.cleanup()

	return sm
}

// blockExpiryTime calculates when a block should expire; zero/negative durations use a minimal interval.
func (s *SecurityManager) blockExpiryTime(now time.Time) time.Time {
	if s.blockDuration <= 0 {
		return now.Add(time.Millisecond)
	}
	return now.Add(s.blockDuration)
}

// IsBlocked checks if an identifier is currently blocked
func (s *SecurityManager) IsBlocked(identifier string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if blockExpiry, exists := s.blockedIPs[identifier]; exists {
		// If the rate-limit window has rolled over, clear the block early
		if s.windowDuration > 0 {
			if windowStart, ok := s.attemptWindowStart[identifier]; ok && time.Since(windowStart) > s.windowDuration {
				delete(s.blockedIPs, identifier)
				return false
			}
		}

		if time.Now().Before(blockExpiry) {
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

	// Ensure maps are initialized when SecurityManager is constructed manually in tests
	if s.attemptWindowStart == nil {
		s.attemptWindowStart = make(map[string]time.Time)
	}

	if success {
		delete(s.attemptCounts, identifier)
		delete(s.lastAttempts, identifier)
		delete(s.blockedIPs, identifier)
		delete(s.attemptWindowStart, identifier)
		return true
	}

	windowStart, exists := s.attemptWindowStart[identifier]
	if !exists || now.Sub(windowStart) > s.windowDuration {
		s.attemptWindowStart[identifier] = now
		s.attemptCounts[identifier] = 1
		delete(s.blockedIPs, identifier)
	} else {
		s.attemptCounts[identifier]++
	}
	s.lastAttempts[identifier] = now

	if s.attemptCounts[identifier] >= s.maxAttempts {
		s.blockedIPs[identifier] = s.blockExpiryTime(now)

		// Log security violation per CLAUDE.md standards
		ctx := context.Background()
		logger := infrastructure.LoggerWithContext(ctx)
		logger.WarnContext(ctx, "IP blocked due to too many failed attempts",
			slog.String("action", "security_violation"),
			slog.String("ip_address", identifier),
			slog.Int("attempt_count", s.attemptCounts[identifier]),
			slog.Int("max_attempts", s.maxAttempts),
		)

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

// Stop gracefully stops the security manager cleanup goroutine
func (s *SecurityManager) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// cleanup periodically removes old entries and performs security maintenance
func (s *SecurityManager) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			now := time.Now()

			// Legacy cleanup
			for identifier, lastAttempt := range s.lastAttempts {
				if now.Sub(lastAttempt) > s.windowDuration {
					delete(s.attemptCounts, identifier)
					delete(s.lastAttempts, identifier)
					delete(s.attemptWindowStart, identifier)
				}
			}

			for identifier, blockExpiry := range s.blockedIPs {
				if now.After(blockExpiry) {
					delete(s.blockedIPs, identifier)
				}
			}

			// Enhanced cleanup
			s.cleanupSlidingWindows(now)
			s.cleanupDeviceTrackers(now)
			s.cleanupSuspiciousActivity(now)
			s.cleanupHoneypotLicenses(now)

			s.mutex.Unlock()
		case <-s.stopChan:
			return
		}
	}
}

// SlidingWindow implements a sliding window rate limiter
type SlidingWindow struct {
	attempts    []time.Time
	maxAttempts int
	windowSize  time.Duration
	lastCleanup time.Time
}

// DeviceAttemptTracker tracks attempts per device fingerprint
type DeviceAttemptTracker struct {
	fingerprint     string
	attempts        []AttemptRecord
	totalAttempts   int
	failedAttempts  int
	lastAttempt     time.Time
	firstSeen       time.Time
	suspiciousScore int
	blocked         bool
	blockExpiry     time.Time
}

// AttemptRecord represents a single attempt record
type AttemptRecord struct {
	timestamp  time.Time
	success    bool
	licenseKey string
	clientIP   string
	userAgent  string
	errorType  string
}

// SuspiciousActivity tracks suspicious IP behavior
type SuspiciousActivity struct {
	ip             string
	attempts       []AttemptRecord
	uniqueDevices  map[string]bool
	uniqueLicenses map[string]bool
	patterns       []string
	riskScore      int
	firstSeen      time.Time
	lastActivity   time.Time
	blocked        bool
	permanent      bool
}

// IsBlocked checks if an identifier is currently blocked (enhanced version)
func (s *SecurityManager) IsBlockedEnhanced(identifier, deviceFingerprint, clientIP string) (bool, string, time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check legacy blocking first
	if blockExpiry, exists := s.blockedIPs[identifier]; exists {
		if time.Now().Before(blockExpiry) {
			remaining := time.Until(blockExpiry)
			return true, "legacy_rate_limit", remaining
		}
		delete(s.blockedIPs, identifier)
	}

	// Check IP-based sliding window
	if window, exists := s.slidingWindows[clientIP]; exists {
		if s.isWindowExceeded(window) {
			return true, "ip_rate_limit", s.slidingWindowSize
		}
	}

	// Check device-based rate limiting
	if tracker, exists := s.deviceAttempts[deviceFingerprint]; exists {
		if tracker.blocked && time.Now().Before(tracker.blockExpiry) {
			remaining := time.Until(tracker.blockExpiry)
			return true, "device_blocked", remaining
		}
	}

	// Check suspicious IP blocking
	if suspicious, exists := s.suspiciousIPs[clientIP]; exists {
		if suspicious.blocked {
			if suspicious.permanent {
				return true, "permanent_ban", 0
			}
			// Temporary suspicious activity block
			return true, "suspicious_activity", 24 * time.Hour
		}
	}

	// Check honeypot detection
	if s.honeypotDetectionEnabled {
		// This would be checked when a license key is provided
	}

	return false, "", 0
}

// RecordAttemptEnhanced records an attempt with enhanced tracking
func (s *SecurityManager) RecordAttemptEnhanced(identifier, deviceFingerprint, clientIP, licenseKey, userAgent string, success bool, errorType string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	ctx := context.Background()
	logger := infrastructure.LoggerWithContext(ctx)

	// Record legacy attempt
	legacyResult := s.recordLegacyAttempt(identifier, success, logger, ctx)

	// Record in sliding window for IP
	s.recordSlidingWindowAttempt(clientIP, now)

	// Record device attempt
	s.recordDeviceAttempt(deviceFingerprint, clientIP, licenseKey, userAgent, success, errorType, now)

	// Update suspicious activity tracking
	s.updateSuspiciousActivity(clientIP, deviceFingerprint, licenseKey, userAgent, success, errorType, now)

	// Check for honeypot license
	if s.honeypotDetectionEnabled && s.isHoneypotLicense(licenseKey) {
		s.handleHoneypotDetection(clientIP, deviceFingerprint, licenseKey, logger, ctx)
		return false
	}

	// Pattern detection
	if s.patternDetectionEnabled {
		s.detectSuspiciousPatterns(clientIP, deviceFingerprint, licenseKey, userAgent)
	}

	// Security event logging
	s.logSecurityEvent(SecurityEventType{
		Type:              "license_attempt",
		Identifier:        identifier,
		DeviceFingerprint: deviceFingerprint,
		ClientIP:          clientIP,
		LicenseKey:        licenseKey,
		UserAgent:         userAgent,
		Success:           success,
		ErrorType:         errorType,
		Timestamp:         now,
	}, logger, ctx)

	return legacyResult
}

// recordLegacyAttempt handles the original rate limiting logic
func (s *SecurityManager) recordLegacyAttempt(identifier string, success bool, logger *slog.Logger, ctx context.Context) bool {
	now := time.Now()

	if success {
		delete(s.attemptCounts, identifier)
		delete(s.lastAttempts, identifier)
		delete(s.blockedIPs, identifier)
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
		s.blockedIPs[identifier] = s.blockExpiryTime(now)

		logger.WarnContext(ctx, "IP blocked due to too many failed attempts",
			slog.String("action", "security_violation"),
			slog.String("ip_address", identifier),
			slog.Int("attempt_count", s.attemptCounts[identifier]),
			slog.Int("max_attempts", s.maxAttempts),
		)

		return false
	}

	return true
}

// recordSlidingWindowAttempt updates the sliding window for IP-based rate limiting
func (s *SecurityManager) recordSlidingWindowAttempt(clientIP string, timestamp time.Time) {
	if window, exists := s.slidingWindows[clientIP]; exists {
		// Clean old attempts outside the window
		s.cleanupWindow(window, timestamp)
		window.attempts = append(window.attempts, timestamp)
	} else {
		s.slidingWindows[clientIP] = &SlidingWindow{
			attempts:    []time.Time{timestamp},
			maxAttempts: s.maxAttemptsPerIP,
			windowSize:  s.slidingWindowSize,
			lastCleanup: timestamp,
		}
	}
}

// recordDeviceAttempt tracks attempts per device fingerprint
func (s *SecurityManager) recordDeviceAttempt(fingerprint, clientIP, licenseKey, userAgent string, success bool, errorType string, timestamp time.Time) {
	record := AttemptRecord{
		timestamp:  timestamp,
		success:    success,
		licenseKey: licenseKey,
		clientIP:   clientIP,
		userAgent:  userAgent,
		errorType:  errorType,
	}

	if tracker, exists := s.deviceAttempts[fingerprint]; exists {
		tracker.attempts = append(tracker.attempts, record)
		tracker.totalAttempts++
		tracker.lastAttempt = timestamp

		if !success {
			tracker.failedAttempts++
		}

		// Check if device should be blocked
		if s.shouldBlockDevice(tracker) {
			tracker.blocked = true
			tracker.blockExpiry = timestamp.Add(s.blockDuration)
		}
	} else {
		s.deviceAttempts[fingerprint] = &DeviceAttemptTracker{
			fingerprint:   fingerprint,
			attempts:      []AttemptRecord{record},
			totalAttempts: 1,
			failedAttempts: func() int {
				if success {
					return 0
				} else {
					return 1
				}
			}(),
			lastAttempt:     timestamp,
			firstSeen:       timestamp,
			suspiciousScore: 0,
			blocked:         false,
		}
	}
}

// updateSuspiciousActivity updates suspicious activity tracking for IPs
func (s *SecurityManager) updateSuspiciousActivity(clientIP, deviceFingerprint, licenseKey, userAgent string, success bool, errorType string, timestamp time.Time) {
	record := AttemptRecord{
		timestamp:  timestamp,
		success:    success,
		licenseKey: licenseKey,
		clientIP:   clientIP,
		userAgent:  userAgent,
		errorType:  errorType,
	}

	if suspicious, exists := s.suspiciousIPs[clientIP]; exists {
		suspicious.attempts = append(suspicious.attempts, record)
		suspicious.lastActivity = timestamp

		// Track unique devices and licenses
		suspicious.uniqueDevices[deviceFingerprint] = true
		suspicious.uniqueLicenses[licenseKey] = true

		// Update risk score
		s.updateRiskScore(suspicious, record)

		// Check for permanent ban
		if suspicious.riskScore >= s.permanentBanThreshold {
			suspicious.blocked = true
			suspicious.permanent = true
		}
	} else {
		s.suspiciousIPs[clientIP] = &SuspiciousActivity{
			ip:             clientIP,
			attempts:       []AttemptRecord{record},
			uniqueDevices:  map[string]bool{deviceFingerprint: true},
			uniqueLicenses: map[string]bool{licenseKey: true},
			patterns:       []string{},
			riskScore:      s.calculateInitialRiskScore(record),
			firstSeen:      timestamp,
			lastActivity:   timestamp,
			blocked:        false,
			permanent:      false,
		}
	}
}

// isWindowExceeded checks if the sliding window rate limit is exceeded
func (s *SecurityManager) isWindowExceeded(window *SlidingWindow) bool {
	now := time.Now()
	s.cleanupWindow(window, now)
	return len(window.attempts) >= window.maxAttempts
}

// cleanupWindow removes old attempts from sliding window
func (s *SecurityManager) cleanupWindow(window *SlidingWindow, now time.Time) {
	cutoff := now.Add(-window.windowSize)
	validAttempts := make([]time.Time, 0, len(window.attempts))

	for _, attempt := range window.attempts {
		if attempt.After(cutoff) {
			validAttempts = append(validAttempts, attempt)
		}
	}

	window.attempts = validAttempts
	window.lastCleanup = now
}

// shouldBlockDevice determines if a device should be blocked
func (s *SecurityManager) shouldBlockDevice(tracker *DeviceAttemptTracker) bool {
	// Block if too many failed attempts in recent period
	recentFailed := 0
	cutoff := time.Now().Add(-1 * time.Hour)

	for _, attempt := range tracker.attempts {
		if attempt.timestamp.After(cutoff) && !attempt.success {
			recentFailed++
		}
	}

	return recentFailed >= s.maxAttemptsPerDevice
}

// updateRiskScore updates the risk score for suspicious activity
func (s *SecurityManager) updateRiskScore(suspicious *SuspiciousActivity, record AttemptRecord) {
	// Increase risk for failed attempts
	if !record.success {
		suspicious.riskScore += 10
	}

	// Increase risk for multiple devices from same IP
	if len(suspicious.uniqueDevices) > 5 {
		suspicious.riskScore += 50
	}

	// Increase risk for trying many different license keys
	if len(suspicious.uniqueLicenses) > 10 {
		suspicious.riskScore += 100
	}

	// Increase risk for certain error types
	switch record.errorType {
	case "invalid_format":
		suspicious.riskScore += 5
	case "not_found":
		suspicious.riskScore += 15
	case "already_activated":
		suspicious.riskScore += 20
	}
}

// calculateInitialRiskScore calculates initial risk score for new IP
func (s *SecurityManager) calculateInitialRiskScore(record AttemptRecord) int {
	score := 0

	if !record.success {
		score += 10
	}

	// Check if IP looks suspicious
	if s.isIPSuspicious(record.clientIP) {
		score += 50
	}

	return score
}

// isIPSuspicious performs basic IP reputation checks
func (s *SecurityManager) isIPSuspicious(ip string) bool {
	// Parse IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // Invalid IP is suspicious
	}

	// Check for private/local addresses (could be proxy/VPN)
	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return false // Local IPs are generally safe
	}

	// Add more sophisticated checks here (GeoIP, known bot networks, etc.)
	return false
}

// isHoneypotLicense checks if a license key is a honeypot
func (s *SecurityManager) isHoneypotLicense(licenseKey string) bool {
	_, exists := s.honeypotLicenses[licenseKey]
	return exists
}

// handleHoneypotDetection handles detection of honeypot license usage
func (s *SecurityManager) handleHoneypotDetection(clientIP, deviceFingerprint, licenseKey string, logger *slog.Logger, ctx context.Context) {
	// Immediately block IP and device
	if suspicious, exists := s.suspiciousIPs[clientIP]; exists {
		suspicious.blocked = true
		suspicious.permanent = true
		suspicious.riskScore = s.permanentBanThreshold
	} else {
		s.suspiciousIPs[clientIP] = &SuspiciousActivity{
			ip:        clientIP,
			blocked:   true,
			permanent: true,
			riskScore: s.permanentBanThreshold,
			firstSeen: time.Now(),
		}
	}

	// Block device
	if tracker, exists := s.deviceAttempts[deviceFingerprint]; exists {
		tracker.blocked = true
		tracker.blockExpiry = time.Now().Add(365 * 24 * time.Hour) // Block for 1 year
	}

	// Log critical security event
	logger.ErrorContext(ctx, "Honeypot license detected - automated attack suspected",
		slog.String("action", "honeypot_detection"),
		slog.String("client_ip", clientIP),
		slog.String("device_fingerprint", deviceFingerprint),
		slog.String("honeypot_license", licenseKey),
		slog.String("security_response", "permanent_ban"),
	)
}

// detectSuspiciousPatterns detects suspicious patterns in requests
func (s *SecurityManager) detectSuspiciousPatterns(clientIP, deviceFingerprint, licenseKey, userAgent string) {
	// Check license key patterns
	for _, pattern := range s.suspiciousPatterns {
		if pattern.MatchString(licenseKey) {
			if suspicious, exists := s.suspiciousIPs[clientIP]; exists {
				suspicious.riskScore += 25
				suspicious.patterns = append(suspicious.patterns, pattern.String())
			}
		}
	}

	// Check user agent patterns
	if strings.Contains(strings.ToLower(userAgent), "bot") ||
		strings.Contains(strings.ToLower(userAgent), "crawler") ||
		strings.Contains(strings.ToLower(userAgent), "script") {
		if suspicious, exists := s.suspiciousIPs[clientIP]; exists {
			suspicious.riskScore += 30
		}
	}
}

// initializeSuspiciousPatterns sets up regex patterns for suspicious license keys
func (s *SecurityManager) initializeSuspiciousPatterns() {
	patterns := []string{
		`^(test|demo|sample|example)`, // Test/demo keys
		`^(admin|root|system)`,        // Administrative keys
		`^[0-9]{10,}$`,                // All numbers
		`^[a-zA-Z]{1,3}$`,             // Very short keys
		`^(..)\\1+$`,                  // Repeated patterns
		`^(ISX-0000-0000-0000|ISX-1111-1111-1111|ISX-XXXX-XXXX-XXXX)$`, // Common test patterns
	}

	s.suspiciousPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			s.suspiciousPatterns = append(s.suspiciousPatterns, compiled)
		}
	}
}

// initializeHoneypotLicenses creates honeypot license keys
func (s *SecurityManager) initializeHoneypotLicenses() {
	honeypots := []string{
		"ISX-TRAP-TRAP-TRAP",
		"ISX-FAKE-FAKE-FAKE",
		"ISX-TEST-TEST-TEST",
		"ISX-DEMO-DEMO-DEMO",
		"ISX-0000-0000-0000",
		"ISX-1111-1111-1111",
		"ISX-AAAA-AAAA-AAAA",
		"ISX-ZZZZ-ZZZZ-ZZZZ",
	}

	now := time.Now()
	for _, honeypot := range honeypots {
		s.honeypotLicenses[honeypot] = now
	}
}

// Security event logging structure
type SecurityEventType struct {
	Type              string
	Identifier        string
	DeviceFingerprint string
	ClientIP          string
	LicenseKey        string
	UserAgent         string
	Success           bool
	ErrorType         string
	Timestamp         time.Time
}

// logSecurityEvent logs structured security events
func (s *SecurityManager) logSecurityEvent(event SecurityEventType, logger *slog.Logger, ctx context.Context) {
	// Sanitize license key for logging (only show prefix)
	licensePrefix := "unknown"
	if len(event.LicenseKey) >= 8 {
		licensePrefix = event.LicenseKey[:8]
	}

	// Determine log level
	level := slog.LevelInfo
	if !event.Success {
		level = slog.LevelWarn
	}

	logger.Log(ctx, level, "Security event recorded",
		slog.String("event_type", event.Type),
		slog.String("identifier", event.Identifier),
		slog.String("device_fingerprint", event.DeviceFingerprint),
		slog.String("client_ip", event.ClientIP),
		slog.String("license_key_prefix", licensePrefix),
		slog.String("user_agent", event.UserAgent),
		slog.Bool("success", event.Success),
		slog.String("error_type", event.ErrorType),
		slog.Time("timestamp", event.Timestamp),
	)
}

// Enhanced cleanup methods
func (s *SecurityManager) cleanupSlidingWindows(now time.Time) {
	for ip, window := range s.slidingWindows {
		s.cleanupWindow(window, now)
		// Remove empty windows
		if len(window.attempts) == 0 && now.Sub(window.lastCleanup) > 24*time.Hour {
			delete(s.slidingWindows, ip)
		}
	}
}

func (s *SecurityManager) cleanupDeviceTrackers(now time.Time) {
	for fingerprint, tracker := range s.deviceAttempts {
		// Remove old attempts
		validAttempts := make([]AttemptRecord, 0, len(tracker.attempts))
		cutoff := now.Add(-24 * time.Hour)

		for _, attempt := range tracker.attempts {
			if attempt.timestamp.After(cutoff) {
				validAttempts = append(validAttempts, attempt)
			}
		}

		tracker.attempts = validAttempts

		// Remove inactive trackers
		if len(tracker.attempts) == 0 && now.Sub(tracker.lastAttempt) > 7*24*time.Hour {
			delete(s.deviceAttempts, fingerprint)
		}
	}
}

func (s *SecurityManager) cleanupSuspiciousActivity(now time.Time) {
	for ip, suspicious := range s.suspiciousIPs {
		// Remove old attempts
		validAttempts := make([]AttemptRecord, 0, len(suspicious.attempts))
		cutoff := now.Add(-7 * 24 * time.Hour)

		for _, attempt := range suspicious.attempts {
			if attempt.timestamp.After(cutoff) {
				validAttempts = append(validAttempts, attempt)
			}
		}

		suspicious.attempts = validAttempts

		// Remove inactive, non-blocked entries
		if !suspicious.blocked && len(suspicious.attempts) == 0 && now.Sub(suspicious.lastActivity) > 30*24*time.Hour {
			delete(s.suspiciousIPs, ip)
		}
	}
}

func (s *SecurityManager) cleanupHoneypotLicenses(now time.Time) {
	// Honeypot licenses don't need cleanup, they're permanent
	// But we could add rotation logic here if needed
}
