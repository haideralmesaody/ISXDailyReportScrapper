package stages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"isxcli/internal/common"
	"isxcli/internal/pipeline"
)

// IndicesStage implements the index extraction stage
type IndicesStage struct {
	pipeline.BaseStage
	executableDir string
	logger        *common.Logger
}

// NewIndicesStage creates a new indices extraction stage
func NewIndicesStage(executableDir string, logger *common.Logger) *IndicesStage {
	return &IndicesStage{
		BaseStage:     pipeline.NewBaseStage(pipeline.StageIDIndices, pipeline.StageNameIndices, []string{pipeline.StageIDProcessing}),
		executableDir: executableDir,
		logger:        logger,
	}
}

// Execute runs the index extractor to create indexes.csv
func (i *IndicesStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
	stageState := state.GetStage(i.ID())
	
	// Get configuration
	downloadDirI, _ := state.GetConfig(pipeline.ContextKeyDownloadDir)
	downloadDir, ok := downloadDirI.(string)
	if !ok || downloadDir == "" {
		downloadDir = "data/downloads"
	}
	
	reportDirI, _ := state.GetConfig(pipeline.ContextKeyReportDir)
	reportDir, ok := reportDirI.(string)
	if !ok || reportDir == "" {
		reportDir = "data/reports"
	}
	
	// Build indexcsv arguments
	indexPath := filepath.Join(i.executableDir, "indexcsv.exe")
	outputFile := filepath.Join(reportDir, "indexes.csv")
	args := []string{
		fmt.Sprintf("-dir=%s", downloadDir),
		fmt.Sprintf("-out=%s", outputFile),
	}
	
	// Update progress
	stageState.UpdateProgress(10, "Starting index extraction...")
	stageState.Metadata["command"] = fmt.Sprintf("%s %s", indexPath, strings.Join(args, " "))
	stageState.Metadata["output_file"] = outputFile
	
	// Execute indexcsv
	cmd := exec.CommandContext(ctx, indexPath, args...)
	cmd.Dir = i.executableDir
	
	// Create simple output handler (indices extraction is usually quick)
	outputHandler := NewIndicesProgressParser(stageState)
	cmd.Stdout = outputHandler
	cmd.Stderr = outputHandler
	
	// Start execution
	stageState.UpdateProgress(20, "Extracting ISX60 index data...")
	err := cmd.Start()
	if err != nil {
		return pipeline.NewExecutionError(i.ID(), err, false)
	}
	
	// Wait for completion
	err = cmd.Wait()
	
	// Check results
	if err != nil {
		if ctx.Err() == context.Canceled {
			return pipeline.NewCancellationError(i.ID())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return pipeline.NewExecutionError(i.ID(), 
				fmt.Errorf("indexcsv exited with code %d", exitErr.ExitCode()), false)
		}
		return pipeline.NewExecutionError(i.ID(), err, false)
	}
	
	// Verify output file was created
	if _, err := os.Stat(outputFile); err != nil {
		return pipeline.NewExecutionError(i.ID(), 
			fmt.Errorf("indexes.csv was not created: %v", err), false)
	}
	
	stageState.UpdateProgress(100, "Index extraction completed successfully")
	stageState.Metadata["indices_extracted"] = 2 // ISX60 and ISX15
	
	return nil
}

// Validate checks if the stage can be executed
func (i *IndicesStage) Validate(state *pipeline.PipelineState) error {
	// Check if indexcsv.exe exists
	indexPath := filepath.Join(i.executableDir, "indexcsv.exe")
	if _, err := os.Stat(indexPath); err != nil {
		return pipeline.NewValidationError(i.ID(), "indexcsv.exe not found")
	}
	
	// Check if report directory exists or can be created
	reportDirI, _ := state.GetConfig(pipeline.ContextKeyReportDir)
	reportDir, ok := reportDirI.(string)
	if !ok || reportDir == "" {
		reportDir = "data/reports"
	}
	
	// Create report directory if it doesn't exist
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return pipeline.NewValidationError(i.ID(), fmt.Sprintf("failed to create report directory: %v", err))
	}
	
	// Check if we have files to extract indices from
	downloadDirI, _ := state.GetConfig(pipeline.ContextKeyDownloadDir)
	downloadDir, ok := downloadDirI.(string)
	if !ok || downloadDir == "" {
		downloadDir = "data/downloads"
	}
	
	files, err := filepath.Glob(filepath.Join(downloadDir, "*.xlsx"))
	if err != nil {
		return pipeline.NewValidationError(i.ID(), fmt.Sprintf("failed to check download directory: %v", err))
	}
	
	if len(files) == 0 {
		return pipeline.NewValidationError(i.ID(), "no Excel files found for index extraction")
	}
	
	return nil
}

// IndicesProgressParser handles progress for indexcsv.exe
type IndicesProgressParser struct {
	*ProgressParser
}

// NewIndicesProgressParser creates a parser for indices stage
func NewIndicesProgressParser(stageState *pipeline.StageState) *IndicesProgressParser {
	return &IndicesProgressParser{
		ProgressParser: NewProgressParser(stageState),
	}
}

// parseLine parses indices-specific output
func (p *IndicesProgressParser) parseLine(line string) {
	// Let base parser handle common patterns
	p.ProgressParser.parseLine(line)
	
	// Handle indices-specific patterns
	line = strings.TrimSpace(line)
	switch {
	case strings.Contains(line, "Extracting ISX60"):
		p.stageState.UpdateProgress(40, "Extracting ISX60 index data...")
		
	case strings.Contains(line, "Extracting ISX15"):
		p.stageState.UpdateProgress(70, "Extracting ISX15 index data...")
		
	case strings.Contains(line, "Writing to"):
		p.stageState.UpdateProgress(90, "Writing index data to CSV...")
		
	case strings.Contains(line, "Index extraction complete"):
		p.stageState.UpdateProgress(95, "Finalizing index data...")
	}
}