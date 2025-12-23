package operations

import (
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// Analyzer defines the interface for data analysis operations
type Analyzer interface {
	// Analyze performs analysis on trade records
	Analyze(records []domain.TradeRecord) (interface{}, error)
}

// AnalysisOptions configures analysis behavior
type AnalysisOptions struct {
	// IncludeForwardFilled includes forward-filled records in analysis
	IncludeForwardFilled bool

	// DateRange limits analysis to specific date range
	StartDate string
	EndDate   string

	// Tickers limits analysis to specific tickers
	Tickers []string
}

// Statistics represents general trading statistics
type Statistics struct {
	TotalRecords     int
	UniqueTickers    int
	DateRange        string
	TotalVolume      int64
	TotalValue       float64
	MostActiveTicker string
	MostActiveVolume int64
}
