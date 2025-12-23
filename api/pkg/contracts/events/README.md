# Event Contracts

WebSocket and event message contracts for real-time communication.

## Message Types

All WebSocket messages follow this structure:

```go
type WebSocketMessage struct {
    Type      MessageType     `json:"type"`
    Timestamp time.Time       `json:"timestamp"`
    TraceID   string          `json:"trace_id,omitempty"`
    Data      json.RawMessage `json:"data"`
    Channel   ChannelType     `json:"channel,omitempty"`
}
```

## Event Types

### Operations Events
- **operation:start** - Operation initiated
- **operation:progress** - Progress update with percentage
- **operation:complete** - Operation finished successfully
- **operation:failed** - Operation encountered error
- **operation:cancelled** - Operation cancelled by user

### operation Events (Legacy - use Operations)
- **operation:start** - operation execution started
- **operation:progress** - step progress update
- **operation:complete** - operation finished
- **operation:error** - operation error occurred

### Data Events
- **data:update** - New data available
- **data:processed** - Data processing complete
- **data:error** - Data processing error
- **report:ready** - Report generation complete

### System Events
- **system:status** - System health status
- **system:alert** - System alert or warning
- **system:metrics** - Performance metrics update
- **connection:established** - WebSocket connected
- **connection:closed** - WebSocket disconnected

### User Events
- **auth:required** - Authentication needed
- **auth:success** - Authentication successful
- **auth:failed** - Authentication failed
- **license:expired** - License has expired

## Message Type Constants

```go
const (
    // Operations events
    MessageTypeOperationStart    MessageType = "operation:start"
    MessageTypeOperationProgress MessageType = "operation:progress"
    MessageTypeOperationComplete MessageType = "operation:complete"
    MessageTypeOperationFailed   MessageType = "operation:failed"
    
    // System events
    MessageTypeSystemStatus  MessageType = "system:status"
    MessageTypeSystemAlert   MessageType = "system:alert"
    MessageTypeSystemMetrics MessageType = "system:metrics"
)
```

## Channel Types

Messages can be broadcast to specific channels:

```go
const (
    ChannelTypeGlobal     ChannelType = "global"
    ChannelTypeOperations ChannelType = "operations"
    ChannelTypeData       ChannelType = "data"
    ChannelTypeSystem     ChannelType = "system"
)
```

## Event Payloads

### Operation Progress Event
```go
type OperationProgressData struct {
    OperationID string  `json:"operation_id"`
    Progress    float64 `json:"progress"`
    CurrentStep string  `json:"current_step"`
    TotalSteps  int     `json:"total_steps"`
    Message     string  `json:"message,omitempty"`
}
```

### System Status Event
```go
type SystemStatusData struct {
    Status      string            `json:"status"`
    Uptime      int64             `json:"uptime"`
    ActiveOps   int               `json:"active_operations"`
    Metrics     map[string]float64 `json:"metrics"`
}
```

### Error Event
```go
type ErrorData struct {
    Code        string `json:"code"`
    Message     string `json:"message"`
    Details     string `json:"details,omitempty"`
    Recoverable bool   `json:"recoverable"`
    Hint        string `json:"hint,omitempty"`
}
```

## Guidelines

1. **Trace Correlation**: All events must include trace_id for request correlation
2. **Consistent Naming**: Use `<domain>:<action>` format (e.g., "operation:progress")
3. **UTC Timestamps**: Always use UTC for timestamp field
4. **Payload Size**: Keep payloads under 64KB for WebSocket efficiency
5. **Type Safety**: Use MessageType constants, not string literals
6. **Channel Scoping**: Use appropriate channel for targeted broadcasting
7. **Error Context**: Include recovery hints in error events

## Backward Compatibility

For migration from operation to operations terminology:
```go
// Legacy aliases (deprecated)
const (
    MessageTypePipelineStart    = MessageTypeOperationStart
    MessageTypePipelineProgress = MessageTypeOperationProgress
    MessageTypePipelineComplete = MessageTypeOperationComplete
)
```

## Change Log
- 2025-07-30: Added operations event types for Phase 4 implementation
- 2025-07-30: Enhanced event payloads with structured data types
- 2025-07-30: Added channel types for scoped broadcasting
- 2025-07-30: Updated documentation with comprehensive examples
- 2025-01-30: Renamed operation events to operations events
- 2025-01-30: Added backward compatibility aliases