package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"isxcli/internal/common"
	"isxcli/internal/license"
	"isxcli/internal/progress"

	"github.com/chromedp/chromedp"
)

var logger *common.Logger

const (
	baseURL  = "http://www.isx-iq.net"
	startURL = "http://www.isx-iq.net/isxportal/portal/uploadedFilesList.html?currLanguage=en"
)

// sendProgress sends a structured progress message using the progress calculator
func sendProgress(calc *progress.EnhancedCalculator, message string, details map[string]interface{}) {
	jsonData, err := calc.ToJSON(message, details)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create progress message: %v\n", err)
		return
	}
	fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonData)
	// Force flush immediately
	os.Stdout.Sync()
}

// sendStatus sends a structured status message
func sendStatus(stage, status, message string) {
	jsonData, err := progress.StatusToJSON(stage, status, message)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create status message: %v\n", err)
		return
	}
	
	fmt.Printf("[WEBSOCKET_STATUS] %s\n", jsonData)
	// Force flush immediately
	os.Stdout.Sync()
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
	logger = common.NewLoggerWithComponent("scraper")
	
	mode := flag.String("mode", "initial", "scrape mode: initial | accumulative")
	fromStr := flag.String("from", "", "start date (YYYY-MM-DD) - REQUIRED when called from web interface")
	toStr := flag.String("to", "", "optional end date (YYYY-MM-DD); leave blank to keep site default")
	outDir := flag.String("out", "data/downloads", "directory to save reports")
	headless := flag.Bool("headless", true, "run browser headless")
	flag.Parse()
	
	logger.LogStart(common.CategoryPipeline, "ISX Daily Reports Scraper")
	logger.DebugCategory(common.CategorySystem, "Command line args: mode=%s, out=%s, from=%s, to=%s, headless=%v", 
		*mode, *outDir, *fromStr, *toStr, *headless)

	// Initialize license system
	fmt.Println("🔐 ISX Daily Reports Scraper - Licensed Version")
	fmt.Println("═══════════════════════════════════════════════")

	logger.LogStart(common.CategoryLicense, "license validation")
	if !checkLicense() {
		logger.ErrorCategory(common.CategoryLicense, "License validation failed")
		fmt.Println("❌ License validation failed. Application will exit.")
		fmt.Println("📞 Contact The Iraqi Investor Group to get a new license.")
		os.Exit(1)
	}
	logger.LogComplete(common.CategoryLicense, "license validation")

	// Create output directory if it doesn't exist (but don't delete existing files)
	logger.DebugCategory(common.CategoryFile, "Creating output directory: %s", *outDir)
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		logger.ErrorCategory(common.CategoryFile, "Failed to create output directory: %v", err)
		fmt.Printf("failed to create output dir: %v\n", err)
		os.Exit(1)
	}
	logger.DebugCategory(common.CategoryFile, "Output directory ready: %s", *outDir)

	// determine fromSite depending on mode
	var fromSite string
	logger.LogDecision(common.CategoryPipeline, "Determining date range", fmt.Sprintf("mode=%s", *mode))
	
	if *mode == "accumulative" {
		logger.DebugCategory(common.CategoryPipeline, "Accumulative mode - scanning for latest downloaded file")
		// scan downloads for latest file
		if d, ok := latestDownloadedDate(*outDir); ok {
			fromSite = d.AddDate(0, 0, 1).Format("02/01/2006") // next day
			logger.InfoCategory(common.CategoryPipeline, "Last report date: %s, will start from: %s", 
				d.Format("2006-01-02"), fromSite)
			fmt.Printf("[MODE accumulative] Detected last report date %s. Will start from %s.\n", d.Format("2006-01-02"), fromSite)
		} else {
			logger.WarnCategory(common.CategoryPipeline, "No existing files found in accumulative mode")
		}
	}

	if fromSite == "" {
		// HTML form is the single source of truth - require from date when not in accumulative mode
		if *fromStr == "" {
			fmt.Printf("ERROR: --from date is required for initial mode\n")
			fmt.Printf("       The web interface should always provide this date\n")
			fmt.Printf("       HTML form is the single source of truth for date ranges\n")
			os.Exit(1)
		}
		
		startDate, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			fmt.Printf("invalid --from date format: %v\n", err)
			os.Exit(1)
		}
		
		fromSite = startDate.Format("02/01/2006")
		fmt.Printf("[MODE initial] Using from date provided by web interface: %s\n", startDate.Format("2006-01-02"))
	}

	var toSite string
	if *toStr != "" {
		endDate, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			fmt.Printf("invalid --to date: %v\n", err)
			os.Exit(1)
		}
		toSite = endDate.Format("02/01/2006")
	}

	// setup ChromeDP
	logger.LogStart(common.CategorySystem, "ChromeDP setup")
	opts := chromedp.DefaultExecAllocatorOptions[:]
	if *headless {
		opts = append(opts, chromedp.Flag("headless", true))
		logger.DebugCategory(common.CategorySystem, "Running in headless mode")
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
		logger.DebugCategory(common.CategorySystem, "Running in headed mode (browser visible)")
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	logger.LogStart(common.CategoryPipeline, "web scraping")
	logger.DebugCategory(common.CategoryPipeline, "Starting ChromeDP scraper with dates: from=%s to=%s", fromSite, toSite)
	
	if err := chromedp.Run(ctx, runScraper(fromSite, toSite, *outDir)); err != nil {
		logger.ErrorCategory(common.CategoryPipeline, "Scraping failed: %v", err)
		fmt.Fprintf(os.Stderr, "scraping failed: %v\n", err)
		
		// Send error status via structured message
		sendStatus("scraping", "error", fmt.Sprintf("Scraping failed: %v", err))
		
		// Send error details
		sendError("SCRAPE_FAILED", "Failed to complete scraping process", err.Error(), 
			"scraping", true, "Check network connection and try again")
		
		// Ensure all output is flushed
		os.Stdout.Sync()
		os.Exit(1)
	}
	
	// ChromeDP completed successfully
	// The scraper has already printed summary and sent progress updates
	// Just ensure everything is flushed before exiting
	logger.LogComplete(common.CategoryPipeline, "web scraping")
	logger.InfoCategory(common.CategoryPipeline, "All operations completed successfully")
	
	os.Stdout.Sync()
	
	// Small delay to ensure parent process reads all output
	time.Sleep(100 * time.Millisecond)
}

// calculateExpectedFiles estimates how many files should be downloaded based on date range
func calculateExpectedFiles(fromSite, toSite string) int {
	if fromSite == "" {
		return 50 // Default estimate if no date range
	}
	
	fromDate, err := time.Parse("02/01/2006", fromSite)
	if err != nil {
		return 50 // Default if parsing fails
	}
	
	var toDate time.Time
	if toSite == "" {
		toDate = time.Now() // Current date if no end date
	} else {
		toDate, err = time.Parse("02/01/2006", toSite)
		if err != nil {
			toDate = time.Now()
		}
	}
	
	// Calculate business days (ISX operates Sunday-Thursday)
	days := int(toDate.Sub(fromDate).Hours() / 24)
	businessDays := 0
	
	for i := 0; i <= days; i++ {
		currentDate := fromDate.AddDate(0, 0, i)
		weekday := currentDate.Weekday()
		// ISX operates Sunday (0) through Thursday (4)
		if weekday >= time.Sunday && weekday <= time.Thursday {
			businessDays++
		}
	}
	
	return businessDays
}

// countExistingFiles counts how many files we already have in the date range
func countExistingFiles(outDir string, fromDate, toDate time.Time) (int, map[string]bool) {
	foundDates := make(map[string]bool)
	pattern := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return 0, foundDates
	}
	
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, err := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		if err != nil {
			continue
		}
		
		// Check if file is within date range
		if !fromDate.IsZero() && !toDate.IsZero() {
			if t.After(fromDate.AddDate(0, 0, -1)) && t.Before(toDate.AddDate(0, 0, 1)) {
				count++
				foundDates[t.Format("2006-01-02")] = true
			}
		} else {
			count++
			foundDates[t.Format("2006-01-02")] = true
		}
	}
	
	return count, foundDates
}

// calculateActualExpectedFiles calculates expected files minus those we already have
func calculateActualExpectedFiles(fromDate, toDate time.Time, foundDates map[string]bool) int {
	if fromDate.IsZero() || toDate.IsZero() {
		return 50 // Default estimate
	}
	
	days := int(toDate.Sub(fromDate).Hours() / 24)
	expectedCount := 0
	
	for i := 0; i <= days; i++ {
		currentDate := fromDate.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		weekday := currentDate.Weekday()
		
		// ISX operates Sunday (0) through Thursday (4)
		if weekday >= time.Sunday && weekday <= time.Thursday {
			// Only count if we don't already have this file
			if !foundDates[dateStr] {
				expectedCount++
			}
		}
	}
	
	return expectedCount
}

// printProgress outputs structured progress information for the web interface
func printProgress(calc *progress.EnhancedCalculator, totalDownloaded, totalExisting int, currentPage int) {
	totalProcessed := totalDownloaded + totalExisting
	calc.Update(totalProcessed)
	
	// Create detailed metadata
	details := map[string]interface{}{
		"downloaded":     totalDownloaded,
		"existing":       totalExisting,
		"current_page":   currentPage,
		"elapsed":        time.Since(calc.StartTime).String(),
		"expected_remaining": calc.TotalItems - calc.ProcessedItems,
	}
	
	// Get ETA (will use historical data if available)
	eta := calc.GetEnhancedETA()
	if eta != "" {
		details["eta_seconds"] = time.Since(calc.StartTime).Seconds()
	}
	
	message := fmt.Sprintf("Downloading Excel reports (%d new, %d existing)", totalDownloaded, totalExisting)
	
	// Send structured progress
	sendProgress(calc, message, details)
}


func runScraper(fromSite, toSite, outDir string) chromedp.Tasks {
	logger.LogStart(common.CategoryPipeline, "runScraper function")
	logger.DebugCategory(common.CategoryPipeline, "Parameters: fromSite=%s, toSite=%s, outDir=%s", fromSite, toSite, outDir)
	
	// Calculate expected files and initialize progress tracking
	expectedFiles := calculateExpectedFiles(fromSite, toSite)
	logger.DebugCategory(common.CategoryData, "Expected files to process: %d", expectedFiles)
	
	totalDownloaded := 0
	totalExisting := 0
	foundDates := make(map[string]bool) // Track which dates we've found files for
	
	// Initialize metrics manager for historical data
	dataPath := filepath.Dir(outDir) // Parent of downloads directory
	metricsManager := progress.NewMetricsManager(dataPath)
	
	// Create enhanced calculator with historical metrics support
	calc := progress.NewEnhancedCalculator("scraping", expectedFiles, metricsManager)
	
	// Parse date range for validation
	var fromDate, toDate time.Time
	var err error
	
	if fromSite != "" {
		fromDate, err = time.Parse("02/01/2006", fromSite)
		if err != nil {
			fmt.Printf("[ERROR] Invalid from date format: %s\n", fromSite)
		}
	}
	
	if toSite != "" {
		toDate, err = time.Parse("02/01/2006", toSite)
		if err != nil {
			fmt.Printf("[ERROR] Invalid to date format: %s\n", toSite)
		}
	} else {
		toDate = time.Now() // Default to current date if not specified
	}
	
	// Count existing files and get dates we already have
	existingCount, existingDates := countExistingFiles(outDir, fromDate, toDate)
	for date := range existingDates {
		foundDates[date] = true
	}
	
	// Recalculate expected files based on what we already have
	actualExpected := calculateActualExpectedFiles(fromDate, toDate, foundDates)
	calc.TotalItems = actualExpected // Update the calculator with actual expected count
	
	fmt.Printf("[INIT] ISX Daily Reports Scraper Starting\n")
	fmt.Printf("[INIT] Date Range: %s to %s\n", fromSite, toSite)
	fmt.Printf("[INIT] Total Business Days: %d (based on ISX schedule)\n", expectedFiles)
	fmt.Printf("[INIT] Existing Files: %d (already downloaded)\n", existingCount)
	fmt.Printf("[INIT] Files to Download: %d (excluding existing)\n", actualExpected)
	fmt.Printf("[INIT] Output Directory: %s\n", outDir)
	
	// Send pipeline status
	sendStatus("scraping", "active", "Starting ISX data download...")
	
	actions := []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.InfoCategory(common.CategoryPipeline, "Navigating to ISX portal: %s", startURL)
			return nil
		}),
		timedAction("Navigate", chromedp.Navigate(startURL)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.DebugCategory(common.CategoryPipeline, "Waiting for date field to be visible")
			return nil
		}),
		chromedp.WaitVisible(`#date`, chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.DebugCategory(common.CategoryPipeline, "Setting from date: %s", fromSite)
			return nil
		}),
		chromedp.SetValue(`#date`, fromSite, chromedp.ByID),
	}
	if toSite != "" {
		actions = append(actions, chromedp.SetValue(`#toDate`, toSite, chromedp.ByID))
	}
	actions = append(actions,
		chromedp.SetValue(`#reporttype`, "40", chromedp.ByID),
		timedAction("ExecuteSearch", chromedp.Click(`/html/body/div[2]/div/div[3]/div[3]/div[2]/div[4]/div/div[1]/form/div[8]/input`, chromedp.BySearch)),
		chromedp.WaitVisible(`#report`, chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Printf("[STATUS] Search completed, beginning file discovery...\n")
			page := 1
			
			for {
				pageStartTime := time.Now()
				fmt.Printf("[PAGE] Scanning page %d for downloadable reports...\n", page)
				
				shouldContinue, err := scrapePageWithProgress(ctx, outDir, &totalDownloaded, &totalExisting, fromDate, toDate, foundDates, calc)
				if err != nil {
					fmt.Printf("[ERROR] Failed to process page %d: %v\n", page, err)
					
					// Send error status
					sendStatus("scraping", "error", fmt.Sprintf("Failed to process page %d", page))
					
					// Send structured error
					sendError("PAGE_PROCESS_ERROR", fmt.Sprintf("Failed to process page %d", page),
						err.Error(), "scraping", true, "Page processing error - retrying may help")
					
					return err
				}
				
				// Print progress after each page
				printProgress(calc, totalDownloaded, totalExisting, page)
				
				pageTime := time.Since(pageStartTime)
				fmt.Printf("[PAGE] Page %d completed in %.1f seconds\n", page, pageTime.Seconds())
				
				if !shouldContinue {
					fmt.Printf("[COMPLETE] Reached existing files - download process finished\n")
					break
				}
				
				// Check if next arrow exists
				var nextExists bool
				err = chromedp.Run(ctx, chromedp.Evaluate(`!!document.querySelector('a img[src*="next.gif"]')`, &nextExists))
				if err != nil || !nextExists {
					fmt.Printf("[COMPLETE] No more pages available - all reports processed\n")
					break
				}
				
				// Click the parent anchor of the img
				fmt.Printf("[NAVIGATE] Moving to page %d...\n", page+1)
				if err := chromedp.Click(`a img[src*='next.gif']`, chromedp.ByQuery).Do(ctx); err != nil {
					fmt.Printf("[COMPLETE] Cannot navigate further - download finished\n")
					break
				}
				
				// Wait for table refresh
				if err := chromedp.WaitVisible(`#report`, chromedp.ByID).Do(ctx); err != nil {
					return err
				}
				
				page++
			}
			
			// Final summary
			totalTime := time.Since(calc.StartTime)
			fmt.Printf("\n[SUMMARY] ====== Download Complete ======\n")
			fmt.Printf("[SUMMARY] Total New Downloads: %d files\n", totalDownloaded)
			fmt.Printf("[SUMMARY] Total Existing Files: %d files\n", totalExisting)
			fmt.Printf("[SUMMARY] Total Processing Time: %.1f minutes\n", totalTime.Minutes())
			fmt.Printf("[SUMMARY] Pages Processed: %d\n", page)
			
			if totalDownloaded > 0 {
				avgTimePerFile := totalTime.Seconds() / float64(totalDownloaded)
				fmt.Printf("[SUMMARY] Average Time per Download: %.1f seconds\n", avgTimePerFile)
			}
			
			// Calculate actual coverage based on found dates
			totalFilesFound := len(foundDates)
			if expectedFiles > 0 {
				coveragePercent := float64(totalFilesFound) / float64(expectedFiles) * 100
				fmt.Printf("[SUMMARY] Date Range Coverage: %.1f%% (%d of %d business days)\n", 
					coveragePercent, totalFilesFound, expectedFiles)
			}
			
			// Show missing dates if there are gaps
			missingDays := expectedFiles - totalFilesFound
			if missingDays > 0 {
				fmt.Printf("[SUMMARY] Missing Days: %d (may be holidays or no trading)\n", missingDays)
			}
			
			fmt.Printf("[SUMMARY] ================================\n")
			
			// Send scraping completion status
			sendStatus("scraping", "completed", "Scraping phase completed successfully")
			
			// Record completion metrics for future ETA predictions
			if err := calc.Complete(); err != nil {
				fmt.Printf("[WARN] Failed to save metrics: %v\n", err)
			}
			
			// Final progress update
			sendProgress(calc, "Scraping completed", map[string]interface{}{
				"total_downloaded": totalDownloaded,
				"total_existing": totalExisting,
				"total_processed": totalFilesFound,
			})
			
			// Ensure all messages are flushed
			os.Stdout.Sync()
			
			return nil
		}),
	)

	return chromedp.Tasks(actions)
}

func scrapePage(ctx context.Context, outDir string) (bool, error) {
	logger.DebugCategory(common.CategoryPipeline, "Scraping current page for downloadable files")
	
	// Retrieve rows data: href, date text, type text
	var rows []struct {
		Href string `json:"href"`
		Date string `json:"date"`
		Typ  string `json:"typ"`
	}

	js := `Array.from(document.querySelectorAll('#report tbody tr')).map(tr => {
		const link = tr.querySelector('td.report-download a');
		if (!link) return null;
		const dateCell = tr.querySelector('td.report-titledata1');
		const typeCell = tr.querySelector('td.report-titledata3');
		return {href: link.getAttribute('href'), date: dateCell ? dateCell.innerText.trim() : '', typ: typeCell ? typeCell.innerText.trim() : ''};
	}).filter(Boolean)`

	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &rows)); err != nil {
		return false, err
	}

	foundExistingFiles := 0
	newDownloads := 0

	for _, r := range rows {
		// We only care about Daily type and xlsx file extension
		if strings.ToLower(r.Typ) != "daily" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(r.Href), ".xlsx") {
			continue
		}

		fullURL := r.Href
		if !strings.HasPrefix(r.Href, "http") {
			fullURL = baseURL + r.Href
		}

		// Parse date dd/mm/yyyy
		t, err := time.Parse("02/01/2006", r.Date)
		if err != nil {
			// fallback to original filename
			fmt.Printf(" !! unable to parse date '%s': %v\n", r.Date, err)
		}

		var fname string
		if err == nil {
			fname = fmt.Sprintf("%s ISX Daily Report.xlsx", t.Format("2006 01 02"))
		} else {
			fname = filepath.Base(r.Href)
		}

		destPath := filepath.Join(outDir, fname)
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf(" --> already have %s, skipping\n", fname)
			foundExistingFiles++
			continue
		}

		fmt.Printf(" --> downloading %s\n", fname)
		if err := downloadFile(fullURL, destPath); err != nil {
			fmt.Printf("failed to download %s: %v\n", fname, err)
		} else {
			newDownloads++
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Page summary: %d new downloads, %d existing files\n", newDownloads, foundExistingFiles)

	// If we found more existing files than new downloads, and we found at least some existing files,
	// it means we're getting into already-downloaded territory, so we should stop
	if foundExistingFiles > 0 && foundExistingFiles >= newDownloads {
		return false, nil // Stop scraping
	}

	return true, nil // Continue scraping
}

// scrapePageWithProgress is an enhanced version of scrapePage that updates progress counters
func scrapePageWithProgress(ctx context.Context, outDir string, totalDownloaded, totalExisting *int, fromDate, toDate time.Time, foundDates map[string]bool, calc *progress.EnhancedCalculator) (bool, error) {
	// Retrieve rows data: href, date text, type text
	var rows []struct {
		Href string `json:"href"`
		Date string `json:"date"`
		Type string `json:"type"`
	}

	const jsCode = `
	Array.from(document.querySelectorAll('#report tbody tr')).map(tr => {
		const link = tr.querySelector('td.report-download a');
		if (!link) return null;
		const dateCell = tr.querySelector('td.report-titledata1');
		const typeCell = tr.querySelector('td.report-titledata3');
		return {
			href: link.getAttribute('href'), 
			date: dateCell ? dateCell.innerText.trim() : '', 
			type: typeCell ? typeCell.innerText.trim() : ''
		};
	}).filter(Boolean);
	`

	if err := chromedp.Evaluate(jsCode, &rows).Do(ctx); err != nil {
		return false, err
	}

	pageDownloaded := 0
	pageExisting := 0
	filesOutsideRange := 0

	fmt.Printf("[FILES] Found %d files on this page\n", len(rows))

	// Filter for Daily xlsx files first
	var dailyFiles []struct {
		Href string `json:"href"`
		Date string `json:"date"`
		Type string `json:"type"`
	}
	
	for _, r := range rows {
		// We only care about Daily type and xlsx file extension
		if strings.ToLower(r.Type) != "daily" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(r.Href), ".xlsx") {
			continue
		}
		dailyFiles = append(dailyFiles, r)
	}

	fmt.Printf("[FILES] Found %d daily Excel files to process\n", len(dailyFiles))

	for i, r := range dailyFiles {

		fullURL := r.Href
		if !strings.HasPrefix(r.Href, "http") {
			fullURL = baseURL + r.Href
		}
		
		// Parse the date and create standardized filename
		t, err := time.Parse("02/01/2006", r.Date)
		if err != nil {
			fmt.Printf("[WARN] Unable to parse date '%s': %v\n", r.Date, err)
		}

		var fname string
		dateStr := ""
		if err == nil {
			fname = fmt.Sprintf("%s ISX Daily Report.xlsx", t.Format("2006 01 02"))
			dateStr = t.Format("2006-01-02")
			
			// Check if file date is within the specified range
			if !fromDate.IsZero() && t.Before(fromDate) {
				fmt.Printf("[SKIP] File %d/%d: %s (before date range - %s)\n", i+1, len(dailyFiles), fname, t.Format("2006-01-02"))
				filesOutsideRange++
				continue
			}
			
			if !toDate.IsZero() && t.After(toDate) {
				fmt.Printf("[SKIP] File %d/%d: %s (after date range - %s)\n", i+1, len(dailyFiles), fname, t.Format("2006-01-02"))
				filesOutsideRange++
				continue
			}
		} else {
			fname = filepath.Base(r.Href)
		}

		destPath := filepath.Join(outDir, fname)
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("[SKIP] File %d/%d: %s (already exists)\n", i+1, len(dailyFiles), fname)
			pageExisting++
			// Mark this date as found
			if dateStr != "" {
				foundDates[dateStr] = true
			}
			continue
		}

		fmt.Printf("[DOWNLOAD] File %d/%d: %s\n", i+1, len(dailyFiles), fname)
		downloadStart := time.Now()
		
		if err := downloadFile(fullURL, destPath); err != nil {
			fmt.Printf("[ERROR] Failed to download %s: %v\n", fname, err)
			
			// Send download error as structured message
			sendError("DOWNLOAD_ERROR", fmt.Sprintf("Failed to download %s", fname),
				err.Error(), "scraping", true, "Check network connection or file permissions")
		} else {
			pageDownloaded++
			downloadTime := time.Since(downloadStart)
			fmt.Printf("[SUCCESS] Downloaded %s in %.1f seconds\n", fname, downloadTime.Seconds())
			
			// Mark this date as found
			if dateStr != "" {
				foundDates[dateStr] = true
			}
			
			// Recalculate expected files based on what we've found so far
			// This is only for progress estimation, not for completion logic
			actualExpected := calculateActualExpectedFiles(fromDate, toDate, foundDates)
			if actualExpected != calc.TotalItems && actualExpected > 0 {
				fmt.Printf("[ADJUST] Updated expected remaining files from %d to %d for better ETA\n", calc.TotalItems - calc.ProcessedItems, actualExpected)
				calc.TotalItems = calc.ProcessedItems + actualExpected
			}
		}
		
		// Brief pause between downloads to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	// Update totals
	*totalDownloaded += pageDownloaded
	*totalExisting += pageExisting

	fmt.Printf("[PAGE_SUMMARY] Downloaded: %d | Existing: %d | Outside Range: %d | Total daily files: %d\n", 
		pageDownloaded, pageExisting, filesOutsideRange, len(dailyFiles))

	// Stop scraping if:
	// 1. We found more existing files than new downloads (already downloaded territory)
	// 2. OR we found files outside the date range (reached the boundary)
	// This ensures completion is based on actual overlap, not expected counts
	if (pageExisting > 0 && pageExisting >= pageDownloaded) || filesOutsideRange > 0 {
		if filesOutsideRange > 0 {
			fmt.Printf("[COMPLETE] Reached date range boundary - stopping download\n")
		} else {
			fmt.Printf("[COMPLETE] Reached existing files - stopping download\n")
		}
		return false, nil // Stop scraping
	}

	return true, nil // Continue scraping
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func timedAction(name string, act chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		start := time.Now()
		err := act.Do(ctx)
		fmt.Printf("[TIME] %s took %s\n", name, time.Since(start))
		return err
	})
}

// latestDownloadedDate looks for files named "YYYY MM DD ISX Daily Report.xlsx" in dir and returns the most recent date.
func latestDownloadedDate(dir string) (time.Time, bool) {
	pattern := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}, false
	}
	var dates []time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, err := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		if err == nil {
			dates = append(dates, t)
		}
	}
	if len(dates) == 0 {
		return time.Time{}, false
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates[len(dates)-1], true
}

func checkLicense() bool {
	// Initialize license manager
	licenseManager, err := license.NewManager("license.dat")
	if err != nil {
		fmt.Printf("⚠️  License system initialization failed: %v\n", err)
		return false
	}

	// Check if license is valid
	valid, err := licenseManager.ValidateLicense()
	if valid {
		// Get license info for display
		info, infoErr := licenseManager.GetLicenseInfo()
		if infoErr == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			fmt.Printf("✅ License Valid - %d days remaining\n", daysLeft)
			if daysLeft <= 7 {
				fmt.Printf("⚠️  License expires soon: %s\n", info.ExpiryDate.Format("2006-01-02"))
				fmt.Println("📞 Contact The Iraqi Investor Group for license renewal.")
			}
		}
		fmt.Println("═══════════════════════════════════════════════")
		return true
	}

	// License is invalid or expired
	fmt.Println("❌ Invalid or Expired License")
	fmt.Println("═══════════════════════════════════════════════")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Prompt for license key activation
	fmt.Println("\n🔑 Please enter your ISX license key to activate:")
	fmt.Println("   (License keys look like: ISX3M-ABC123DEF456GHI789JKL)")
	fmt.Print("License Key: ")

	reader := bufio.NewReader(os.Stdin)
	licenseKey, _ := reader.ReadString('\n')
	licenseKey = strings.TrimSpace(licenseKey)

	if licenseKey == "" {
		fmt.Println("❌ No license key provided.")
		return false
	}

	// Validate license key format
	if !isValidLicenseFormat(licenseKey) {
		fmt.Println("❌ Invalid license key format.")
		fmt.Println("   License keys should start with ISX1M, ISX3M, ISX6M, or ISX1Y")
		return false
	}

	// Activate license
	fmt.Println("🔄 Activating license...")
	if err := licenseManager.ActivateLicense(licenseKey); err != nil {
		fmt.Printf("❌ License activation failed: %v\n", err)
		fmt.Println("📞 Please contact The Iraqi Investor Group if you believe this is an error.")
		return false
	}

	fmt.Println("✅ License activated successfully!")
	fmt.Println("🎉 Welcome to ISX Daily Reports Scraper!")
	fmt.Println("═══════════════════════════════════════════════")
	return true
}

func isValidLicenseFormat(licenseKey string) bool {
	// Check if license key starts with valid prefixes
	validPrefixes := []string{"ISX1M", "ISX3M", "ISX6M", "ISX1Y"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(licenseKey, prefix) {
			return true
		}
	}
	return false
}
