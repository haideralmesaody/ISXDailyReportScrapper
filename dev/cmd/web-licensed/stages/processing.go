package stages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"isxcli/internal/common"
	"isxcli/internal/pipeline"
)

// ProcessingStage implements the data processing stage
type ProcessingStage struct {
	pipeline.BaseStage
	executableDir string
	logger        *common.Logger
}

// NewProcessingStage creates a new processing stage
func NewProcessingStage(executableDir string, logger *common.Logger) *ProcessingStage {
	return &ProcessingStage{
		BaseStage:     pipeline.NewBaseStage(pipeline.StageIDProcessing, pipeline.StageNameProcessing, []string{pipeline.StageIDScraping}),
		executableDir: executableDir,
		logger:        logger,
	}
}

// Execute runs the data processor to convert Excel files to CSV
func (p *ProcessingStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
	p.logger.Info("[PROCESSING] Starting Execute method")
	stageState := state.GetStage(p.ID())
	
	// Get configuration
	downloadDirI, _ := state.GetConfig(pipeline.ContextKeyDownloadDir)
	downloadDir, ok := downloadDirI.(string)
	if !ok || downloadDir == "" {
		downloadDir = "data/downloads"
	}
	p.logger.Info("[PROCESSING] Download directory: %s", downloadDir)
	
	modeI, _ := state.GetConfig(pipeline.ContextKeyMode)
	mode, _ := modeI.(string)
	p.logger.Info("[PROCESSING] Processing mode: %s", mode)
	
	// Build processor arguments
	processPath := filepath.Join(p.executableDir, "process.exe")
	p.logger.Info("[PROCESSING] Process executable path: %s", processPath)
	
	// Check if process.exe exists
	if _, err := os.Stat(processPath); err != nil {
		p.logger.Error("[PROCESSING] process.exe not found at %s: %v", processPath, err)
		return pipeline.NewExecutionError(p.ID(), fmt.Errorf("process.exe not found: %v", err), false)
	}
	
	args := []string{
		fmt.Sprintf("-in=%s", downloadDir),
	}
	
	// Add full mode if requested
	if mode == pipeline.ModeFull {
		args = append(args, "-full")
		stageState.UpdateProgress(5, "Full rework mode enabled - will reprocess all data")
	}
	
	// Update progress
	stageState.UpdateProgress(10, "Starting data processing...")
	fullCommand := fmt.Sprintf("%s %s", processPath, strings.Join(args, " "))
	stageState.Metadata["command"] = fullCommand
	p.logger.Info("[PROCESSING] Full command: %s", fullCommand)
	p.logger.Info("[PROCESSING] Working directory: %s", p.executableDir)
	
	// Execute processor
	cmd := exec.CommandContext(ctx, processPath, args...)
	cmd.Dir = p.executableDir
	
	// Create buffers to capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	
	// Create output handler for processing
	outputHandler := NewProcessingProgressParser(stageState)
	
	// Use MultiWriter to capture output to both buffer and progress parser
	cmd.Stdout = io.MultiWriter(&stdoutBuf, outputHandler)
	cmd.Stderr = io.MultiWriter(&stderrBuf, outputHandler)
	
	// Start execution
	stageState.UpdateProgress(15, "Loading Excel files...")
	p.logger.Info("[PROCESSING] Starting process.exe...")
	err := cmd.Start()
	if err != nil {
		p.logger.Error("[PROCESSING] Failed to start process.exe: %v", err)
		return pipeline.NewExecutionError(p.ID(), fmt.Errorf("failed to start process.exe: %v", err), false)
	}
	
	p.logger.Info("[PROCESSING] Process started with PID: %d", cmd.Process.Pid)
	
	// Wait for completion
	p.logger.Info("[PROCESSING] Waiting for process to complete...")
	err = cmd.Wait()
	
	// Check results
	if err != nil {
		if ctx.Err() == context.Canceled {
			p.logger.Warn("[PROCESSING] Process canceled by context")
			return pipeline.NewCancellationError(p.ID())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			p.logger.Error("[PROCESSING] Process exited with code %d", exitCode)
			
			// Log captured output
			if stdoutBuf.Len() > 0 {
				p.logger.Error("[PROCESSING] Stdout: %s", stdoutBuf.String())
			}
			if stderrBuf.Len() > 0 {
				p.logger.Error("[PROCESSING] Stderr: %s", stderrBuf.String())
			}
			
			return pipeline.NewExecutionError(p.ID(), 
				fmt.Errorf("processor exited with code %d", exitCode), false)
		}
		p.logger.Error("[PROCESSING] Process failed: %v", err)
		
		// Log any captured output
		if stdoutBuf.Len() > 0 {
			p.logger.Error("[PROCESSING] Stdout: %s", stdoutBuf.String())
		}
		if stderrBuf.Len() > 0 {
			p.logger.Error("[PROCESSING] Stderr: %s", stderrBuf.String())
		}
		
		return pipeline.NewExecutionError(p.ID(), fmt.Errorf("process failed: %v", err), false)
	}
	
	// Success - update context
	p.logger.Info("[PROCESSING] Process completed successfully")
	
	// Log any captured output
	if lastOutput, ok := stageState.Metadata["last_output"]; ok {
		p.logger.Info("[PROCESSING] Last output: %v", lastOutput)
	}
	
	filesProcessed := stageState.Metadata["files_processed"]
	if filesProcessed != nil {
		state.SetContext(pipeline.ContextKeyFilesProcessed, filesProcessed)
		p.logger.Info("[PROCESSING] Files processed: %v", filesProcessed)
	}
	
	stageState.UpdateProgress(100, "Data processing completed successfully")
	
	return nil
}

// Validate checks if the stage can be executed
func (p *ProcessingStage) Validate(state *pipeline.PipelineState) error {
	// Check if process.exe exists
	processPath := filepath.Join(p.executableDir, "process.exe")
	if _, err := os.Stat(processPath); err != nil {
		return pipeline.NewValidationError(p.ID(), "process.exe not found")
	}
	
	// Check if we have files to process
	downloadDirI, _ := state.GetConfig(pipeline.ContextKeyDownloadDir)
	downloadDir, ok := downloadDirI.(string)
	if !ok || downloadDir == "" {
		downloadDir = "data/downloads"
	}
	
	// Count Excel files
	files, err := filepath.Glob(filepath.Join(downloadDir, "*.xlsx"))
	if err != nil {
		return pipeline.NewValidationError(p.ID(), fmt.Sprintf("failed to check download directory: %v", err))
	}
	
	if len(files) == 0 {
		return pipeline.NewValidationError(p.ID(), "no Excel files found to process")
	}
	
	return nil
}