package stages

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"isxcli/internal/common"
	"isxcli/internal/pipeline"
)

// Test utilities
func createTestLogger() *common.Logger {
	return common.NewLoggerWithComponent("test")
}

func createTestState(stageID string) *pipeline.PipelineState {
	return &pipeline.PipelineState{
		ID:     "test-pipeline",
		Status: pipeline.PipelineStatusRunning,
		Stages: map[string]*pipeline.StageState{
			stageID: {
				ID:       stageID,
				Status:   pipeline.StageStatusPending,
				Progress: 0,
				Metadata: make(map[string]interface{}),
			},
		},
		Context: make(map[string]interface{}),
		Config: map[string]interface{}{
			pipeline.ContextKeyDownloadDir: "testdata/downloads",
			pipeline.ContextKeyReportDir:   "testdata/reports",
			pipeline.ContextKeyFromDate:    "2025-01-01",
			pipeline.ContextKeyToDate:      "2025-01-10",
		},
	}
}

func setupTestDirectories(t *testing.T) {
	os.MkdirAll("testdata/downloads", 0755)
	os.MkdirAll("testdata/reports", 0755)
}

func cleanupTestDirectories(t *testing.T) {
	os.RemoveAll("testdata")
}

// Test ScrapingStage
func TestScrapingStage_Creation(t *testing.T) {
	stage := NewScrapingStage(".", createTestLogger())
	
	if stage.ID() != pipeline.StageIDScraping {
		t.Errorf("Expected stage ID %s, got %s", pipeline.StageIDScraping, stage.ID())
	}
	
	if stage.Name() != pipeline.StageNameScraping {
		t.Errorf("Expected stage name %s, got %s", pipeline.StageNameScraping, stage.Name())
	}
	
	if len(stage.GetDependencies()) != 0 {
		t.Error("Scraping stage should have no dependencies")
	}
}

func TestScrapingStage_Validate(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := NewScrapingStage("testdata", createTestLogger())
	state := createTestState(stage.ID())
	
	// Create mock scraper.exe
	scraperPath := filepath.Join("testdata", "scraper.exe")
	file, _ := os.Create(scraperPath)
	file.Close()
	
	err := stage.Validate(state)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
	
	// Test with missing scraper
	os.Remove(scraperPath)
	err = stage.Validate(state)
	if err == nil {
		t.Error("Expected validation to fail with missing scraper.exe")
	}
}

// Test ProcessingStage
func TestProcessingStage_Creation(t *testing.T) {
	stage := NewProcessingStage(".", createTestLogger())
	
	if stage.ID() != pipeline.StageIDProcessing {
		t.Errorf("Expected stage ID %s, got %s", pipeline.StageIDProcessing, stage.ID())
	}
	
	deps := stage.GetDependencies()
	if len(deps) != 1 || deps[0] != pipeline.StageIDScraping {
		t.Error("Processing stage should depend on scraping stage")
	}
}

func TestProcessingStage_Validate(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := NewProcessingStage("testdata", createTestLogger())
	state := createTestState(stage.ID())
	
	// Create mock process.exe
	processPath := filepath.Join("testdata", "process.exe")
	file, _ := os.Create(processPath)
	file.Close()
	
	// Create test Excel file
	excelFile := filepath.Join("testdata/downloads", "test.xlsx")
	file2, _ := os.Create(excelFile)
	file2.Close()
	
	err := stage.Validate(state)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
	
	// Test with no Excel files
	os.Remove(excelFile)
	err = stage.Validate(state)
	if err == nil {
		t.Error("Expected validation to fail with no Excel files")
	}
}

// Test IndicesStage
func TestIndicesStage_Creation(t *testing.T) {
	stage := NewIndicesStage(".", createTestLogger())
	
	if stage.ID() != pipeline.StageIDIndices {
		t.Errorf("Expected stage ID %s, got %s", pipeline.StageIDIndices, stage.ID())
	}
	
	deps := stage.GetDependencies()
	if len(deps) != 1 || deps[0] != pipeline.StageIDProcessing {
		t.Error("Indices stage should depend on processing stage")
	}
}

func TestIndicesStage_Validate(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := NewIndicesStage("testdata", createTestLogger())
	state := createTestState(stage.ID())
	
	// Create mock indexcsv.exe
	indexPath := filepath.Join("testdata", "indexcsv.exe")
	file, _ := os.Create(indexPath)
	file.Close()
	
	// Create test Excel file
	excelFile := filepath.Join("testdata/downloads", "test.xlsx")
	file2, _ := os.Create(excelFile)
	file2.Close()
	
	err := stage.Validate(state)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

// Test AnalysisStage
func TestAnalysisStage_Creation(t *testing.T) {
	stage := NewAnalysisStage(".", createTestLogger())
	
	if stage.ID() != pipeline.StageIDAnalysis {
		t.Errorf("Expected stage ID %s, got %s", pipeline.StageIDAnalysis, stage.ID())
	}
	
	deps := stage.GetDependencies()
	if len(deps) != 1 || deps[0] != pipeline.StageIDIndices {
		t.Error("Analysis stage should depend on indices stage")
	}
}

func TestAnalysisStage_Validate(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := NewAnalysisStage(".", createTestLogger())
	state := createTestState(stage.ID())
	
	// Create test combined data file
	csvFile := filepath.Join("testdata/reports", "isx_combined_data.csv")
	file, _ := os.Create(csvFile)
	file.Close()
	
	err := stage.Validate(state)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
	
	// Test with missing CSV file
	os.Remove(csvFile)
	err = stage.Validate(state)
	if err == nil {
		t.Error("Expected validation to fail with missing combined data file")
	}
}

// Test Progress Parser
func TestProgressParser(t *testing.T) {
	state := &pipeline.StageState{
		ID:       "test",
		Progress: 0,
		Metadata: make(map[string]interface{}),
	}
	
	parser := NewProgressParser(state)
	
	// First set the total files
	initLine := "[INIT] Files to Download: 10\n"
	parser.Write([]byte(initLine))
	
	// Test download progress parsing
	testLine := "[DOWNLOAD] File 5/10: 2025 01 05 ISX Daily Report.xlsx\n"
	parser.Write([]byte(testLine))
	
	if state.Metadata["current_file"] != "2025 01 05 ISX Daily Report.xlsx" {
		t.Errorf("Failed to parse current file, got: %v", state.Metadata["current_file"])
	}
	
	// Test completion parsing
	testLine2 := "[SUMMARY] ====== Download Complete ======\n"
	parser.Write([]byte(testLine2))
	
	if state.Progress != 95 {
		t.Errorf("Expected progress 95, got %.1f", state.Progress)
	}
}

func TestProcessingProgressParser(t *testing.T) {
	state := &pipeline.StageState{
		ID:       "processing",
		Progress: 0,
		Metadata: make(map[string]interface{}),
	}
	
	parser := NewProcessingProgressParser(state)
	
	// Test file processing parsing - need to access the actual parseLine method
	testLine := "Processing file 3 of 10: 2025 01 03 ISX Daily Report.xlsx\n"
	n, err := parser.Write([]byte(testLine))
	
	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	
	if n != len(testLine) {
		t.Errorf("Write returned %d, expected %d", n, len(testLine))
	}
	
	// The parseLine method is overridden in ProcessingProgressParser
	// Let's directly call it
	parser.parseLine("Processing file 3 of 10: 2025 01 03 ISX Daily Report.xlsx")
	
	if state.Progress != 30 {
		t.Errorf("Expected progress 30, got %.1f", state.Progress)
	}
	
	if state.Metadata["current_file"] != "2025 01 03 ISX Daily Report.xlsx" {
		t.Errorf("Failed to parse current file, got: %v", state.Metadata["current_file"])
	}
	
	if state.Metadata["files_processed"] != 3 {
		t.Errorf("Failed to parse files processed count, got: %v", state.Metadata["files_processed"])
	}
}

// Test file checking functions
func TestCheckNeedsDownload(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := &ScrapingStage{
		executableDir: "testdata",
	}
	
	// Test with no files
	needsDownload := stage.checkNeedsDownload("2025-01-01", "2025-01-10", "testdata/downloads")
	if !needsDownload {
		t.Error("Should need download when no files exist")
	}
	
	// Create a test file
	file, _ := os.Create(filepath.Join("testdata/downloads", "test.xlsx"))
	file.Close()
	
	// Test with specific date range
	needsDownload = stage.checkNeedsDownload("2025-01-01", "2025-01-10", "testdata/downloads")
	if !needsDownload {
		t.Error("Should always download for specific date ranges")
	}
	
	// Test with no date range
	needsDownload = stage.checkNeedsDownload("", "", "testdata/downloads")
	if needsDownload {
		t.Error("Should not need download when files exist and no date range specified")
	}
}

// Integration test for stage dependencies
func TestStageDependencyChain(t *testing.T) {
	stages := []pipeline.Stage{
		NewScrapingStage(".", createTestLogger()),
		NewProcessingStage(".", createTestLogger()),
		NewIndicesStage(".", createTestLogger()),
		NewAnalysisStage(".", createTestLogger()),
	}
	
	// Verify dependency chain
	if len(stages[0].GetDependencies()) != 0 {
		t.Error("Scraping should have no dependencies")
	}
	
	if stages[1].GetDependencies()[0] != stages[0].ID() {
		t.Error("Processing should depend on scraping")
	}
	
	if stages[2].GetDependencies()[0] != stages[1].ID() {
		t.Error("Indices should depend on processing")
	}
	
	if stages[3].GetDependencies()[0] != stages[2].ID() {
		t.Error("Analysis should depend on indices")
	}
}

// Test error creation
func TestAnalysisStage_ErrorHandling(t *testing.T) {
	setupTestDirectories(t)
	defer cleanupTestDirectories(t)
	
	stage := NewAnalysisStage(".", createTestLogger())
	state := createTestState(stage.ID())
	
	// Create invalid CSV file
	csvFile := filepath.Join("testdata/reports", "isx_combined_data.csv")
	file, _ := os.Create(csvFile)
	file.WriteString("invalid,csv,data")
	file.Close()
	
	// Execute should handle errors gracefully
	ctx := context.Background()
	err := stage.Execute(ctx, state)
	
	if err == nil {
		t.Error("Expected error when processing invalid CSV")
	}
	
	// Check if it's the right type of error
	if pErr, ok := err.(*pipeline.PipelineError); !ok {
		t.Error("Expected PipelineError type")
	} else if pErr.Type != pipeline.ErrorTypeExecution {
		t.Errorf("Expected error type %s, got %s", pipeline.ErrorTypeExecution, pErr.Type)
	}
}