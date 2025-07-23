package processor

import (
	"isxcli/internal/parser"
)

// Processor defines the interface for data processing operations
type Processor interface {
	// Process takes raw trade records and returns processed records
	Process(records []parser.TradeRecord) ([]parser.TradeRecord, error)
}

// ProcessingOptions configures processing behavior
type ProcessingOptions struct {
	// EnableForwardFill enables forward-fill for missing data
	EnableForwardFill bool
	
	// SkipWeekends excludes weekends from forward-fill
	SkipWeekends bool
	
	// MaxFillDays limits how many days to forward-fill
	MaxFillDays int
}

// DefaultOptions returns default processing options
func DefaultOptions() ProcessingOptions {
	return ProcessingOptions{
		EnableForwardFill: true,
		SkipWeekends:      true,
		MaxFillDays:       0, // 0 means no limit
	}
}