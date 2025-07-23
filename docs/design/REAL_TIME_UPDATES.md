# Real-Time Data Update Design Pattern

## Overview

This document defines a standardized pattern for propagating data changes from CSV files and backend processes to the frontend automatically. This pattern ensures that users always see the latest data without manual refresh.

## Architecture

### 1. Event-Driven Architecture

```
CSV Files → File Watcher → Backend Process → WebSocket → Frontend UI
     ↓                           ↓                 ↓
File Change Event          Process Event      UI Update Event
```

### 2. Core Components

#### A. File System Watcher (Backend)
- Monitor changes to CSV files in `data/reports/` directory
- Emit events when files are created, modified, or deleted
- Filter relevant file types (*.csv, *.json)

#### B. Data Processing Pipeline
- Process changed files
- Update in-memory caches
- Emit data update events

#### C. WebSocket Communication Layer
- Maintain persistent connection between backend and frontend
- Push real-time updates to connected clients
- Handle reconnection and error recovery

#### D. Frontend Update Manager
- Listen for WebSocket messages
- Update UI components without page refresh
- Manage local state synchronization

## Implementation Pattern

### 1. Backend Implementation

```go
// Event types for data changes
type DataChangeEvent struct {
    Type      string    // "file_created", "file_updated", "data_processed"
    Timestamp time.Time
    Data      interface{}
}

// File watcher service
type FileWatcherService struct {
    watchPath string
    events    chan DataChangeEvent
}

func (fw *FileWatcherService) Start() {
    // Watch for file changes
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(fw.watchPath)
    
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                fw.events <- DataChangeEvent{
                    Type:      "file_updated",
                    Timestamp: time.Now(),
                    Data:      event.Name,
                }
            }
        }
    }
}

// WebSocket hub for broadcasting updates
type UpdateHub struct {
    clients    map[*Client]bool
    broadcast  chan DataChangeEvent
    register   chan *Client
    unregister chan *Client
}

func (h *UpdateHub) run() {
    for {
        select {
        case event := <-h.broadcast:
            // Send to all connected clients
            for client := range h.clients {
                client.send <- event
            }
        }
    }
}
```

### 2. WebSocket Message Protocol

```javascript
// Standard message format
{
    "type": "data_update",
    "subtype": "ticker_summary|daily_report|combined_data|index_data",
    "action": "created|updated|deleted",
    "data": {
        // Specific data payload
    },
    "timestamp": "2025-07-18T10:30:00Z"
}
```

### 3. Frontend Implementation

```javascript
class DataUpdateManager {
    constructor() {
        this.ws = null;
        this.handlers = new Map();
        this.reconnectInterval = 5000;
        this.connect();
    }
    
    connect() {
        this.ws = new WebSocket('ws://localhost:8080/ws');
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleUpdate(message);
        };
        
        this.ws.onclose = () => {
            // Auto-reconnect
            setTimeout(() => this.connect(), this.reconnectInterval);
        };
    }
    
    handleUpdate(message) {
        if (message.type === 'data_update') {
            const handler = this.handlers.get(message.subtype);
            if (handler) {
                handler(message.data);
            }
        }
    }
    
    // Register update handlers
    onTickerUpdate(callback) {
        this.handlers.set('ticker_summary', callback);
    }
    
    onDailyReportUpdate(callback) {
        this.handlers.set('daily_report', callback);
    }
}

// Usage in UI components
const updateManager = new DataUpdateManager();

// Register handlers for different data types
updateManager.onTickerUpdate((data) => {
    // Update ticker list UI
    refreshTickerList(data);
});

updateManager.onDailyReportUpdate((data) => {
    // Update reports table
    addNewReportRow(data);
});
```

## Implementation Guidelines

### 1. Data Change Detection Points

All processes that modify data must emit update events:

- **CSV File Writers**: After writing CSV files
- **JSON Generators**: After creating/updating JSON files
- **Data Processors**: After completing processing tasks
- **Scrapers**: After downloading new data

### 2. Update Granularity

- **Full Refresh**: For major data structure changes
- **Incremental Updates**: For adding new records
- **Partial Updates**: For modifying specific records

### 3. Performance Considerations

- Debounce rapid file changes (wait 100ms before processing)
- Batch multiple updates within a time window
- Send only changed data, not entire datasets
- Use compression for large payloads

### 4. Error Handling

- Graceful degradation when WebSocket unavailable
- Queue updates during disconnection
- Provide manual refresh fallback
- Log all update failures

## Integration Points

### 1. Existing Code Modifications

#### A. data-processor.go
```go
// After saving combined CSV
if err := saveCombinedCSV(combinedFile, allRecords); err != nil {
    return err
}
// Emit update event
hub.Broadcast(DataChangeEvent{
    Type:    "data_update",
    Subtype: "combined_data",
    Action:  "updated",
})
```

#### B. summary.go
```go
// After generating ticker summary
if err := s.writeSummaryJSON(jsonFile, summaries); err != nil {
    return err
}
// Emit update event
hub.Broadcast(DataChangeEvent{
    Type:    "data_update",
    Subtype: "ticker_summary",
    Action:  "updated",
    Data:    summaries,
})
```

#### C. index.html
```javascript
// Replace manual refresh with auto-update
updateManager.onTickerUpdate(() => {
    // Existing refreshTickers() logic
    loadTickerSummary();
});

// Remove setTimeout refresh calls
// Use event-driven updates instead
```

### 2. New Components to Add

1. **File Watcher Service**: Monitor data directory
2. **WebSocket Server**: Handle real-time connections
3. **Update Hub**: Coordinate message broadcasting
4. **Client Update Manager**: Handle frontend updates

## Benefits

1. **Real-Time Updates**: Users see changes immediately
2. **Reduced Server Load**: No polling required
3. **Better UX**: No manual refresh needed
4. **Scalability**: Efficient for multiple clients
5. **Consistency**: All clients see same data state

## About BOM (Byte Order Mark)

The BOM is a special character sequence at the beginning of text files:
- **UTF-8 BOM**: `0xEF 0xBB 0xBF` (3 bytes)
- **Purpose**: Indicates file encoding and byte order
- **Issue**: Can cause parsing problems if not handled
- **Excel Compatibility**: Excel expects BOM for UTF-8 CSV files

In our code:
- We ADD BOM when writing CSV files (for Excel)
- We REMOVE BOM when reading CSV files (for parsing)

Example:
```go
// Writing with BOM
file.Write([]byte{0xEF, 0xBB, 0xBF}) // Add BOM
writer.Write(headers)                  // Then write data

// Reading with BOM removal
if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
    content = content[3:] // Skip BOM bytes
}
```

## Next Steps

1. Implement file watcher service
2. Add WebSocket server to web application
3. Create update hub for message broadcasting
4. Modify all data writers to emit events
5. Update frontend to use event-driven updates
6. Remove manual refresh code
7. Test real-time synchronization