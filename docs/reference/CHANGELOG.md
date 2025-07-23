# Changelog

All notable changes to the ISX Daily Reports Scrapper project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Chi Framework Migration (INFRA-019)**: Migration from gorilla/mux to Chi framework
  - Resolved critical HTTP 206 partial content issues causing browser page hanging
  - Replaced problematic custom LoggingMiddleware with Chi's battle-tested middleware stack
  - Implemented proper request ID tracing and structured logging
  - Fixed WebSocket upgrade interference from middleware
  - Improved performance with lightweight routing and composable middleware
  - Added proper static file serving with compression and caching
  - Organized routes into logical groups with Chi's route mounting
  - Enhanced error handling and recovery mechanisms

### Fixed
- **HTTP Response Issues**: Resolved browser hanging and incomplete page loading
  - Fixed HTTP 206 (Partial Content) responses that prevented complete HTML delivery
  - Eliminated response hanging issues in license page and main application
  - Resolved middleware interference with WebSocket upgrade process
  - Fixed custom responseWriter wrapper compatibility issues

### Technical
- Updated dependencies to include Chi v5.1.0 framework
- Refactored HTTP routing architecture for better maintainability
- Improved middleware composition and selective application
- Enhanced request tracing with unique request IDs
- Optimized static asset serving with proper caching headers

### Planning
- Future features and improvements are tracked in DEVELOPMENT_TASKS.md

## [0.3.1-alpha] - 2025-07-21

### Added
- **Market Movers Analytics Integration**: Complete Market Movers functionality with real-time data
  - Daily, weekly, and monthly percentage change calculations
  - Volume and value tracking for trading activity analysis
  - 52-week high/low price tracking for historical context
  - Previous close price data for accurate change calculations
  - Real-time data updates via WebSocket integration

### Fixed
- **Critical Market Movers Data Flow**: Resolved analytics integration preventing Market Movers page from displaying data
  - Root cause: Web application had duplicate `generateTickerSummary()` function overriding analytics package
  - Solution: Replaced 200+ line duplicate code with proper analytics package integration
  - Fixed path construction bug causing invalid file paths (`C::` error) in analytics generation
  - Ensured consistent data structure between analytics package and web application
- **Analytics Package Integration**: Streamlined ticker summary generation
  - Added proper analytics import to web-licensed application
  - Eliminated code duplication between analytics package and web application
  - Improved maintainability by centralizing Market Movers calculations in analytics package

## [0.3.0-alpha] - 2025-01-20

### Added
- **Pipeline Manager Implementation (PIPE-002)**: Complete pipeline orchestration overhaul
  - Centralized pipeline orchestration system in `internal/pipeline`
  - Stage-based architecture with automatic dependency management  
  - Individual stage implementations in `cmd/web-licensed/stages/`
  - WebSocket adapter for real-time progress updates maintaining frontend compatibility
- **Enhanced Progress Tracking System**:
  - Historical metrics storage in `data/metrics/` directory
  - `EnhancedCalculator` with immediate ETA estimates using past performance data
  - Dynamic progress adjustment based on actual findings (non-trading days, existing files)
  - Standardized WebSocket message format with `[WEBSOCKET_PROGRESS]`, `[WEBSOCKET_STATUS]`, `[WEBSOCKET_ERROR]` prefixes
- **Smart File Count Calculation**: Sophisticated approach to estimate remaining downloads
  - Initial estimate counts all calendar business days (Sunday-Thursday) in date range
  - Real-time discovery of non-trading days and market holidays
  - Completion logic based on actual file overlap rather than expected counts
- **UI and Chart Improvements**:
  - Fixed indices chart Y-axis scaling to properly display both ISX60 and ISX15
  - Enhanced horizontal axis to show all available dates with intelligent skipping
  - Ticker list auto-refresh functionality when CSV files are updated
  - Fixed pipeline stage status messages to show proper completion states

### Changed
- **BREAKING**: Completely replaced complex handleScrape implementation
  - Reduced from 200+ lines to ~50 lines using Pipeline Manager
  - Eliminated fragile scraperSuccess variable propagation that caused critical bug
  - Pipeline control now centralized and automatic based on stage dependencies
- **Enhanced WebSocket Communication**: All executables now use standardized message format
- **Improved Chart Configuration**: Better aspect ratios and responsive design for indices charts
- **Progress Display**: Pipeline stages show original descriptions when completed instead of last progress message

### Fixed
- **Critical Bug**: Pipeline now correctly executes all stages (was stopping after scraping)
  - Root cause: scraperSuccess variable not propagating through complex control flow
  - Solution: Centralized pipeline manager with automatic stage progression  
- **Pipeline Stage Status**: Fixed final status messages showing incorrect states
- **Chart Display Issues**: Indices chart now properly scales to show both ISX60 and ISX15 values
- **Auto-refresh**: Ticker list and charts automatically update when underlying data changes
- **Race Conditions**: Eliminated timing issues and state propagation problems with centralized state management

### Technical Improvements
- **Clean Error Boundaries**: Each pipeline stage properly isolated with individual error handling
- **Consistent Progress Tracking**: All stages use standardized progress reporting
- **Easy Extension**: Adding new pipeline stages now requires only implementing the Stage interface
- **Maintainable Code**: Significantly reduced complexity in core pipeline logic
- **Historical Learning**: System learns actual ISX trading patterns and improves ETA accuracy over time

## [0.2.0-alpha] - 2025-01-19

### Added
- Real-time WebSocket update system with file watchers
- Enhanced progress tracking with historical metrics and ETA calculations
- Data processing improvements with BOM handling
- Build system enhancements with progress reporting
- Auto-refresh capabilities for UI components
- Standardized column matching for CSV data
- Historical timing metrics for improved ETA accuracy

### Changed
- Updated version numbering to follow SemVer (0.2.0-alpha)
- Enhanced development roadmap with planned features
- Reorganized development documentation structure

### Fixed
- BOM handling in CSV readers
- Ticker list JSON format (snake_case fields)  
- Pipeline status transitions
- Canvas variable conflicts in UI
- Build script reliability

### Documentation
- Added ARCHITECTURE_PRINCIPLES.md - Core three-layer architecture guidelines
- Added PIPELINE_ARCHITECTURE.md - Pipeline stage specifications
- Added FEATURE_TEMPLATE.md - Template for new feature development
- Added LIQUIDITY_CALCULATION_APPROACH.md - Comprehensive methodology
- Added LIQUIDITY_SCORING_SPECS.md - Technical implementation details
- Added STRATEGY_MODULE_DESIGN.md - Trading strategy module architecture
- Added STRATEGY_MODULE_SPECS.md - Technical specifications for strategies
- Created developer/README.md - Developer guide and index
- Updated WEBSOCKET_MESSAGE_SPECS.md - Clarified WebSocket is for status only
- Updated DEVELOPMENT_BEST_PRACTICES.md - Added architecture patterns section
- Updated CLAUDE.md - Added architecture principles for AI context
- Updated TASK_TEMPLATE.md - Added architecture compliance checklist

## [0.2.0-alpha] - 2025-01-19

### Added
- Standardized WebSocket message creation utilities in `internal/progress` package
  - Calculator for basic progress tracking with real-time ETA
  - EnhancedCalculator with historical metrics for improved predictions
  - MetricsManager for persisting timing data across runs
  - Helper functions for consistent message formatting
- Automatic ETA estimation based on historical processing times
- Metrics persistence in `data/metrics/` directory
- Unified error message format with recovery hints across all executables
- Progress messages now include detailed metadata for better tracking

### Changed
- All executables (scraper, processor, indexcsv) now use standardized progress message format
- Progress calculations are more accurate using historical timing data
- ETA estimates show "(estimated)" when using historical data, switching to actual once processing starts
- Error messages now include structured fields: code, details, stage, recoverable flag, and hint
- WebSocket messages are created using centralized utilities instead of inline JSON formatting

### Improved
- First-run experience shows "Calculating..." until actual timing data is available
- Subsequent runs show immediate ETA estimates based on past performance
- ETA accuracy improves over time as more historical data is collected
- Progress tracking is consistent across all pipeline stages

## [0.1.0-alpha] - 2025-01-19

### Added
- Real-time update system with WebSocket support
- File watcher for automatic data synchronization
- DataUpdateManager in frontend for handling real-time updates
- Comprehensive data specifications documentation
- Column name mapping documentation
- Data validation checklist
- Build summary with success/failure statistics

### Changed
- Restructured project: moved source to dev/ directory
- Updated all import paths to use "isxcli" module name
- Enhanced build script with detailed progress reporting
- Improved ticker summary generation with BOM handling
- Standardized data formats across all components

### Fixed
- BOM (Byte Order Mark) handling in CSV files
- Column matching issues with PascalCase headers
- TradingStatus format consistency (using "true"/"false" strings)
- JSON field naming to use snake_case consistently
- Ticker list not updating after processing

## [0.0.2-pre] - 2024-07-15

### Changed
- Modularized codebase into internal packages
- Separated concerns: parser, processor, exporter, analytics
- Improved error handling and logging
- Added comprehensive file management utilities

### Package Structure
- `internal/parser`: Excel file parsing
- `internal/processor`: Data processing and forward-filling
- `internal/exporter`: CSV and Google Sheets export
- `internal/analytics`: Data analysis and summaries
- `internal/files`: File system operations
- `internal/scraper`: Web scraping logic
- `internal/license`: License management
- `internal/updater`: Auto-update functionality

## [0.0.1-pre] - 2024-07-01

### Added
- Initial prototype release
- Web scraper for ISX daily reports
- Excel to CSV conversion
- Historical data processing with forward-filling
- Basic web interface for data visualization
- Ticker summary generation
- Index value extraction (ISX60, ISX15)
- License management system
- Google Sheets integration

### Known Issues Fixed Since Release
- Special characters in headers (/, &, etc.)
- Memory optimization for large datasets
- Duplicate removal in combined data
- Date parsing for various formats
- Unicode handling in company names

## Data Compliance Updates

### Completed Fixes
- ✅ Fixed JSON field naming in ticker summaries (camelCase → snake_case)
- ✅ Updated all CSV writers to include UTF-8 BOM
- ✅ Standardized TradingStatus format across all components
- ✅ Fixed column name transformations in parser
- ✅ Added comprehensive error handling for missing columns
- ✅ Implemented consistent date formatting (YYYY-MM-DD)
- ✅ Fixed float precision (prices: 3 decimals, percentages: 2 decimals)

### Data Format Standards Established
- CSV headers use PascalCase (e.g., CompanyName)
- JSON fields use snake_case (e.g., company_name)
- Empty numeric values default to 0 or 0.0
- Empty arrays in JSON never use null
- Boolean values in CSV use lowercase "true"/"false"
- All CSV files include UTF-8 BOM for Excel compatibility

## Migration Notes

### From v1.x to v2.0
1. Update all import paths from local references to "isxcli" module
2. Rebuild all executables using the new build script
3. Data files are backward compatible - no migration needed
4. License files remain in the same location

### Real-time Updates (v2.0+)
- No configuration needed - works automatically
- WebSocket connection established on web interface startup
- File changes detected and broadcast within 100ms
- All connected clients receive updates simultaneously