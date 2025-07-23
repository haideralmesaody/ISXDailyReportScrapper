package processor

import (
	"sort"
	"time"

	"isxcli/internal/parser"
)

// ForwardFillProcessor handles forward-fill operations for missing trading data
type ForwardFillProcessor struct{}

// NewForwardFillProcessor creates a new forward-fill processor
func NewForwardFillProcessor() *ForwardFillProcessor {
	return &ForwardFillProcessor{}
}

// FillMissingData fills in missing trading data for symbols that don't trade on certain days
// It uses the last known trading data to fill gaps, marking filled records with TradingStatus=false
func (f *ForwardFillProcessor) FillMissingData(records []parser.TradeRecord) []parser.TradeRecord {
	if len(records) == 0 {
		return records
	}

	// Group records by symbol and date
	symbolsByDate := make(map[string]map[string]parser.TradeRecord) // date -> symbol -> record
	allSymbols := make(map[string]bool)
	allDates := make(map[string]bool)

	for _, record := range records {
		dateStr := record.Date.Format("2006-01-02")
		symbol := record.CompanySymbol

		if symbolsByDate[dateStr] == nil {
			symbolsByDate[dateStr] = make(map[string]parser.TradeRecord)
		}
		symbolsByDate[dateStr][symbol] = record
		allSymbols[symbol] = true
		allDates[dateStr] = true
	}

	// Convert to sorted slices
	dates := f.getSortedKeys(allDates)
	symbols := f.getSortedKeys(allSymbols)

	// Keep track of last known data for each symbol
	lastKnownData := make(map[string]parser.TradeRecord)

	var result []parser.TradeRecord

	for _, dateStr := range dates {
		date, _ := time.Parse("2006-01-02", dateStr)
		dayRecords := symbolsByDate[dateStr]

		for _, symbol := range symbols {
			if record, exists := dayRecords[symbol]; exists {
				// Symbol traded on this day - use actual data
				result = append(result, record)
				lastKnownData[symbol] = record
			} else if lastRecord, hasHistory := lastKnownData[symbol]; hasHistory {
				// Symbol didn't trade - forward fill from last known data
				filledRecord := f.createFilledRecord(lastRecord, symbol, date)
				result = append(result, filledRecord)
				// Don't update lastKnownData since this is filled data
			}
			// If no history exists, skip this symbol for this date
		}
	}

	return result
}

// createFilledRecord creates a forward-filled record based on the last known data
func (f *ForwardFillProcessor) createFilledRecord(lastRecord parser.TradeRecord, symbol string, date time.Time) parser.TradeRecord {
	return parser.TradeRecord{
		CompanyName:      lastRecord.CompanyName,
		CompanySymbol:    symbol,
		Date:             date,
		OpenPrice:        lastRecord.ClosePrice,   // Open = previous close
		HighPrice:        lastRecord.ClosePrice,   // High = previous close
		LowPrice:         lastRecord.ClosePrice,   // Low = previous close
		AveragePrice:     lastRecord.ClosePrice,   // Average = previous close
		PrevAveragePrice: lastRecord.AveragePrice, // Keep previous average
		ClosePrice:       lastRecord.ClosePrice,   // Close = previous close
		PrevClosePrice:   lastRecord.ClosePrice,   // Prev close = previous close
		Change:           0.0,                     // No change
		ChangePercent:    0.0,                     // No change %
		NumTrades:        0,                       // No trades
		Volume:           0,                       // No volume
		Value:            0.0,                     // No value
		TradingStatus:    false,                   // Forward-filled data
	}
}

// getSortedKeys extracts and sorts keys from a map[string]bool
func (f *ForwardFillProcessor) getSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Statistics represents forward-fill operation statistics
type Statistics struct {
	TotalRecords       int
	ActiveRecords      int
	ForwardFilledCount int
	SymbolsProcessed   int
	DatesProcessed     int
}

// FillMissingDataWithStats performs forward-fill and returns statistics
func (f *ForwardFillProcessor) FillMissingDataWithStats(records []parser.TradeRecord) ([]parser.TradeRecord, Statistics) {
	originalCount := len(records)
	filledRecords := f.FillMissingData(records)
	
	// Count unique symbols and dates
	uniqueSymbols := make(map[string]bool)
	uniqueDates := make(map[string]bool)
	for _, record := range filledRecords {
		uniqueSymbols[record.CompanySymbol] = true
		uniqueDates[record.Date.Format("2006-01-02")] = true
	}
	
	stats := Statistics{
		TotalRecords:       len(filledRecords),
		ActiveRecords:      originalCount,
		ForwardFilledCount: len(filledRecords) - originalCount,
		SymbolsProcessed:   len(uniqueSymbols),
		DatesProcessed:     len(uniqueDates),
	}
	
	return filledRecords, stats
}