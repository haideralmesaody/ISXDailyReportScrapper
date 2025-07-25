# Changelog

All notable changes to the ISX Daily Reports Scrapper will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive test suite for WebSocket implementation with race detection
- Handler tests for health, client logging, and data endpoints
- JavaScript unit tests using Jest for core modules (Logger, EventBus, WebSocket)
- Race detector setup for Windows ARM64 development
- Testing guide documentation with best practices
- Test coverage improvements across critical packages

### Changed
- WebSocket Hub implementation now thread-safe with mutex protection
- All broadcast methods now include timestamps for consistency
- Logger module exports class for better testability

### Fixed
- Race conditions in WebSocket Hub ClientCount method
- WebSocket test timing issues with proper synchronization
- Handler test compilation errors with proper interface implementations

### Documentation
- Added comprehensive TESTING_GUIDE.md
- Updated CLAUDE.md compliance for all test files
- Added inline documentation for test patterns

## [0.5.0] - 2025-07-25

### Added
- Comprehensive Playwright E2E test suite for automated testing
- Date parameter validation tests
- Pipeline status real-time updates via WebSocket
- Enhanced logging for parameter transformation
- MCP (Model Context Protocol) browser automation support
- Test infrastructure with license activation automation

### Changed
- WebSocket message types updated to match frontend expectations (e.g., `pipeline:progress` instead of `pipeline_progress`)
- Parameter extraction in pipeline service to handle nested JSON structure
- Improved error handling with proper parameter validation
- Enhanced pipeline stage tracking with start events

### Fixed
- Pipeline status updates not displaying in UI - fixed WebSocket message format mismatch
- Date parameter communication failure - scraper now correctly respects date ranges
- Frontend sending `{args: {from, to}}` but backend expecting flat structure
- Scraper downloading all files instead of date-filtered subset
- WebSocket adapter not transforming message types correctly

### Technical Details
- Updated `internal/websocket/types.go` with correct message type constants
- Fixed parameter extraction in `internal/services/pipeline_service.go`
- Added parameter transformation from `from`/`to` to `from_date`/`to_date`
- Created automated tests for date parameter validation

## [0.4.0] - 2025-07-24

### Added
- WebSocket real-time progress tracking for all pipeline stages
- RFC 7807 compliant error responses
- Enhanced error display component
- Market movers functionality
- Ticker charts with historical data
- Market indices tracking (ISX60, ISX15)
- Structured logging with slog
- Request ID propagation
- Panic recovery middleware

### Changed
- API endpoints aligned with RESTful patterns
- File paths standardized to `data/downloads` structure
- Improved Chi middleware organization using route groups
- WebSocket route registration moved before middleware
- Frontend API service updated to match backend routes

### Fixed
- WebSocket connection issues with middleware - used Chi route groups
- JavaScript APIError global access - exported to window object
- Frontend API endpoint mismatches - updated all endpoints
- File path inconsistencies - standardized to `data/` structure
- Chi middleware ordering panic - proper route group implementation

### Security
- All routes protected by license validation
- CORS properly configured
- WebSocket origin validation

## [0.3.0] - 2025-07-15

### Added
- License management system with AES-GCM encryption
- Web-based license activation interface
- Pipeline orchestration with stage dependencies
- WebSocket hub for real-time updates
- Data analysis and reporting features

### Changed
- Migrated from Gorilla Mux to Chi router
- Restructured project layout for better organization
- Updated build process to create `release` directory

### Deprecated
- Old Gorilla Mux routing (to be removed in v1.0.0)

## [0.2.0] - 2025-07-01

### Added
- Initial web interface
- Basic scraping functionality
- Excel to CSV conversion
- Index extraction (ISX60, ISX15)

### Fixed
- Excel parsing for Arabic content
- Date formatting issues

## [0.1.0] - 2025-06-15

### Added
- Initial release
- Command-line scraper for ISX daily reports
- Basic Excel file processing
- Simple CSV output

[Unreleased]: https://github.com/haideralmesaody/ISXDailyReportScrapper/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/haideralmesaody/ISXDailyReportScrapper/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/haideralmesaody/ISXDailyReportScrapper/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/haideralmesaody/ISXDailyReportScrapper/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/haideralmesaody/ISXDailyReportScrapper/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/haideralmesaody/ISXDailyReportScrapper/releases/tag/v0.1.0