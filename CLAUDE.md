# ISX Daily Reports Scrapper - Project Information

## Project Overview
ISX Daily Reports Scrapper is a licensed web application that scrapes daily reports from the Iraq Stock Exchange (ISX) website, processes the data, and provides analytics through a web interface.

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
│   │   └── license/       # License management
│   ├── web/               # Web interface assets
│   └── scraper.go         # Main scraper application
├── release/               # Built executables and runtime files
│   ├── data/             
│   │   ├── downloads/     # Downloaded Excel files
│   │   └── reports/       # Processed CSV reports
│   ├── logs/              # Application logs
│   └── web/               # Web interface files
├── tests/                 # Test suite
│   ├── unit/             # Unit tests
│   ├── integration/      # Integration tests
│   ├── e2e/              # End-to-end tests
│   ├── performance/      # Performance benchmarks
│   └── scripts/          # Test execution scripts
├── docs/                  # Documentation
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
- Real-time WebSocket communication
- License validation and management
- Pipeline orchestration (scraping → processing → indexing)

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

## Build Process
```bash
# Use build.bat to compile all executables
build.bat
```

The build script:
1. Backs up existing data and license
2. Creates release directory structure
3. Copies web assets
4. Builds all Go executables with `-ldflags "-s -w"` (release mode)
5. Restores backed up data

## Common Commands

### Running the Application
```bash
cd release
web-licensed.exe
```

### Manual Operations
```bash
# Scrape new reports (initial mode)
scraper.exe --mode initial --from 2024-01-01 --to 2024-01-31

# Process downloaded files
process.exe

# Extract indices
indexcsv.exe
```

## Development Guidelines

### Security Considerations
- Never log sensitive data (license keys, private keys)
- Use conditional debug logging with ISX_DEBUG environment variable
- All paths should be relative to executable location
- Embedded credentials are stripped in release builds

### Code Style
- Use existing utilities from internal/common package
- Follow Go conventions
- Maintain backward compatibility
- Test thoroughly before committing

### Testing

#### Test Structure
All tests are organized in the `/tests/` directory:
- `unit/` - Unit tests for individual components
- `integration/` - Integration tests for component interactions
- `e2e/` - End-to-end tests for complete workflows
- `performance/` - Performance benchmarks
- `scripts/` - Test execution scripts
- `coverage/` - Coverage reports (generated)
- `results/` - Test results (generated)

#### Running Tests
```bash
# Run all tests
cd tests/scripts
run_tests.bat

# Run specific test types
run_unit_tests.bat
run_integration_tests.bat
run_e2e_tests.bat

# Generate coverage report
generate_coverage.bat

# Skip E2E tests (faster)
set SKIP_E2E=true
run_tests.bat

# Run with debug logging
set ISX_DEBUG=true
web-licensed.exe
```

#### Test Guidelines
- Write unit tests for new functions
- Add integration tests for component interactions
- Ensure all tests pass before committing
- Maintain 80%+ code coverage
- Use test utilities from `/tests/testutil/`

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

## Documentation Structure

All documentation is organized in the `docs/` directory:

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
- **docs/README.md** - Documentation index and guide
- **docs/user/README.md** - Main user documentation
- **docs/specifications/DATA_SPECIFICATIONS.md** - Data format specifications
- **docs/specifications/COLUMN_NAME_MAPPING.md** - Column name transformations
- **docs/reference/CHANGELOG.md** - Version history
- **CLAUDE.md** - This file (AI assistant context)

### Archived Documentation
- `archived-docs/` - Historical development plans and completed task lists

## Recent Updates (v0.3.0-alpha - 2025-07-23)

### Chi Framework Migration (INFRA-019)
1. **Replaced gorilla/mux with Chi framework** - Lightweight, Google-aligned HTTP router
2. **Fixed critical HTTP 206 issues** - Resolved browser page hanging and incomplete loading
3. **Improved middleware architecture** - Chi's battle-tested middleware stack
4. **Enhanced request tracing** - Unique request IDs for better debugging
5. **Optimized static file serving** - Compression and proper caching headers
6. **Fixed WebSocket handling** - Eliminated middleware interference with upgrades

### Web Application Architecture
- **Framework**: Chi v5.1.0 (replaced gorilla/mux)
- **Middleware Stack**: RequestID, RealIP, Logger, Recoverer, Timeout, Compression
- **Route Organization**: Logical grouping with `/api/license`, `/api/pipeline`, `/api/data`
- **Static Serving**: FileServer with ETags and gzip compression
- **WebSocket**: Clean upgrade process without middleware interference

### Technical Improvements
1. **HTTP Response Reliability** - All endpoints return proper HTTP 200 (no more 206)
2. **Performance Optimization** - Faster routing with composable middleware
3. **Error Handling** - Centralized error middleware with structured responses
4. **Logging Enhancement** - Structured logging with request correlation
5. **Development Experience** - Better debugging with request tracing

### Previous Updates (v2.0.0 - 2025-07-18)

### Real-time Update System
1. Implemented WebSocket hub for real-time communication
2. Added file watcher to monitor data directory changes
3. Frontend DataUpdateManager handles automatic UI updates
4. No manual refresh needed - all changes propagate instantly

### Data Processing Improvements
1. Fixed BOM handling in CSV readers
2. Standardized column matching for PascalCase headers
3. Consistent TradingStatus format ("true"/"false" strings)
4. All JSON fields use snake_case naming

### Build System Enhancements
1. Added detailed progress reporting with success/failure counts
2. Build summary shows status of each component
3. Automatic backup and restore of data during builds

## Recent Fixes (July 2024)
1. Removed sensitive logging from license manager
2. Created common utilities (logger, paths, errors)
3. Fixed duplicate canvas variable in index.html
4. Fixed pipeline status transitions
5. Fixed indices chart data loading
6. Consolidated build scripts (using build.bat)
7. Created modular internal packages (files, exporter, processor, analytics)
8. Fixed ticker list JSON format (snake_case fields, proper structure)
9. Implemented actual trading days calculation (not forward-filled)

## Known Issues
- None currently

## Contact
The Iraqi Investor Group - For license renewals and support