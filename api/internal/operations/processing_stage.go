package operations

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

func parseFileInfoModTime(modTime string) (time.Time, bool) {
	if modTime == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339, modTime); err == nil {
		return parsed, true
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", modTime, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

// ProcessingStage runs the processor.exe step and emits simple progress updates
// that mirror the scraping stage's VelocityTrackingBroadcaster pattern.
type ProcessingStage struct {
	BaseStage
	executableDir string
	logger        *slog.Logger
	options       *StageOptions
	broadcaster   *VelocityTrackingBroadcaster
	fileList      []string
}

// NewProcessingStage creates a new processing stage with lightweight progress broadcasting.
func NewProcessingStage(executableDir string, logger *slog.Logger, options *StageOptions) *ProcessingStage {
	if options == nil {
		options = &StageOptions{}
	}

	if logger != nil {
		logger = logger.With(slog.String("stage", StageIDProcessing))
		logger.Info("Processing stage initialized", slog.String("executable_dir", executableDir))
	}

	var broadcaster *VelocityTrackingBroadcaster
	if options.WebSocketManager != nil {
		broadcaster = NewVelocityTrackingBroadcaster("temp", StageIDProcessing, options.WebSocketManager, logger, options.EnableProgress)
	}

	return &ProcessingStage{
		BaseStage:     NewBaseStage(StageIDProcessing, StageNameProcessing, []string{StageIDScraping}),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
		broadcaster:   broadcaster,
	}
}

// Processing Phase Constants
const (
	ProcessingPhasePreparing    = "preparing"
	ProcessingPhaseReading      = "reading"
	ProcessingPhaseTransforming = "transforming"
	ProcessingPhaseWriting      = "writing"
	ProcessingPhaseComplete     = "complete"
)

// setOperationID refreshes the broadcaster with the real operation ID.
func (p *ProcessingStage) setOperationID(operationID string) {
	if p.broadcaster == nil || operationID == "" || operationID == "temp" {
		return
	}
	if p.broadcaster.operationID == operationID {
		return
	}

	currentHub := p.broadcaster.hub
	isEnabled := p.broadcaster.IsEnabled()
	p.broadcaster = NewVelocityTrackingBroadcaster(operationID, StageIDProcessing, currentHub, p.logger, isEnabled)
}

// Execute runs processor.exe and streams simplified progress updates.
func (p *ProcessingStage) Execute(ctx context.Context, state *OperationState) (err error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("processing stage panic: %v", r)
			if p.logger != nil {
				p.logger.Error("Processing stage panic",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())))
			}
			if state != nil {
				if stepState := state.GetStage(p.ID()); stepState != nil {
					p.updateSimpleProgress(stepState, 0, panicErr.Error(), "failed", ProcessingPhaseComplete, 0, 0, 0, "", 0, 0, 0, nil)
				}
			}
			if p.broadcaster != nil {
				p.broadcaster.Fail(panicErr)
			}
			err = panicErr
		}
	}()

	p.setOperationID(state.ID)

	// Ensure broadcaster exists and is enabled so progress reaches the UI
	if p.broadcaster == nil && p.options != nil && p.options.WebSocketManager != nil {
		p.broadcaster = NewVelocityTrackingBroadcaster(state.ID, StageIDProcessing, p.options.WebSocketManager, p.logger, true)
	} else if p.broadcaster != nil {
		// Update existing broadcaster's operation ID if needed, instead of recreating it
		// This prevents race conditions where a new broadcaster might be created while the old one is still active
		if p.broadcaster.operationID != state.ID && state.ID != "" && state.ID != "temp" {
			p.broadcaster.operationID = state.ID
		}
		if !p.broadcaster.IsEnabled() {
			p.broadcaster.SetEnabled(true)
		}
	}
	if p.broadcaster == nil {
		return fmt.Errorf("processing broadcaster not configured (WebSocket hub missing)")
	}

	stepState := state.GetStage(p.ID())
	if stepState == nil {
		return fmt.Errorf("processing stage state not initialized")
	}
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	stepState.Metadata["operation_type"] = "processing"
	stepState.Metadata["stage_id"] = StageIDProcessing
	stepState.Metadata["target_stage_id"] = StageIDProcessing

	downloadsDir := p.absolutePath(relativeDownloadsRoot())
	reportsDir := p.absolutePath(relativeReportsRoot())

	// Phase: Preparing
	p.updateSimpleProgress(stepState, 0, "Initializing data processor...", "running", ProcessingPhasePreparing, 0, 0, 0, "", 0, 0, 0, nil)

	analysis := p.detectUnprocessedFiles(downloadsDir, reportsDir)
	totalFiles := analysis.FilesToProcessCount
	if totalFiles < 0 {
		totalFiles = 0
	}

	// Build file list from analysis results for consistent segmented progress
	fileList := make([]string, len(analysis.FilesToProcess))
	for i, f := range analysis.FilesToProcess {
		fileList[i] = relativeDownloadFilePath(f.Name)
	}
	p.fileList = fileList
	stepState.Metadata["file_list"] = fileList
	p.applyAnalysisMetadata(stepState, analysis)

	// Initialize file statuses for UI granular progress tracking
	if len(fileList) > 0 {
		statuses := make([]*FileProcessingStatus, len(fileList))
		for i, fileName := range fileList {
			statuses[i] = &FileProcessingStatus{
				FileName: filepath.Base(fileName),
				Status:   "pending",
				Progress: 0,
			}
		}
		stepState.Metadata["file_statuses"] = statuses
	}

	// Log analysis summary
	if p.logger != nil {
		p.logger.Info("Processing file analysis result",
			slog.Int("total_excel_files", analysis.TotalExcelFiles),
			slog.Int("total_csv_files", analysis.TotalCSVFiles),
			slog.Int("files_to_process", analysis.FilesToProcessCount),
			slog.Bool("has_work", analysis.HasWork),
			slog.String("reason", analysis.Reason))
	}

	// CRITICAL: Only send "No new files" completion if we are absolutely sure.
	// If HasWork is true, we proceed.
	if !analysis.HasWork {
		// Use Skip instead of Complete to prevent UI from showing a full progress bar for an empty run
		stepState.Status = StepStatusSkipped
		stepState.Message = "No new files to process"
		if p.broadcaster != nil {
			p.broadcaster.Skip("No new files to process")
		}
		return nil
	}

	// Phase: Reading - Found files
	p.updateSimpleProgress(stepState, 5, "Loading Excel files...", "running", ProcessingPhaseReading, 0, totalFiles, 0, "", 0, 0, 0, nil)

	lastActivity := atomic.Int64{}
	lastActivity.Store(time.Now().UnixNano())
	markActivity := func() {
		lastActivity.Store(time.Now().UnixNano())
	}
	markActivity()

	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("create reports directory: %w", err)
	}

	processorPath := filepath.Join(p.executableDir, "processor.exe")
	if _, err := os.Stat(processorPath); err != nil {
		return fmt.Errorf("processor.exe not found: %w", err)
	}

	args := p.buildProcessorArgs(state)

	if p.logger != nil {
		p.logger.Info("Starting processor",
			slog.String("path", processorPath),
			slog.Any("args", args),
			slog.Int("total_files", totalFiles))
	}

	p.updateSimpleProgress(
		stepState,
		5,
		fmt.Sprintf("Found %d files to process", totalFiles),
		"running",
		ProcessingPhaseReading,
		0,
		totalFiles,
		0,
		"",
		0,
		0,
		0,
		nil,
	)
	markActivity()

	cmdCtx, cmdCancel := context.WithCancel(ctx)
	defer cmdCancel()

	cmd := exec.CommandContext(cmdCtx, processorPath, args...)
	cmd.Dir = p.executableDir

	var processedAtomic atomic.Int64
	var failedAtomic atomic.Int64
	var generatedAtomic atomic.Int64
	var totalOutputsAtomic atomic.Int64
	var generationProgressAtomic atomic.Int64
	var currentFileAtomic atomic.Value
	currentFileAtomic.Store("")
	processedTotal := totalFiles
	if processedTotal == 0 && len(fileList) > 0 {
		processedTotal = len(fileList)
	}
	if processedTotal == 0 {
		// Avoid division by zero in heartbeats
		processedTotal = 1
	}
	failOnce := atomic.Bool{}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	// Start the processor
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		msg := fmt.Sprintf("start processor: %v", err)
		p.updateSimpleProgress(stepState, 10, msg, "failed", ProcessingPhaseTransforming, 0, processedTotal, 0, "", int(generatedAtomic.Load()), int(totalOutputsAtomic.Load()), int(generationProgressAtomic.Load()), nil)
		return fmt.Errorf(msg)
	}
	if p.logger != nil {
		p.logger.Info("Processor started",
			slog.String("path", processorPath),
			slog.Any("args", args),
			slog.Int("total_files", totalFiles))
	}
	markActivity()

	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()

	// Emit an early "started" heartbeat to avoid UI jumping straight to 100%
	p.updateSimpleProgress(stepState, 10, "Processor started", "running", ProcessingPhaseTransforming, 0, processedTotal, 0, "", int(generatedAtomic.Load()), int(totalOutputsAtomic.Load()), int(generationProgressAtomic.Load()), nil)

	// Heartbeat monitor to keep UI updated even when stdout is quiet
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-monitorCtx.Done():
				return
			case <-ticker.C:
				fp := int(processedAtomic.Load())
				ff := int(failedAtomic.Load())
				tf := processedTotal
				gen := int(generatedAtomic.Load())
				genTotal := int(totalOutputsAtomic.Load())
				genProg := int(generationProgressAtomic.Load())
				cur := ""
				if v, ok := currentFileAtomic.Load().(string); ok {
					cur = v
				}
				msg := p.processingMessage(fp, tf, cur)
				progress := 0
				if tf > 0 {
					ratio := float64(fp+ff) / float64(tf)
					if ratio > 1 {
						ratio = 1
					}
					progress = clampProgress(int(math.Round(ratio * 100)))
				}

				// Determine phase for heartbeat
				currentPhase := ProcessingPhaseTransforming
				if genTotal > 0 && gen > 0 {
					currentPhase = ProcessingPhaseWriting
				}

				p.updateSimpleProgress(stepState, progress, msg, "running", currentPhase, fp, tf, ff, cur, gen, genTotal, genProg, nil)
			}
		}
	}()

	// Watchdog to detect silent stalls and fail fast
	stallThreshold := 60 * time.Second
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-monitorCtx.Done():
				return
			case <-ticker.C:
				last := time.Unix(0, lastActivity.Load())
				if time.Since(last) > stallThreshold {
					if !failOnce.CompareAndSwap(false, true) {
						return
					}
					fp := int(processedAtomic.Load())
					ff := int(failedAtomic.Load())
					gen := int(generatedAtomic.Load())
					genTotal := int(totalOutputsAtomic.Load())
					genProg := int(generationProgressAtomic.Load())
					progress := 0
					if processedTotal > 0 {
						ratio := float64(fp+ff) / float64(processedTotal)
						if ratio > 1 {
							ratio = 1
						}
						progress = clampProgress(int(math.Round(ratio * 100)))
					}
					cur := ""
					if v, ok := currentFileAtomic.Load().(string); ok {
						cur = v
					}
					msg := fmt.Sprintf("processor stalled: no progress for %s", stallThreshold)
					p.updateSimpleProgress(stepState, progress, msg, "failed", ProcessingPhaseTransforming, fp, processedTotal, ff, cur, gen, genTotal, genProg, nil)
					if p.logger != nil {
						p.logger.Error("processing watchdog triggered", slog.String("reason", msg))
					}
					cmdCancel()
					monitorCancel()
					return
				}
			}
		}
	}()

	// Collect stderr for diagnostics without blocking progress updates.
	var stderrBuf strings.Builder
	errChan := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			stderrBuf.WriteString(scanner.Text())
			stderrBuf.WriteByte('\n')
		}
		errChan <- scanner.Err()
	}()

	parser := NewProcessingParser(totalFiles)
	stdoutScanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	stdoutScanner.Buffer(buf, 1024*1024)

	stageStart := time.Now()
	if stepState.StartTime != nil {
		stageStart = *stepState.StartTime
	}

	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()

		if p.logger != nil {
			p.logger.Debug("processor output", slog.String("line", line))
		}

		// Debug: Log any generation-related output
		if strings.Contains(line, "Total outputs") || strings.Contains(line, "Generated output") {
			if p.logger != nil {
				p.logger.Info("Found generation output", slog.String("line", line))
			}
		}

		if snap, changed := parser.ParseLine(line); changed {
			// Keep atomic counters aligned with parser snapshot (monotonic across multiple sources)
			if next := int64(snap.FilesProcessed); next > processedAtomic.Load() {
				processedAtomic.Store(next)
			}
			if next := int64(snap.FailedFiles); next > failedAtomic.Load() {
				failedAtomic.Store(next)
			}
			currentFileAtomic.Store(snap.CurrentFile)
			if next := int64(snap.OutputsGenerated); next > generatedAtomic.Load() {
				generatedAtomic.Store(next)
			}
			if next := int64(snap.TotalOutputs); next > totalOutputsAtomic.Load() {
				totalOutputsAtomic.Store(next)
			}
			if next := int64(snap.GenerationProgress); next > generationProgressAtomic.Load() {
				generationProgressAtomic.Store(next)
			}
			message := snap.Message
			if message == "" {
				message = p.processingMessage(snap.FilesProcessed, snap.TotalFiles, snap.CurrentFile)
			}

			// Update file statuses based on current progress
			if len(p.fileList) > 0 {
				statuses := make([]*FileProcessingStatus, len(p.fileList))
				// Copy existing statuses or create new ones
				if existingStatuses, ok := stepState.Metadata["file_statuses"].([]*FileProcessingStatus); ok && len(existingStatuses) == len(p.fileList) {
					copy(statuses, existingStatuses)
				} else {
					for i, fileName := range p.fileList {
						statuses[i] = &FileProcessingStatus{
							FileName: filepath.Base(fileName),
							Status:   "pending",
							Progress: 0,
						}
					}
				}

				// Update statuses based on progress
				for i := 0; i < snap.FilesProcessed && i < len(statuses); i++ {
					statuses[i].Status = "completed"
					statuses[i].Progress = 100
				}

				// Mark current file as processing
				if snap.CurrentFile != "" && snap.FilesProcessed < len(statuses) {
					for i, fileName := range p.fileList {
						if filepath.Base(fileName) == snap.CurrentFile {
							statuses[i].Status = "processing"
							statuses[i].Progress = 50 // Mid-progress
							break
						}
					}
				}

				// Update metadata with new statuses
				stepState.Metadata["file_statuses"] = statuses
			}

			markActivity()
			markActivity()

			// Determine phase based on what's happening
			currentPhase := ProcessingPhaseTransforming
			if snap.TotalOutputs > 0 || snap.OutputsGenerated > 0 {
				currentPhase = ProcessingPhaseWriting
			}

			p.updateSimpleProgress(
				stepState,
				snap.Progress,
				message,
				"running",
				currentPhase,
				int(processedAtomic.Load()), // Use atomic high-water mark
				processedTotal,              // Use local variable
				int(failedAtomic.Load()),    // Use atomic high-water mark
				snap.CurrentFile,
				int(generatedAtomic.Load()),          // Use atomic high-water mark
				int(totalOutputsAtomic.Load()),       // Use atomic high-water mark
				int(generationProgressAtomic.Load()), // Use atomic high-water mark
				snap.OutputStatuses,
			)
		}
	}

	// Fallback ticker: emit progress based on files written to disk when stdout is sparse.
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-monitorCtx.Done():
				return
			case <-ticker.C:
				// Count processed files by looking for corresponding CSVs in daily directory.
				fileDetector := NewFileDetector(p.logger)
				dailyDir := filepath.Join(reportsDir, "daily")
				csvFiles, _ := fileDetector.GetCSVFileDetails(dailyDir)

				processedFromDisk := 0
				expected := make(map[string]struct{}, len(p.fileList))
				for _, f := range p.fileList {
					base := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
					expected[base] = struct{}{}
				}
				for _, csv := range csvFiles {
					base := strings.TrimSuffix(csv.Name, filepath.Ext(csv.Name))
					// Ignore files created before this stage started to avoid counting previous runs.
					if csv.ModTime != "" {
						if mt, ok := parseFileInfoModTime(csv.ModTime); ok && mt.Before(stageStart) {
							continue
						}
					}
					// Ignore empty or very small files (likely placeholders created at start)
					if csv.Size < 100 {
						continue
					}
					if _, ok := expected[base]; ok {
						processedFromDisk++
					}
				}

				// Outputs generated fallback (all CSVs we see for this run).
				outputsFromDisk := 0
				for _, csv := range csvFiles {
					if csv.ModTime != "" {
						if mt, ok := parseFileInfoModTime(csv.ModTime); ok && mt.Before(stageStart) {
							continue
						}
					}
					if csv.Size < 100 {
						continue
					}
					outputsFromDisk++
				}

				currentProcessed := int(processedAtomic.Load())
				currentFailed := int(failedAtomic.Load())
				currentOutputs := int(generatedAtomic.Load())
				currentTotalOutputs := int(totalOutputsAtomic.Load())

				// Only emit if we have new progress.
				if processedFromDisk <= currentProcessed && outputsFromDisk <= currentOutputs {
					continue
				}

				// Keep counters monotonic.
				if processedFromDisk < currentProcessed {
					processedFromDisk = currentProcessed
				}
				if outputsFromDisk < currentOutputs {
					outputsFromDisk = currentOutputs
				}
				if outputsFromDisk > currentTotalOutputs {
					totalOutputsAtomic.Store(int64(outputsFromDisk))
					currentTotalOutputs = outputsFromDisk
				}
				generatedAtomic.Store(int64(outputsFromDisk))

				if processedFromDisk > processedTotal {
					processedFromDisk = processedTotal
				}
				processedAtomic.Store(int64(processedFromDisk))

				progress := clampProgress(int(math.Round((float64(processedFromDisk+currentFailed) / float64(maxInt(processedTotal, 1))) * 100)))
				genProg := 0
				if currentTotalOutputs > 0 {
					genProg = clampProgress(int(math.Round((float64(outputsFromDisk) / float64(currentTotalOutputs)) * 100)))
				}
				if next := int64(genProg); next > generationProgressAtomic.Load() {
					generationProgressAtomic.Store(next)
				}

				curFile := ""
				if v, ok := currentFileAtomic.Load().(string); ok {
					curFile = v
				}
				message := p.processingMessage(processedFromDisk, processedTotal, curFile)

				// Update file statuses in fallback ticker
				if len(p.fileList) > 0 && processedFromDisk > currentProcessed {
					statuses := make([]*FileProcessingStatus, len(p.fileList))
					// Copy existing statuses
					if existingStatuses, ok := stepState.Metadata["file_statuses"].([]*FileProcessingStatus); ok && len(existingStatuses) == len(p.fileList) {
						copy(statuses, existingStatuses)
					} else {
						for i, fileName := range p.fileList {
							statuses[i] = &FileProcessingStatus{
								FileName: filepath.Base(fileName),
								Status:   "pending",
								Progress: 0,
							}
						}
					}

					// Update statuses based on disk-detected progress
					for i := 0; i < processedFromDisk && i < len(statuses); i++ {
						statuses[i].Status = "completed"
						statuses[i].Progress = 100
					}

					stepState.Metadata["file_statuses"] = statuses
				}

				// Determine phase for fallback
				currentPhase := ProcessingPhaseTransforming
				if outputsFromDisk > 0 {
					currentPhase = ProcessingPhaseWriting
				}

				p.updateSimpleProgress(
					stepState,
					progress,
					message,
					"running",
					currentPhase,
					processedFromDisk,
					processedTotal,
					currentFailed,
					curFile,
					outputsFromDisk,
					currentTotalOutputs,
					genProg,
					nil,
				)
			}
		}
	}()

	if scanErr := stdoutScanner.Err(); scanErr != nil && p.logger != nil {
		p.logger.Warn("processor stdout scan error", slog.String("error", scanErr.Error()))
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			p.logger.Error("Processor exited with error",
				slog.Int("exit_code", exitErr.ExitCode()),
				slog.String("stderr", string(exitErr.Stderr)))
		} else {
			p.logger.Error("Processor execution failed", slog.Any("error", waitErr))
		}
	}
	stderrErr := <-errChan
	latest := parser.Snapshot()

	if waitErr != nil || stderrErr != nil {
		var failureErr error
		switch {
		case waitErr != nil && stderrErr != nil:
			failureErr = fmt.Errorf("processing failed: wait error: %w; stderr scan: %v", waitErr, stderrErr)
		case waitErr != nil:
			failureErr = fmt.Errorf("processing failed: %w", waitErr)
		default:
			failureErr = fmt.Errorf("processing failed: stderr scan: %w", stderrErr)
		}
		stderrSnippet := strings.TrimSpace(stderrBuf.String())
		if serr := strings.TrimSpace(stderrBuf.String()); serr != "" {
			stepState.Metadata["stderr_output"] = serr
		}

		if !failOnce.CompareAndSwap(false, true) {
			return failureErr
		}
		p.updateSimpleProgress(
			stepState,
			latest.Progress,
			failureErr.Error(),
			"failed",
			ProcessingPhaseTransforming, // Stay on transforming if failed
			latest.FilesProcessed,
			latest.TotalFiles,
			latest.FailedFiles,
			latest.CurrentFile,
			latest.OutputsGenerated,
			latest.TotalOutputs,
			latest.GenerationProgress,
			latest.OutputStatuses,
		)
		if p.logger != nil {
			p.logger.Error("Processor failed",
				slog.String("message", failureErr.Error()),
				slog.String("stderr_snippet", truncate(stderrSnippet, 2000)))
		}
		if p.broadcaster != nil {
			p.broadcaster.Fail(failureErr)
		}
		return failureErr
	}

	// Ensure final metadata is complete for UI
	finalFilesProcessed := latest.FilesProcessed
	finalTotalFiles := latest.TotalFiles
	if finalTotalFiles < processedTotal {
		finalTotalFiles = processedTotal
	}
	if finalFilesProcessed > finalTotalFiles {
		finalFilesProcessed = finalTotalFiles
	}
	finalOutputsGenerated := maxInt(latest.OutputsGenerated, int(generatedAtomic.Load()))
	finalTotalOutputs := maxInt(latest.TotalOutputs, int(totalOutputsAtomic.Load()))

	// Fallback: if parser saw nothing, derive counts from file list length
	if finalFilesProcessed == 0 && len(fileList) > 0 {
		finalFilesProcessed = len(fileList)
		finalTotalFiles = len(fileList)
	}

	message := p.processingMessage(finalFilesProcessed, finalTotalFiles, "")
	if latest.FailedFiles > 0 {
		message = fmt.Sprintf("Processed %d/%d files (%d failed)", finalFilesProcessed, finalTotalFiles, latest.FailedFiles)
	} else {
		message = fmt.Sprintf("Processed %d/%d files - CSVs ready", finalFilesProcessed, finalTotalFiles)
	}
	p.updateSimpleProgress(
		stepState,
		100,
		message,
		"completed",
		ProcessingPhaseComplete,
		finalFilesProcessed,
		finalTotalFiles,
		latest.FailedFiles,
		"",
		finalOutputsGenerated,
		finalTotalOutputs,
		clampProgress(int(math.Round((float64(finalOutputsGenerated)/float64(maxInt(finalTotalOutputs, 1)))*100))),
		latest.OutputStatuses,
	)

	if p.broadcaster != nil {
		p.broadcaster.CompleteWithMetadata(message, p.metadataSnapshot(stepState))
	}

	if p.logger != nil {
		p.logger.Info("Processing completed",
			slog.Int("files_processed", maxInt(latest.FilesProcessed, totalFiles)),
			slog.Int("failed_files", latest.FailedFiles),
			slog.Int("total_files", maxInt(latest.TotalFiles, totalFiles)))
	}

	return nil
}

func (p *ProcessingStage) processingMessage(processed, total int, currentFile string) string {
	if processed > 0 && total > 0 {
		return fmt.Sprintf("Processing files... (%d/%d files)", processed, total)
	}
	if currentFile != "" {
		return fmt.Sprintf("Processing %s", filepath.Base(currentFile))
	}
	return "Processing files..."
}

func (p *ProcessingStage) updateSimpleProgress(
	stepState *StepState,
	progress int,
	message,
	status string,
	phase string, // New argument: explicit phase
	filesProcessed,
	totalFiles,
	failedFiles int,
	currentFile string,
	outputsGenerated int,
	totalOutputs int,
	generationProgress int,
	outputStatuses []*FileProcessingStatus,
) {
	if stepState == nil {
		return
	}

	// Normalize counts to avoid exceeding totalFiles in UI
	boundedTotal := totalFiles
	if boundedTotal < 0 {
		boundedTotal = 0
	}
	boundedProcessed := filesProcessed
	boundedFailed := failedFiles
	if boundedTotal > 0 {
		if boundedProcessed > boundedTotal {
			boundedProcessed = boundedTotal
		}
		if boundedFailed > boundedTotal {
			boundedFailed = boundedTotal
		}
		// Ensure combined processed+failed does not exceed total
		if boundedProcessed+boundedFailed > boundedTotal {
			over := (boundedProcessed + boundedFailed) - boundedTotal
			if boundedFailed >= over {
				boundedFailed -= over
			} else {
				boundedProcessed -= over - boundedFailed
				boundedFailed = 0
			}
		}
	}

	remaining := 0
	if boundedTotal > 0 {
		remaining = boundedTotal - (boundedProcessed + boundedFailed)
		if remaining < 0 {
			remaining = 0
		}
	}

	// Recompute progress based on bounded counts to keep bar accurate
	fileProgress := progress
	if boundedTotal > 0 {
		fileProgress = clampProgress(int(math.Round((float64(boundedProcessed+boundedFailed) / float64(boundedTotal)) * 100)))
	} else {
		fileProgress = clampProgress(progress)
	}

	// Blend file progress with generation progress so we don't report 100% until outputs are generated.
	genProgress := generationProgress
	if genProgress == 0 && totalOutputs > 0 && outputsGenerated > 0 {
		genProgress = clampProgress(int(math.Round((float64(outputsGenerated) / float64(totalOutputs)) * 100)))
	}
	// Weight parsing and generation to avoid premature 100%: 60% files, 40% generation.
	clamped := fileProgress
	if totalOutputs > 0 {
		clamped = clampProgress(int(math.Round((float64(fileProgress)*0.6 + float64(genProgress)*0.4))))
	}

	if status == "failed" {
		phase = string(StagePhaseFailed)
	}

	// Normalize the message so UI never shows over-total counts.
	if status == "running" {
		message = p.processingMessage(boundedProcessed, boundedTotal, currentFile)
	}

	stepState.mu.Lock()
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	stepState.Metadata["stage_id"] = StageIDProcessing
	stepState.Metadata["target_stage_id"] = StageIDProcessing
	stepState.Metadata["operation_type"] = "processing"
	stepState.Metadata["files_processed"] = boundedProcessed
	stepState.Metadata["total_files"] = boundedTotal
	stepState.Metadata["remaining_files"] = remaining
	stepState.Metadata["failed_files"] = boundedFailed
	stepState.Metadata["current_file"] = currentFile
	stepState.Metadata["progress_percent"] = clamped
	stepState.Metadata["files_generated"] = outputsGenerated
	stepState.Metadata["total_outputs"] = totalOutputs
	stepState.Metadata["generation_progress_percent"] = generationProgress
	if len(outputStatuses) > 0 {
		stepState.Metadata["generation_file_statuses"] = outputStatuses
	}
	stepState.Metadata["status"] = status
	stepState.Metadata["stage_message"] = message
	stepState.Metadata["phase_message"] = message
	stepState.Metadata["phase"] = string(phase)
	stepState.Metadata["current_phase"] = string(phase)
	stepState.Metadata["telemetry_missing"] = false
	if stepState.StartTime != nil {
		stepState.Metadata["stage_started_at"] = stepState.StartTime.Format(time.RFC3339)
	}
	stepState.Metadata["stage_updated_at"] = time.Now().Format(time.RFC3339)

	// Build lightweight file status segments for the UI segmented bar
	if boundedTotal > 0 {
		statuses := make([]*FileProcessingStatus, 0, boundedTotal)
		// Completed
		for i := 0; i < boundedProcessed && i < boundedTotal; i++ {
			name := fmt.Sprintf("file_%d", i+1)
			if i < len(p.fileList) {
				name = p.fileList[i]
			}
			statuses = append(statuses, &FileProcessingStatus{
				FileName: name,
				Status:   "completed",
				Progress: 100,
			})
		}
		// Failed
		for i := 0; i < boundedFailed && len(statuses) < boundedTotal; i++ {
			name := fmt.Sprintf("failed_%d", i+1)
			idx := boundedProcessed + i
			if idx < len(p.fileList) {
				name = p.fileList[idx]
			}
			statuses = append(statuses, &FileProcessingStatus{
				FileName: name,
				Status:   "failed",
				Progress: 0,
			})
		}
		// Current processing
		if currentFile != "" && status == "running" && len(statuses) < totalFiles {
			statuses = append(statuses, &FileProcessingStatus{
				FileName: currentFile,
				Status:   "processing",
				Progress: 0,
			})
		}
		// Pending
		for len(statuses) < totalFiles {
			name := fmt.Sprintf("pending_%d", len(statuses)+1)
			if len(statuses) < len(p.fileList) {
				name = p.fileList[len(statuses)]
			}
			statuses = append(statuses, &FileProcessingStatus{
				FileName: name,
				Status:   "pending",
				Progress: 0,
			})
		}
		stepState.Metadata["file_statuses"] = statuses
	}
	metadataCopy := cloneMetadata(stepState.Metadata)
	if status == "completed" {
		now := time.Now()
		stepState.EndTime = &now
		stepState.Status = StepStatusCompleted
	} else if status == "failed" {
		now := time.Now()
		stepState.EndTime = &now
		stepState.Status = StepStatusFailed
	}
	stepState.mu.Unlock()

	stepState.UpdateProgress(float64(clamped), message)

	if p.broadcaster != nil {
		p.broadcaster.UpdateProgressWithMetadata(clamped, message, metadataCopy)
	}
}

func (p *ProcessingStage) metadataSnapshot(stepState *StepState) map[string]interface{} {
	if stepState == nil {
		return nil
	}
	stepState.mu.RLock()
	defer stepState.mu.RUnlock()

	if len(stepState.Metadata) == 0 {
		return nil
	}
	return cloneMetadata(stepState.Metadata)
}

func (p *ProcessingStage) buildProcessorArgs(state *OperationState) []string {
	args := []string{}
	args = append(args, "--in", relativeDownloadsRoot())
	// Use absolute reports path to avoid double-prefixing by the CSV writer.
	args = append(args, "--out", p.absolutePath(relativeReportsRoot()))
	return args
}

func (p *ProcessingStage) absolutePath(rel string) string {
	return AbsolutePath(p.executableDir, rel)
}

// detectUnprocessedFiles analyzes downloaded Excel files vs existing CSV outputs.
func (p *ProcessingStage) detectUnprocessedFiles(downloadsDir, reportsDir string) ProcessingAnalysis {
	analysis := ProcessingAnalysis{
		FilesToProcess:   make([]FileInfo, 0),
		AlreadyProcessed: make([]FileInfo, 0),
		Reason:           "",
		HasWork:          false,
	}

	if p.logger != nil {
		p.logger.Info("Analyzing files for processing",
			slog.String("downloads_dir", downloadsDir),
			slog.String("reports_dir", reportsDir))
	}

	fileDetector := NewFileDetector(p.logger)

	excelFiles, err := fileDetector.GetExcelFileDetails(downloadsDir)
	if err != nil {
		analysis.Reason = fmt.Sprintf("Failed to detect Excel files: %s", err.Error())
		return analysis
	}

	dailyDir := filepath.Join(reportsDir, "daily")

	// Debug logging: Check what's in the daily directory
	if entries, err := os.ReadDir(dailyDir); err == nil {
		p.logger.Info("Daily directory scan",
			slog.String("path", dailyDir),
			slog.Int("total_entries", len(entries)))
	} else {
		p.logger.Warn("Daily directory not accessible",
			slog.String("path", dailyDir),
			slog.String("error", err.Error()))
	}

	analysis.TotalExcelFiles = len(excelFiles)

	// Use filepath.Walk to find existing daily CSVs - exactly like processor.exe does
	existingDates := make(map[string]bool)
	walkError := filepath.Walk(dailyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if p.logger != nil {
				p.logger.Warn("filepath.Walk error",
					slog.String("path", path),
					slog.String("error", err.Error()))
			}
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

	if walkError != nil && p.logger != nil {
		p.logger.Error("filepath.Walk failed",
			slog.String("daily_dir", dailyDir),
			slog.String("error", walkError.Error()))
	}

	analysis.TotalCSVFiles = len(existingDates)

	if p.logger != nil {
		p.logger.Info("CSV date detection (using filepath.Walk)",
			slog.String("daily_dir", dailyDir),
			slog.Int("existing_dates_found", len(existingDates)),
			slog.Bool("walk_succeeded", walkError == nil))
	}

	// Extract date from Excel filename using regex: "YYYY MM DD ..." -> "YYYY_MM_DD"
	excelDateRegex := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2})`)

	for _, excelFile := range excelFiles {
		matches := excelDateRegex.FindStringSubmatch(excelFile.Name)
		if len(matches) != 4 {
			// Malformed filename, treat as needing processing
			if p.logger != nil {
				p.logger.Debug("Excel file has non-standard name format",
					slog.String("filename", excelFile.Name))
			}
			analysis.FilesToProcess = append(analysis.FilesToProcess, excelFile)
			continue
		}

		// Format date as YYYY_MM_DD to match CSV naming convention
		dateStr := fmt.Sprintf("%s_%s_%s", matches[1], matches[2], matches[3])

		if existingDates[dateStr] {
			// CSV exists for this date - already processed
			analysis.AlreadyProcessed = append(analysis.AlreadyProcessed, excelFile)
		} else {
			// No CSV for this date - needs processing
			analysis.FilesToProcess = append(analysis.FilesToProcess, excelFile)
		}
	}

	analysis.FilesToProcessCount = len(analysis.FilesToProcess)
	analysis.HasWork = analysis.FilesToProcessCount > 0

	if !analysis.HasWork {
		if analysis.TotalExcelFiles == 0 {
			analysis.Reason = "No Excel files found in downloads directory"
		} else {
			analysis.Reason = fmt.Sprintf("All %d Excel files already processed", analysis.TotalExcelFiles)
		}
	} else {
		analysis.Reason = fmt.Sprintf("Found %d of %d Excel files needing processing", analysis.FilesToProcessCount, analysis.TotalExcelFiles)
	}

	if p.logger != nil {
		p.logger.Info("File Detection Analysis",
			slog.Int("total_excel_files", analysis.TotalExcelFiles),
			slog.Int("total_csv_files", analysis.TotalCSVFiles),
			slog.Int("files_to_process", analysis.FilesToProcessCount),
			slog.Int("already_processed", len(analysis.AlreadyProcessed)),
			slog.String("reason", analysis.Reason))

		if len(analysis.AlreadyProcessed) > 0 {
			sample := analysis.AlreadyProcessed
			if len(sample) > 5 {
				sample = sample[:5]
			}
			names := make([]string, len(sample))
			for i, f := range sample {
				names[i] = f.Name
			}
			p.logger.Debug("Already processed files sample", slog.Any("files", names))
		}

		if len(analysis.FilesToProcess) > 0 {
			sample := analysis.FilesToProcess
			if len(sample) > 5 {
				sample = sample[:5]
			}
			names := make([]string, len(sample))
			for i, f := range sample {
				names[i] = f.Name
			}
			p.logger.Debug("Files to process sample", slog.Any("files", names))
		}
	}

	return analysis
}

// ProcessingAnalysis represents the result of file analysis.
type ProcessingAnalysis struct {
	FilesToProcess      []FileInfo // Files that need to be processed
	AlreadyProcessed    []FileInfo // Files already processed as CSV
	TotalExcelFiles     int        // Total Excel files found
	TotalCSVFiles       int        // Total CSV files found
	FilesToProcessCount int        // Count of files needing processing
	Reason              string     // Reason for any exclusions
	HasWork             bool       // Whether there are files to process
}

// ID returns the stage identifier.
func (p *ProcessingStage) ID() string {
	return StageIDProcessing
}

// Name returns the stage name.
func (p *ProcessingStage) Name() string {
	return StageNameProcessing
}

// Dependencies returns the stage dependencies.
func (p *ProcessingStage) Dependencies() []string {
	return []string{StageIDScraping}
}

// CanRun checks if the stage can run.
func (p *ProcessingStage) CanRun() bool {
	return true
}

// deriveProcessingPhase determines the current phase based on progress percentage.
func deriveProcessingPhase(progress int) StagePhase {
	switch {
	case progress < 20:
		return StagePhaseStarting
	case progress < 85:
		return StagePhaseRunning
	case progress < 100:
		return StagePhaseCompleting
	default:
		return StagePhaseCompleted
	}
}

func (p *ProcessingStage) applyAnalysisMetadata(stepState *StepState, analysis ProcessingAnalysis) {
	stepState.mu.Lock()
	defer stepState.mu.Unlock()

	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	stepState.Metadata["total_excel_files"] = analysis.TotalExcelFiles
	stepState.Metadata["total_csv_files"] = analysis.TotalCSVFiles
	stepState.Metadata["files_to_process"] = analysis.FilesToProcessCount
	stepState.Metadata["already_processed"] = len(analysis.AlreadyProcessed)
	stepState.Metadata["has_work"] = analysis.HasWork
	stepState.Metadata["analysis_reason"] = analysis.Reason
}
