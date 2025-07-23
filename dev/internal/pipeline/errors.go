package pipeline

import (
	"fmt"
)

// ErrorType represents the type of pipeline error
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeDependency   ErrorType = "dependency"
	ErrorTypeExecution    ErrorType = "execution"
	ErrorTypeTimeout      ErrorType = "timeout"
	ErrorTypeCancellation ErrorType = "cancellation"
	ErrorTypeRetryable    ErrorType = "retryable"
	ErrorTypeFatal        ErrorType = "fatal"
)

// PipelineError represents a pipeline-specific error
type PipelineError struct {
	Type      ErrorType              `json:"type"`
	Stage     string                 `json:"stage,omitempty"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Retryable bool                   `json:"retryable"`
}

// Error implements the error interface
func (e *PipelineError) Error() string {
	if e.Stage != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Type, e.Stage, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *PipelineError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(stage, message string) *PipelineError {
	return &PipelineError{
		Type:      ErrorTypeValidation,
		Stage:     stage,
		Message:   message,
		Retryable: false,
	}
}

// NewDependencyError creates a new dependency error
func NewDependencyError(stage, dependsOn, message string) *PipelineError {
	return &PipelineError{
		Type:    ErrorTypeDependency,
		Stage:   stage,
		Message: message,
		Context: map[string]interface{}{
			"depends_on": dependsOn,
		},
		Retryable: false,
	}
}

// NewExecutionError creates a new execution error
func NewExecutionError(stage string, cause error, retryable bool) *PipelineError {
	return &PipelineError{
		Type:      ErrorTypeExecution,
		Stage:     stage,
		Message:   "Stage execution failed",
		Cause:     cause,
		Retryable: retryable,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(stage string, timeout string) *PipelineError {
	return &PipelineError{
		Type:    ErrorTypeTimeout,
		Stage:   stage,
		Message: fmt.Sprintf("Stage exceeded timeout of %s", timeout),
		Context: map[string]interface{}{
			"timeout": timeout,
		},
		Retryable: true,
	}
}

// NewCancellationError creates a new cancellation error
func NewCancellationError(stage string) *PipelineError {
	return &PipelineError{
		Type:      ErrorTypeCancellation,
		Stage:     stage,
		Message:   "Pipeline was cancelled",
		Retryable: false,
	}
}

// NewFatalError creates a new fatal error
func NewFatalError(message string, cause error) *PipelineError {
	return &PipelineError{
		Type:      ErrorTypeFatal,
		Message:   message,
		Cause:     cause,
		Retryable: false,
	}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if pErr, ok := err.(*PipelineError); ok {
		return pErr.Retryable
	}
	return false
}

// GetErrorType returns the type of the error
func GetErrorType(err error) ErrorType {
	if err == nil {
		return ""
	}
	if pErr, ok := err.(*PipelineError); ok {
		return pErr.Type
	}
	return ErrorTypeExecution
}

// WrapError wraps an error with pipeline context
func WrapError(err error, stage string, message string) *PipelineError {
	if err == nil {
		return nil
	}
	
	// If it's already a PipelineError, enhance it
	if pErr, ok := err.(*PipelineError); ok {
		if pErr.Stage == "" {
			pErr.Stage = stage
		}
		if message != "" {
			pErr.Message = fmt.Sprintf("%s: %s", message, pErr.Message)
		}
		return pErr
	}
	
	// Otherwise create a new execution error
	return &PipelineError{
		Type:      ErrorTypeExecution,
		Stage:     stage,
		Message:   message,
		Cause:     err,
		Retryable: false,
	}
}

// ErrorList represents multiple errors
type ErrorList struct {
	Errors []*PipelineError `json:"errors"`
}

// Error implements the error interface
func (e *ErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors: %d errors occurred", len(e.Errors))
}

// Add adds an error to the list
func (e *ErrorList) Add(err *PipelineError) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (e *ErrorList) HasErrors() bool {
	return len(e.Errors) > 0
}

// GetByStage returns errors for a specific stage
func (e *ErrorList) GetByStage(stage string) []*PipelineError {
	var stageErrors []*PipelineError
	for _, err := range e.Errors {
		if err.Stage == stage {
			stageErrors = append(stageErrors, err)
		}
	}
	return stageErrors
}