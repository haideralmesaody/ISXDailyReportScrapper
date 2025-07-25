package websocket

import (
	"encoding/json"
	"time"
)

// BaseMessage represents the common structure for all WebSocket messages
type BaseMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// OutputMessage represents console output from pipeline operations
type OutputMessage struct {
	BaseMessage
	Level   string `json:"level"`
	Message string `json:"message"`
	Stage   string `json:"stage,omitempty"`
}

// DataUpdateMessage represents data update notifications
type DataUpdateMessage struct {
	BaseMessage
	Subtype string      `json:"subtype"`
	Action  string      `json:"action"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// PipelineStatusMessage represents pipeline status updates
type PipelineStatusMessage struct {
	BaseMessage
	PipelineID string                 `json:"pipeline_id"`
	Status     string                 `json:"status"`
	Stage      string                 `json:"stage,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// PipelineProgressMessage represents pipeline progress updates
type PipelineProgressMessage struct {
	BaseMessage
	PipelineID string  `json:"pipeline_id"`
	Stage      string  `json:"stage"`
	Progress   float64 `json:"progress"` // 0-100
	Message    string  `json:"message,omitempty"`
	Current    int     `json:"current,omitempty"`
	Total      int     `json:"total,omitempty"`
}

// ErrorMessage represents error notifications with RFC 7807 alignment
type ErrorMessage struct {
	BaseMessage
	ErrorCode string                 `json:"error_code"`
	Title     string                 `json:"title"`
	Detail    string                 `json:"detail"`
	Stage     string                 `json:"stage,omitempty"`
	Hint      string                 `json:"hint,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectionMessage represents WebSocket connection status
type ConnectionMessage struct {
	BaseMessage
	Status    string `json:"status"` // connected, disconnected, reconnecting
	ClientID  string `json:"client_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// LogMessage represents structured log entries
type LogMessage struct {
	BaseMessage
	Level    string                 `json:"level"`
	Message  string                 `json:"message"`
	Source   string                 `json:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StageProgressMessage represents individual stage progress (legacy support)
// Deprecated: Use PipelineProgressMessage instead
type StageProgressMessage struct {
	Stage    string  `json:"stage"`
	Progress float64 `json:"progress"`
	Message  string  `json:"message,omitempty"`
}

// Message creation helpers

// NewOutputMessage creates a new output message
func NewOutputMessage(level, message, stage string) *OutputMessage {
	return &OutputMessage{
		BaseMessage: BaseMessage{
			Type:      TypeOutput,
			Timestamp: time.Now(),
		},
		Level:   level,
		Message: message,
		Stage:   stage,
	}
}

// NewDataUpdateMessage creates a new data update message
func NewDataUpdateMessage(subtype, action string, data interface{}) *DataUpdateMessage {
	msg := &DataUpdateMessage{
		BaseMessage: BaseMessage{
			Type:      TypeDataUpdate,
			Timestamp: time.Now(),
		},
		Subtype: subtype,
		Action:  action,
		Data:    data,
	}
	
	// Calculate count if data is countable
	if countable, ok := data.([]interface{}); ok {
		msg.Count = len(countable)
	}
	
	return msg
}

// NewPipelineStatusMessage creates a new pipeline status message
func NewPipelineStatusMessage(pipelineID, status, stage string) *PipelineStatusMessage {
	return &PipelineStatusMessage{
		BaseMessage: BaseMessage{
			Type:      TypePipelineStatus,
			Timestamp: time.Now(),
		},
		PipelineID: pipelineID,
		Status:     status,
		Stage:      stage,
	}
}

// NewPipelineProgressMessage creates a new pipeline progress message
func NewPipelineProgressMessage(pipelineID, stage string, progress float64) *PipelineProgressMessage {
	return &PipelineProgressMessage{
		BaseMessage: BaseMessage{
			Type:      TypePipelineProgress,
			Timestamp: time.Now(),
		},
		PipelineID: pipelineID,
		Stage:      stage,
		Progress:   progress,
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(errorCode, title, detail, stage string) *ErrorMessage {
	msg := &ErrorMessage{
		BaseMessage: BaseMessage{
			Type:      TypeError,
			Timestamp: time.Now(),
		},
		ErrorCode: errorCode,
		Title:     title,
		Detail:    detail,
		Stage:     stage,
	}
	
	// Add recovery hint if available
	if hint, ok := ErrorRecoveryHints[errorCode]; ok {
		msg.Hint = hint
	}
	
	return msg
}

// NewConnectionMessage creates a new connection message
func NewConnectionMessage(status, clientID string) *ConnectionMessage {
	return &ConnectionMessage{
		BaseMessage: BaseMessage{
			Type:      TypeConnection,
			Timestamp: time.Now(),
		},
		Status:   status,
		ClientID: clientID,
	}
}

// NewLogMessage creates a new log message
func NewLogMessage(level, message, source string) *LogMessage {
	return &LogMessage{
		BaseMessage: BaseMessage{
			Type:      TypeLog,
			Timestamp: time.Now(),
		},
		Level:   level,
		Message: message,
		Source:  source,
	}
}

// Message marshaling helpers

// MarshalJSON ensures consistent JSON formatting
func (m BaseMessage) MarshalJSON() ([]byte, error) {
	type Alias BaseMessage
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: m.Timestamp.Format(time.RFC3339),
		Alias:     (*Alias)(&m),
	})
}

// ToJSON converts any message to JSON bytes
func ToJSON(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

// Legacy message support for backward compatibility

// LegacyMessage represents the old message format
type LegacyMessage struct {
	Type    string      `json:"type"`
	Stage   string      `json:"stage,omitempty"`
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ConvertFromLegacy converts legacy message format to new format
func ConvertFromLegacy(legacy *LegacyMessage) interface{} {
	switch legacy.Type {
	case "stage_progress":
		// Convert to PipelineProgressMessage
		return &PipelineProgressMessage{
			BaseMessage: BaseMessage{
				Type:      TypePipelineProgress,
				Timestamp: time.Now(),
			},
			Stage:   legacy.Stage,
			Message: legacy.Message,
		}
	case "pipeline_status":
		return &PipelineStatusMessage{
			BaseMessage: BaseMessage{
				Type:      TypePipelineStatus,
				Timestamp: time.Now(),
			},
			Status: legacy.Status,
			Stage:  legacy.Stage,
		}
	case "error":
		return &ErrorMessage{
			BaseMessage: BaseMessage{
				Type:      TypeError,
				Timestamp: time.Now(),
			},
			Title:  "Error",
			Detail: legacy.Message,
			Stage:  legacy.Stage,
		}
	default:
		// Return as output message
		return &OutputMessage{
			BaseMessage: BaseMessage{
				Type:      TypeOutput,
				Timestamp: time.Now(),
			},
			Level:   LevelInfo,
			Message: legacy.Message,
			Stage:   legacy.Stage,
		}
	}
}