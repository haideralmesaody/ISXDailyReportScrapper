# July Development Plan - ISX Daily Reports Scrapper

## Overview
This document outlines the development plan for improving the ISX Daily Reports Scrapper codebase structure and security. The plan maintains the current dev/release folder structure where:
- `dev/` contains all source code
- `release/` contains compiled binaries and runtime files
- Both folders maintain the same structure for consistency
- All file paths are relative to ensure portability

## Status Update (July 18, 2024)
✅ **Phase 1 Completed**: Critical security improvements implemented
✅ **Phase 2 Partially Completed**: Common utilities created, path handling improved
✅ **Build Process**: Consolidated to build.bat with proper flags

## Current Structure (Maintained)
```
ISXDailyReportsScrapper/
├── dev/                     # Source code
│   ├── cmd/                 # Command entry points
│   ├── internal/            # Internal packages
│   ├── web/                 # Web assets
│   └── *.go                 # Main applications
└── release/                 # Compiled binaries + runtime
    ├── *.exe                # Compiled executables
    ├── web/                 # Web assets (copied from dev/web/)
    ├── downloads/           # Downloaded Excel files
    ├── reports/             # Generated CSV reports
    └── logs/                # Application logs
```

## Phase 1: Critical Security Improvements (Week 1)

### 1.1 Remove Sensitive Data from Logs ✅ COMPLETED
- **Priority**: HIGH
- **Task**: Remove all debug logging that exposes sensitive information
- **Files modified**:
  - `dev/internal/license/manager.go` - Removed private key logging
  - `dev/cmd/web-licensed/web-application.go` - Removed license key logging
- **Implementation**:
  - Removed all sensitive `fmt.Printf` debug statements
  - Build scripts use `-ldflags "-s -w"` to strip debug info
  - Created conditional logging with ISX_DEBUG environment variable

### 1.2 Improve Path Handling ✅ COMPLETED
- **Priority**: HIGH
- **Task**: Ensure all paths are relative and work from both dev and release
- **Implementation**:
  - Created `dev/internal/common/paths.go` with GetPaths() utility
  - All paths now relative to executable location
  - Fixed file download handler to use getProjectPath()

## Phase 2: Code Organization Improvements (Week 2-3)

### 2.1 Refactor Within dev/ Folder - DETAILED BREAKDOWN
- **Priority**: MEDIUM  
- **Task**: Better organize code while maintaining dev/release structure
- **Detailed Tasks**: See PHASE_2_TASKS.md for complete breakdown

**Task Categories (15 tasks total)**:
- **A. File Management Utilities (HIGH priority, 2 tasks)**
  - A1: Create internal/files package (1 hour)
  - A2: Create internal/exporter package (1.5 hours)
- **B. Processing Algorithms (HIGH priority, 2 tasks)**
  - B1: Extract forward-fill algorithm (1 hour)
  - B2: Extract ticker summary generation (1 hour)
- **C. Web Scraping Components (MEDIUM priority, 3 tasks)**
  - C1: Create scraper package structure (1 hour)
  - C2: Extract browser automation (2 hours)
  - C3: Extract download management (1.5 hours)
- **D. Index Extraction (MEDIUM priority, 1 task)**
  - D1: Create internal/indices package (1.5 hours)
- **E. Web API Organization (LOW priority, 3 tasks)**
  - E1: Create API package structure (1 hour)
  - E2: Extract WebSocket handling (1.5 hours)
  - E3: Extract API handlers (2 hours)
- **F. Command Execution (LOW priority, 1 task)**
  - F1: Create internal/executor package (1 hour)

**Improved dev/ Structure**:
```
dev/
├── cmd/                      # Command applications
│   ├── process/             # Keep existing
│   ├── indexcsv/            # Keep existing
│   └── web-licensed/        # Keep existing
├── internal/                # Internal packages
│   ├── license/             # Keep existing
│   ├── parser/              # Keep existing
│   ├── updater/             # Keep existing
│   ├── scraper/             # NEW: Extract from scraper.go
│   │   ├── scraper.go       # Core scraping logic
│   │   ├── progress.go      # Progress reporting
│   │   └── download.go      # Download handling
│   ├── processor/           # NEW: Extract from data-processor.go
│   │   ├── processor.go     # Core processing logic
│   │   ├── forwardfill.go   # Forward-fill algorithm
│   │   └── csv.go           # CSV operations
│   ├── web/                 # NEW: Extract from web-application.go
│   │   ├── server.go        # HTTP server setup
│   │   ├── handlers.go      # Route handlers
│   │   ├── websocket.go     # WebSocket handling
│   │   └── api.go           # API endpoints
│   └── common/              # NEW: Shared utilities
│       ├── paths.go         # Path resolution utilities
│       ├── errors.go        # Error types
│       └── logger.go        # Logging utilities
├── web/                     # Static web assets (unchanged)
└── scraper.go              # Thin wrapper using internal/scraper
```

### 2.2 Create Path Resolution Utilities ✅ COMPLETED
- **Priority**: HIGH
- **Task**: Centralize path handling for dev/release compatibility
- **Implementation**: Created in `dev/internal/common/paths.go`
```go
// internal/common/paths.go
package common

import (
    "os"
    "path/filepath"
)

type Paths struct {
    ExecutableDir string
    WebDir        string
    DownloadsDir  string
    ReportsDir    string
    LogsDir       string
}

func GetPaths() (*Paths, error) {
    exe, err := os.Executable()
    if err != nil {
        return nil, err
    }
    
    exeDir := filepath.Dir(exe)
    
    return &Paths{
        ExecutableDir: exeDir,
        WebDir:        filepath.Join(exeDir, "web"),
        DownloadsDir:  filepath.Join(exeDir, "downloads"),
        ReportsDir:    filepath.Join(exeDir, "reports"),
        LogsDir:       filepath.Join(exeDir, "logs"),
    }, nil
}
```

### 2.3 Simplify Main Entry Points
- **Priority**: MEDIUM
- **Task**: Move business logic to internal packages
- **Example for scraper.go**:
```go
// dev/scraper.go - Simplified entry point
package main

import (
    "github.com/haideralmesaody/ISXDailyReportsScrapper/internal/scraper"
    "github.com/haideralmesaody/ISXDailyReportsScrapper/internal/common"
)

func main() {
    paths, err := common.GetPaths()
    if err != nil {
        log.Fatal(err)
    }
    
    app := scraper.New(paths)
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Phase 3: Build Process Improvements (Week 3)

### 3.1 Improve Build Scripts
- **Priority**: MEDIUM
- **Task**: Enhance build process while maintaining dev/release structure
- **Implementation**:
  - Update `build.bat` and `build.ps1` to use build tags
  - Add `-tags release` for production builds (excludes debug code)
  - Ensure web assets are properly copied to release/web/
  - Preserve existing data directories during build

### 3.2 Add Version Information
- **Priority**: LOW
- **Task**: Embed version information in binaries
- **Implementation**:
  - Use `-ldflags` to inject version at build time
  - Display version in application startup
  - Add version to web interface

## Phase 4: Code Quality Improvements (Week 3-4)

### 4.1 Implement Conditional Logging ✅ COMPLETED
- **Priority**: HIGH
- **Task**: Replace debug prints with proper logging
- **Implementation**: Created in `dev/internal/common/logger.go`
```go
// internal/common/logger.go
package common

import (
    "log"
    "os"
)

type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)

type Logger struct {
    level LogLevel
}

func NewLogger() *Logger {
    level := INFO
    if os.Getenv("ISX_DEBUG") == "true" {
        level = DEBUG
    }
    return &Logger{level: level}
}

func (l *Logger) Debug(format string, v ...interface{}) {
    if l.level <= DEBUG {
        log.Printf("[DEBUG] "+format, v...)
    }
}
```

### 4.2 Error Handling Improvements
- **Priority**: MEDIUM
- **Task**: Standardize error handling
- **Implementation**:
  - Create error types for different scenarios
  - Add context to errors
  - Improve user-facing error messages

### 4.3 Add Tests
- **Priority**: MEDIUM
- **Task**: Add tests for core functionality
- **Focus on**:
  - Path resolution utilities
  - Data processing logic
  - CSV/Excel operations
  - License validation (mock Google Sheets)

## Implementation Order

1. **Pre-implementation**: Git commit current state ✅
2. **Week 1**: ✅ COMPLETED
   - Remove sensitive logging (Phase 1.1) ✅
   - Create path utilities (Phase 2.2) ✅
   - Implement conditional logging (Phase 4.1) ✅
3. **Week 2-3**: IN PROGRESS
   - Refactor code organization (Phase 2.1) - See PHASE_2_TASKS.md
     - Week 2: Tasks A1-A2, B1-B2 (File utilities and algorithms)
     - Week 2: Tasks C1-C3, D1 (Scraping and indices)
   - Improve build scripts (Phase 3.1) ✅
4. **Week 3-4**:
   - Complete refactoring: Tasks E1-E3, F1 (Web API and execution)
   - Error handling improvements (Phase 4.2)
   - Add tests (Phase 4.3)
   - Add version information (Phase 3.2)

## Key Principles

1. **Maintain dev/release structure**: All changes preserve the current folder organization
2. **Relative paths only**: No absolute paths in code
3. **Backward compatibility**: Existing functionality unchanged
4. **Self-contained**: Application remains standalone with embedded credentials
5. **Security focus**: Remove all sensitive data from logs while keeping embedded credentials functional

## Success Metrics
- [x] No sensitive data exposed in logs
- [x] All paths are relative and work from both dev/ and release/
- [x] Code is better organized within existing structure (common utilities created)
- [x] Debug logging can be completely excluded from release builds
- [x] Improved error messages for users
- [ ] Basic test coverage for core functionality

## Notes
- This plan maintains the exact dev/release structure as requested
- All file references will be relative to ensure portability
- The embedded credentials approach is maintained but secured
- Focus is on organization within existing constraints

## Completion Summary (July 18, 2024)

### Completed Tasks:
1. **Security Improvements**:
   - Removed all sensitive logging from license manager and web application
   - Implemented conditional debug logging with ISX_DEBUG environment variable
   - Build process strips debug information with `-ldflags "-s -w"`

2. **Common Utilities Created**:
   - `dev/internal/common/logger.go` - Conditional logging system
   - `dev/internal/common/paths.go` - Path resolution utilities
   - `dev/internal/common/errors.go` - Common error types

3. **UI/UX Fixes**:
   - Fixed duplicate canvas variable declaration in index.html
   - Fixed pipeline status transitions (scraping → processing)
   - Fixed indices chart data loading issue
   - Added processing progress indicators

4. **Build Process**:
   - Consolidated to single `build.bat` script
   - Proper directory structure creation and data preservation
   - Release mode compilation flags

### Remaining Tasks:
- Complete code refactoring into internal packages (Phase 2.1)
- Add comprehensive test coverage (Phase 4.3)
- Implement version information in binaries (Phase 3.2)

### Next Steps:
1. Continue with Phase 2.1 - Extract business logic into internal packages
2. Add unit tests for critical components
3. Consider adding integration tests for the full pipeline