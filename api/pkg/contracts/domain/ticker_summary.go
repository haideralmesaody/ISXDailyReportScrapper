package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TickerSummary represents the Single Source of Truth (SSOT) for ticker summary data.
// This structure defines the authoritative format for ticker summaries across the entire
// ISX Daily Reports Scrapper system. All consumers, exporters, APIs, and processors
// must use this structure for ticker summary operations.
//
// Design Principles:
// - Single Source of Truth for all ticker summary data
// - Support both CSV and JSON serialization with proper field mapping
// - Backward compatibility with existing formats
// - Extensible for future metrics without breaking changes
// - Clear field documentation and validation rules
//
// Usage:
//   summary := &TickerSummary{
//       Ticker: "BBOB",
//       CompanyName: "Bank of Baghdad",
//       LastPrice: 1.250,
//       LastDate: "2024-01-15", // Last date where TradingStatus=true
//       TradingDays: 120,
//       Last10Days: []float64{1.200, 1.210, 1.220, 1.240, 1.250},
//   }
type TickerSummary struct {
	// === CORE FIELDS (always required) ===
	
	// Ticker is the stock symbol/ticker code (e.g., "BBOB", "BMNS")
	// Must be 3-5 uppercase letters, no spaces or special characters
	Ticker string `json:"ticker" csv:"Ticker" validate:"required,min=2,max=10"`
	
	// CompanyName is the full company name in English
	// Used for display purposes and human readability
	CompanyName string `json:"company_name" csv:"CompanyName" validate:"required,min=3,max=255"`
	
	// LastPrice is the most recent trading price from the last actual trading day
	// This is the closing price from the last date where TradingStatus=true
	// Precision: 3 decimal places, minimum value: 0.001
	LastPrice float64 `json:"last_price" csv:"LastPrice" validate:"required,min=0.001"`
	
	// LastDate is the last calendar date with actual trading activity
	// Format: "2006-01-02" (ISO 8601 date format)
	// This MUST be the most recent date where TradingStatus=true in the source data
	LastDate string `json:"last_date" csv:"LastDate" validate:"required"`
	
	// TradingDays is the count of days with actual trading activity (not calendar days)
	// Only counts days where TradingStatus=true OR (Volume > 0 OR NumTrades > 0)
	// Used to understand liquidity and market activity levels
	TradingDays int `json:"trading_days" csv:"TradingDays" validate:"min=0"`
	
	// Last10Days contains the last 10 actual trading day closing prices
	// Array format for JSON: [1.200, 1.210, 1.220, ...]
	// CSV format: comma-separated string "1.200,1.210,1.220,..."
	// Ordered chronologically (oldest to newest)
	// Maximum 10 elements, may be fewer if insufficient trading history
	Last10Days []float64 `json:"last_10_days" csv:"Last10Days" validate:"max=10"`
	
	// === EXTENDED METRICS (optional, for enhanced analysis) ===
	
	// TotalVolume is the aggregate trading volume across all trading days
	// Includes only days with actual trading activity (TradingStatus=true)
	TotalVolume int64 `json:"total_volume,omitempty" csv:"TotalVolume,omitempty" validate:"min=0"`
	
	// TotalValue is the aggregate trading value (price × volume) across all trading days
	// Calculated only for days with actual trading activity
	// Precision: 3 decimal places
	TotalValue float64 `json:"total_value,omitempty" csv:"TotalValue,omitempty" validate:"min=0"`
	
	// AveragePrice is the mean closing price across all trading days
	// Calculated as sum of closing prices / number of trading days
	// Excludes days without trading activity
	AveragePrice float64 `json:"average_price,omitempty" csv:"AveragePrice,omitempty" validate:"min=0"`
	
	// HighestPrice is the highest closing price across all available data
	// Considers only days with actual trading activity
	HighestPrice float64 `json:"highest_price,omitempty" csv:"HighestPrice,omitempty" validate:"min=0"`
	
	// LowestPrice is the lowest closing price across all available data
	// Considers only days with actual trading activity
	// Excludes zero values and invalid prices
	LowestPrice float64 `json:"lowest_price,omitempty" csv:"LowestPrice,omitempty" validate:"min=0"`
	
	// Change is the absolute price change from the previous trading day
	// Calculated as: LastPrice - PreviousPrice
	// Will be 0 if stock didn't trade on LastDate
	Change float64 `json:"change,omitempty" csv:"Change,omitempty"`
	
	// ChangePercent is the percentage change from previous trading day
	// Calculated as: ((LastPrice - PreviousPrice) / PreviousPrice) * 100
	// Will be 0 if stock didn't trade on LastDate
	// Use LastTradingStatus to distinguish between "no change" and "no trading"
	ChangePercent float64 `json:"change_percent,omitempty" csv:"ChangePercent,omitempty"`
	
	// LastTradingStatus indicates whether the stock actually traded on LastDate
	// true = stock traded (TradingStatus=true in source data)
	// false = stock did not trade (forward-filled data)
	// Used to distinguish between "0% change" and "no trading activity"
	LastTradingStatus bool `json:"last_trading_status" csv:"LastTradingStatus"`
	
	// === METADATA FIELDS (system information) ===
	
	// GeneratedAt is the timestamp when this summary was created
	// ISO 8601 format with timezone information
	GeneratedAt time.Time `json:"generated_at,omitempty" csv:"GeneratedAt,omitempty"`
	
	// DataSource indicates the source of the summary data
	// Examples: "daily_reports", "real_time", "manual"
	DataSource string `json:"data_source,omitempty" csv:"DataSource,omitempty"`
	
	// Version tracks the structure version for backward compatibility
	// Current version: "1.0"
	Version string `json:"version,omitempty" csv:"Version,omitempty"`
}

// TickerSummaryValidationRules defines validation constraints for TickerSummary fields.
// These rules ensure data integrity and consistency across the system.
var TickerSummaryValidationRules = struct {
	TickerPattern     *regexp.Regexp
	MinTickerLength   int
	MaxTickerLength   int
	MinCompanyLength  int
	MaxCompanyLength  int
	MinPrice          float64
	MaxLast10Days     int
	RequiredDateFormat string
}{
	TickerPattern:     regexp.MustCompile(`^[A-Z]{2,10}$`),
	MinTickerLength:   2,
	MaxTickerLength:   10,
	MinCompanyLength:  3,
	MaxCompanyLength:  255,
	MinPrice:          0.001,
	MaxLast10Days:     10,
	RequiredDateFormat: "2006-01-02",
}

// ValidateTickerSummary performs comprehensive validation on a TickerSummary instance.
// It checks all business rules, data constraints, and format requirements.
//
// Validation Rules:
// - Ticker: 2-10 uppercase letters only
// - CompanyName: 3-255 characters, not empty
// - LastPrice: >= 0.001 (minimum tradeable price)
// - LastDate: valid date in "2006-01-02" format
// - TradingDays: >= 0
// - Last10Days: maximum 10 elements, all positive values
//
// Returns:
//   - nil if validation passes
//   - error with detailed description if validation fails
//
// Example:
//   if err := ValidateTickerSummary(summary); err != nil {
//       return fmt.Errorf("invalid ticker summary: %w", err)
//   }
func ValidateTickerSummary(summary *TickerSummary) error {
	if summary == nil {
		return fmt.Errorf("ticker summary cannot be nil")
	}

	// Validate Ticker
	if summary.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}
	if !TickerSummaryValidationRules.TickerPattern.MatchString(summary.Ticker) {
		return fmt.Errorf("ticker '%s' must be 2-10 uppercase letters only", summary.Ticker)
	}

	// Validate CompanyName
	if summary.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}
	if len(summary.CompanyName) < TickerSummaryValidationRules.MinCompanyLength {
		return fmt.Errorf("company name must be at least %d characters", TickerSummaryValidationRules.MinCompanyLength)
	}
	if len(summary.CompanyName) > TickerSummaryValidationRules.MaxCompanyLength {
		return fmt.Errorf("company name must not exceed %d characters", TickerSummaryValidationRules.MaxCompanyLength)
	}

	// Validate LastPrice
	if summary.LastPrice < TickerSummaryValidationRules.MinPrice {
		return fmt.Errorf("last price %.6f must be at least %.3f", summary.LastPrice, TickerSummaryValidationRules.MinPrice)
	}

	// Validate LastDate format
	if summary.LastDate == "" {
		return fmt.Errorf("last date is required")
	}
	if _, err := time.Parse(TickerSummaryValidationRules.RequiredDateFormat, summary.LastDate); err != nil {
		return fmt.Errorf("last date '%s' must be in format '%s': %w", 
			summary.LastDate, TickerSummaryValidationRules.RequiredDateFormat, err)
	}

	// Validate TradingDays
	if summary.TradingDays < 0 {
		return fmt.Errorf("trading days cannot be negative: %d", summary.TradingDays)
	}

	// Validate Last10Days
	if len(summary.Last10Days) > TickerSummaryValidationRules.MaxLast10Days {
		return fmt.Errorf("last 10 days cannot have more than %d elements: got %d", 
			TickerSummaryValidationRules.MaxLast10Days, len(summary.Last10Days))
	}
	for i, price := range summary.Last10Days {
		if price < 0 {
			return fmt.Errorf("last 10 days price at index %d cannot be negative: %.6f", i, price)
		}
	}

	// Validate extended metrics (if present)
	if summary.TotalVolume < 0 {
		return fmt.Errorf("total volume cannot be negative: %d", summary.TotalVolume)
	}
	if summary.TotalValue < 0 {
		return fmt.Errorf("total value cannot be negative: %.6f", summary.TotalValue)
	}
	if summary.AveragePrice < 0 {
		return fmt.Errorf("average price cannot be negative: %.6f", summary.AveragePrice)
	}
	if summary.HighestPrice < 0 {
		return fmt.Errorf("highest price cannot be negative: %.6f", summary.HighestPrice)
	}
	if summary.LowestPrice < 0 {
		return fmt.Errorf("lowest price cannot be negative: %.6f", summary.LowestPrice)
	}

	return nil
}

// FormatLast10DaysForCSV formats the Last10Days slice as a comma-separated string
// for CSV export. This ensures consistent CSV formatting across the system.
//
// Format: "1.200,1.210,1.220,1.240,1.250"
// Precision: 3 decimal places
// Empty slice returns empty string
//
// Example:
//   prices := []float64{1.200, 1.210, 1.220}
//   csvString := summary.FormatLast10DaysForCSV() // "1.200,1.210,1.220"
func (ts *TickerSummary) FormatLast10DaysForCSV() string {
	if len(ts.Last10Days) == 0 {
		return ""
	}

	parts := make([]string, len(ts.Last10Days))
	for i, price := range ts.Last10Days {
		parts[i] = fmt.Sprintf("%.3f", price)
	}
	return strings.Join(parts, ",")
}

// ParseLast10DaysFromCSV parses a comma-separated string into the Last10Days slice.
// This is used when reading ticker summaries from CSV files.
//
// Input format: "1.200,1.210,1.220,1.240,1.250"
// Handles empty strings and malformed values gracefully
//
// Example:
//   err := summary.ParseLast10DaysFromCSV("1.200,1.210,1.220")
//   if err != nil {
//       return fmt.Errorf("parse prices: %w", err)
//   }
func (ts *TickerSummary) ParseLast10DaysFromCSV(csvString string) error {
	csvString = strings.TrimSpace(csvString)
	if csvString == "" {
		ts.Last10Days = []float64{}
		return nil
	}

	parts := strings.Split(csvString, ",")
	if len(parts) > TickerSummaryValidationRules.MaxLast10Days {
		return fmt.Errorf("too many price values: expected max %d, got %d", 
			TickerSummaryValidationRules.MaxLast10Days, len(parts))
	}

	prices := make([]float64, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue // Skip empty values
		}

		var price float64
		if _, err := fmt.Sscanf(part, "%f", &price); err != nil {
			return fmt.Errorf("invalid price at position %d: '%s': %w", i, part, err)
		}

		if price < 0 {
			return fmt.Errorf("negative price at position %d: %.6f", i, price)
		}

		prices = append(prices, price)
	}

	ts.Last10Days = prices
	return nil
}

// IsValidTicker checks if a ticker symbol meets the validation requirements.
// This is a convenience function for validating ticker symbols before creating summaries.
//
// Requirements:
// - 2-10 characters in length
// - Uppercase letters only (A-Z)
// - No spaces, numbers, or special characters
//
// Example:
//   if !IsValidTicker("BBOB") {
//       return errors.New("invalid ticker format")
//   }
func IsValidTicker(ticker string) bool {
	return TickerSummaryValidationRules.TickerPattern.MatchString(ticker)
}

// NewTickerSummary creates a new TickerSummary with required fields and validation.
// This is the recommended way to create TickerSummary instances to ensure
// proper initialization and data consistency.
//
// Parameters:
//   - ticker: Stock symbol (validated for format)
//   - companyName: Full company name
//   - lastPrice: Most recent trading price
//   - lastDate: Last trading date in "2006-01-02" format
//   - tradingDays: Count of actual trading days
//
// Returns:
//   - *TickerSummary: Initialized and validated summary
//   - error: Validation error if any parameter is invalid
//
// Example:
//   summary, err := NewTickerSummary("BBOB", "Bank of Baghdad", 1.250, "2024-01-15", 120)
//   if err != nil {
//       return fmt.Errorf("create summary: %w", err)
//   }
func NewTickerSummary(ticker, companyName string, lastPrice float64, lastDate string, tradingDays int) (*TickerSummary, error) {
	summary := &TickerSummary{
		Ticker:      strings.ToUpper(strings.TrimSpace(ticker)),
		CompanyName: strings.TrimSpace(companyName),
		LastPrice:   lastPrice,
		LastDate:    lastDate,
		TradingDays: tradingDays,
		Last10Days:  make([]float64, 0),
		GeneratedAt: time.Now(),
		Version:     "1.0",
	}

	if err := ValidateTickerSummary(summary); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return summary, nil
}

// TickerSummaryFilter represents filters for querying ticker summaries.
// This structure provides comprehensive filtering capabilities for summary data.
type TickerSummaryFilter struct {
	// Tickers filters by specific ticker symbols
	// Empty slice means no ticker filtering
	Tickers []string `json:"tickers,omitempty"`
	
	// CompanyNamePattern filters by company name using substring matching
	// Case-insensitive partial matching
	CompanyNamePattern string `json:"company_name_pattern,omitempty"`
	
	// MinLastPrice filters summaries with last price >= this value
	MinLastPrice float64 `json:"min_last_price,omitempty"`
	
	// MaxLastPrice filters summaries with last price <= this value
	MaxLastPrice float64 `json:"max_last_price,omitempty"`
	
	// MinTradingDays filters summaries with trading days >= this value
	MinTradingDays int `json:"min_trading_days,omitempty"`
	
	// MaxTradingDays filters summaries with trading days <= this value
	MaxTradingDays int `json:"max_trading_days,omitempty"`
	
	// DateFrom filters summaries with last date >= this date
	DateFrom *time.Time `json:"date_from,omitempty"`
	
	// DateTo filters summaries with last date <= this date
	DateTo *time.Time `json:"date_to,omitempty"`
	
	// MinChangePercent filters summaries with change % >= this value
	MinChangePercent *float64 `json:"min_change_percent,omitempty"`
	
	// MaxChangePercent filters summaries with change % <= this value
	MaxChangePercent *float64 `json:"max_change_percent,omitempty"`
	
	// SortBy specifies the field to sort results by
	// Valid values: "ticker", "company_name", "last_price", "last_date", "trading_days", "change_percent"
	SortBy string `json:"sort_by,omitempty"`
	
	// SortDesc specifies sort direction (true = descending, false = ascending)
	SortDesc bool `json:"sort_desc,omitempty"`
	
	// Limit specifies maximum number of results to return
	Limit int `json:"limit,omitempty"`
	
	// Offset specifies number of results to skip (for pagination)
	Offset int `json:"offset,omitempty"`
}

// TickerSummaryResponse represents a paginated response for ticker summary queries.
// This structure provides metadata along with the actual summary data.
type TickerSummaryResponse struct {
	// Summaries contains the actual ticker summary data
	Summaries []TickerSummary `json:"summaries"`
	
	// TotalCount is the total number of summaries matching the filter (before pagination)
	TotalCount int `json:"total_count"`
	
	// Page is the current page number (1-based)
	Page int `json:"page"`
	
	// PageSize is the number of summaries per page
	PageSize int `json:"page_size"`
	
	// TotalPages is the total number of pages available
	TotalPages int `json:"total_pages"`
	
	// GeneratedAt is when this response was created
	GeneratedAt time.Time `json:"generated_at"`
	
	// Filter contains the filter parameters used for this query
	Filter *TickerSummaryFilter `json:"filter,omitempty"`
}