package dataprocessing

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// IntegrationExample demonstrates how to use the new Summarizer to replace
// the existing duplicate ticker summary implementations.
// This shows the migration path from the old implementations to the SSOT.
type IntegrationExample struct {
	summarizer *Summarizer
	logger     *slog.Logger
}

// NewIntegrationExample creates a new integration example.
func NewIntegrationExample(logger *slog.Logger) *IntegrationExample {
	if logger == nil {
		logger = slog.Default()
	}

	// Create summarizer with extended metrics for comprehensive output
	config := ExtendedSummarizerConfig()
	summarizer := NewSummarizer(logger, config)

	return &IntegrationExample{
		summarizer: summarizer,
		logger:     logger,
	}
}

// GenerateTickerSummaryFromCombinedCSV demonstrates replacing the logic in
// cmd/processor/main.go (lines 718-885) with the new SSOT implementation.
// This method reads a combined CSV and generates ticker summaries using
// the correct last trading date logic.
func (ie *IntegrationExample) GenerateTickerSummaryFromCombinedCSV(ctx context.Context, combinedFile, outputDir string) error {
	ie.logger.InfoContext(ctx, "generating ticker summary from combined CSV using SSOT",
		slog.String("combined_file", combinedFile),
		slog.String("output_dir", outputDir))

	// Ensure output directory exists
	summaryDir := filepath.Join(outputDir, "summary", "ticker")
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return errors.NewStorageError("failed to create summary directory", err)
	}

	// Read and parse the combined CSV file
	records, err := ie.readCombinedCSV(combinedFile)
	if err != nil {
		return fmt.Errorf("read combined CSV: %w", err)
	}

	// Generate summaries using the SSOT implementation
	summaries, err := ie.summarizer.GenerateFromRecords(ctx, records)
	if err != nil {
		return fmt.Errorf("generate summaries: %w", err)
	}

	// Write both CSV and JSON outputs
	csvPath := filepath.Join(summaryDir, "ticker_summary.csv")
	jsonPath := filepath.Join(summaryDir, "ticker_summary.json")

	if err := ie.summarizer.WriteCSV(ctx, csvPath, summaries); err != nil {
		return fmt.Errorf("write CSV summary: %w", err)
	}

	if err := ie.summarizer.WriteJSON(ctx, jsonPath, summaries); err != nil {
		return fmt.Errorf("write JSON summary: %w", err)
	}

	ie.logger.InfoContext(ctx, "successfully generated ticker summary using SSOT",
		slog.String("csv_path", csvPath),
		slog.String("json_path", jsonPath),
		slog.Int("ticker_count", len(summaries)))

	return nil
}

// GenerateTickerSummaryFromTradeRecords demonstrates how to use the SSOT
// when you already have domain.TradeRecord slices (e.g., from database queries).
func (ie *IntegrationExample) GenerateTickerSummaryFromTradeRecords(ctx context.Context, records []domain.TradeRecord, outputPath string) error {
	ie.logger.InfoContext(ctx, "generating ticker summary from trade records using SSOT",
		slog.String("output_path", outputPath),
		slog.Int("record_count", len(records)))

	// Generate summaries using the SSOT implementation
	summaries, err := ie.summarizer.GenerateFromRecords(ctx, records)
	if err != nil {
		return fmt.Errorf("generate summaries: %w", err)
	}

	// Write CSV output
	if err := ie.summarizer.WriteCSV(ctx, outputPath, summaries); err != nil {
		return fmt.Errorf("write CSV summary: %w", err)
	}

	ie.logger.InfoContext(ctx, "successfully generated ticker summary from trade records",
		slog.String("output_path", outputPath),
		slog.Int("ticker_count", len(summaries)))

	return nil
}

// readCombinedCSV reads and parses a combined CSV file into TradeRecord slices.
// This replaces the manual CSV parsing logic from the old implementations.
func (ie *IntegrationExample) readCombinedCSV(filePath string) ([]domain.TradeRecord, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("combined CSV file not found: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open combined file: %w", err)
	}
	defer file.Close()

	// Read file content to handle BOM
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Remove BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}

	// Create CSV reader from cleaned content
	reader := csv.NewReader(strings.NewReader(string(content)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	// Parse header to find column indices
	header := records[0]
	columnMap := ie.findColumnIndices(header)

	// Convert CSV rows to TradeRecord structs
	var tradeRecords []domain.TradeRecord
	for i := 1; i < len(records); i++ {
		record, err := ie.parseCSVRowToTradeRecord(records[i], columnMap)
		if err != nil {
			ie.logger.Warn("failed to parse CSV row, skipping",
				slog.Int("row", i),
				slog.String("error", err.Error()))
			continue
		}
		tradeRecords = append(tradeRecords, record)
	}

	return tradeRecords, nil
}

// columnMapping holds the indices of CSV columns
type columnMapping struct {
	ticker        int
	company       int
	date          int
	open          int
	high          int
	low           int
	close         int
	avgPrice      int
	prevAvgPrice  int
	prevClose     int
	change        int
	changePercent int
	numTrades     int
	volume        int
	value         int
	tradingStatus int
}

// findColumnIndices finds the indices of required columns in the CSV header.
// This handles various column name formats used in ISX CSV files.
func (ie *IntegrationExample) findColumnIndices(header []string) columnMapping {
	mapping := columnMapping{
		ticker: -1, company: -1, date: -1, open: -1, high: -1, low: -1,
		close: -1, avgPrice: -1, prevAvgPrice: -1, prevClose: -1,
		change: -1, changePercent: -1, numTrades: -1, volume: -1,
		value: -1, tradingStatus: -1,
	}

	for i, col := range header {
		// Clean and normalize column name
		cleanCol := strings.TrimSpace(col)
		cleanCol = strings.TrimPrefix(cleanCol, "\ufeff") // Remove BOM
		lowerCol := strings.ToLower(cleanCol)

		// Map to appropriate fields
		switch {
		case lowerCol == "symbol" || lowerCol == "ticker" || lowerCol == "company_symbol":
			mapping.ticker = i
		case lowerCol == "companyname" || lowerCol == "company_name" || lowerCol == "company" || lowerCol == "name":
			mapping.company = i
		case lowerCol == "date":
			mapping.date = i
		case lowerCol == "openprice" || lowerCol == "open_price" || lowerCol == "open":
			mapping.open = i
		case lowerCol == "highprice" || lowerCol == "high_price" || lowerCol == "high":
			mapping.high = i
		case lowerCol == "lowprice" || lowerCol == "low_price" || lowerCol == "low":
			mapping.low = i
		case lowerCol == "closeprice" || lowerCol == "close_price" || lowerCol == "close":
			mapping.close = i
		case lowerCol == "averageprice" || lowerCol == "average_price" || lowerCol == "avgprice" || lowerCol == "avg_price":
			mapping.avgPrice = i
		case lowerCol == "prevaverageprice" || lowerCol == "prev_average_price" || lowerCol == "prevavgprice" || lowerCol == "prev_avg_price":
			mapping.prevAvgPrice = i
		case lowerCol == "prevcloseprice" || lowerCol == "prev_close_price" || lowerCol == "prevclose" || lowerCol == "prev_close":
			mapping.prevClose = i
		case lowerCol == "change":
			mapping.change = i
		case lowerCol == "changepercent" || lowerCol == "change_percent" || lowerCol == "change%":
			mapping.changePercent = i
		case lowerCol == "numtrades" || lowerCol == "num_trades" || lowerCol == "trades":
			mapping.numTrades = i
		case lowerCol == "volume":
			mapping.volume = i
		case lowerCol == "value":
			mapping.value = i
		case lowerCol == "tradingstatus" || lowerCol == "trading_status" || lowerCol == "status":
			mapping.tradingStatus = i
		}
	}

	return mapping
}

// parseCSVRowToTradeRecord converts a CSV row to a TradeRecord using the column mapping.
func (ie *IntegrationExample) parseCSVRowToTradeRecord(row []string, mapping columnMapping) (domain.TradeRecord, error) {
	if mapping.ticker == -1 || mapping.company == -1 || mapping.date == -1 || mapping.close == -1 {
		return domain.TradeRecord{}, fmt.Errorf("required columns missing")
	}

	// Extract required fields
	ticker := strings.TrimSpace(row[mapping.ticker])
	company := strings.TrimSpace(row[mapping.company])
	dateStr := strings.TrimSpace(row[mapping.date])

	if ticker == "" || company == "" || dateStr == "" {
		return domain.TradeRecord{}, fmt.Errorf("empty required field")
	}

	// Parse date (try multiple formats)
	date, err := ie.parseDate(dateStr)
	if err != nil {
		return domain.TradeRecord{}, fmt.Errorf("parse date: %w", err)
	}

	// Parse prices and volumes with safe defaults
	record := domain.TradeRecord{
		CompanySymbol: ticker,
		CompanyName:   company,
		Date:          date,
		ClosePrice:    ie.parseFloat(row, mapping.close),
		OpenPrice:     ie.parseFloat(row, mapping.open),
		HighPrice:     ie.parseFloat(row, mapping.high),
		LowPrice:      ie.parseFloat(row, mapping.low),
		AveragePrice:  ie.parseFloat(row, mapping.avgPrice),
		PrevAveragePrice: ie.parseFloat(row, mapping.prevAvgPrice),
		PrevClosePrice:   ie.parseFloat(row, mapping.prevClose),
		Change:           ie.parseFloat(row, mapping.change),
		ChangePercent:    ie.parseFloat(row, mapping.changePercent),
		NumTrades:        ie.parseInt(row, mapping.numTrades),
		Volume:           ie.parseInt(row, mapping.volume),
		Value:            ie.parseFloat(row, mapping.value),
		TradingStatus:    ie.parseBool(row, mapping.tradingStatus),
	}

	return record, nil
}

// parseDate parses date strings in various formats used by ISX files.
func (ie *IntegrationExample) parseDate(dateStr string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"02-01-2006",
		"01-02-2006",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseFloat safely parses a float from CSV row with default fallback.
func (ie *IntegrationExample) parseFloat(row []string, index int) float64 {
	if index == -1 || index >= len(row) {
		return 0.0
	}
	
	value := strings.TrimSpace(row[index])
	if value == "" || value == "-" || value == "N/A" {
		return 0.0
	}
	
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	
	return parsed
}

// parseInt safely parses an int from CSV row with default fallback.
func (ie *IntegrationExample) parseInt(row []string, index int) int64 {
	if index == -1 || index >= len(row) {
		return 0
	}
	
	value := strings.TrimSpace(row[index])
	if value == "" || value == "-" || value == "N/A" {
		return 0
	}
	
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	
	return parsed
}

// parseBool safely parses a boolean from CSV row with default fallback.
func (ie *IntegrationExample) parseBool(row []string, index int) bool {
	if index == -1 || index >= len(row) {
		return false
	}
	
	value := strings.TrimSpace(strings.ToLower(row[index]))
	return value == "true" || value == "1" || value == "yes" || value == "active"
}

// MigrationGuide provides documentation for replacing the old implementations.
func MigrationGuide() string {
	return `
MIGRATION GUIDE: Replacing Duplicate Ticker Summary Implementations
==================================================================

OLD IMPLEMENTATIONS TO REPLACE:
1. cmd/processor/main.go (lines 718-885) - generateTickerSummary()
2. internal/dataprocessing/analytics.go - SummaryGenerator methods
3. internal/exporter/ticker.go - GenerateTickerSummaries()

NEW SINGLE SOURCE OF TRUTH:
- internal/dataprocessing/summarizer.go - Summarizer struct

MIGRATION STEPS:

1. Replace cmd/processor/main.go logic:
   OLD: generateTickerSummary(outDir string)
   NEW: Use IntegrationExample.GenerateTickerSummaryFromCombinedCSV()

2. Replace analytics.go logic:
   OLD: SummaryGenerator.GenerateFromCombinedCSV()
   NEW: Use Summarizer.GenerateFromRecords() + CSV/JSON writers

3. Replace exporter/ticker.go logic:
   OLD: TickerExporter.GenerateTickerSummaries()
   NEW: Use Summarizer.GenerateFromRecords()

KEY IMPROVEMENTS:
- Correctly checks TradingStatus field for last trading date
- Falls back to volume/numTrades if TradingStatus unavailable
- Counts only actual trading days, not calendar days
- Provides both CSV and JSON output
- Follows CLAUDE.md standards (slog, RFC 7807 errors, context)
- Comprehensive test coverage

EXAMPLE USAGE:
  ctx := context.Background()
  logger := slog.Default()
  config := ExtendedSummarizerConfig()
  summarizer := NewSummarizer(logger, config)
  
  summaries, err := summarizer.GenerateFromRecords(ctx, records)
  if err != nil {
      return err
  }
  
  err = summarizer.WriteCSV(ctx, "ticker_summary.csv", summaries)
  err = summarizer.WriteJSON(ctx, "ticker_summary.json", summaries)

TESTING:
  cd api && go test ./internal/dataprocessing -v -run TestSummarizer
`
}