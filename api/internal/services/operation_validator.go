package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	licenseErrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
)

type OperationValidator struct {
	licenseManager      license.ManagerInterface
	logger              *slog.Logger
	validateExecutables func(context.Context) error
}

func NewOperationValidator(
	licenseManager license.ManagerInterface,
	logger *slog.Logger,
	validateExecutables func(ctx context.Context) error,
) *OperationValidator {
	return &OperationValidator{
		licenseManager:      licenseManager,
		logger:              logger,
		validateExecutables: validateExecutables,
	}
}

func (ov *OperationValidator) Validate(ctx context.Context, operationType string) error {
	if ov == nil {
		return fmt.Errorf("operation validator not initialized")
	}

	start := time.Now()
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = infrastructure.GenerateTraceID()
		ctx = infrastructure.WithTraceID(ctx, traceID)
	}

	logger := ov.logger.With(
		slog.String("trace_id", traceID),
		slog.String("operation_type", operationType),
		slog.String("component", "operations_service"),
		slog.String("method", "ValidateLicenseForOperation"),
	)

	logger.InfoContext(ctx, "License validation started for operation",
		slog.String("operation_type", operationType))

	info, status, statusErr := ov.licenseManager.GetLicenseStatus()
	if statusErr == nil && ov.isLicenseStatusActive(info, status) {
		logger.InfoContext(ctx, "License status is active - using cached validation result",
			slog.String("cached_status", status),
			slog.Duration("validation_duration", time.Since(start)))

		if validationErr := ov.validateOperationSpecificRequirements(ctx, operationType, info, status); validationErr != nil {
			logger.ErrorContext(ctx, "Operation-specific validation failed with cached status",
				slog.String("error", validationErr.Error()),
				slog.Duration("total_validation_duration", time.Since(start)))
			return validationErr
		}

		return nil
	}

	valid, err := ov.licenseManager.ValidateLicense()
	if err != nil {
		logger.ErrorContext(ctx, "License validation failed",
			slog.String("error", err.Error()),
			slog.Duration("validation_duration", time.Since(start)))

		return licenseErrors.NewLicenseValidationFailed(
			ctx,
			"",
			fmt.Sprintf("License validation failed for %s operation: %v", operationType, err),
			err,
		).WithExtension("operation_type", operationType).
			WithExtension("validation_duration_ms", time.Since(start).Milliseconds()).
			WithExtension("component", "operations_service")
	}

	if !valid {
		logger.ErrorContext(ctx, "License validation returned invalid status",
			slog.Bool("valid", valid),
			slog.Duration("validation_duration", time.Since(start)))

		info, status, _ := ov.licenseManager.GetLicenseStatus()

		var licenseErr error
		switch status {
		case "Not Activated":
			licenseErr = licenseErrors.NewLicenseNotActivatedError(ctx).
				WithExtension("operation_type", operationType).
				WithExtension("component", "operations_service")
		case "Expired":
			if info != nil {
				licenseErr = licenseErrors.NewLicenseExpiredError(ctx, info.ExpiryDate, info.LicenseKey).
					WithExtension("operation_type", operationType).
					WithExtension("component", "operations_service")
			} else {
				licenseErr = licenseErrors.NewLicenseNotActivatedError(ctx).
					WithExtension("operation_type", operationType).
					WithExtension("component", "operations_service")
			}
		default:
			licenseErr = licenseErrors.NewLicenseValidationFailed(
				ctx,
				"",
				fmt.Sprintf("License is not valid for %s operation. Current status: %s", operationType, status),
				licenseErrors.ErrLicenseValidationFailed,
			).WithExtension("operation_type", operationType).
				WithExtension("license_status", status).
				WithExtension("component", "operations_service")
		}

		return licenseErr
	}

	logger.InfoContext(ctx, "License validation successful",
		slog.Duration("validation_duration", time.Since(start)))

	if err := ov.validateOperationSpecificRequirements(ctx, operationType, info, status); err != nil {
		logger.ErrorContext(ctx, "Operation-specific validation failed",
			slog.String("error", err.Error()),
			slog.Duration("total_validation_duration", time.Since(start)))
		return err
	}

	logger.InfoContext(ctx, "License validation and operation-specific checks completed successfully",
		slog.Duration("total_validation_duration", time.Since(start)))

	return nil
}

func (ov *OperationValidator) validateOperationSpecificRequirements(ctx context.Context, operationType string, info *license.LicenseInfo, status string) error {
	traceID := infrastructure.GetTraceID(ctx)
	logger := ov.logger.With(
		slog.String("trace_id", traceID),
		slog.String("operation_type", operationType),
		slog.String("component", "operations_service"),
		slog.String("method", "validateOperationSpecificRequirements"),
	)

	if info == nil {
		var err error
		info, status, err = ov.licenseManager.GetLicenseStatus()
		if err != nil {
			logger.WarnContext(ctx, "Could not get license status for operation-specific validation",
				slog.String("error", err.Error()))
		}
	}

	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus == "" && info != nil {
		normalizedStatus = strings.TrimSpace(info.Status)
	}
	if normalizedStatus == "" {
		normalizedStatus = "Unknown"
	}

	switch operationType {
	case "run_stage":
		return ov.validateActiveLicenseRequirement(ctx, "individual stage execution", info, normalizedStatus)
	case "scraping", "scrape":
		return ov.validateActiveLicenseRequirement(ctx, "ISX data scraping", info, normalizedStatus)
	case "processing", "process":
		return ov.validateActiveLicenseRequirement(ctx, "data processing", info, normalizedStatus)
	case "indices", "index", "indexcsv":
		return ov.validateActiveLicenseRequirement(ctx, "index extraction", info, normalizedStatus)
	case "liquidity":
		return ov.validateActiveLicenseRequirement(ctx, "liquidity calculation", info, normalizedStatus)
	case "upload", "sheets":
		return ov.validateUploadOperationRequirements(ctx, info, normalizedStatus)
	case "", "general":
		logger.DebugContext(ctx, "General operation - basic license validation sufficient")
		return nil
	default:
		if ov.isValidStageName(operationType) {
			return ov.validateActiveLicenseRequirement(ctx, fmt.Sprintf("stage execution: %s", operationType), info, normalizedStatus)
		}

		logger.WarnContext(ctx, "Unknown operation type, applying active license requirement",
			slog.String("operation_type", operationType))
		return ov.validateActiveLicenseRequirement(ctx, "custom operation", info, normalizedStatus)
	}
}

func (ov *OperationValidator) validateActiveLicenseRequirement(ctx context.Context, operationName string, info *license.LicenseInfo, status string) error {
	traceID := infrastructure.GetTraceID(ctx)
	logger := ov.logger.With(
		slog.String("trace_id", traceID),
		slog.String("operation_name", operationName),
		slog.String("license_status", status),
	)

	switch status {
	case "Active", "Activated", "Valid":
		logger.DebugContext(ctx, "License is active for operation")
		return nil
	case "Warning":
		logger.InfoContext(ctx, "License in warning status but operation allowed",
			slog.String("message", "License expires soon. Consider renewal to avoid interruption."))
		return nil
	case "Critical":
		logger.WarnContext(ctx, "License in critical status but operation allowed",
			slog.String("message", "License expires very soon. Immediate renewal recommended to avoid service interruption."))
		return nil
	case "Not Activated":
		return licenseErrors.NewLicenseNotActivatedError(ctx).
			WithExtension("operation_name", operationName).
			WithExtension("required_action", "activate_license").
			WithExtension("component", "operations_service")
	case "Expired":
		if info != nil {
			return licenseErrors.NewLicenseExpiredError(ctx, info.ExpiryDate, info.LicenseKey).
				WithExtension("operation_name", operationName).
				WithExtension("required_action", "renew_license").
				WithExtension("component", "operations_service")
		}
		return licenseErrors.NewLicenseNotActivatedError(ctx).
			WithExtension("operation_name", operationName).
			WithExtension("required_action", "renew_license").
			WithExtension("component", "operations_service")
	default:
		return licenseErrors.NewLicenseValidationFailed(
			ctx,
			"",
			fmt.Sprintf("License status '%s' is not valid for %s. Active license required.", status, operationName),
			licenseErrors.ErrLicenseValidationFailed,
		).WithExtension("operation_name", operationName).
			WithExtension("license_status", status).
			WithExtension("required_status", "Active").
			WithExtension("component", "operations_service")
	}
}

func (ov *OperationValidator) validateUploadOperationRequirements(ctx context.Context, info *license.LicenseInfo, status string) error {
	if err := ov.validateActiveLicenseRequirement(ctx, "Google Sheets upload", info, status); err != nil {
		return err
	}

	traceID := infrastructure.GetTraceID(ctx)
	logger := ov.logger.With(
		slog.String("trace_id", traceID),
		slog.String("operation_name", "Google Sheets upload"),
	)

	logger.DebugContext(ctx, "Google Sheets upload requirements validated")
	return nil
}

func (ov *OperationValidator) isValidStageName(operationType string) bool {
	knownStages := []string{
		"scraping", "processing", "indices", "liquidity",
	}

	for _, stage := range knownStages {
		if operationType == stage {
			return true
		}
	}

	return false
}

func (ov *OperationValidator) isLicenseStatusActive(info *license.LicenseInfo, status string) bool {
	normalized := strings.TrimSpace(strings.ToLower(status))
	if normalized == "" && info != nil {
		normalized = strings.TrimSpace(strings.ToLower(info.Status))
	}

	switch normalized {
	case "active", "activated", "valid", "warning", "critical":
		if info != nil && !info.ExpiryDate.IsZero() && time.Now().After(info.ExpiryDate) {
			return false
		}
		return true
	default:
		return false
	}
}
