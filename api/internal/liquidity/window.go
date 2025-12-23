package liquidity

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AssembleWindow loads and organizes trading data for the specified window and tickers
// This function handles CSV file loading, data validation, and calendar alignment
//
// Parameters:
//   - csvDir: directory containing CSV files with trading data
//   - window: time window for the calculation (20, 60, or 120 days)
//   - tickers: list of ticker symbols to load (empty slice loads all available)
//
// Returns: map of ticker symbol to sorted trading data
func AssembleWindow(ctx context.Context, csvDir string, window Window, tickers []string) (map[string][]TradingDay, error) {
	logger := slog.Default()
	
	logger.InfoContext(ctx, "assembling trading data window",
		"csv_dir", csvDir,
		"window", window.String(),
		"num_tickers", len(tickers),
	)
	
	// Validate input directory
	if _, err := os.Stat(csvDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("CSV directory does not exist: %s", csvDir)
	}
	
	// Find all CSV files in the directory
	csvFiles, err := findCSVFiles(csvDir)
	if err != nil {
		return nil, fmt.Errorf("find CSV files: %w", err)
	}
	
	if len(csvFiles) == 0 {
		return nil, fmt.Errorf("no CSV files found in directory: %s", csvDir)
	}
	
	logger.InfoContext(ctx, "found CSV files", "count", len(csvFiles))
	
	// Load trading data from CSV files
	allData := make(map[string][]TradingDay)
	
	for _, csvFile := range csvFiles {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during data loading: %w", ctx.Err())
		default:
		}
		
		tickerData, err := LoadTradingData(csvFile)
		if err != nil {
			logger.WarnContext(ctx, "failed to load CSV file",
				"file", csvFile,
				"error", err,
			)
			continue // Skip problematic files
		}
		
		// Group data by ticker symbol
		for _, td := range tickerData {
			if len(tickers) > 0 && !contains(tickers, td.Symbol) {
				continue // Skip tickers not in the requested list
			}
			
			allData[td.Symbol] = append(allData[td.Symbol], td)
		}
	}
	
	if len(allData) == 0 {
		return nil, fmt.Errorf("no valid trading data loaded")
	}
	
	// Sort data by date for each ticker and validate
	validatedData := make(map[string][]TradingDay)
	
	for symbol, data := range allData {
		// Sort by date
		sort.Slice(data, func(i, j int) bool {
			return data[i].Date.Before(data[j].Date)
		})
		
		// Validate and filter data
		validData := filterValidData(data, window)
		if len(validData) >= window.Days() {
			validatedData[symbol] = validData
		} else {
			logger.WarnContext(ctx, "insufficient data for ticker",
				"symbol", symbol,
				"data_points", len(validData),
				"required", window.Days(),
			)
		}
	}
	
	logger.InfoContext(ctx, "data assembly completed",
		"valid_tickers", len(validatedData),
		"window_days", window.Days(),
	)
	
	return validatedData, nil
}

// LoadTradingData loads trading data from a single CSV file
// Expected CSV format: Date,Symbol,Open,High,Low,Close,Volume,NumTrades,Status
func LoadTradingData(csvPath string) ([]TradingDay, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("open CSV file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV records: %w", err)
	}
	
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}
	
	// Check if first row is header
	var dataStart int
	if isHeaderRow(records[0]) {
		dataStart = 1
	}
	
	if len(records) <= dataStart {
		return nil, fmt.Errorf("CSV file contains only header")
	}
	
	var tradingDays []TradingDay
	
	for i := dataStart; i < len(records); i++ {
		record := records[i]
		
		td, err := parseTradingDayRecord(record, i+1)
		if err != nil {
			// Log warning but continue with other records
			slog.Warn("failed to parse CSV record",
				"file", filepath.Base(csvPath),
				"line", i+1,
				"error", err,
			)
			continue
		}
		
		tradingDays = append(tradingDays, td)
	}
	
	return tradingDays, nil
}

// parseTradingDayRecord parses a single CSV record into TradingDay struct
func parseTradingDayRecord(record []string, lineNum int) (TradingDay, error) {
	if len(record) < 6 {
		return TradingDay{}, fmt.Errorf("insufficient columns in record (line %d): expected at least 6, got %d", lineNum, len(record))
	}
	
	// Parse date (multiple formats supported)
	date, err := parseDate(strings.TrimSpace(record[0]))
	if err != nil {
		return TradingDay{}, fmt.Errorf("parse date (line %d): %w", lineNum, err)
	}
	
	// Parse symbol
	symbol := strings.TrimSpace(strings.ToUpper(record[1]))
	if symbol == "" {
		return TradingDay{}, fmt.Errorf("empty symbol (line %d)", lineNum)
	}
	
	// Parse OHLC prices
	open, err := parseFloat(record[2], "open", lineNum)
	if err != nil {
		return TradingDay{}, err
	}
	
	high, err := parseFloat(record[3], "high", lineNum)
	if err != nil {
		return TradingDay{}, err
	}
	
	low, err := parseFloat(record[4], "low", lineNum)
	if err != nil {
		return TradingDay{}, err
	}
	
	close, err := parseFloat(record[5], "close", lineNum)
	if err != nil {
		return TradingDay{}, err
	}
	
	// Parse volume (default to 0 if missing)
	var volume float64
	if len(record) > 6 && strings.TrimSpace(record[6]) != "" {
		volume, err = parseFloat(record[6], "volume", lineNum)
		if err != nil {
			return TradingDay{}, err
		}
	}
	
	// Parse number of trades (default to 0 if missing)
	var numTrades int
	if len(record) > 7 && strings.TrimSpace(record[7]) != "" {
		numTrades, err = strconv.Atoi(strings.TrimSpace(record[7]))
		if err != nil {
			return TradingDay{}, fmt.Errorf("parse num_trades (line %d): %w", lineNum, err)
		}
	}
	
	// Parse trading status (default to "ACTIVE" if missing)
	status := "ACTIVE"
	if len(record) > 8 && strings.TrimSpace(record[8]) != "" {
		status = strings.TrimSpace(strings.ToUpper(record[8]))
	}
	
	return TradingDay{
		Date:          date,
		Symbol:        symbol,
		Open:          open,
		High:          high,
		Low:           low,
		Close:         close,
		Volume:        volume,
		NumTrades:     numTrades,
		TradingStatus: status,
	}, nil
}

// parseDate attempts to parse date strings in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	dateFormats := []string{
		"2006-01-02",           // ISO format
		"01/02/2006",           // US format
		"02/01/2006",           // European format
		"2006/01/02",           // Alternative ISO
		"2006-01-02 15:04:05",  // With time
		"01-02-2006",           // US with dashes
		"02-01-2006",           // European with dashes
	}
	
	for _, format := range dateFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseFloat safely parses a float64 value from string
func parseFloat(str, fieldName string, lineNum int) (float64, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return 0, fmt.Errorf("empty %s (line %d)", fieldName, lineNum)
	}
	
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s (line %d): %w", fieldName, lineNum, err)
	}
	
	return value, nil
}

// isHeaderRow checks if the first row contains headers
func isHeaderRow(record []string) bool {
	if len(record) == 0 {
		return false
	}
	
	// Common header indicators
	headers := []string{"date", "symbol", "open", "high", "low", "close", "volume"}
	firstCell := strings.ToLower(strings.TrimSpace(record[0]))
	
	for _, header := range headers {
		if strings.Contains(firstCell, header) {
			return true
		}
	}
	
	// Try parsing as date - if it fails, likely a header
	_, err := parseDate(strings.TrimSpace(record[0]))
	return err != nil
}

// findCSVFiles finds all CSV files in the specified directory
func findCSVFiles(dir string) ([]string, error) {
	var csvFiles []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			csvFiles = append(csvFiles, path)
		}
		
		return nil
	})
	
	return csvFiles, err
}

// filterValidData filters and validates trading data
func filterValidData(data []TradingDay, window Window) []TradingDay {
	var validData []TradingDay
	
	for _, td := range data {
		if td.IsValid() {
			validData = append(validData, td)
		}
	}
	
	return validData
}

// BuildCalendar creates a trading calendar from the loaded data
// This helps identify trading days vs. non-trading days
func BuildCalendar(data []TradingDay) []time.Time {
	dateSet := make(map[string]bool)
	
	for _, td := range data {
		if td.IsTrading() {
			dateKey := td.Date.Format("2006-01-02")
			dateSet[dateKey] = true
		}
	}
	
	var calendar []time.Time
	for dateKey := range dateSet {
		if date, err := time.Parse("2006-01-02", dateKey); err == nil {
			calendar = append(calendar, date)
		}
	}
	
	sort.Slice(calendar, func(i, j int) bool {
		return calendar[i].Before(calendar[j])
	})
	
	return calendar
}

// GetDateRange returns the date range covered by the data
func GetDateRange(data map[string][]TradingDay) (startDate, endDate time.Time, err error) {
	var allDates []time.Time
	
	for _, tickerData := range data {
		for _, td := range tickerData {
			allDates = append(allDates, td.Date)
		}
	}
	
	if len(allDates) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("no trading data available")
	}
	
	sort.Slice(allDates, func(i, j int) bool {
		return allDates[i].Before(allDates[j])
	})
	
	return allDates[0], allDates[len(allDates)-1], nil
}

// ValidateDataQuality performs comprehensive data quality checks
func ValidateDataQuality(data map[string][]TradingDay, window Window) map[string]interface{} {
	report := make(map[string]interface{})
	
	totalTickers := len(data)
	totalDataPoints := 0
	validDataPoints := 0
	tradingDataPoints := 0
	
	tickerStats := make(map[string]map[string]interface{})
	
	for symbol, tickerData := range data {
		totalDataPoints += len(tickerData)
		
		tickerValid := 0
		tickerTrading := 0
		
		for _, td := range tickerData {
			if td.IsValid() {
				validDataPoints++
				tickerValid++
			}
			if td.IsTrading() {
				tradingDataPoints++
				tickerTrading++
			}
		}
		
		tickerStats[symbol] = map[string]interface{}{
			"total_days":    len(tickerData),
			"valid_days":    tickerValid,
			"trading_days":  tickerTrading,
			"valid_ratio":   float64(tickerValid) / float64(len(tickerData)),
			"trading_ratio": float64(tickerTrading) / float64(len(tickerData)),
			"sufficient":    len(tickerData) >= window.Days() && tickerTrading >= MinTradingDaysForCalc,
		}
	}
	
	report["summary"] = map[string]interface{}{
		"total_tickers":       totalTickers,
		"total_data_points":   totalDataPoints,
		"valid_data_points":   validDataPoints,
		"trading_data_points": tradingDataPoints,
		"valid_ratio":         float64(validDataPoints) / float64(totalDataPoints),
		"trading_ratio":       float64(tradingDataPoints) / float64(totalDataPoints),
		"window_requirement":  window.Days(),
		"min_trading_days":    MinTradingDaysForCalc,
	}
	
	report["tickers"] = tickerStats
	
	return report
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ExportDataSample exports a sample of loaded data for inspection
func ExportDataSample(data map[string][]TradingDay, sampleSize int, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create sample file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header
	header := []string{"Date", "Symbol", "Open", "High", "Low", "Close", "Volume", "NumTrades", "Status"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	
	// Write sample data
	count := 0
	for _, tickerData := range data {
		for _, td := range tickerData {
			if count >= sampleSize {
				return nil
			}
			
			record := []string{
				td.Date.Format("2006-01-02"),
				td.Symbol,
				fmt.Sprintf("%.4f", td.Open),
				fmt.Sprintf("%.4f", td.High),
				fmt.Sprintf("%.4f", td.Low),
				fmt.Sprintf("%.4f", td.Close),
				fmt.Sprintf("%.0f", td.Volume),
				strconv.Itoa(td.NumTrades),
				td.TradingStatus,
			}
			
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("write record: %w", err)
			}
			
			count++
		}
	}
	
	return nil
}