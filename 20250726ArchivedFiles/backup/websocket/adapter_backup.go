package websocket

import (
	"log/slog"
	"sync"
)

// MessageAdapter converts between old and new WebSocket message formats
type MessageAdapter struct {
	hub    *Hub
	logger *slog.Logger
	
	// Pipeline tracking for proper message context
	activePipelines map[string]string // pipelineID -> current stage
	mutex           sync.RWMutex
}

// NewMessageAdapter creates a new message format adapter
func NewMessageAdapter(hub *Hub, logger *slog.Logger) *MessageAdapter {
	return &MessageAdapter{
		hub:             hub,
		logger:          logger.With(slog.String("component", "websocket_adapter")),
		activePipelines: make(map[string]string),
	}
}

// BroadcastUpdate adapts old BroadcastUpdate calls to new message format
func (a *MessageAdapter) BroadcastUpdate(updateType, subtype, action string, data interface{}) {
	switch updateType {
	case "stage_progress":
		// Convert stage progress to pipeline progress
		a.handleStageProgress(subtype, data)
		
	case "pipeline_status":
		// Convert to pipeline status message
		a.handlePipelineStatus(data)
		
	case TypeDataUpdate:
		// Standard data update
		a.hub.BroadcastDataUpdate(subtype, action, data)
		
	case TypeOutput:
		// Convert to output message
		a.handleOutput(data)
		
	case TypeError:
		// Convert to error message
		a.handleError(data)
		
	default:
		// Log unknown type and broadcast as-is
		a.logger.Warn("unknown message type, broadcasting as legacy",
			slog.String("type", updateType),
			slog.String("subtype", subtype),
		)
		a.hub.BroadcastUpdate(updateType, subtype, action, data)
	}
}

// handleStageProgress converts stage progress messages
func (a *MessageAdapter) handleStageProgress(stage string, data interface{}) {
	// Extract progress data
	var progress float64
	var message string
	var current, total int
	
	if m, ok := data.(map[string]interface{}); ok {
		if p, ok := m["progress"].(int); ok {
			progress = float64(p)
		} else if p, ok := m["progress"].(float64); ok {
			progress = p
		}
		
		if msg, ok := m["message"].(string); ok {
			message = msg
		}
		
		if c, ok := m["current"].(int); ok {
			current = c
		}
		if t, ok := m["total"].(int); ok {
			total = t
		}
	}
	
	// Get pipeline ID for this stage
	pipelineID := a.getPipelineForStage(stage)
	
	// Send new format message
	a.hub.BroadcastPipelineProgress(pipelineID, stage, progress, message, current, total)
}

// handlePipelineStatus converts pipeline status messages
func (a *MessageAdapter) handlePipelineStatus(data interface{}) {
	var pipelineID, status, stage string
	var details map[string]interface{}
	
	if m, ok := data.(map[string]interface{}); ok {
		if id, ok := m["pipeline_id"].(string); ok {
			pipelineID = id
		}
		if s, ok := m["status"].(string); ok {
			status = s
		}
		if st, ok := m["stage"].(string); ok {
			stage = st
		}
		
		// Store stage for pipeline tracking
		if pipelineID != "" && stage != "" {
			a.setPipelineStage(pipelineID, stage)
		}
		
		// Copy other fields as details
		details = make(map[string]interface{})
		for k, v := range m {
			if k != "pipeline_id" && k != "status" && k != "stage" {
				details[k] = v
			}
		}
	}
	
	a.hub.BroadcastPipelineStatus(pipelineID, status, stage, details)
}

// handleOutput converts output messages
func (a *MessageAdapter) handleOutput(data interface{}) {
	var level, message, stage string
	
	if m, ok := data.(map[string]interface{}); ok {
		if l, ok := m["level"].(string); ok {
			level = l
		} else {
			level = LevelInfo
		}
		
		if msg, ok := m["message"].(string); ok {
			message = msg
		}
		
		if s, ok := m["stage"].(string); ok {
			stage = s
		}
	}
	
	a.hub.BroadcastOutput(level, message, stage)
}

// handleError converts error messages
func (a *MessageAdapter) handleError(data interface{}) {
	var errorCode, title, detail, stage string
	var metadata map[string]interface{}
	
	if m, ok := data.(map[string]interface{}); ok {
		if code, ok := m["code"].(string); ok {
			errorCode = code
		} else if code, ok := m["error_code"].(string); ok {
			errorCode = code
		} else {
			errorCode = "ERR_UNKNOWN"
		}
		
		if t, ok := m["title"].(string); ok {
			title = t
		} else if msg, ok := m["message"].(string); ok {
			title = msg
		} else {
			title = "Error"
		}
		
		if d, ok := m["detail"].(string); ok {
			detail = d
		} else if d, ok := m["details"].(string); ok {
			detail = d
		}
		
		if s, ok := m["stage"].(string); ok {
			stage = s
		}
		
		// Copy other fields as metadata
		metadata = make(map[string]interface{})
		for k, v := range m {
			if k != "code" && k != "error_code" && k != "title" && 
			   k != "message" && k != "detail" && k != "details" && k != "stage" {
				metadata[k] = v
			}
		}
	}
	
	a.hub.BroadcastError(errorCode, title, detail, stage, metadata)
}

// Pipeline tracking helpers

func (a *MessageAdapter) setPipelineStage(pipelineID, stage string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.activePipelines[pipelineID] = stage
}

func (a *MessageAdapter) getPipelineForStage(stage string) string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	// Find pipeline with matching stage
	for pipelineID, currentStage := range a.activePipelines {
		if currentStage == stage {
			return pipelineID
		}
	}
	
	// Return empty if not found (legacy compatibility)
	return ""
}

// Legacy compatibility methods (implement old Hub interface)

// BroadcastProgress implements old progress broadcast
func (a *MessageAdapter) BroadcastProgress(stage string, progress int, message string) {
	a.handleStageProgress(stage, map[string]interface{}{
		"progress": progress,
		"message":  message,
	})
}

// BroadcastStatus implements old status broadcast
func (a *MessageAdapter) BroadcastStatus(status, message string) {
	a.hub.BroadcastPipelineStatus("", status, "", map[string]interface{}{
		"message": message,
	})
}

// BroadcastOutput implements old output broadcast
func (a *MessageAdapter) BroadcastOutput(message, level string) {
	a.hub.BroadcastOutput(level, message, "")
}

// BroadcastError implements old error broadcast
func (a *MessageAdapter) BroadcastError(code, message, details, stage string, recoverable bool) {
	metadata := map[string]interface{}{
		"recoverable": recoverable,
	}
	
	if details != "" {
		metadata["details"] = details
	}
	
	a.hub.BroadcastError(code, message, details, stage, metadata)
}

// BroadcastJSON implements old JSON broadcast
func (a *MessageAdapter) BroadcastJSON(message map[string]interface{}) {
	a.hub.BroadcastJSON(message)
}

// Broadcast implements services.WebSocketHub interface
func (a *MessageAdapter) Broadcast(messageType string, data interface{}) {
	a.hub.Broadcast(messageType, data)
}

// ClientCount returns the number of connected clients
func (a *MessageAdapter) ClientCount() int {
	return a.hub.ClientCount()
}

// Register adds a client to the hub
func (a *MessageAdapter) Register(client *Client) {
	a.hub.Register(client)
}