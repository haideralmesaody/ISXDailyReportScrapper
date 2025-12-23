package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/dataprocessing"
	"github.com/isxcli/isxcli/internal/exporter"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/operations"
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// ExcelFileInfo holds information about an Excel file
type ExcelFileInfo struct {
	Name string
	Date time.Time
}

func main() {
	inDir := flag.String("in", "", "input directory for .xlsx files (defaults to data/downloads relative to executable)")
	outDir := flag.String("out", "", "output directory for CSV files (defaults to data/reports relative to executable)")
	fullRework := flag.Bool("full", false, "force full rework of all files")
	fromStr := flag.String("from", "", "start date (YYYY-MM-DD) (defaults to January 1st of current year)")
	toStr := flag.String("to", "", "end date (YYYY-MM-DD) (defaults to today)")
	flag.Parse()

	// Initialize paths first to get default directories
	paths, err := config.GetPaths()
	if err != nil {
		slog.Error("Failed to initialize paths", "error", err)
		os.Exit(1)
	}

	// Initialize license service for validation
	ctx := context.Background()
	slog.Info("Starting ISX Daily Reports processor - validating license...")

	// Use existing license service
	licenseManager, err := license.NewManager()
	if err != nil {
		slog.Error("Failed to initialize license manager", "error", err)
		fmt.Printf("\n❌ LICENSE VALIDATION FAILED\n")
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	// Simple license validation
	if valid, err := licenseManager.ValidateLicenseWithContext(ctx); !valid {
		slog.Error("Processor license validation failed - exiting",
			"error", err,
			"operation", "processor_startup_validation")

		fmt.Printf("\n❌ LICENSE VALIDATION FAILED\n")
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	slog.Info("✅ License validation successful - processor starting")

	// Set default date range if not provided (matching scraper behavior)
	if *fromStr == "" {
		currentYear := time.Now().Year()
		*fromStr = fmt.Sprintf("%d-01-01", currentYear)
	}
	if *toStr == "" {
		*toStr = time.Now().Format("2006-01-02")
	}

	// Parse date range for filtering
	fromDate, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		slog.Error("Invalid from date format", "date", *fromStr, "error", err)
		os.Exit(1)
	}
	toDate, err := time.Parse("2006-01-02", *toStr)
	if err != nil {
		slog.Error("Invalid to date format", "date", *toStr, "error", err)
		os.Exit(1)
	}

	// Use centralized directories as defaults if not specified
	if *inDir == "" {
		*inDir = paths.DownloadsDir
	}
	if *outDir == "" {
		*outDir = paths.ReportsDir
	}
	// Normalize relative outDir to absolute to prevent double-prefixing by CSV writer.
	if !filepath.IsAbs(*outDir) {
		*outDir = filepath.Join(paths.ExecutableDir, *outDir)
	}

	// Ensure all required directories exist
	if err := paths.EnsureDirectories(); err != nil {
		slog.Error("Failed to create required directories", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger per CLAUDE.md
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("Failed to load config, using defaults", "error", err)
		cfg = &config.Config{
			Logging: config.LoggingConfig{
				Level:       "info",
				Format:      "json",
				Output:      "both",
				FilePath:    paths.GetLogPath("process.log"),
				Development: false,
			},
		}
	}

	logger, err := infrastructure.InitializeLogger(cfg.Logging)
	if err != nil {
		slog.Warn("Failed to initialize logger, using default", "error", err)
		logger = slog.Default()
	}

	logger.Info("Starting ISX Daily Reports processing",
		slog.String("input_dir", *inDir),
		slog.String("output_dir", *outDir),
		slog.Bool("full_rework", *fullRework),
		slog.String("date_range", fmt.Sprintf("%s to %s", *fromStr, *toStr)),
		slog.String("executable_dir", paths.ExecutableDir))

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		logger.Error("Error creating output directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Starting ISX Daily Reports processing...")

	// Get all available Excel files
	files, err := ioutil.ReadDir(*inDir)
	if err != nil {
		logger.Error("Failed to read input directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Parse and filter files by date range
	var excelFiles []ExcelFileInfo
	var skippedFiles []string
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
			logger.Warn("Could not parse date from filename",
				slog.String("filename", file.Name()),
				slog.String("error", err.Error()))
			continue
		}

		// Filter by date range
		if date.Before(fromDate) || date.After(toDate) {
			skippedFiles = append(skippedFiles, file.Name())
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

	logger.Info("Excel files discovered",
		slog.Int("count", len(excelFiles)),
		slog.Int("skipped_outside_range", len(skippedFiles)),
		slog.String("date_range", fmt.Sprintf("%s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))))

	if len(skippedFiles) > 0 {
		logger.Debug("Files skipped (outside date range)",
			slog.Int("count", len(skippedFiles)),
			slog.String("sample_files", strings.Join(skippedFiles[:min(3, len(skippedFiles))], ", ")))
	}

	// Output progress message for stages.go to parse
	fmt.Printf("Found %d Excel files\n", len(excelFiles))

	// Graceful exit if no Excel files found
	if len(excelFiles) == 0 {
		logger.Warn("No Excel files found in input directory",
			slog.String("input_dir", *inDir),
			slog.String("pattern", "*.xlsx"))

		fmt.Println("Processing complete: 0 files")
		return
	}

	// Output file list for stages.go to parse
	if len(excelFiles) > 0 {
		var fileNames []string
		for _, f := range excelFiles {
			fileNames = append(fileNames, f.Name)
		}
		fmt.Printf("Files to process: %s\n", strings.Join(fileNames, "|"))
	}

	// Check what needs to be processed
	var filesToProcess []ExcelFileInfo
	var existingRecords []domain.TradeRecord

	if *fullRework {
		logger.Info("Full rework requested - processing all files in date range")
		filesToProcess = excelFiles
	} else {
		// Smart update: check what's already processed
		logger.Info("Smart processing mode - detecting unprocessed files")
		filesToProcess, existingRecords = detectUnprocessedFiles(excelFiles, *outDir, logger)
		logger.Info("Processing analysis complete",
			slog.Int("files_to_process", len(filesToProcess)),
			slog.Int("files_already_processed", len(excelFiles)-len(filesToProcess)),
			slog.Int("existing_records_loaded", len(existingRecords)))
	}

	// Process required files
	var newRecords []domain.TradeRecord
	totalFiles := len(filesToProcess)

	for i, fileInfo := range filesToProcess {
		logger.Info("Processing file",
			slog.Int("current", i+1),
			slog.Int("total", totalFiles),
			slog.String("filename", fileInfo.Name))

		// Output progress message for stages.go to parse
		fmt.Printf("Processing file %d of %d: %s\n", i+1, totalFiles, fileInfo.Name)
		_ = os.Stdout.Sync()

		// Add intermediate progress message to show file being actively processed
		fmt.Printf("Starting to parse %s...\n", fileInfo.Name)
		_ = os.Stdout.Sync()

		report, err := dataprocessing.ParseFile(filepath.Join(*inDir, fileInfo.Name))
		if err != nil {
			logger.Error("Error parsing file",
				slog.String("filename", fileInfo.Name),
				slog.String("error", err.Error()))
			// Emit stdout signal so processing_stage can count failures live
			fmt.Printf("Failed file: %s - %v\n", fileInfo.Name, err)
			_ = os.Stdout.Sync()
			continue
		}

		// Update all records with correct date
		for i := range report.Records {
			report.Records[i].Date = fileInfo.Date
		}

		logger.Info("Records processed from file",
			slog.Int("record_count", len(report.Records)),
			slog.String("filename", fileInfo.Name))

		// Add to new records
		newRecords = append(newRecords, report.Records...)

		// Emit stdout signal so processing_stage can track progress live
		fmt.Printf("Completed file: %s\n", fileInfo.Name)
		_ = os.Stdout.Sync()
	}

	// Combine existing and new records
	allRecords := append(existingRecords, newRecords...)

	// Apply forward-fill and generate all output files
	if len(allRecords) > 0 {
		slog.Info("Generating dataset with forward-fill...")
		filledRecords := forwardFillMissingData(allRecords)

		logger.Info("Record processing summary",
			slog.Int("total_records", len(filledRecords)),
			slog.Int("active_trading_records", len(allRecords)),
			slog.Int("forward_filled_records", len(filledRecords)-len(allRecords)))

		// Generate output files using existing services
		// Flush stdout before generation to ensure all file processing messages are sent
		_ = os.Stdout.Sync()

		if err := generateOutputFiles(filledRecords, *outDir, logger); err != nil {
			logger.Error("Error generating output files", "error", err)
			os.Exit(1)
		}
	}

	// Output completion message for stages.go to parse
	fmt.Println("All files processed")

	// Final processing summary with intelligent processing insights
	totalFilesInRange := len(excelFiles)
	filesProcessed := len(filesToProcess)
	filesSkipped := totalFilesInRange - filesProcessed

	logger.Info("Processing completed successfully",
		slog.Int("total_files_in_date_range", totalFilesInRange),
		slog.Int("files_processed", filesProcessed),
		slog.Int("files_skipped_already_processed", filesSkipped),
		slog.String("processing_mode", func() string {
			if *fullRework {
				return "full_rework"
			}
			return "intelligent_incremental"
		}()),
		slog.String("efficiency_note", fmt.Sprintf("Avoided reprocessing %d files", filesSkipped)))

	if filesSkipped > 0 && !*fullRework {
		logger.Info("Intelligent processing avoided redundant work",
			slog.Int("files_with_existing_csv", filesSkipped),
			slog.String("recommendation", "Use -full flag to force reprocessing all files"))
	}

	// Output completion message for stages.go to parse
	fmt.Printf("Processing complete: %d files\n", len(filesToProcess))
}

// detectUnprocessedFiles intelligently detects which files haven't been processed yet
func detectUnprocessedFiles(excelFiles []ExcelFileInfo, outDir string, logger *slog.Logger) ([]ExcelFileInfo, []domain.TradeRecord) {
	var filesToProcess []ExcelFileInfo
	var existingRecords []domain.TradeRecord

	// Check which daily CSV files already exist in new directory structure
	existingDates := make(map[string]bool)
	dailyDir := filepath.Join(outDir, "daily")

	// Walk through all daily subdirectories
	filepath.Walk(dailyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "isx_daily_") && strings.HasSuffix(info.Name(), ".csv") {
			// Extract date from filename: isx_daily_YYYY_MM_DD.csv
			dateStr := strings.TrimPrefix(info.Name(), "isx_daily_")
			dateStr = strings.TrimSuffix(dateStr, ".csv")
			existingDates[dateStr] = true
		}
		return nil
	})

	logger.Info("Found existing daily CSV files",
		slog.Int("count", len(existingDates)),
		slog.String("daily_dir", dailyDir))

	// Load existing records from combined CSV if it exists
	combinedCSVPath := filepath.Join(outDir, "combined", "isx_combined_data.csv")
	if _, err := os.Stat(combinedCSVPath); err == nil {
		logger.Info("Loading existing combined CSV data")
		if records, err := loadExistingRecords(combinedCSVPath); err == nil {
			existingRecords = records
			logger.Info("Loaded existing records", slog.Int("count", len(existingRecords)))
		} else {
			logger.Warn("Could not load existing combined CSV", slog.String("error", err.Error()))
		}
	}

	// Determine which files need processing by comparing Excel inputs with CSV outputs
	var processedFiles []string
	var unprocessedFiles []string

	for _, fileInfo := range excelFiles {
		dateStr := fileInfo.Date.Format("2006_01_02")
		if !existingDates[dateStr] {
			// No CSV exists for this Excel file - needs processing
			filesToProcess = append(filesToProcess, fileInfo)
			unprocessedFiles = append(unprocessedFiles, fileInfo.Name)
			logger.Debug("Detected unprocessed file",
				slog.String("filename", fileInfo.Name),
				slog.String("date", dateStr),
				slog.String("reason", "no_corresponding_csv"))
		} else {
			// CSV exists for this Excel file - skip processing
			processedFiles = append(processedFiles, fileInfo.Name)
			logger.Debug("Skipping processed file",
				slog.String("filename", fileInfo.Name),
				slog.String("date", dateStr),
				slog.String("reason", "csv_already_exists"))
		}
	}

	// Log summary of processing decisions
	logger.Info("File processing analysis",
		slog.Int("total_files_in_range", len(excelFiles)),
		slog.Int("already_processed", len(processedFiles)),
		slog.Int("need_processing", len(unprocessedFiles)))

	if len(unprocessedFiles) > 0 {
		logger.Info("Files requiring processing",
			slog.String("sample_files", strings.Join(unprocessedFiles[:min(5, len(unprocessedFiles))], ", ")))
	}
	if len(processedFiles) > 0 {
		logger.Info("Files being skipped (already processed)",
			slog.String("sample_files", strings.Join(processedFiles[:min(5, len(processedFiles))], ", ")))
	}

	// If we have existing records but files to process, we need to filter out records for dates we're reprocessing
	if len(existingRecords) > 0 && len(filesToProcess) > 0 {
		logger.Info("Filtering existing records to avoid duplicates")
		reprocessDates := make(map[string]bool)
		for _, fileInfo := range filesToProcess {
			reprocessDates[fileInfo.Date.Format("2006-01-02")] = true
		}

		var filteredRecords []domain.TradeRecord
		originalCount := len(existingRecords)
		for _, record := range existingRecords {
			if !reprocessDates[record.Date.Format("2006-01-02")] {
				filteredRecords = append(filteredRecords, record)
			}
		}
		existingRecords = filteredRecords
		logger.Info("Filtered existing records",
			slog.Int("remaining_records", len(existingRecords)),
			slog.Int("removed_records", originalCount-len(filteredRecords)))
	}

	return filesToProcess, existingRecords
}

// loadExistingRecords loads records from an existing combined CSV file
func loadExistingRecords(filePath string) ([]domain.TradeRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var tradeRecords []domain.TradeRecord
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 16 {
			continue // Skip malformed records
		}

		// Parse record
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

		tradeRecord := domain.TradeRecord{
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

// forwardFillMissingData fills in missing trading data for symbols that don't trade on certain days
func forwardFillMissingData(records []domain.TradeRecord) []domain.TradeRecord {
	if len(records) == 0 {
		return records
	}

	// Group records by symbol and date
	symbolsByDate := make(map[string]map[string]domain.TradeRecord) // date -> symbol -> record
	allSymbols := make(map[string]bool)

	// Track min/max dates to generate complete range
	var minDate, maxDate time.Time

	for i, record := range records {
		dateStr := record.Date.Format("2006-01-02")
		symbol := record.CompanySymbol

		if symbolsByDate[dateStr] == nil {
			symbolsByDate[dateStr] = make(map[string]domain.TradeRecord)
		}
		symbolsByDate[dateStr][symbol] = record
		allSymbols[symbol] = true

		// Track date range
		if i == 0 {
			minDate = record.Date
			maxDate = record.Date
		} else {
			if record.Date.Before(minDate) {
				minDate = record.Date
			}
			if record.Date.After(maxDate) {
				maxDate = record.Date
			}
		}
	}

	// Generate ALL dates in range (including missing calendar dates)
	var dates []string
	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}

	var symbols []string
	for symbol := range allSymbols {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	// Keep track of last known data for each symbol
	lastKnownData := make(map[string]domain.TradeRecord)

	var result []domain.TradeRecord

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
				filledRecord := domain.TradeRecord{
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
					NumTrades:        0,                         // No trades
					Volume:           0,                         // No volume
					Value:            0.0,                       // No value
					TradingStatus:    false,                     // Forward-filled data
				}
				result = append(result, filledRecord)
				// Don't update lastKnownData since this is filled data
			}
			// If no history exists, skip this symbol for this date
		}
	}

	return result
}

// generateOutputFiles creates all necessary output files using shared services
func generateOutputFiles(records []domain.TradeRecord, outDir string, logger *slog.Logger) error {
	// Initialize paths for CSV writer
	paths, _ := config.GetPaths()
	csvWriter := exporter.NewCSVWriter(paths)

	// Create output directories
	combinedDir := filepath.Join(outDir, "combined")
	dailyDir := filepath.Join(outDir, "daily")
	tickerDir := filepath.Join(outDir, "ticker")

	for _, dir := range []string{combinedDir, dailyDir, tickerDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Pre-calculate total outputs for telemetry (1 combined + daily + ticker files)
	datesSet := uniqueDates(records)
	tickersSet := uniqueSymbols(records)
	totalOutputs := 1 + len(datesSet) + len(tickersSet)
	if totalOutputs < 1 {
		totalOutputs = 1
	}
	outputsGenerated := 0

	fmt.Printf("Total outputs to generate: %d\n", totalOutputs)
	_ = os.Stdout.Sync()

	// 1. Save combined CSV
	combinedCSVPath := filepath.Join(combinedDir, "isx_combined_data.csv")
	headers := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}

	var csvRecords [][]string
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
		csvRecords = append(csvRecords, row)
	}

	if err := csvWriter.WriteSimpleCSV(combinedCSVPath, headers, csvRecords); err != nil {
		return fmt.Errorf("failed to write combined CSV: %w", err)
	}
	logger.Info("Saved combined report", "path", combinedCSVPath)
	outputsGenerated++
	fmt.Printf("Generated output %d of %d: %s\n", outputsGenerated, totalOutputs, combinedCSVPath)
	_ = os.Stdout.Sync()

	// 2. Generate daily files
	recordsByDate := make(map[string][]domain.TradeRecord)
	for _, record := range records {
		dateStr := record.Date.Format("2006_01_02")
		recordsByDate[dateStr] = append(recordsByDate[dateStr], record)
	}

	for dateStr, dailyRecords := range recordsByDate {
		dailyCSVPath := filepath.Join(dailyDir, fmt.Sprintf("isx_daily_%s.csv", dateStr))

		var dailyRecordsStr [][]string
		for _, record := range dailyRecords {
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
			dailyRecordsStr = append(dailyRecordsStr, row)
		}

		if err := csvWriter.WriteSimpleCSV(dailyCSVPath, headers, dailyRecordsStr); err != nil {
			logger.Error("Error saving daily CSV",
				slog.String("path", dailyCSVPath),
				slog.String("error", err.Error()))
		} else {
			outputsGenerated++
			fmt.Printf("Generated output %d of %d: %s\n", outputsGenerated, totalOutputs, dailyCSVPath)
			_ = os.Stdout.Sync()
		}
	}

	// 3. Generate ticker files
	tickers := make(map[string]bool)
	for _, record := range records {
		tickers[record.CompanySymbol] = true
	}

	for ticker := range tickers {
		tickerCSVPath := filepath.Join(tickerDir, fmt.Sprintf("%s_trading_history.csv", ticker))

		var tickerRecords []domain.TradeRecord
		for _, record := range records {
			if record.CompanySymbol == ticker {
				tickerRecords = append(tickerRecords, record)
			}
		}

		var tickerRecordsStr [][]string
		for _, record := range tickerRecords {
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
			tickerRecordsStr = append(tickerRecordsStr, row)
		}

		if err := csvWriter.WriteSimpleCSV(tickerCSVPath, headers, tickerRecordsStr); err != nil {
			logger.Error("Error saving ticker CSV",
				slog.String("ticker", ticker),
				slog.String("path", tickerCSVPath),
				slog.String("error", err.Error()))
		} else {
			outputsGenerated++
			fmt.Printf("Generated output %d of %d: %s\n", outputsGenerated, totalOutputs, tickerCSVPath)
			_ = os.Stdout.Sync()
		}
	}

	// 4. Generate ticker summary files
	logger.Info("Generating ticker summary files...")

	// Initialize summarizer
	summarizer := operations.NewSummarizer(logger, operations.SummarizerConfig{
		IncludeExtendedMetrics: true,
		MaxLast10Days:          10,
		DateFormat:             "2006-01-02",
	})

	// Generate summaries from records
	summaries, err := summarizer.GenerateFromRecords(context.Background(), records)
	if err != nil {
		logger.Error("Failed to generate ticker summaries", "error", err)
		return fmt.Errorf("failed to generate ticker summaries: %w", err)
	}

	// Create summary directory
	summaryDir := filepath.Join(outDir, "summary", "ticker")
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create summary directory: %w", err)
	}

	// Write CSV summary
	summaryCSVPath := filepath.Join(summaryDir, "ticker_summary.csv")
	if err := summarizer.WriteCSV(context.Background(), summaryCSVPath, summaries); err != nil {
		logger.Error("Failed to write ticker summary CSV", "error", err)
		return fmt.Errorf("failed to write ticker summary CSV: %w", err)
	}
	logger.Info("Generated ticker summary CSV", "path", summaryCSVPath)

	// Write JSON summary
	summaryJSONPath := filepath.Join(summaryDir, "ticker_summary.json")
	if err := summarizer.WriteJSON(context.Background(), summaryJSONPath, summaries); err != nil {
		logger.Error("Failed to write ticker summary JSON", "error", err)
		return fmt.Errorf("failed to write ticker summary JSON: %w", err)
	}
	logger.Info("Generated ticker summary JSON", "path", summaryJSONPath)

	fmt.Printf("Generated ticker summary files: %s and %s\n", summaryCSVPath, summaryJSONPath)
	_ = os.Stdout.Sync()

	return nil
}

// count helpers for telemetry
func uniqueDates(records []domain.TradeRecord) map[string]struct{} {
	result := make(map[string]struct{})
	for _, r := range records {
		result[r.Date.Format("2006_01_02")] = struct{}{}
	}
	return result
}

func uniqueSymbols(records []domain.TradeRecord) map[string]struct{} {
	result := make(map[string]struct{})
	for _, r := range records {
		result[r.CompanySymbol] = struct{}{}
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
