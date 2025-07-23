# Troubleshooting Guide

## Common Issues and Solutions

### 1. Web Application Won't Start

#### Symptom
`web-licensed.exe` fails to start or immediately closes

#### Possible Causes & Solutions

**Port 8080 Already in Use**
```bash
# Check what's using port 8080
netstat -ano | findstr :8080

# Solution: Kill the process or change port
taskkill /PID <process_id> /F
```

**Missing License File**
- Ensure `license.dat` exists in the release folder
- Re-activate license through the web interface
- Check `logs/license.log` for details

**Antivirus Blocking**
- Add exception for `web-licensed.exe`
- Whitelist the entire `release/` folder

### 2. Scraper Not Downloading Files

#### Symptom
Scraper runs but no files appear in `data/downloads/`

#### Solutions

**Chrome Not Installed**
- Install Google Chrome (required for scraping)
- Scraper uses Chrome in headless mode

**ISX Website Changed**
- Check if ISX website structure changed
- View `logs/scraper.log` for specific errors
- May need code update for new selectors

**Network Issues**
```bash
# Test connectivity to ISX
ping www.isx-iq.net

# Check DNS resolution
nslookup www.isx-iq.net
```

### 3. Data Processing Errors

#### Symptom
"Required columns not found" error during processing

#### Common Causes

**BOM in CSV Files**
- Already handled in v2.0+
- If persists, check file encoding

**Column Name Changes**
- ISX changed Excel column headers
- Check `docs/specifications/COLUMN_NAME_MAPPING.md`
- Update parser mappings if needed

**Corrupted Excel File**
```bash
# Try opening file manually
# Re-download if corrupted
scraper.exe --date=2024-01-15
```

### 4. Ticker List Not Updating

#### Symptom
Web interface shows old data after processing

#### Pre v2.0 Solution
- Hard refresh browser (Ctrl+F5)
- Clear browser cache

#### v2.0+ (Should Auto-Update)
If not updating automatically:
- Check WebSocket connection in browser console
- Ensure file watcher is running
- Check `data/reports/` permissions

### 5. Real-time Updates Not Working

#### Symptom
Changes to files don't reflect in UI automatically

#### Diagnostics
```javascript
// In browser console, check WebSocket
ws.readyState
// Should be 1 (OPEN)
```

#### Solutions

**WebSocket Connection Failed**
- Check firewall settings
- Ensure port 8080 allows WebSocket
- Try different browser

**File Watcher Issues**
- Check logs for watcher errors
- Ensure proper read permissions on `data/reports/`
- Windows file system delays (rare)

### 6. Build Failures

#### Symptom
`build.bat` fails with errors

#### Common Issues

**Go Not Installed**
```bash
# Check Go installation
go version
# Should be 1.23 or higher
```

**Missing Dependencies**
```bash
cd dev
go mod download
go mod tidy
```

**Import Errors**
- Ensure you're in the correct directory
- All imports should use "isxcli" module name

### 7. License Activation Issues

#### Symptom
"Invalid license" or "License expired"

#### Solutions

**Invalid License Key**
- Verify key from license provider
- Check for typos or extra spaces
- Keys are case-sensitive

**Machine ID Changed**
- License tied to machine
- Contact support for transfer
- Hardware changes may trigger this

**Clock Skew**
- Ensure system time is correct
- License validation uses system time

### 8. Performance Issues

#### Symptom
Slow processing or high memory usage

#### Optimizations

**Large Dataset Processing**
```bash
# Process specific date range
process.exe --start-date=2024-01-01 --end-date=2024-01-31

# Increase memory limit if needed
set GOGC=200
```

**Browser Performance**
- Limit ticker list display
- Clear old log entries periodically
- Use pagination for file list

### 9. Data Inconsistencies

#### Symptom
Data in CSV doesn't match Excel source

#### Validation Steps

1. **Check Column Mappings**
   - Review `COLUMN_NAME_MAPPING.md`
   - Verify Excel headers match expected

2. **Verify Processing Logic**
   - Forward-filling should only apply to missing dates
   - TradingStatus should be "false" for filled data

3. **Run Validation**
   - Use `DATA_VALIDATION.md` checklist
   - Compare sample records manually

### 10. Logging and Debugging

#### Enable Debug Mode
```bash
# Windows
set ISX_DEBUG=true
web-licensed.exe

# Or for specific component
set ISX_DEBUG=true
process.exe
```

#### Log Locations
- `logs/audit.log` - License and security events
- `logs/license.log` - License validation details
- Console output - Real-time processing information

#### Common Log Messages

**"Failed to generate ticker summary"**
- Usually column mismatch
- Check combined CSV has all required columns

**"WebSocket upgrade error"**
- Client connection issue
- Check browser compatibility

**"License validation failed"**
- See license.log for specific reason
- Common: expired, wrong machine, invalid key

## Getting Help

If issues persist:

1. **Collect Information**
   - Error messages (exact text)
   - Log files
   - Steps to reproduce
   - System information (Windows version, Go version)

2. **Check Documentation**
   - Review relevant specification docs
   - Check CHANGELOG for recent changes

3. **Report Issue**
   - GitHub Issues with collected information
   - Include minimal reproduction steps
   - Attach relevant log excerpts (not full logs)