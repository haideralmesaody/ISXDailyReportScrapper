package operations

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	iso8601Date = "2006-01-02"
	// File validation constants
	minFileSizeBytes = 1024             // Minimum file size: 1KB
	maxFileSizeBytes = 50 * 1024 * 1024 // Maximum file size: 50MB
	// Checksum constants
	checksumExtension = ".md5" // Extension for checksum files
)

// EventDrivenFileMonitor handles event-driven file monitoring using fsnotify
type EventDrivenFileMonitor struct {
	watcher     *fsnotify.Watcher
	fileEvents  chan fsnotify.Event
	downloadDir string
	cutoff      time.Time
	filesSeen   map[string]bool
	mutex       sync.RWMutex
	logger      *slog.Logger
}

// NewEventDrivenFileMonitor creates a new file monitor for the specified directory
func NewEventDrivenFileMonitor(downloadDir string, logger *slog.Logger) (*EventDrivenFileMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	monitor := &EventDrivenFileMonitor{
		watcher:     watcher,
		fileEvents:  make(chan fsnotify.Event, 100),
		downloadDir: downloadDir,
		filesSeen:   make(map[string]bool),
		logger:      logger,
	}

	// Add the download directory to the watcher
	if err := watcher.Add(downloadDir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch directory %s: %w", downloadDir, err)
	}

	// Also watch subdirectories
	if err := monitor.watchSubdirectories(downloadDir); err != nil && logger != nil {
		logger.Warn("Failed to watch some subdirectories", "error", err.Error())
	}

	return monitor, nil
}

// watchSubdirectories recursively adds subdirectories to the watcher
func (m *EventDrivenFileMonitor) watchSubdirectories(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != dir {
			if err := m.watcher.Add(path); err != nil {
				if m.logger != nil {
					m.logger.Debug("Failed to watch subdirectory", "path", path, "error", err.Error())
				}
				return nil // Continue walking other directories
			}
		}
		return nil
	})
}

// Start begins the event monitoring loop
func (m *EventDrivenFileMonitor) Start() {
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				m.fileEvents <- event
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				if m.logger != nil {
					m.logger.Warn("File watcher error", "error", err.Error())
				}
			}
		}
	}()
}

// Close stops the file monitoring
func (m *EventDrivenFileMonitor) Close() error {
	return m.watcher.Close()
}

// IsRelevantEvent checks if the file event is relevant for download monitoring
func (m *EventDrivenFileMonitor) IsRelevantEvent(event fsnotify.Event) bool {
	// Only interested in Excel files
	if !strings.HasSuffix(strings.ToLower(event.Name), ".xlsx") {
		return false
	}

	// Check if file is after cutoff time
	if info, err := os.Stat(event.Name); err == nil {
		if info.ModTime().Before(m.cutoff) {
			return false
		}
	}

	return true
}

// MarkFileSeen marks a file as been processed
func (m *EventDrivenFileMonitor) MarkFileSeen(filename string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.filesSeen[filepath.Base(filename)] = true
}

// HasFileSeen checks if a file has been seen before
func (m *EventDrivenFileMonitor) HasFileSeen(filename string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.filesSeen[filepath.Base(filename)]
}

// GetSeenFiles returns a list of all seen files
func (m *EventDrivenFileMonitor) GetSeenFiles() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	files := make([]string, 0, len(m.filesSeen))
	for file := range m.filesSeen {
		files = append(files, file)
	}
	return files
}

// SetCutoff updates the cutoff time for file filtering
func (m *EventDrivenFileMonitor) SetCutoff(cutoff time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cutoff = cutoff
}

// DiskSpaceInfo contains information about disk usage
type DiskSpaceInfo struct {
	TotalGB     float64 `json:"total_gb"`
	UsedGB      float64 `json:"used_gb"`
	AvailableGB float64 `json:"available_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// ScrapingStage handles the scraping of ISX daily reports
type ScrapingStage struct {
	BaseStage
	executableDir string
	logger        *slog.Logger
	options       *StageOptions
	broadcaster   *VelocityTrackingBroadcaster
}

const (
	maxScraperLogBytes = 8 * 1024
)

// NewScrapingStage creates a new scraping Stage
func NewScrapingStage(executableDir string, logger *slog.Logger, options *StageOptions) *ScrapingStage {
	if options == nil {
		options = &StageOptions{}
	}

	// Create logger with Stage context
	if logger != nil {
		logger = logger.With(slog.String("Stage", StageIDScraping))
		logger.Info("Scraping Stage initialized",
			slog.String("executable_dir", executableDir))
	}

	// Create velocity tracking broadcaster for scraping operations
	var broadcaster *VelocityTrackingBroadcaster
	if options.WebSocketManager != nil {
		broadcaster = NewVelocityTrackingBroadcaster("temp", StageIDScraping, options.WebSocketManager, logger, options.EnableProgress)
	}

	return &ScrapingStage{
		BaseStage:     NewBaseStage(StageIDScraping, StageNameScraping, nil),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
		broadcaster:   broadcaster,
	}
}

// setOperationID updates the broadcaster with the actual operation ID
func (s *ScrapingStage) setOperationID(operationID string) {
	if s.broadcaster == nil || operationID == "" {
		return
	}

	// CRITICAL: Validate operation ID to prevent "temp" ID broadcasts
	if operationID == "temp" {
		if s.logger != nil {
			s.logger.Warn("setOperationID skipped - invalid operation ID",
				"operation_id", operationID,
				"reason", "invalid_operation_id")
		}
		return
	}

	if s.broadcaster.operationID == operationID {
		return
	}

	currentHub := s.broadcaster.hub
	isEnabled := s.broadcaster.IsEnabled()
	newBroadcaster := NewVelocityTrackingBroadcaster(
		operationID,
		StageIDScraping,
		currentHub,
		s.logger,
		isEnabled,
	)
	// Directly assign velocityMetrics to avoid unnecessary map copy
	newBroadcaster.velocityMetrics = s.broadcaster.velocityMetrics
	newBroadcaster.lastUpdate = s.broadcaster.lastUpdate

	s.broadcaster = newBroadcaster

	if s.logger != nil {
		s.logger.Info("setOperationID completed successfully",
			"operation_id", operationID,
			"stage", StageIDScraping,
			"broadcaster_enabled", isEnabled)
	}
}

// Execute runs the scraper to download ISX daily reports
func (s *ScrapingStage) Execute(ctx context.Context, state *OperationState) error {
	// Update broadcaster with actual operation ID
	s.setOperationID(state.ID)

	stepState := state.GetStage(s.ID())

	// Validate stepState
	if stepState == nil {
		return fmt.Errorf("scraping stage state not initialized")
	}

	var (
		monitorCtx     context.Context
		monitorCancel  context.CancelFunc
		watchdog       *scraperWatchdog
		completionOnce sync.Once
	)

	completeStage := func(status, message string, metadata map[string]interface{}) {
		completionOnce.Do(func() {
			// Normalize step status for UI consumption
			if stepState != nil {
				switch status {
				case "completed", "complete":
					stepState.Complete()
				case "failed":
					stepState.Fail(fmt.Errorf(message))
				}
			}

			if monitorCancel != nil {
				monitorCancel()
				monitorCancel = nil
			}
			if watchdog != nil {
				watchdog.Stop()
				watchdog = nil
			}
			metadataForBroadcast := cloneMetadata(metadata)
			metadataForEmit := cloneMetadata(metadata)
			if s.broadcaster != nil {
				s.broadcaster.CompleteWithMetadata(message, metadataForBroadcast)
			}
			s.emitScrapingTelemetry(state.ID, status, metadataForEmit)
			if s.logger != nil {
				s.logger.Info("Scraping stage completion finalized",
					slog.String("operation_id", state.ID),
					slog.String("status", status),
					slog.String("message", message))
			}
		})
	}

	// Log Stage execution start
	if s.logger != nil {
		s.logger.Info("Starting scraping Stage",
			slog.String("operation_id", state.ID))
	}

	// Check license if required
	if s.options.LicenseChecker != nil && s.options.LicenseChecker.RequiresLicense() {
		if err := s.options.LicenseChecker.CheckLicense(); err != nil {
			return fmt.Errorf("invalid or expired license for scraping operations: %w", err)
		}
	}

	// Initialize progress
	s.updateProgress(stepState, 10, "Initializing scraper...")

	// Prepare download directory
	downloadDir := AbsoluteDownloadsPath(s.executableDir, "")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Check disk space after ensuring directory exists
	s.updateProgress(stepState, 20, "Checking disk space...")
	diskInfo, err := s.checkDiskSpace(downloadDir)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("Disk space check failed",
				slog.String("operation_id", state.ID),
				slog.String("path", downloadDir),
				slog.String("error", err.Error()))
		}
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	// Add disk space info to metadata
	withStepMetadata(stepState, func(md map[string]interface{}) {
		md["disk_total_gb"] = diskInfo.TotalGB
		md["disk_used_gb"] = diskInfo.UsedGB
		md["disk_available_gb"] = diskInfo.AvailableGB
	})

	if s.logger != nil {
		s.logger.Info("Disk space check completed",
			slog.String("operation_id", state.ID),
			slog.String("path", downloadDir),
			slog.Float64("total_gb", diskInfo.TotalGB),
			slog.Float64("used_gb", diskInfo.UsedGB),
			slog.Float64("available_gb", diskInfo.AvailableGB),
			slog.Float64("used_percent", diskInfo.UsedPercent))
	}

	// Check if we have enough space
	if diskInfo.AvailableGB < 1.0 { // Need at least 1GB
		return fmt.Errorf("insufficient disk space: only %.2fGB available", diskInfo.AvailableGB)
	}

	fromDate := s.getConfigString(state, ContextKeyFromDate)
	toDate := s.getConfigString(state, ContextKeyToDate)
	mode := s.getConfigString(state, ContextKeyMode)

	// Normalize date range (swap if inverted) to avoid zero trading-day totals
	if fromDate != "" && toDate != "" {
		fromTime, errFrom := time.Parse("2006-01-02", fromDate)
		toTime, errTo := time.Parse("2006-01-02", toDate)
		if errFrom == nil && errTo == nil && fromTime.After(toTime) {
			// swap
			fromDate, toDate = toDate, fromDate
		}
	}

	// Store metadata for UI diagnostics
	withStepMetadata(stepState, func(md map[string]interface{}) {
		md["download_dir"] = downloadDir
		if fromDate != "" {
			md["from_date"] = fromDate
		}
		if toDate != "" {
			md["to_date"] = toDate
		}
		if mode != "" {
			md["mode"] = mode
		}
	})

	stageStart := time.Now()
	// Use simplified file count validation
	expectedFiles, validatedTotal := s.validateFileCount(fromDate, toDate)
	totalTradingDays := validatedTotal // Use validated count for progress

	// Store validation info in metadata
	withStepMetadata(stepState, func(md map[string]interface{}) {
		md["total_trading_days"] = totalTradingDays
		md["expected_files_validated"] = expectedFiles
		md["file_validation_method"] = "trading_days_count"
	})

	initialTelemetry := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
		FromDate:             fromDate,
		ToDate:               toDate,
		DownloadedFiles:      []string{},
		SkippedFiles:         []string{},
		TradingDaysTotal:     totalTradingDays,
		TradingDaysCompleted: 0,
		CurrentFile:          "",
		StageMessage:         "Initializing scraper...",
		ProgressPercent:      0,
		Status:               "running",
	})
	s.applyTelemetryToStep(stepState, initialTelemetry)
	if s.broadcaster != nil {
		s.broadcaster.UpdateProgressWithMetadata(0, "Initializing scraper...", initialTelemetry)
	}
	s.emitScrapingTelemetry(state.ID, "running", initialTelemetry)

	heartbeat := newDownloadHeartbeat()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("scraping cancelled: %w", ctx.Err())
	default:
	}

	// Execute the scraping process
	s.updateProgress(stepState, 40, "Starting download process...")

	executablePath, err := s.resolveScraperExecutable()
	if err != nil {
		return err
	}

	scraperArgs := s.buildScraperArgs(fromDate, toDate, mode, downloadDir)
	withStepMetadata(stepState, func(md map[string]interface{}) {
		md["scraper_command"] = map[string]interface{}{
			"executable": filepath.Base(executablePath),
			"args":       scraperArgs,
		}
	})

	monitorCtx, monitorCancel = context.WithCancel(ctx)
	defer func() {
		if monitorCancel != nil {
			monitorCancel()
		}
	}()
	// Monitor context also carries a channel to signal scraper exit
	scraperDone := make(chan struct{})
	go s.monitorDownloadProgressEventDriven(monitorCtx, stepState, state.ID, downloadDir, fromDate, toDate, stageStart, totalTradingDays, heartbeat, scraperDone, completeStage)

	timeoutPerFile := 10 * time.Second
	calculatedTimeout := timeoutPerFile
	if totalTradingDays > 1 {
		calculatedTimeout = time.Duration(totalTradingDays) * timeoutPerFile
	}
	scraperCtx, scraperCancel := s.prepareScraperContext(ctx, calculatedTimeout)
	defer scraperCancel()

	watchdog = s.startScraperWatchdog(ctx, heartbeat, scraperCancel, monitorCancel, totalTradingDays)
	defer func() {
		if watchdog != nil {
			watchdog.Stop()
		}
	}()

	output, err := s.runScraperExecutable(scraperCtx, state.ID, executablePath, scraperArgs)

	// If the scraper reports an error but we already processed all trading days, downgrade to success.
	if err != nil && totalTradingDays > 0 {
		beat := heartbeat.Snapshot()
		if beat.completed >= totalTradingDays {
			if s.logger != nil {
				s.logger.Warn("Scraper error after completing all trading days - treating as success",
					slog.String("operation_id", state.ID),
					slog.String("error", err.Error()),
					slog.Int("completed_trading_days", beat.completed),
					slog.Int("total_trading_days", totalTradingDays))
			}
			err = nil
		}
	}

	if output != "" {
		withStepMetadata(stepState, func(md map[string]interface{}) {
			md["scraper_output"] = truncateOutput(output, maxScraperLogBytes)
		})
	}
	if err != nil {
		lastBeat := heartbeat.Snapshot()
		outputTail := lastLines(output, 10)
		if outputTail == "" {
			outputTail = "scraper exited with no stdout/stderr output"
		}
		withStepMetadata(stepState, func(md map[string]interface{}) {
			md["scraper_error"] = err.Error()
			if output != "" {
				md["scraper_output_tail"] = outputTail
			}
			md["last_heartbeat_completed"] = lastBeat.completed
			md["last_heartbeat_expected"] = totalTradingDays
			if lastBeat.lastFile != "" {
				md["last_heartbeat_file"] = lastBeat.lastFile
			}
			if !lastBeat.lastTime.IsZero() {
				md["last_heartbeat_at"] = lastBeat.lastTime.Format(time.RFC3339)
				md["last_heartbeat_seconds_ago"] = int(time.Since(lastBeat.lastTime).Seconds())
			}
			if output == "" {
				md["scraper_output_tail"] = "scraper exited with no stdout/stderr output"
			}
		})
		if s.logger != nil {
			s.logger.Error("Scraper execution failed with diagnostics",
				slog.String("operation_id", state.ID),
				slog.String("error", err.Error()),
				slog.String("output_tail", outputTail),
				slog.Int("last_heartbeat_completed", lastBeat.completed),
				slog.Int("last_heartbeat_expected", totalTradingDays),
				slog.String("last_heartbeat_file", lastBeat.lastFile),
				slog.String("last_heartbeat_at", func() string {
					if lastBeat.lastTime.IsZero() {
						return ""
					}
					return lastBeat.lastTime.Format(time.RFC3339)
				}()))
		}
		if watchdog != nil && watchdog.Triggered() {
			lastBeat := heartbeat.Snapshot()
			withStepMetadata(stepState, func(md map[string]interface{}) {
				md["watchdog_status"] = "timeout"
				if !lastBeat.lastTime.IsZero() {
					md["watchdog_last_seen_at"] = lastBeat.lastTime.Format(time.RFC3339)
					md["watchdog_seconds_idle"] = int(time.Since(lastBeat.lastTime).Seconds())
				}
				if lastBeat.lastFile != "" {
					md["watchdog_last_file"] = lastBeat.lastFile
				}
			})
			allFiles, _, _ := s.collectDownloadedFiles(downloadDir, time.Time{})
			downloadedDates := s.extractDatesFromFilenames(allFiles)
			skippedFiles := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)
			// Use holiday-aware progress calculation
			progressPercent := s.calculateHolidayAwareProgress(lastBeat.completed, totalTradingDays, len(skippedFiles))
			timeoutMessage := "Scraper watchdog timeout - no files detected"
			timeoutTelemetry := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
				FromDate:         fromDate,
				ToDate:           toDate,
				DownloadedFiles:  allFiles,
				SkippedFiles:     skippedFiles,
				TradingDaysTotal: totalTradingDays,
				TradingDaysCompleted: func(a, b int) int {
					if a < b {
						return a
					}
					return b
				}(lastBeat.completed, totalTradingDays),
				CurrentFile:     lastBeat.lastFile,
				StageMessage:    timeoutMessage,
				ProgressPercent: progressPercent,
				WatchdogStatus:  "timeout",
				Status:          "failed",
			})
			s.applyTelemetryToStep(stepState, timeoutTelemetry)
			if s.broadcaster != nil {
				s.broadcaster.UpdateProgressWithMetadata(progressPercent, timeoutMessage, timeoutTelemetry)
			}
			s.emitScrapingTelemetry(state.ID, "failed", timeoutTelemetry)
			completeStage("failed", timeoutMessage, timeoutTelemetry)
			return fmt.Errorf("scraper watchdog timeout: no downloads for %s", watchdog.Timeout())
		}
		failureMessage := "Scraper execution failed"
		partialFiles, partialTotalMB, _ := s.collectDownloadedFiles(downloadDir, stageStart)
		downloadedDates := s.extractDatesFromFilenames(partialFiles)
		skippedFiles := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)
		beat := heartbeat.Snapshot()
		completedTradingDays := func(a, b int) int {
			if a < b {
				return a
			}
			return b
		}(beat.completed, totalTradingDays)
		// Use holiday-aware progress calculation (skippedFiles should be calculated from beat data or passed in)
		// Use detected skipped files for accurate progress calculation
		progressPercent := s.calculateHolidayAwareProgress(completedTradingDays, totalTradingDays, len(skippedFiles))
		failureTelemetry := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
			FromDate:             fromDate,
			ToDate:               toDate,
			DownloadedFiles:      partialFiles,
			SkippedFiles:         skippedFiles,
			TradingDaysTotal:     totalTradingDays,
			TradingDaysCompleted: completedTradingDays,
			CurrentFile:          beat.lastFile,
			TotalMB:              partialTotalMB,
			ScrapedFilesCount:    len(partialFiles),
			StageMessage:         failureMessage,
			ProgressPercent:      progressPercent,
			Phase:                "failed",
			Status:               "failed",
		})
		s.applyTelemetryToStep(stepState, failureTelemetry)
		if s.broadcaster != nil {
			s.broadcaster.UpdateProgressWithMetadata(progressPercent, failureMessage, failureTelemetry)
		}
		s.emitScrapingTelemetry(state.ID, "failed", failureTelemetry)
		completeStage("failed", failureMessage, failureTelemetry)
		close(scraperDone)
		return fmt.Errorf("scraper execution failed: %w", err)
	}

	state.SetContext(ContextKeyScraperSuccess, true)
	close(scraperDone)

	s.updateProgress(stepState, 80, "Scraper completed, verifying downloads...")

	downloadedFiles, totalSizeMB, err := s.collectDownloadedFiles(downloadDir, stageStart)
	if err != nil && s.logger != nil {
		s.logger.Warn("Failed to collect downloaded files",
			slog.String("operation_id", state.ID),
			slog.String("error", err.Error()))
	}

	// Validate downloaded files and remove corrupted ones
	validFiles, corruptedFiles := s.validateAndProcessDownloads(downloadDir, downloadedFiles)
	if len(corruptedFiles) > 0 && s.logger != nil {
		s.logger.Warn("Corrupted files detected and removed",
			slog.String("operation_id", state.ID),
			slog.Int("corrupted_count", len(corruptedFiles)),
			slog.Int("valid_count", len(validFiles)))
	}

	// Detect skipped trading days for segmented UI if dates available
	downloadedDates := s.extractDatesFromFilenames(validFiles)
	skippedFiles := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)
	if len(validFiles) > 1 {
		sortFilesChronologically(validFiles)
	}
	completedTradingDays := len(validFiles)
	if totalTradingDays > 0 && completedTradingDays > totalTradingDays {
		completedTradingDays = totalTradingDays
	}

	// Update final progress
	s.updateProgress(stepState, 100, "Scraping completed successfully")

	// Store scraped files info
	finalMetadata := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
		FromDate:             fromDate,
		ToDate:               toDate,
		DownloadedFiles:      validFiles,
		SkippedFiles:         skippedFiles,
		TradingDaysTotal:     totalTradingDays,
		TradingDaysCompleted: completedTradingDays,
		CurrentFile:          "",
		TotalMB:              totalSizeMB,
		ScrapedFilesCount:    len(validFiles),
		StageMessage:         "Scraping completed successfully",
		ProgressPercent:      100,
		Phase:                "complete",
		Status:               "completed",
	})
	s.applyTelemetryToStep(stepState, finalMetadata)
	// Emit a final scraping telemetry update marked as completed to ensure clients see 100% and completed status.
	s.emitScrapingTelemetry(state.ID, "completed", finalMetadata)
	completeStage("completed", "Scraping completed successfully", finalMetadata)

	if s.logger != nil {
		s.logger.Info("Scraping completed successfully",
			slog.String("operation_id", state.ID),
			slog.Int("files_scraped", len(validFiles)),
			slog.Float64("total_size_mb", totalSizeMB))
	}

	return nil
}

// updateProgress updates progress through the decentralized VelocityTrackingBroadcaster
func (s *ScrapingStage) updateProgress(stepState *StepState, progress int, message string) {
	if stepState == nil {
		return
	}

	var metadataSnapshot map[string]interface{}
	withStepMetadata(stepState, func(md map[string]interface{}) {
		setStageMessageMetadata(md, message, progress)
		// Add simplified progress calculation indicator
		md["progress_calculation"] = "basic_percentage"
		metadataSnapshot = cloneMetadata(md)
	})
	stepState.UpdateProgress(float64(progress), message)

	// Use decentralized velocity tracking broadcaster for all updates
	if s.broadcaster != nil {
		// Update through the broadcaster - single source of truth
		// Pass the metadata to ensure frontend receives all the details
		s.broadcaster.UpdateProgressWithMetadata(progress, message, metadataSnapshot)
	}
}

// checkDiskSpace checks available disk space using OS-native system calls
func (s *ScrapingStage) checkDiskSpace(targetPath string) (*DiskSpaceInfo, error) {
	if targetPath == "" {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}

	info, err := queryDiskUsage(absPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// bytesToGB converts bytes to gigabytes (used by disk_usage_windows.go)
func bytesToGB(value uint64) float64 {
	return float64(value) / (1024 * 1024 * 1024)
}

// buildScraperArgs constructs command-line arguments for the scraper executable
func (s *ScrapingStage) buildScraperArgs(fromDate, toDate, mode, downloadDir string) []string {
	// Use -headless=true as a single argument so Go's flag parser doesn't stop parsing subsequent flags.
	args := []string{"-out", downloadDir, "-headless=true"}

	if strings.TrimSpace(mode) != "" {
		args = append(args, "-mode", mode)
	}
	if strings.TrimSpace(fromDate) != "" {
		args = append(args, "-from", fromDate, "-actual-from", fromDate)
	}
	if strings.TrimSpace(toDate) != "" {
		args = append(args, "-to", toDate, "-actual-to", toDate)
	}

	return args
}

// runScraperExecutable runs the scraper executable with the specified arguments and captures its output
func (s *ScrapingStage) runScraperExecutable(ctx context.Context, operationID, executablePath string, args []string) (string, error) {
	if executablePath == "" {
		return "", fmt.Errorf("scraper executable path is empty")
	}

	ctxToUse := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// Use default timeout when no deadline is set
		ctxToUse, cancel = context.WithTimeout(ctx, DefaultScrapingTimeout)

		if s.logger != nil {
			s.logger.Info("Using default scraper timeout",
				slog.String("operation_id", operationID),
				slog.Duration("default_timeout", DefaultScrapingTimeout))
		}
	}
	defer cancel()

	cmd := exec.CommandContext(ctxToUse, executablePath, args...)
	cmd.Dir = s.executableDir

	env := os.Environ()
	if operationID != "" {
		env = append(env, fmt.Sprintf("ISX_OPERATION_ID=%s", operationID))
	}
	cmd.Env = env

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if s.logger != nil {
		s.logger.Info("Launching scraper executable",
			slog.String("operation_id", operationID),
			slog.String("executable", executablePath),
			slog.Any("args", args))
	}

	err := cmd.Run()
	combinedOutput := strings.TrimSpace(stdoutBuf.String() + "\n" + stderrBuf.String())

	if err != nil {
		if s.logger != nil {
			s.logger.Error("Scraper executable failed",
				slog.String("operation_id", operationID),
				slog.String("executable", executablePath),
				slog.Any("args", args),
				slog.String("error", err.Error()))
		}
		return combinedOutput, err
	}

	// Parse and broadcast holiday updates from scraper output
	s.parseAndBroadcastHolidayUpdates(combinedOutput, operationID)

	if s.logger != nil {
		s.logger.Info("Scraper executable completed",
			slog.String("operation_id", operationID),
			slog.String("executable", executablePath))
	}

	return combinedOutput, nil
}

// resolveScraperExecutable locates the scraper executable in the configured directory
func (s *ScrapingStage) resolveScraperExecutable() (string, error) {
	if s.executableDir == "" {
		return "", fmt.Errorf("executable directory not configured for scraping stage")
	}

	var candidates []string
	if strings.EqualFold(runtime.GOOS, "windows") {
		candidates = []string{"scraper.exe", "scraper"}
	} else {
		candidates = []string{"scraper", "scraper.exe"}
	}

	for _, candidate := range candidates {
		path := filepath.Join(s.executableDir, candidate)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("scraper executable not found in %s", s.executableDir)
}

// validateFileCount validates that downloaded files count is accurate and provides expected total
func (s *ScrapingStage) validateFileCount(fromDate, toDate string) (int, int) {
	if fromDate == "" || toDate == "" {
		return 0, 0
	}

	// Count trading days (excluding Friday/Saturday for Iraqi market)
	tradingDays := s.countTradingDays(fromDate, toDate)

	// Validate trading days count is reasonable
	if tradingDays <= 0 {
		if s.logger != nil {
			s.logger.Warn("Invalid date range for file count validation",
				"from_date", fromDate,
				"to_date", toDate,
				"trading_days", tradingDays)
		}
		return 0, 0
	}

	if s.logger != nil {
		s.logger.Info("File count validation completed",
			"from_date", fromDate,
			"to_date", toDate,
			"expected_files", tradingDays,
			"validation_method", "trading_days_count")
	}

	return tradingDays, tradingDays
}

// collectDownloadedFiles scans the download directory for .xlsx files modified since the specified time
func (s *ScrapingStage) collectDownloadedFiles(downloadDir string, since time.Time) ([]string, float64, error) {
	fileNames := make([]string, 0)
	var totalSizeMB float64

	cutoff := since
	includeAll := cutoff.IsZero()
	if includeAll {
		cutoff = time.Time{}
	} else {
		cutoff = cutoff.Add(-2 * time.Second)
	}

	err := filepath.WalkDir(downloadDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".xlsx") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !includeAll && info.ModTime().Before(cutoff) {
			return nil
		}

		sizeMB := float64(info.Size()) / (1024 * 1024)

		fileNames = append(fileNames, d.Name())
		totalSizeMB += sizeMB

		return nil
	})

	if err != nil {
		return fileNames, totalSizeMB, err
	}

	return fileNames, totalSizeMB, nil
}

// getConfigString retrieves a string configuration value from the operation state
func (s *ScrapingStage) getConfigString(state *OperationState, key string) string {
	if state == nil {
		return ""
	}

	if value, ok := state.GetConfig(key); ok {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str)
		}
	}

	return ""
}

// truncateOutput truncates the output string to the specified limit, adding "..." if truncated
func truncateOutput(output string, limit int) string {
	if limit <= 0 || len(output) <= limit {
		return output
	}
	if limit <= 3 {
		return output[:limit]
	}
	return output[:limit-3] + "..."
}

// lastLines returns up to n lines from the end of the string for diagnostics.
func lastLines(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(strings.Join(lines[len(lines)-n:], "\n"))
}

// extractDatesFromFilenames extracts unique dates from the given list of filenames
func (s *ScrapingStage) extractDatesFromFilenames(files []string) map[string]struct{} {
	dates := make(map[string]struct{})
	for _, name := range files {
		if date, ok := extractDateFromFilename(name); ok {
			dates[date] = struct{}{}
		}
	}
	return dates
}

// detectSkippedTradingDays identifies trading days that were skipped based on downloaded dates
func (s *ScrapingStage) detectSkippedTradingDays(fromDate, toDate string, downloadedDates map[string]struct{}) []string {
	if fromDate == "" || toDate == "" || len(downloadedDates) < 2 {
		return nil
	}

	start, err := time.Parse(iso8601Date, fromDate)
	if err != nil {
		return nil
	}
	end, err := time.Parse(iso8601Date, toDate)
	if err != nil {
		return nil
	}
	if end.Before(start) {
		start, end = end, start
	}

	downloadedList := make([]time.Time, 0, len(downloadedDates))
	for date := range downloadedDates {
		if parsed, err := time.Parse(iso8601Date, date); err == nil {
			if parsed.Before(start) || parsed.After(end) {
				continue
			}
			downloadedList = append(downloadedList, parsed)
		}
	}
	if len(downloadedList) < 2 {
		return nil
	}

	sort.Slice(downloadedList, func(i, j int) bool {
		return downloadedList[i].Before(downloadedList[j])
	})

	skipped := make([]string, 0)
	prev := downloadedList[0]

	for i := 1; i < len(downloadedList); i++ {
		next := downloadedList[i]
		for day := prev.AddDate(0, 0, 1); day.Before(next); day = day.AddDate(0, 0, 1) {
			weekday := day.Weekday()
			if weekday == time.Friday || weekday == time.Saturday {
				continue
			}
			key := day.Format(iso8601Date)
			if _, exists := downloadedDates[key]; exists {
				continue
			}
			skipped = append(skipped, key)
		}
		prev = next
	}

	return skipped
}

var filenameDateMatcher = regexp.MustCompile(`(\d{4})[-_\s](\d{2})[-_\s](\d{2})`)

// extractDateFromFilename extracts a date in YYYY-MM-DD format from the given filename
func extractDateFromFilename(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	matches := filenameDateMatcher.FindStringSubmatch(name)
	if len(matches) != 4 {
		return "", false
	}
	return fmt.Sprintf("%s-%s-%s", matches[1], matches[2], matches[3]), true
}

func (s *ScrapingStage) countTradingDays(fromDate, toDate string) int {
	dates := s.enumerateTradingDays(fromDate, toDate)
	return len(dates)
}

// enumerateTradingDays returns a list of trading days (excluding Fridays and Saturdays) between fromDate and toDate
func (s *ScrapingStage) enumerateTradingDays(fromDate, toDate string) []string {
	if fromDate == "" || toDate == "" {
		return nil
	}
	start, err := time.Parse(iso8601Date, fromDate)
	if err != nil {
		return nil
	}
	end, err := time.Parse(iso8601Date, toDate)
	if err != nil {
		return nil
	}
	if end.Before(start) {
		start, end = end, start
	}
	var days []string
	for day := end; !day.Before(start); day = day.AddDate(0, 0, -1) {
		if weekday := day.Weekday(); weekday == time.Friday || weekday == time.Saturday {
			continue
		}
		days = append(days, day.Format(iso8601Date))
	}
	return days
}

// filterFilesInRange filters the given list of filenames to only include those within the specified date range
func (s *ScrapingStage) filterFilesInRange(files []string, fromDate, toDate string) []string {
	if fromDate == "" || toDate == "" {
		return files
	}
	start, err := time.Parse(iso8601Date, fromDate)
	if err != nil {
		return files
	}
	end, err := time.Parse(iso8601Date, toDate)
	if err != nil {
		return files
	}
	if end.Before(start) {
		start, end = end, start
	}

	filtered := make([]string, 0, len(files))
	for _, name := range files {
		if date, ok := extractDateFromFilename(name); ok {
			if t, err := time.Parse(iso8601Date, date); err == nil {
				if !t.Before(start) && !t.After(end) {
					filtered = append(filtered, name)
					continue
				}
			}
		}
	}
	return filtered
}

// sortFilesChronologically sorts the given list of filenames in chronological order based on embedded dates
func sortFilesChronologically(files []string) {
	sort.SliceStable(files, func(i, j int) bool {
		leftDate, leftOK := extractDateFromFilename(files[i])
		rightDate, rightOK := extractDateFromFilename(files[j])

		switch {
		case leftOK && rightOK:
			return leftDate < rightDate
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return files[i] < files[j]
		}
	})
}

// findMostRecentFile finds the most recently modified file from the given list in the download directory
func (s *ScrapingStage) findMostRecentFile(downloadDir string, files []string) string {
	var latestFile string
	var latestTime time.Time

	for _, name := range files {
		info, err := os.Stat(filepath.Join(downloadDir, name))
		if err != nil {
			continue
		}
		if latestFile == "" || info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = name
		}
	}

	if latestFile == "" && len(files) > 0 {
		latestFile = files[len(files)-1]
	}

	return latestFile
}

// prepareScraperContext prepares a context with timeout for scraper execution
func (s *ScrapingStage) prepareScraperContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, deadline := parent.Deadline(); deadline {
		return parent, func() {}
	}
	if timeout <= 0 {
		timeout = DefaultScrapingTimeout
	}
	return context.WithTimeout(parent, timeout)
}

// scrapingTelemetry holds detailed telemetry data for the scraping stage
type scrapingTelemetry struct {
	FromDate             string
	ToDate               string
	DownloadedFiles      []string
	SkippedFiles         []string
	TradingDaysTotal     int
	TradingDaysCompleted int
	CurrentFile          string
	TotalMB              float64
	ScrapedFilesCount    int
	WatchdogStatus       string
	StageMessage         string
	ProgressPercent      int
	Phase                string
	Status               string
}

// buildScrapingTelemetry constructs a telemetry map that keeps canonical progress fields
// via StageTelemetry and namespaces scraping-specific details under the "scraping" key
// (also mirrored at top-level for backward compatibility).
func (s *ScrapingStage) buildScrapingTelemetry(stepState *StepState, data *scrapingTelemetry) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{}
	}

	holidaysDetected := len(data.SkippedFiles)

	// Derive adjusted totals and remaining using holiday awareness
	adjustedTotal := data.TradingDaysTotal - holidaysDetected
	if adjustedTotal < 0 {
		adjustedTotal = 0
	}
	remaining := 0
	if adjustedTotal > data.TradingDaysCompleted {
		remaining = adjustedTotal - data.TradingDaysCompleted
	}

	progressPercent := clampProgress(data.ProgressPercent)
	if progressPercent == 0 && data.TradingDaysTotal > 0 {
		progressPercent = s.calculateHolidayAwareProgress(data.TradingDaysCompleted, data.TradingDaysTotal, holidaysDetected)
	}
	stageMessage := data.StageMessage
	if stageMessage == "" {
		stageMessage = "Stage running"
	}

	phase := StagePhase(strings.ToLower(data.Phase))
	if phase == "" {
		if progressPercent >= 100 {
			phase = StagePhaseCompleted
		} else {
			phase = StagePhaseRunning
		}
	}
	status := strings.ToLower(data.Status)
	if status == "" {
		switch {
		case phase == StagePhaseCompleted || progressPercent >= 100:
			status = "completed"
		case phase == StagePhaseFailed:
			status = "failed"
		default:
			status = "running"
		}
	}

	// Canonical stage telemetry payload
	stageTelemetry := NewStageTelemetry(
		StageIDScraping,
		stepState,
		phase,
		stageMessage,
		float64(progressPercent),
		data.TradingDaysCompleted,
		data.TradingDaysTotal,
		data.CurrentFile,
		nil,
	)

	md := make(map[string]interface{})
	stageTelemetry.ApplyToMetadata(md)
	md["status"] = status

	// Scraping-specific section (namespaced) for UI consumers
	scrapingSection := map[string]interface{}{
		"from_date":                   data.FromDate,
		"to_date":                     data.ToDate,
		"downloaded_files":            data.DownloadedFiles,
		"skipped_files":               data.SkippedFiles,
		"trading_days_total":          data.TradingDaysTotal,
		"trading_days_completed":      data.TradingDaysCompleted,
		"trading_days_remaining":      remaining,
		"adjusted_trading_days_total": adjustedTotal,
		"scraped_files_count":         data.ScrapedFilesCount,
		"progress_percent":            progressPercent,
		"stage_message":               stageMessage,
		"current_file":                data.CurrentFile,
		"total_downloaded_mb":         data.TotalMB,
		"holidays_detected":           holidaysDetected,
		"watchdog_status":             data.WatchdogStatus,
		"telemetry_missing":           data.TradingDaysTotal == 0 || data.TradingDaysCompleted == 0,
		"files_downloaded":            len(data.DownloadedFiles),
		"files_processed":             data.TradingDaysCompleted,
		"scraping_status":             status,
	}

	md["scraping"] = scrapingSection

	// Backward-compatible flattening of scraping-specific fields
	for k, v := range scrapingSection {
		md[k] = v
	}

	return md
}

// applyTelemetryToStep merges the given telemetry data into the step's metadata
func (s *ScrapingStage) applyTelemetryToStep(stepState *StepState, telemetry map[string]interface{}) {
	if stepState == nil || telemetry == nil {
		return
	}
	withStepMetadata(stepState, func(md map[string]interface{}) {
		for k, v := range telemetry {
			md[k] = v
		}
	})
}

// withStepMetadata safely accesses and modifies the step's metadata map
func withStepMetadata(stepState *StepState, fn func(map[string]interface{})) {
	if stepState == nil || fn == nil {
		return
	}
	stepState.mu.Lock()
	defer stepState.mu.Unlock()
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	fn(stepState.Metadata)
}

// setStageMessageMetadata sets the stage message and related metadata fields
func setStageMessageMetadata(metadata map[string]interface{}, message string, progress int) {
	if metadata == nil {
		return
	}
	metadata["stage_message"] = message
	metadata["phase_message"] = message
	phase := "running"
	if progress >= 100 {
		phase = "complete"
	} else if progress <= 0 {
		phase = "starting"
	}
	metadata["phase"] = phase
	metadata["current_phase"] = phase
	metadata["telemetry_missing"] = false
}

// emitScrapingTelemetry sends the scraping telemetry data through the WebSocket manager
func (s *ScrapingStage) emitScrapingTelemetry(operationID, status string, telemetry map[string]interface{}) {
	if telemetry == nil || s == nil || s.options == nil || s.options.WebSocketManager == nil {
		return
	}
	if _, exists := telemetry["operation_id"]; !exists {
		telemetry["operation_id"] = operationID
	}
	s.options.WebSocketManager.BroadcastUpdate("scraping:telemetry", operationID, status, telemetry)
}

// downloadHeartbeat tracks the last activity time, last file processed, and count of completed downloads
type downloadHeartbeat struct {
	mu        sync.RWMutex
	lastTime  time.Time
	lastFile  string
	completed int
}

// Update updates the heartbeat with new completion information
func (h *downloadHeartbeat) Update(completed int, lastFile string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.completed = completed
	h.lastFile = lastFile
	h.lastTime = time.Now()
}

// heartbeatSnapshot captures a snapshot of the downloadHeartbeat's state
type heartbeatSnapshot struct {
	lastTime  time.Time
	lastFile  string
	completed int
}

// newDownloadHeartbeat creates a new instance of downloadHeartbeat with the current time as the last activity time
func newDownloadHeartbeat() *downloadHeartbeat {
	return &downloadHeartbeat{
		lastTime: time.Now(),
	}
}

// Touch updates the last activity time, last file processed, and count of completed downloads
func (h *downloadHeartbeat) Touch(file string, count int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastTime = time.Now()
	h.lastFile = file
	h.completed = count
}

// Snapshot captures a snapshot of the downloadHeartbeat's state
func (h *downloadHeartbeat) Snapshot() heartbeatSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return heartbeatSnapshot{
		lastTime:  h.lastTime,
		lastFile:  h.lastFile,
		completed: h.completed,
	}
}

// scraperWatchdog monitors the scraper's activity and triggers cancellation if a timeout occurs
type scraperWatchdog struct {
	triggered           atomic.Bool
	stopCtx             context.CancelFunc
	cancelScraper       context.CancelFunc
	cancelMonitor       context.CancelFunc
	logger              *slog.Logger
	expectedTradingDays int
	adaptiveWatchdog    *AdaptiveWatchdog // New field for adaptive watchdog
}

// Stop stops the scraperWatchdog, cancelling its internal context
func (w *scraperWatchdog) Stop() {
	if w == nil {
		return
	}
	if w.stopCtx != nil {
		w.stopCtx()
	}
}

// Triggered returns whether the scraperWatchdog has been triggered
func (w *scraperWatchdog) Triggered() bool {
	if w == nil {
		return false
	}
	return w.triggered.Load()
}

// Timeout returns the timeout duration set for the scraperWatchdog
func (w *scraperWatchdog) Timeout() time.Duration {
	if w == nil {
		return 0
	}
	// Read current timeout from adaptive watchdog
	if w.adaptiveWatchdog != nil {
		return w.adaptiveWatchdog.GetTimeout()
	}
	// Fallback when no adaptive watchdog is available
	return DefaultScrapingTimeout
}

// startScraperWatchdog starts an adaptive watchdog that uses the new AdaptiveWatchdog with 10s per file dynamic timeout
func (s *ScrapingStage) startScraperWatchdog(ctx context.Context, heartbeat *downloadHeartbeat, cancelScraper, cancelMonitor context.CancelFunc, tradingDays int) *scraperWatchdog {
	if heartbeat == nil || cancelScraper == nil {
		return nil
	}

	// Create adaptive watchdog with 10s per file timeout as requested
	adaptiveWatchdog := NewAdaptiveWatchdog(tradingDays, StageIDScraping, s.logger)

	// Create wrapper watchdog for compatibility with existing code
	wd := &scraperWatchdog{
		cancelScraper:       cancelScraper,
		cancelMonitor:       cancelMonitor,
		logger:              s.logger,
		expectedTradingDays: tradingDays,
		adaptiveWatchdog:    adaptiveWatchdog,
	}

	// Start the adaptive watchdog monitoring
	adaptiveWatchdog.StartMonitoring(ctx, cancelScraper)

	// Start legacy monitoring for backward compatibility
	watchCtx, stop := context.WithCancel(ctx)
	wd.stopCtx = stop

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		defer stop()

		for {
			select {
			case <-watchCtx.Done():
				// Stop adaptive watchdog when context is done
				adaptiveWatchdog.StopMonitoring()
				return
			case <-ticker.C:
				snap := heartbeat.Snapshot()

				// Update adaptive watchdog with latest activity
				if !snap.lastTime.IsZero() {
					adaptiveWatchdog.UpdateActivity()
				}

				// Check for completion
				if wd.expectedTradingDays > 0 && snap.completed >= wd.expectedTradingDays {
					// Record successful performance
					adaptiveWatchdog.RecordPerformance(time.Since(snap.lastTime), true)
					return
				}

				// Check if adaptive watchdog triggered timeout
				if adaptiveWatchdog.Triggered() {
					if wd.triggered.CompareAndSwap(false, true) {
						if wd.logger != nil {
							wd.logger.Warn("adaptive scraper watchdog timeout",
								slog.Duration("timeout", adaptiveWatchdog.GetTimeout()),
								slog.String("last_file", snap.lastFile),
								slog.Time("last_activity", snap.lastTime))
						}

						// Record timeout performance
						adaptiveWatchdog.RecordPerformance(time.Since(snap.lastTime), false)

						cancelScraper()
						if wd.cancelMonitor != nil {
							wd.cancelMonitor()
						}
					}
					return
				}
			}
		}
	}()

	return wd
}

// monitorDownloadProgress continuously monitors the download progress and updates the step state and telemetry accordingly
func (s *ScrapingStage) monitorDownloadProgress(ctx context.Context, stepState *StepState, operationID, downloadDir, fromDate, toDate string, since time.Time, tradingDays int, heartbeat *downloadHeartbeat, scraperDone <-chan struct{}, completeStage func(status, message string, metadata map[string]interface{})) {
	// Initialize with fast update interval for initial responsiveness
	updateInterval := 1 * time.Second
	ticker := time.NewTicker(updateInterval)
	defer func() {
		ticker.Stop() // will stop *the last* ticker
	}()

	lastCount := -1
	lastSkipped := -1
	consecutiveNoChange := 0

	// Optimized update interval calculation
	calculateOptimalInterval := func(progressPercent int, hasNewFiles bool, consecutiveNoChange int) time.Duration {
		// Phase-based interval optimization:
		// Starting phase (0-10%): 1s updates for responsive UI
		// Active phase (10-80%): 2-3s updates when downloading, 5s when idle
		// Completion phase (80-95%): 2s updates for final progress
		// Final phase (95-100%): 1s updates for completion confirmation

		if progressPercent < 10 {
			return 1 * time.Second // Fast updates during initialization
		}

		if progressPercent >= 95 {
			return 1 * time.Second // Fast updates near completion
		}

		if hasNewFiles {
			// Active downloading - moderate update frequency
			if progressPercent < 80 {
				return 2 * time.Second // Standard interval during bulk download
			} else {
				return 1 * time.Second // Faster updates near completion
			}
		} else {
			// No new files - reduce update frequency to prevent noise
			if consecutiveNoChange < 3 {
				return 3 * time.Second // Allow some time for downloads to appear
			} else {
				return 5 * time.Second // Slower updates during idle periods
			}
		}
	}

	// Update ticker interval dynamically
	updateTicker := func(newInterval time.Duration) {
		if newInterval != updateInterval {
			ticker.Stop()
			updateInterval = newInterval
			ticker = time.NewTicker(updateInterval)

			if s.logger != nil {
				s.logger.Debug("Adjusted WebSocket update interval",
					slog.String("operation_id", operationID),
					slog.Duration("new_interval", updateInterval))
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-scraperDone:
			if s.logger != nil {
				s.logger.Debug("scraperDone received - stopping monitor loop",
					slog.String("operation_id", operationID))
			}
			return
		case <-ticker.C:
			fileNames, _, err := s.collectDownloadedFiles(downloadDir, since)
			if err != nil {
				if s.logger != nil {
					s.logger.Warn("Download progress scan failed",
						slog.String("operation_id", operationID),
						slog.String("error", err.Error()))
				}
				continue
			}

			filtered := s.filterFilesInRange(fileNames, fromDate, toDate)
			if len(filtered) > 1 {
				sortFilesChronologically(filtered)
			}

			count := len(filtered)

			downloadedDates := s.extractDatesFromFilenames(filtered)
			skipped := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)

			var completedTradingDays int
			completedTradingDays = count
			if tradingDays > 0 && completedTradingDays > tradingDays {
				completedTradingDays = tradingDays
			}

			currentFile := ""
			if count > 0 {
				currentFile = s.findMostRecentFile(downloadDir, filtered)
			}

			// Track if we have new files for interval optimization
			hasNewFiles := count > lastCount
			if !hasNewFiles {
				consecutiveNoChange++
			} else {
				consecutiveNoChange = 0
			}

			// Skip update if no changes and we're in an optimized idle period
			if count == lastCount && len(skipped) == lastSkipped && consecutiveNoChange > 3 {
				// If scraper is done and no progress, finalize to avoid hanging UI
				select {
				case <-scraperDone:
					if s.logger != nil {
						s.logger.Debug("scraperDone received during idle window - exiting monitor",
							slog.String("operation_id", operationID))
					}
					return
				default:
					continue
				}
			}
			lastCount = count
			lastSkipped = len(skipped)

			var progress int
			var message string

			// Use holiday-aware progress calculation
			progress = s.calculateHolidayAwareProgress(completedTradingDays, tradingDays, len(skipped))

			// Update message with holiday information
			message = "Downloading files..."
			if len(skipped) > 0 {
				message = fmt.Sprintf("Downloading files... (%d/%d trading days, %d holidays)", completedTradingDays, tradingDays, len(skipped))
			} else if tradingDays > 0 && completedTradingDays < tradingDays {
				message = fmt.Sprintf("Downloading files... (%d/%d trading days processed)", completedTradingDays, tradingDays)
			} else if tradingDays <= 0 {
				if count == 1 {
					message = "Downloading files... (1 file processed)"
				} else {
					message = fmt.Sprintf("Downloading files... (%d files processed)", count)
				}
			}

			// Update StepState.Progress to keep it in sync
			stepState.UpdateProgress(float64(progress), message)

			// Account for both downloaded trading days and detected holidays
			accounted := completedTradingDays + len(skipped)
			status := "running"
			if len(skipped) > 0 {
				message = fmt.Sprintf("Downloading files... (%d/%d trading days, %d holidays)", completedTradingDays, tradingDays, len(skipped))
			} else if tradingDays > 0 && completedTradingDays < tradingDays {
				message = fmt.Sprintf("Downloading files... (%d/%d trading days processed)", completedTradingDays, tradingDays)
			}

			if tradingDays > 0 && accounted >= tradingDays {
				progress = 100
				message = fmt.Sprintf("Scraping completed (%d/%d trading days)", completedTradingDays, tradingDays)
				// Update final step state with proper status
				stepState.Status = StepStatusCompleted
				stepState.Progress = float64(progress)
			}

			// Calculate and apply optimal update interval based on current progress
			optimalInterval := calculateOptimalInterval(progress, hasNewFiles, consecutiveNoChange)
			updateTicker(optimalInterval)

			metadata := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
				FromDate:             fromDate,
				ToDate:               toDate,
				DownloadedFiles:      filtered,
				SkippedFiles:         skipped,
				TradingDaysTotal:     tradingDays,
				TradingDaysCompleted: completedTradingDays,
				CurrentFile:          currentFile,
				StageMessage:         message,
				ProgressPercent:      progress,
				Status:               status,
			})

			s.applyTelemetryToStep(stepState, metadata)

			if status == "completed" {
				if s.broadcaster != nil {
					s.broadcaster.UpdateProgressWithMetadata(progress, message, metadata)
				}
				s.emitScrapingTelemetry(operationID, status, metadata)
				if s.logger != nil {
					s.logger.Info("Download monitor detected scraping completion",
						slog.String("operation_id", operationID),
						slog.Int("trading_days_completed", completedTradingDays),
						slog.Int("trading_days_total", tradingDays))
				}
				return
			}

			if s.broadcaster != nil {
				s.broadcaster.UpdateProgressWithMetadata(progress, message, metadata)
			}
			s.emitScrapingTelemetry(operationID, status, metadata)
		}
	}
}

// monitorDownloadProgressEventDriven uses event-driven file monitoring for instant progress updates
// Replaces the polling-based approach with fsnotify for 10x faster file detection
func (s *ScrapingStage) monitorDownloadProgressEventDriven(ctx context.Context, stepState *StepState, operationID, downloadDir, fromDate, toDate string, since time.Time, tradingDays int, heartbeat *downloadHeartbeat, scraperDone <-chan struct{}, completeStage func(status, message string, metadata map[string]interface{})) {
	// Initialize event-driven file monitor
	fileMonitor, err := NewEventDrivenFileMonitor(downloadDir, s.logger)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to create file monitor", "error", err.Error(), "operation_id", operationID)
		}
		// Fallback to polling-based approach
		s.monitorDownloadProgress(ctx, stepState, operationID, downloadDir, fromDate, toDate, since, tradingDays, heartbeat, scraperDone, completeStage)
		return
	}
	defer fileMonitor.Close()

	// Set cutoff time for file filtering
	cutoff := since
	if cutoff.IsZero() {
		cutoff = time.Now().Add(-2 * time.Second) // Only look at recent files
	} else {
		cutoff = cutoff.Add(-2 * time.Second)
	}
	fileMonitor.SetCutoff(cutoff)

	// Start event monitoring
	fileMonitor.Start()

	// Initial scan to catch existing files
	initialFiles, _, err := s.collectDownloadedFiles(downloadDir, since)
	if err != nil && s.logger != nil {
		s.logger.Warn("Initial file scan failed", "error", err.Error(), "operation_id", operationID)
	}

	// Mark existing files as seen
	for _, file := range initialFiles {
		fileMonitor.MarkFileSeen(file)
	}

	// Filter and process initial files
	filtered := s.filterFilesInRange(initialFiles, fromDate, toDate)
	if len(filtered) > 1 {
		sortFilesChronologically(filtered)
	}

	lastCount := len(filtered)
	lastSkipped := 0
	consecutiveNoChange := 0

	// Update with initial state
	if len(filtered) > 0 {
		downloadedDates := s.extractDatesFromFilenames(filtered)
		skipped := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)
		lastSkipped = len(skipped)

		var completedTradingDays int
		completedTradingDays = len(filtered)
		if tradingDays > 0 && completedTradingDays > tradingDays {
			completedTradingDays = tradingDays
		}

		currentFile := s.findMostRecentFile(downloadDir, filtered)
		progress := s.calculateHolidayAwareProgress(completedTradingDays, tradingDays, len(skipped))

		message := "Monitoring file downloads..."
		if tradingDays > 0 {
			message = fmt.Sprintf("Monitoring file downloads... (%d/%d trading days)", completedTradingDays, tradingDays)
		}

		metadata := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
			FromDate:             fromDate,
			ToDate:               toDate,
			DownloadedFiles:      filtered,
			SkippedFiles:         skipped,
			TradingDaysTotal:     tradingDays,
			TradingDaysCompleted: completedTradingDays,
			CurrentFile:          currentFile,
			StageMessage:         message,
			ProgressPercent:      progress,
			Status:               "running",
		})

		s.applyTelemetryToStep(stepState, metadata)
		if s.broadcaster != nil {
			s.broadcaster.UpdateProgressWithMetadata(progress, message, metadata)
		}
		s.emitScrapingTelemetry(operationID, "running", metadata)
	}

	// Event-driven monitoring loop
	for {
		select {
		case <-ctx.Done():
			return
		case <-scraperDone:
			if s.logger != nil {
				s.logger.Debug("scraperDone received (event-driven monitor) - stopping without completing stage",
					slog.String("operation_id", operationID))
			}
			return

		case event, ok := <-fileMonitor.fileEvents:
			if !ok {
				// File events channel closed
				return
			}

			// Process only relevant file events
			if !fileMonitor.IsRelevantEvent(event) {
				continue
			}

			// Mark new file as seen
			fileMonitor.MarkFileSeen(filepath.Base(event.Name))

			// Get all seen files and process
			seenFiles := fileMonitor.GetSeenFiles()
			filtered := s.filterFilesInRange(seenFiles, fromDate, toDate)
			if len(filtered) > 1 {
				sortFilesChronologically(filtered)
			}

			count := len(filtered)
			downloadedDates := s.extractDatesFromFilenames(filtered)
			skipped := s.detectSkippedTradingDays(fromDate, toDate, downloadedDates)

			var completedTradingDays int
			completedTradingDays = count
			if tradingDays > 0 && completedTradingDays > tradingDays {
				completedTradingDays = tradingDays
			}

			currentFile := ""
			if count > 0 {
				currentFile = s.findMostRecentFile(downloadDir, filtered)
			}

			// Track changes for progress calculation
			hasNewFiles := count > lastCount
			if !hasNewFiles && len(skipped) == lastSkipped {
				consecutiveNoChange++
			} else {
				consecutiveNoChange = 0
			}

			// Update progress immediately on file events (no polling delay)
			progress := s.calculateHolidayAwareProgress(completedTradingDays, tradingDays, len(skipped))
			message := "File download detected"
			if tradingDays > 0 {
				message = fmt.Sprintf("Downloading files... (%d/%d trading days)", completedTradingDays, tradingDays)
			}

			// Add file name to message for real-time feedback
			if currentFile != "" {
				fileName := filepath.Base(currentFile)
				message = fmt.Sprintf("%s - %s", message, fileName)
			}

			// Update status using holidays as non-trading
			accounted := completedTradingDays + len(skipped)
			status := "running"
			if tradingDays > 0 && accounted >= tradingDays {
				progress = 100
				message = fmt.Sprintf("Scraping completed (%d/%d trading days)", completedTradingDays, tradingDays)
			}

			metadata := s.buildScrapingTelemetry(stepState, &scrapingTelemetry{
				FromDate:             fromDate,
				ToDate:               toDate,
				DownloadedFiles:      filtered,
				SkippedFiles:         skipped,
				TradingDaysTotal:     tradingDays,
				TradingDaysCompleted: completedTradingDays,
				CurrentFile:          currentFile,
				StageMessage:         message,
				ProgressPercent:      progress,
				Status:               status,
			})

			s.applyTelemetryToStep(stepState, metadata)

			// Update heartbeat with new file info
			if heartbeat != nil {
				heartbeat.Update(completedTradingDays, currentFile)
			}

			// Broadcast updates immediately on file events
			if s.broadcaster != nil {
				s.broadcaster.UpdateProgressWithMetadata(progress, message, metadata)
			}
			s.emitScrapingTelemetry(operationID, status, metadata)

			lastCount = count
			lastSkipped = len(skipped)

			// Log new file detection for debugging
			if s.logger != nil && hasNewFiles {
				s.logger.Debug("New file detected via event monitoring",
					slog.String("operation_id", operationID),
					slog.String("file_path", event.Name),
					slog.String("file_name", filepath.Base(event.Name)),
					slog.Int("total_files", count),
					slog.Int("progress_percent", progress))
			}

			// Do not finalize here; main Execute handles completion to avoid race with scraper result.
		}
	}
}

// GetPhase returns the current phase of the scraping operation
func (s *ScrapingStage) GetPhase(stepState *StepState) (StagePhase, string) {
	if stepState == nil {
		return StagePhasePending, "Ready to start scraping"
	}

	stepState.mu.RLock()
	progress := int(stepState.Progress)
	stepState.mu.RUnlock()
	if progress < 20 {
		return StagePhaseStarting, "Initializing scraper"
	} else if progress < 40 {
		return StagePhaseStarting, "Checking system resources"
	} else if progress < 90 {
		return StagePhaseRunning, fmt.Sprintf("Downloading files (%d%%)", progress)
	} else if progress < 100 {
		return StagePhaseCompleting, "Finalizing scraping"
	} else {
		return StagePhaseCompleted, "Scraping completed successfully"
	}
}

// RequiredInputs returns the data requirements for this stage
func (s *ScrapingStage) RequiredInputs() []DataRequirement {
	// Scraping stage doesn't require any input data
	return nil
}

// ProducedOutputs returns the data outputs this stage produces
func (s *ScrapingStage) ProducedOutputs() []DataOutput {
	return []DataOutput{
		{
			Type:        DataTypeExcelFiles,
			FilePattern: "*.xlsx",
			Location:    AbsoluteDownloadsPath(s.executableDir, ""),
			Required:    true,
			MinCount:    1,
		},
	}
}

// Dependencies returns the stage dependencies
func (s *ScrapingStage) Dependencies() []string {
	return nil // No dependencies - this is the first stage
}

// calculateHolidayAwareProgress computes progress accounting for detected holidays/non-trading days.
// This ensures accurate progress calculation when holidays reduce the expected file count.
func (s *ScrapingStage) calculateHolidayAwareProgress(completedTradingDays, totalTradingDays, holidaysDetected int) int {
	if totalTradingDays <= 0 {
		return 0
	}

	// Calculate adjusted total accounting for holidays
	adjustedTotal := totalTradingDays - holidaysDetected
	if adjustedTotal <= 0 {
		// All days are holidays - consider complete
		return 100
	}

	// Calculate progress based on adjusted total
	progress := int(math.Round((float64(completedTradingDays) / float64(adjustedTotal)) * 100))

	// Clamp progress to valid range
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}

	return progress
}

// parseAndBroadcastHolidayUpdates parses scraper output for holiday detection messages
// and broadcasts them via WebSocket for real-time frontend updates
func (s *ScrapingStage) parseAndBroadcastHolidayUpdates(output, operationID string) {
	if output == "" || operationID == "" || operationID == "temp" {
		return
	}

	// Look for SCRAPER_HOLIDAY_UPDATE lines in the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SCRAPER_HOLIDAY_UPDATE:") {
			jsonData := strings.TrimPrefix(line, "SCRAPER_HOLIDAY_UPDATE:")
			jsonData = strings.TrimSpace(jsonData)

			var holidayUpdate map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &holidayUpdate); err == nil {
				// Broadcast holiday update via WebSocket hub
				if s.options != nil && s.options.WebSocketManager != nil {
					s.options.WebSocketManager.BroadcastUpdate(
						"holiday_detected",
						operationID,
						"detected",
						holidayUpdate,
					)
				}

				// Log the holiday update
				if s.logger != nil {
					s.logger.Info("Holiday update broadcasted",
						slog.String("operation_id", operationID),
						slog.String("date", getStringFromMap(holidayUpdate, "date", "unknown")),
						slog.String("reason", getStringFromMap(holidayUpdate, "reason", "unknown")),
						slog.Int("total_holidays", getIntFromMap(holidayUpdate, "total_holidays", 0)),
						slog.Int("previous_expected", getIntFromMap(holidayUpdate, "previous_expected", 0)),
						slog.Int("new_expected", getIntFromMap(holidayUpdate, "new_expected", 0)),
					)
				}
			}
		}
	}
}

// Helper functions for parsing map values
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
	}
	return defaultValue
}

// CanRun checks if the stage can run
func (s *ScrapingStage) CanRun() bool {
	return true // Can always run - no dependencies
}

// validateDownloadedFile performs comprehensive validation of a downloaded file
func (s *ScrapingStage) validateDownloadedFile(filePath string) *FileValidationResult {
	result := &FileValidationResult{
		FilePath:       filePath,
		ValidationTime: time.Now(),
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("File does not exist: %v", err)
		if s.logger != nil {
			s.logger.Warn("File validation failed - file not found",
				slog.String("file", filePath),
				slog.String("error", err.Error()))
		}
		return result
	}

	// Check file size
	result.FileSize = info.Size()
	if result.FileSize < minFileSizeBytes {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("File too small: %d bytes (minimum: %d bytes)", result.FileSize, minFileSizeBytes)
		if s.logger != nil {
			s.logger.Warn("File validation failed - file too small",
				slog.String("file", filePath),
				slog.Int64("size", result.FileSize),
				slog.Int("min_size", minFileSizeBytes))
		}
		return result
	}

	if result.FileSize > maxFileSizeBytes {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("File too large: %d bytes (maximum: %d bytes)", result.FileSize, maxFileSizeBytes)
		if s.logger != nil {
			s.logger.Warn("File validation failed - file too large",
				slog.String("file", filePath),
				slog.Int64("size", result.FileSize),
				slog.Int("max_size", maxFileSizeBytes))
		}
		return result
	}

	// Calculate MD5 checksum
	hash, err := s.calculateFileChecksum(filePath)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("Failed to calculate checksum: %v", err)
		if s.logger != nil {
			s.logger.Warn("File validation failed - checksum error",
				slog.String("file", filePath),
				slog.String("error", err.Error()))
		}
		return result
	}

	result.Checksum = hash
	result.IsValid = true

	if s.logger != nil {
		s.logger.Debug("File validation completed successfully",
			slog.String("file", filePath),
			slog.Int64("size", result.FileSize),
			slog.String("checksum", result.Checksum))
	}

	return result
}

// calculateFileChecksum calculates MD5 checksum of a file
func (s *ScrapingStage) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// saveChecksum saves checksum to a .md5 file
func (s *ScrapingStage) saveChecksum(filePath, checksum string) error {
	checksumFile := filePath + checksumExtension
	file, err := os.Create(checksumFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(checksum); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.Debug("Checksum saved",
			slog.String("file", checksumFile),
			slog.String("checksum", checksum))
	}

	return nil
}

// loadChecksum loads checksum from a .md5 file
func (s *ScrapingStage) loadChecksum(filePath string) (string, error) {
	checksumFile := filePath + checksumExtension
	data, err := os.ReadFile(checksumFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// removeCorruptedFile removes a corrupted file and its checksum
func (s *ScrapingStage) removeCorruptedFile(filePath string) error {
	var errors []string

	// Remove main file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("failed to remove file: %v", err))
	}

	// Remove checksum file
	checksumFile := filePath + checksumExtension
	if err := os.Remove(checksumFile); err != nil && !os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("failed to remove checksum file: %v", err))
	}

	if len(errors) > 0 {
		combinedError := strings.Join(errors, "; ")
		if s.logger != nil {
			s.logger.Error("Failed to remove corrupted file",
				slog.String("file", filePath),
				slog.String("errors", combinedError))
		}
		return fmt.Errorf("cleanup failed: %s", combinedError)
	}

	if s.logger != nil {
		s.logger.Info("Corrupted file removed successfully",
			slog.String("file", filePath))
	}

	return nil
}

// validateAndProcessDownloads validates all downloaded files and removes corrupted ones
func (s *ScrapingStage) validateAndProcessDownloads(downloadDir string, files []string) ([]string, []string) {
	var validFiles []string
	var corruptedFiles []string

	if s.logger != nil {
		s.logger.Info("Starting file validation",
			slog.String("directory", downloadDir),
			slog.Int("files_to_validate", len(files)))
	}

	for _, fileName := range files {
		filePath := filepath.Join(downloadDir, fileName)

		// Validate file
		result := s.validateDownloadedFile(filePath)

		if result.IsValid {
			// Save checksum for future verification
			if err := s.saveChecksum(filePath, result.Checksum); err != nil {
				if s.logger != nil {
					s.logger.Warn("Failed to save checksum",
						slog.String("file", filePath),
						slog.String("error", err.Error()))
				}
			}
			validFiles = append(validFiles, fileName)
		} else {
			corruptedFiles = append(corruptedFiles, fileName)

			// Remove corrupted file
			if err := s.removeCorruptedFile(filePath); err != nil {
				if s.logger != nil {
					s.logger.Error("Failed to remove corrupted file",
						slog.String("file", filePath),
						slog.String("error", err.Error()))
				}
			}
		}
	}

	if s.logger != nil {
		s.logger.Info("File validation completed",
			slog.String("directory", downloadDir),
			slog.Int("valid_files", len(validFiles)),
			slog.Int("corrupted_files", len(corruptedFiles)))
	}

	return validFiles, corruptedFiles
}
