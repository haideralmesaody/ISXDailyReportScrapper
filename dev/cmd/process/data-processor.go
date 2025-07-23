package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/analytics"
	"isxcli/internal/common"
	"isxcli/internal/exporter"
	"isxcli/internal/files"
	"isxcli/internal/parser"
	"isxcli/internal/processor"
	"isxcli/internal/progress"
)

var logger *common.Logger

// ExcelFileInfo holds information about an Excel file
type ExcelFileInfo struct {
	Name string
	Date time.Time
}

// sendProgress sends a structured progress message using the progress calculator
func sendProgress(calc *progress.EnhancedCalculator, message string, details map[string]interface{}) {
	jsonData, err := calc.ToJSON(message, details)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create progress message: %v\n", err)
		return
	}
	fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonData)
}

// sendStatus sends a structured status message
func sendStatus(stage, status, message string) {
	jsonData, err := progress.StatusToJSON(stage, status, message)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create status message: %v\n", err)
		return
	}
	fmt.Printf("[WEBSOCKET_STATUS] %s\n", jsonData)
}

// sendError sends a structured error message
func sendError(code, message, details, stage string, recoverable bool, hint string) {
	jsonData, err := progress.ErrorToJSON(code, message, details, stage, recoverable, hint)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create error message: %v\n", err)
		return
	}
	fmt.Printf("[WEBSOCKET_ERROR] %s\n", jsonData)
}

func main() {
	// Initialize logger
	logger = common.NewLoggerWithComponent("processor")
	logger.LogStart(common.CategoryPipeline, "ISX Data Processor")
	
	inDir := flag.String("in", "data/downloads", "input directory for .xlsx files")
	outDir := flag.String("out", "data/reports", "output directory for CSV files")
	fullRework := flag.Bool("full", false, "force full rework of all files")
	flag.Parse()
	
	logger.DebugCategory(common.CategorySystem, "Command line args: in=%s, out=%s, full=%v", 
		*inDir, *outDir, *fullRework)

	// Create file manager instance
	fileManager := files.NewManager(".")
	logger.DebugCategory(common.CategoryFile, "File manager initialized")
	
	// Create output directory if it doesn't exist
	logger.DebugCategory(common.CategoryFile, "Creating output directory: %s", *outDir)
	if err := fileManager.CreateDirectory(*outDir); err != nil {
		logger.ErrorCategory(common.CategoryFile, "Failed to create output directory: %v", err)
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	logger.DebugCategory(common.CategoryFile, "Output directory ready")

	fmt.Printf("[INIT] ISX Daily Reports Processor Starting\n")
	fmt.Printf("[INIT] Input Directory: %s\n", *inDir)
	fmt.Printf("[INIT] Output Directory: %s\n", *outDir)
	fmt.Printf("[INIT] Full Rework Mode: %v\n", *fullRework)
	
	// Send initial status
	sendStatus("processing", "active", "Starting data processing...")
	
	startTime := time.Now()

	// Create file discovery instance
	discovery := files.NewDiscovery(".")
	logger.LogStart(common.CategoryFile, "Excel file discovery")
	
	// Get all available Excel files
	discoveredFiles, err := discovery.FindExcelFiles(*inDir)
	if err != nil {
		logger.ErrorCategory(common.CategoryFile, "Failed to discover Excel files: %v", err)
		fmt.Printf("failed to discover Excel files: %v\n", err)
		os.Exit(1)
	}
	logger.InfoCategory(common.CategoryFile, "Discovered %d Excel files", len(discoveredFiles))

	// Parse and sort all available files by date
	var excelFiles []ExcelFileInfo
	for _, file := range discoveredFiles {
		// Skip temporary Excel files
		if strings.HasPrefix(file.Name, "~$") {
			continue
		}

		// Extract date from filename (e.g., "YYYY MM DD ISX Daily Report.xlsx")
		parts := strings.Split(file.Name, " ")
		if len(parts) < 4 {
			continue // Skip malformed filenames
		}

		dateStr := strings.Join(parts[0:3], " ")
		date, err := time.Parse("2006 01 02", dateStr)
		if err != nil {
			fmt.Printf("Warning: Could not parse date from filename %s: %v\n", file.Name, err)
			continue
		}

		excelFiles = append(excelFiles, ExcelFileInfo{
			Name: file.Name,
			Date: date,
		})
	}

	// Sort files by date
	sort.Slice(excelFiles, func(i, j int) bool {
		return excelFiles[i].Date.Before(excelFiles[j].Date)
	})

	fmt.Printf("[DISCOVERY] %d Excel files discovered\n", len(excelFiles))

	// Check what needs to be processed
	var filesToProcess []ExcelFileInfo
	var existingRecords []parser.TradeRecord

	logger.LogDecision(common.CategoryPipeline, "Processing mode", fmt.Sprintf("fullRework=%v", *fullRework))
	
	if *fullRework {
		logger.InfoCategory(common.CategoryPipeline, "Full rework mode - will process all files")
		fmt.Printf("[MODE] Full rework requested - processing all files\n")
		filesToProcess = excelFiles
	} else {
		// Smart update: check what's already processed
		logger.InfoCategory(common.CategoryPipeline, "Smart update mode - checking existing processed files")
		fmt.Printf("[MODE] Smart update mode - checking existing files\n")
		filesToProcess, existingRecords = determineFilesToProcess(excelFiles, *outDir)
		logger.InfoCategory(common.CategoryPipeline, "Smart update: %d new files to process, %d existing records loaded", 
			len(filesToProcess), len(existingRecords))
		fmt.Printf("[MODE] Smart update: %d files need processing\n", len(filesToProcess))
	}

	if len(filesToProcess) == 0 {
		fmt.Printf("[COMPLETE] No files need processing - all data is up to date\n")
		fmt.Printf("[SUMMARY] Processing completed in %.1f seconds\n", time.Since(startTime).Seconds())
		
		// Send completion status
		sendStatus("processing", "completed", "All files are already up to date")
		
		// Ensure all output is flushed before exiting
		os.Stdout.Sync()
		return
	}

	// Process the required files
	var newRecords []parser.TradeRecord
	totalFiles := len(filesToProcess)
	
	// Initialize metrics manager for historical data
	dataPath := filepath.Dir(*outDir) // Parent of reports directory
	metricsManager := progress.NewMetricsManager(dataPath)
	
	// Create enhanced calculator with historical metrics support
	calc := progress.NewEnhancedCalculator("processing", totalFiles, metricsManager)
	
	fmt.Printf("[PROCESSING] Starting to process %d files...\n", totalFiles)

	for i, fileInfo := range filesToProcess {
		fileStartTime := time.Now()
		
		// Update calculator
		calc.Update(i)
		
		logger.LogProgress(common.CategoryData, "Processing files", i+1, totalFiles)
		logger.DebugCategory(common.CategoryData, "Processing file: %s (date: %s)", 
			fileInfo.Name, fileInfo.Date.Format("2006-01-02"))
		
		fmt.Printf("[PROGRESS] File %d/%d (%.1f%%): %s\n", i+1, totalFiles, calc.GetProgress(), fileInfo.Name)
		
		// Send structured progress with enhanced ETA
		details := map[string]interface{}{
			"current_file": fileInfo.Name,
			"file_date": fileInfo.Date.Format("2006-01-02"),
			"processed_files": i,
			"total_files": totalFiles,
		}
		sendProgress(calc, fmt.Sprintf("Processing file: %s", fileInfo.Name), details)

		filePath := filepath.Join(*inDir, fileInfo.Name)
		logger.DebugCategory(common.CategoryData, "Parsing Excel file: %s", filePath)
		
		report, err := parser.ParseFile(filePath)
		if err != nil {
			logger.ErrorCategory(common.CategoryData, "Failed to parse file %s: %v", fileInfo.Name, err)
			fmt.Printf("[ERROR] Failed to parse file %s: %v\n", fileInfo.Name, err)
			
			// Send error message
			sendError("PARSE_ERROR", fmt.Sprintf("Failed to parse file %s", fileInfo.Name),
				err.Error(), "processing", true, 
				"Check if the Excel file is corrupted or has an unexpected format")
			
			continue
		}
		
		logger.DebugCategory(common.CategoryData, "Successfully parsed %d records from %s", 
			len(report.Records), fileInfo.Name)
		
		fileProcessTime := time.Since(fileStartTime)
		fmt.Printf("[SUCCESS] File processed in %.1f seconds - %d records found\n", fileProcessTime.Seconds(), len(report.Records))

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
	var filledRecords []parser.TradeRecord
	if len(allRecords) > 0 {
		logger.LogStart(common.CategoryData, "forward-fill processing")
		logger.DebugCategory(common.CategoryData, "Total records before forward-fill: %d", len(allRecords))
		
		fmt.Printf("[PROCESSING] Generating dataset with forward-fill algorithm...\n")
		fillStartTime := time.Now()
		
		// Create a new calculator for forward-fill phase
		fillCalc := progress.NewEnhancedCalculator("processing", 100, metricsManager)
		
		// Send progress for forward-fill
		sendProgress(fillCalc, "Applying forward-fill algorithm to dataset",
			map[string]interface{}{
				"phase": "forward-fill",
				"total_records": len(allRecords),
			})
		
		// Use the new processor
		fillProcessor := processor.NewForwardFillProcessor()
		logger.DebugCategory(common.CategoryData, "Starting forward-fill algorithm")
		filledRecords, stats := fillProcessor.FillMissingDataWithStats(allRecords)
		fillTime := time.Since(fillStartTime)

		fmt.Printf("[RESULTS] Forward-fill completed in %.1f seconds\n", fillTime.Seconds())
		fmt.Printf("[RESULTS] Total records processed: %d\n", stats.TotalRecords)
		fmt.Printf("[RESULTS] Active trading records: %d\n", stats.ActiveRecords)
		fmt.Printf("[RESULTS] Forward-filled records: %d\n", stats.ForwardFilledCount)
		
		// Update progress to show forward-fill completed
		fillCalc.Update(100)
		sendProgress(fillCalc, "Forward-fill completed",
			map[string]interface{}{
				"phase": "forward-fill",
				"total_records": stats.TotalRecords,
				"active_records": stats.ActiveRecords,
				"forward_filled_count": stats.ForwardFilledCount,
			})

		// Save combined CSV with forward-fill
		fmt.Printf("[OUTPUT] Saving combined CSV file...\n")
		combinedCSVPath := filepath.Join(*outDir, "isx_combined_data.csv")
		combinedStartTime := time.Now()
		
		// Use the new exporter
		dailyExporter := exporter.NewDailyExporter(".")
		if err := dailyExporter.ExportCombinedData(filledRecords, combinedCSVPath); err != nil {
			fmt.Printf("[ERROR] Failed to save combined CSV: %v\n", err)
		} else {
			combinedTime := time.Since(combinedStartTime)
			fmt.Printf("[SUCCESS] Combined CSV saved in %.1f seconds: %s\n", combinedTime.Seconds(), combinedCSVPath)
		}

		// Generate daily CSV files with forward-fill
		fmt.Printf("[OUTPUT] Generating daily CSV files with forward-fill...\n")
		dailyStartTime := time.Now()
		
		// Get existing daily files to avoid regenerating
		discovery := files.NewDiscovery(".")
		existingDates := make(map[string]bool)
		if dailyFiles, err := discovery.FindDailyCSVFiles(*outDir); err == nil {
			for dateStr := range dailyFiles {
				existingDates[dateStr] = true
			}
		}
		
		if err := dailyExporter.ExportDailyReportsStreaming(filledRecords, *outDir, existingDates); err != nil {
			fmt.Printf("[ERROR] Failed to generate daily files: %v\n", err)
		} else {
			dailyTime := time.Since(dailyStartTime)
			fmt.Printf("[SUCCESS] Daily files generated in %.1f seconds\n", dailyTime.Seconds())
		}

		// Generate individual ticker CSV files with forward-fill
		fmt.Printf("[OUTPUT] Generating individual ticker CSV files with forward-fill...\n")
		tickerStartTime := time.Now()
		
		tickerExporter := exporter.NewTickerExporter(".")
		if err := tickerExporter.ExportTickerFiles(filledRecords, *outDir); err != nil {
			fmt.Printf("[ERROR] Failed to generate ticker files: %v\n", err)
		} else {
			tickerTime := time.Since(tickerStartTime)
			fmt.Printf("[SUCCESS] Ticker files generated in %.1f seconds\n", tickerTime.Seconds())
		}
	}

	fmt.Println("Processing complete.")

	// Record completion metrics for future ETA predictions
	if err := calc.Complete(); err != nil {
		fmt.Printf("[WARN] Failed to save metrics: %v\n", err)
	}
	
	// Send completion status
	totalTime := time.Since(startTime)
	sendStatus("processing", "completed", 
		fmt.Sprintf("Processing completed: %d files processed in %.1f minutes", 
			len(filesToProcess), totalTime.Minutes()))
	
	// Send final progress
	calc.Update(len(filesToProcess))
	sendProgress(calc, "All files processed successfully",
		map[string]interface{}{
			"total_files": len(filesToProcess),
			"new_records": len(newRecords),
			"duration": totalTime.String(),
			"average_time_per_file": totalTime.Seconds() / float64(len(filesToProcess)),
		})

	// Generate ticker summary for web interface
	logger.LogStart(common.CategoryData, "ticker summary generation")
	fmt.Printf("[SUMMARY] Generating ticker summary for web interface...\n")
	summaryStartTime := time.Now()
	
	// Use the new analytics package
	summaryGenerator := analytics.NewSummaryGenerator(".")
	combinedFile := "data/reports/isx_combined_data.csv"
	summaryFile := "data/reports/ticker_summary.csv"
	
	logger.DebugCategory(common.CategoryData, "Generating summary from: %s", combinedFile)
	if err := summaryGenerator.GenerateFromCombinedCSV(combinedFile, summaryFile); err != nil {
		logger.ErrorCategory(common.CategoryData, "Failed to generate ticker summary: %v", err)
		fmt.Printf("[ERROR] Failed to generate ticker summary: %v\n", err)
	} else {
		summaryTime := time.Since(summaryStartTime)
		logger.LogComplete(common.CategoryData, "ticker summary generation")
		logger.LogTiming(common.CategoryData, "Ticker summary generation", summaryStartTime)
		fmt.Printf("[SUCCESS] Ticker summary generated in %.1f seconds\n", summaryTime.Seconds())
	}
	
	// Final processing summary
	totalProcessingTime := time.Since(startTime)
	logger.LogComplete(common.CategoryPipeline, "data processing")
	logger.InfoCategory(common.CategoryPipeline, "Total processing time: %.1f minutes", totalProcessingTime.Minutes())
	
	fmt.Printf("\n[COMPLETE] ====== Processing Summary ======\n")
	fmt.Printf("[COMPLETE] Total Files Processed: %d\n", totalFiles)
	fmt.Printf("[COMPLETE] Total Processing Time: %.1f minutes\n", totalProcessingTime.Minutes())
	fmt.Printf("[COMPLETE] Average Time per File: %.1f seconds\n", totalProcessingTime.Seconds()/float64(totalFiles))
	fmt.Printf("[COMPLETE] Total Records Generated: %d\n", len(allRecords))
	
	if len(filledRecords) > 0 {
		logger.InfoCategory(common.CategoryData, "Data quality: %.1f%% active trading records", 
			float64(len(allRecords))/float64(len(filledRecords))*100)
		fmt.Printf("[COMPLETE] Forward-filled Records: %d\n", len(filledRecords)-len(allRecords))
		fmt.Printf("[COMPLETE] Data Quality: %.1f%% active trading records\n", 
			float64(len(allRecords))/float64(len(filledRecords))*100)
	}
	
	fmt.Printf("[COMPLETE] ================================\n")
	
	logger.InfoCategory(common.CategoryPipeline, "All processing completed successfully")
	
	// Ensure all output is flushed before exiting
	os.Stdout.Sync()
}

// determineFilesToProcess checks which files need to be processed based on existing CSV files
func determineFilesToProcess(excelFiles []ExcelFileInfo, outDir string) ([]ExcelFileInfo, []parser.TradeRecord) {
	var filesToProcess []ExcelFileInfo
	var existingRecords []parser.TradeRecord

	// Create discovery instance for finding existing files
	discovery := files.NewDiscovery(".")
	
	// Check which daily CSV files already exist
	existingDates := make(map[string]bool)
	if dailyFiles, err := discovery.FindDailyCSVFiles(outDir); err == nil {
		for dateStr := range dailyFiles {
			existingDates[dateStr] = true
		}
	}

	fmt.Printf("Found %d existing daily CSV files\n", len(existingDates))

	// Load existing records from combined CSV if it exists
	fileManager := files.NewManager(".")
	combinedCSVPath := filepath.Join(outDir, "isx_combined_data.csv")
	if fileManager.FileExists(combinedCSVPath) {
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

	// Read file content to handle BOM
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}
	
	// Remove BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}
	
	// Create CSV reader from cleaned content
	reader := csv.NewReader(strings.NewReader(string(content)))
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

// formatTradingStatus formats boolean as "true"/"false" string
func formatTradingStatus(status bool) string {
	if status {
		return "true"
	}
	return "false"
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
			formatTradingStatus(record.TradingStatus),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
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
			formatTradingStatus(record.TradingStatus),
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
	fileManager := files.NewManager(".")
	if err := fileManager.CreateDirectory(outDir); err != nil {
		return err
	}

	// Generate CSV files for each date
	count := 0
	total := len(recordsByDate)
	
	for dateStr, dailyRecords := range recordsByDate {
		count++
		fmt.Printf("[DAILY] Processing date %d/%d: %s (%d records)\n", count, total, dateStr, len(dailyRecords))

		// Save CSV for the current date
		dailyCSVPath := filepath.Join(outDir, fmt.Sprintf("isx_daily_%s.csv", dateStr))
		if err := saveDailyCSV(dailyCSVPath, dailyRecords); err != nil {
			fmt.Printf("[ERROR] Failed to save daily CSV for %s: %v\n", dateStr, err)
		} else {
			fmt.Printf("[SUCCESS] Daily CSV saved: %s\n", dailyCSVPath)
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
	fileManager := files.NewManager(".")
	if err := fileManager.CreateDirectory(outDir); err != nil {
		return err
	}

	// Generate CSV files for each ticker
	count := 0
	total := len(tickers)
	
	for ticker := range tickers {
		count++
		
		// Filter records for the current ticker
		var tickerRecords []parser.TradeRecord
		for _, record := range records {
			if record.CompanySymbol == ticker {
				tickerRecords = append(tickerRecords, record)
			}
		}
		
		fmt.Printf("[TICKER] Processing %d/%d: %s (%d records)\n", count, total, ticker, len(tickerRecords))

		// Save CSV for the current ticker
		tickerCSVPath := filepath.Join(outDir, fmt.Sprintf("%s_trading_history.csv", ticker))
		if err := saveDailyCSV(tickerCSVPath, tickerRecords); err != nil {
			fmt.Printf("[ERROR] Failed to save ticker CSV for %s: %v\n", ticker, err)
		} else {
			fmt.Printf("[SUCCESS] Ticker CSV saved: %s\n", tickerCSVPath)
		}
	}

	return nil
}

