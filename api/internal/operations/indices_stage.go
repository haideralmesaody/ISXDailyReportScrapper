package operations

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// IndicesStage handles index extraction
type IndicesStage struct {
	BaseStage
	executableDir string
	logger        *slog.Logger
	options       *StageOptions
	broadcaster   *VelocityTrackingBroadcaster
	fileList      []string

	// Add atomic counters for thread safety
	filesProcessed   atomic.Int64
	totalFiles       atomic.Int64
	indicesExtracted atomic.Int64 // Count of ISX60 values extracted
	isx15Extracted   atomic.Int64 // Count of ISX15 values extracted
}

// NewIndicesStage creates a new indices extraction Step
func NewIndicesStage(executableDir string, logger *slog.Logger, options *StageOptions) *IndicesStage {
	if options == nil {
		options = &StageOptions{}
	}

	// Create logger with Step context
	if logger != nil {
		logger = logger.With(slog.String("Step", StageIDIndices))
		logger.Info("Indices Step initialized",
			slog.String("executable_dir", executableDir))
	}

	// Create velocity tracking broadcaster
	var broadcaster *VelocityTrackingBroadcaster
	if options.WebSocketManager != nil {
		broadcaster = NewVelocityTrackingBroadcaster("temp", StageIDIndices, options.WebSocketManager, logger, options.EnableProgress)
	}

	return &IndicesStage{
		BaseStage:     NewBaseStage(StageIDIndices, StageNameIndices, []string{StageIDScraping}), // Depends on scraping
		executableDir: executableDir,
		logger:        logger,
		options:       options,
		broadcaster:   broadcaster,
	}
}

// setOperationID updates the broadcaster with the actual operation ID
func (i *IndicesStage) setOperationID(operationID string) {
	if i.broadcaster != nil && operationID != "" && operationID != "temp" {
		if i.broadcaster.operationID != operationID {
			currentHub := i.broadcaster.hub
			isEnabled := i.broadcaster.IsEnabled()
			i.broadcaster = NewVelocityTrackingBroadcaster(operationID, StageIDIndices, currentHub, i.logger, isEnabled)
		}
	}
}

// Execute runs the index extractor
func (i *IndicesStage) Execute(ctx context.Context, state *OperationState) error {
	// Update broadcaster with actual operation ID
	i.setOperationID(state.ID)

	stepState := state.GetStage(i.ID())
	if stepState == nil {
		return fmt.Errorf("stage state not found for stage %s", i.ID())
	}
	if stepState.StartTime == nil {
		now := time.Now()
		stepState.StartTime = &now
	}

	// Initialize metadata
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	stepState.Metadata["operation_type"] = "indices_extraction"
	stepState.Metadata["stage_id"] = StageIDIndices

	// Detect Excel files before starting
	downloadsDir := AbsoluteDownloadsPath(i.executableDir, "")
	indexOutPath := AbsolutePath(i.executableDir, filepath.Join("data", "reports", "indexes", "indexes.csv"))
	lastProcessedDate, hasLastDate := loadLastIndexDate(indexOutPath)
	files, err := filepath.Glob(filepath.Join(downloadsDir, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("failed to scan downloads directory: %w", err)
	}

	// Filter for valid ISX report files
	var validFiles []string
	fileRe := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	for _, file := range files {
		base := filepath.Base(file)
		if !fileRe.MatchString(base) {
			continue
		}
		if hasLastDate {
			if matches := fileRe.FindStringSubmatch(base); len(matches) == 4 {
				if fileDate, err := time.Parse("2006 01 02", strings.Join(matches[1:4], " ")); err == nil {
					if !fileDate.After(lastProcessedDate) {
						continue
					}
				}
			}
		}
		validFiles = append(validFiles, file)
	}

	i.fileList = validFiles
	totalFiles := len(validFiles)
	i.totalFiles.Store(int64(totalFiles))

	// Initialize file statuses
	if totalFiles > 0 {
		statuses := make([]*FileProcessingStatus, totalFiles)
		for idx, file := range validFiles {
			statuses[idx] = &FileProcessingStatus{
				FileName: filepath.Base(file),
				Status:   "pending",
				Progress: 0,
			}
		}
		stepState.Metadata["file_list"] = validFiles
		stepState.Metadata["file_statuses"] = statuses
	}

	if i.logger != nil {
		i.logger.Info("Index extraction stage started",
			slog.String("operation_id", state.ID),
			slog.Int("total_files", totalFiles))
	}

	updateStageProgress(i.ID(), stepState, i.broadcaster, 10,
		fmt.Sprintf("Found %d Excel files to process", totalFiles),
		"running", 0, totalFiles, 0, "")

	if totalFiles == 0 {
		updateStageProgress(i.ID(), stepState, i.broadcaster, 100,
			"No Excel files found to process", "completed", 0, 0, 0, "")
		finalizeStageSuccess(i.ID(), stepState, i.broadcaster, "No files to process")
		return nil
	}

	indexPath := filepath.Join(i.executableDir, "indexcsv.exe")
	if _, err := os.Stat(indexPath); err != nil {
		if i.logger != nil {
			i.logger.Error("Index extractor executable not found",
				slog.String("path", indexPath),
				slog.String("error", err.Error()))
		}
		return fmt.Errorf("indexcsv.exe not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, indexPath,
		"-mode", "accumulative",
		"-dir", downloadsDir,
		"-out", indexOutPath,
	)
	cmd.Dir = i.executableDir

	// Execute with progress tracking
	if err := i.executeWithProgress(ctx, cmd, state.ID, stepState, state); err != nil {
		if i.logger != nil {
			i.logger.Error("Index extraction failed",
				slog.String("error", err.Error()))
		}
		return fmt.Errorf("index extraction failed: %w", err)
	}

	return nil
}

// executeWithProgress runs the command with real-time progress tracking
func (i *IndicesStage) executeWithProgress(ctx context.Context, cmd *exec.Cmd, operationID string, stepState *StepState, state *OperationState) error {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start index extractor: %w", err)
	}

	// Track progress
	var currentFile string
	var startTime = time.Now()

	// Use goroutine to read stdout and parse progress
	done := make(chan bool, 2)
	stderrBuffer := &strings.Builder{}

	// Read stderr for error logging
	go func() {
		defer func() { done <- true }()
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuffer.WriteString(line + "\n")
			if i.logger != nil {
				i.logger.Debug("Index extractor stderr",
					slog.String("line", line),
					slog.String("operation_id", operationID))
			}
		}
	}()

	// Read stdout for progress tracking
	go func() {
		defer func() { done <- true }()
		scanner := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if i.logger != nil {
				i.logger.Debug("Index extractor stdout",
					slog.String("line", line),
					slog.String("operation_id", operationID))
			}

			// Parse progress from stdout using new patterns
			i.parseIndexOutputNew(line, &currentFile, stepState)
		}
	}()

	// Wait for stdout and stderr to finish
	<-done
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	stderrOutput := stderrBuffer.String()

	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			if i.logger != nil {
				i.logger.Warn("Index extractor cancelled",
					slog.String("operation_id", operationID))
			}
			return fmt.Errorf("index extractor cancelled: %w", err)
		}
		if i.logger != nil {
			i.logger.Error("Index extractor process failed",
				slog.String("error", err.Error()),
				slog.String("stderr", truncate(stderrOutput, 2000)),
				slog.String("operation_id", operationID))
		}
		return fmt.Errorf("index extractor process failed: %w, stderr: %s", err, truncate(stderrOutput, 2000))
	}

	// Final progress update
	filesProcessed := int(i.filesProcessed.Load())
	totalFiles := int(i.totalFiles.Load())
	indicesExtracted := int(i.indicesExtracted.Load())
	isx15Count := int(i.isx15Extracted.Load())

	// Update final file statuses - mark ALL files as completed since we're at 100%
	if statuses, ok := stepState.Metadata["file_statuses"].([]*FileProcessingStatus); ok {
		for idx := range statuses {
			// Mark all files as completed (including the last one that was still "processing")
			statuses[idx].Status = "completed"
			statuses[idx].Progress = 100
			statuses[idx].ProcessingMs = time.Since(startTime).Milliseconds()
		}
		stepState.Metadata["file_statuses"] = statuses
	}

	// Update metadata with correct field names
	stepState.Metadata["files_processed"] = filesProcessed
	stepState.Metadata["records_processed"] = filesProcessed // For frontend compatibility
	stepState.Metadata["indices_extracted"] = indicesExtracted
	stepState.Metadata["isx15_extracted"] = isx15Count
	stepState.Metadata["total_files"] = totalFiles
	stepState.Metadata["failed_files"] = 0

	completionMsg := fmt.Sprintf("Index extraction completed: %d ISX60 indices from %d files", indicesExtracted, filesProcessed)
	if isx15Count > 0 {
		completionMsg += fmt.Sprintf(", %d ISX15 indices", isx15Count)
	}

	updateStageProgress(i.ID(), stepState, i.broadcaster, 100, completionMsg, "completed",
		filesProcessed, totalFiles, 0, currentFile)
	finalizeStageSuccess(i.ID(), stepState, i.broadcaster, completionMsg)

	if i.logger != nil {
		i.logger.Info("Index extraction completed successfully",
			slog.Int("files_processed", filesProcessed),
			slog.Int("indices_extracted", indicesExtracted),
			slog.Int("isx15_extracted", isx15Count),
			slog.String("operation_id", operationID))
	}

	return nil
}

// parseIndexOutputNew parses output from index extractor using updated patterns
func (i *IndicesStage) parseIndexOutputNew(line string, currentFile *string, stepState *StepState) {
	// Parse "Found N Excel files"
	if matches := reIndexFound.FindStringSubmatch(line); matches != nil {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			i.totalFiles.Store(int64(count))
			if i.logger != nil {
				i.logger.Debug("Detected total file count",
					slog.Int("total_files", count))
			}
		}
	}

	// Parse "Processing file N of M: filename"
	if matches := reIndexProcessing.FindStringSubmatch(line); matches != nil {
		*currentFile = matches[3]
		if currentNum, err := strconv.Atoi(matches[1]); err == nil {
			totalNum, _ := strconv.Atoi(matches[2])
			if totalNum > 0 {
				i.totalFiles.Store(int64(totalNum))
			}

			// Update file statuses
			if statuses, ok := stepState.Metadata["file_statuses"].([]*FileProcessingStatus); ok {
				// Mark previous files as completed (including the one that was "processing")
				for idx := 0; idx < currentNum-1 && idx < len(statuses); idx++ {
					if statuses[idx].Status == "pending" || statuses[idx].Status == "processing" {
						statuses[idx].Status = "completed"
						statuses[idx].Progress = 100
					}
				}
				// Mark current file as processing
				if currentNum-1 < len(statuses) {
					statuses[currentNum-1].Status = "processing"
					statuses[currentNum-1].Progress = 50
					statuses[currentNum-1].FileName = *currentFile
				}
				stepState.Metadata["file_statuses"] = statuses
			}

			i.filesProcessed.Store(int64(currentNum - 1))

			// Calculate and broadcast intermediate progress
			total := int(i.totalFiles.Load())
			processed := currentNum - 1
			indicesCount := int(i.indicesExtracted.Load())
			isx15Count := int(i.isx15Extracted.Load())

			if total > 0 {
				// Progress from 10% (start) to 90% (before completion)
				progress := 10 + (processed * 80 / total)
				message := fmt.Sprintf("Extracting indices from file %d of %d: %s", currentNum, total, *currentFile)

				// Update metadata with current counts
				stepState.Metadata["files_processed"] = processed
				stepState.Metadata["total_files"] = total
				stepState.Metadata["indices_extracted"] = indicesCount
				stepState.Metadata["isx15_extracted"] = isx15Count
				stepState.Metadata["current_file"] = *currentFile
				stepState.Metadata["progress_percent"] = progress

				// Broadcast the progress update
				updateStageProgress(i.ID(), stepState, i.broadcaster, progress, message, "running",
					processed, total, 0, *currentFile)
			}
		}

		if i.logger != nil {
			i.logger.Debug("Processing file",
				slog.String("file", *currentFile))
		}
	}

	// Parse file list from "Files to process: file1|file2|..."
	if matches := reIndexList.FindStringSubmatch(line); matches != nil {
		fileList := strings.Split(matches[1], "|")
		i.totalFiles.Store(int64(len(fileList)))

		// Update file list in metadata
		stepState.Metadata["file_list"] = fileList

		if i.logger != nil {
			i.logger.Debug("Received file list",
				slog.Int("count", len(fileList)))
		}
	}

	// Parse extraction success messages
	if matches := reIndexExtracted15.FindStringSubmatch(line); matches != nil {
		// Extracted both ISX60 and ISX15
		i.indicesExtracted.Add(1)
		i.isx15Extracted.Add(1)
		if i.logger != nil {
			i.logger.Debug("Extracted ISX60 and ISX15",
				slog.String("file", matches[1]),
				slog.String("isx60", matches[2]),
				slog.String("isx15", matches[3]))
		}
	} else if matches := reIndexExtracted.FindStringSubmatch(line); matches != nil {
		// Extracted only ISX60
		i.indicesExtracted.Add(1)
		if i.logger != nil {
			i.logger.Debug("Extracted ISX60 only",
				slog.String("file", matches[1]),
				slog.String("isx60", matches[2]))
		}
	}

	// Parse completion message
	if matches := reIndexComplete.FindStringSubmatch(line); matches != nil {
		if processedCount, err := strconv.Atoi(matches[1]); err == nil {
			i.filesProcessed.Store(int64(processedCount))

			// Count extracted indices from the output CSV file as backup
			if i.indicesExtracted.Load() == 0 {
				i.countExtractedIndices()
			}

			if i.logger != nil {
				i.logger.Debug("Processing complete",
					slog.Int("files_processed", processedCount),
					slog.Int("indices_extracted", int(i.indicesExtracted.Load())),
					slog.Int("isx15_extracted", int(i.isx15Extracted.Load())))
			}
		}
	}
}

// countExtractedIndices reads the generated index CSV and counts extracted values
func (i *IndicesStage) countExtractedIndices() {
	indexPath := filepath.Join(i.executableDir, "data", "reports", "indexes.csv")
	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	file, err := os.Open(indexPath)
	if err != nil {
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return
	}

	// Skip header and count data rows
	isx60Count := 0
	isx15Count := 0

	for _, record := range records[1:] {
		if len(record) > 0 && record[0] != "" {
			if len(record) > 1 && record[1] != "" {
				isx60Count++
			}
			if len(record) > 2 && record[2] != "" {
				isx15Count++
			}
		}
	}

	i.indicesExtracted.Store(int64(isx60Count))
	i.isx15Extracted.Store(int64(isx15Count))

	if i.logger != nil {
		i.logger.Info("Counted extracted indices",
			slog.Int("isx60_count", isx60Count),
			slog.Int("isx15_count", isx15Count))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// updateProgress updates progress through the stage broadcaster
// GetPhase returns the current phase of the index extraction operation
func (i *IndicesStage) GetPhase(stepState *StepState) (StagePhase, string) {
	if stepState == nil {
		return StagePhasePending, "Ready to start"
	}

	progress := int(stepState.Progress)
	if progress < 20 {
		return StagePhaseStarting, "Initializing index extractor"
	} else if progress < 90 {
		return StagePhaseRunning, fmt.Sprintf("Extracting index data (%d%%)", progress)
	} else if progress < 100 {
		return StagePhaseCompleting, "Finalizing extraction"
	} else {
		return StagePhaseCompleted, "Index extraction completed successfully"
	}
}

// CanRun checks if the indices stage can run
func (i *IndicesStage) CanRun() bool {
	return true
}

// RequiredInputs returns the data requirements for this stage
func (i *IndicesStage) RequiredInputs() []DataRequirement {
	return []DataRequirement{
		{
			Type:     DataTypeExcelFiles,
			Required: true,
			MinCount: 1, // At least one Excel file is required
		},
	}
}

// ProducedOutputs returns the data outputs this stage produces
func (i *IndicesStage) ProducedOutputs() []DataOutput {
	return []DataOutput{
		{
			Type:        DataTypeIndexFiles,
			Location:    filepath.Join(i.executableDir, "data", "reports"),
			FilePattern: "*index*.csv",
			Required:    true,
			MinCount:    1, // Should produce at least one index file
		},
	}
}

// Validate checks if the stage can be executed with the current state
func (i *IndicesStage) Validate(state *OperationState) error {
	// Check if executable exists
	indexPath := filepath.Join(i.executableDir, "indexcsv.exe")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("indexcsv.exe not found at %s: %w", indexPath, err)
	}

	// Check if Excel files exist from scraping stage
	downloadsDir := AbsoluteDownloadsPath(i.executableDir, "")
	files, err := filepath.Glob(filepath.Join(downloadsDir, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("failed to check downloads directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no Excel files found in downloads directory - scraping stage may not have completed successfully")
	}

	return nil
}

func loadLastIndexDate(path string) (time.Time, bool) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	reader := csv.NewReader(f)
	var lastDate time.Time
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return time.Time{}, false
		}
		if len(record) == 0 {
			continue
		}
		dateStr := strings.TrimSpace(record[0])
		if dateStr == "" || strings.EqualFold(dateStr, "Date") {
			continue
		}
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		lastDate = parsed
	}

	if lastDate.IsZero() {
		return time.Time{}, false
	}
	return lastDate, true
}

var (
	reIndexFound       = regexp.MustCompile(`^Found (\d+) Excel files`)
	reIndexProcessing  = regexp.MustCompile(`^Processing file (\d+) of (\d+): (.+)$`)
	reIndexComplete    = regexp.MustCompile(`^Index extraction complete: (\d+) files`)
	reIndexList        = regexp.MustCompile(`^Files to process: (.+)$`)
	reIndexExtracted   = regexp.MustCompile(`^Extracted (?:index|indices) from (.+): ISX60=([0-9.,]+)`)                   // Captures filename and ISX60 value
	reIndexExtracted15 = regexp.MustCompile(`^Extracted (?:index|indices) from (.+): ISX60=([0-9.,]+), ISX15=([0-9.,]+)`) // Captures ISX15 too
)
