// Package exporter provides CSV export functionality for the ISX Daily Reports Scrapper.
//
// This package contains three main components:
//
// CSVWriter: Core CSV writing functionality with support for headers, streaming,
// and UTF-8 BOM for Excel compatibility.
//
// DailyExporter: Handles generation of daily report CSV files grouped by date,
// as well as combined data exports.
//
// TickerExporter: Manages ticker-specific exports including individual ticker
// history files and summary statistics.
//
// Example usage:
//
//	// Create a daily exporter
//	dailyExporter := exporter.NewDailyExporter("/path/to/base")
//	
//	// Export daily reports
//	err := dailyExporter.ExportDailyReports(records, "data/reports")
//	
//	// Create a ticker exporter
//	tickerExporter := exporter.NewTickerExporter("/path/to/base")
//	
//	// Generate and export ticker summaries
//	summaries := tickerExporter.GenerateTickerSummaries(records)
//	err = tickerExporter.ExportTickerSummary(summaries, "data/reports/ticker_summary.csv")
package exporter