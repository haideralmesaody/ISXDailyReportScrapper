package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"isxcli/internal/pipeline"
)

// PipelineService manages pipeline operations
type PipelineService struct {
	manager *pipeline.Manager
	logger  Logger
	paths   string
}

// WebSocketPipelineAdapter adapts WebSocket communication for pipeline

type WebSocketPipelineAdapter struct {
	hub WebSocketHub
}

// NewWebSocketPipelineAdapter creates a new WebSocket pipeline adapter
func NewWebSocketPipelineAdapter(hub WebSocketHub) *WebSocketPipelineAdapter {
	return &WebSocketPipelineAdapter{hub: hub}
}

// WebSocketHub interface for WebSocket communication
type WebSocketHub interface {
	Broadcast(messageType string, data interface{})
}

// WebSocketPipelineAdapter implements PipelineAdapter
func (w *WebSocketPipelineAdapter) SendProgress(stageID, message string, progress int) {
	w.hub.Broadcast("pipeline_progress", map[string]interface{}{
		"stage":    stageID,
		"message":  message,
		"progress": progress,
		"status":   "active",
	})
}

func (w *WebSocketPipelineAdapter) SendComplete(stageID, message string, success bool) {
	status := "completed"
	if !success {
		status = "failed"
	}
	w.hub.Broadcast("pipeline_complete", map[string]interface{}{
		"stage":   stageID,
		"message": message,
		"status":  status,
		"success": success,
	})
}

func (w *WebSocketPipelineAdapter) SendError(stageID, error string) {
	w.hub.Broadcast("pipeline_error", map[string]interface{}{
		"stage": stageID,
		"error": error,
		"status": "error",
	})
}

// BroadcastUpdate implements pipeline.WebSocketHub interface
func (w *WebSocketPipelineAdapter) BroadcastUpdate(eventType, stage, status string, metadata interface{}) {
	data := map[string]interface{}{
		"eventType": eventType,
		"stage":     stage,
		"status":    status,
	}
	if metadata != nil {
		data["metadata"] = metadata
	}
	w.hub.Broadcast(eventType, data)
}

// PipelineLogger implements Logger interface for pipeline

type PipelineLogger struct {
	name string
	hub  WebSocketHub
}

// NewPipelineLogger creates a new pipeline logger
func NewPipelineLogger(name string, hub WebSocketHub) *PipelineLogger {
	return &PipelineLogger{
		name: name,
		hub:  hub,
	}
}

// Debug logs debug messages
func (pl *PipelineLogger) Debug(format string, v ...interface{}) {
	pl.log("debug", fmt.Sprintf(format, v...))
}

// Info logs info messages
func (pl *PipelineLogger) Info(format string, v ...interface{}) {
	pl.log("info", fmt.Sprintf(format, v...))
}

// Warn logs warning messages
func (pl *PipelineLogger) Warn(format string, v ...interface{}) {
	pl.log("warn", fmt.Sprintf(format, v...))
}

// Error logs error messages
func (pl *PipelineLogger) Error(format string, v ...interface{}) {
	pl.log("error", fmt.Sprintf(format, v...))
}

// log is internal method for logging
func (pl *PipelineLogger) log(level, message string) {
	pl.hub.Broadcast("log", map[string]interface{}{
		"level":     level,
		"message":   fmt.Sprintf("[%s] %s", pl.name, message),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// NewPipelineService creates a new pipeline service
func NewPipelineService(adapter *WebSocketPipelineAdapter, logger *PipelineLogger, appLogger Logger) (*PipelineService, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	paths := filepath.Dir(execPath)

	pipelineAdapter := adapter
	pipelineLogger := logger
	
	manager := pipeline.NewManager(pipelineAdapter, pipelineLogger)
	
	// Register pipeline stages
	if err := registerStages(manager, paths, appLogger); err != nil {
		return nil, fmt.Errorf("failed to register stages: %w", err)
	}

	return &PipelineService{
		manager: manager,
		logger:  appLogger,
		paths:   paths,
	}, nil
}

// registerStages registers all pipeline stages
func registerStages(manager *pipeline.Manager, paths string, logger Logger) error {
	// Create stages
	// Create stages without WebSocket integration (service layer)
	scraper := pipeline.NewScrapingStage(paths, logger, nil)
	processor := pipeline.NewProcessingStage(paths, logger, nil)
	indices := pipeline.NewIndicesStage(paths, logger, nil)
	analysis := pipeline.NewAnalysisStage(paths, logger, nil)

	// Register stages
	manager.GetRegistry().Register(scraper)
	manager.GetRegistry().Register(processor)
	manager.GetRegistry().Register(indices)
	manager.GetRegistry().Register(analysis)

	return nil
}

// StartPipeline starts a new pipeline execution
func (ps *PipelineService) StartPipeline(params map[string]interface{}) (string, error) {
	ctx := context.Background()
	
	// Create pipeline request
	request := pipeline.PipelineRequest{
		ID:         fmt.Sprintf("pipeline-%d", time.Now().Unix()),
		Mode:       "full",
		Parameters: params,
	}

	resp, err := ps.manager.Execute(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to start pipeline: %w", err)
	}

	ps.logger.Info("Pipeline started", "id", resp.ID, "status", resp.Status)
	return resp.ID, nil
}

// StartScraping starts the scraping stage
func (ps *PipelineService) StartScraping(params map[string]interface{}) (string, error) {
	// Log received parameters for debugging
	ps.logger.Info("StartScraping received params", "params", params)
	
	// Extract args from the params structure
	args, ok := params["args"].(map[string]interface{})
	if !ok {
		// Fallback to direct params if no args wrapper
		args = params
		ps.logger.Warn("No 'args' wrapper found, using params directly")
	}
	
	// Build pipeline parameters with correct field names
	scrapingParams := map[string]interface{}{
		"mode":      getValue(args, "mode", "initial"),
		"from_date": getValue(args, "from", ""),  // Map 'from' to 'from_date'
		"to_date":   getValue(args, "to", ""),    // Map 'to' to 'to_date'
		"headless":  getValue(args, "headless", true),
		"stage":     "scraping",
	}
	
	// Log transformed parameters
	ps.logger.Info("Transformed scraping params", "scrapingParams", scrapingParams)

	return ps.StartPipeline(scrapingParams)
}

// StartProcessing starts the processing stage
func (ps *PipelineService) StartProcessing(params map[string]interface{}) (string, error) {
	processingParams := map[string]interface{}{
		"stage":     "processing",
		"input_dir": params["input_dir"],
		"mode":      params["mode"],
	}

	return ps.StartPipeline(processingParams)
}

// StartIndexExtraction starts the index extraction stage
func (ps *PipelineService) StartIndexExtraction(params map[string]interface{}) (string, error) {
	indexParams := map[string]interface{}{
		"stage": "indices",
		"mode":  "full",
	}

	return ps.StartPipeline(indexParams)
}

// StopPipeline stops a running pipeline
func (ps *PipelineService) StopPipeline(pipelineID string) error {
	if err := ps.manager.CancelPipeline(pipelineID); err != nil {
		return fmt.Errorf("failed to stop pipeline: %w", err)
	}

	ps.logger.Info("Pipeline stopped", "id", pipelineID)
	return nil
}

// GetStatus returns pipeline status
func (ps *PipelineService) GetStatus(pipelineID string) (*pipeline.PipelineState, error) {
	if pipelineID == "" {
		return nil, fmt.Errorf("pipeline ID is required")
	}

	state, err := ps.manager.GetPipeline(pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	return state, nil
}

// ListPipelines returns all pipelines
func (ps *PipelineService) ListPipelines() []*pipeline.PipelineState {
	return ps.manager.ListPipelines()
}

// CancelAll stops all running pipelines
func (ps *PipelineService) CancelAll() error {
	pipelines := ps.manager.ListPipelines()
	for _, p := range pipelines {
		if p.Status == pipeline.PipelineStatusRunning {
			if err := ps.manager.CancelPipeline(p.ID); err != nil {
				ps.logger.Error("Failed to cancel pipeline", "id", p.ID, "error", err)
				return err
			}
		}
	}
	return nil
}

// ExecuteStage executes a specific stage
func (ps *PipelineService) ExecuteStage(stageID string, ctx context.Context) error {
	// This would execute individual stages - implement as needed
	return fmt.Errorf("individual stage execution not implemented")
}

// ValidateExecutables checks if required executables exist
func (ps *PipelineService) ValidateExecutables() error {
	executables := []string{
		"scraper.exe",
		"process.exe",
		"indexcsv.exe",
	}

	for _, exe := range executables {
		path := filepath.Join(ps.paths, exe)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("required executable not found: %s", exe)
		}
	}

	return nil
}

// GetManager returns the underlying pipeline manager
func (ps *PipelineService) GetManager() *pipeline.Manager {
	return ps.manager
}

// GetStageInfo returns information about available stages
func (ps *PipelineService) GetStageInfo() map[string]interface{} {
	return map[string]interface{}{
		"stages": []map[string]interface{}{
			{
				"id":   "scraping",
				"name": "Scraping",
				"description": "Download daily reports from ISX website",
				"executable":  "scraper.exe",
			},
			{
				"id":   "processing",
				"name": "Processing",
				"description": "Process Excel files into CSV format",
				"executable":  "process.exe",
			},
			{
				"id":   "indices",
				"name": "Index Extraction",
				"description": "Extract market indices from processed data",
				"executable":  "indexcsv.exe",
			},
			{
				"id":   "analysis",
				"name": "Analysis",
				"description": "Generate analytical reports",
				"executable":  "",
			},
		},
	}
}

// getValue safely extracts a value from a map with a default
func getValue(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := m[key]; ok && val != nil {
		return val
	}
	return defaultValue
}