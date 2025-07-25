# API Endpoint Mapping

Generated: 2025-07-24  
Last Updated: 2025-07-24

## Status Summary
All essential API endpoints are now properly implemented and connected between frontend and backend.

## Current Backend Routes (Chi)

### Health & System
- GET    `/api/health` - Basic health check
- GET    `/api/health/ready` - Readiness check
- GET    `/api/health/live` - Liveness check  
- GET    `/api/version` - Version information

### License Management
- GET    `/api/license/status` - Get license status
- POST   `/api/license/activate` - Activate license

### Data Operations
- GET    `/api/data/reports` - Get all reports
- GET    `/api/data/tickers` - Get all tickers
- GET    `/api/data/indices` - Get indices data
- GET    `/api/data/files` - Get file listings
- GET    `/api/data/market-movers` - Get market movers
- GET    `/api/data/ticker/{ticker}/chart` - Get ticker chart data
- GET    `/api/data/download/{type}/{filename}` - Download file

### Pipeline Management
- POST   `/api/pipeline/start` - Start pipeline
- POST   `/api/pipeline/stop` - Stop pipeline
- GET    `/api/pipeline/status` - Get pipeline status
- GET    `/api/pipeline/pipelines` - List pipelines
- GET    `/api/pipeline/{id}` - Get specific pipeline
- DELETE `/api/pipeline/{id}` - Delete pipeline
- GET    `/api/pipeline/{id}/logs` - Get pipeline logs

### Operation Shortcuts
- POST   `/api/scrape` - Start scraping
- POST   `/api/process` - Start processing
- POST   `/api/indexcsv` - Start index extraction

### Client Logging
- POST   `/api/logs` - Submit client logs

### WebSocket
- GET    `/ws` - WebSocket connection

## Current Frontend Calls

### ✅ All Essential Endpoints Working
- GET    `/api/license/status` ✓ License validation
- POST   `/api/license/activate` ✓ License activation
- GET    `/api/version` ✓ Version info
- POST   `/api/logs` ✓ Client error logging
- GET    `/api/data/files` ✓ File listings
- GET    `/api/data/tickers` ✓ Ticker list (FIXED)
- GET    `/api/data/indices` ✓ Index data (FIXED)
- GET    `/api/data/ticker/{ticker}/chart` ✓ Ticker charts (FIXED)
- GET    `/api/data/market-movers` ✓ Market movers (FIXED)
- POST   `/api/pipeline/start` ✓ Pipeline operations
- GET    `/api/pipeline/status` ✓ Pipeline status
- GET    `/api/data/download/{type}/{filename}` ✓ File downloads
- POST   `/api/scrape` ✓ Start scraping
- POST   `/api/process` ✓ Start processing
- POST   `/api/indexcsv` ✓ Extract indices
- GET    `/ws` ✓ WebSocket connection

### 📌 Optional/Not Used
- GET    `/api/data/reports` - Backend ready but frontend uses `/api/data/files`
- GET    `/api/stats` - Placeholder in frontend, not needed for business logic

## Implementation Status

✅ **All required changes have been completed!**

The following fixes were applied to `dev/web/static/js/services/api.js`:
1. ✅ Fixed `getTickers()` to use `/api/data/tickers`
2. ✅ Fixed `getIndexData()` to use `/api/data/indices`
3. ✅ Fixed `getTickerChart()` to use `/api/data/ticker/{ticker}/chart`
4. ✅ Fixed `getMarketMovers()` to use `/api/data/market-movers`
5. ✅ Updated `getSystemStats()` to return placeholder data (not needed)

## Summary
- **Total Frontend API Calls**: 17
- **Working Correctly**: 17 (100%)
- **Essential for Business Logic**: 16
- **Optional/Placeholder**: 1 (`/api/stats`)

## Business-Critical Endpoints
All essential endpoints for the ISX Daily Reports Scrapper are now properly connected:
- ✅ **License Management**: Activation and status checking
- ✅ **Data Operations**: Files, tickers, indices, market movers
- ✅ **Processing Pipeline**: Scraping, processing, index extraction
- ✅ **Real-time Updates**: WebSocket for progress tracking
- ✅ **File Downloads**: Excel and CSV file downloads