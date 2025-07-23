package exporter

import (
	"fmt"
	"path/filepath"
	"sort"

	"isxcli/internal/parser"
)

// TickerExporter handles ticker-specific report generation
type TickerExporter struct {
	csvWriter *CSVWriter
}

// NewTickerExporter creates a new ticker report exporter
func NewTickerExporter(basePath string) *TickerExporter {
	return &TickerExporter{
		csvWriter: NewCSVWriter(basePath),
	}
}

// TickerSummary represents summary statistics for a ticker
type TickerSummary struct {
	Ticker         string
	CompanyName    string
	LastPrice      float64
	LastDate       string
	TradingDays    int
	Last10Days     string
	TotalVolume    int64
	TotalValue     float64
	AveragePrice   float64
	HighestPrice   float64
	LowestPrice    float64
}

// ExportTickerFiles generates individual CSV files for each ticker
func (t *TickerExporter) ExportTickerFiles(records []parser.TradeRecord, outputDir string) error {
	// Group records by ticker
	recordsByTicker := make(map[string][]parser.TradeRecord)
	for _, record := range records {
		recordsByTicker[record.CompanySymbol] = append(recordsByTicker[record.CompanySymbol], record)
	}
	
	// Export each ticker's data
	for ticker, tickerRecords := range recordsByTicker {
		// Sort by date (oldest to newest)
		sort.Slice(tickerRecords, func(i, j int) bool {
			return tickerRecords[i].Date.Before(tickerRecords[j].Date)
		})
		
		// Generate filename
		filename := fmt.Sprintf("%s_trading_history.csv", ticker)
		filePath := filepath.Join(outputDir, filename)
		
		// Convert records to CSV format
		var csvRecords [][]string
		for _, record := range tickerRecords {
			csvRecords = append(csvRecords, t.recordToCSVRow(record))
		}
		
		// Write CSV file
		if err := t.csvWriter.WriteSimpleCSV(filePath, t.getHeaders(), csvRecords); err != nil {
			return fmt.Errorf("failed to write ticker file for %s: %w", ticker, err)
		}
	}
	
	return nil
}

// ExportTickerSummary exports a summary CSV with statistics for all tickers
func (t *TickerExporter) ExportTickerSummary(summaries []TickerSummary, outputPath string) error {
	// Sort by ticker symbol
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})
	
	// Convert summaries to CSV format
	var csvRecords [][]string
	for _, summary := range summaries {
		csvRecords = append(csvRecords, t.summaryToCSVRow(summary))
	}
	
	// Write summary CSV
	headers := []string{
		"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days",
		"TotalVolume", "TotalValue", "AveragePrice", "HighestPrice", "LowestPrice",
	}
	
	return t.csvWriter.WriteSimpleCSV(outputPath, headers, csvRecords)
}

// GenerateTickerSummaries creates summary statistics from trade records
func (t *TickerExporter) GenerateTickerSummaries(records []parser.TradeRecord) []TickerSummary {
	// Group by ticker
	tickerData := make(map[string][]parser.TradeRecord)
	for _, record := range records {
		tickerData[record.CompanySymbol] = append(tickerData[record.CompanySymbol], record)
	}
	
	var summaries []TickerSummary
	for ticker, tickerRecords := range tickerData {
		// Sort by date
		sort.Slice(tickerRecords, func(i, j int) bool {
			return tickerRecords[i].Date.Before(tickerRecords[j].Date)
		})
		
		// Calculate statistics
		summary := TickerSummary{
			Ticker:      ticker,
			CompanyName: tickerRecords[0].CompanyName,
		}
		
		var totalValue float64
		var totalVolume int64
		var priceSum float64
		var tradingDays int
		highestPrice := 0.0
		lowestPrice := 999999999.0
		
		for _, record := range tickerRecords {
			if record.ClosePrice > 0 {
				tradingDays++
				totalVolume += record.Volume
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
		
		if tradingDays > 0 {
			summary.TradingDays = tradingDays
			summary.TotalVolume = totalVolume
			summary.TotalValue = totalValue
			summary.AveragePrice = priceSum / float64(tradingDays)
			summary.HighestPrice = highestPrice
			summary.LowestPrice = lowestPrice
			
			// Get last trading info
			for i := len(tickerRecords) - 1; i >= 0; i-- {
				if tickerRecords[i].ClosePrice > 0 {
					summary.LastPrice = tickerRecords[i].ClosePrice
					summary.LastDate = tickerRecords[i].Date.Format("2006-01-02")
					break
				}
			}
			
			// Calculate last 10 trading days
			var last10Count int
			for i := len(tickerRecords) - 1; i >= 0 && last10Count < 10; i-- {
				if tickerRecords[i].ClosePrice > 0 {
					last10Count++
				}
			}
			summary.Last10Days = fmt.Sprintf("%d/10", last10Count)
		}
		
		summaries = append(summaries, summary)
	}
	
	return summaries
}

// getHeaders returns the CSV headers for ticker trade records
// Using the same format as daily CSV for consistency
func (t *TickerExporter) getHeaders() []string {
	return []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
}

// recordToCSVRow converts a trade record to a ticker CSV row
func (t *TickerExporter) recordToCSVRow(record parser.TradeRecord) []string {
	return []string{
		record.Date.Format("2006-01-02"),
		record.CompanyName,
		record.CompanySymbol,
		formatFloat(record.OpenPrice),
		formatFloat(record.HighPrice),
		formatFloat(record.LowPrice),
		formatFloat(record.AveragePrice),
		formatFloat(record.PrevAveragePrice),
		formatFloat(record.ClosePrice),
		formatFloat(record.PrevClosePrice),
		formatFloat(record.Change),
		formatFloat(record.ChangePercent),
		formatInt(record.NumTrades),
		formatInt(record.Volume),
		formatFloat(record.Value),
		formatBool(record.TradingStatus),
	}
}

// summaryToCSVRow converts a ticker summary to a CSV row
func (t *TickerExporter) summaryToCSVRow(summary TickerSummary) []string {
	return []string{
		summary.Ticker,
		summary.CompanyName,
		formatFloat(summary.LastPrice),
		summary.LastDate,
		fmt.Sprintf("%d", summary.TradingDays),
		summary.Last10Days,
		formatInt(summary.TotalVolume),
		formatFloat(summary.TotalValue),
		formatFloat(summary.AveragePrice),
		formatFloat(summary.HighestPrice),
		formatFloat(summary.LowestPrice),
	}
}