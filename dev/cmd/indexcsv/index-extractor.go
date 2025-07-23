package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/common"
	"isxcli/internal/progress"
	"github.com/xuri/excelize/v2"
)

var logger *common.Logger

// regex for filenames like "2025 06 24 ISX Daily Report.xlsx"
var fileRe = regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)

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
	logger = common.NewLoggerWithComponent("indexcsv")
	logger.LogStart(common.CategoryPipeline, "ISX Index Extractor")
	
	mode := flag.String("mode", "initial", "initial | accumulative")
	dir := flag.String("dir", "data/downloads", "directory containing xlsx reports")
	out := flag.String("out", "indexes.csv", "output csv file path")
	flag.Parse()
	
	logger.DebugCategory(common.CategorySystem, "Command line args: mode=%s, dir=%s, out=%s", 
		*mode, *dir, *out)

	startTime := time.Now()
	
	fmt.Printf("[INIT] ISX Index Extractor Starting\n")
	fmt.Printf("[INIT] Mode: %s\n", *mode)
	fmt.Printf("[INIT] Input Directory: %s\n", *dir)
	fmt.Printf("[INIT] Output File: %s\n", *out)
	
	// Send initial status
	sendStatus("indices", "active", "Starting index extraction...")

	var lastDate time.Time
	logger.LogDecision(common.CategoryPipeline, "Processing mode", fmt.Sprintf("mode=%s", *mode))
	
	if *mode == "accumulative" {
		logger.DebugCategory(common.CategoryPipeline, "Checking for existing CSV file: %s", *out)
		if d, err := loadLastDate(*out); err == nil {
			lastDate = d
			logger.InfoCategory(common.CategoryPipeline, "Accumulative mode: last processed date %s", lastDate.Format("2006-01-02"))
			fmt.Printf("[MODE] Accumulative mode: Last processed date %s\n", lastDate.Format("2006-01-02"))
		} else {
			logger.WarnCategory(common.CategoryPipeline, "No existing CSV found, switching to initial mode: %v", err)
			fmt.Printf("[MODE] No existing CSV found, switching to initial mode\n")
			*mode = "initial"
		}
	}

	if *mode == "initial" {
		// initial mode: create/truncate csv with header
		fmt.Printf("[SETUP] Creating new CSV file with headers...\n")
		f, err := os.Create(*out)
		if err != nil {
			fmt.Printf("[ERROR] Cannot create %s: %v\n", *out, err)
			os.Exit(1)
		}
		w := csv.NewWriter(f)
		w.Write([]string{"Date", "ISX60", "ISX15"})
		w.Flush()
		_ = f.Close()
		fmt.Printf("[SUCCESS] Created new CSV file: %s\n", *out)
	}

	logger.LogStart(common.CategoryFile, "directory scanning")
	fmt.Printf("[DISCOVERY] Scanning directory for Excel files...\n")
	entries, err := os.ReadDir(*dir)
	if err != nil {
		logger.ErrorCategory(common.CategoryFile, "Failed to read directory %s: %v", *dir, err)
		fmt.Printf("[ERROR] Failed to read directory: %v\n", err)
		os.Exit(1)
	}
	logger.DebugCategory(common.CategoryFile, "Found %d directory entries", len(entries))

	type fileInfo struct {
		path string
		date time.Time
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := fileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, _ := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		if !lastDate.IsZero() && !t.After(lastDate) {
			continue // already processed
		}
		files = append(files, fileInfo{path: filepath.Join(*dir, e.Name()), date: t})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].date.Before(files[j].date) })

	fmt.Printf("[DISCOVERY] Found %d Excel files to process\n", len(files))
	if len(files) == 0 {
		fmt.Printf("[COMPLETE] No new files to process - all data is up to date\n")
		sendStatus("indices", "completed", "All indices are already up to date")
		
		// Ensure all output is flushed before exiting
		os.Stdout.Sync()
		return
	}

	outF, err := os.OpenFile(*out, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to open output file: %v\n", err)
		os.Exit(1)
	}
	defer outF.Close()
	writer := csv.NewWriter(outF)

	processedCount := 0
	fmt.Printf("[PROCESSING] Starting index extraction from %d files...\n", len(files))
	
	// Initialize metrics manager for historical data
	dataPath := filepath.Dir(*out) // Parent of output file
	if dataPath == "" {
		dataPath = "."
	}
	metricsManager := progress.NewMetricsManager(dataPath)
	
	// Create enhanced calculator with historical metrics support
	calc := progress.NewEnhancedCalculator("indices", len(files), metricsManager)
	
	for i, fi := range files {
		fileStartTime := time.Now()
		
		// Update calculator
		calc.Update(i)
		
		logger.LogProgress(common.CategoryData, "Extracting indices", i+1, len(files))
		logger.DebugCategory(common.CategoryData, "Processing file: %s (date: %s)", 
			filepath.Base(fi.path), fi.date.Format("2006-01-02"))
		
		fmt.Printf("[PROGRESS] File %d/%d (%.1f%%): %s\n", i+1, len(files), calc.GetProgress(), filepath.Base(fi.path))
		
		// Send structured progress with enhanced ETA
		details := map[string]interface{}{
			"current_file": filepath.Base(fi.path),
			"file_date": fi.date.Format("2006-01-02"),
			"indices_extracted": i,
			"total_files": len(files),
		}
		sendProgress(calc, fmt.Sprintf("Extracting indices from %s", filepath.Base(fi.path)), details)

		logger.DebugCategory(common.CategoryData, "Extracting indices from Excel file")
		isx60, isx15, err := extractIndices(fi.path)
		if err != nil {
			logger.ErrorCategory(common.CategoryData, "Failed to extract indices from %s: %v", filepath.Base(fi.path), err)
			fmt.Printf("[ERROR] Failed to extract indices from %s: %v\n", filepath.Base(fi.path), err)
			
			// Send error message
			sendError("INDEX_EXTRACT_ERROR", 
				fmt.Sprintf("Failed to extract indices from %s", filepath.Base(fi.path)),
				err.Error(), "indices", true, 
				"Check if the Excel file contains ISX60/ISX15 index data")
			
			continue
		}
		
		logger.DebugCategory(common.CategoryData, "Successfully extracted indices: ISX60=%.2f, ISX15=%.2f", isx60, isx15)

		rec := []string{fi.date.Format("2006-01-02"), formatFloat(isx60)}
		if isx15 > 0 {
			rec = append(rec, formatFloat(isx15))
		} else {
			rec = append(rec, "")
		}
		writer.Write(rec)
		processedCount++
		
		fileProcessTime := time.Since(fileStartTime)
		if isx15 > 0 {
			fmt.Printf("[SUCCESS] Processed in %.1f seconds - Date: %s | ISX60: %.2f | ISX15: %.2f\n", 
				fileProcessTime.Seconds(), fi.date.Format("2006-01-02"), isx60, isx15)
		} else {
			fmt.Printf("[SUCCESS] Processed in %.1f seconds - Date: %s | ISX60: %.2f | ISX15: N/A\n", 
				fileProcessTime.Seconds(), fi.date.Format("2006-01-02"), isx60)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Printf("[ERROR] Failed to write CSV: %v\n", err)
		os.Exit(1)
	}

	// Final processing summary
	totalProcessingTime := time.Since(startTime)
	logger.LogComplete(common.CategoryPipeline, "index extraction")
	logger.InfoCategory(common.CategoryPipeline, "Total processing time: %.1f minutes", totalProcessingTime.Minutes())
	logger.InfoCategory(common.CategoryData, "Successfully extracted indices from %d files", processedCount)
	
	fmt.Printf("\n[COMPLETE] ====== Index Extraction Summary ======\n")
	fmt.Printf("[COMPLETE] Total Files Processed: %d\n", processedCount)
	fmt.Printf("[COMPLETE] Total Processing Time: %.1f minutes\n", totalProcessingTime.Minutes())
	if processedCount > 0 {
		fmt.Printf("[COMPLETE] Average Time per File: %.1f seconds\n", totalProcessingTime.Seconds()/float64(processedCount))
	}
	fmt.Printf("[COMPLETE] Output File: %s\n", *out)
	fmt.Printf("[COMPLETE] =========================================\n")
	
	// Record completion metrics for future ETA predictions
	if err := calc.Complete(); err != nil {
		fmt.Printf("[WARN] Failed to save metrics: %v\n", err)
	}
	
	// Send completion status
	completionMessage := fmt.Sprintf("Index extraction completed: %d files processed", processedCount)
	sendStatus("indices", "completed", completionMessage)
	
	// Send final progress
	calc.Update(processedCount)
	sendProgress(calc, completionMessage,
		map[string]interface{}{
			"files_processed": processedCount,
			"output_file": *out,
			"duration": totalProcessingTime.String(),
			"average_time_per_file": totalProcessingTime.Seconds() / float64(processedCount),
		})
		
	// Ensure all output is flushed before exiting
	os.Stdout.Sync()
}

func loadLastDate(csvPath string) (time.Time, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	var last string
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return time.Time{}, err
		}
		if rec[0] == "Date" {
			continue
		}
		last = rec[0]
	}
	if last == "" {
		return time.Time{}, fmt.Errorf("no data rows")
	}
	t, err := time.Parse("2006-01-02", last)
	return t, err
}

func extractIndices(path string) (isx60, isx15 float64, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Build list of sheets to inspect: prefer "Indices" if exists, otherwise all
	var sheets []string
	hasIndices := false
	for _, sh := range f.GetSheetList() {
		if strings.EqualFold(sh, "indices") {
			hasIndices = true
			break
		}
	}
	if hasIndices {
		sheets = []string{"Indices"}
	} else {
		sheets = f.GetSheetList()
	}

	joinRe := regexp.MustCompile(`\s+`)
	for _, sheet := range sheets {
		rows, _ := f.GetRows(sheet)
		for _, row := range rows {
			line := strings.TrimSpace(joinRe.ReplaceAllString(strings.Join(row, " "), " "))
			if line == "" {
				continue
			}
			// Case 1: Both 60 and 15 on the same line
			if strings.Contains(line, "ISX Index 60") && strings.Contains(line, "ISX Index 15") {
				numRe := regexp.MustCompile(`ISX Index 60\s+([0-9.,]+).*?ISX Index 15\s+([0-9.,]+)`) // non-greedy
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1])
					isx15, _ = parseFloat(m[2])
					return isx60, isx15, nil
				}
			}

			// Case 2: Only 60 present (older reports)
			if strings.Contains(line, "ISX Index 60") {
				numRe := regexp.MustCompile(`ISX Index 60\s+([0-9.,]+)`)
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1])
					return isx60, 0, nil
				}
			}

			// Case 3: Very old format – "ISX Price Index"
			if strings.Contains(line, "ISX Price Index") {
				numRe := regexp.MustCompile(`ISX Price Index\s+([0-9.,]+)`)
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1]) // treat as 60 index
					return isx60, 0, nil
				}
			}
		}
	}
	return 0, 0, fmt.Errorf("ISX indices not found in any sheet of %s", filepath.Base(path))
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
