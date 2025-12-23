# ISX Daily Reports Scrapper Commands

This directory contains the executable commands for the ISX Daily Reports Scrapper system.

## Commands

### scraper
Downloads ISX daily Excel reports from the official website.
- Supports initial and accumulative modes
- Requires valid license
- Saves to `{exe_dir}/data/downloads/`

### process
Processes downloaded Excel files into CSV format.
- Generates combined, daily, and per-ticker CSV files
- Implements forward-fill for missing trading data
- Reads from `{exe_dir}/data/downloads/`
- Writes to `{exe_dir}/data/reports/`

### indexcsv
Extracts ISX60 and ISX15 index values from Excel files.
- Creates time-series CSV of index values
- Supports accumulative mode
- Outputs to `{exe_dir}/data/reports/indexes.csv`

### web-licensed
Main web server with embedded frontend.
- Serves Next.js frontend
- Provides REST API and WebSocket
- Requires license validation
- Default port 8080

## Build Instructions
```bash
# Build all commands
cd dev
go build ./cmd/...

# Or use the build script
../build.bat
```

## Path Resolution
All commands use centralized path management from `internal/config/paths.go`:
- Paths are relative to the executable location
- Automatic directory creation on startup
- Consistent structure across all commands

## Change Log
- 2025-07-31: Renamed pipeline to operations for business-friendly terminology
- 2025-07-31: Updated all references from stages to steps
- 2025-07-31: Simplified implementation - removed unnecessary adapter pattern and duplicate types
- 2025-01-29: All commands updated to use centralized path management
- 2025-01-29: Added consistent startup logging for path resolution
- 2025-01-29: Documented path resolution behavior for all commands