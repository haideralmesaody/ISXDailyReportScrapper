package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	
	apierrors "isxcli/internal/errors"
)

// ValidationMiddleware provides request validation using struct tags
type ValidationMiddleware struct {
	validator    *validator.Validate
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
	maxBodySize  int64
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(logger *slog.Logger, errorHandler *apierrors.ErrorHandler) *ValidationMiddleware {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("iso8601", isISO8601)
	v.RegisterValidation("ticker", isValidTicker)
	v.RegisterValidation("filename", isValidFilename)
	
	// Use JSON tag names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &ValidationMiddleware{
		validator:    v,
		logger:       logger.With(slog.String("component", "validation_middleware")),
		errorHandler: errorHandler,
		maxBodySize:  10 * 1024 * 1024, // 10MB default
	}
}

// ValidateRequest validates request body against a struct with validation tags
func (m *ValidationMiddleware) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip validation for GET, HEAD, OPTIONS
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		
		// Check content length
		if r.ContentLength > m.maxBodySize {
			m.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusRequestEntityTooLarge,
				"PAYLOAD_TOO_LARGE",
				"Request body exceeds maximum allowed size",
				map[string]interface{}{
					"max_size": m.maxBodySize,
					"size":     r.ContentLength,
				},
			))
			return
		}
		
		// Read and validate body if present
		if r.Body != nil && r.ContentLength > 0 {
			// Read body
			body, err := io.ReadAll(io.LimitReader(r.Body, m.maxBodySize))
			if err != nil {
				m.logger.ErrorContext(r.Context(), "failed to read request body",
					slog.String("error", err.Error()),
					slog.String("request_id", middleware.GetReqID(r.Context())),
				)
				m.errorHandler.HandleError(w, r, apierrors.InvalidRequestWithError(err))
				return
			}
			
			// Restore body for handlers
			r.Body = io.NopCloser(bytes.NewReader(body))
			
			// Validate JSON structure
			if !json.Valid(body) && len(body) > 0 {
				m.errorHandler.HandleError(w, r, apierrors.New(
					http.StatusBadRequest,
					"INVALID_JSON",
					"Request body contains invalid JSON",
				))
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// ValidateStruct validates a struct and returns validation errors
func (m *ValidationMiddleware) ValidateStruct(v interface{}) error {
	if err := m.validator.Struct(v); err != nil {
		// Convert to validation errors
		var validationErrors []apierrors.ValidationError
		
		for _, err := range err.(validator.ValidationErrors) {
			ve := apierrors.ValidationError{
				Field:   err.Field(),
				Message: m.formatValidationError(err),
			}
			validationErrors = append(validationErrors, ve)
		}
		
		return apierrors.NewValidationErrors(validationErrors)
	}
	return nil
}

// ContentTypeValidator ensures requests have proper content type
func ContentTypeValidator(contentTypes ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for GET, HEAD, DELETE
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}
			
			// Check content type
			contentType := r.Header.Get("Content-Type")
			if contentType == "" {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, apierrors.New(
					http.StatusBadRequest,
					"MISSING_CONTENT_TYPE",
					"Content-Type header is required",
				))
				return
			}
			
			// Validate against allowed types
			valid := false
			for _, allowed := range contentTypes {
				if strings.HasPrefix(contentType, allowed) {
					valid = true
					break
				}
			}
			
			if !valid {
				render.Status(r, http.StatusUnsupportedMediaType)
				render.JSON(w, r, apierrors.NewWithDetails(
					http.StatusUnsupportedMediaType,
					"UNSUPPORTED_MEDIA_TYPE",
					"Unsupported content type",
					map[string]interface{}{
						"content_type": contentType,
						"allowed":      contentTypes,
					},
				))
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// formatValidationError formats validation error messages
func (m *ValidationMiddleware) formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Replace(param, " ", ", ", -1))
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "iso8601":
		return fmt.Sprintf("%s must be a valid ISO8601 date", field)
	case "ticker":
		return fmt.Sprintf("%s must be a valid ticker symbol", field)
	case "filename":
		return fmt.Sprintf("%s must be a valid filename", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	default:
		return fmt.Sprintf("%s failed %s validation", field, tag)
	}
}

// Custom validators

// isISO8601 validates ISO8601 date format
func isISO8601(fl validator.FieldLevel) bool {
	date := fl.Field().String()
	// Simple ISO8601 validation (YYYY-MM-DD)
	if len(date) != 10 {
		return false
	}
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return false
	}
	// Basic validation - could be enhanced
	return len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2
}

// isValidTicker validates ticker symbol format
func isValidTicker(fl validator.FieldLevel) bool {
	ticker := fl.Field().String()
	if len(ticker) < 1 || len(ticker) > 10 {
		return false
	}
	// Only allow letters, numbers, and dots
	for _, ch := range ticker {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '.') {
			return false
		}
	}
	return true
}

// isValidFilename validates filename format
func isValidFilename(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if filename == "" {
		return false
	}
	// Prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}
	// Basic filename validation
	return len(filename) <= 255
}

// QueryParamValidator validates query parameters
type QueryParamValidator struct {
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
}

// NewQueryParamValidator creates a new query parameter validator
func NewQueryParamValidator(logger *slog.Logger, errorHandler *apierrors.ErrorHandler) *QueryParamValidator {
	return &QueryParamValidator{
		logger:       logger.With(slog.String("component", "query_validator")),
		errorHandler: errorHandler,
	}
}

// ValidateInt validates an integer query parameter
func (v *QueryParamValidator) ValidateInt(w http.ResponseWriter, r *http.Request, param string, min, max int, defaultValue int) (int, bool) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue, true
	}
	
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		v.errorHandler.HandleError(w, r, apierrors.ErrValidation(param, fmt.Sprintf("%s must be a valid integer", param)))
		return 0, false
	}
	
	if intValue < min || intValue > max {
		v.errorHandler.HandleError(w, r, apierrors.ErrValidation(param, fmt.Sprintf("%s must be between %d and %d", param, min, max)))
		return 0, false
	}
	
	return intValue, true
}

// ValidateEnum validates an enum query parameter
func (v *QueryParamValidator) ValidateEnum(w http.ResponseWriter, r *http.Request, param string, allowed []string, defaultValue string) (string, bool) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue, true
	}
	
	for _, a := range allowed {
		if value == a {
			return value, true
		}
	}
	
	v.errorHandler.HandleError(w, r, apierrors.ErrValidation(param, fmt.Sprintf("%s must be one of: %s", param, strings.Join(allowed, ", "))))
	return "", false
}