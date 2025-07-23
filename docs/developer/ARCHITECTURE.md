# System Architecture

## Overview

The ISX Daily Reports Scrapper is a modular Go application designed for automated data collection, processing, and visualization of Iraq Stock Exchange trading data.

## Architecture Diagram

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   ISX Website   │────▶│  Web Scraper     │────▶│  Excel Files    │
│                 │     │  (scraper.exe)   │     │  (.xlsx)        │
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                           │
                        ┌──────────────────┐               │
                        │ Data Processor   │◀──────────────┘
                        │ (process.exe)    │
                        └────────┬─────────┘
                                 │
                    ┌────────────┴────────────┐
                    ▼                         ▼
            ┌───────────────┐         ┌───────────────┐
            │ CSV Reports   │         │ Index Values  │
            │               │         │ (indexcsv.exe)│
            └───────┬───────┘         └───────┬───────┘
                    │                         │
                    └────────┬────────────────┘
                             ▼
                    ┌─────────────────┐      ┌─────────────────┐
                    │ File Watcher    │─────▶│ WebSocket Hub   │
                    │                 │      │                 │
                    └─────────────────┘      └────────┬────────┘
                                                      │
                    ┌─────────────────┐               │
                    │ Web Interface   │◀──────────────┘
                    │ (web-licensed)  │
                    └─────────────────┘
```

## Core Components

### 1. Web Scraper (`cmd/scraper/`)
**Purpose**: Automated downloading of Excel reports from ISX website
- Uses Chromedp for browser automation
- Configurable concurrent workers
- Handles authentication and navigation
- Downloads to `data/downloads/`

### 2. Data Processor (`cmd/process/`)
**Purpose**: Converts Excel files to structured CSV format
- Parses Excel using excelize library
- Applies data transformations
- Handles forward-filling for non-trading days
- Generates multiple output formats

### 3. Index Extractor (`cmd/indexes/`)
**Purpose**: Extracts market index values from Excel files
- Specifically handles ISX60 and ISX15 indices
- Creates time-series data for indices
- Outputs to `indexes.csv`

### 4. Web Application (`cmd/web-licensed/`)
**Purpose**: User interface and API server
- Serves web interface on port 8080
- Provides REST API endpoints
- Manages WebSocket connections
- Handles license validation

## Internal Packages

### Data Processing
- `internal/parser` - Excel file parsing logic
- `internal/processor` - Data transformation and forward-filling
- `internal/analytics` - Statistical analysis and summaries
- `internal/exporter` - CSV and JSON export utilities

### Infrastructure
- `internal/websocket` - Real-time communication hub
- `internal/files` - File management utilities
- `internal/license` - License validation system
- `internal/pipeline` - Pipeline orchestration manager and stage framework
- `internal/progress` - Progress tracking with historical metrics and WebSocket message utilities

### Web Scraping
- `internal/common` - Shared utilities (logger, paths, errors)

## Data Flow

1. **Collection Phase**
   - Scraper downloads Excel files from ISX
   - Files stored in `data/downloads/`
   - Named by date: `YYYY-MM-DD.xlsx`

2. **Processing Phase**
   - Processor reads Excel files
   - Transforms to internal TradeRecord structure
   - Applies forward-filling for gaps
   - Generates CSV outputs

3. **Analysis Phase**
   - Analytics module generates summaries
   - Creates ticker statistics
   - Calculates trading day counts

4. **Distribution Phase**
   - File watcher detects changes
   - WebSocket broadcasts updates
   - Web UI refreshes automatically

### Metrics Storage
Historical performance data is stored in JSON format for ETA calculations:
- `data/metrics/scraping_metrics.json` - Download timing history
- `data/metrics/processing_metrics.json` - Processing timing history
- `data/metrics/indices_metrics.json` - Index extraction timing history

Each metrics file contains timing data that improves ETA accuracy over time.

## Pipeline Manager Architecture

### Overview
The Pipeline Manager (`internal/pipeline`) provides centralized orchestration of the data processing pipeline, replacing the previous complex conditional logic with a clean, stage-based architecture.

### Key Components

1. **Pipeline Manager**
   - Central orchestrator for all pipeline executions
   - Manages pipeline state and stage transitions
   - Handles stage dependencies and execution order
   - Provides retry and error recovery mechanisms

2. **Stage Interface**
   ```go
   type Stage interface {
       ID() string
       Name() string
       Dependencies() []string
       Execute(ctx context.Context, state *PipelineState) error
       Validate(state *PipelineState) error
   }
   ```

3. **Stage Implementations** (`cmd/web-licensed/stages/`)
   - `ScrapingStage` - Wraps scraper.exe
   - `ProcessingStage` - Wraps process.exe
   - `IndicesStage` - Wraps indexcsv.exe
   - `AnalysisStage` - Performs ticker analysis

4. **WebSocket Adapter**
   - Bridges pipeline events to existing WebSocket hub
   - Maintains backward compatibility with frontend
   - Provides real-time progress updates

5. **Enhanced Progress Tracking** (`internal/progress`)
   - **EnhancedCalculator**: Progress calculation with historical metrics
   - **MetricsManager**: Persists timing data for future ETA predictions
   - **Message Utilities**: Standardized WebSocket message creation
   - **Historical Learning**: Improves ETA accuracy over time
   - **Dynamic Adjustment**: Updates progress based on actual findings

### Execution Flow
```
PipelineRequest → Manager.Execute()
    ↓
Validate all stages
    ↓
Execute stages in dependency order
    ↓
Monitor progress and update state
    ↓
Return PipelineResponse
```

### Benefits
- Automatic stage progression based on dependencies
- Centralized error handling and recovery
- Consistent progress tracking across all stages
- Easy to add new stages or modify execution flow
- Eliminates race conditions and timing issues

## Technology Stack

- **Language**: Go 1.23+
- **Web Framework**: Gorilla Mux
- **WebSocket**: Gorilla WebSocket
- **Excel Processing**: Excelize v2
- **Web Scraping**: Chromedp
- **File Watching**: fsnotify
- **Frontend**: Vanilla JavaScript, Bootstrap 5, Chart.js

## Security Considerations

- License validation on startup
- Machine ID binding for licenses
- No sensitive data in logs
- Input validation on all endpoints
- CORS configured for local access only

## Performance Optimizations

- Concurrent scraping with worker pools
- Streaming CSV writers for large datasets
- In-memory caching for ticker summaries
- Debounced file watching (100ms)
- Efficient Excel parsing with selective sheet reading

## Deployment Architecture

### Single Machine Deployment
All components run on a single Windows machine:
```
web-licensed.exe (port 8080)
  ├── File Watcher (monitors data/reports/)
  ├── WebSocket Server
  └── HTTP API Server
```

### Data Storage
```
release/
├── data/
│   ├── downloads/     # Raw Excel files
│   └── reports/       # Processed CSV/JSON
├── logs/              # Application logs
└── license.dat        # License file
```

## Future Scalability

The architecture supports future enhancements:
- Database backend (replace CSV storage)
- Distributed processing (message queue)
- Cloud deployment (containerization)
- Multi-user support (authentication)
- API rate limiting and caching