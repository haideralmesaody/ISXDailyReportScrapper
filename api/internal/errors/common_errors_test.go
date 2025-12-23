package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{
			name:     "license error type",
			errType:  ErrTypeLicense,
			expected: "LICENSE",
		},
		{
			name:     "network error type",
			errType:  ErrTypeNetwork,
			expected: "NETWORK",
		},
		{
			name:     "parsing error type",
			errType:  ErrTypeParsing,
			expected: "PARSING",
		},
		{
			name:     "storage error type",
			errType:  ErrTypeStorage,
			expected: "STORAGE",
		},
		{
			name:     "validation error type",
			errType:  ErrTypeValidation,
			expected: "VALIDATION",
		},
		{
			name:     "not found error type",
			errType:  ErrTypeNotFound,
			expected: "NOT_FOUND",
		},
		{
			name:     "permission error type",
			errType:  ErrTypePermission,
			expected: "PERMISSION",
		},
		{
			name:     "config error type",
			errType:  ErrTypeConfig,
			expected: "CONFIG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.errType))
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name        string
		appError    *AppError
		wantMessage string
	}{
		{
			name: "error without cause",
			appError: &AppError{
				Type:    ErrTypeLicense,
				Message: "License validation failed",
				Cause:   nil,
			},
			wantMessage: "[LICENSE] License validation failed",
		},
		{
			name: "error with cause",
			appError: &AppError{
				Type:    ErrTypeNetwork,
				Message: "Failed to connect to server",
				Cause:   fmt.Errorf("connection refused"),
			},
			wantMessage: "[NETWORK] Failed to connect to server: connection refused",
		},
		{
			name: "error with complex cause",
			appError: &AppError{
				Type:    ErrTypeStorage,
				Message: "Database operation failed",
				Cause:   errors.New("table does not exist"),
			},
			wantMessage: "[STORAGE] Database operation failed: table does not exist",
		},
		{
			name: "error with empty message",
			appError: &AppError{
				Type:    ErrTypeValidation,
				Message: "",
				Cause:   nil,
			},
			wantMessage: "[VALIDATION] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appError.Error()
			assert.Equal(t, tt.wantMessage, got)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		wantErr  error
	}{
		{
			name: "unwrap with cause",
			appError: &AppError{
				Type:    ErrTypeLicense,
				Message: "License error",
				Cause:   fmt.Errorf("original error"),
			},
			wantErr: fmt.Errorf("original error"),
		},
		{
			name: "unwrap without cause",
			appError: &AppError{
				Type:    ErrTypeNetwork,
				Message: "Network error",
				Cause:   nil,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appError.Unwrap()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), got.Error())
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	tests := []struct {
		name          string
		appError      *AppError
		key           string
		value         interface{}
		expectedValue interface{}
	}{
		{
			name: "add string context",
			appError: &AppError{
				Type:    ErrTypeLicense,
				Message: "License error",
			},
			key:           "user_id",
			value:         "12345",
			expectedValue: "12345",
		},
		{
			name: "add integer context",
			appError: &AppError{
				Type:    ErrTypeNetwork,
				Message: "Network error",
			},
			key:           "retry_count",
			value:         3,
			expectedValue: 3,
		},
		{
			name: "add complex object context",
			appError: &AppError{
				Type:    ErrTypeStorage,
				Message: "Storage error",
			},
			key:           "query",
			value:         map[string]string{"table": "users", "operation": "select"},
			expectedValue: map[string]string{"table": "users", "operation": "select"},
		},
		{
			name: "add context to error with existing context",
			appError: &AppError{
				Type:    ErrTypeValidation,
				Message: "Validation error",
				Context: map[string]interface{}{"field": "email"},
			},
			key:           "value",
			value:         "invalid@",
			expectedValue: "invalid@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.WithContext(tt.key, tt.value)
			
			// Should return the same instance
			assert.Same(t, tt.appError, result)
			
			// Should have the context value
			require.Contains(t, result.Context, tt.key)
			assert.Equal(t, tt.expectedValue, result.Context[tt.key])
			
			// Should initialize context if it was nil
			assert.NotNil(t, result.Context)
		})
	}
}

func TestAppError_WithContext_NilContext(t *testing.T) {
	t.Run("add context to error with nil context", func(t *testing.T) {
		appError := &AppError{
			Type:    ErrTypeLicense,
			Message: "Test error",
			Context: nil,
		}
		
		result := appError.WithContext("test_key", "test_value")
		
		assert.NotNil(t, result.Context)
		assert.Equal(t, "test_value", result.Context["test_key"])
	})
}

func TestNewAppError(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		message  string
		cause    error
		wantType ErrorType
		wantMsg  string
		wantCause error
	}{
		{
			name:      "create license error",
			errType:   ErrTypeLicense,
			message:   "License validation failed",
			cause:     fmt.Errorf("expired"),
			wantType:  ErrTypeLicense,
			wantMsg:   "License validation failed", 
			wantCause: fmt.Errorf("expired"),
		},
		{
			name:      "create error without cause",
			errType:   ErrTypeNetwork,
			message:   "Connection failed",
			cause:     nil,
			wantType:  ErrTypeNetwork,
			wantMsg:   "Connection failed",
			wantCause: nil,
		},
		{
			name:      "create validation error",
			errType:   ErrTypeValidation,
			message:   "Invalid input",
			cause:     errors.New("field required"),
			wantType:  ErrTypeValidation,
			wantMsg:   "Invalid input",
			wantCause: errors.New("field required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAppError(tt.errType, tt.message, tt.cause)
			
			assert.Equal(t, tt.wantType, got.Type)
			assert.Equal(t, tt.wantMsg, got.Message)
			
			if tt.wantCause != nil {
				require.NotNil(t, got.Cause)
				assert.Equal(t, tt.wantCause.Error(), got.Cause.Error())
			} else {
				assert.Nil(t, got.Cause)
			}
			
			// Should initialize empty context
			assert.NotNil(t, got.Context)
			assert.Empty(t, got.Context)
		})
	}
}

func TestNewLicenseError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "license error with cause",
			message: "License validation failed",
			cause:   fmt.Errorf("expired"),
		},
		{
			name:    "license error without cause",
			message: "License not found",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLicenseError(tt.message, tt.cause)
			
			assert.Equal(t, ErrTypeLicense, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewNetworkError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "network error with cause",
			message: "Connection failed",
			cause:   fmt.Errorf("timeout"),
		},
		{
			name:    "network error without cause",
			message: "Network unavailable",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNetworkError(tt.message, tt.cause)
			
			assert.Equal(t, ErrTypeNetwork, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewParsingError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "parsing error with cause",
			message: "Failed to parse JSON",
			cause:   fmt.Errorf("invalid character"),
		},
		{
			name:    "parsing error without cause",
			message: "Parse failed",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewParsingError(tt.message, tt.cause)
			
			assert.Equal(t, ErrTypeParsing, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewStorageError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "storage error with cause",
			message: "Database connection failed",
			cause:   fmt.Errorf("connection refused"),
		},
		{
			name:    "storage error without cause",
			message: "Storage unavailable",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStorageError(tt.message, tt.cause)
			
			assert.Equal(t, ErrTypeStorage, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewAppValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "validation error",
			message: "Field validation failed",
		},
		{
			name:    "empty validation message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAppValidationError(tt.message)
			
			assert.Equal(t, ErrTypeValidation, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Nil(t, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantMsg  string
	}{
		{
			name:     "user not found",
			resource: "user",
			wantMsg:  "user not found",
		},
		{
			name:     "license not found",
			resource: "license",
			wantMsg:  "license not found",
		},
		{
			name:     "file not found",
			resource: "file",
			wantMsg:  "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNotFoundError(tt.resource)
			
			assert.Equal(t, ErrTypeNotFound, got.Type)
			assert.Equal(t, tt.wantMsg, got.Message)
			assert.Nil(t, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewPermissionError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "access denied",
			message: "Access denied to resource",
		},
		{
			name:    "insufficient permissions",
			message: "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPermissionError(tt.message)
			
			assert.Equal(t, ErrTypePermission, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Nil(t, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestNewConfigError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "config error with cause",
			message: "Failed to load configuration",
			cause:   fmt.Errorf("file not found"),
		},
		{
			name:    "config error without cause",
			message: "Invalid configuration",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfigError(tt.message, tt.cause)
			
			assert.Equal(t, ErrTypeConfig, got.Type)
			assert.Equal(t, tt.message, got.Message)
			assert.Equal(t, tt.cause, got.Cause)
			assert.NotNil(t, got.Context)
		})
	}
}

func TestAppError_ErrorsIntegration(t *testing.T) {
	t.Run("errors.Is works with AppError", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		appErr := NewLicenseError("License failed", originalErr)
		
		// Should work with errors.Is
		assert.True(t, errors.Is(appErr, originalErr))
		
		// Should not match different error
		otherErr := fmt.Errorf("other error")
		assert.False(t, errors.Is(appErr, otherErr))
	})

	t.Run("errors.As works with AppError", func(t *testing.T) {
		originalErr := &AppError{
			Type:    ErrTypeNetwork,
			Message: "Network error",
		}
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)
		
		var appErr *AppError
		assert.True(t, errors.As(wrappedErr, &appErr))
		assert.Equal(t, ErrTypeNetwork, appErr.Type)
		assert.Equal(t, "Network error", appErr.Message)
	})
}

func TestAppError_ContextChaining(t *testing.T) {
	t.Run("chain multiple context values", func(t *testing.T) {
		appErr := NewLicenseError("License validation failed", nil)
		
		result := appErr.
			WithContext("user_id", "12345").
			WithContext("license_key", "ISX1Y-XXXXX").
			WithContext("attempt", 3)
		
		// Should be the same instance
		assert.Same(t, appErr, result)
		
		// Should have all context values
		assert.Equal(t, "12345", result.Context["user_id"])
		assert.Equal(t, "ISX1Y-XXXXX", result.Context["license_key"])
		assert.Equal(t, 3, result.Context["attempt"])
	})

	t.Run("overwrite existing context value", func(t *testing.T) {
		appErr := NewNetworkError("Connection failed", nil)
		
		result := appErr.
			WithContext("retry_count", 1).
			WithContext("retry_count", 2) // Overwrite
		
		assert.Equal(t, 2, result.Context["retry_count"])
	})
}

func TestAppError_ComplexScenarios(t *testing.T) {
	t.Run("nested error unwrapping", func(t *testing.T) {
		// Create a chain of errors
		rootErr := fmt.Errorf("root cause")
		appErr1 := NewStorageError("Database error", rootErr)
		appErr2 := NewNetworkError("Connection error", appErr1)
		
		// Should unwrap correctly
		assert.True(t, errors.Is(appErr2, appErr1))
		assert.True(t, errors.Is(appErr2, rootErr))
		
		// Should match AppError types - errors.As finds the first match (appErr2, not appErr1)
		var networkErr *AppError
		assert.True(t, errors.As(appErr2, &networkErr))
		assert.Equal(t, ErrTypeNetwork, networkErr.Type)
	})

	t.Run("error with rich context", func(t *testing.T) {
		appErr := NewParsingError("Failed to parse JSON", fmt.Errorf("invalid syntax")).
			WithContext("file_path", "/data/input.json").
			WithContext("line_number", 42).
			WithContext("column", 15).
			WithContext("parser_version", "1.2.3")
		
		expected := "[PARSING] Failed to parse JSON: invalid syntax"
		assert.Equal(t, expected, appErr.Error())
		
		// Verify context is preserved
		assert.Equal(t, "/data/input.json", appErr.Context["file_path"])
		assert.Equal(t, 42, appErr.Context["line_number"])
		assert.Equal(t, 15, appErr.Context["column"])
		assert.Equal(t, "1.2.3", appErr.Context["parser_version"])
	})
}

func TestAppError_EdgeCases(t *testing.T) {
	t.Run("nil cause unwrap", func(t *testing.T) {
		appErr := &AppError{
			Type:    ErrTypeValidation,
			Message: "Validation failed",
			Cause:   nil,
		}
		
		assert.Nil(t, appErr.Unwrap())
	})

	t.Run("empty context handling", func(t *testing.T) {
		appErr := &AppError{
			Type:    ErrTypeConfig,
			Message: "Config error",
			Context: make(map[string]interface{}),
		}
		
		result := appErr.WithContext("key", "value")
		assert.Equal(t, "value", result.Context["key"])
	})

	t.Run("context with nil values", func(t *testing.T) {
		appErr := NewLicenseError("License error", nil)
		
		result := appErr.WithContext("nullable_field", nil)
		assert.Contains(t, result.Context, "nullable_field")
		assert.Nil(t, result.Context["nullable_field"])
	})
}