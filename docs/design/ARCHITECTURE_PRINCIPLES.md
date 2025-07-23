# Architecture Principles

## Overview
This document defines the core architectural principles for the ISX Daily Reports Scrapper project. All development must follow these patterns to ensure maintainability, reliability, and clear separation of concerns.

## Three-Layer Architecture

### 1. Frontend Layer (HTML/JavaScript)
**Purpose**: User interface for input collection and status display

**Responsibilities**:
- Collect user inputs (date ranges, options, modes)
- Display status updates received via WebSocket
- Send user commands to backend via HTTP API
- Update UI elements based on backend messages

**Strict Rules**:
- NO business logic or decision making
- NO direct control of pipeline flow
- NO data validation beyond basic UI constraints
- NO direct execution of backend processes

**Example - Correct Pattern**:
```javascript
// CORRECT: Send user input to backend
async function startScraping() {
    const dates = {
        from: document.getElementById('fromDate').value,
        to: document.getElementById('toDate').value
    };
    
    // Send to backend for processing
    const response = await fetch('/api/scrape', {
        method: 'POST',
        body: JSON.stringify(dates)
    });
    
    // Backend handles all logic and decisions
}

// CORRECT: Display status from backend
websocket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'pipeline_status') {
        updateStatusDisplay(message.data);
    }
};
```

**Example - Incorrect Pattern**:
```javascript
// WRONG: Frontend making decisions
websocket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'pipeline_status' && message.data.status === 'completed') {
        // WRONG: Frontend should not control pipeline
        startNextStage(); 
    }
};
```

### 2. Backend Layer (Go)
**Purpose**: All business logic, processing, and decision making

**Responsibilities**:
- Receive and validate user inputs
- Execute all business logic
- Manage pipeline stages and transitions
- Control process execution
- Send status updates via WebSocket
- Handle all error conditions and recovery

**Strict Rules**:
- ALL logic must be in Go code
- Pipeline transitions happen sequentially in backend
- Process success determined by exit codes, not WebSocket
- WebSocket used only for status broadcasting

**Example - Correct Pattern**:
```go
// CORRECT: Backend controls entire pipeline
func handleScrape(w http.ResponseWriter, r *http.Request) {
    // 1. Validate input
    dates := validateDates(r.Body)
    
    // 2. Execute scraper
    scraperCmd := exec.Command("scraper.exe", args...)
    err := scraperCmd.Run()
    
    // 3. Check success via exit code
    if err == nil {
        // 4. Backend decides next step
        broadcastStatus("scraping", "completed", "Scraping finished")
        
        // 5. Start next stage automatically
        processCmd := exec.Command("process.exe")
        err = processCmd.Run()
        
        if err == nil {
            broadcastStatus("processing", "completed", "Processing finished")
            // Continue pipeline...
        }
    }
}
```

**Example - Incorrect Pattern**:
```go
// WRONG: Using WebSocket for control flow
func handleWebSocketMessage(msg Message) {
    if msg.Type == "stage_completed" {
        // WRONG: Pipeline control via WebSocket
        startNextStage(msg.Stage)
    }
}
```

### 3. Communication Layer

#### HTTP API (Frontend → Backend)
**Purpose**: User commands and inputs

**Pattern**:
- Frontend sends user actions via HTTP POST/GET
- Backend validates and processes
- Backend returns immediate response
- Long-running tasks tracked via WebSocket

**Examples**:
- `POST /api/scrape` - Start scraping with date range
- `POST /api/process` - Start processing with options
- `GET /api/status` - Get current pipeline status

#### WebSocket (Backend → Frontend)
**Purpose**: Real-time status updates only

**Strict Rules**:
- One-way communication (Backend to Frontend)
- Never used for control flow
- Only status, progress, and informational messages
- Frontend only displays, never acts on messages

**Message Types**:
- `pipeline_status` - Current stage status
- `progress` - Progress within a stage
- `output` - Log messages for display
- `error` - Error information for display

## Data Flow Patterns

### User Action Flow
```
User Input → Frontend → HTTP API → Backend Validation → Backend Processing
                                          ↓
                                    WebSocket Status → Frontend Display
```

### Pipeline Execution Flow
```
Backend Start Stage → Execute Process → Check Exit Code → Decide Next Action
        ↓                    ↓                               ↓
   Send Status          Send Progress                   Start Next Stage
        ↓                    ↓                               ↓
   WebSocket            WebSocket                       (Internal Logic)
        ↓                    ↓
Frontend Display    Frontend Display
```

## Anti-Patterns to Avoid

### 1. WebSocket Control Flow
**Wrong**:
```go
// Executable sends completion
sendWebSocketMessage("completed")

// Web app waits for WebSocket to decide
if websocketMsg == "completed" {
    startNextStage()
}
```

**Right**:
```go
// Executable returns exit code
os.Exit(0) // success

// Web app checks exit code
if cmd.Wait() == nil {
    startNextStage()
}
```

### 2. Frontend Logic
**Wrong**:
```javascript
// Frontend decides when to refresh
if (progressPercent >= 100) {
    refreshData();
}
```

**Right**:
```javascript
// Backend tells frontend when to refresh
if (message.type === 'refresh') {
    refreshData();
}
```

### 3. Circular Dependencies
**Wrong**:
- Frontend waits for WebSocket to enable buttons
- Backend waits for frontend to trigger next stage
- WebSocket messages trigger backend actions

**Right**:
- Backend controls everything
- Frontend only reflects backend state
- WebSocket only carries status updates

## Implementation Guidelines

### Adding a New Pipeline Stage

1. **Define Stage in Backend**:
```go
// In handleScrape or similar
func runPipeline() {
    // Stage 1
    if err := runScraper(); err != nil {
        handleError(err)
        return
    }
    
    // Stage 2
    if err := runProcessor(); err != nil {
        handleError(err)
        return
    }
    
    // New Stage 3
    if err := runNewStage(); err != nil {
        handleError(err)
        return
    }
}
```

2. **Create Status Updates**:
```go
func runNewStage() error {
    broadcastStatus("newstage", "active", "Starting new stage...")
    
    cmd := exec.Command("newstage.exe")
    err := cmd.Run()
    
    if err != nil {
        broadcastStatus("newstage", "error", "New stage failed")
        return err
    }
    
    broadcastStatus("newstage", "completed", "New stage completed")
    return nil
}
```

3. **Update Frontend Display**:
```javascript
// Only display logic
const stageElements = {
    'newstage': document.getElementById('newStageStatus')
};

// In WebSocket handler
if (message.data.stage === 'newstage') {
    updateStageDisplay(message.data);
}
```

### Error Handling

**Backend Responsibility**:
- Detect all error conditions
- Decide on recovery strategy
- Control retry logic
- Send appropriate status to frontend

**Frontend Responsibility**:
- Display error messages
- Show recovery hints to user
- Enable/disable UI based on backend state

## Testing Architecture Compliance

### Checklist for Code Review
- [ ] All business logic is in Go code?
- [ ] Frontend only has display logic?
- [ ] WebSocket never used for control flow?
- [ ] Pipeline transitions happen in backend?
- [ ] Exit codes determine process success?
- [ ] No circular dependencies?
- [ ] Status updates are display-only?

### Common Mistakes
1. Waiting for WebSocket messages to proceed
2. Frontend triggering next pipeline stage
3. Using WebSocket completion instead of exit codes
4. Business logic in JavaScript
5. Validation in frontend beyond UI constraints

## Summary

The architecture ensures:
- **Reliability**: No timing issues or race conditions
- **Maintainability**: Clear separation of concerns
- **Testability**: Backend logic can be tested independently
- **Scalability**: Easy to add new stages or features

Remember: 
- **Frontend** = Display and Input
- **Backend** = All Logic and Control
- **WebSocket** = Status Updates Only