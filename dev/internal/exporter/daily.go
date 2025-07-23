package exporter

import (
	"fmt"
	"path/filepath"
	"sort"

	"isxcli/internal/parser"
)

// DailyExporter handles daily report generation
type DailyExporter struct {
	csvWriter *CSVWriter
}

// NewDailyExporter creates a new daily report exporter
func NewDailyExporter(basePath string) *DailyExporter {
	return &DailyExporter{
		csvWriter: NewCSVWriter(basePath),
	}
}

// ExportDailyReports generates daily CSV files grouped by date
func (d *DailyExporter) ExportDailyReports(records []parser.TradeRecord, outputDir string) error {
	// Group records by date
	recordsByDate := make(map[string][]parser.TradeRecord)
	for _, record := range records {
		dateKey := record.Date.Format("2006_01_02")
		recordsByDate[dateKey] = append(recordsByDate[dateKey], record)
	}
	
	// Get sorted dates
	var dates []string
	for date := range recordsByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	
	// Export each day's data
	for _, dateKey := range dates {
		dayRecords := recordsByDate[dateKey]
		
		// Sort by symbol for consistent output
		sort.Slice(dayRecords, func(i, j int) bool {
			return dayRecords[i].CompanySymbol < dayRecords[j].CompanySymbol
		})
		
		// Generate filename
		filename := fmt.Sprintf("isx_daily_%s.csv", dateKey)
		filePath := filepath.Join(outputDir, filename)
		
		// Convert records to CSV format
		var csvRecords [][]string
		for _, record := range dayRecords {
			csvRecords = append(csvRecords, d.recordToCSVRow(record))
		}
		
		// Write CSV file
		if err := d.csvWriter.WriteSimpleCSV(filePath, d.getHeaders(), csvRecords); err != nil {
			return fmt.Errorf("failed to write daily report for %s: %w", dateKey, err)
		}
	}
	
	return nil
}

// ExportCombinedData exports all records to a single combined CSV file
func (d *DailyExporter) ExportCombinedData(records []parser.TradeRecord, outputPath string) error {
	// Sort records by date and symbol
	sort.Slice(records, func(i, j int) bool {
		if records[i].Date.Equal(records[j].Date) {
			return records[i].CompanySymbol < records[j].CompanySymbol
		}
		return records[i].Date.Before(records[j].Date)
	})
	
	// Convert all records to CSV format
	var csvRecords [][]string
	for _, record := range records {
		csvRecords = append(csvRecords, d.recordToCSVRow(record))
	}
	
	// Write combined CSV file without BOM for better compatibility with analysis tools
	return d.csvWriter.WriteCSV(outputPath, WriteOptions{
		Headers: d.getHeaders(),
		Records: csvRecords,
		Append: false,
		BOMPrefix: false, // No BOM for combined CSV to avoid parsing issues
	})
}

// ExportDailyReportsStreaming exports daily reports using streaming for large datasets
func (d *DailyExporter) ExportDailyReportsStreaming(records []parser.TradeRecord, outputDir string, existingDates map[string]bool) error {
	// Group records by date
	recordsByDate := make(map[string][]parser.TradeRecord)
	for _, record := range records {
		dateKey := record.Date.Format("2006_01_02")
		recordsByDate[dateKey] = append(recordsByDate[dateKey], record)
	}
	
	// Process each date
	for dateKey, dayRecords := range recordsByDate {
		// Skip if already exists
		if existingDates != nil && existingDates[dateKey] {
			continue
		}
		
		// Sort by symbol
		sort.Slice(dayRecords, func(i, j int) bool {
			return dayRecords[i].CompanySymbol < dayRecords[j].CompanySymbol
		})
		
		// Generate filename
		filename := fmt.Sprintf("isx_daily_%s.csv", dateKey)
		filePath := filepath.Join(outputDir, filename)
		
		// Create stream writer
		stream, err := d.csvWriter.CreateStreamWriter(filePath, d.getHeaders())
		if err != nil {
			return fmt.Errorf("failed to create stream writer for %s: %w", dateKey, err)
		}
		
		// Write records
		for _, record := range dayRecords {
			if err := stream.WriteRecord(d.recordToCSVRow(record)); err != nil {
				stream.Close()
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
		
		// Close stream
		if err := stream.Close(); err != nil {
			return fmt.Errorf("failed to close stream for %s: %w", dateKey, err)
		}
	}
	
	return nil
}

// getHeaders returns the CSV headers for trade records
func (d *DailyExporter) getHeaders() []string {
	return []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
}

// recordToCSVRow converts a trade record to a CSV row
func (d *DailyExporter) recordToCSVRow(record parser.TradeRecord) []string {
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

