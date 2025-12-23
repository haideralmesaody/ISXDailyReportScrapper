package errors

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrTypeLicense     ErrorType = "LICENSE"
	ErrTypeNetwork     ErrorType = "NETWORK"
	ErrTypeParsing     ErrorType = "PARSING"
	ErrTypeStorage     ErrorType = "STORAGE"
	ErrTypeValidation  ErrorType = "VALIDATION"
	ErrTypeNotFound    ErrorType = "NOT_FOUND"
	ErrTypePermission  ErrorType = "PERMISSION"
	ErrTypeConfig      ErrorType = "CONFIG"
)

// AppError represents an application-specific error
type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap allows errors.Is and errors.As to work with AppError
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ToError returns the error interface (implemented for compatibility)
func (e *AppError) ToError() error {
	return e
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewAppError creates a new application error
func NewAppError(errType ErrorType, message string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Helper functions for common error types

// NewLicenseError creates a license-related error
func NewLicenseError(message string, cause error) *AppError {
	return NewAppError(ErrTypeLicense, message, cause)
}

// NewNetworkError creates a network-related error
func NewNetworkError(message string, cause error) *AppError {
	return NewAppError(ErrTypeNetwork, message, cause)
}

// NewParsingError creates a parsing-related error
func NewParsingError(message string, cause error) *AppError {
	return NewAppError(ErrTypeParsing, message, cause)
}

// NewStorageError creates a storage-related error
func NewStorageError(message string, cause error) *AppError {
	return NewAppError(ErrTypeStorage, message, cause)
}

// NewAppValidationError creates a validation error for AppError type
func NewAppValidationError(message string) *AppError {
	return NewAppError(ErrTypeValidation, message, nil)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrTypeNotFound, fmt.Sprintf("%s not found", resource), nil)
}

// NewPermissionError creates a permission error
func NewPermissionError(message string) *AppError {
	return NewAppError(ErrTypePermission, message, nil)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, cause error) *AppError {
	return NewAppError(ErrTypeConfig, message, cause)
}

// NewStageDependencyError creates a stage dependency error
func NewStageDependencyError(stage, dependency string) *AppError {
	return NewAppError(ErrTypeConfig, fmt.Sprintf("Stage dependency error: '%s' depends on '%s'", stage, dependency), nil)
}