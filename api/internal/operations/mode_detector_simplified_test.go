package operations

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModeDetector_SimplifiedLogic tests the core simplified mode detection logic
// as documented in CLAUDE.md: if ANY files exist, use accumulative; otherwise initial
func TestModeDetector_SimplifiedLogic(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "mode_detector_simplified_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create detector and override paths for testing
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = tempDir

	tests := []struct {
		name           string
		existingFiles  []string
		fromDate       string
		toDate         string
		expectedMode   string
		expectedReason string
		description    string
	}{
		{
			name:           "no files - initial mode",
			existingFiles:  []string{},
			fromDate:       "2024-01-01",
			toDate:         "2024-01-05",
			expectedMode:   ModeInitial,
			expectedReason: "No existing files found - using initial mode for fresh download",
			description:    "When no files exist, should recommend initial mode",
		},
		{
			name: "1 file - accumulative mode",
			existingFiles: []string{
				"2024 01 01 ISX Daily Report.xlsx",
			},
			fromDate:       "2024-01-01",
			toDate:         "2024-01-05",
			expectedMode:   ModeAccumulative,
			expectedReason: "Found 1 existing files in requested range - using accumulative mode to add new data",
			description:    "With just 1 file existing, should recommend accumulative mode",
		},
		{
			name: "50 files - accumulative mode",
			existingFiles: func() []string {
				files := make([]string, 50)
				for i := 0; i < 50; i++ {
					files[i] = fmt.Sprintf("2024 01 %02d ISX Daily Report.xlsx", i+1)
				}
				return files
			}(),
			fromDate:       "2024-01-01",
			toDate:         "2024-03-31",
			expectedMode:   ModeAccumulative,
			expectedReason: "Found 31 existing files in requested range - using accumulative mode to add new data",
			description:    "With 50 files existing, should recommend accumulative mode",
		},
		{
			name: "200 files - accumulative mode",
			existingFiles: func() []string {
				files := make([]string, 200)
				for i := 0; i < 200; i++ {
					// Spread files across multiple months to avoid invalid dates
					month := (i / 30) + 1
					day := (i % 30) + 1
					if month > 12 {
						month = 12
						day = (i % 31) + 1
					}
					files[i] = fmt.Sprintf("2024 %02d %02d ISX Daily Report.xlsx", month, day)
				}
				return files
			}(),
			fromDate:       "2024-01-01",
			toDate:         "2024-12-31",
			expectedMode:   ModeAccumulative,
			expectedReason: "Found 199 existing files in requested range - using accumulative mode to add new data",
			description:    "With 200 files existing, should recommend accumulative mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp directory
			os.RemoveAll(tempDir)
			err := os.MkdirAll(tempDir, 0755)
			require.NoError(t, err)

			// Create test files
			for _, filename := range tt.existingFiles {
				filePath := filepath.Join(tempDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			// Run mode detection
			result, err := detector.DetectOptimalMode(ctx, tt.fromDate, tt.toDate)
			require.NoError(t, err, "DetectOptimalMode should not fail")
			require.NotNil(t, result, "Result should not be nil")

			// Validate simplified logic assertions
			assert.Equal(t, tt.expectedMode, result.RecommendedMode,
				"Mode should match expected: %s", tt.description)
			assert.Equal(t, tt.expectedReason, result.Reason,
				"Reason should match expected format")

			// NOTE: result.ExistingFiles shows files IN REQUESTED RANGE, which may be
			// less than len(tt.existingFiles) if some files fall outside the range
			assert.LessOrEqual(t, result.ExistingFiles, len(tt.existingFiles),
				"Existing files count should not exceed created files")
			assert.False(t, result.SkipRecommendation,
				"Simplified logic never recommends skipping")

			// Validate that coverage is calculated correctly
			if result.ExistingFiles > 0 {
				assert.Greater(t, result.CoveragePercent, 0.0,
					"Coverage should be > 0 when files exist in range")
			} else {
				assert.Equal(t, 0.0, result.CoveragePercent,
					"Coverage should be 0 when no files exist in range")
			}

			// Validate trading days calculation
			assert.Greater(t, result.TradingDays, 0,
				"Should have at least 1 trading day for valid date range")
			assert.Equal(t, result.TradingDays-result.ExistingFiles, result.MissingDays,
				"Missing days calculation should be correct")
		})
	}
}

// TestModeDetector_GetExistingFileDates_Focused tests file date extraction with table-driven approach
func TestModeDetector_GetExistingFileDates_Focused(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mode_detector_dates_focused_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = tempDir

	tests := []struct {
		name          string
		createFiles   []string
		expectedDates []string
		description   string
	}{
		{
			name:          "empty directory",
			createFiles:   []string{},
			expectedDates: []string{},
			description:   "Should return empty slice when no files exist",
		},
		{
			name: "valid ISX files only",
			createFiles: []string{
				"2024 01 15 ISX Daily Report.xlsx",
				"2024 02 20 ISX Daily Report.xlsx",
				"2023 12 25 ISX Daily Report.xls",
			},
			expectedDates: []string{
				"2024-01-15",
				"2024-02-20",
				"2023-12-25",
			},
			description: "Should extract dates from valid ISX report filenames",
		},
		{
			name: "mixed valid and invalid files",
			createFiles: []string{
				"2024 01 15 ISX Daily Report.xlsx",
				"invalid_filename.xlsx",
				"readme.txt",
				"2024 03 10 ISX Daily Report.xls",
				"not_a_date.xlsx",
				"subdir", // This will be created as a directory
			},
			expectedDates: []string{
				"2024-01-15",
				"2024-03-10",
			},
			description: "Should only extract dates from valid ISX files, ignoring others",
		},
		{
			name: "files with invalid date formats",
			createFiles: []string{
				"24 01 15 ISX Daily Report.xlsx",   // Year too short
				"2024 13 15 ISX Daily Report.xlsx", // Invalid month
				"2024 01 32 ISX Daily Report.xlsx", // Invalid day
				"2024 01 ISX Daily Report.xlsx",    // Missing day
				"not_a_date_file.xlsx",             // No date pattern
			},
			expectedDates: []string{},
			description:   "Should return empty when all files have invalid date formats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp directory
			os.RemoveAll(tempDir)
			err := os.MkdirAll(tempDir, 0755)
			require.NoError(t, err)

			// Create test files and directories
			for _, name := range tt.createFiles {
				path := filepath.Join(tempDir, name)
				if name == "subdir" {
					// Create a subdirectory
					err := os.MkdirAll(path, 0755)
					require.NoError(t, err)
				} else {
					// Create a file
					err := os.WriteFile(path, []byte("test content"), 0644)
					require.NoError(t, err)
				}
			}

			// Test the function
			dates, err := detector.GetExistingFileDates(ctx)
			require.NoError(t, err, "GetExistingFileDates should not fail")

			// Validate results
			assert.ElementsMatch(t, tt.expectedDates, dates,
				"Extracted dates should match expected: %s", tt.description)
		})
	}
}

// TestModeDetector_ValidateDateRange_Focused tests date range validation with comprehensive cases
func TestModeDetector_ValidateDateRange_Focused(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		fromDate      string
		toDate        string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "valid single day",
			fromDate:    "2024-01-01", // Monday
			toDate:      "2024-01-01",
			expectError: false,
			description: "Single trading day should be valid",
		},
		{
			name:        "valid week range",
			fromDate:    "2024-01-01", // Monday
			toDate:      "2024-01-07", // Sunday
			expectError: false,
			description: "Week range with multiple trading days should be valid",
		},
		{
			name:          "invalid from date format",
			fromDate:      "invalid-date",
			toDate:        "2024-01-01",
			expectError:   true,
			errorContains: "invalid from_date format",
			description:   "Should reject invalid from_date format",
		},
		{
			name:          "invalid to date format",
			fromDate:      "2024-01-01",
			toDate:        "not-a-date",
			expectError:   true,
			errorContains: "invalid to_date format",
			description:   "Should reject invalid to_date format",
		},
		{
			name:          "reversed date range",
			fromDate:      "2024-01-10",
			toDate:        "2024-01-01",
			expectError:   true,
			errorContains: "cannot be before",
			description:   "Should reject when to_date is before from_date",
		},
		{
			name:          "weekend only range",
			fromDate:      "2024-01-05", // Friday
			toDate:        "2024-01-06", // Saturday
			expectError:   true,
			errorContains: "contains no trading days",
			description:   "Should reject range with only weekends",
		},
		{
			name:        "large date range warning",
			fromDate:    "2024-01-01",
			toDate:      "2024-12-31", // Full year
			expectError: false,
			description: "Large range should pass but log warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := detector.ValidateDateRange(ctx, tt.fromDate, tt.toDate)

			if tt.expectError {
				assert.Error(t, err, "Should return error: %s", tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Should not return error: %s", tt.description)
			}
		})
	}
}

// TestModeDetector_EdgeCases_Focused tests edge cases and error conditions
func TestModeDetector_EdgeCases_Focused(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name          string
		setupFunc     func() *ModeDetector
		testFunc      func(*ModeDetector) error
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name: "nil detector - DetectOptimalMode",
			setupFunc: func() *ModeDetector {
				return nil
			},
			testFunc: func(detector *ModeDetector) error {
				_, err := detector.DetectOptimalMode(ctx, "2024-01-01", "2024-01-02")
				return err
			},
			expectError:   true,
			errorContains: "ModeDetector is nil",
			description:   "Should handle nil detector gracefully",
		},
		{
			name: "nil detector - GetExistingFileDates",
			setupFunc: func() *ModeDetector {
				return nil
			},
			testFunc: func(detector *ModeDetector) error {
				_, err := detector.GetExistingFileDates(ctx)
				return err
			},
			expectError:   true,
			errorContains: "ModeDetector is nil",
			description:   "Should handle nil detector gracefully",
		},
		{
			name: "nil detector - ValidateDateRange",
			setupFunc: func() *ModeDetector {
				return nil
			},
			testFunc: func(detector *ModeDetector) error {
				return detector.ValidateDateRange(ctx, "2024-01-01", "2024-01-02")
			},
			expectError:   true,
			errorContains: "ModeDetector is nil",
			description:   "Should handle nil detector gracefully",
		},
		{
			name: "nonexistent downloads directory",
			setupFunc: func() *ModeDetector {
				detector, _ := NewModeDetector(logger)
				detector.paths.DownloadsDir = "/nonexistent/directory/path"
				return detector
			},
			testFunc: func(detector *ModeDetector) error {
				dates, err := detector.GetExistingFileDates(ctx)
				if err != nil {
					return err
				}
				// Should return empty slice, not error
				if len(dates) != 0 {
					return fmt.Errorf("expected empty slice, got %v", dates)
				}
				return nil
			},
			expectError: false,
			description: "Should handle nonexistent directory gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := tt.setupFunc()
			err := tt.testFunc(detector)

			if tt.expectError {
				assert.Error(t, err, "Should return error: %s", tt.description)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Should not return error: %s", tt.description)
			}
		})
	}
}

// BenchmarkModeDetection_Simplified benchmarks the core mode detection logic
func BenchmarkModeDetection_Simplified(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "mode_detector_bench_simplified")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	detector, err := NewModeDetector(logger)
	require.NoError(b, err)
	detector.paths.DownloadsDir = tempDir

	// Create varying numbers of test files to benchmark different scenarios
	benchmarkCases := []struct {
		name      string
		fileCount int
	}{
		{"no_files", 0},
		{"few_files", 10},
		{"many_files", 100},
		{"large_dataset", 500},
	}

	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			// Setup: create test files
			for i := 1; i <= bc.fileCount; i++ {
				filename := fmt.Sprintf("2024 01 %02d ISX Daily Report.xlsx", (i%28)+1) // Cycle through valid days
				filePath := filepath.Join(tempDir, filename)
				os.WriteFile(filePath, []byte("test"), 0644)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := detector.DetectOptimalMode(ctx, "2024-01-01", "2024-01-31")
				if err != nil {
					b.Fatalf("DetectOptimalMode failed: %v", err)
				}
			}

			// Cleanup for next benchmark
			os.RemoveAll(tempDir)
			os.MkdirAll(tempDir, 0755)
		})
	}
}

// BenchmarkGetExistingFileDates benchmarks file date extraction performance
func BenchmarkGetExistingFileDates(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "file_dates_bench")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	detector, err := NewModeDetector(logger)
	require.NoError(b, err)
	detector.paths.DownloadsDir = tempDir

	// Create test files with valid and invalid names
	for i := 1; i <= 100; i++ {
		// Create mix of valid ISX files and other files
		if i%3 == 0 {
			filename := fmt.Sprintf("2024 %02d %02d ISX Daily Report.xlsx", (i%12)+1, (i%28)+1)
			filePath := filepath.Join(tempDir, filename)
			os.WriteFile(filePath, []byte("test"), 0644)
		} else {
			filename := fmt.Sprintf("other_file_%d.xlsx", i)
			filePath := filepath.Join(tempDir, filename)
			os.WriteFile(filePath, []byte("test"), 0644)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := detector.GetExistingFileDates(ctx)
		if err != nil {
			b.Fatalf("GetExistingFileDates failed: %v", err)
		}
	}
}

// BenchmarkValidateDateRange benchmarks date range validation
func BenchmarkValidateDateRange(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(b, err)

	testCases := []struct {
		name     string
		fromDate string
		toDate   string
	}{
		{"single_day", "2024-01-01", "2024-01-01"},
		{"week", "2024-01-01", "2024-01-07"},
		{"month", "2024-01-01", "2024-01-31"},
		{"year", "2024-01-01", "2024-12-31"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := detector.ValidateDateRange(ctx, tc.fromDate, tc.toDate)
				if err != nil {
					b.Fatalf("ValidateDateRange failed: %v", err)
				}
			}
		})
	}
}
