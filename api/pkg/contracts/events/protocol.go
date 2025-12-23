// Package events contains event contract definitions for WebSocket communication
// in the ISX Daily Reports Scrapper system.
package events

import (
	"encoding/json"
	"time"
)

// Protocol version
const (
	ProtocolVersion = "1.0"
	ProtocolName    = "isx-websocket-protocol"
)

// Connection states
type ConnectionState string

const (
	ConnectionStateConnecting   ConnectionState = "connecting"
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateDisconnecting ConnectionState = "disconnecting"
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateReconnecting ConnectionState = "reconnecting"
)

// Quality of Service levels
type QoSLevel int

const (
	QoSAtMostOnce  QoSLevel = 0 // Fire and forget
	QoSAtLeastOnce QoSLevel = 1 // Acknowledged delivery
	QoSExactlyOnce QoSLevel = 2 // Guaranteed single delivery
)

// Channel types
type ChannelType string

const (
	ChannelTypeGlobal     ChannelType = "global"
	ChannelTypeOperations ChannelType = "operations"
	ChannelTypeTicker     ChannelType = "ticker"
	ChannelTypeMarket     ChannelType = "market"
	ChannelTypeAnalytics  ChannelType = "analytics"
	ChannelTypeSystem     ChannelType = "system"
	ChannelTypeUser       ChannelType = "user"
	// Legacy channel type - deprecated
	ChannelTypePipeline   ChannelType = "operation" // Deprecated: use ChannelTypeOperations
)

// Frame represents a WebSocket protocol frame
type Frame struct {
	Version   string          `json:"version"`
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	QoS       QoSLevel        `json:"qos,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Sequence  int64           `json:"sequence,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	ReplyTo   string          `json:"reply_to,omitempty"`
	TTL       int             `json:"ttl,omitempty"` // Time to live in seconds
}

// Envelope represents a message envelope for routing
type Envelope struct {
	From      string                 `json:"from"`      // Sender identifier
	To        string                 `json:"to"`        // Recipient identifier
	MessageID string                 `json:"message_id"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Frame     Frame                  `json:"frame"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ProtocolError represents a protocol-level error
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fatal   bool   `json:"fatal"`
}

// Protocol error codes
const (
	ErrCodeInvalidFrame      = "INVALID_FRAME"
	ErrCodeInvalidChannel    = "INVALID_CHANNEL"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeRateLimited       = "RATE_LIMITED"
	ErrCodeChannelClosed     = "CHANNEL_CLOSED"
	ErrCodeMessageTooLarge   = "MESSAGE_TOO_LARGE"
	ErrCodeUnsupportedType   = "UNSUPPORTED_TYPE"
	ErrCodeProtocolViolation = "PROTOCOL_VIOLATION"
	ErrCodeTimeout           = "TIMEOUT"
	ErrCodeServerError       = "SERVER_ERROR"
)

// Handshake represents the initial handshake message
type Handshake struct {
	Version      string            `json:"version"`
	ClientID     string            `json:"client_id,omitempty"`
	Auth         *AuthData         `json:"auth,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	UserAgent    string            `json:"user_agent,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AuthData represents authentication data
type AuthData struct {
	Type        string `json:"type"` // token, api_key, certificate
	Credentials string `json:"credentials"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// HandshakeResponse represents the server's handshake response
type HandshakeResponse struct {
	Success      bool              `json:"success"`
	SessionID    string            `json:"session_id,omitempty"`
	ServerTime   time.Time         `json:"server_time"`
	Heartbeat    int               `json:"heartbeat_interval"` // seconds
	Capabilities []string          `json:"capabilities,omitempty"`
	Limits       *ConnectionLimits `json:"limits,omitempty"`
	Error        *ProtocolError    `json:"error,omitempty"`
}

// ConnectionLimits represents connection limits
type ConnectionLimits struct {
	MaxMessageSize    int64 `json:"max_message_size"`    // bytes
	MaxMessagesPerSec int   `json:"max_messages_per_sec"`
	MaxSubscriptions  int   `json:"max_subscriptions"`
	MaxQueueSize      int   `json:"max_queue_size"`
	IdleTimeout       int   `json:"idle_timeout"` // seconds
}

// ChannelInfo represents information about a channel
type ChannelInfo struct {
	Name         string              `json:"name"`
	Type         ChannelType         `json:"type"`
	Description  string              `json:"description,omitempty"`
	Public       bool                `json:"public"`
	Subscribers  int                 `json:"subscribers,omitempty"`
	MessageTypes []MessageType       `json:"message_types,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SubscriptionResult represents the result of a subscription request
type SubscriptionResult struct {
	Channel      string           `json:"channel"`
	Success      bool             `json:"success"`
	Subscribed   bool             `json:"subscribed"`
	Position     string           `json:"position,omitempty"` // For resuming
	HistoryItems []WebSocketMessage `json:"history,omitempty"`
	Error        *ProtocolError   `json:"error,omitempty"`
}

// MessageAck represents a message acknowledgment
type MessageAck struct {
	MessageID string         `json:"message_id"`
	Status    AckStatus      `json:"status"`
	Error     *ProtocolError `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// AckStatus represents acknowledgment status
type AckStatus string

const (
	AckStatusReceived  AckStatus = "received"
	AckStatusProcessed AckStatus = "processed"
	AckStatusFailed    AckStatus = "failed"
	AckStatusIgnored   AckStatus = "ignored"
)

// ReconnectInfo represents reconnection information
type ReconnectInfo struct {
	SessionID      string    `json:"session_id"`
	LastSequence   int64     `json:"last_sequence"`
	LastMessageID  string    `json:"last_message_id"`
	Subscriptions  []string  `json:"subscriptions"`
	DisconnectedAt time.Time `json:"disconnected_at"`
}

// Compression types
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGzip   CompressionType = "gzip"
	CompressionZstd   CompressionType = "zstd"
	CompressionSnappy CompressionType = "snappy"
)

// Encoding types
type EncodingType string

const (
	EncodingJSON     EncodingType = "json"
	EncodingMsgPack  EncodingType = "msgpack"
	EncodingProtobuf EncodingType = "protobuf"
)

// TransportConfig represents transport configuration
type TransportConfig struct {
	Compression      CompressionType `json:"compression"`
	Encoding         EncodingType    `json:"encoding"`
	HeartbeatInterval int            `json:"heartbeat_interval"` // seconds
	ReconnectDelay   int             `json:"reconnect_delay"`    // seconds
	MaxReconnects    int             `json:"max_reconnects"`
	BufferSize       int             `json:"buffer_size"`
}

// MetricsSnapshot represents WebSocket connection metrics
type MetricsSnapshot struct {
	SessionID        string        `json:"session_id"`
	ConnectedAt      time.Time     `json:"connected_at"`
	Duration         time.Duration `json:"duration"`
	MessagesSent     int64         `json:"messages_sent"`
	MessagesReceived int64         `json:"messages_received"`
	BytesSent        int64         `json:"bytes_sent"`
	BytesReceived    int64         `json:"bytes_received"`
	ErrorCount       int64         `json:"error_count"`
	ReconnectCount   int           `json:"reconnect_count"`
	Latency          int64         `json:"latency_ms"`
	Subscriptions    int           `json:"subscription_count"`
}

// ChannelPattern represents a channel subscription pattern
type ChannelPattern struct {
	Pattern     string                 `json:"pattern"`      // e.g., "ticker:*", "operations:123:*"
	Type        PatternType            `json:"type"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Options     *SubscriptionOptions   `json:"options,omitempty"`
}

// PatternType represents the type of channel pattern
type PatternType string

const (
	PatternTypeExact    PatternType = "exact"    // Exact match
	PatternTypePrefix   PatternType = "prefix"   // Prefix match
	PatternTypeWildcard PatternType = "wildcard" // Wildcard match
	PatternTypeRegex    PatternType = "regex"    // Regular expression
)

// BatchMessage represents a batch of messages
type BatchMessage struct {
	BaseMessage
	Messages []WebSocketMessage `json:"messages"`
	Count    int                `json:"count"`
}

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Sequence  int64     `json:"sequence"`
	Latency   int64     `json:"latency_ms,omitempty"`
}

// StreamControl represents stream control commands
type StreamControl struct {
	Command   StreamCommand          `json:"command"`
	Channel   string                 `json:"channel,omitempty"`
	Position  string                 `json:"position,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// StreamCommand represents stream control commands
type StreamCommand string

const (
	StreamCommandPause    StreamCommand = "pause"
	StreamCommandResume   StreamCommand = "resume"
	StreamCommandRewind   StreamCommand = "rewind"
	StreamCommandFastForward StreamCommand = "fast_forward"
	StreamCommandSeek     StreamCommand = "seek"
)