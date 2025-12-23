package license

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Simple OpenTelemetry Tests (to avoid complex setup issues)
// =============================================================================

func TestClassifyLicenseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "expired license",
			err:      errors.New("license expired on 2024-01-01"),
			expected: "license_expired",
		},
		{
			name:     "machine mismatch",
			err:      errors.New("machine_mismatch: license bound to different machine"),
			expected: "machine_mismatch",
		},
		{
			name:     "license not found",
			err:      errors.New("license not found in database"),
			expected: "license_not_found",
		},
		{
			name:     "invalid license",
			err:      errors.New("invalid license key format"),
			expected: "invalid_license",
		},
		{
			name:     "network error",
			err:      errors.New("network timeout while connecting"),
			expected: "network_error",
		},
		{
			name:     "timeout error",
			err:      errors.New("operation timeout"),
			expected: "network_error",
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded"),
			expected: "rate_limited",
		},
		{
			name:     "unauthorized error",
			err:      errors.New("unauthorized access"),
			expected: "unauthorized",
		},
		{
			name:     "unknown error",
			err:      errors.New("some random error"),
			expected: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyLicenseError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClassifyNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "timeout error",
			err:      errors.New("timeout connecting to server"),
			expected: "timeout",
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: "connection_refused",
		},
		{
			name:     "DNS error",
			err:      errors.New("no such host"),
			expected: "dns_error",
		},
		{
			name:     "403 forbidden",
			err:      errors.New("HTTP 403 Forbidden"),
			expected: "forbidden",
		},
		{
			name:     "401 unauthorized",
			err:      errors.New("HTTP 401 Unauthorized"),
			expected: "unauthorized",
		},
		{
			name:     "500 server error",
			err:      errors.New("HTTP 500 Internal Server Error"),
			expected: "server_error",
		},
		{
			name:     "generic network error",
			err:      errors.New("network is unreachable"),
			expected: "network_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyNetworkError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "timeout",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "substring at beginning",
			s:        "timeout error occurred",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "connection timeout",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "network timeout error",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "not found",
			s:        "connection refused",
			substr:   "timeout",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "any string",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "timeout",
			expected: false,
		},
		{
			name:     "both empty",
			s:        "",
			substr:   "",
			expected: true,
		},
		{
			name:     "case sensitive",
			s:        "TIMEOUT",
			substr:   "timeout",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsInner(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "found in middle",
			s:        "abcdefghijk",
			substr:   "cde",
			expected: true,
		},
		{
			name:     "found at start",
			s:        "abcdefghijk",
			substr:   "abc",
			expected: true,
		},
		{
			name:     "found at end",
			s:        "abcdefghijk",
			substr:   "ijk",
			expected: true,
		},
		{
			name:     "not found",
			s:        "abcdefghijk",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "abc",
			substr:   "abcdef",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "abc",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "abc",
			substr:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsInner(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkClassifyLicenseErrorSimple(b *testing.B) {
	err := errors.New("license expired on 2024-01-01")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = classifyLicenseError(err)
	}
}

func BenchmarkContainsFunctionSimple(b *testing.B) {
	s := "this is a long string with timeout error in the middle"
	substr := "timeout"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = contains(s, substr)
	}
}
