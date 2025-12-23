package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/operations"

	"github.com/chromedp/chromedp"
)

const (
	baseURL                  = "http://www.isx-iq.net"
	startURL                 = "http://www.isx-iq.net/isxportal/portal/uploadedFilesList.html?currLanguage=en"
	downloadCooldownInterval = 10 * time.Millisecond
	fileStabilityTimeout     = 3 * time.Second
)

// chromeSession keeps a single Chrome instance alive across the entire scraping flow.
type chromeSession struct {
	allocCtx     context.Context
	allocCancel  context.CancelFunc
	browserCtx   context.Context
	browserClose context.CancelFunc
	logger       *slog.Logger
}

// newChromeSession applies aggressive performance flags and spins up a single browser process.
// It prefers the modern headless mode, but automatically falls back if Chrome rejects the flag.
func newChromeSession(headless bool, logger *slog.Logger) (*chromeSession, error) {
	if !headless {
		return startAndValidateSession(false, false, logger)
	}

	session, err := startAndValidateSession(true, true, logger)
	if err == nil {
		return session, nil
	}
	if isUnsupportedHeadlessErr(err) {
		logger.Warn("Chrome headless 'new' mode unsupported, falling back to classic headless",
			slog.String("error", err.Error()))
		return startAndValidateSession(true, false, logger)
	}
	return nil, err
}

func startAndValidateSession(headless bool, preferNew bool, logger *slog.Logger) (*chromeSession, error) {
	session, err := launchChromeSession(headless, preferNew, logger)
	if err != nil {
		return nil, err
	}
	if err := session.ensureHealthy(); err != nil {
		session.Close()
		return nil, err
	}
	return session, nil
}

func launchChromeSession(headless bool, preferNew bool, logger *slog.Logger) (*chromeSession, error) {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	switch {
	case !headless:
		opts = append(opts, chromedp.Flag("headless", false))
	case preferNew:
		opts = append(opts, chromedp.Flag("headless", "new"))
	default:
		opts = append(opts, chromedp.Flag("headless", true))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	return &chromeSession{
		allocCtx:     allocCtx,
		allocCancel:  allocCancel,
		browserCtx:   browserCtx,
		browserClose: browserCancel,
		logger:       logger,
	}, nil
}

// Close tears down Chrome cleanly with timeout to prevent hanging.
func (s *chromeSession) Close() {
	if s.browserClose != nil {
		// Add timeout to prevent hanging on Chrome cleanup
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.browserClose()
		}()

		select {
		case <-done:
			// Normal cleanup completed
			if s.logger != nil {
				s.logger.Debug("Chrome session closed normally")
			}
		case <-time.After(30 * time.Second):
			// Force timeout - Chrome cleanup is hanging
			if s.logger != nil {
				s.logger.Warn("Chrome session cleanup timeout, forcing exit")
			}
		}
	}

	if s.allocCancel != nil {
		s.allocCancel()
	}
}

func (s *chromeSession) ensureHealthy() error {
	ctx, cancel := s.NewTab()
	defer cancel()
	return chromedp.Run(ctx, chromedp.ActionFunc(func(context.Context) error {
		return nil
	}))
}

// NewTab returns a context bound to the existing browser process.
func (s *chromeSession) NewTab() (context.Context, context.CancelFunc) {
	return chromedp.NewContext(s.browserCtx)
}

func isUnsupportedHeadlessErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid exec pool flag") ||
		(strings.Contains(msg, "headless") && strings.Contains(msg, "invalid"))
}

func main() {
	// Add panic recovery at the very start to catch any crashes
	var logger *slog.Logger // Declare logger early for use in panic handler
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			// Log the panic with full stack trace
			fmt.Printf("PANIC RECOVERED: %v\n", r)
			fmt.Printf("Stack trace:\n%s\n", stack)

			// Try to log to file if logger is available
			if logger != nil {
				logger.Error("Scraper panicked",
					slog.Any("panic", r),
					slog.String("stack", string(stack)))
			}
			os.Exit(1)
		}
	}()

	mode := flag.String("mode", "", "scrape mode: initial | accumulative | skip (empty = auto-detect)")
	fromStr := flag.String("from", "", "start date (YYYY-MM-DD, required)")
	toStr := flag.String("to", "", "end date (YYYY-MM-DD, required)")
	// Actual dates for progress tracking (not for scraper logic)
	actualFromStr := flag.String("actual-from", "", "actual from date for progress calculation")
	actualToStr := flag.String("actual-to", "", "actual to date for progress calculation")
	outDir := flag.String("out", "", "directory to save reports (defaults to data/downloads relative to executable)")
	headless := flag.Bool("headless", true, "run browser headless")
	stateFile := flag.String("state-file", "", "path to license state file (for validation bypass)")
	flag.Parse()

	// Initialize paths first to get default directories
	paths, err := config.GetPaths()
	if err != nil {
		fmt.Printf("Error: Failed to initialize paths: %v\n", err)
		os.Exit(1)
	}

	// Set default dates if not provided
	*fromStr = strings.TrimSpace(*fromStr)
	*toStr = strings.TrimSpace(*toStr)

	if *fromStr == "" {
		fmt.Println("Error: --from date is required and cannot fall back to defaults")
		os.Exit(1)
	}
	if *toStr == "" {
		fmt.Println("Error: --to date is required and cannot fall back to defaults")
		os.Exit(1)
	}

	// Use centralized downloads directory as default if not specified
	if *outDir == "" {
		*outDir = paths.DownloadsDir
	}

	// Ensure all required directories exist
	if err := paths.EnsureDirectories(); err != nil {
		fmt.Printf("Error: Failed to create required directories: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger per CLAUDE.md
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		cfg = &config.Config{
			Logging: config.LoggingConfig{
				Level:       "info",
				Format:      "json",
				Output:      "both",
				FilePath:    paths.GetLogPath("scraper.log"),
				Development: false,
			},
		}
	}

	// Assign to pre-declared logger variable for panic handler
	var err2 error
	logger, err2 = infrastructure.InitializeLogger(cfg.Logging)
	if err2 != nil {
		fmt.Printf("Warning: Failed to initialize logger, using default: %v\n", err2)
		logger = slog.Default()
	}

	// Start resource monitoring in background
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			logger.Info("Resource usage",
				slog.Uint64("memory_alloc_mb", m.Alloc/1024/1024),
				slog.Uint64("memory_sys_mb", m.Sys/1024/1024),
				slog.Int("goroutines", runtime.NumGoroutine()))
		}
	}()

	// Initialize license system
	// Keep console output for user-facing messages
	slog.Info("🔐 ISX Daily Reports Scraper - Licensed Version")
	slog.Info("═══════════════════════════════════════════════")

	// Initialize mode detector for automatic mode detection
	ctx := context.Background()
	var finalMode string
	var modeDetector *operations.ModeDetector

	if *mode == "" {
		// Auto-detect mode using ModeDetector service
		modeDetector, err = operations.NewModeDetector(logger)
		if err != nil {
			logger.Error("Failed to initialize mode detector", slog.String("error", err.Error()))
			fmt.Printf("Error: Failed to initialize mode detector: %v\n", err)
			os.Exit(1)
		}

		logger.Info("Auto-detecting optimal scraping mode",
			slog.String("from_date", *fromStr),
			slog.String("to_date", *toStr))

		result, err := modeDetector.DetectOptimalMode(ctx, *fromStr, *toStr)
		if err != nil {
			logger.Error("Mode detection failed", slog.String("error", err.Error()))
			fmt.Printf("Error: Mode detection failed: %v\n", err)
			os.Exit(1)
		}

		finalMode = result.RecommendedMode
		logger.Info("Mode auto-detected",
			slog.String("detected_mode", finalMode),
			slog.String("reason", result.Reason),
			slog.Float64("coverage_percent", result.CoveragePercent),
			slog.Int("existing_files", result.ExistingFiles),
			slog.Int("expected_files", result.ExpectedFiles))

		slog.Info("Mode Detection Result",
			"mode", finalMode,
			"reason", result.Reason,
			"coverage", fmt.Sprintf("%.1f%%", result.CoveragePercent))

		// Handle skip mode - exit early if all files already exist
		if finalMode == operations.ModeSkip {
			logger.Info("Skip mode detected - all required files already exist",
				slog.Int("existing_files", result.ExistingFiles),
				slog.Int("expected_files", result.ExpectedFiles),
				slog.Float64("coverage_percent", result.CoveragePercent))

			slog.Info("SCRAPER_COMPLETE: All required files already exist")
			slog.Info("✅ All files already downloaded - skipping scraper execution")
			return
		}
	} else {
		finalMode = *mode
		logger.Info("Using explicitly specified mode", slog.String("mode", finalMode))
	}

	logger.Info("ISX Daily Reports Scraper starting",
		slog.String("final_mode", finalMode),
		slog.String("from", *fromStr),
		slog.String("to", *toStr),
		slog.String("actual_from", *actualFromStr),
		slog.String("actual_to", *actualToStr),
		slog.String("output_dir", *outDir),
		slog.String("executable_dir", paths.ExecutableDir))

	// Log resolved paths for debugging
	slog.Info("Output directory", "path", *outDir)
	slog.Info("Executable directory", "path", paths.ExecutableDir)

	if !checkLicense(*stateFile, logger) {
		slog.Info("❌ License validation failed. Application will exit.")
		slog.Info("📞 Contact The Iraqi Investor Group to get a new license.")
		logger.Error("License validation failed")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist (but don't delete existing files)
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		logger.Error("failed to create output dir", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// determine fromSite depending on final mode
	var fromSite string
	if finalMode == "accumulative" {
		// scan downloads for latest file
		if d, ok := latestDownloadedDate(*outDir); ok {
			// Start from the same day to re-download if missing
			fromSite = d.Format("02/01/2006")
			slog.Info("[MODE accumulative] Starting from last report date (inclusive)",
				"start_from", d.Format("2006-01-02"))
			logger.Info("Accumulative mode starting from last date",
				slog.String("start_from", d.Format("2006-01-02")),
				slog.String("note", "will re-download if missing"))
		}
	}

	if fromSite == "" {
		// fallback to user provided from
		startDate, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			logger.Error("invalid --from date", slog.String("error", err.Error()))
			fmt.Printf("Error: Invalid --from date: %v\n", err)
			os.Exit(1)
		}
		fromSite = startDate.Format("02/01/2006")
		slog.Info("[MODE initial] Starting from date (preserving existing files)", "from_date", startDate.Format("2006-01-02"))
		logger.Info("Initial mode starting",
			slog.String("from_date", startDate.Format("2006-01-02")),
			slog.String("mode", finalMode),
			slog.String("behavior", "preserving existing files"))
	}

	var toSite string
	if *toStr != "" {
		endDate, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			logger.Error("invalid --to date", slog.String("error", err.Error()))
			os.Exit(1)
		}
		toSite = endDate.Format("02/01/2006")
	}

	// Calculate expected files based on actual date range (not buffered range)
	// Use actual dates if provided, otherwise fall back to buffer dates
	expectedFromStr := *fromStr
	expectedToStr := *toStr
	if *actualFromStr != "" {
		expectedFromStr = *actualFromStr
		logger.Info("Using actual-from date for expected files calculation",
			slog.String("actual_from", *actualFromStr))
	}
	if *actualToStr != "" {
		expectedToStr = *actualToStr
		logger.Info("Using actual-to date for expected files calculation",
			slog.String("actual_to", *actualToStr))
	}

	expectedFiles := calculateExpectedFiles(expectedFromStr, expectedToStr)
	slog.Info("Expected files to download", "count", expectedFiles, "from", expectedFromStr, "to", expectedToStr)
	logger.Info("Calculated expected files",
		slog.Int("expected_files", expectedFiles),
		slog.String("calculation_from", expectedFromStr),
		slog.String("calculation_to", expectedToStr),
		slog.String("buffer_from", *fromStr),
		slog.String("buffer_to", *toStr))

	// Output for parsing by stages.go
	slog.Info("Total expected files", "count", expectedFiles, "from", *fromStr, "to", *toStr, "mode", finalMode)

	// Parse dates for scanning existing files
	fromDateForScan, err := time.Parse("2006-01-02", expectedFromStr)
	if err != nil {
		logger.Warn("Failed to parse from date for scan", slog.String("error", err.Error()))
		fromDateForScan = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	}

	toDateForScan := time.Now()
	if expectedToStr != "" {
		toDateForScan, err = time.Parse("2006-01-02", expectedToStr)
		if err != nil {
			logger.Warn("Failed to parse to date for scan", slog.String("error", err.Error()))
			toDateForScan = time.Now()
		}
	}

	// Scan for existing files first
	existingFiles, existingHolidays := scanExistingFiles(*outDir, fromDateForScan, toDateForScan, logger)
	logger.Info("Pre-scan found existing files",
		slog.Int("existing_files", existingFiles),
		slog.Int("holidays_detected", existingHolidays))

	// Check if we already have all needed files
	if existingFiles >= expectedFiles {
		logger.Info("All required files already exist",
			slog.Int("existing_files", existingFiles),
			slog.Int("holidays", existingHolidays),
			slog.Int("expected", expectedFiles))

		// Signal completion to stages.go
		slog.Info("SCRAPER_COMPLETE: All required dates processed")

		// Exit successfully without launching browser
		return
	}

	// setup Chrome session once and reuse the same browser/tab through the whole run
	session, err := newChromeSession(*headless, logger)
	if err != nil {
		logger.Error("Failed to initialize Chrome session", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer session.Close()

	ctx, cancelCtx := session.NewTab()
	defer cancelCtx()

	// Pass actual dates for progress tracking (if provided)
	// These are only for progress calculation, not for stopping logic
	if *actualFromStr != "" {
		logger.Info("Actual from date for progress", slog.String("actual_from", *actualFromStr))
	}
	if *actualToStr != "" {
		logger.Info("Actual to date for progress", slog.String("actual_to", *actualToStr))
	}

	actualFromDate, actualToDate := parseActualDateBounds(*actualFromStr, *actualToStr, logger)

	if err := runScraper(ctx, fromSite, toSite, *outDir, finalMode, logger, expectedFiles, actualFromDate, actualToDate); err != nil {
		logger.Error("scraping failed", slog.String("error", err.Error()))
		slog.Error("SCRAPER_FAILED: " + err.Error())
		os.Exit(1)
	}

	// GUARANTEED: Always emit completion signal for consistent parsing
	// This ensures the operations stage always knows when scraping is complete
	slog.Info("SCRAPER_COMPLETE: All required dates processed")
	logger.Info("Scraper completed successfully")
}

// scanExistingFiles scans the output directory for existing Excel files within the date range
func scanExistingFiles(outDir string, fromDate, toDate time.Time, logger *slog.Logger) (filesFound int, holidaysDetected int) {
	pattern := filepath.Join(outDir, "*.xlsx")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.Warn("Failed to scan existing files", slog.String("error", err.Error()))
		return 0, 0
	}

	var lastDate *time.Time
	datePattern := regexp.MustCompile(`(\d{4})\s+(\d{2})\s+(\d{2})`)

	for _, file := range files {
		fname := filepath.Base(file)
		matches := datePattern.FindStringSubmatch(fname)
		if len(matches) != 4 {
			continue
		}

		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		fileDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

		// Check if file is in date range
		if !fileDate.Before(fromDate) && !fileDate.After(toDate) {
			filesFound++

			// Check for holidays (gaps in dates)
			if lastDate != nil {
				var older, newer time.Time
				if fileDate.Before(*lastDate) {
					older = fileDate
					newer = *lastDate
				} else {
					older = *lastDate
					newer = fileDate
				}
				for missing := older.AddDate(0, 0, 1); missing.Before(newer); missing = missing.AddDate(0, 0, 1) {
					if missing.Weekday() != time.Friday && missing.Weekday() != time.Saturday {
						holidaysDetected++
					}
				}
			}
			lastDate = &fileDate

			// Log for stages.go to parse
			slog.Info("Already exists", "file", fname)
		}
	}

	return filesFound, holidaysDetected
}

func parseActualDateBounds(actualFromStr, actualToStr string, logger *slog.Logger) (*time.Time, *time.Time) {
	var actualFromDate *time.Time
	var actualToDate *time.Time

	if trimmed := strings.TrimSpace(actualFromStr); trimmed != "" {
		if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
			actualFromDate = &parsed
		} else if logger != nil {
			logger.Warn("Invalid actual-from date for progress tracking",
				slog.String("actual_from", trimmed),
				slog.String("error", err.Error()))
		}
	}

	if trimmed := strings.TrimSpace(actualToStr); trimmed != "" {
		if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
			actualToDate = &parsed
		} else if logger != nil {
			logger.Warn("Invalid actual-to date for progress tracking",
				slog.String("actual_to", trimmed),
				slog.String("error", err.Error()))
		}
	}

	return actualFromDate, actualToDate
}

func runScraper(ctx context.Context, fromSite, toSite, outDir, finalMode string, logger *slog.Logger, expectedFiles int, actualFromDate, actualToDate *time.Time) error {
	// Track progress
	totalDownloaded := 0
	totalExisting := 0
	filesInRange := 0                // Files within actual date range
	holidaysInRange := 0             // Holidays within actual date range
	var lastProcessedDate *time.Time // Track for holiday detection
	actions := []chromedp.Action{
		timedAction("Navigate", chromedp.Navigate(startURL)),
		chromedp.WaitVisible(`#date`, chromedp.ByID),
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
			page := 1
			for {
				pageStart := time.Now()
				slog.Info("Scraping page", "page", page)
				logger.Info("Scraping page", slog.Int("page", page))
				_, _, shouldContinue, err := scrapePage(ctx, outDir, finalMode, logger, &totalDownloaded, &totalExisting, &filesInRange, &holidaysInRange, expectedFiles, actualFromDate, actualToDate, &lastProcessedDate)
				if err != nil {
					return err
				}
				if !shouldContinue {
					slog.Info("Found existing files, stopping scraping process", "page", page)
					logger.Info("Found existing files, stopping scraping", slog.Int("page", page))
					return nil
				}
				// Only check expectedFiles in initial mode
				// Accumulative mode relies on overlap detection
				if finalMode != "accumulative" && (filesInRange+holidaysInRange) >= expectedFiles {
					logger.Info("Initial mode: completion criteria met",
						slog.Int("files_in_range", filesInRange),
						slog.Int("holidays_in_range", holidaysInRange),
						slog.Int("total_accounted", filesInRange+holidaysInRange),
						slog.Int("expected_files", expectedFiles),
						slog.String("mode", finalMode))
					// Signal completion
					slog.Info("SCRAPER_COMPLETE: All required dates processed")
					return nil
				}

				// check if next arrow exists
				var nextHref string
				var ok bool
				err = chromedp.Run(ctx, chromedp.AttributeValue(`a img[src*='next.gif']`, "src", &nextHref, &ok))
				if err != nil || !ok {
					// No next arrow or not clickable
					return nil
				}
				// Click the parent anchor of the img
				if err := chromedp.Click(`a img[src*='next.gif']`, chromedp.ByQuery).Do(ctx); err != nil {
					return nil // assume finished when can't click
				}
				// wait for table refresh
				if err := chromedp.WaitVisible(`#report`, chromedp.ByID).Do(ctx); err != nil {
					return err
				}
				logger.Debug("Page processed",
					slog.Int("page", page),
					slog.Duration("duration", time.Since(pageStart)))
				page++
			}
		}),
	)

	return chromedp.Run(ctx, actions...)
}

func scrapePage(ctx context.Context, outDir, finalMode string, logger *slog.Logger, totalDownloaded, totalExisting, filesInRange, holidaysInRange *int, expectedFiles int, actualFromDate, actualToDate *time.Time, lastProcessedDate **time.Time) (int, int, bool, error) {
	// Add panic recovery for this function
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in scrapePage",
				slog.Any("panic", r),
				slog.Int("files_downloaded", *totalDownloaded),
				slog.Int("files_existing", *totalExisting),
				slog.String("stack", string(debug.Stack())))
			panic(r) // Re-panic to be caught by main
		}
	}()

	// Helper to check if a date is within the actual range
	isDateInRange := func(t time.Time) bool {
		if actualFromDate != nil && t.Before(*actualFromDate) {
			return false
		}
		if actualToDate != nil && t.After(*actualToDate) {
			return false
		}
		return true
	}

	// Add progress checkpoint every 5 files
	totalProcessed := *totalDownloaded + *totalExisting
	if totalProcessed > 0 && totalProcessed%5 == 0 {
		progressPct := float64(totalProcessed) / float64(expectedFiles) * 100
		logger.Info("Progress checkpoint",
			slog.Int("total_processed", totalProcessed),
			slog.Int("downloaded", *totalDownloaded),
			slog.Int("existing", *totalExisting),
			slog.Int("expected", expectedFiles),
			slog.Float64("percentage", progressPct))
	}

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

	rowsStart := time.Now()
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &rows)); err != nil {
		return 0, 0, false, err
	}
	logger.Debug("Fetched table rows",
		slog.Int("row_count", len(rows)),
		slog.Float64("duration_seconds", time.Since(rowsStart).Seconds()))

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
			logger.Warn("unable to parse date",
				slog.String("date", r.Date),
				slog.String("error", err.Error()))
		}

		// Check for holiday gaps with previous file
		// Since files are served newest to oldest, lastProcessedDate is newer than current t
		if err == nil && *lastProcessedDate != nil {
			// Calculate days between current (older) and last (newer)
			daysDiff := (*lastProcessedDate).Sub(t).Hours() / 24
			if daysDiff > 1 {
				// Found gap - report holidays between t and lastProcessedDate
				// Start from day after current file (older) to day before last file (newer)
				for d := t.AddDate(0, 0, 1); d.Before(**lastProcessedDate); d = d.AddDate(0, 0, 1) {
					// Skip weekends (Friday=5, Saturday=6 in Iraq)
					if d.Weekday() != time.Friday && d.Weekday() != time.Saturday {
						// Check if this holiday is in our actual date range
						if isDateInRange(d) {
							*holidaysInRange++
							// CRITICAL: Adjust expected files count dynamically
							expectedFiles--

							logger.Info("Detected non-trading day - adjusting expected files count",
								slog.String("date", d.Format("2006-01-02")),
								slog.Int("holidays_in_range", *holidaysInRange),
								slog.Int("remaining_expected_files", expectedFiles),
								slog.String("reason", "gap_analysis"))

							// Emit structured JSON for operations stage to parse and broadcast
							operationID := os.Getenv("ISX_OPERATION_ID")
							if operationID != "" {
								holidayUpdate := map[string]interface{}{
									"type": "holiday_detected",
									"operation_id": operationID,
									"date": d.Format("2006-01-02"),
									"reason": "non_trading_day_gap_detected",
									"previous_expected": expectedFiles + 1,
									"new_expected": expectedFiles,
									"total_holidays": *holidaysInRange,
									"timestamp": time.Now().UnixMilli(),
									"weekday": d.Weekday().String(),
								}

								if jsonData, err := json.Marshal(holidayUpdate); err == nil {
									fmt.Printf("SCRAPER_HOLIDAY_UPDATE: %s\n", string(jsonData))
								}
							}

							// Also log to console for immediate feedback
							fmt.Printf("📅 NON-TRADING DAY DETECTED: %s (Expected files adjusted to %d)\n",
								d.Format("2006-01-02"), expectedFiles)
						}
					}
				}
			}
		}

		var fname string
		if err == nil {
			fname = fmt.Sprintf("%s ISX Daily Report.xlsx", t.Format("2006 01 02"))
		} else {
			fname = filepath.Base(r.Href)
		}

		destPath := filepath.Join(outDir, fname)
		if _, err := os.Stat(destPath); err == nil {
			foundExistingFiles++
			*totalExisting++
			// Check if this existing file is in range
			if err == nil && isDateInRange(t) {
				*filesInRange++
			}
			totalFiles := *totalDownloaded + *totalExisting
			progressMsg := fmt.Sprintf("File %d of %d already exists, skipping", totalFiles, expectedFiles)
			slog.Info(progressMsg, "file", fname)
			logger.Debug("File already exists",
				slog.String("file", fname),
				slog.Int("total_processed", totalFiles),
				slog.Int("expected_files", expectedFiles),
				slog.Int("files_in_range", *filesInRange))
			continue
		}

		newDownloads++
		*totalDownloaded++
		totalFiles := *totalDownloaded + *totalExisting
		progressMsg := fmt.Sprintf("Downloading file %d of %d", totalFiles, expectedFiles)
		slog.Info(progressMsg, "file", fname)
		logger.Info("Downloading file",
			slog.String("file", fname),
			slog.Int("file_number", totalFiles),
			slog.Int("expected_files", expectedFiles))

		downloadStart := time.Now()
		if err := downloadFile(fullURL, destPath, logger); err != nil {
			slog.Error("Failed to download file", "file", fname, "error", err)
			logger.Error("Failed to download file",
				slog.String("file", fname),
				slog.String("error", err.Error()))
			logger.Debug("Download attempt timing",
				slog.String("file", fname),
				slog.Float64("download_seconds", time.Since(downloadStart).Seconds()))
			// Revert counts on failure
			newDownloads--
			*totalDownloaded--
		} else {
			logger.Debug("Download completed",
				slog.String("file", fname),
				slog.Float64("download_seconds", time.Since(downloadStart).Seconds()))
			stabilityStart := time.Now()
			if waitErr := waitForFileStability(ctx, destPath, logger); waitErr != nil {
				return newDownloads, foundExistingFiles, false, waitErr
			}
			logger.Debug("File stability wait completed",
				slog.String("file", fname),
				slog.Float64("stability_wait_seconds", time.Since(stabilityStart).Seconds()))
			// Successfully downloaded - check if in range
			if err == nil && isDateInRange(t) {
				*filesInRange++
				logger.Info("Downloaded file in range",
					slog.String("file", fname),
					slog.Int("files_in_range", *filesInRange))
			}
		}

		throttleStart := time.Now()
		if err := throttleBetweenDownloads(ctx, downloadCooldownInterval); err != nil {
			return newDownloads, foundExistingFiles, false, err
		}
		logger.Debug("Download throttle completed",
			slog.Float64("throttle_seconds", time.Since(throttleStart).Seconds()))
		cleanupBetweenDownloads(*totalDownloaded+*totalExisting, logger)

		// Check if this file was before actual-from date (buffer zone)
		if err == nil && actualFromDate != nil && t.Before(*actualFromDate) {
			// This file is in the buffer zone - we've processed all files in range
			logger.Info("Reached buffer zone after processing files in range",
				slog.String("file_date", t.Format("2006-01-02")),
				slog.String("actual_from", actualFromDate.Format("2006-01-02")),
				slog.Int("files_downloaded", newDownloads),
				slog.Int("files_existing", foundExistingFiles),
				slog.Int("files_in_range", *filesInRange),
				slog.Int("holidays_in_range", *holidaysInRange))

			// Check if we have accounted for all expected files
			if (*filesInRange + *holidaysInRange) >= expectedFiles {
				logger.Info("Completion criteria met",
					slog.Int("files_in_range", *filesInRange),
					slog.Int("holidays_in_range", *holidaysInRange),
					slog.Int("total_accounted", *filesInRange+*holidaysInRange),
					slog.Int("expected_files", expectedFiles))
				// Signal completion
				slog.Info("SCRAPER_COMPLETE: All required dates processed")
			}

			return newDownloads, foundExistingFiles, false, nil // Stop scraping
		}

		// Update last processed date for holiday detection
		if err == nil {
			*lastProcessedDate = &t
		}
	}

	slog.Info("Page summary", "new_downloads", newDownloads, "existing_files", foundExistingFiles)
	logger.Info("Page summary",
		slog.Int("new_downloads", newDownloads),
		slog.Int("existing_files", foundExistingFiles))

	// Output total progress summary
	totalFiles := *totalDownloaded + *totalExisting
	slog.Info("Progress summary",
		"processed", totalFiles,
		"expected", expectedFiles,
		"downloaded", *totalDownloaded,
		"existing", *totalExisting)

	// If the page had nothing actionable, just advance immediately.
	if newDownloads == 0 && foundExistingFiles == 0 {
		slog.Info("No files found on page, moving to next page")
		logger.Info("No files found on page, moving to next page")
		return newDownloads, foundExistingFiles, true, nil
	}

	// Smart overlap detection based on mode
	if finalMode == "accumulative" {
		// In accumulative mode with newest-first ordering:
		// If we found ANY existing files, we've reached our overlap point
		// This works because:
		// 1. Files are served newest to oldest
		// 2. We start from the last downloaded date (inclusive)
		// 3. Once we hit existing files, all older files will also exist
		if foundExistingFiles > 0 {
			logger.Info("Accumulative mode: reached overlap zone",
				slog.Int("existing_on_page", foundExistingFiles),
				slog.Int("new_on_page", newDownloads),
				slog.String("mode", "accumulative"),
				slog.String("reason", "found existing files - all newer files have been downloaded"))

			// Only signal completion if we actually downloaded something
			// or if this is the first page with all existing files
			if *totalDownloaded > 0 || (newDownloads == 0 && foundExistingFiles > 0) {
				slog.Info("SCRAPER_COMPLETE: All required dates processed")
			}

			return newDownloads, foundExistingFiles, false, nil // Stop
		}
		// Continue only if NO existing files found yet
		logger.Debug("Accumulative mode: continuing, no existing files found yet",
			slog.Int("new_on_page", newDownloads),
			slog.Int("existing_on_page", foundExistingFiles))
	} else {
		// Initial mode: use ratio-based detection for robustness
		// This handles gaps, holidays, and partial downloads
		if foundExistingFiles > 0 && foundExistingFiles > newDownloads*3 {
			logger.Info("Initial mode: found mostly existing files, stopping",
				slog.Int("existing", foundExistingFiles),
				slog.Int("new", newDownloads),
				slog.String("mode", "initial"))
			return newDownloads, foundExistingFiles, false, nil // Stop
		}
	}

	return newDownloads, foundExistingFiles, true, nil // Continue scraping
}

func downloadFile(url, dest string, logger *slog.Logger) error {
	logger.Debug("Starting file download",
		slog.String("url", url),
		slog.String("destination", dest))

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("HTTP GET failed",
			slog.String("url", url),
			slog.String("error", err.Error()),
			slog.String("error_type", fmt.Sprintf("%T", err)))
		return fmt.Errorf("download failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Bad HTTP status",
			slog.String("url", url),
			slog.Int("status_code", resp.StatusCode),
			slog.String("status", resp.Status))
		return fmt.Errorf("bad status for %s: %s", url, resp.Status)
	}

	// Log file creation attempt
	logger.Debug("Creating output file",
		slog.String("path", dest))

	out, err := os.Create(dest)
	if err != nil {
		logger.Error("Failed to create file",
			slog.String("path", dest),
			slog.String("error", err.Error()))
		return fmt.Errorf("create file %s: %w", dest, err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		logger.Error("Failed to write file content",
			slog.String("path", dest),
			slog.Int64("bytes_written", written),
			slog.String("error", err.Error()))
		return fmt.Errorf("write file %s: %w", dest, err)
	}

	logger.Info("File downloaded successfully",
		slog.String("file", filepath.Base(dest)),
		slog.Int64("size_bytes", written),
		slog.Float64("size_mb", float64(written)/1024/1024))

	return nil
}

// waitForFileStability polls the filesystem quickly to ensure the download is flushed before moving on.
func waitForFileStability(ctx context.Context, filePath string, logger *slog.Logger) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(fileStabilityTimeout)
	defer timeout.Stop()

	var lastSize int64 = -1
	stableTicks := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			if info.Size() == lastSize {
				stableTicks++
				if stableTicks >= 2 {
					return nil
				}
				continue
			}
			lastSize = info.Size()
			stableTicks = 0
		case <-timeout.C:
			logger.Debug("File stability timeout reached, continuing",
				slog.String("file", filepath.Base(filePath)))
			return nil
		}
	}
}

// throttleBetweenDownloads replaces the previous fixed 500ms waits with a shorter, cancellable pause.
func throttleBetweenDownloads(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// cleanupBetweenDownloads performs lightweight resource maintenance without pausing the scraper.
func cleanupBetweenDownloads(processed int, logger *slog.Logger) {
	if processed == 0 || processed%10 != 0 {
		return
	}

	runtime.GC()
	logger.Debug("Triggered GC after batch",
		slog.Int("processed_files", processed))
}

func timedAction(name string, act chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return act.Do(ctx)
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

func checkLicense(stateFilePath string, logger *slog.Logger) bool {
	// Initialize license manager
	licenseManager, err := license.NewManager()
	if err != nil {
		logger.Error("License system initialization failed", slog.String("error", err.Error()))
		return false
	}
	activator := license.NewActivationService(licenseManager)

	// Check state file first if provided
	if stateFilePath != "" {
		logger.Info("Checking license state file", slog.String("path", stateFilePath))
		valid, err := licenseManager.ValidateStateFile(stateFilePath)
		if err != nil {
			logger.Warn("State file validation error", slog.String("error", err.Error()))
			// Continue with normal validation
		} else if valid {
			slog.Info("✅ License validated via state file")
			slog.Info("═══════════════════════════════════════════════")
			logger.Info("License validated via state file")
			return true
		} else {
			slog.Info("⚠️  State file invalid or expired, proceeding with normal validation")
			logger.Warn("State file invalid or expired, proceeding with normal validation")
		}
	}

	// Check if license is valid
	valid, err := licenseManager.ValidateLicense()
	if valid {
		// Get license info for display
		info, infoErr := licenseManager.GetLicenseInfo()
		if infoErr == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			slog.Info("License Valid", "days_remaining", daysLeft)
			logger.Info("License Valid", slog.Int("days_remaining", daysLeft))
			if daysLeft <= 7 {
				slog.Warn("License expires soon", "expiry_date", info.ExpiryDate.Format("2006-01-02"))
				slog.Info("Contact The Iraqi Investor Group for license renewal")
				logger.Warn("License expires soon",
					slog.String("expiry_date", info.ExpiryDate.Format("2006-01-02")),
					slog.String("action", "Contact The Iraqi Investor Group for license renewal"))
			}
		}
		slog.Info("═══════════════════════════════════════════════")
		return true
	}

	// License is invalid or expired
	slog.Info("❌ Invalid or Expired License")
	slog.Info("═══════════════════════════════════════════════")
	logger.Error("Invalid or Expired License")

	if err != nil {
		logger.Error("License validation error", slog.String("error", err.Error()))
		fmt.Printf("Error: %v\n", err)
	}

	// Prompt for license key activation
	slog.Info("Please enter your ISX license key to activate")
	slog.Info("License keys look like: ISX3M-ABC123DEF456GHI789JKL")
	slog.Info("License Key: (waiting for input...)")

	reader := bufio.NewReader(os.Stdin)
	licenseKey, _ := reader.ReadString('\n')
	licenseKey = strings.TrimSpace(licenseKey)

	if licenseKey == "" {
		slog.Info("❌ No license key provided.")
		logger.Error("No license key provided")
		return false
	}

	// Validate license key format
	if !isValidLicenseFormat(licenseKey) {
		slog.Info("❌ Invalid license key format.")
		slog.Info("   License keys should start with ISX1M, ISX3M, ISX6M, or ISX1Y")
		logger.Error("Invalid license key format")
		return false
	}

	// Activate license
	logger.Info("Activating license...")
	if err := activator.Activate(context.Background(), licenseKey); err != nil {
		logger.Error("License activation failed", slog.String("error", err.Error()))
		fmt.Printf("License activation failed: %v\n", err)
		slog.Info("Please contact The Iraqi Investor Group if you believe this is an error.")
		return false
	}

	slog.Info("✅ License activated successfully!")
	slog.Info("🎉 Welcome to ISX Daily Reports Scraper!")
	slog.Info("═══════════════════════════════════════════════")
	logger.Info("License activated successfully")
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

// calculateExpectedFiles calculates the expected number of files based on date range
// ISX publishes reports on working days (Sunday-Thursday in Iraq)
func calculateExpectedFiles(fromStr, toStr string) int {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return 0
	}

	now := time.Now()
	endDate := now
	if toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			endDate = parsed
		}
	}

	// Don't count future dates
	if endDate.After(now) {
		endDate = now
	}

	// Count working days between start and end
	count := 0
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// ISX is closed on Friday and Saturday
		if d.Weekday() != time.Friday && d.Weekday() != time.Saturday {
			count++
		}
	}

	return count
}
