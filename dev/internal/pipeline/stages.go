package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScrapingStage handles the scraping process
type ScrapingStage struct {
	BaseStage
	executableDir string
	logger        Logger
	options       *StageOptions
}

// NewScrapingStage creates a new scraping stage
func NewScrapingStage(executableDir string, logger Logger, options *StageOptions) *ScrapingStage {
	if options == nil {
		options = &StageOptions{}
	}
	return &ScrapingStage{
		BaseStage:     NewBaseStage(StageIDScraping, StageNameScraping, nil),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
	}
}

// Execute runs the scraper to download ISX daily reports
func (s *ScrapingStage) Execute(ctx context.Context, state *PipelineState) error {
	stageState := state.GetStage(s.ID())
	
	// Check license if required
	if s.options.LicenseChecker != nil && s.options.LicenseChecker.RequiresLicense() {
		if err := s.options.LicenseChecker.CheckLicense(); err != nil {
			return fmt.Errorf("license check failed: %w", err)
		}
	}
	
	s.updateProgress(stageState, 10, "Starting scraper...")

	scraperPath := filepath.Join(s.executableDir, "scraper.exe")
	if _, err := os.Stat(scraperPath); err != nil {
		return fmt.Errorf("scraper.exe not found: %w", err)
	}

	// Build command arguments
	args := s.buildScraperArgs(state)
	cmd := exec.CommandContext(ctx, scraperPath, args...)
	cmd.Dir = s.executableDir

	s.updateProgress(stageState, 50, "Running scraper...")
	
	// Execute with progress tracking if enabled
	if s.options.EnableProgress && s.options.WebSocketManager != nil {
		if err := s.executeWithProgress(ctx, cmd, stageState); err != nil {
			return fmt.Errorf("scraper failed: %w", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("scraper failed: %w, output: %s", err, string(output))
		}
	}

	s.updateProgress(stageState, 100, "Scraping completed")
	return nil
}

// buildScraperArgs builds command line arguments from pipeline state
func (s *ScrapingStage) buildScraperArgs(state *PipelineState) []string {
	args := []string{}
	
	// Get configuration from pipeline state
	if fromDateI, exists := state.GetConfig(ContextKeyFromDate); exists {
		if fromDate, ok := fromDateI.(string); ok && fromDate != "" {
			args = append(args, "--from", fromDate)
			s.logger.Info("Added from date to scraper args: %s", fromDate)
		}
	}
	if toDateI, exists := state.GetConfig(ContextKeyToDate); exists {
		if toDate, ok := toDateI.(string); ok && toDate != "" {
			args = append(args, "--to", toDate)
			s.logger.Info("Added to date to scraper args: %s", toDate)
		}
	}
	if modeI, exists := state.GetConfig(ContextKeyMode); exists {
		if mode, ok := modeI.(string); ok && mode != "" {
			args = append(args, "--mode", mode)
		} else {
			args = append(args, "--mode", "full")
		}
	} else {
		args = append(args, "--mode", "full")
	}
	
	s.logger.Info("Final scraper args: %v", args)
	return args
}

// executeWithProgress runs the command with real-time progress tracking
func (s *ScrapingStage) executeWithProgress(ctx context.Context, cmd *exec.Cmd, stageState *StageState) error {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start scraper: %w", err)
	}

	// Track progress
	var totalFiles, downloadedFiles, existingFiles int
	var currentPage int
	progressChan := make(chan string, 100)
	errChan := make(chan error, 2)

	// Read stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			progressChan <- line
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stdout scan error: %w", err)
		}
		close(progressChan)
	}()

	// Read stderr in goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		var errOutput strings.Builder
		for scanner.Scan() {
			errOutput.WriteString(scanner.Text() + "\n")
		}
		if errOutput.Len() > 0 {
			errChan <- fmt.Errorf("stderr output: %s", errOutput.String())
		}
	}()

	// Process output lines
	for line := range progressChan {
		s.logger.Debug("Scraper output: %s", line)

		// Parse different types of messages
		switch {
		case strings.Contains(line, "Processing page"):
			// Extract page number
			if _, err := fmt.Sscanf(line, "Processing page %d", &currentPage); err == nil {
				s.updateProgress(stageState, 20+currentPage*5, fmt.Sprintf("Processing page %d", currentPage))
			}

		case strings.Contains(line, "downloading"):
			// File being downloaded
			downloadedFiles++
			totalProcessed := downloadedFiles + existingFiles
			message := fmt.Sprintf("Downloading file %d (%d existing)", downloadedFiles, existingFiles)
			
			// Calculate progress (20-80% range for file downloads)
			progress := 20 + int(float64(totalProcessed)*60/float64(totalProcessed+10))
			if progress > 80 {
				progress = 80
			}
			s.updateProgress(stageState, progress, message)
			
			// Update metadata
			stageState.Metadata["files_downloaded"] = downloadedFiles
			stageState.Metadata["files_existing"] = existingFiles
			stageState.Metadata["current_page"] = currentPage

		case strings.Contains(line, "already have"):
			// Existing file found
			existingFiles++
			totalProcessed := downloadedFiles + existingFiles
			message := fmt.Sprintf("Found existing file (%d downloaded, %d existing)", downloadedFiles, existingFiles)
			
			// Update progress
			progress := 20 + int(float64(totalProcessed)*60/float64(totalProcessed+10))
			if progress > 80 {
				progress = 80
			}
			s.updateProgress(stageState, progress, message)
			
			// Update metadata
			stageState.Metadata["files_existing"] = existingFiles

		case strings.Contains(line, "Page summary"):
			// Page processing complete
			var pageNew, pageExisting int
			if _, err := fmt.Sscanf(line, "Page summary: %d new downloads, %d existing files", &pageNew, &pageExisting); err == nil {
				s.logger.Info("Page %d complete: %d new, %d existing", currentPage, pageNew, pageExisting)
			}

		case strings.Contains(line, "All pages processed"):
			// Scraping complete
			s.updateProgress(stageState, 90, "Finalizing downloads")
			totalFiles = downloadedFiles + existingFiles
			stageState.Metadata["total_files"] = totalFiles
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		select {
		case stderr := <-errChan:
			return fmt.Errorf("scraper failed: %w, stderr: %v", err, stderr)
		default:
			return fmt.Errorf("scraper failed: %w", err)
		}
	}

	// Final progress update
	s.updateProgress(stageState, 95, fmt.Sprintf("Scraping complete: %d files downloaded, %d existing", downloadedFiles, existingFiles))
	
	// Verify files were actually downloaded
	if downloadedFiles == 0 && existingFiles == 0 {
		return fmt.Errorf("no files were downloaded or found")
	}

	return nil
}

// updateProgress updates progress and optionally sends WebSocket updates
func (s *ScrapingStage) updateProgress(stageState *StageState, progress int, message string) {
	stageState.UpdateProgress(float64(progress), message)
	
	if s.options.WebSocketManager != nil {
		s.options.WebSocketManager.BroadcastUpdate(
			"stage_progress",
			s.ID(),
			"progress",
			map[string]interface{}{
				"progress": progress,
				"message":  message,
			},
		)
	}
}

// ProcessingStage handles data processing
type ProcessingStage struct {
	BaseStage
	executableDir string
	logger        Logger
	options       *StageOptions
}

// NewProcessingStage creates a new processing stage
func NewProcessingStage(executableDir string, logger Logger, options *StageOptions) *ProcessingStage {
	if options == nil {
		options = &StageOptions{}
	}
	return &ProcessingStage{
		BaseStage:     NewBaseStage(StageIDProcessing, StageNameProcessing, []string{StageIDScraping}),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
	}
}

// Execute runs the processor to convert Excel files to CSV
func (p *ProcessingStage) Execute(ctx context.Context, state *PipelineState) error {
	stageState := state.GetStage(p.ID())
	p.updateProgress(stageState, 10, "Starting processor...")

	processorPath := filepath.Join(p.executableDir, "process.exe")
	if _, err := os.Stat(processorPath); err != nil {
		return fmt.Errorf("process.exe not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, processorPath)
	cmd.Dir = p.executableDir

	p.updateProgress(stageState, 50, "Processing data...")
	
	if p.options.EnableProgress && p.options.WebSocketManager != nil {
		if err := p.executeWithProgress(ctx, cmd, stageState); err != nil {
			return fmt.Errorf("processor failed: %w", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("processor failed: %w, output: %s", err, string(output))
		}
	}

	p.updateProgress(stageState, 100, "Processing completed")
	return nil
}

// executeWithProgress runs the command with real-time progress tracking
func (p *ProcessingStage) executeWithProgress(ctx context.Context, cmd *exec.Cmd, stageState *StageState) error {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}

	// Track progress
	var processedFiles, totalFiles int
	progressChan := make(chan string, 100)
	errChan := make(chan error, 2)

	// Read stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			progressChan <- line
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stdout scan error: %w", err)
		}
		close(progressChan)
	}()

	// Read stderr in goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		var errOutput strings.Builder
		for scanner.Scan() {
			errOutput.WriteString(scanner.Text() + "\n")
		}
		if errOutput.Len() > 0 {
			errChan <- fmt.Errorf("stderr output: %s", errOutput.String())
		}
	}()

	// Process output lines
	for line := range progressChan {
		p.logger.Debug("Processor output: %s", line)

		// Parse different types of messages
		switch {
		case strings.Contains(line, "Processing") && strings.Contains(line, ".xlsx"):
			// Processing a file
			processedFiles++
			fileName := ""
			if idx := strings.Index(line, "Processing"); idx >= 0 {
				fileName = strings.TrimSpace(line[idx+10:])
			}
			
			// Calculate progress (20-80% range)
			progress := 20
			if totalFiles > 0 {
				progress = 20 + int(float64(processedFiles)*60/float64(totalFiles))
				if progress > 80 {
					progress = 80
				}
			}
			
			message := fmt.Sprintf("Processing file %d: %s", processedFiles, fileName)
			p.updateProgress(stageState, progress, message)
			
			// Update metadata
			stageState.Metadata["files_processed"] = processedFiles

		case strings.Contains(line, "Found") && strings.Contains(line, "Excel files"):
			// Total files detected
			if _, err := fmt.Sscanf(line, "Found %d Excel files", &totalFiles); err == nil {
				p.updateProgress(stageState, 15, fmt.Sprintf("Found %d Excel files to process", totalFiles))
				stageState.Metadata["total_files"] = totalFiles
			}

		case strings.Contains(line, "Converted") && strings.Contains(line, "to CSV"):
			// File conversion complete
			p.logger.Info("File conversion: %s", line)

		case strings.Contains(line, "All files processed"):
			// Processing complete
			p.updateProgress(stageState, 90, "Finalizing CSV conversion")
			stageState.Metadata["csv_files_created"] = processedFiles
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		select {
		case stderr := <-errChan:
			return fmt.Errorf("processor failed: %w, stderr: %v", err, stderr)
		default:
			return fmt.Errorf("processor failed: %w", err)
		}
	}

	// Final progress update
	p.updateProgress(stageState, 95, fmt.Sprintf("Processing complete: %d files converted to CSV", processedFiles))
	
	// Verify files were processed
	if processedFiles == 0 {
		return fmt.Errorf("no files were processed")
	}

	return nil
}

// updateProgress updates progress and optionally sends WebSocket updates
func (p *ProcessingStage) updateProgress(stageState *StageState, progress int, message string) {
	stageState.UpdateProgress(float64(progress), message)
	
	if p.options.WebSocketManager != nil {
		p.options.WebSocketManager.BroadcastUpdate(
			"stage_progress",
			p.ID(),
			"progress",
			map[string]interface{}{
				"progress": progress,
				"message":  message,
			},
		)
	}
}

// IndicesStage handles index extraction
type IndicesStage struct {
	BaseStage
	executableDir string
	logger        Logger
	options       *StageOptions
}

// NewIndicesStage creates a new indices extraction stage
func NewIndicesStage(executableDir string, logger Logger, options *StageOptions) *IndicesStage {
	if options == nil {
		options = &StageOptions{}
	}
	return &IndicesStage{
		BaseStage:     NewBaseStage(StageIDIndices, StageNameIndices, []string{StageIDProcessing}),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
	}
}

// Execute runs the index extractor
func (i *IndicesStage) Execute(ctx context.Context, state *PipelineState) error {
	stageState := state.GetStage(i.ID())
	i.updateProgress(stageState, 10, "Starting index extractor...")

	indexPath := filepath.Join(i.executableDir, "indexcsv.exe")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("indexcsv.exe not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, indexPath)
	cmd.Dir = i.executableDir

	i.updateProgress(stageState, 50, "Extracting indices...")
	
	if i.options.EnableProgress && i.options.WebSocketManager != nil {
		if err := i.executeWithProgress(ctx, cmd, stageState); err != nil {
			return fmt.Errorf("index extraction failed: %w", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("index extraction failed: %w, output: %s", err, string(output))
		}
	}

	i.updateProgress(stageState, 100, "Index extraction completed")
	return nil
}

// executeWithProgress runs the command with real-time progress tracking
func (i *IndicesStage) executeWithProgress(ctx context.Context, cmd *exec.Cmd, stageState *StageState) error {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start index extractor: %w", err)
	}

	// Track progress
	var processedFiles int
	var indicesExtracted []string
	progressChan := make(chan string, 100)
	errChan := make(chan error, 2)

	// Read stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			progressChan <- line
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("stdout scan error: %w", err)
		}
		close(progressChan)
	}()

	// Read stderr in goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		var errOutput strings.Builder
		for scanner.Scan() {
			errOutput.WriteString(scanner.Text() + "\n")
		}
		if errOutput.Len() > 0 {
			errChan <- fmt.Errorf("stderr output: %s", errOutput.String())
		}
	}()

	// Process output lines
	for line := range progressChan {
		i.logger.Debug("Index extractor output: %s", line)

		// Parse different types of messages
		switch {
		case strings.Contains(line, "Processing CSV file"):
			// Processing a CSV file
			processedFiles++
			message := fmt.Sprintf("Processing CSV file %d", processedFiles)
			progress := 20 + (processedFiles * 10)
			if progress > 70 {
				progress = 70
			}
			i.updateProgress(stageState, progress, message)
			stageState.Metadata["csv_files_processed"] = processedFiles

		case strings.Contains(line, "Extracting ISX60"):
			// Extracting ISX60 index
			i.updateProgress(stageState, 75, "Extracting ISX60 index data")
			indicesExtracted = append(indicesExtracted, "ISX60")

		case strings.Contains(line, "Extracting ISX15"):
			// Extracting ISX15 index
			i.updateProgress(stageState, 85, "Extracting ISX15 index data")
			indicesExtracted = append(indicesExtracted, "ISX15")

		case strings.Contains(line, "Index data saved"):
			// Index extraction complete
			i.updateProgress(stageState, 90, "Index data saved to indexes.csv")
			stageState.Metadata["indices_extracted"] = indicesExtracted

		case strings.Contains(line, "Index extraction complete"):
			// All done
			i.updateProgress(stageState, 95, "Index extraction complete")
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		select {
		case stderr := <-errChan:
			return fmt.Errorf("index extractor failed: %w, stderr: %v", err, stderr)
		default:
			return fmt.Errorf("index extractor failed: %w", err)
		}
	}

	// Verify index file was created
	indexFile := filepath.Join(i.executableDir, "data", "reports", "indexes.csv")
	if _, err := os.Stat(indexFile); err != nil {
		return fmt.Errorf("index file not created: %w", err)
	}

	// Final progress update
	i.updateProgress(stageState, 100, fmt.Sprintf("Index extraction complete: %d indices extracted", len(indicesExtracted)))
	
	return nil
}

// updateProgress updates progress and optionally sends WebSocket updates
func (i *IndicesStage) updateProgress(stageState *StageState, progress int, message string) {
	stageState.UpdateProgress(float64(progress), message)
	
	if i.options.WebSocketManager != nil {
		i.options.WebSocketManager.BroadcastUpdate(
			"stage_progress",
			i.ID(),
			"progress",
			map[string]interface{}{
				"progress": progress,
				"message":  message,
			},
		)
	}
}

// AnalysisStage handles ticker analysis
type AnalysisStage struct {
	BaseStage
	executableDir string
	logger        Logger
	options       *StageOptions
}

// NewAnalysisStage creates a new analysis stage
func NewAnalysisStage(executableDir string, logger Logger, options *StageOptions) *AnalysisStage {
	if options == nil {
		options = &StageOptions{}
	}
	return &AnalysisStage{
		BaseStage:     NewBaseStage(StageIDAnalysis, StageNameAnalysis, []string{StageIDIndices}),
		executableDir: executableDir,
		logger:        logger,
		options:       options,
	}
}

// Execute runs the ticker analysis
func (a *AnalysisStage) Execute(ctx context.Context, state *PipelineState) error {
	stageState := state.GetStage(a.ID())
	a.updateProgress(stageState, 10, "Starting ticker analysis...")

	// TODO: Implement actual analysis logic
	// This would include:
	// - Calculate market statistics
	// - Generate ticker summaries
	// - Create performance reports

	a.updateProgress(stageState, 100, "Analysis completed")
	return nil
}

// updateProgress updates progress and optionally sends WebSocket updates
func (a *AnalysisStage) updateProgress(stageState *StageState, progress int, message string) {
	stageState.UpdateProgress(float64(progress), message)
	
	if a.options.WebSocketManager != nil {
		a.options.WebSocketManager.BroadcastUpdate(
			"stage_progress",
			a.ID(),
			"progress",
			map[string]interface{}{
				"progress": progress,
				"message":  message,
			},
		)
	}
}

// StageFactory creates pipeline stages with optional configuration
func StageFactory(executableDir string, logger Logger, options *StageOptions) map[string]Stage {
	return map[string]Stage{
		StageIDScraping:   NewScrapingStage(executableDir, logger, options),
		StageIDProcessing: NewProcessingStage(executableDir, logger, options),
		StageIDIndices:    NewIndicesStage(executableDir, logger, options),
		StageIDAnalysis:   NewAnalysisStage(executableDir, logger, options),
	}
}