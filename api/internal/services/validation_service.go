package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/operations"
)

// ValidationService provides centralized validation for all operations and stages
type ValidationService struct {
	logger *slog.Logger
}

// NewValidationService creates a new validation service
func NewValidationService(logger *slog.Logger) *ValidationService {
	return &ValidationService{
		logger: logger,
	}
}

// ValidateStageParameters validates parameters specific to each stage
func (vs *ValidationService) ValidateStageParameters(ctx context.Context, stageID string, parameters map[string]interface{}) error {
	logger := vs.logger.With(
		slog.String("component", "validation_service"),
		slog.String("method", "ValidateStageParameters"),
		slog.String("stage_id", stageID),
	)

	switch stageID {
	case operations.StageIDScraping:
		return vs.validateScrapingParameters(ctx, parameters)
	case operations.StageIDProcessing, operations.StageIDIndices, operations.StageIDLiquidity:
		// These stages don't require specific parameters
		logger.DebugContext(ctx, "stage_does_not_require_parameter_validation",
			slog.String("stage_id", stageID))
		return nil
	default:
		logger.WarnContext(ctx, "unknown_stage_type_for_validation",
			slog.String("stage_id", stageID))
		return nil
	}
}

// validateScrapingParameters validates scraping-specific parameters
func (vs *ValidationService) validateScrapingParameters(ctx context.Context, parameters map[string]interface{}) error {
	_ = vs.logger.With(
		slog.String("component", "validation_service"),
		slog.String("method", "validateScrapingParameters"),
	)

	// Validate mode parameter
	if mode, ok := parameters["mode"].(string); ok {
		validModes := []string{"initial", "incremental", "backfill"}
		valid := false
		for _, validMode := range validModes {
			if mode == validMode {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid scraping mode '%s'. Valid modes: %v", mode, validModes)
		}
	}

	// Validate date parameters
	if from, ok := parameters["from"].(string); ok {
		if err := vs.validateDateParameter(from, "from"); err != nil {
			return fmt.Errorf("invalid from date for scraping: %w", err)
		}
	}
	if to, ok := parameters["to"].(string); ok {
		if err := vs.validateDateParameter(to, "to"); err != nil {
			return fmt.Errorf("invalid to date for scraping: %w", err)
		}
	}

	return nil
}


// validateDateParameter validates a date parameter in YYYY-MM-DD format
func (vs *ValidationService) validateDateParameter(dateStr, paramName string) error {
	if dateStr == "" {
		return nil // Empty dates are allowed
	}

	// Parse date to validate format
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format for %s: '%s'. Expected format: YYYY-MM-DD", paramName, dateStr)
	}

	// Check if date is reasonable (not too far in past or future)
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err // Shouldn't happen given above check
	}

	now := time.Now()
	minDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	maxDate := now.AddDate(1, 0, 0) // Allow dates up to 1 year in future

	if parsedDate.Before(minDate) {
		return fmt.Errorf("date %s for %s is too old (minimum: 2020-01-01)", dateStr, paramName)
	}

	if parsedDate.After(maxDate) {
		return fmt.Errorf("date %s for %s is too far in future (maximum: %s)", dateStr, paramName, maxDate.Format("2006-01-02"))
	}

	return nil
}

// validateDateRange validates that from_date is before or equal to to_date
func (vs *ValidationService) validateDateRange(fromDate, toDate string) error {
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return fmt.Errorf("invalid from_date format: %w", err)
	}

	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return fmt.Errorf("invalid to_date format: %w", err)
	}

	if from.After(to) {
		return fmt.Errorf("from_date (%s) cannot be after to_date (%s)", fromDate, toDate)
	}

	return nil
}

// validatePeriodList validates a comma-separated list of periods
func (vs *ValidationService) validatePeriodList(periodsStr, paramName string) error {
	if periodsStr == "" {
		return nil
	}

	periods := strings.Split(periodsStr, ",")
	for i, periodStr := range periods {
		periodStr = strings.TrimSpace(periodStr)
		if periodStr == "" {
			continue
		}

		period, err := strconv.Atoi(periodStr)
		if err != nil {
			return fmt.Errorf("invalid period in %s at position %d: '%s' is not a number", paramName, i+1, periodStr)
		}

		if period < 1 || period > 500 {
			return fmt.Errorf("period in %s at position %d is out of range (1-500): %d", paramName, i+1, period)
		}
	}

	return nil
}

// validateMACDParameters validates MACD parameters (fast,slow,signal)
func (vs *ValidationService) validateMACDParameters(macdParams string) error {
	if macdParams == "" {
		return nil
	}

	parts := strings.Split(macdParams, ",")
	if len(parts) != 3 {
		return fmt.Errorf("MACD parameters must have exactly 3 values (fast,slow,signal), got: %s", macdParams)
	}

	// Parse and validate each parameter
	fast, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return fmt.Errorf("invalid MACD fast period: '%s'", parts[0])
	}
	if fast < 1 || fast > 100 {
		return fmt.Errorf("MACD fast period must be between 1 and 100, got: %d", fast)
	}

	slow, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return fmt.Errorf("invalid MACD slow period: '%s'", parts[1])
	}
	if slow < 1 || slow > 200 {
		return fmt.Errorf("MACD slow period must be between 1 and 200, got: %d", slow)
	}

	signal, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return fmt.Errorf("invalid MACD signal period: '%s'", parts[2])
	}
	if signal < 1 || signal > 50 {
		return fmt.Errorf("MACD signal period must be between 1 and 50, got: %d", signal)
	}

	// Validate logical relationships
	if fast >= slow {
		return fmt.Errorf("MACD fast period (%d) must be less than slow period (%d)", fast, slow)
	}

	if signal >= slow {
		return fmt.Errorf("MACD signal period (%d) must be less than slow period (%d)", signal, slow)
	}

	return nil
}

// ValidateOperationRequest validates a complete operation request
func (vs *ValidationService) ValidateOperationRequest(ctx context.Context, request *operations.OperationRequest) error {
	logger := vs.logger.With(
		slog.String("component", "validation_service"),
		slog.String("method", "ValidateOperationRequest"),
		slog.String("operation_id", request.ID),
	)

	if request == nil {
		return errors.NewValidationError("operation request cannot be nil")
	}

	if request.ID == "" {
		return errors.NewValidationError("operation ID cannot be empty")
	}

	// Validate date range if both dates are provided
	if request.FromDate != "" && request.ToDate != "" {
		if err := vs.validateDateRange(request.FromDate, request.ToDate); err != nil {
			return fmt.Errorf("invalid date range: %w", err)
		}
	}

	// Validate individual date parameters
	if request.FromDate != "" {
		if err := vs.validateDateParameter(request.FromDate, "from_date"); err != nil {
			return fmt.Errorf("invalid from_date: %w", err)
		}
	}
	if request.ToDate != "" {
		if err := vs.validateDateParameter(request.ToDate, "to_date"); err != nil {
			return fmt.Errorf("invalid to_date: %w", err)
		}
	}

	logger.DebugContext(ctx, "operation_request_validation_passed",
		slog.String("operation_id", request.ID))

	return nil
}
