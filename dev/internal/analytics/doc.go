// Package analytics provides data analysis and summary generation for ISX trade records.
//
// The main component is the SummaryGenerator which creates ticker summaries
// from the combined CSV file, including statistics like last price, trading days,
// and recent price history.
//
// Example usage:
//
//	generator := analytics.NewSummaryGenerator("/path/to/base")
//	
//	// Generate ticker summary from combined CSV
//	err := generator.GenerateFromCombinedCSV(
//	    "data/reports/isx_combined_data.csv",
//	    "data/reports/ticker_summary.csv",
//	)
//	
//	if err != nil {
//	    log.Fatal(err)
//	}
package analytics