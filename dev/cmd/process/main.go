package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/parser"
)

// ExcelFileInfo holds information about an Excel file
type ExcelFileInfo struct {
	Name string
	Date time.Time
}

func main() {
	inDir := flag.String("in", "data/downloads", "input directory for .xlsx files")
	outDir := flag.String("out", "data/reports", "output directory for CSV files")
	fullRework := flag.Bool("full", false, "force full rework of all files")
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting ISX Daily Reports processing...\n")
	fmt.Printf("Input directory: %s\n", *inDir)
	fmt.Printf("Output directory: %s\n", *outDir)
	fmt.Printf("Full rework: %v\n", *fullRework)

	// Get all available Excel files
	files, err := ioutil.ReadDir(*inDir)
	if err != nil {
		fmt.Printf("failed to read input dir: %v\n", err)
		os.Exit(1)
	}

	// Parse and sort all available files by date
	var excelFiles []ExcelFileInfo
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".xlsx") || strings.HasPrefix(file.Name(), "~$") {
			continue
		}

		// Extract date from filename (e.g., "YYYY MM DD ISX Daily Report.xlsx")
		parts := strings.Split(file.Name(), " ")
		if len(parts) < 4 {
			continue // Skip malformed filenames
		}

		dateStr := strings.Join(parts[0:3], " ")
		date, err := time.Parse("2006 01 02", dateStr)
		if err != nil {
			fmt.Printf("Warning: Could not parse date from filename %s: %v\n", file.Name(), err)
			continue
		}

		excelFiles = append(excelFiles, ExcelFileInfo{
			Name: file.Name(),
			Date: date,
		})
	}

	// Sort files by date
	sort.Slice(excelFiles, func(i, j int) bool {
		return excelFiles[i].Date.Before(excelFiles[j].Date)
	})

	fmt.Printf("%d Excel files discovered\n", len(excelFiles))

	// Check what needs to be processed
	var filesToProcess []ExcelFileInfo
	var existingRecords []parser.TradeRecord

	if *fullRework {
		fmt.Printf("Full rework requested - processing all files\n")
		filesToProcess = excelFiles
	} else {
		// Smart update: check what's already processed
		filesToProcess, existingRecords = determineFilesToProcess(excelFiles, *outDir)
		fmt.Printf("Smart update: %d files need processing\n", len(filesToProcess))
	}

	// Process the required files
	var newRecords []parser.TradeRecord
	totalFiles := len(filesToProcess)

	for i, fileInfo := range filesToProcess {
		fmt.Printf("Processing file %d/%d: %s\n", i+1, totalFiles, fileInfo.Name)
		fmt.Printf("Processing: %s\n", fileInfo.Name)

		report, err := parser.ParseFile(filepath.Join(*inDir, fileInfo.Name))
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", fileInfo.Name, err)
			continue
		}

		// Update all records with the correct date
		for i := range report.Records {
			report.Records[i].Date = fileInfo.Date
		}

		fmt.Printf("%d records processed from %s\n", len(report.Records), fileInfo.Name)

		// Note: Daily CSV files will be generated after forward-fill processing
		// to ensure they include forward-filled data with proper trading status

		// Add to new records
		newRecords = append(newRecords, report.Records...)

		// Print a few sample records
		for i, record := range report.Records {
			if i >= 3 { // Print up to 3 records
				break
			}
			fmt.Printf("  Symbol: %s (%s), Date: %s, Close: %.3f, Volume: %d\n",
				record.CompanySymbol, record.CompanyName, record.Date.Format("2006-01-02"),
				record.ClosePrice, record.Volume)
		}
	}

	// Combine existing and new records
	allRecords := append(existingRecords, newRecords...)

	// Apply forward-fill and generate all output files
	if len(allRecords) > 0 {
		fmt.Printf("Generating dataset with forward-fill...\n")
		filledRecords := forwardFillMissingData(allRecords)

		fmt.Printf("%d records processed\n", len(filledRecords))
		fmt.Printf("%d active trading records\n", len(allRecords))
		fmt.Printf("%d forward-filled records\n", len(filledRecords)-len(allRecords))

		// Save combined CSV with forward-fill
		combinedCSVPath := filepath.Join(*outDir, "isx_combined_data.csv")
		if err := saveCombinedCSV(combinedCSVPath, filledRecords); err != nil {
			fmt.Printf("Error saving combined CSV: %v\n", err)
		} else {
			fmt.Printf("Saved combined report: %s\n", combinedCSVPath)
		}

		// Generate daily CSV files with forward-fill
		fmt.Printf("Generating daily CSV files with forward-fill...\n")
		if err := generateDailyFiles(filledRecords, *outDir); err != nil {
			fmt.Printf("Error generating daily files: %v\n", err)
		} else {
			fmt.Printf("Daily files generated successfully\n")
		}

		// Generate individual ticker CSV files with forward-fill
		fmt.Printf("Generating individual ticker CSV files with forward-fill...\n")
		if err := generateTickerFiles(filledRecords, *outDir); err != nil {
			fmt.Printf("Error generating ticker files: %v\n", err)
		} else {
			fmt.Printf("Ticker files generated successfully\n")
		}
	}

	fmt.Println("Processing complete.")

	// Generate ticker summary CSV
	fmt.Println("Generating ticker summary...")
	if err := generateTickerSummary(*outDir); err != nil {
		fmt.Printf("Warning: Failed to generate ticker summary CSV: %v\n", err)
	} else {
		fmt.Println("Ticker summary CSV generated successfully")
	}
	
	// Generate ticker summary JSON for web interface
	if err := generateTickerSummaryJSON(*outDir); err != nil {
		fmt.Printf("Warning: Failed to generate ticker summary JSON: %v\n", err)
	} else {
		fmt.Println("Ticker summary JSON generated successfully")
	}
}

// determineFilesToProcess checks which files need to be processed based on existing CSV files
func determineFilesToProcess(excelFiles []ExcelFileInfo, outDir string) ([]ExcelFileInfo, []parser.TradeRecord) {
	var filesToProcess []ExcelFileInfo
	var existingRecords []parser.TradeRecord

	// Check which daily CSV files already exist
	existingDates := make(map[string]bool)
	if entries, err := ioutil.ReadDir(outDir); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "isx_daily_") && strings.HasSuffix(entry.Name(), ".csv") {
				// Extract date from filename: isx_daily_YYYY_MM_DD.csv
				dateStr := strings.TrimPrefix(entry.Name(), "isx_daily_")
				dateStr = strings.TrimSuffix(dateStr, ".csv")
				existingDates[dateStr] = true
			}
		}
	}

	fmt.Printf("Found %d existing daily CSV files\n", len(existingDates))

	// Load existing records from combined CSV if it exists
	combinedCSVPath := filepath.Join(outDir, "isx_combined_data.csv")
	if _, err := os.Stat(combinedCSVPath); err == nil {
		fmt.Printf("Loading existing combined CSV data...\n")
		if records, err := loadExistingRecords(combinedCSVPath); err == nil {
			existingRecords = records
			fmt.Printf("Loaded %d existing records\n", len(existingRecords))
		} else {
			fmt.Printf("Warning: Could not load existing combined CSV: %v\n", err)
		}
	}

	// Determine which files need processing
	for _, fileInfo := range excelFiles {
		dateStr := fileInfo.Date.Format("2006_01_02")
		if !existingDates[dateStr] {
			filesToProcess = append(filesToProcess, fileInfo)
			fmt.Printf("  Need to process: %s (date: %s)\n", fileInfo.Name, dateStr)
		} else {
			fmt.Printf("  Already processed: %s (date: %s)\n", fileInfo.Name, dateStr)
		}
	}

	// If we have existing records but files to process, we need to filter out records for dates we're reprocessing
	if len(existingRecords) > 0 && len(filesToProcess) > 0 {
		fmt.Printf("Filtering existing records to avoid duplicates...\n")
		reprocessDates := make(map[string]bool)
		for _, fileInfo := range filesToProcess {
			reprocessDates[fileInfo.Date.Format("2006-01-02")] = true
		}

		var filteredRecords []parser.TradeRecord
		for _, record := range existingRecords {
			if !reprocessDates[record.Date.Format("2006-01-02")] {
				filteredRecords = append(filteredRecords, record)
			}
		}
		existingRecords = filteredRecords
		fmt.Printf("Filtered to %d existing records (removed reprocessing dates)\n", len(existingRecords))
	}

	return filesToProcess, existingRecords
}

// loadExistingRecords loads records from an existing combined CSV file
func loadExistingRecords(filePath string) ([]parser.TradeRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var tradeRecords []parser.TradeRecord
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 16 {
			continue // Skip malformed records
		}

		// Parse the record
		date, _ := time.Parse("2006-01-02", record[0])
		openPrice, _ := strconv.ParseFloat(record[3], 64)
		highPrice, _ := strconv.ParseFloat(record[4], 64)
		lowPrice, _ := strconv.ParseFloat(record[5], 64)
		avgPrice, _ := strconv.ParseFloat(record[6], 64)
		prevAvgPrice, _ := strconv.ParseFloat(record[7], 64)
		closePrice, _ := strconv.ParseFloat(record[8], 64)
		prevClosePrice, _ := strconv.ParseFloat(record[9], 64)
		change, _ := strconv.ParseFloat(record[10], 64)
		changePct, _ := strconv.ParseFloat(record[11], 64)
		numTrades, _ := strconv.ParseInt(record[12], 10, 64)
		volume, _ := strconv.ParseInt(record[13], 10, 64)
		value, _ := strconv.ParseFloat(record[14], 64)
		tradingStatus, _ := strconv.ParseBool(record[15])

		tradeRecord := parser.TradeRecord{
			CompanyName:      record[1],
			CompanySymbol:    record[2],
			Date:             date,
			OpenPrice:        openPrice,
			HighPrice:        highPrice,
			LowPrice:         lowPrice,
			AveragePrice:     avgPrice,
			PrevAveragePrice: prevAvgPrice,
			ClosePrice:       closePrice,
			PrevClosePrice:   prevClosePrice,
			Change:           change,
			ChangePercent:    changePct,
			NumTrades:        numTrades,
			Volume:           volume,
			Value:            value,
			TradingStatus:    tradingStatus,
		}
		tradeRecords = append(tradeRecords, tradeRecord)
	}

	return tradeRecords, nil
}

func saveDailyCSV(filePath string, records []parser.TradeRecord) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with all fields
	header := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := []string{
			record.Date.Format("2006-01-02"),
			record.CompanyName,
			record.CompanySymbol,
			fmt.Sprintf("%.3f", record.OpenPrice),
			fmt.Sprintf("%.3f", record.HighPrice),
			fmt.Sprintf("%.3f", record.LowPrice),
			fmt.Sprintf("%.3f", record.AveragePrice),
			fmt.Sprintf("%.3f", record.PrevAveragePrice),
			fmt.Sprintf("%.3f", record.ClosePrice),
			fmt.Sprintf("%.3f", record.PrevClosePrice),
			fmt.Sprintf("%.3f", record.Change),
			fmt.Sprintf("%.2f", record.ChangePercent),
			fmt.Sprintf("%d", record.NumTrades),
			fmt.Sprintf("%d", record.Volume),
			fmt.Sprintf("%.2f", record.Value),
			fmt.Sprintf("%t", record.TradingStatus),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// forwardFillMissingData fills in missing trading data for symbols that don't trade on certain days
func forwardFillMissingData(records []parser.TradeRecord) []parser.TradeRecord {
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
	var dates []string
	for date := range allDates {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var symbols []string
	for symbol := range allSymbols {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

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
				filledRecord := parser.TradeRecord{
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
				result = append(result, filledRecord)
				// Don't update lastKnownData since this is filled data
			}
			// If no history exists, skip this symbol for this date
		}
	}

	return result
}

func saveCombinedCSV(filePath string, records []parser.TradeRecord) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with all fields
	header := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := []string{
			record.Date.Format("2006-01-02"),
			record.CompanyName,
			record.CompanySymbol,
			fmt.Sprintf("%.3f", record.OpenPrice),
			fmt.Sprintf("%.3f", record.HighPrice),
			fmt.Sprintf("%.3f", record.LowPrice),
			fmt.Sprintf("%.3f", record.AveragePrice),
			fmt.Sprintf("%.3f", record.PrevAveragePrice),
			fmt.Sprintf("%.3f", record.ClosePrice),
			fmt.Sprintf("%.3f", record.PrevClosePrice),
			fmt.Sprintf("%.3f", record.Change),
			fmt.Sprintf("%.2f", record.ChangePercent),
			fmt.Sprintf("%d", record.NumTrades),
			fmt.Sprintf("%d", record.Volume),
			fmt.Sprintf("%.2f", record.Value),
			fmt.Sprintf("%t", record.TradingStatus),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// generateDailyFiles generates daily CSV files grouped by date from forward-filled records
func generateDailyFiles(records []parser.TradeRecord, outDir string) error {
	// Group records by date
	recordsByDate := make(map[string][]parser.TradeRecord)
	for _, record := range records {
		dateStr := record.Date.Format("2006_01_02")
		recordsByDate[dateStr] = append(recordsByDate[dateStr], record)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Generate CSV files for each date
	for dateStr, dailyRecords := range recordsByDate {
		fmt.Printf("Generating daily CSV for date: %s\n", dateStr)

		// Save CSV for the current date
		dailyCSVPath := filepath.Join(outDir, fmt.Sprintf("isx_daily_%s.csv", dateStr))
		if err := saveDailyCSV(dailyCSVPath, dailyRecords); err != nil {
			fmt.Printf("Error saving daily CSV: %v\n", err)
		} else {
			fmt.Printf("Saved daily CSV: %s (%d records)\n", dailyCSVPath, len(dailyRecords))
		}
	}

	return nil
}

// generateTickerFiles generates individual CSV files for each ticker with their complete trading history
func generateTickerFiles(records []parser.TradeRecord, outDir string) error {
	// Extract all unique tickers
	tickers := make(map[string]bool)
	for _, record := range records {
		tickers[record.CompanySymbol] = true
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Generate CSV files for each ticker
	for ticker := range tickers {
		fmt.Printf("Generating CSV for ticker: %s\n", ticker)

		// Filter records for the current ticker
		var tickerRecords []parser.TradeRecord
		for _, record := range records {
			if record.CompanySymbol == ticker {
				tickerRecords = append(tickerRecords, record)
			}
		}

		// Save CSV for the current ticker
		tickerCSVPath := filepath.Join(outDir, fmt.Sprintf("%s_trading_history.csv", ticker))
		if err := saveDailyCSV(tickerCSVPath, tickerRecords); err != nil {
			fmt.Printf("Error saving ticker CSV: %v\n", err)
		} else {
			fmt.Printf("Saved ticker CSV: %s\n", tickerCSVPath)
		}
	}

	return nil
}

// generateTickerSummary creates a ticker summary CSV from the combined CSV file
func generateTickerSummary(outDir string) error {
	combinedFile := filepath.Join(outDir, "isx_combined_data.csv")
	summaryFile := filepath.Join(outDir, "ticker_summary.csv")

	// Check if combined file exists
	if _, err := os.Stat(combinedFile); os.IsNotExist(err) {
		return fmt.Errorf("combined CSV file not found: %s", combinedFile)
	}

	// Read combined CSV
	file, err := os.Open(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to open combined file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read combined CSV: %v", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("combined CSV has no data rows")
	}

	// Parse header to find column indices
	header := records[0]
	tickerCol := -1
	companyCol := -1
	dateCol := -1
	closeCol := -1

	for i, col := range header {
		switch strings.ToLower(col) {
		case "ticker", "company_symbol", "symbol":
			tickerCol = i
		case "company_name", "companyname", "company", "name":
			companyCol = i
		case "date":
			dateCol = i
		case "close_price", "closeprice", "close":
			closeCol = i
		}
	}

	if tickerCol == -1 || companyCol == -1 || dateCol == -1 || closeCol == -1 {
		return fmt.Errorf("required columns not found in combined CSV. Found: %v", header)
	}

	// Group data by ticker
	tickerData := make(map[string][]map[string]string)

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) <= tickerCol || len(record) <= companyCol || len(record) <= dateCol || len(record) <= closeCol {
			continue
		}

		ticker := strings.TrimSpace(record[tickerCol])
		if ticker == "" {
			continue
		}

		rowData := map[string]string{
			"ticker":       ticker,
			"company_name": strings.TrimSpace(record[companyCol]),
			"date":         strings.TrimSpace(record[dateCol]),
			"close_price":  strings.TrimSpace(record[closeCol]),
		}

		tickerData[ticker] = append(tickerData[ticker], rowData)
	}

	// Create ticker summaries
	type TickerSummary struct {
		Ticker      string
		CompanyName string
		LastPrice   float64
		LastDate    string
		TradingDays int
		Last10Days  []float64
	}

	var summaries []TickerSummary

	for ticker, data := range tickerData {
		if len(data) == 0 {
			continue
		}

		// Sort by date
		sort.Slice(data, func(i, j int) bool {
			return data[i]["date"] < data[j]["date"]
		})

		lastRecord := data[len(data)-1]
		lastPrice, _ := strconv.ParseFloat(lastRecord["close_price"], 64)

		// Get last 10 trading days
		var last10Days []float64
		start := len(data) - 10
		if start < 0 {
			start = 0
		}

		for i := start; i < len(data); i++ {
			price, _ := strconv.ParseFloat(data[i]["close_price"], 64)
			last10Days = append(last10Days, price)
		}

		summary := TickerSummary{
			Ticker:      ticker,
			CompanyName: lastRecord["company_name"],
			LastPrice:   lastPrice,
			LastDate:    lastRecord["date"],
			TradingDays: len(data),
			Last10Days:  last10Days,
		}

		summaries = append(summaries, summary)
	}

	// Sort summaries by ticker
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})

	// Write ticker summary CSV
	outFile, err := os.Create(summaryFile)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days"})

	// Write data
	for _, summary := range summaries {
		last10DaysStr := ""
		for i, price := range summary.Last10Days {
			if i > 0 {
				last10DaysStr += ","
			}
			last10DaysStr += fmt.Sprintf("%.3f", price)
		}

		writer.Write([]string{
			summary.Ticker,
			summary.CompanyName,
			fmt.Sprintf("%.3f", summary.LastPrice),
			summary.LastDate,
			fmt.Sprintf("%d", summary.TradingDays),
			last10DaysStr,
		})
	}

	fmt.Printf("Generated ticker summary with %d tickers\n", len(summaries))
	return nil
}
