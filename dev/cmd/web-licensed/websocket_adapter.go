package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"isxcli/internal/pipeline"
	ws "isxcli/internal/websocket"
)

var debugMode = os.Getenv("ISX_DEBUG") == "true"

// CommandResponse represents the response from a command execution
type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// WebSocketAdapter adapts the pipeline.WebSocketHub interface to the existing ws.Manager
type WebSocketAdapter struct {
	manager *ws.Manager
}

// NewWebSocketAdapter creates a new adapter
func NewWebSocketAdapter(manager *ws.Manager) *WebSocketAdapter {
	return &WebSocketAdapter{
		manager: manager,
	}
}

// BroadcastUpdate implements the pipeline.WebSocketHub interface
func (w *WebSocketAdapter) BroadcastUpdate(eventType, stage, status string, metadata interface{}) {
	// Log the WebSocket message being sent with more detail
	log.Printf("[WEBSOCKET ADAPTER] Sending update - Type: %s, Stage: %s, Status: %s", eventType, stage, status)
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			log.Printf("[WEBSOCKET ADAPTER] Metadata: %s", string(data))
		}
	}
	
	// Ensure stdout is flushed for log visibility
	os.Stdout.Sync()
	
	// The event types are already in frontend format now (pipeline:progress, etc)
	// So we just use them directly
	frontendEventType := eventType
	
	// Transform stage ID from backend to frontend format
	frontendStage := stage
	stageMapping := map[string]string{
		"scraping":   "scrape",
		"processing": "process",
		"indices":    "index",
		"analysis":   "complete",
	}
	if mapped, ok := stageMapping[stage]; ok {
		frontendStage = mapped
	}
	
	// Transform metadata to include frontend stage ID
	if metadataMap, ok := metadata.(map[string]interface{}); ok {
		// Add the frontend stage ID to metadata if not present
		if _, hasStage := metadataMap["stage"]; !hasStage && frontendStage != "" {
			metadataMap["stage"] = frontendStage
		} else if existingStage, hasStage := metadataMap["stage"].(string); hasStage {
			// Transform existing stage ID
			if mapped, ok := stageMapping[existingStage]; ok {
				metadataMap["stage"] = mapped
			}
		}
		metadata = metadataMap
	}
	
	// Use the manager to broadcast the update
	w.manager.Broadcast(ws.Message{
		Type:    frontendEventType,
		Stage:   frontendStage,
		Message: status,
		Data:    metadata,
	})
}

// PipelineLogger adapts the common.Logger to the pipeline.Logger interface
type PipelineLogger struct {
	source    string
	wsManager *ws.Manager
}

// NewPipelineLogger creates a new pipeline logger
func NewPipelineLogger(source string, wsManager *ws.Manager) *PipelineLogger {
	return &PipelineLogger{
		source:    source,
		wsManager: wsManager,
	}
}

// Debug logs a debug message
func (l *PipelineLogger) Debug(format string, v ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] [%s] %s", l.source, fmt.Sprintf(format, v...))
	}
}

// Info logs an info message
func (l *PipelineLogger) Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("[INFO] [%s] %s", l.source, message)
	if l.wsManager != nil {
		l.wsManager.SendLog(ws.LevelInfo, message, map[string]interface{}{
			"source": l.source,
		})
	}
}

// Warn logs a warning message
func (l *PipelineLogger) Warn(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("[WARN] [%s] %s", l.source, message)
	if l.wsManager != nil {
		l.wsManager.SendLog(ws.LevelWarning, message, map[string]interface{}{
			"source": l.source,
		})
	}
}

// Error logs an error message
func (l *PipelineLogger) Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("[ERROR] [%s] %s", l.source, message)
	if l.wsManager != nil {
		l.wsManager.SendLog(ws.LevelError, message, map[string]interface{}{
			"source": l.source,
		})
	}
}

// PipelineEventHandler handles pipeline events and converts them to the existing format
type PipelineEventHandler struct {
	manager *pipeline.Manager
}

// NewPipelineEventHandler creates a new event handler
func NewPipelineEventHandler(manager *pipeline.Manager) *PipelineEventHandler {
	return &PipelineEventHandler{
		manager: manager,
	}
}

// ConvertPipelineResponse converts a pipeline response to the existing CommandResponse format
func ConvertPipelineResponse(resp *pipeline.PipelineResponse) CommandResponse {
	return CommandResponse{
		Success: resp.Status == pipeline.PipelineStatusCompleted,
		Output:  fmt.Sprintf("Pipeline completed with status: %s", resp.Status),
		Error:   resp.Error,
	}
}

// SendPipelineUpdate sends a pipeline update in the format expected by the frontend
func SendPipelineUpdate(hub pipeline.WebSocketHub, pipelineID string, resp *pipeline.PipelineResponse) {
	// Send overall pipeline status
	hub.BroadcastUpdate(ws.TypePipelineStatus, "", "", map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      string(resp.Status),
		"duration":    resp.Duration.Seconds(),
		"stages":      convertStageStates(resp.Stages),
	})
	
	// Send individual stage updates
	for stageID, stage := range resp.Stages {
		hub.BroadcastUpdate(ws.TypePipelineProgress, "", "", map[string]interface{}{
			"pipeline_id": pipelineID,
			"stage":       stageID,
			"status":      string(stage.Status),
			"progress":    stage.Progress,
			"message":     stage.Message,
			"metadata":    stage.Metadata,
		})
	}
}

// convertStageStates converts stage states to a format suitable for JSON
func convertStageStates(stages map[string]*pipeline.StageState) map[string]interface{} {
	result := make(map[string]interface{})
	for id, stage := range stages {
		result[id] = map[string]interface{}{
			"name":       stage.Name,
			"status":     string(stage.Status),
			"progress":   stage.Progress,
			"message":    stage.Message,
			"start_time": stage.StartTime,
			"end_time":   stage.EndTime,
			"duration":   stage.Duration().Seconds(),
			"error":      formatError(stage.Error),
			"metadata":   stage.Metadata,
		}
	}
	return result
}

// formatError formats an error for JSON serialization
func formatError(err error) interface{} {
	if err == nil {
		return nil
	}
	return map[string]string{
		"message": err.Error(),
	}
}

// MonitorPipelineProgress monitors a pipeline and sends progress updates
func MonitorPipelineProgress(hub pipeline.WebSocketHub, pipelineID string, manager *pipeline.Manager) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for range ticker.C {
		state, err := manager.GetPipeline(pipelineID)
		if err != nil {
			// Pipeline no longer exists
			return
		}
		
		// Send progress updates for active stages
		for _, stage := range state.GetActiveStages() {
			// Send in the format the frontend expects for progress messages
			hub.BroadcastUpdate(ws.TypeProgress, "", "", map[string]interface{}{
				"stage":       stage.ID,
				"percentage":  stage.Progress,
				"message":     stage.Message,
				"eta":         nil, // Could calculate ETA if needed
				"metadata":    stage.Metadata,
			})
		}
		
		// Check if pipeline is complete
		if state.Status == pipeline.PipelineStatusCompleted || 
		   state.Status == pipeline.PipelineStatusFailed ||
		   state.Status == pipeline.PipelineStatusCancelled {
			// Send final status
			hub.BroadcastUpdate(ws.TypePipelineComplete, "", "", map[string]interface{}{
				"pipeline_id": pipelineID,
				"status":      string(state.Status),
				"duration":    state.Duration().Seconds(),
			})
			return
		}
	}
}

// ParseWebSocketMessage attempts to parse a WebSocket message from command output
func ParseWebSocketMessage(line string) (map[string]interface{}, error) {
	// Look for JSON between markers
	startMarker := "[WEBSOCKET_"
	
	startIdx := strings.Index(line, startMarker)
	if startIdx == -1 {
		return nil, fmt.Errorf("no WebSocket marker found")
	}
	
	// Find the end of the message type
	typeEndIdx := strings.Index(line[startIdx:], "]")
	if typeEndIdx == -1 {
		return nil, fmt.Errorf("no closing bracket for message type")
	}
	
	// Extract message type
	messageType := line[startIdx+len("[WEBSOCKET_") : startIdx+typeEndIdx]
	
	// Find JSON content
	jsonStartIdx := startIdx + typeEndIdx + 1
	jsonEndIdx := strings.LastIndex(line, "}")
	if jsonEndIdx == -1 || jsonEndIdx < jsonStartIdx {
		// No JSON content, just type
		return map[string]interface{}{
			"type": messageType,
		}, nil
	}
	
	// Extract and parse JSON
	jsonContent := line[jsonStartIdx : jsonEndIdx+1]
	jsonContent = strings.TrimSpace(jsonContent)
	
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	
	data["type"] = messageType
	return data, nil
}