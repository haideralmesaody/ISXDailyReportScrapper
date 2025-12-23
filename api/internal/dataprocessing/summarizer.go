package dataprocessing

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// Summarizer provides Single Source of Truth (SSOT) for ticker summary generation.
// It consolidates all ticker summary logic that was previously duplicated across
// cmd/processor/main.go, internal/dataprocessing/analytics.go, and internal/exporter/ticker.go.
type Summarizer struct {
	logger            *slog.Logger
	includeExtendedMetrics bool
	maxLast10Days     int
	dateFormat        string
}

// SummarizerConfig holds configuration options for the Summarizer.
type SummarizerConfig struct {
	IncludeExtendedMetrics bool // Include daily/weekly/monthly change percentages
	MaxLast10Days         int  // Maximum number of last trading days to track
	DateFormat            string // Format for date strings in output
}

// TickerSummary represents comprehensive summary information for a ticker.
// This is the Single Source of Truth data structure for ticker summaries.
type TickerSummary struct {
	// Core fields (always included)
	Ticker      string    `json:"ticker" csv:"Ticker"`
	CompanyName string    `json:"company_name" csv:"CompanyName"`
	LastPrice   float64   `json:"last_price" csv:"LastPrice"`
	LastDate    string    `json:"last_date" csv:"LastDate"`
	TradingDays int       `json:"trading_days" csv:"TradingDays"`
	Last10Days  []float64 `json:"last_10_days" csv:"Last10Days"`
	
	// Change fields for accurate daily change display
	Change            float64 `json:"change" csv:"Change"`
	ChangePercent     float64 `json:"change_percent" csv:"ChangePercent"`
	LastTradingStatus bool    `json:"last_trading_status" csv:"LastTradingStatus"`

	// Extended metrics (optional, enabled via config)
	DailyChangePercent   float64 `json:"daily_change_percent,omitempty"`
	WeeklyChangePercent  float64 `json:"weekly_change_percent,omitempty"`
	MonthlyChangePercent float64 `json:"monthly_change_percent,omitempty"`
	DailyVolume          int64   `json:"daily_volume,omitempty"`
	DailyValue           float64 `json:"daily_value,omitempty"`
	PreviousClose        float64 `json:"previous_close,omitempty"`
	High52Week           float64 `json:"high_52_week,omitempty"`
	Low52Week            float64 `json:"low_52_week,omitempty"`
	TotalVolume          int64   `json:"total_volume,omitempty"`
	TotalValue           float64 `json:"total_value,omitempty"`
	AveragePrice         float64 `json:"average_price,omitempty"`
	HighestPrice         float64 `json:"highest_price,omitempty"`
	LowestPrice          float64 `json:"lowest_price,omitempty"`
}

// NewSummarizer creates a new ticker summarizer with the given configuration.
// This is the primary constructor for creating SSOT summarizer instances.
func NewSummarizer(logger *slog.Logger, config SummarizerConfig) *Summarizer {
	if logger == nil {
		logger = slog.Default()
	}

	// Set default configuration values
	if config.MaxLast10Days <= 0 {
		config.MaxLast10Days = 10
	}
	if config.DateFormat == "" {
		config.DateFormat = "2006-01-02"
	}

	return &Summarizer{
		logger:                logger,
		includeExtendedMetrics: config.IncludeExtendedMetrics,
		maxLast10Days:         config.MaxLast10Days,
		dateFormat:            config.DateFormat,
	}
}

// GenerateFromRecords is the Single Source of Truth method for generating ticker summaries
// from TradeRecord slices. This method implements the correct logic for determining
// last trading dates using TradingStatus field with fallback to volume/numTrades.
func (s *Summarizer) GenerateFromRecords(ctx context.Context, records []domain.TradeRecord) ([]TickerSummary, error) {
	s.logger.InfoContext(ctx, "generating ticker summaries from trade records",
		slog.Int("record_count", len(records)))

	if len(records) == 0 {
		return []TickerSummary{}, nil
	}

	// Group records by ticker symbol
	tickerData := s.groupRecordsByTicker(records)

	// Generate summaries for each ticker
	summaries := make([]TickerSummary, 0, len(tickerData))
	for ticker, tickerRecords := range tickerData {
		summary, err := s.generateTickerSummary(ctx, ticker, tickerRecords)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to generate summary for ticker",
				slog.String("ticker", ticker),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("generate summary for ticker %s: %w", ticker, err)
		}
		summaries = append(summaries, summary)
	}

	// Sort summaries by ticker symbol for consistent output
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})

	s.logger.InfoContext(ctx, "successfully generated ticker summaries",
		slog.Int("ticker_count", len(summaries)))

	return summaries, nil
}

// WriteCSV writes ticker summaries to a CSV file using the standard format.
// This method ensures consistent CSV output format across all usages.
func (s *Summarizer) WriteCSV(ctx context.Context, path string, summaries []TickerSummary) error {
	s.logger.InfoContext(ctx, "writing ticker summaries to CSV",
		slog.String("path", path),
		slog.Int("summary_count", len(summaries)))

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.NewStorageError("failed to create directory for CSV output", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.NewStorageError("failed to create CSV file for ticker summaries", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days"}
	if s.includeExtendedMetrics {
		header = append(header, "TotalVolume", "TotalValue", "AveragePrice", "HighestPrice", "LowestPrice")
	}
	// Always include change fields for frontend display
	header = append(header, "Change", "ChangePercent", "LastTradingStatus")

	if err := writer.Write(header); err != nil {
		return errors.NewStorageError("failed to write CSV header row", err)
	}

	// Write data rows
	for _, summary := range summaries {
		row := []string{
			summary.Ticker,
			summary.CompanyName,
			fmt.Sprintf("%.3f", summary.LastPrice),
			summary.LastDate,
			fmt.Sprintf("%d", summary.TradingDays),
			s.formatLast10Days(summary.Last10Days),
		}

		if s.includeExtendedMetrics {
			row = append(row,
				fmt.Sprintf("%d", summary.TotalVolume),
				fmt.Sprintf("%.3f", summary.TotalValue),
				fmt.Sprintf("%.3f", summary.AveragePrice),
				fmt.Sprintf("%.3f", summary.HighestPrice),
				fmt.Sprintf("%.3f", summary.LowestPrice),
			)
		}
		
		// Always include change fields for frontend display
		row = append(row,
			fmt.Sprintf("%.3f", summary.Change),
			fmt.Sprintf("%.2f", summary.ChangePercent),
			fmt.Sprintf("%t", summary.LastTradingStatus),
		)

		if err := writer.Write(row); err != nil {
			return errors.NewStorageError("failed to write CSV data row", err)
		}
	}

	s.logger.InfoContext(ctx, "successfully wrote ticker summaries to CSV",
		slog.String("path", path))

	return nil
}

// WriteJSON writes ticker summaries to a JSON file with metadata.
// This method provides structured JSON output compatible with web interfaces.
func (s *Summarizer) WriteJSON(ctx context.Context, path string, summaries []TickerSummary) error {
	s.logger.InfoContext(ctx, "writing ticker summaries to JSON",
		slog.String("path", path),
		slog.Int("summary_count", len(summaries)))

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.NewStorageError("failed to create directory for JSON output", err)
	}

	// Create JSON structure with metadata
	jsonData := map[string]interface{}{
		"tickers":      summaries,
		"count":        len(summaries),
		"generated_at": time.Now().Format(time.RFC3339),
		"format":       "ticker_summary_v1",
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.NewStorageError("failed to create JSON file for ticker summaries", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(jsonData); err != nil {
		return errors.NewStorageError("failed to encode ticker summaries to JSON", err)
	}

	s.logger.InfoContext(ctx, "successfully wrote ticker summaries to JSON",
		slog.String("path", path))

	return nil
}

// groupRecordsByTicker groups trade records by company symbol.
func (s *Summarizer) groupRecordsByTicker(records []domain.TradeRecord) map[string][]domain.TradeRecord {
	tickerData := make(map[string][]domain.TradeRecord)

	for _, record := range records {
		ticker := strings.TrimSpace(record.CompanySymbol)
		if ticker == "" {
			continue // Skip records without ticker symbol
		}
		tickerData[ticker] = append(tickerData[ticker], record)
	}

	return tickerData
}

// generateTickerSummary generates a comprehensive summary for a single ticker.
// This implements the correct logic for determining last trading date using TradingStatus.
func (s *Summarizer) generateTickerSummary(ctx context.Context, ticker string, records []domain.TradeRecord) (TickerSummary, error) {
	if len(records) == 0 {
		return TickerSummary{}, fmt.Errorf("no records provided for ticker %s", ticker)
	}

	// Sort records by date (oldest to newest)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.Before(records[j].Date)
	})

	// Initialize summary with basic info
	summary := TickerSummary{
		Ticker:      ticker,
		CompanyName: records[0].CompanyName, // Use company name from first record
		Last10Days:  make([]float64, 0, s.maxLast10Days),
	}

	// Find the last record with actual trading activity (SSOT logic)
	lastTradingRecord, lastTradingIndex := s.findLastTradingRecord(records)
	if lastTradingRecord != nil {
		summary.LastPrice = lastTradingRecord.ClosePrice
		summary.LastDate = lastTradingRecord.Date.Format(s.dateFormat)
		summary.Change = lastTradingRecord.Change
		summary.ChangePercent = lastTradingRecord.ChangePercent
		summary.LastTradingStatus = lastTradingRecord.TradingStatus
	} else {
		// Fallback to last record if no trading activity found
		lastRecord := records[len(records)-1]
		summary.LastPrice = lastRecord.ClosePrice
		summary.LastDate = lastRecord.Date.Format(s.dateFormat)
		summary.Change = 0
		summary.ChangePercent = 0
		summary.LastTradingStatus = false
		lastTradingIndex = len(records) - 1
	}

	// Count actual trading days and collect last N trading prices
	summary.TradingDays = s.countTradingDays(records)
	summary.Last10Days = s.getLastTradingPrices(records, s.maxLast10Days)

	// Calculate extended metrics if enabled
	if s.includeExtendedMetrics {
		s.calculateExtendedMetrics(&summary, records, lastTradingIndex)
	}

	return summary, nil
}

// findLastTradingRecord finds the last record with actual trading activity.
// This implements the Single Source of Truth logic for determining last trading date:
// 1. Check TradingStatus field first (most reliable)
// 2. Fallback to volume > 0 or numTrades > 0 if TradingStatus unavailable
func (s *Summarizer) findLastTradingRecord(records []domain.TradeRecord) (*domain.TradeRecord, int) {
	for i := len(records) - 1; i >= 0; i-- {
		record := &records[i]

		// Primary check: TradingStatus field (most reliable indicator)
		if record.TradingStatus {
			return record, i
		}

		// Fallback check: volume or number of trades (for compatibility)
		if record.Volume > 0 || record.NumTrades > 0 {
			return record, i
		}
	}

	return nil, -1 // No trading activity found
}

// countTradingDays counts the number of actual trading days (not just calendar days).
func (s *Summarizer) countTradingDays(records []domain.TradeRecord) int {
	count := 0
	for _, record := range records {
		// Count days with actual trading activity
		if record.TradingStatus || record.Volume > 0 || record.NumTrades > 0 {
			count++
		}
	}
	return count
}

// getLastTradingPrices gets the last N actual trading day prices in chronological order.
func (s *Summarizer) getLastTradingPrices(records []domain.TradeRecord, maxDays int) []float64 {
	prices := make([]float64, 0, maxDays)

	// Go backwards from the most recent date to find actual trading days
	for i := len(records) - 1; i >= 0 && len(prices) < maxDays; i-- {
		record := &records[i]

		// Only include days with actual trading activity
		if record.TradingStatus || record.Volume > 0 || record.NumTrades > 0 {
			// Prepend to maintain chronological order (oldest to newest)
			prices = append([]float64{record.ClosePrice}, prices...)
		}
	}

	return prices
}

// calculateExtendedMetrics calculates additional metrics for the summary.
func (s *Summarizer) calculateExtendedMetrics(summary *TickerSummary, records []domain.TradeRecord, lastTradingIndex int) {
	// Calculate percentage changes
	summary.DailyChangePercent = s.calculateDailyChangePercent(summary.Last10Days)
	summary.WeeklyChangePercent = s.calculatePeriodChangePercent(summary.Last10Days, 7)
	summary.MonthlyChangePercent = s.calculatePeriodChangePercent(summary.Last10Days, 30)
	summary.PreviousClose = s.getPreviousClose(summary.Last10Days)

	// Calculate 52-week high/low
	summary.High52Week, summary.Low52Week = s.calculate52WeekHighLow(records)

	// Calculate volume and value metrics
	if lastTradingIndex >= 0 {
		lastRecord := &records[lastTradingIndex]
		summary.DailyVolume = lastRecord.Volume
		summary.DailyValue = lastRecord.Value
	}

	// Calculate aggregate statistics
	var totalValue, totalVolume, priceSum float64
	var tradingDayCount int
	var highestPrice, lowestPrice float64 = 0.0, 999999999.0

	for _, record := range records {
		if record.TradingStatus || record.Volume > 0 || record.NumTrades > 0 {
			tradingDayCount++
			totalVolume += float64(record.Volume)
			totalValue += record.Value
			priceSum += record.ClosePrice

			if record.HighPrice > highestPrice {
				highestPrice = record.HighPrice
			}
			if record.LowPrice < lowestPrice && record.LowPrice > 0 {
				lowestPrice = record.LowPrice
			}
		}
	}

	summary.TotalVolume = int64(totalVolume)
	summary.TotalValue = totalValue
	summary.HighestPrice = highestPrice
	summary.LowestPrice = lowestPrice

	if tradingDayCount > 0 {
		summary.AveragePrice = priceSum / float64(tradingDayCount)
	}

	if lowestPrice == 999999999.0 {
		summary.LowestPrice = 0.0
	}
}

// calculateDailyChangePercent calculates the daily percentage change.
func (s *Summarizer) calculateDailyChangePercent(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}
	current := prices[len(prices)-1]
	previous := prices[len(prices)-2]
	if previous == 0 {
		return 0.0
	}
	return ((current - previous) / previous) * 100
}

// calculatePeriodChangePercent calculates percentage change over a period.
func (s *Summarizer) calculatePeriodChangePercent(prices []float64, days int) float64 {
	if len(prices) < 2 {
		return 0.0
	}

	current := prices[len(prices)-1]

	// Get price from 'days' ago, or earliest available
	pastIndex := len(prices) - 1 - days
	if pastIndex < 0 {
		pastIndex = 0
	}

	past := prices[pastIndex]
	if past == 0 {
		return 0.0
	}

	return ((current - past) / past) * 100
}

// getPreviousClose gets the previous trading day's close price.
func (s *Summarizer) getPreviousClose(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}
	return prices[len(prices)-2]
}

// calculate52WeekHighLow calculates 52-week high and low from all available data.
func (s *Summarizer) calculate52WeekHighLow(records []domain.TradeRecord) (float64, float64) {
	if len(records) == 0 {
		return 0.0, 0.0
	}

	var high, low float64 = 0.0, 999999999.0

	// Look at up to 252 trading days (roughly 52 weeks)
	startIndex := 0
	if len(records) > 252 {
		startIndex = len(records) - 252
	}

	for i := startIndex; i < len(records); i++ {
		record := &records[i]
		
		// Only consider records with trading activity
		if record.TradingStatus || record.Volume > 0 || record.NumTrades > 0 {
			if record.HighPrice > high {
				high = record.HighPrice
			}
			if record.LowPrice < low && record.LowPrice > 0 {
				low = record.LowPrice
			}
		}
	}

	if low == 999999999.0 {
		low = 0.0
	}

	return high, low
}

// formatLast10Days formats the last 10 days prices as a comma-separated string.
func (s *Summarizer) formatLast10Days(prices []float64) string {
	if len(prices) == 0 {
		return ""
	}

	parts := make([]string, len(prices))
	for i, price := range prices {
		parts[i] = fmt.Sprintf("%.3f", price)
	}
	return strings.Join(parts, ",")
}

// DefaultSummarizerConfig returns a default configuration for typical use cases.
func DefaultSummarizerConfig() SummarizerConfig {
	return SummarizerConfig{
		IncludeExtendedMetrics: false,
		MaxLast10Days:         10,
		DateFormat:            "2006-01-02",
	}
}

// ExtendedSummarizerConfig returns a configuration with all extended metrics enabled.
func ExtendedSummarizerConfig() SummarizerConfig {
	return SummarizerConfig{
		IncludeExtendedMetrics: true,
		MaxLast10Days:         10,
		DateFormat:            "2006-01-02",
	}
}