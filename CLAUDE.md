# ISX Daily Reports Scrapper - Complete Project Guide

## Project Overview
ISX Daily Reports Scrapper is a licensed web application that scrapes daily reports from the Iraq Stock Exchange (ISX) website, processes the data, and provides analytics through a web interface.

## Quick Start

### 1. Build the Application
```bash
# Windows (PowerShell/Terminal)
.\build.bat

# Git Bash (manual)
cd dev
go build -ldflags "-s -w" -o ../release/scraper.exe scraper.go
go build -ldflags "-s -w" -o ../release/process.exe cmd/process/data-processor.go
go build -ldflags "-s -w" -o ../release/indexcsv.exe cmd/indexcsv/index-extractor.go
go build -ldflags "-s -w" -o ../release/web-licensed.exe cmd/web-licensed/web-application.go
```

### 2. Run the Application
```bash
cd release
web-licensed.exe
```

### 3. Access the Web Interface
Open http://localhost:8080 in your browser

## Project Structure
```
ISXDailyReportsScrapper/
├── dev/                    # Source code directory
│   ├── cmd/               # Command-line applications
│   │   ├── indexcsv/      # Index extractor
│   │   ├── process/       # Data processor
│   │   └── web-licensed/  # Web application
│   ├── internal/          # Internal packages
│   │   ├── common/        # Common utilities (logger, paths, errors)
│   │   ├── license/       # License management
│   │   ├── pipeline/      # Pipeline manager and orchestration
│   │   ├── progress/      # Progress tracking and WebSocket message utilities
│   │   ├── analytics/     # Data analysis and summary generation
│   │   ├── exporter/      # CSV and data export functionality
│   │   ├── files/         # File system operations
│   │   ├── parser/        # Excel file parsing
│   │   ├── processor/     # Data processing and forward-filling
│   │   └── websocket/     # WebSocket hub and client management
│   ├── web/               # Web interface assets
│   └── scraper.go         # Main scraper application
├── release/               # Built executables and runtime files
│   ├── data/             
│   │   ├── downloads/     # Downloaded Excel files
│   │   ├── reports/       # Processed CSV reports
│   │   └── metrics/       # Historical timing metrics
│   ├── logs/              # Application logs
│   └── web/               # Web interface files
└── build.bat              # Build script
```

## Key Components

### 1. Scraper (scraper.exe)
- Downloads daily Excel reports from ISX website
- Supports initial and accumulative modes
- Uses ChromeDP for web automation
- License validation required

### 2. Data Processor (process.exe)
- Processes downloaded Excel files into CSV format
- Combines multiple reports into unified dataset
- Extracts trading data for analysis

### 3. Index Extractor (indexcsv.exe)
- Extracts market indices (ISX60, ISX15) from Excel files
- Creates indexes.csv for charting

### 4. Web Application (web-licensed.exe)
- Provides web interface on http://localhost:8080
- Real-time WebSocket communication with enhanced progress tracking
- License validation and management
- Centralized pipeline orchestration using Pipeline Manager
- Auto-refresh data components when files are updated

## Important Files and Paths

### Configuration
- **License File**: `release/license.dat`
- **Service Account**: Embedded in executables (Google Sheets API)

### Data Directories
- **Downloads**: `release/data/downloads/` - Excel files (YYYY MM DD ISX Daily Report.xlsx)
- **Reports**: `release/data/reports/` - Processed CSV files
  - `isx_combined_data.csv` - All trading data
  - `indexes.csv` - Market indices data
  - `ticker_summary.json` - Ticker statistics

### Web Interface
- **Main HTML**: `dev/web/index.html`
- **Static Assets**: `dev/web/static/`

## Common Commands

### Manual Operations
```bash
# Scrape new reports (initial mode)
scraper.exe --mode initial --from 2024-01-01 --to 2024-01-31

# Process downloaded files
process.exe

# Extract indices
indexcsv.exe
```

### Testing Commands
```bash
# Run all tests with coverage
cd dev && go test ./... -cover

# Run tests with race detection
cd dev && go test ./... -race

# Generate coverage report
cd dev && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
cd dev && go test ./internal/analytics -v

# Run with debug logging
set ISX_DEBUG=true
web-licensed.exe

# Run single test
cd dev && go test -v ./internal/pipeline -run TestManager
```

## Architecture Principles (CRITICAL)
**All development MUST follow the three-layer architecture:**

1. **Frontend (HTML/JS)**: Display and input collection ONLY
   - NO business logic
   - NO pipeline control
   - Only shows what backend tells it

2. **Backend (Go)**: ALL logic and control
   - Pipeline stage transitions
   - Business logic and validation
   - Process management
   - Uses exit codes (not WebSocket) for control

3. **WebSocket**: Status updates ONLY
   - One-way: Backend → Frontend
   - Never used for control flow
   - Display purposes only

**Key Rules:**
- Pipeline progression based on process exit codes (0 = success)
- WebSocket messages are for UI updates only
- All stage transitions happen in backend Go code
- See `docs/design/ARCHITECTURE_PRINCIPLES.md` for details

## Development Guidelines

### Task Management Workflow
We use a structured task management system. All tasks are tracked in `docs/developer/DEVELOPMENT_TASKS.md`.

#### Task States:
- **BACKLOG**: Task identified but not yet prioritized
- **READY**: Task is prioritized and ready to be worked on
- **IN_PROGRESS**: Currently being developed
- **IN_REVIEW**: Code complete, undergoing review
- **DONE**: Fully completed and deployed
- **BLOCKED**: Cannot proceed due to dependencies

#### Task ID Format:
Tasks use format `[EPIC-XXX]`:
- **COMM**: Communication & Real-time Updates
- **DATA**: Data Processing & Analytics
- **INFRA**: Infrastructure & Deployment
- **SEC**: Security & Authentication
- **UI**: User Interface & Experience
- **API**: API & Integrations
- **BUG**: Bug Reporting System

### Code Style Guidelines
- **Imports**: Standard library first, then third-party, then internal packages
- **Naming**: PascalCase for exported, camelCase for unexported, snake_case for JSON
- **Errors**: Use `common.NewAppError()` with typed errors from `common/errors.go`
- **Logging**: Use `common.NewLogger()` with categories (PIPELINE, WEBSOCKET, etc.)
- **Comments**: No unnecessary comments - code should be self-documenting
- **Data**: Dates=YYYY-MM-DD, timestamps=RFC3339, CSV headers=PascalCase, JSON=snake_case

### Testing Best Practices

#### Unit Testing Guidelines
Every new feature MUST include comprehensive unit tests:

```go
// Test naming: Test<FunctionName>_<Scenario>
func TestPipelineManager_ExecuteAllStages(t *testing.T) {
    // Arrange
    manager := NewManager()
    
    // Act
    result := manager.Execute()
    
    // Assert
    assert.Equal(t, expected, result)
}

// Table-driven tests for multiple scenarios
func TestProgressCalculator_CalculateETA(t *testing.T) {
    tests := []struct {
        name     string
        input    ProgressData
        expected time.Duration
    }{
        {"Normal case", ProgressData{Files: 10, Processed: 5}, 30 * time.Minute},
        {"Empty case", ProgressData{Files: 0, Processed: 0}, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateETA(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Test Coverage Requirements
- **Minimum 80% coverage** for new packages
- **100% coverage** for critical business logic (pipeline, data validation)
- **Integration tests** for WebSocket message flows
- **Mock external dependencies** (file system, network calls)

### Development Best Practices

#### MANDATORY: Test-Driven Development Workflow
**⚠️ CRITICAL: All tasks MUST follow the Test-Driven Development Workflow**
- See `docs/developer/TEST_DRIVEN_DEVELOPMENT_WORKFLOW.md` for complete requirements
- NO CODE can be merged without comprehensive automated tests
- Minimum 100% test coverage for new code, 95% for modified code
- All test categories must be covered: Unit, Integration, Security, Performance

#### Test Requirements for EVERY Task
1. **Before Coding**: Create test plan with all scenarios
2. **Write Tests First**: Tests must fail before implementation
3. **Test Categories** (all mandatory):
   - Unit Tests: 100% coverage of new functions
   - Integration Tests: Component interactions
   - Security Tests: Input validation, auth, injection prevention
   - Communication Tests: WebSocket/API message validation
   - Data Integrity Tests: No data loss, format validation
   - Performance Tests: Benchmarks, no regression
   - E2E Tests: Complete user workflows

4. **Automated Enforcement**:
   - Pre-commit hooks run all tests
   - CI/CD blocks merge if tests fail
   - Coverage below threshold blocks merge

#### Code Organization
1. **Single Responsibility**: Each function/type should have one clear purpose
2. **Interface Segregation**: Create focused interfaces
3. **Dependency Injection**: Pass dependencies as interfaces for testability
4. **Error Handling**: Always check errors, wrap with context

#### Logging Standards
```go
// Use structured logging with categories
logger := common.NewLogger("PIPELINE")
logger.Info("Starting stage", "stage", "scraping", "files", len(files))

// Debug logging behind flag
if os.Getenv("ISX_DEBUG") == "true" {
    logger.Debug("Detailed processing info", "file", filename, "records", count)
}
```

#### Git Workflow
1. **Feature branches** named after task ID (e.g., `comm-001-websocket-messages`)
2. **Atomic commits** with clear messages referencing task ID
3. **Pull requests** for code review before merging
4. **Rebase** to keep clean history

### Security Best Practices
- **Never commit secrets** (use environment variables)
- **Input validation** for all user inputs
- **File path sanitization** to prevent directory traversal
- **Rate limiting** for external API calls
- **Secure defaults** for all configurations

### Data Validation Requirements
When modifying code that generates or consumes data, you MUST:

1. **Review Data Specifications**: Check DATA_SPECIFICATIONS.md for the correct format
2. **Use Validation Checklist**: Follow DATA_VALIDATION_CHECKLIST.md before committing
3. **Ensure Consistency**:
   - CSV headers must match exactly (case-sensitive, correct order)
   - JSON fields must use snake_case (e.g., `last_price`, not `lastPrice`)
   - Dates: YYYY-MM-DD format everywhere
   - Timestamps: ISO 8601 format (RFC3339)
   - Booleans in CSV: lowercase "true"/"false" strings
   - Arrays in JSON: Never null, use empty array [] instead
4. **Handle Edge Cases**:
   - CSV readers must handle UTF-8 BOM
   - JavaScript must check null/undefined before .toFixed()
   - TradingStatus: true = actual trading, false = forward-filled
5. **Test Data Flow**: Verify data works end-to-end from generation to display

### Deployment Checklist
Before any release:
- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run` if available)
- [ ] Security scan completed
- [ ] Performance benchmarks meet requirements
- [ ] Documentation updated
- [ ] Breaking changes documented
- [ ] Migration guide provided (if needed)

## Recent Updates

### v0.3.1-alpha - Market Movers Analytics
- **Real-time Market Data**: Complete Market Movers functionality
- **Analytics Integration**: Centralized data processing using `internal/analytics`
- **Professional UI**: Market Movers page with Bootstrap 5 styling

### v0.3.0-alpha - Pipeline Manager
- **Fixed Critical Bug**: Pipeline now correctly executes all stages
- **New Architecture**: Centralized pipeline manager with automatic progression
- **Enhanced Progress**: Historical metrics for better ETA estimates
- **UI Improvements**: Indices chart fixes, auto-refresh components

### v0.2.0-alpha - WebSocket Standardization
- **Standardized Messages**: Consistent WebSocket format across all executables
- **Progress Tracking**: Immediate ETA estimates using historical data
- **Real-time Updates**: File watcher for automatic UI refresh

### v0.1.0-alpha - Real-time System
- **WebSocket Hub**: Real-time communication system
- **Data Processing**: Fixed BOM handling, standardized formats
- **Build Enhancements**: Detailed progress reporting

## Bug Reporting System (v0.8.0-beta - Planned)
- **Google Sheets Integration**: Bug reports stored in `ISX_Bug_Reports_Tracker`
- **Image Upload**: Screenshots to Google Drive
- **Real-time Feedback**: WebSocket-based submission status
- **Structured Data**: Standardized bug report format

## Documentation Structure
```
docs/
├── user/                    # End-user guides
├── developer/              # Developer documentation
├── specifications/         # Technical specifications
├── design/                # Design documents
├── operations/            # Deployment and operations
└── reference/             # Reference materials
```

Key documents:
- **docs/README.md** - Documentation index
- **docs/user/README.md** - Main user documentation
- **docs/specifications/DATA_SPECIFICATIONS.md** - Data format specifications
- **docs/design/FRONTEND_ARCHITECTURE.md** - Frontend modularization guide

## Known Issues
- None currently

## Contact
The Iraqi Investor Group - For license renewals and support