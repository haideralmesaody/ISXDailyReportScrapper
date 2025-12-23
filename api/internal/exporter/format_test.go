package exporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "zero value",
			input:    0.0,
			expected: "0.00", // Fixed 2 decimal format
		},
		{
			name:     "positive integer",
			input:    123.0,
			expected: "123.00", // Fixed 2 decimal format
		},
		{
			name:     "negative integer",
			input:    -456.0,
			expected: "-456.00", // Fixed 2 decimal format
		},
		{
			name:     "positive decimal with trailing zeros",
			input:    123.456000,
			expected: "123.46", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "negative decimal with trailing zeros",
			input:    -789.123000,
			expected: "-789.12", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "small positive decimal",
			input:    0.001234,
			expected: "0.00", // Fixed 2 decimal format (rounded down)
		},
		{
			name:     "small negative decimal",
			input:    -0.005678,
			expected: "-0.01", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "very small positive number",
			input:    0.000001,
			expected: "0.00", // Fixed 2 decimal format (rounded down)
		},
		{
			name:     "very small negative number",
			input:    -0.000001,
			expected: "-0.00", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "large positive number",
			input:    1234567.890123,
			expected: "1234567.89", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "large negative number",
			input:    -9876543.210987,
			expected: "-9876543.21", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "decimal ending in zero",
			input:    123.450000,
			expected: "123.45", // Fixed 2 decimal format
		},
		{
			name:     "all trailing zeros removed",
			input:    100.000000,
			expected: "100.00", // Fixed 2 decimal format
		},
		{
			name:     "six decimal places",
			input:    1.123456,
			expected: "1.12", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "more than six decimal places (should truncate)",
			input:    1.1234567890,
			expected: "1.12", // Fixed 2 decimal format (rounded)
		},
		{
			name:     "scientific notation input",
			input:    1.23e-5,
			expected: "0.00", // Fixed 2 decimal format (rounded down)
		},
		{
			name:     "negative scientific notation input",
			input:    -4.56e-4,
			expected: "-0.00", // Fixed 2 decimal format (rounded)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFloat(tt.input)
			assert.Equal(t, tt.expected, result, "formatFloat(%f) = %s, want %s", tt.input, result, tt.expected)
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "zero value",
			input:    0,
			expected: "0",
		},
		{
			name:     "positive small integer",
			input:    123,
			expected: "123",
		},
		{
			name:     "negative small integer",
			input:    -456,
			expected: "-456",
		},
		{
			name:     "positive large integer",
			input:    9223372036854775807, // max int64
			expected: "9223372036854775807",
		},
		{
			name:     "negative large integer",
			input:    -9223372036854775808, // min int64
			expected: "-9223372036854775808",
		},
		{
			name:     "typical volume value",
			input:    1000000,
			expected: "1000000",
		},
		{
			name:     "typical trade count",
			input:    42,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInt(tt.input)
			assert.Equal(t, tt.expected, result, "formatInt(%d) = %s, want %s", tt.input, result, tt.expected)
		})
	}
}

func TestFormatBool(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true value",
			input:    true,
			expected: "true",
		},
		{
			name:     "false value",
			input:    false,
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBool(tt.input)
			assert.Equal(t, tt.expected, result, "formatBool(%t) = %s, want %s", tt.input, result, tt.expected)
		})
	}
}

// BenchmarkFormatFloat tests the performance of formatFloat function
func BenchmarkFormatFloat(b *testing.B) {
	testValues := []float64{
		0.0,
		123.456789,
		-987.654321,
		1234567.890123,
		0.000001,
		999999.999999,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			_ = formatFloat(val)
		}
	}
}

// BenchmarkFormatInt tests the performance of formatInt function
func BenchmarkFormatInt(b *testing.B) {
	testValues := []int64{
		0,
		123456,
		-987654,
		9223372036854775807,
		-9223372036854775808,
		1000000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			_ = formatInt(val)
		}
	}
}

// BenchmarkFormatBool tests the performance of formatBool function
func BenchmarkFormatBool(b *testing.B) {
	testValues := []bool{true, false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			_ = formatBool(val)
		}
	}
}