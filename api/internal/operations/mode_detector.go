package operations

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

// ModeDetector automatically determines the optimal operation mode based on existing files
type ModeDetector struct {
	logger       *slog.Logger
	fileDetector *FileDetector
	paths        *config.Paths
}

// NewModeDetector creates a new mode detector with proper dependency injection
func NewModeDetector(logger *slog.Logger) (*ModeDetector, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	paths, err := config.GetPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get config paths: %w", err)
	}

	return &ModeDetector{
		logger:       logger,
		fileDetector: NewFileDetector(logger),
		paths:        paths,
	}, nil
}

// ModeDetectionResult contains the result of mode detection analysis
type ModeDetectionResult struct {
	RecommendedMode    string  `json:"recommended_mode"`
	ExpectedFiles      int     `json:"expected_files"`
	ExistingFiles      int     `json:"existing_files"`
	CoveragePercent    float64 `json:"coverage_percent"`
	TradingDays        int     `json:"trading_days"`
	MissingDays        int     `json:"missing_days"`
	Reason             string  `json:"reason"`
	SkipRecommendation bool    `json:"skip_recommendation"`
}

// DetectOptimalMode analyzes the date range and existing files to determine the best operation mode
func (md *ModeDetector) DetectOptimalMode(ctx context.Context, fromDate, toDate string) (*ModeDetectionResult, error) {
	if md == nil {
		return nil, fmt.Errorf("ModeDetector is nil")
	}

	md.logger.InfoContext(ctx, "Starting mode detection analysis",
		slog.String("from_date", fromDate),
		slog.String("to_date", toDate),
		slog.String("downloads_dir", md.paths.DownloadsDir))

	// Parse dates
	startDate, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date format '%s': %w", fromDate, err)
	}

	endDate, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date format '%s': %w", toDate, err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("to_date '%s' cannot be before from_date '%s'", toDate, fromDate)
	}

	// Calculate expected trading days (excluding Iraqi weekends: Friday and Saturday)
	tradingDays := md.calculateTradingDays(startDate, endDate)

	md.logger.InfoContext(ctx, "Calculated expected trading days",
		slog.Int("trading_days", tradingDays),
		slog.String("period", fmt.Sprintf("%s to %s", fromDate, toDate)))

	// Count existing Excel files and normalize for the requested range
	totalExistingFiles, err := md.fileDetector.DetectExcelFiles(md.paths.DownloadsDir)
	if err != nil {
		md.logger.WarnContext(ctx, "Failed to detect existing files, assuming none exist",
			slog.String("error", err.Error()))
		totalExistingFiles = 0
	}

	existingDates, datesErr := md.GetExistingFileDates(ctx)
	if datesErr != nil {
		md.logger.WarnContext(ctx, "Failed to enumerate existing file dates, assuming none exist",
			slog.String("error", datesErr.Error()))
		existingDates = []string{}
	}

	existingFiles := md.countDatesInRange(existingDates, startDate, endDate)
	if existingFiles > totalExistingFiles {
		totalExistingFiles = existingFiles
	}

	// Calculate coverage percentage
	var coveragePercent float64
	if tradingDays > 0 {
		coveragePercent = float64(existingFiles) / float64(tradingDays) * 100.0
		if coveragePercent > 100 {
			coveragePercent = 100
		}
	}

	missingDays := tradingDays - existingFiles
	if missingDays < 0 {
		missingDays = 0
	}

	md.logger.InfoContext(ctx, "File coverage analysis",
		slog.Int("existing_files_in_range", existingFiles),
		slog.Int("existing_files_total", totalExistingFiles),
		slog.Int("expected_files", tradingDays),
		slog.Int("missing_days", missingDays),
		slog.Float64("coverage_percent", coveragePercent))

	// Determine optimal mode based on coverage
	result := &ModeDetectionResult{
		ExpectedFiles:   tradingDays,
		ExistingFiles:   existingFiles,
		CoveragePercent: coveragePercent,
		TradingDays:     tradingDays,
		MissingDays:     missingDays,
	}

	switch {
	case tradingDays > 0 && existingFiles >= tradingDays:
		result.RecommendedMode = ModeSkip
		result.SkipRecommendation = true
		result.Reason = fmt.Sprintf("All %d trading days already have downloaded reports - skipping scrape", tradingDays)

		md.logger.InfoContext(ctx, "Mode detection: skip selected",
			slog.Int("existing_files_in_range", existingFiles),
			slog.Int("trading_days", tradingDays),
			slog.String("operation", "mode_detection"),
			slog.String("trace_id", infrastructure.GetTraceID(ctx)))
	case existingFiles > 0:
		result.RecommendedMode = ModeAccumulative
		result.SkipRecommendation = false
		result.Reason = fmt.Sprintf("Found %d existing files in requested range - using accumulative mode to add new data", existingFiles)

		md.logger.InfoContext(ctx, "Mode detection: accumulative selected",
			slog.Int("existing_files_in_range", existingFiles),
			slog.String("operation", "mode_detection"),
			slog.String("trace_id", infrastructure.GetTraceID(ctx)))
	default:
		result.RecommendedMode = ModeInitial
		result.SkipRecommendation = false
		result.Reason = "No existing files found - using initial mode for fresh download"

		md.logger.InfoContext(ctx, "Mode detection: initial selected",
			slog.String("operation", "mode_detection"),
			slog.String("trace_id", infrastructure.GetTraceID(ctx)))
	}

	md.logger.InfoContext(ctx, "Mode detection completed",
		slog.String("recommended_mode", result.RecommendedMode),
		slog.Bool("skip_recommendation", result.SkipRecommendation),
		slog.String("reason", result.Reason))

	return result, nil
}

// calculateTradingDays calculates the number of expected trading days between two dates
// Excludes Iraqi weekends (Friday and Saturday)
func (md *ModeDetector) calculateTradingDays(startDate, endDate time.Time) int {
	tradingDays := 0
	current := startDate

	// Iterate through each day in the range
	for !current.After(endDate) {
		// Iraqi Stock Exchange is closed on Friday and Saturday
		// In Go's time.Weekday: Sunday=0, Monday=1, ..., Friday=5, Saturday=6
		if current.Weekday() != time.Friday && current.Weekday() != time.Saturday {
			tradingDays++
		}
		current = current.AddDate(0, 0, 1)
	}

	return tradingDays
}

// GetExistingFileDates returns a list of dates for which Excel files already exist
func (md *ModeDetector) GetExistingFileDates(ctx context.Context) ([]string, error) {
	if md == nil {
		return nil, fmt.Errorf("ModeDetector is nil")
	}

	entries, err := os.ReadDir(md.paths.DownloadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			md.logger.WarnContext(ctx, "Downloads directory does not exist",
				slog.String("directory", md.paths.DownloadsDir))
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read downloads directory: %w", err)
	}

	var existingDates []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") &&
			!strings.HasSuffix(strings.ToLower(filename), ".xls") {
			continue
		}

		// Extract date from filename
		// Expected format: "YYYY MM DD ISX Daily Report.xlsx"
		if date := md.extractDateFromFilename(filename); date != "" {
			existingDates = append(existingDates, date)
		}
	}

	md.logger.InfoContext(ctx, "Found existing file dates",
		slog.Int("count", len(existingDates)),
		slog.Any("dates", existingDates))

	return existingDates, nil
}

// countDatesInRange returns how many existing report dates fall within the requested period
func (md *ModeDetector) countDatesInRange(dates []string, startDate, endDate time.Time) int {
	count := 0

	for _, dateStr := range dates {
		if dateStr == "" {
			continue
		}

		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			if md.logger != nil {
				md.logger.Debug("Failed to parse existing file date",
					slog.String("value", dateStr),
					slog.String("error", err.Error()))
			}
			continue
		}

		if (parsed.Equal(startDate) || parsed.After(startDate)) && (parsed.Equal(endDate) || parsed.Before(endDate)) {
			count++
		}
	}

	return count
}

// extractDateFromFilename extracts date from ISX report filename
// Expects format: "YYYY MM DD ISX Daily Report.xlsx"
// Returns date in YYYY-MM-DD format or empty string if not found
func (md *ModeDetector) extractDateFromFilename(filename string) string {
	// Remove extension
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Expected format: "YYYY MM DD ISX Daily Report"
	parts := strings.Fields(baseName)

	if len(parts) >= 3 {
		// Try to parse first three parts as date components
		if len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2 {
			// Validate by attempting to parse as date
			dateStr := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
			if _, err := time.Parse("2006-01-02", dateStr); err == nil {
				return dateStr
			}
		}
	}

	return ""
}

// ValidateDateRange validates that the date range contains at least one trading day
func (md *ModeDetector) ValidateDateRange(ctx context.Context, fromDate, toDate string) error {
	if md == nil {
		return fmt.Errorf("ModeDetector is nil")
	}

	startDate, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return fmt.Errorf("invalid from_date format '%s': expected YYYY-MM-DD", fromDate)
	}

	endDate, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return fmt.Errorf("invalid to_date format '%s': expected YYYY-MM-DD", toDate)
	}

	if endDate.Before(startDate) {
		return fmt.Errorf("to_date '%s' cannot be before from_date '%s'", toDate, fromDate)
	}

	tradingDays := md.calculateTradingDays(startDate, endDate)
	if tradingDays == 0 {
		return fmt.Errorf("date range '%s' to '%s' contains no trading days (only weekends)", fromDate, toDate)
	}

	// Warn if date range is very large
	if tradingDays > 250 { // Approximately 1 year of trading days
		md.logger.WarnContext(ctx, "Large date range detected",
			slog.String("from_date", fromDate),
			slog.String("to_date", toDate),
			slog.Int("trading_days", tradingDays),
			slog.String("warning", "consider breaking into smaller batches"))
	}

	return nil
}

// GetModeRecommendationSummary provides a human-readable summary of the mode recommendation
func (md *ModeDetector) GetModeRecommendationSummary(result *ModeDetectionResult) string {
	if result == nil {
		return "No mode detection result available"
	}

	summary := fmt.Sprintf(
		"Mode Detection Analysis:\n"+
			"• Trading Days: %d\n"+
			"• Existing Files: %d\n"+
			"• Coverage: %.1f%%\n"+
			"• Recommended Mode: %s\n"+
			"• Reason: %s",
		result.TradingDays,
		result.ExistingFiles,
		result.CoveragePercent,
		result.RecommendedMode,
		result.Reason,
	)

	if result.SkipRecommendation {
		summary += "\n• Action: Consider skipping operation as most files exist"
	}

	return summary
}
