package strategy

import (
	"fmt"

	"github.com/isxcli/isxcli/internal/errors"
)

// Strategy-specific error functions using the project's error pattern

// NewStrategyNotFoundError creates a strategy not found error
func NewStrategyNotFoundError(strategyID string) error {
	return errors.NewAppError(errors.ErrTypeValidation,
		fmt.Sprintf("strategy not found: %s", strategyID), nil)
}

// NewInvalidStrategyError creates an invalid strategy error
func NewInvalidStrategyError(message string) error {
	return errors.NewAppError(errors.ErrTypeValidation,
		fmt.Sprintf("invalid strategy: %s", message), nil)
}

// NewStrategyExistsError creates a strategy already exists error
func NewStrategyExistsError(strategyID string) error {
	return errors.NewAppError(errors.ErrTypeValidation,
		fmt.Sprintf("strategy already exists: %s", strategyID), nil)
}

// NewInsufficientDataError creates an insufficient data error
func NewInsufficientDataError(required, available int) error {
	return errors.NewAppError(errors.ErrTypeValidation,
		fmt.Sprintf("insufficient data: need %d points, have %d", required, available), nil)
}

// NewExecutionError creates a strategy execution error
func NewExecutionError(strategyID string, cause error) error {
	return errors.NewAppError(errors.ErrTypeParsing,
		fmt.Sprintf("strategy execution failed: %s", strategyID), cause)
}

// NewBacktestError creates a backtest error
func NewBacktestError(strategyID string, cause error) error {
	return errors.NewAppError(errors.ErrTypeParsing,
		fmt.Sprintf("backtest failed: %s", strategyID), cause)
}

// NewInvalidParametersError creates an invalid parameters error
func NewInvalidParametersError(param string, value interface{}) error {
	return errors.NewAppError(errors.ErrTypeValidation,
		fmt.Sprintf("invalid parameter %s: %v", param, value), nil)
}