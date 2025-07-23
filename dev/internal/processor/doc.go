// Package processor provides data processing algorithms for ISX trade records.
//
// The main component is the ForwardFillProcessor which handles missing data
// by carrying forward the last known trading values for symbols that don't
// trade on certain days.
//
// Forward-Fill Algorithm:
// When a stock doesn't trade on a given day, the processor creates synthetic
// records using the last known closing price as the price for all fields
// (open, high, low, close). These filled records are marked with
// TradingStatus=false to distinguish them from actual trading data.
//
// Example usage:
//
//	processor := processor.NewForwardFillProcessor()
//	filledRecords := processor.FillMissingData(rawRecords)
//	
//	// Or with statistics
//	filledRecords, stats := processor.FillMissingDataWithStats(rawRecords)
//	fmt.Printf("Filled %d records for %d symbols across %d days\n",
//	    stats.ForwardFilledCount, stats.SymbolsProcessed, stats.DatesProcessed)
package processor