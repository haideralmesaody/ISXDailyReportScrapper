package security

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	logger                 *slog.Logger
	maxLicenseKeyLength    int
	maxUsernameLength      int
	maxEmailLength         int
	maxUserAgentLength     int
	maxIPAddressLength     int
	enableStrictValidation bool
	suspiciousPatterns     []*regexp.Regexp
	sqlInjectionPatterns   []*regexp.Regexp
	xssPatterns           []*regexp.Regexp
	commandInjectionPatterns []*regexp.Regexp
}

// ValidationConfig holds configuration for input validation
type ValidationConfig struct {
	MaxLicenseKeyLength    int  `json:"max_license_key_length"`
	MaxUsernameLength      int  `json:"max_username_length"`
	MaxEmailLength         int  `json:"max_email_length"`
	MaxUserAgentLength     int  `json:"max_user_agent_length"`
	MaxIPAddressLength     int  `json:"max_ip_address_length"`
	EnableStrictValidation bool `json:"enable_strict_validation"`
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	SanitizedValue string `json:"sanitized_value"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	RiskScore    int      `json:"risk_score"`
	InputType    string   `json:"input_type"`
	ThreatTypes  []string `json:"threat_types"`
}

// ThreatType represents different types of security threats
type ThreatType string

const (
	ThreatSQLInjection      ThreatType = "sql_injection"
	ThreatXSS              ThreatType = "xss"
	ThreatCommandInjection ThreatType = "command_injection"
	ThreatPathTraversal    ThreatType = "path_traversal"
	ThreatCSRF             ThreatType = "csrf"
	ThreatLDAPInjection    ThreatType = "ldap_injection"
	ThreatXMLInjection     ThreatType = "xml_injection"
	ThreatEmailInjection   ThreatType = "email_injection"
	ThreatSuspiciousPattern ThreatType = "suspicious_pattern"
	ThreatMalformedInput   ThreatType = "malformed_input"
)

// NewInputValidator creates a new input validator with security patterns
func NewInputValidator(config *ValidationConfig) *InputValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	validator := &InputValidator{
		logger:                 slog.Default(),
		maxLicenseKeyLength:    config.MaxLicenseKeyLength,
		maxUsernameLength:      config.MaxUsernameLength,
		maxEmailLength:         config.MaxEmailLength,
		maxUserAgentLength:     config.MaxUserAgentLength,
		maxIPAddressLength:     config.MaxIPAddressLength,
		enableStrictValidation: config.EnableStrictValidation,
	}

	// Initialize security patterns
	validator.initializeSecurityPatterns()

	return validator
}

// DefaultValidationConfig returns secure default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxLicenseKeyLength:    256,  // Maximum license key length
		MaxUsernameLength:      64,   // Maximum username length
		MaxEmailLength:         254,  // RFC 5321 maximum email length
		MaxUserAgentLength:     1024, // Maximum user agent length
		MaxIPAddressLength:     45,   // Maximum IP address length (IPv6)
		EnableStrictValidation: true,
	}
}

// SetLogger sets a custom logger for the validator
func (v *InputValidator) SetLogger(logger *slog.Logger) {
	v.logger = logger
}

// ValidateLicenseKey validates and sanitizes license key input
func (v *InputValidator) ValidateLicenseKey(ctx context.Context, licenseKey string) *ValidationResult {
	result := &ValidationResult{
		InputType:   "license_key",
		Errors:      []string{},
		Warnings:    []string{},
		ThreatTypes: []string{},
	}

	// Basic length check
	if len(licenseKey) == 0 {
		result.Errors = append(result.Errors, "license key cannot be empty")
		return result
	}

	if len(licenseKey) > v.maxLicenseKeyLength {
		result.Errors = append(result.Errors, fmt.Sprintf("license key exceeds maximum length of %d characters", v.maxLicenseKeyLength))
		return result
	}

	// Normalize the input
	sanitized := v.normalizeLicenseKey(licenseKey)
	result.SanitizedValue = sanitized

	// Validate license key format
	if err := v.validateLicenseKeyFormat(sanitized); err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.RiskScore += 20
	}

	// Check for suspicious patterns (but be lenient for license keys)
	// Accept both formats: ISX-XXXX-XXXX-XXXX-XXXX (23 chars) or standard format
	isValidScratchCard := strings.HasPrefix(sanitized, "ISX-") && len(sanitized) == 23
	isValidStandardFormat := strings.HasPrefix(sanitized, "ISX") && !strings.Contains(sanitized, "-") && len(sanitized) >= 9
	
	if isValidScratchCard || isValidStandardFormat {
		// Minimal threat detection for properly formatted license keys
		if !utf8.ValidString(sanitized) {
			threats := []string{string(ThreatMalformedInput)}
			result.ThreatTypes = threats
			result.RiskScore += len(threats) * 15
		}
	} else {
		threats := v.detectThreats(sanitized)
		result.ThreatTypes = threats
		result.RiskScore += len(threats) * 15
	}

	// Check for SQL injection patterns specifically (but skip for valid license format)
	if isValidScratchCard || isValidStandardFormat {
		// Skip SQL injection check for properly formatted license keys
	} else if v.containsSQLInjection(sanitized) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatSQLInjection))
		result.Errors = append(result.Errors, "potential SQL injection detected")
		result.RiskScore += 50
	}

	// Check for XSS patterns (but skip for valid license format)
	if isValidScratchCard || isValidStandardFormat {
		// Skip XSS check for properly formatted license keys
	} else if v.containsXSS(sanitized) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatXSS))
		result.Errors = append(result.Errors, "potential XSS attack detected")
		result.RiskScore += 40
	}

	// Check for command injection (but skip for valid license format)
	if isValidScratchCard || isValidStandardFormat {
		// Skip command injection check for properly formatted license keys
	} else if v.containsCommandInjection(sanitized) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatCommandInjection))
		result.Errors = append(result.Errors, "potential command injection detected")
		result.RiskScore += 60
	}

	// Log suspicious activity
	if result.RiskScore > 30 {
		v.logSuspiciousInput(ctx, "license_key", licenseKey, result)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// ValidateEmail validates and sanitizes email input
func (v *InputValidator) ValidateEmail(ctx context.Context, email string) *ValidationResult {
	result := &ValidationResult{
		InputType:   "email",
		Errors:      []string{},
		Warnings:    []string{},
		ThreatTypes: []string{},
	}

	// Basic checks
	if len(email) == 0 {
		result.Errors = append(result.Errors, "email cannot be empty")
		return result
	}

	if len(email) > v.maxEmailLength {
		result.Errors = append(result.Errors, fmt.Sprintf("email exceeds maximum length of %d characters", v.maxEmailLength))
		return result
	}

	// Sanitize email
	sanitized := v.sanitizeEmail(email)
	result.SanitizedValue = sanitized

	// Validate email format
	if !v.isValidEmailFormat(sanitized) {
		result.Errors = append(result.Errors, "invalid email format")
		result.RiskScore += 15
	}

	// Check for email injection patterns
	if v.containsEmailInjection(sanitized) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatEmailInjection))
		result.Errors = append(result.Errors, "potential email injection detected")
		result.RiskScore += 45
	}

	// Check for other threats
	threats := v.detectThreats(sanitized)
	result.ThreatTypes = append(result.ThreatTypes, threats...)
	result.RiskScore += len(threats) * 10

	if result.RiskScore > 20 {
		v.logSuspiciousInput(ctx, "email", email, result)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// ValidateUserAgent validates and sanitizes user agent input
func (v *InputValidator) ValidateUserAgent(ctx context.Context, userAgent string) *ValidationResult {
	result := &ValidationResult{
		InputType:   "user_agent",
		Errors:      []string{},
		Warnings:    []string{},
		ThreatTypes: []string{},
	}

	// Allow empty user agent
	if len(userAgent) == 0 {
		result.SanitizedValue = "Unknown"
		result.IsValid = true
		return result
	}

	if len(userAgent) > v.maxUserAgentLength {
		result.Errors = append(result.Errors, fmt.Sprintf("user agent exceeds maximum length of %d characters", v.maxUserAgentLength))
		return result
	}

	// Sanitize user agent
	sanitized := v.sanitizeUserAgent(userAgent)
	result.SanitizedValue = sanitized

	// Check for suspicious patterns
	threats := v.detectThreats(sanitized)
	result.ThreatTypes = threats
	result.RiskScore = len(threats) * 10

	// Check for automated tools
	if v.isAutomatedUserAgent(sanitized) {
		result.Warnings = append(result.Warnings, "automated tool detected")
		result.RiskScore += 20
	}

	// Check for malicious patterns
	if v.containsMaliciousUserAgent(sanitized) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatSuspiciousPattern))
		result.Warnings = append(result.Warnings, "potentially malicious user agent")
		result.RiskScore += 30
	}

	if result.RiskScore > 25 {
		v.logSuspiciousInput(ctx, "user_agent", userAgent, result)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// ValidateIPAddress validates and sanitizes IP address input
func (v *InputValidator) ValidateIPAddress(ctx context.Context, ipAddress string) *ValidationResult {
	result := &ValidationResult{
		InputType:   "ip_address",
		Errors:      []string{},
		Warnings:    []string{},
		ThreatTypes: []string{},
	}

	if len(ipAddress) == 0 {
		result.Errors = append(result.Errors, "IP address cannot be empty")
		return result
	}

	if len(ipAddress) > v.maxIPAddressLength {
		result.Errors = append(result.Errors, fmt.Sprintf("IP address exceeds maximum length of %d characters", v.maxIPAddressLength))
		return result
	}

	// Sanitize IP address
	sanitized := v.sanitizeIPAddress(ipAddress)
	result.SanitizedValue = sanitized

	// Validate IP format
	if !v.isValidIPFormat(sanitized) {
		result.Errors = append(result.Errors, "invalid IP address format")
		result.RiskScore += 25
	}

	// Check IP reputation
	reputation := v.checkIPReputation(sanitized)
	result.RiskScore += reputation.riskScore
	result.Warnings = append(result.Warnings, reputation.warnings...)

	if result.RiskScore > 30 {
		v.logSuspiciousInput(ctx, "ip_address", ipAddress, result)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// ValidateGenericInput validates generic string input with threat detection
func (v *InputValidator) ValidateGenericInput(ctx context.Context, input, inputType string, maxLength int) *ValidationResult {
	result := &ValidationResult{
		InputType:   inputType,
		Errors:      []string{},
		Warnings:    []string{},
		ThreatTypes: []string{},
	}

	if len(input) > maxLength {
		result.Errors = append(result.Errors, fmt.Sprintf("input exceeds maximum length of %d characters", maxLength))
		return result
	}

	// Basic sanitization
	sanitized := v.sanitizeGenericInput(input)
	result.SanitizedValue = sanitized

	// Threat detection (check original input for dangerous patterns)
	if v.containsSQLInjection(input) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatSQLInjection))
		result.RiskScore += 50
	}

	if v.containsXSS(input) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatXSS))
		result.RiskScore += 40
	}

	if v.containsCommandInjection(input) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatCommandInjection))
		result.RiskScore += 60
	}

	if v.containsPathTraversal(input) {
		result.ThreatTypes = append(result.ThreatTypes, string(ThreatPathTraversal))
		result.RiskScore += 35
	}

	if result.RiskScore > 20 {
		v.logSuspiciousInput(ctx, inputType, input, result)
	}

	result.IsValid = len(result.Errors) == 0
	return result
}

// normalizeLicenseKey normalizes license key format
func (v *InputValidator) normalizeLicenseKey(licenseKey string) string {
	// Remove all whitespace and convert to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(licenseKey))
	
	// For validation, keep the dashes but clean up whitespace
	return normalized
}

// validateLicenseKeyFormat validates the format of a license key
func (v *InputValidator) validateLicenseKeyFormat(licenseKey string) error {
	// Must start with ISX
	if !strings.HasPrefix(licenseKey, "ISX") {
		return fmt.Errorf("license key must start with 'ISX'")
	}

	// Check for scratch card format: ISX-XXXX-XXXX-XXXX-XXXX (23 characters with dashes)
	if strings.Contains(licenseKey, "-") {
		// Expected scratch card format
		if len(licenseKey) != 23 {
			return fmt.Errorf("scratch card license key must be in format ISX-XXXX-XXXX-XXXX-XXXX")
		}

		// Check dash positions for scratch card format
		if licenseKey[3] != '-' || licenseKey[8] != '-' || licenseKey[13] != '-' || licenseKey[18] != '-' {
			return fmt.Errorf("scratch card license key must be in format ISX-XXXX-XXXX-XXXX-XXXX")
		}

		// Check that segments contain only alphanumeric characters
		segments := []string{
			licenseKey[4:8],   // First segment after ISX-
			licenseKey[9:13],  // Second segment
			licenseKey[14:18], // Third segment
			licenseKey[19:23], // Fourth segment
		}

		for i, segment := range segments {
			for _, char := range segment {
				if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
					return fmt.Errorf("license key segment %d must contain only letters and numbers", i+1)
				}
			}
		}

		return nil
	}

	// Check for standard format without dashes (e.g., ISX1M02LYE1F9QJHR9D7Z)
	// Must have at least ISX + 1 char for duration + 5 chars minimum = 9 total
	if len(licenseKey) < 9 {
		return fmt.Errorf("license key too short")
	}

	// Check for standard duration prefixes
	validPrefixes := []string{"ISX1M", "ISX3M", "ISX6M", "ISX1Y"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(licenseKey, prefix) {
			hasValidPrefix = true
			// Validate remaining characters are alphanumeric
			for _, char := range licenseKey[5:] {
				if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
					return fmt.Errorf("license key must contain only letters and numbers after duration code")
				}
			}
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("invalid license key format. Expected ISX-XXXX-XXXX-XXXX-XXXX or ISX1M/3M/6M/1Y followed by alphanumeric code")
	}

	return nil
}

// sanitizeEmail sanitizes email input
func (v *InputValidator) sanitizeEmail(email string) string {
	// Trim whitespace and convert to lowercase
	sanitized := strings.ToLower(strings.TrimSpace(email))
	
	// Remove null bytes and control characters
	sanitized = v.removeControlCharacters(sanitized)
	
	return sanitized
}

// sanitizeUserAgent sanitizes user agent input
func (v *InputValidator) sanitizeUserAgent(userAgent string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(userAgent)
	
	// Remove null bytes and dangerous control characters
	sanitized = v.removeControlCharacters(sanitized)
	
	// HTML encode to prevent XSS
	sanitized = html.EscapeString(sanitized)
	
	return sanitized
}

// sanitizeIPAddress sanitizes IP address input
func (v *InputValidator) sanitizeIPAddress(ipAddress string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(ipAddress)
	
	// Remove brackets for IPv6
	sanitized = strings.Trim(sanitized, "[]")
	
	// Remove any control characters
	sanitized = v.removeControlCharacters(sanitized)
	
	return sanitized
}

// sanitizeGenericInput provides basic sanitization for generic input
func (v *InputValidator) sanitizeGenericInput(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)
	
	// Remove control characters
	sanitized = v.removeControlCharacters(sanitized)
	
	// HTML encode dangerous characters
	sanitized = html.EscapeString(sanitized)
	
	return sanitized
}

// removeControlCharacters removes null bytes and control characters
func (v *InputValidator) removeControlCharacters(input string) string {
	var result strings.Builder
	
	for _, r := range input {
		// Keep printable characters and common whitespace
		if unicode.IsPrint(r) || r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// isValidEmailFormat validates email format using regex
func (v *InputValidator) isValidEmailFormat(email string) bool {
	// Basic email regex (RFC 5322 compliant)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidIPFormat validates IP address format
func (v *InputValidator) isValidIPFormat(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isAutomatedUserAgent checks if user agent indicates automated tool
func (v *InputValidator) isAutomatedUserAgent(userAgent string) bool {
	lowerUA := strings.ToLower(userAgent)
	automatedPatterns := []string{
		"bot", "crawler", "spider", "scraper", "curl", "wget", "python", "java",
		"script", "automation", "monitor", "check", "test", "headless",
	}
	
	for _, pattern := range automatedPatterns {
		if strings.Contains(lowerUA, pattern) {
			return true
		}
	}
	
	return false
}

// containsMaliciousUserAgent checks for malicious user agent patterns
func (v *InputValidator) containsMaliciousUserAgent(userAgent string) bool {
	maliciousPatterns := []string{
		"<script", "javascript:", "eval(", "alert(", "document.cookie",
		"../../", "../", "cmd.exe", "/bin/sh", "wget", "curl",
	}
	
	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerUA, pattern) {
			return true
		}
	}
	
	return false
}

// IPReputation represents IP reputation information
type IPReputation struct {
	riskScore int
	warnings  []string
}

// checkIPReputation performs basic IP reputation checks
func (v *InputValidator) checkIPReputation(ip string) IPReputation {
	reputation := IPReputation{
		riskScore: 0,
		warnings:  []string{},
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		reputation.riskScore += 50
		reputation.warnings = append(reputation.warnings, "invalid IP format")
		return reputation
	}
	
	// Check for private/local addresses
	if parsedIP.IsPrivate() {
		reputation.warnings = append(reputation.warnings, "private IP address")
	}
	
	if parsedIP.IsLoopback() {
		reputation.warnings = append(reputation.warnings, "loopback IP address")
	}
	
	// Check for known bad IP ranges (simplified)
	if v.isKnownBadIPRange(parsedIP) {
		reputation.riskScore += 40
		reputation.warnings = append(reputation.warnings, "IP in known bad range")
	}
	
	return reputation
}

// isKnownBadIPRange checks if IP is in known bad ranges (simplified implementation)
func (v *InputValidator) isKnownBadIPRange(ip net.IP) bool {
	// This is a simplified implementation
	// In production, you would integrate with threat intelligence feeds
	
	// Check for bogon networks (simplified)
	bogonNetworks := []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"224.0.0.0/4",
		"240.0.0.0/4",
	}
	
	for _, network := range bogonNetworks {
		_, cidr, err := net.ParseCIDR(network)
		if err == nil && cidr.Contains(ip) {
			return true
		}
	}
	
	return false
}

// detectThreats detects various security threats in input
func (v *InputValidator) detectThreats(input string) []string {
	var threats []string
	
	// Check suspicious patterns
	for _, pattern := range v.suspiciousPatterns {
		if pattern.MatchString(input) {
			threats = append(threats, string(ThreatSuspiciousPattern))
			break
		}
	}
	
	// Check for malformed UTF-8
	if !utf8.ValidString(input) {
		threats = append(threats, string(ThreatMalformedInput))
	}
	
	// Check for path traversal
	if v.containsPathTraversal(input) {
		threats = append(threats, string(ThreatPathTraversal))
	}
	
	// Check for LDAP injection
	if v.containsLDAPInjection(input) {
		threats = append(threats, string(ThreatLDAPInjection))
	}
	
	// Check for XML injection
	if v.containsXMLInjection(input) {
		threats = append(threats, string(ThreatXMLInjection))
	}
	
	return threats
}

// containsSQLInjection checks for SQL injection patterns  
func (v *InputValidator) containsSQLInjection(input string) bool {
	// Check both original and sanitized input since patterns might be in either
	inputs := []string{input, strings.ToLower(input)}
	
	for _, testInput := range inputs {
		for _, pattern := range v.sqlInjectionPatterns {
			if pattern.MatchString(testInput) {
				return true
			}
		}
	}
	return false
}

// containsXSS checks for XSS patterns (before and after HTML encoding)
func (v *InputValidator) containsXSS(input string) bool {
	// Check original input
	for _, pattern := range v.xssPatterns {
		if pattern.MatchString(strings.ToLower(input)) {
			return true
		}
	}
	
	// Also check if script tags are present even after encoding
	if strings.Contains(strings.ToLower(input), "script") && 
	   (strings.Contains(input, "&lt;") || strings.Contains(input, "<")) {
		return true
	}
	
	return false
}

// containsCommandInjection checks for command injection patterns
func (v *InputValidator) containsCommandInjection(input string) bool {
	for _, pattern := range v.commandInjectionPatterns {
		if pattern.MatchString(strings.ToLower(input)) {
			return true
		}
	}
	return false
}

// containsEmailInjection checks for email header injection
func (v *InputValidator) containsEmailInjection(input string) bool {
	injectionPatterns := []string{
		"\r", "\n", "\\r", "\\n", "%0a", "%0d", "%0A", "%0D",
		"content-type:", "bcc:", "cc:", "to:", "from:",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}

// containsPathTraversal checks for path traversal patterns
func (v *InputValidator) containsPathTraversal(input string) bool {
	pathTraversalPatterns := []string{
		"../", "..\\", "..%2f", "..%5c", "%2e%2e%2f", "%2e%2e%5c",
		".%2e%2f", ".%2e%5c", "..%252f", "..%255c",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range pathTraversalPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}

// containsLDAPInjection checks for LDAP injection patterns
func (v *InputValidator) containsLDAPInjection(input string) bool {
	ldapPatterns := []string{
		"*)(", "*)|(", "*)(objectclass=*",
		"*)(&", "*))%00", "*)(|(objectclass=*",
	}
	
	for _, pattern := range ldapPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	
	return false
}

// containsXMLInjection checks for XML injection patterns
func (v *InputValidator) containsXMLInjection(input string) bool {
	xmlPatterns := []string{
		"<!entity", "<!doctype", "<!element", "<![cdata[",
		"&lt;!entity", "&lt;!doctype", "&#x", "&#",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range xmlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}

// initializeSecurityPatterns initializes regex patterns for threat detection
func (v *InputValidator) initializeSecurityPatterns() {
	// Suspicious patterns (very specific to avoid false positives)
	suspiciousPatterns := []string{
		`(?i)(union\\s+(all\\s+)?select)`,
		`(?i)(<script[^>]*>)`,
		`(?i)(javascript:|eval\\s*\\()`,
		`(?i)(cmd\\.exe|powershell\\.exe)`,
		`(?i)(\\.\\.[\\/\\\\])`,  // Path traversal
	}
	
	v.suspiciousPatterns = make([]*regexp.Regexp, 0, len(suspiciousPatterns))
	for _, pattern := range suspiciousPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			v.suspiciousPatterns = append(v.suspiciousPatterns, compiled)
		}
	}
	
	// SQL injection patterns (specific and targeted)
	sqlPatterns := []string{
		`(?i)('\\s*;\\s*(drop|delete|insert|update|create|alter))`,
		`(?i)(union\\s+(all\\s+)?select)`,
		`(?i)(--\\s*(drop|delete))`,  // Comment-based injections
		`(?i)(or\\s+1\\s*=\\s*1|and\\s+1\\s*=\\s*1)`,
		`(?i)(exec\\s*\\(|sp_\\w+|xp_\\w+)`,
		`(?i)(';.*drop.*table)`,  // More generic drop table pattern
	}
	
	v.sqlInjectionPatterns = make([]*regexp.Regexp, 0, len(sqlPatterns))
	for _, pattern := range sqlPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			v.sqlInjectionPatterns = append(v.sqlInjectionPatterns, compiled)
		}
	}
	
	// XSS patterns
	xssPatterns := []string{
		`(?i)(<script[^>]*>|</script>)`,
		`(?i)(javascript:|vbscript:|data:)`,
		`(?i)(on\\w+\\s*=|style\\s*=.*expression)`,
		`(?i)(<iframe|<object|<embed|<applet)`,
		`(?i)(eval\\s*\\(|alert\\s*\\(|confirm\\s*\\(|prompt\\s*\\()`,
	}
	
	v.xssPatterns = make([]*regexp.Regexp, 0, len(xssPatterns))
	for _, pattern := range xssPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			v.xssPatterns = append(v.xssPatterns, compiled)
		}
	}
	
	// Command injection patterns
	cmdPatterns := []string{
		`(?i)(;|\\|\\||&&|\\||\\$\\(|\\`+"`"+`|<\\(|>\\()`,
		`(?i)(cmd\\.exe|powershell|bash|sh|zsh|fish)`,
		`(?i)(wget|curl|nc|netcat|telnet|ssh)`,
		`(?i)(cat|ls|dir|type|echo|ping)`,
		`(?i)(>\\s*\\/|>\\s*[a-z]:|<\\s*\\/|<\\s*[a-z]:)`,
	}
	
	v.commandInjectionPatterns = make([]*regexp.Regexp, 0, len(cmdPatterns))
	for _, pattern := range cmdPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			v.commandInjectionPatterns = append(v.commandInjectionPatterns, compiled)
		}
	}
}

// logSuspiciousInput logs suspicious input for security monitoring
func (v *InputValidator) logSuspiciousInput(ctx context.Context, inputType, originalInput string, result *ValidationResult) {
	// Truncate long inputs for logging
	truncatedInput := originalInput
	if len(truncatedInput) > 100 {
		truncatedInput = truncatedInput[:100] + "..."
	}
	
	v.logger.WarnContext(ctx, "Suspicious input detected",
		slog.String("input_type", inputType),
		slog.String("original_input", truncatedInput),
		slog.String("sanitized_input", result.SanitizedValue),
		slog.Int("risk_score", result.RiskScore),
		slog.Any("threat_types", result.ThreatTypes),
		slog.Any("errors", result.Errors),
		slog.Any("warnings", result.Warnings),
		slog.Time("timestamp", time.Now()),
	)
}

// GetValidationMetrics returns validation metrics for monitoring
func (v *InputValidator) GetValidationMetrics() map[string]interface{} {
	return map[string]interface{}{
		"max_license_key_length":    v.maxLicenseKeyLength,
		"max_username_length":       v.maxUsernameLength,
		"max_email_length":          v.maxEmailLength,
		"max_user_agent_length":     v.maxUserAgentLength,
		"max_ip_address_length":     v.maxIPAddressLength,
		"strict_validation_enabled": v.enableStrictValidation,
		"suspicious_patterns_count": len(v.suspiciousPatterns),
		"sql_patterns_count":        len(v.sqlInjectionPatterns),
		"xss_patterns_count":        len(v.xssPatterns),
		"cmd_patterns_count":        len(v.commandInjectionPatterns),
	}
}