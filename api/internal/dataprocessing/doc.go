// Package dataprocessing provides comprehensive data processing capabilities for ISX daily reports.
// It consolidates parsing, processing, and analysis functionality into a cohesive package
// that handles the complete data lifecycle from Excel ingestion to analytical insights.
//
// # Architecture
//
// The package is organized into three main components:
//
// 1. Parser: Reads ISX Excel files and extracts trading data
// 2. Processor: Applies transformations like forward-fill for missing data
// 3. Analytics: Generates summaries and statistical analysis
//
// # Usage
//
// Basic parsing example:
//
//	report, err := dataprocessing.ParseFile("2024_01_15 ISX Daily Report.xlsx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Processing with forward-fill:
//
//	processor := dataprocessing.NewForwardFillProcessor()
//	filledRecords := processor.FillMissingData(report.Records)
//
// Generate summaries:
//
//	generator := dataprocessing.NewSummaryGenerator(paths)
//	err := generator.GenerateFromCombinedCSV("combined.csv", "summary.csv")
//
// # Data Flow
//
// The typical data flow through this package:
//
//	Excel File → Parser → TradeRecords → Processor → Enhanced Records → Analytics → Reports
//
// # Error Handling
//
// All functions return detailed errors that include context about what failed:
//
//	- File parsing errors include the problematic row/column
//	- Processing errors indicate which record caused the issue
//	- Analytics errors specify which calculation failed
//
// # Performance Considerations
//
// The package is designed to handle large datasets efficiently:
//
//	- Streaming CSV writers for memory efficiency
//	- Concurrent processing where applicable
//	- Minimal allocations in hot paths
//
// # Testing
//
// The package includes comprehensive tests for all components.
// Use table-driven tests when adding new functionality.
package dataprocessing