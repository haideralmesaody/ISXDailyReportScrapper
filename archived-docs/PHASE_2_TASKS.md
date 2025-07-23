# Phase 2: Code Refactoring - Detailed Task List

## Overview
This document breaks down Phase 2 of the July Development Plan into manageable tasks. Each task is designed to be completed in 1-2 hours and maintains backward compatibility.

## Task Categories

### A. File Management Utilities (Priority: HIGH)
These utilities are used across multiple components and should be refactored first.

#### Task A1: Create internal/files package
- **Duration**: 1 hour
- **Description**: Create file discovery and management utilities
- **Files to create**:
  - `dev/internal/files/discovery.go` - Find files by pattern, date range
  - `dev/internal/files/manager.go` - File operations (exists, copy, move)
- **Extract from**:
  - `data-processor.go`: findExcelFiles(), findReportFiles()
  - `index-extractor.go`: file discovery logic
  - `web-application.go`: handleListFiles()

#### Task A2: Create internal/exporter package for CSV operations
- **Duration**: 1.5 hours
- **Description**: Centralize all CSV writing operations
- **Files to create**:
  - `dev/internal/exporter/csv.go` - Generic CSV writing utilities
  - `dev/internal/exporter/daily.go` - Daily report generation
  - `dev/internal/exporter/ticker.go` - Ticker-specific exports
- **Extract from**:
  - `data-processor.go`: writeCSV(), writeDailyCSVGrouped()
  - `index-extractor.go`: CSV writing logic

### B. Processing Algorithms (Priority: HIGH)
Core business logic that should be testable in isolation.

#### Task B1: Extract forward-fill algorithm
- **Duration**: 1 hour
- **Description**: Move forward-fill logic to dedicated package
- **Files to create**:
  - `dev/internal/processor/forwardfill.go` - Forward-fill implementation
  - `dev/internal/processor/types.go` - Data structures
- **Extract from**:
  - `data-processor.go`: forwardFillMissingData()

#### Task B2: Extract ticker summary generation
- **Duration**: 1 hour
- **Description**: Move analytics logic to dedicated package
- **Files to create**:
  - `dev/internal/analytics/summary.go` - Ticker summary generation
  - `dev/internal/analytics/types.go` - Summary data structures
- **Extract from**:
  - `data-processor.go`: generateTickerSummary()
  - `web-application.go`: generateTickerSummary() calls

### C. Web Scraping Components (Priority: MEDIUM)
Separate browser automation from business logic.

#### Task C1: Create internal/scraper package structure
- **Duration**: 1 hour
- **Description**: Initialize scraper package with types
- **Files to create**:
  - `dev/internal/scraper/types.go` - Config and data structures
  - `dev/internal/scraper/progress.go` - Progress tracking

#### Task C2: Extract browser automation
- **Duration**: 2 hours
- **Description**: Move ChromeDP logic to dedicated module
- **Files to create**:
  - `dev/internal/scraper/browser.go` - Browser initialization and control
  - `dev/internal/scraper/navigation.go` - Page navigation logic
- **Extract from**:
  - `scraper.go`: ChromeDP setup and navigation

#### Task C3: Extract download management
- **Duration**: 1.5 hours
- **Description**: Separate file download logic
- **Files to create**:
  - `dev/internal/scraper/downloader.go` - HTTP download functionality
  - `dev/internal/scraper/detector.go` - Duplicate detection
- **Extract from**:
  - `scraper.go`: downloadFile(), latestDownloadedDate()

### D. Index Extraction (Priority: MEDIUM)
Consolidate market index extraction logic.

#### Task D1: Create internal/indices package
- **Duration**: 1.5 hours
- **Description**: Extract index processing logic
- **Files to create**:
  - `dev/internal/indices/extractor.go` - Index extraction from Excel
  - `dev/internal/indices/types.go` - Index data structures
- **Extract from**:
  - `index-extractor.go`: Core extraction logic

### E. Web API Organization (Priority: LOW)
Reorganize web handlers for better maintainability.

#### Task E1: Create internal/api package structure
- **Duration**: 1 hour
- **Description**: Set up API package organization
- **Files to create**:
  - `dev/internal/api/types.go` - Request/response types
  - `dev/internal/api/middleware.go` - License validation middleware

#### Task E2: Extract WebSocket handling
- **Duration**: 1.5 hours
- **Description**: Separate WebSocket logic
- **Files to create**:
  - `dev/internal/api/websocket.go` - WebSocket handler
  - `dev/internal/api/broadcast.go` - Message broadcasting
- **Extract from**:
  - `web-application.go`: WebSocket-related functions

#### Task E3: Extract API handlers
- **Duration**: 2 hours
- **Description**: Move HTTP handlers to dedicated files
- **Files to create**:
  - `dev/internal/api/handlers.go` - Main API handlers
  - `dev/internal/api/admin.go` - Admin endpoints
- **Extract from**:
  - `web-application.go`: All handler functions

### F. Command Execution (Priority: LOW)
Centralize subprocess management.

#### Task F1: Create internal/executor package
- **Duration**: 1 hour
- **Description**: Create command execution utilities
- **Files to create**:
  - `dev/internal/executor/executor.go` - Subprocess management
  - `dev/internal/executor/pipeline.go` - Pipeline orchestration
- **Extract from**:
  - `web-application.go`: executeCommand(), pipeline logic

## Implementation Guidelines

### For Each Task:
1. **Create the package directory** first
2. **Define interfaces** before implementation
3. **Move code incrementally** - start with types, then functions
4. **Update imports** in the original files
5. **Test manually** after each extraction
6. **Keep original files working** during transition

### Code Migration Pattern:
```go
// Step 1: Create new package with interface
package files

type FileManager interface {
    FindExcelFiles(dir string) ([]string, error)
    FileExists(path string) bool
}

// Step 2: Implement interface
type manager struct {
    basePath string
}

func (m *manager) FindExcelFiles(dir string) ([]string, error) {
    // Move existing implementation here
}

// Step 3: Update original file to use new package
import "isxcli/internal/files"

fm := files.NewManager(basePath)
excelFiles, err := fm.FindExcelFiles(downloadDir)
```

## Testing Strategy

### After Each Task:
1. **Build all executables** using `build.bat`
2. **Test affected functionality**:
   - Task A: File listing in web UI
   - Task B: Data processing pipeline
   - Task C: Scraping functionality
   - Task D: Index extraction
   - Task E: Web API endpoints
   - Task F: Pipeline execution

### Integration Testing:
- Run full pipeline: scraping → processing → index extraction
- Verify web UI functionality
- Check WebSocket real-time updates
- Ensure backward compatibility

## Success Criteria

Each task is complete when:
- [ ] Code is extracted to new package
- [ ] Original functionality still works
- [ ] No compilation errors
- [ ] Manual testing passes
- [ ] Code is cleaner and more focused

## Recommended Order

1. **Week 1**: Complete Tasks A1-A2, B1-B2 (File utilities and algorithms)
2. **Week 2**: Complete Tasks C1-C3, D1 (Scraping and indices)
3. **Week 3**: Complete Tasks E1-E3, F1 (Web API and execution)

## Notes
- Start with high-priority tasks that provide immediate value
- Each task builds on previous ones
- Keep backward compatibility throughout
- Document any breaking changes
- Consider adding unit tests as you refactor