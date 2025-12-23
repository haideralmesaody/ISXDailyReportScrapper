package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name: "simple message",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "INVALID_REQUEST",
				Message:    "Invalid request format",
			},
			want: "Invalid request format",
		},
		{
			name: "empty message",
			apiError: &APIError{
				StatusCode: http.StatusInternalServerError,
				ErrorCode:  "INTERNAL_ERROR",
				Message:    "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAPIError_Render(t *testing.T) {
	tests := []struct {
		name       string
		apiError   *APIError
		wantStatus int
	}{
		{
			name: "bad request error",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "INVALID_REQUEST",
				Message:    "Invalid request format",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal server error",
			apiError: &APIError{
				StatusCode: http.StatusInternalServerError,
				ErrorCode:  "INTERNAL_ERROR",
				Message:    "Internal server error",
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "not found error",
			apiError: &APIError{
				StatusCode: http.StatusNotFound,
				ErrorCode:  "NOT_FOUND",
				Message:    "Resource not found",
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			err := tt.apiError.Render(w, r)
			assert.NoError(t, err)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		message    string
		want       *APIError
	}{
		{
			name:       "create bad request error",
			statusCode: http.StatusBadRequest,
			errorCode:  "INVALID_REQUEST",
			message:    "Invalid request format",
			want: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "INVALID_REQUEST",
				Message:    "Invalid request format",
				Details:    nil,
			},
		},
		{
			name:       "create internal error",
			statusCode: http.StatusInternalServerError,
			errorCode:  "INTERNAL_ERROR",
			message:    "Something went wrong",
			want: &APIError{
				StatusCode: http.StatusInternalServerError,
				ErrorCode:  "INTERNAL_ERROR",
				Message:    "Something went wrong",
				Details:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.statusCode, tt.errorCode, tt.message)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewWithDetails(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		message    string
		details    interface{}
		want       *APIError
	}{
		{
			name:       "create error with string details",
			statusCode: http.StatusBadRequest,
			errorCode:  "VALIDATION_FAILED",
			message:    "Validation failed",
			details:    "field 'name' is required",
			want: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "VALIDATION_FAILED",
				Message:    "Validation failed",
				Details:    "field 'name' is required",
			},
		},
		{
			name:       "create error with map details",
			statusCode: http.StatusBadRequest,
			errorCode:  "VALIDATION_FAILED",
			message:    "Validation failed",
			details:    map[string]string{"field": "name", "error": "required"},
			want: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "VALIDATION_FAILED",
				Message:    "Validation failed",
				Details:    map[string]string{"field": "name", "error": "required"},
			},
		},
		{
			name:       "create error with validation error details",
			statusCode: http.StatusBadRequest,
			errorCode:  "VALIDATION_FAILED",
			message:    "Validation failed",
			details:    ValidationError{Field: "email", Message: "invalid format"},
			want: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "VALIDATION_FAILED",
				Message:    "Validation failed",
				Details:    ValidationError{Field: "email", Message: "invalid format"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithDetails(tt.statusCode, tt.errorCode, tt.message, tt.details)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *APIError
		wantStatus int
		wantCode   string
	}{
		{
			name:       "ErrInvalidRequest",
			err:        ErrInvalidRequest,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "ErrValidationFailed",
			err:        ErrValidationFailed,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_FAILED",
		},
		{
			name:       "ErrUnauthorized",
			err:        ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "ErrForbidden",
			err:        ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantCode:   "FORBIDDEN",
		},
		{
			name:       "ErrNotFound",
			err:        ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "ErrInternalServer",
			err:        ErrInternalServer,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_SERVER_ERROR",
		},
		{
			name:       "ErrRateLimitExceeded",
			err:        ErrRateLimitExceeded,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "RATE_LIMIT_EXCEEDED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, tt.err.StatusCode)
			assert.Equal(t, tt.wantCode, tt.err.ErrorCode)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestInvalidRequestWithError(t *testing.T) {
	tests := []struct {
		name      string
		inputErr  error
		wantCode  string
		wantStatus int
	}{
		{
			name:       "with simple error",
			inputErr:   assert.AnError,
			wantCode:   "INVALID_REQUEST",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "with custom error",
			inputErr:   New(http.StatusBadRequest, "CUSTOM", "custom error"),
			wantCode:   "INVALID_REQUEST", 
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InvalidRequestWithError(tt.inputErr)
			
			assert.Equal(t, tt.wantStatus, got.StatusCode)
			assert.Equal(t, tt.wantCode, got.ErrorCode)
			assert.Equal(t, "Invalid request format", got.Message)
			assert.Equal(t, tt.inputErr.Error(), got.Details)
		})
	}
}

func TestErrValidation(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		message   string
		wantField string
		wantMsg   string
	}{
		{
			name:      "email validation error",
			field:     "email",
			message:   "invalid email format",
			wantField: "email",
			wantMsg:   "invalid email format",
		},
		{
			name:      "password validation error",
			field:     "password",
			message:   "password too short",
			wantField: "password",
			wantMsg:   "password too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrValidation(tt.field, tt.message)
			
			assert.Equal(t, http.StatusBadRequest, got.StatusCode)
			assert.Equal(t, "VALIDATION_FAILED", got.ErrorCode)
			assert.Equal(t, "Request validation failed", got.Message)
			
			// Check details contain ValidationError
			validationErr, ok := got.Details.(ValidationError)
			require.True(t, ok, "Details should be ValidationError type")
			assert.Equal(t, tt.wantField, validationErr.Field)
			assert.Equal(t, tt.wantMsg, validationErr.Message)
		})
	}
}

func TestNotFoundError(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NotFoundError(tt.resource)
			
			assert.Equal(t, http.StatusNotFound, got.StatusCode)
			assert.Equal(t, "NOT_FOUND", got.ErrorCode)
			assert.Equal(t, tt.wantMsg, got.Message)
			assert.Equal(t, tt.resource, got.Details)
		})
	}
}

func TestLicenseNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		inputErr error
	}{
		{
			name:     "with file not found error",
			inputErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LicenseNotFoundError(tt.inputErr)
			
			assert.Equal(t, http.StatusNotFound, got.StatusCode)
			assert.Equal(t, "LICENSE_NOT_FOUND", got.ErrorCode)
			assert.Equal(t, "License file not found", got.Message)
			assert.Equal(t, tt.inputErr.Error(), got.Details)
		})
	}
}

func TestErrLicenseActivation(t *testing.T) {
	tests := []struct {
		name     string
		inputErr error
	}{
		{
			name:     "activation failed",
			inputErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrLicenseActivation(tt.inputErr)
			
			assert.Equal(t, http.StatusBadRequest, got.StatusCode)
			assert.Equal(t, "LICENSE_ACTIVATION_FAILED", got.ErrorCode)
			assert.Equal(t, "Failed to activate license", got.Message)
			assert.Equal(t, tt.inputErr.Error(), got.Details)
		})
	}
}

func TestErrPipelineExecution(t *testing.T) {
	tests := []struct {
		name     string
		inputErr error
	}{
		{
			name:     "pipeline failed",
			inputErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrPipelineExecution(tt.inputErr)
			
			assert.Equal(t, http.StatusInternalServerError, got.StatusCode)
			assert.Equal(t, "PIPELINE_EXECUTION_FAILED", got.ErrorCode)
			assert.Equal(t, "operation execution failed", got.Message)
			assert.Equal(t, tt.inputErr.Error(), got.Details)
		})
	}
}

func TestFileSystemError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		inputErr  error
		wantMsg   string
	}{
		{
			name:      "read operation failed",
			operation: "read",
			inputErr:  assert.AnError,
			wantMsg:   "File system error during read",
		},
		{
			name:      "write operation failed",
			operation: "write",
			inputErr:  assert.AnError,
			wantMsg:   "File system error during write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileSystemError(tt.operation, tt.inputErr)
			
			assert.Equal(t, http.StatusInternalServerError, got.StatusCode)
			assert.Equal(t, "FILESYSTEM_ERROR", got.ErrorCode)
			assert.Equal(t, tt.wantMsg, got.Message)
			assert.Equal(t, tt.inputErr.Error(), got.Details)
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
	}{
		{
			name: "create error response",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "TEST_ERROR",
				Message:    "Test error message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponse(tt.apiError)
			
			assert.False(t, got.Success)
			assert.Equal(t, tt.apiError, got.Error)
		})
	}
}

func TestErrorResponse_Render(t *testing.T) {
	tests := []struct {
		name       string
		apiError   *APIError
		wantStatus int
	}{
		{
			name: "render error response",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "TEST_ERROR",
				Message:    "Test error message",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errResp := NewErrorResponse(tt.apiError)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			err := errResp.Render(w, r)
			assert.NoError(t, err)
		})
	}
}

func TestNewValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors []ValidationError
	}{
		{
			name: "single validation error",
			errors: []ValidationError{
				{Field: "email", Message: "invalid format"},
			},
		},
		{
			name: "multiple validation errors",
			errors: []ValidationError{
				{Field: "email", Message: "invalid format"},
				{Field: "password", Message: "too short"},
			},
		},
		{
			name:   "empty validation errors",
			errors: []ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidationErrors(tt.errors)
			
			assert.Equal(t, http.StatusBadRequest, got.StatusCode)
			assert.Equal(t, "VALIDATION_FAILED", got.ErrorCode)
			assert.Equal(t, "Request validation failed", got.Message)
			
			// Check details
			validationErrs, ok := got.Details.(ValidationErrors)
			require.True(t, ok, "Details should be ValidationErrors type")
			assert.Equal(t, tt.errors, validationErrs.Errors)
		})
	}
}

func TestErrPanic(t *testing.T) {
	tests := []struct {
		name      string
		recovered interface{}
		wantMsg   string
	}{
		{
			name:      "string panic",
			recovered: "something went wrong",
			wantMsg:   "something went wrong",
		},
		{
			name:      "error panic",
			recovered: assert.AnError,
			wantMsg:   assert.AnError.Error(),
		},
		{
			name:      "integer panic",
			recovered: 42,
			wantMsg:   "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrPanic(tt.recovered)
			
			assert.Equal(t, http.StatusInternalServerError, got.StatusCode)
			assert.Equal(t, "INTERNAL_SERVER_ERROR", got.ErrorCode)
			assert.Equal(t, "Internal server error", got.Message)
			
			// Check details
			panicRecovery, ok := got.Details.(PanicRecovery)
			require.True(t, ok, "Details should be PanicRecovery type")
			assert.Equal(t, tt.wantMsg, panicRecovery.Message)
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		apiError   *APIError
		wantStatus int
	}{
		{
			name: "write bad request error",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "TEST_ERROR",
				Message:    "Test error message",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "write internal error",
			apiError: &APIError{
				StatusCode: http.StatusInternalServerError,
				ErrorCode:  "INTERNAL_ERROR",
				Message:    "Internal server error",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteError(w, tt.apiError)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			
			// Decode response body
			var response ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			
			assert.False(t, response.Success)
			assert.Equal(t, tt.apiError.StatusCode, response.Error.StatusCode)
			assert.Equal(t, tt.apiError.ErrorCode, response.Error.ErrorCode)
			assert.Equal(t, tt.apiError.Message, response.Error.Message)
		})
	}
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple validation error",
			message: "field is required",
		},
		{
			name:    "empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidationError(tt.message)
			
			assert.Equal(t, http.StatusBadRequest, got.StatusCode)
			assert.Equal(t, "VALIDATION_FAILED", got.ErrorCode)
			assert.Equal(t, tt.message, got.Message)
		})
	}
}

func TestNewInternalError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple internal error",
			message: "database connection failed",
		},
		{
			name:    "empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInternalError(tt.message)
			
			assert.Equal(t, http.StatusInternalServerError, got.StatusCode)
			assert.Equal(t, "INTERNAL_SERVER_ERROR", got.ErrorCode)
			assert.Equal(t, tt.message, got.Message)
		})
	}
}

func TestAPIError_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
	}{
		{
			name: "serialize basic error",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "TEST_ERROR",
				Message:    "Test message",
			},
		},
		{
			name: "serialize error with details",
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "VALIDATION_FAILED",
				Message:    "Validation failed",
				Details:    ValidationError{Field: "email", Message: "invalid"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.apiError)
			require.NoError(t, err)
			
			// Test JSON unmarshaling
			var unmarshaled APIError
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			
			assert.Equal(t, tt.apiError.StatusCode, unmarshaled.StatusCode)
			assert.Equal(t, tt.apiError.ErrorCode, unmarshaled.ErrorCode)
			assert.Equal(t, tt.apiError.Message, unmarshaled.Message)
		})
	}
}

func TestAPIErrorsIntegrationWithRender(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
	}{
		{
			name: "render APIError directly", 
			apiError: &APIError{
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "TEST_ERROR",
				Message:    "Test message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			err := render.Render(w, r, tt.apiError)
			assert.NoError(t, err)
			
			// Verify the response was written properly
			var response APIError
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			
			assert.Equal(t, tt.apiError.StatusCode, response.StatusCode)
			assert.Equal(t, tt.apiError.ErrorCode, response.ErrorCode)
			assert.Equal(t, tt.apiError.Message, response.Message)
		})
	}
}