package stages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"isxcli/internal/common"
	"isxcli/internal/pipeline"
)

// ScrapingStage implements the data collection stage
type ScrapingStage struct {
	pipeline.BaseStage
	executableDir string
	logger        *common.Logger
}

// NewScrapingStage creates a new scraping stage
func NewScrapingStage(executableDir string, logger *common.Logger) *ScrapingStage {
	return &ScrapingStage{
		BaseStage:     pipeline.NewBaseStage(pipeline.StageIDScraping, pipeline.StageNameScraping, nil),
		executableDir: executableDir,
		logger:        logger,
	}
}

// Execute runs the scraper to download ISX daily reports
func (s *ScrapingStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
	s.logger.Info("[SCRAPING] Starting Execute method")
	stageState := state.GetStage(s.ID())
	
	// Log initial stage state
	s.logger.Info("[SCRAPING] Stage ID: %s, Stage Name: %s", s.ID(), s.Name())
	
	// Get configuration from pipeline state
	fromDateI, _ := state.GetConfig(pipeline.ContextKeyFromDate)
	fromDate, _ := fromDateI.(string)
	toDateI, _ := state.GetConfig(pipeline.ContextKeyToDate)
	toDate, _ := toDateI.(string)
	modeI, _ := state.GetConfig(pipeline.ContextKeyMode)
	mode, _ := modeI.(string)
	downloadDirI, _ := state.GetConfig(pipeline.ContextKeyDownloadDir)
	downloadDir, ok := downloadDirI.(string)
	
	// Default download directory
	if !ok || downloadDir == "" {
		downloadDir = "data/downloads"
	}
	
	s.logger.Info("[SCRAPING] Configuration - From: %s, To: %s, Mode: %s, Dir: %s", 
		fromDate, toDate, mode, downloadDir)
	s.logger.Info("[SCRAPING] Executable directory: %s", s.executableDir)
	
	// Check if we need to download
	needsDownload := s.checkNeedsDownload(fromDate, toDate, downloadDir)
	if !needsDownload {
		stageState.UpdateProgress(100, "Using existing Excel files - no download needed")
		state.SetContext(pipeline.ContextKeyScraperSuccess, true)
		state.SetContext(pipeline.ContextKeyFilesFound, s.countExistingFiles(downloadDir))
		return nil
	}
	
	// Check if license file exists
	licensePath := filepath.Join(s.executableDir, "license.dat")
	if _, err := os.Stat(licensePath); err != nil {
		s.logger.Error("[SCRAPING] License file not found at: %s", licensePath)
		// Also check in current directory
		altLicensePath := "license.dat"
		if _, err2 := os.Stat(altLicensePath); err2 == nil {
			s.logger.Info("[SCRAPING] License file found at alternative location: %s", altLicensePath)
			licensePath = altLicensePath
		} else {
			return pipeline.NewExecutionError(s.ID(), fmt.Errorf("license.dat not found at %s or %s", licensePath, altLicensePath), false)
		}
	} else {
		s.logger.Info("[SCRAPING] License file found at: %s", licensePath)
	}
	
	// Build scraper arguments
	scraperPath := filepath.Join(s.executableDir, "scraper.exe")
	
	// Check if scraper exists
	if _, err := os.Stat(scraperPath); err != nil {
		s.logger.Error("[SCRAPING] scraper.exe not found at: %s", scraperPath)
		return pipeline.NewExecutionError(s.ID(), fmt.Errorf("scraper.exe not found at %s", scraperPath), false)
	}
	s.logger.Info("[SCRAPING] scraper.exe found at: %s", scraperPath)
	
	args := []string{
		fmt.Sprintf("-mode=%s", mode),
		fmt.Sprintf("-out=%s", downloadDir),
	}
	
	if fromDate != "" {
		args = append(args, fmt.Sprintf("-from=%s", fromDate))
	}
	if toDate != "" {
		args = append(args, fmt.Sprintf("-to=%s", toDate))
	}
	
	// Log the full command
	fullCommand := fmt.Sprintf("%s %s", scraperPath, strings.Join(args, " "))
	s.logger.Info("[SCRAPING] Executing command: %s", fullCommand)
	s.logger.Info("[SCRAPING] Working directory: %s", s.executableDir)
	
	// Update progress
	stageState.UpdateProgress(10, "Starting ISX data download...")
	stageState.Metadata["command"] = fullCommand
	
	// Execute scraper
	cmd := exec.CommandContext(ctx, scraperPath, args...)
	cmd.Dir = s.executableDir // Set working directory for license.dat
	
	// Create output handlers
	outputHandler := NewProgressParser(stageState)
	cmd.Stdout = outputHandler
	cmd.Stderr = outputHandler
	
	// Start execution
	stageState.UpdateProgress(20, "Connecting to ISX website...")
	s.logger.Info("[SCRAPING] Starting scraper process...")
	err := cmd.Start()
	if err != nil {
		s.logger.Error("[SCRAPING] Failed to start scraper: %v", err)
		return pipeline.NewExecutionError(s.ID(), fmt.Errorf("failed to start scraper: %v", err), false)
	}
	s.logger.Info("[SCRAPING] Scraper process started with PID: %d", cmd.Process.Pid)
	
	// Wait for completion
	s.logger.Info("[SCRAPING] Waiting for scraper to complete...")
	err = cmd.Wait()
	
	// Check results
	if err != nil {
		s.logger.Error("[SCRAPING] Scraper failed with error: %v", err)
		
		// Get any output from the progress parser
		if lastOutput, ok := stageState.Metadata["last_raw_output"].(string); ok && lastOutput != "" {
			s.logger.Error("[SCRAPING] Last scraper output: %s", lastOutput)
		}
		
		if ctx.Err() == context.Canceled {
			s.logger.Info("[SCRAPING] Scraper was cancelled")
			return pipeline.NewCancellationError(s.ID())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			s.logger.Error("[SCRAPING] Scraper exited with code %d", exitCode)
			// Check for common exit codes
			switch exitCode {
			case 1:
				return pipeline.NewExecutionError(s.ID(), 
					fmt.Errorf("scraper failed: likely license validation error (exit code %d)", exitCode), false)
			default:
				return pipeline.NewExecutionError(s.ID(), 
					fmt.Errorf("scraper exited with code %d - check logs for details", exitCode), false)
			}
		}
		return pipeline.NewExecutionError(s.ID(), fmt.Errorf("scraper execution failed: %v", err), false)
	}
	
	s.logger.Info("[SCRAPING] Scraper completed successfully")
	
	// Success - store results in context
	filesDownloaded := outputHandler.GetFilesDownloaded()
	state.SetContext(pipeline.ContextKeyScraperSuccess, true)
	state.SetContext(pipeline.ContextKeyFilesFound, filesDownloaded)
	
	stageState.UpdateProgress(100, fmt.Sprintf("Downloaded %d files successfully", filesDownloaded))
	stageState.Metadata["files_downloaded"] = filesDownloaded
	
	return nil
}

// Validate checks if the stage can be executed
func (s *ScrapingStage) Validate(state *pipeline.PipelineState) error {
	// Check if scraper.exe exists
	scraperPath := filepath.Join(s.executableDir, "scraper.exe")
	if _, err := os.Stat(scraperPath); err != nil {
		return pipeline.NewValidationError(s.ID(), fmt.Sprintf("scraper.exe not found at %s", scraperPath))
	}
	
	// Check if license.dat exists
	licensePath := filepath.Join(s.executableDir, "license.dat")
	if _, err := os.Stat(licensePath); err != nil {
		return pipeline.NewValidationError(s.ID(), fmt.Sprintf("license.dat not found at %s", licensePath))
	}
	
	// Validate date format if provided
	fromDateI, _ := state.GetConfig(pipeline.ContextKeyFromDate)
	fromDate, _ := fromDateI.(string)
	toDateI, _ := state.GetConfig(pipeline.ContextKeyToDate)
	toDate, _ := toDateI.(string)
	
	if fromDate != "" && !isValidDateFormat(fromDate) {
		return pipeline.NewValidationError(s.ID(), "invalid from date format (expected: YYYY-MM-DD)")
	}
	
	if toDate != "" && !isValidDateFormat(toDate) {
		return pipeline.NewValidationError(s.ID(), "invalid to date format (expected: YYYY-MM-DD)")
	}
	
	return nil
}

// checkNeedsDownload determines if we need to run the scraper
func (s *ScrapingStage) checkNeedsDownload(fromDate, toDate interface{}, downloadDir string) bool {
	// If no date range specified, check if any files exist
	fromDateStr, _ := fromDate.(string)
	toDateStr, _ := toDate.(string)
	if fromDateStr == "" && toDateStr == "" {
		count := s.countExistingFiles(downloadDir)
		return count == 0
	}
	
	// For specific date ranges, always download (scraper will handle existing files)
	return true
}

// countExistingFiles counts Excel files in the download directory
func (s *ScrapingStage) countExistingFiles(downloadDir string) int {
	files, err := filepath.Glob(filepath.Join(downloadDir, "*.xlsx"))
	if err != nil {
		return 0
	}
	return len(files)
}

// isValidDateFormat checks if a date string is in YYYY-MM-DD format
func isValidDateFormat(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}