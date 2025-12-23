package operations

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModeDetector(t *testing.T) {
	tests := []struct {
		name      string
		logger    *slog.Logger
		expectErr bool
	}{
		{
			name:      "valid logger",
			logger:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			expectErr: false,
		},
		{
			name:      "nil logger",
			logger:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector, err := NewModeDetector(tt.logger)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, detector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, detector)
				assert.NotNil(t, detector.fileDetector)
				assert.NotNil(t, detector.paths)
			}
		})
	}
}

func TestModeDetector_calculateTradingDays(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)

	tests := []struct {
		name         string
		fromDate     string
		toDate       string
		expectedDays int
	}{
		{
			name:         "single weekday",
			fromDate:     "2024-01-01", // Monday
			toDate:       "2024-01-01",
			expectedDays: 1,
		},
		{
			name:         "single friday (weekend in Iraq)",
			fromDate:     "2024-01-05", // Friday
			toDate:       "2024-01-05",
			expectedDays: 0,
		},
		{
			name:         "single saturday (weekend in Iraq)",
			fromDate:     "2024-01-06", // Saturday
			toDate:       "2024-01-06",
			expectedDays: 0,
		},
		{
			name:         "full week monday to sunday",
			fromDate:     "2024-01-01", // Monday
			toDate:       "2024-01-07", // Sunday
			expectedDays: 5,            // Mon, Tue, Wed, Thu, Sun (excluding Fri, Sat)
		},
		{
			name:         "two full weeks",
			fromDate:     "2024-01-01", // Monday
			toDate:       "2024-01-14", // Sunday
			expectedDays: 10,           // 5 days per week * 2 weeks
		},
		{
			name:         "entire january 2024",
			fromDate:     "2024-01-01",
			toDate:       "2024-01-31",
			expectedDays: 23, // 31 days - 8 weekend days (4 Fridays + 4 Saturdays)
		},
		{
			name:         "weekend only range",
			fromDate:     "2024-01-05", // Friday
			toDate:       "2024-01-06", // Saturday
			expectedDays: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromDate, err := time.Parse("2006-01-02", tt.fromDate)
			require.NoError(t, err)

			toDate, err := time.Parse("2006-01-02", tt.toDate)
			require.NoError(t, err)

			days := detector.calculateTradingDays(fromDate, toDate)
			assert.Equal(t, tt.expectedDays, days)
		})
	}
}

func TestModeDetector_extractDateFromFilename(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "valid ISX report filename",
			filename: "2024 01 15 ISX Daily Report.xlsx",
			expected: "2024-01-15",
		},
		{
			name:     "valid ISX report with xls extension",
			filename: "2023 12 25 ISX Daily Report.xls",
			expected: "2023-12-25",
		},
		{
			name:     "filename with extra text",
			filename: "2024 03 10 ISX Daily Report - Final.xlsx",
			expected: "2024-03-10",
		},
		{
			name:     "invalid date format",
			filename: "24 01 15 ISX Daily Report.xlsx",
			expected: "",
		},
		{
			name:     "missing date components",
			filename: "2024 01 ISX Daily Report.xlsx",
			expected: "",
		},
		{
			name:     "invalid filename structure",
			filename: "random_file.xlsx",
			expected: "",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "invalid date values",
			filename: "2024 13 35 ISX Daily Report.xlsx",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractDateFromFilename(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases to improve coverage
func TestModeDetector_DetectOptimalMode_EdgeCases(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "mode_detector_edge_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create detector and override paths for testing
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = tempDir

	t.Run("file detection error fallback", func(t *testing.T) {
		// Create a directory that will cause fileDetector.DetectExcelFiles to return an error
		// by making the directory unreadable
		unreadableDir := filepath.Join(tempDir, "unreadable")
		err := os.MkdirAll(unreadableDir, 0000) // No read permissions
		require.NoError(t, err)

		detector.paths.DownloadsDir = unreadableDir

		result, err := detector.DetectOptimalMode(ctx, "2024-01-01", "2024-01-02")
		require.NoError(t, err) // Should not fail, should assume 0 files
		assert.Equal(t, 0, result.ExistingFiles)
		assert.Equal(t, ModeInitial, result.RecommendedMode)
	})

	t.Run("large date range warning", func(t *testing.T) {
		detector.paths.DownloadsDir = tempDir

		// Test with a very large date range (more than 250 trading days)
		err := detector.ValidateDateRange(ctx, "2024-01-01", "2024-12-31")
		require.NoError(t, err) // Should pass but log a warning
	})
}

func TestModeDetector_ValidateDateRange(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		fromDate  string
		toDate    string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid date range",
			fromDate:  "2024-01-01",
			toDate:    "2024-01-10",
			expectErr: false,
		},
		{
			name:      "same date range",
			fromDate:  "2024-01-01",
			toDate:    "2024-01-01",
			expectErr: false,
		},
		{
			name:      "reversed date range",
			fromDate:  "2024-01-10",
			toDate:    "2024-01-01",
			expectErr: true,
			errMsg:    "cannot be before",
		},
		{
			name:      "invalid from date format",
			fromDate:  "invalid-date",
			toDate:    "2024-01-01",
			expectErr: true,
			errMsg:    "invalid from_date format",
		},
		{
			name:      "invalid to date format",
			fromDate:  "2024-01-01",
			toDate:    "invalid-date",
			expectErr: true,
			errMsg:    "invalid to_date format",
		},
		{
			name:      "weekend only range",
			fromDate:  "2024-01-05", // Friday
			toDate:    "2024-01-06", // Saturday
			expectErr: true,
			errMsg:    "contains no trading days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := detector.ValidateDateRange(ctx, tt.fromDate, tt.toDate)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModeDetector_DetectOptimalMode(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "mode_detector_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create detector and override paths for testing
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = tempDir

	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name             string
		fromDate         string
		toDate           string
		createFiles      []string
		expectedMode     string
		expectedSkip     bool
		expectedCoverage float64
		expectedExisting *int
		expectErr        bool
	}{
		{
			name:             "no existing files - initial mode",
			fromDate:         "2024-01-01", // Monday
			toDate:           "2024-01-05", // Friday (weekend in Iraq)
			createFiles:      []string{},
			expectedMode:     ModeInitial,
			expectedSkip:     false,
			expectedCoverage: 0.0,
		},
		{
			name:     "full coverage - skip mode",
			fromDate: "2024-01-01", // Monday
			toDate:   "2024-01-04", // Thursday (4 trading days)
			createFiles: []string{
				"2024 01 01 ISX Daily Report.xlsx",
				"2024 01 02 ISX Daily Report.xlsx",
				"2024 01 03 ISX Daily Report.xlsx",
				"2024 01 04 ISX Daily Report.xlsx",
			},
			expectedMode:     ModeSkip,
			expectedSkip:     true,
			expectedCoverage: 100.0,
		},
		{
			name:     "full coverage with extra history - skip mode",
			fromDate: "2024-01-01", // Monday
			toDate:   "2024-01-04", // Thursday (4 trading days)
			createFiles: []string{
				"2024 01 01 ISX Daily Report.xlsx",
				"2024 01 02 ISX Daily Report.xlsx",
				"2024 01 03 ISX Daily Report.xlsx",
				"2024 01 04 ISX Daily Report.xlsx",
				"2023 12 31 ISX Daily Report.xlsx",
			},
			expectedMode:     ModeSkip,
			expectedSkip:     true,
			expectedCoverage: 100.0,
			expectedExisting: intPtr(4),
		},
		{
			name:     "good coverage - accumulative mode",
			fromDate: "2024-01-01", // Monday
			toDate:   "2024-01-07", // Sunday (5 trading days: Mon,Tue,Wed,Thu,Sun)
			createFiles: []string{
				"2024 01 01 ISX Daily Report.xlsx",
				"2024 01 02 ISX Daily Report.xlsx",
				"2024 01 03 ISX Daily Report.xlsx",
				"2024 01 07 ISX Daily Report.xlsx",
			},
			expectedMode:     ModeAccumulative,
			expectedSkip:     false,
			expectedCoverage: 80.0,
		},
		{
			name:     "partial coverage - accumulative mode",
			fromDate: "2024-01-01", // Monday
			toDate:   "2024-01-07", // Sunday (5 trading days)
			createFiles: []string{
				"2024 01 01 ISX Daily Report.xlsx",
				"2024 01 02 ISX Daily Report.xlsx",
			},
			expectedMode:     ModeAccumulative,
			expectedSkip:     false,
			expectedCoverage: 40.0,
		},
		{
			name:      "invalid date range",
			fromDate:  "invalid-date",
			toDate:    "2024-01-01",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp directory
			os.RemoveAll(tempDir)
			os.MkdirAll(tempDir, 0755)

			// Create test files
			for _, filename := range tt.createFiles {
				filePath := filepath.Join(tempDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			result, err := detector.DetectOptimalMode(ctx, tt.fromDate, tt.toDate)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedMode, result.RecommendedMode)
			assert.Equal(t, tt.expectedSkip, result.SkipRecommendation)
			assert.InDelta(t, tt.expectedCoverage, result.CoveragePercent, 0.1)
			expectedExisting := len(tt.createFiles)
			if tt.expectedExisting != nil {
				expectedExisting = *tt.expectedExisting
			}
			assert.Equal(t, expectedExisting, result.ExistingFiles)
			assert.GreaterOrEqual(t, result.MissingDays, 0)
			if tt.expectedSkip {
				assert.Equal(t, 0, result.MissingDays)
			}
			assert.NotEmpty(t, result.Reason)
		})
	}
}

func TestModeDetector_GetExistingFileDates(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mode_detector_dates_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = tempDir

	// Create test files
	testFiles := []string{
		"2024 01 15 ISX Daily Report.xlsx",
		"2024 01 16 ISX Daily Report.xlsx",
		"2023 12 25 ISX Daily Report.xls",
		"invalid_filename.xlsx",
		"readme.txt", // Should be ignored
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	dates, err := detector.GetExistingFileDates(ctx)
	require.NoError(t, err)

	expectedDates := []string{
		"2024-01-15",
		"2024-01-16",
		"2023-12-25",
	}

	assert.ElementsMatch(t, expectedDates, dates)
}

func TestModeDetector_GetExistingFileDates_NonexistentDirectory(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	detector, err := NewModeDetector(logger)
	require.NoError(t, err)
	detector.paths.DownloadsDir = "/nonexistent/directory"

	dates, err := detector.GetExistingFileDates(ctx)
	require.NoError(t, err)
	assert.Empty(t, dates)
}

func TestModeDetector_GetModeRecommendationSummary(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		result   *ModeDetectionResult
		expected string
	}{
		{
			name:     "nil result",
			result:   nil,
			expected: "No mode detection result available",
		},
		{
			name: "skip recommendation",
			result: &ModeDetectionResult{
				RecommendedMode:    "skip",
				TradingDays:        5,
				ExistingFiles:      5,
				CoveragePercent:    100.0,
				Reason:             "High coverage",
				SkipRecommendation: true,
			},
			expected: "Mode Detection Analysis:",
		},
		{
			name: "accumulative mode",
			result: &ModeDetectionResult{
				RecommendedMode:    ModeAccumulative,
				TradingDays:        10,
				ExistingFiles:      8,
				CoveragePercent:    80.0,
				Reason:             "Good coverage",
				SkipRecommendation: false,
			},
			expected: "Mode Detection Analysis:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := detector.GetModeRecommendationSummary(tt.result)
			assert.Contains(t, summary, tt.expected)

			if tt.result != nil {
				assert.Contains(t, summary, tt.result.RecommendedMode)
				if tt.result.SkipRecommendation {
					assert.Contains(t, summary, "Consider skipping operation")
				}
			}
		})
	}
}

// Edge case tests for defensive programming
func TestModeDetector_NilChecks(t *testing.T) {
	ctx := context.Background()

	t.Run("DetectOptimalMode with nil detector", func(t *testing.T) {
		var detector *ModeDetector
		result, err := detector.DetectOptimalMode(ctx, "2024-01-01", "2024-01-02")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ModeDetector is nil")
	})

	t.Run("GetExistingFileDates with nil detector", func(t *testing.T) {
		var detector *ModeDetector
		dates, err := detector.GetExistingFileDates(ctx)
		assert.Error(t, err)
		assert.Nil(t, dates)
		assert.Contains(t, err.Error(), "ModeDetector is nil")
	})

	t.Run("ValidateDateRange with nil detector", func(t *testing.T) {
		var detector *ModeDetector
		err := detector.ValidateDateRange(ctx, "2024-01-01", "2024-01-02")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ModeDetector is nil")
	})
}

// Benchmark tests for performance validation
func BenchmarkModeDetector_calculateTradingDays(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	detector, err := NewModeDetector(logger)
	require.NoError(b, err)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.calculateTradingDays(startDate, endDate)
	}
}

func BenchmarkModeDetector_DetectOptimalMode(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create temporary directory with some test files
	tempDir, err := os.MkdirTemp("", "mode_detector_bench")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	detector, err := NewModeDetector(logger)
	require.NoError(b, err)
	detector.paths.DownloadsDir = tempDir

	// Create some test files
	for i := 1; i <= 50; i++ {
		filename := fmt.Sprintf("2024 01 %02d ISX Daily Report.xlsx", i)
		filePath := filepath.Join(tempDir, filename)
		os.WriteFile(filePath, []byte("test"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectOptimalMode(ctx, "2024-01-01", "2024-01-31")
	}
}
