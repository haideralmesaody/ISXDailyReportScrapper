package license

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Logging Integration Tests
// =============================================================================

// LoggingTestSuite tests logging functionality in license manager
type LoggingTestSuite struct {
	suite.Suite
	tempDir     string
	licenseFile string
	manager     *Manager
}

func (suite *LoggingTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
	suite.licenseFile = filepath.Join(suite.tempDir, "test_license.dat")

	var err error
	suite.manager, err = NewManagerWithPath(suite.licenseFile)
	require.NoError(suite.T(), err)
}

func (suite *LoggingTestSuite) TearDownTest() {
	if suite.manager != nil {
		suite.manager.Close()
	}
}

// TestLogInfo tests info level logging
func (suite *LoggingTestSuite) TestLogInfo() {
	ctx := context.Background()

	// These should not panic and should work with default logger
	suite.manager.logInfo(ctx, "test_action", "Test info message",
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
		slog.Bool("key3", true),
	)

	// Test with various data types
	suite.manager.logInfo(ctx, "test_types", "Testing different types",
		slog.String("string_val", "test"),
		slog.Int("int_val", 123),
		slog.Int64("int64_val", 456789),
		slog.Float64("float_val", 3.14159),
		slog.Bool("bool_val", false),
		slog.Duration("duration_val", 5*time.Minute),
		slog.Time("time_val", time.Now()),
	)
}

// TestLogError tests error level logging
func (suite *LoggingTestSuite) TestLogError() {
	ctx := context.Background()

	// Test basic error logging
	suite.manager.logError(ctx, "test_error", "Test error message",
		slog.String("error", "something went wrong"),
		slog.String("component", "test"),
	)

	// Test error with stack trace context
	suite.manager.logError(ctx, "detailed_error", "Detailed error with context",
		slog.String("operation", "license_validation"),
		slog.String("license_key", "ISX1M***"),
		slog.String("error", "network timeout"),
		slog.Int("attempt", 3),
		slog.String("endpoint", "https://sheets.googleapis.com"),
	)
}

// TestLogWarn tests warning level logging
func (suite *LoggingTestSuite) TestLogWarn() {
	ctx := context.Background()

	suite.manager.logWarn(ctx, "test_warning", "Test warning message",
		slog.String("reason", "approaching limit"),
		slog.Int("current_value", 85),
		slog.Int("limit", 100),
	)
}

// TestLogDebug tests debug level logging
func (suite *LoggingTestSuite) TestLogDebug() {
	ctx := context.Background()

	suite.manager.logDebug(ctx, "test_debug", "Test debug message",
		slog.String("debug_info", "detailed internal state"),
		slog.String("cache_status", "hit"),
		slog.Duration("operation_time", 50*time.Millisecond),
	)
}

// TestLogLicenseAction tests license-specific action logging
func (suite *LoggingTestSuite) TestLogLicenseAction() {
	ctx := context.Background()

	// Test various license actions
	suite.manager.logLicenseAction(ctx, slog.LevelInfo, "license_activation", "License activated successfully",
		"ISX1M123", "test@example.com",
		slog.String("expiry_date", "2025-12-31"),
		slog.String("duration", "1m"),
		slog.Int("days_left", 365),
	)

	suite.manager.logLicenseAction(ctx, slog.LevelWarn, "license_expiring", "License expiring soon",
		"ISX1M456", "warn@example.com",
		slog.Int("days_left", 5),
		slog.String("status", "critical"),
	)

	suite.manager.logLicenseAction(ctx, slog.LevelError, "license_expired", "License has expired",
		"ISX1M789", "expired@example.com",
		slog.String("expired_date", "2024-01-01"),
		slog.Int("days_expired", 30),
	)
}

// TestLoggingWithNilContext tests logging with nil context
func (suite *LoggingTestSuite) TestLoggingWithNilContext() {
	// Should handle nil context gracefully
	suite.manager.logInfo(nil, "nil_context_test", "Testing with nil context")
	suite.manager.logError(nil, "nil_context_error", "Error with nil context")
	suite.manager.logWarn(nil, "nil_context_warn", "Warning with nil context")
	suite.manager.logDebug(nil, "nil_context_debug", "Debug with nil context")
}

// TestLoggingWithEmptyFields tests logging with no additional fields
func (suite *LoggingTestSuite) TestLoggingWithEmptyFields() {
	ctx := context.Background()

	// Test with no additional fields
	suite.manager.logInfo(ctx, "empty_fields", "Message with no additional fields")
	suite.manager.logError(ctx, "empty_error", "Error with no additional fields")
	suite.manager.logWarn(ctx, "empty_warn", "Warning with no additional fields")
	suite.manager.logDebug(ctx, "empty_debug", "Debug with no additional fields")
}

// TestLoggingWithSpecialCharacters tests logging with special characters
func (suite *LoggingTestSuite) TestLoggingWithSpecialCharacters() {
	ctx := context.Background()

	// Test with special characters in messages and fields
	suite.manager.logInfo(ctx, "special_chars", "Message with special chars: !@#$%^&*()",
		slog.String("field_with_special", "value!@#$%^&*()"),
		slog.String("unicode", "测试🔒日本語"),
		slog.String("json_like", `{"key": "value", "nested": {"inner": true}}`),
		slog.String("multiline", "line1\nline2\nline3"),
	)
}

// TestLoggingWithLargeData tests logging with large data
func (suite *LoggingTestSuite) TestLoggingWithLargeData() {
	ctx := context.Background()

	// Test with large strings
	largeString := strings.Repeat("x", 10000)
	suite.manager.logInfo(ctx, "large_data", "Testing large data logging",
		slog.String("large_field", largeString),
		slog.String("normal_field", "normal_value"),
	)

	// Test with many fields
	fields := make([]slog.Attr, 50)
	for i := 0; i < 50; i++ {
		fields[i] = slog.String(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Convert to variadic call - use individual fields
	suite.manager.logInfo(ctx, "many_fields", "Testing many fields",
		slog.String("field_0", "value_0"),
		slog.String("field_1", "value_1"),
		slog.String("field_2", "value_2"),
		slog.Int("total_fields", 50),
	)
}

// TestConcurrentLogging tests concurrent logging operations
func (suite *LoggingTestSuite) TestConcurrentLogging() {
	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent logging should be safe
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Mix different log levels
			suite.manager.logInfo(ctx, "concurrent_info", fmt.Sprintf("Info from goroutine %d", id),
				slog.Int("goroutine_id", id),
			)

			suite.manager.logWarn(ctx, "concurrent_warn", fmt.Sprintf("Warning from goroutine %d", id),
				slog.Int("goroutine_id", id),
			)

			suite.manager.logError(ctx, "concurrent_error", fmt.Sprintf("Error from goroutine %d", id),
				slog.Int("goroutine_id", id),
			)

			suite.manager.logDebug(ctx, "concurrent_debug", fmt.Sprintf("Debug from goroutine %d", id),
				slog.Int("goroutine_id", id),
			)

			// License-specific logging
			suite.manager.logLicenseAction(ctx, slog.LevelInfo, "concurrent_license",
				fmt.Sprintf("License action from goroutine %d", id),
				fmt.Sprintf("ISX1M%03d", id),
				fmt.Sprintf("user%d@test.com", id),
				slog.Int("goroutine_id", id),
			)
		}(i)
	}

	wg.Wait()
	// Should complete without deadlocks or panics
}

// TestLoggingPerformance tests logging performance
func (suite *LoggingTestSuite) TestLoggingPerformance() {
	ctx := context.Background()

	// Time multiple logging operations
	start := time.Now()

	for i := 0; i < 1000; i++ {
		suite.manager.logInfo(ctx, "performance_test", "Performance test message",
			slog.Int("iteration", i),
			slog.String("test_type", "performance"),
			slog.Duration("elapsed", time.Since(start)),
		)
	}

	elapsed := time.Since(start)
	suite.T().Logf("1000 log operations took %v (%.2f μs per operation)",
		elapsed, float64(elapsed.Nanoseconds())/1000.0/1000.0)

	// Logging should be reasonably fast
	suite.Less(elapsed, 5*time.Second, "Logging should be performant")
}

// TestLoggingWithContextValues tests logging with context values
func (suite *LoggingTestSuite) TestLoggingWithContextValues() {
	// Create context with values
	ctx := context.WithValue(context.Background(), "request_id", "req-12345")
	ctx = context.WithValue(ctx, "user_id", "user-67890")
	ctx = context.WithValue(ctx, "trace_id", "trace-abcdef")

	// Logging should work with context values (though our logger may not extract them)
	suite.manager.logInfo(ctx, "context_values", "Testing with context values",
		slog.String("additional_field", "extra_data"),
	)
}

// TestLoggingFieldTypes tests various field types
func (suite *LoggingTestSuite) TestLoggingFieldTypes() {
	ctx := context.Background()

	// Test all supported slog field types
	suite.manager.logInfo(ctx, "field_types", "Testing all field types",
		slog.String("string", "text_value"),
		slog.Int("int", 42),
		slog.Int64("int64", 9223372036854775807),
		slog.Uint64("uint64", 18446744073709551615),
		slog.Float64("float64", 3.14159265359),
		slog.Bool("bool_true", true),
		slog.Bool("bool_false", false),
		slog.Duration("duration", 2*time.Hour+30*time.Minute+45*time.Second),
		slog.Time("time", time.Date(2024, 8, 1, 12, 30, 45, 0, time.UTC)),
		slog.Group("group",
			slog.String("nested_string", "nested_value"),
			slog.Int("nested_int", 123),
		),
	)
}

// TestErrorLoggingEdgeCases tests edge cases in error logging
func (suite *LoggingTestSuite) TestErrorLoggingEdgeCases() {
	ctx := context.Background()

	// Test with nil error
	suite.manager.logError(ctx, "nil_error", "Testing with nil error",
		slog.String("error", ""),
	)

	// Test with empty error message
	suite.manager.logError(ctx, "empty_error", "",
		slog.String("component", "test"),
	)

	// Test with very long error message
	longError := strings.Repeat("very long error message ", 1000)
	suite.manager.logError(ctx, "long_error", longError,
		slog.String("component", "test"),
	)
}

// Run the logging test suite
func TestLoggingTestSuite(t *testing.T) {
	suite.Run(t, new(LoggingTestSuite))
}

// =============================================================================
// Unit Tests for Logging Helper Functions
// =============================================================================

func TestManagerGetWorkingDir(t *testing.T) {
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "test_license.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(t, err)
	defer manager.Close()

	workingDir := manager.getWorkingDir()
	assert.NotEmpty(t, workingDir)
	assert.NotEqual(t, "unknown", workingDir)

	// Should be a valid directory path
	_, err = os.Stat(workingDir)
	assert.NoError(t, err)
}

func TestIsWritable(t *testing.T) {
	tempDir := t.TempDir()

	// Test writable directory
	assert.True(t, isWritable(tempDir))

	// Test non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	assert.False(t, isWritable(nonExistentDir))

	// Skip read-only test on Windows (Unix permissions don't work the same)
	if runtime.GOOS != "windows" && os.Getenv("CI") != "true" {
		// Test read-only directory (skip in CI and on Windows)
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555) // Read-only
		require.NoError(t, err)

		assert.False(t, isWritable(readOnlyDir))

		// Cleanup
		os.Chmod(readOnlyDir, 0755)
		os.RemoveAll(readOnlyDir)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkLogInfo(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "bench_license.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.logInfo(ctx, "benchmark_test", "Benchmark log message",
			slog.Int("iteration", i),
			slog.String("test_type", "benchmark"),
		)
	}
}

func BenchmarkLogLicenseAction(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "bench_license.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.logLicenseAction(ctx, slog.LevelInfo, "benchmark_action", "Benchmark license action",
			"ISX1MBENCH123456", "bench@example.com",
			slog.Int("iteration", i),
		)
	}
}

func BenchmarkConcurrentLogging(b *testing.B) {
	tempDir := b.TempDir()
	licenseFile := filepath.Join(tempDir, "bench_license.dat")

	manager, err := NewManagerWithPath(licenseFile)
	require.NoError(b, err)
	defer manager.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			manager.logInfo(ctx, "concurrent_benchmark", "Concurrent benchmark log",
				slog.Int("iteration", i),
			)
			i++
		}
	})
}
