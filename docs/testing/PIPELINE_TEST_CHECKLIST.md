# Pipeline Test Checklist

## Pre-Test Setup
- [ ] Ensure all executables are built with latest changes
- [ ] Check that license.dat exists and is valid
- [ ] Verify data directories exist (data/downloads, data/reports, data/metrics)
- [ ] Close any running instances of web-licensed.exe

## Test Execution Steps

### 1. Start Web Application
```bash
cd release
web-licensed.exe
```

**Check:**
- [ ] Application starts on http://localhost:8080
- [ ] No errors in console
- [ ] License status shows correctly
- [ ] WebSocket connection established

### 2. Run Pipeline (Files Already Exist)
- Click "Run Complete Pipeline" when files already exist in downloads

**Expected Behavior:**
- [ ] Scraping status shows "active"
- [ ] Console shows "Using existing Excel files"
- [ ] Scraping status changes to "completed"
- [ ] Processing starts automatically
- [ ] Processing status shows "active"
- [ ] Indices extraction starts after processing
- [ ] Analysis runs and completes
- [ ] All stages show "completed" status

**Console Checks:**
- [ ] Look for `[PIPELINE]` log messages
- [ ] Check for `[WEBSOCKET_STATUS]` messages
- [ ] Verify no error messages

### 3. Run Pipeline (Fresh Download)
- Delete some Excel files from data/downloads
- Click "Run Complete Pipeline"

**Expected Behavior:**
- [ ] Scraping status shows "active"
- [ ] Progress bar updates during download
- [ ] Files download successfully
- [ ] Scraping status changes to "completed"
- [ ] Processing starts automatically
- [ ] Full pipeline completes

### 4. Error Scenarios

#### Test: Network Disconnection
- Disconnect network during scraping
- [ ] Error message appears
- [ ] Status shows "failed"
- [ ] Recoverable error hint provided

#### Test: Missing Executable
- Temporarily rename process.exe
- [ ] Appropriate error when trying to process
- [ ] Clear error message about missing file

## Common Issues and Solutions

### Issue: Pipeline Stops After Scraping
**Symptoms:**
- Scraping completes but processing doesn't start
- Frontend shows scraping as still "active"

**Checks:**
1. Check browser console for WebSocket errors
2. Look for `[PIPELINE]` logs in web-licensed console
3. Verify scraper exit code: `echo %ERRORLEVEL%`
4. Check if web-licensed.exe is still running

**Solutions:**
- Ensure WebSocket messages are being sent/received
- Check for JavaScript errors in browser
- Verify all executables have correct permissions

### Issue: Status Not Updating
**Symptoms:**
- Pipeline progresses but UI doesn't update

**Checks:**
1. WebSocket connection status
2. Browser console for errors
3. Network tab for WebSocket frames

**Solutions:**
- Refresh browser page
- Check for firewall/antivirus blocking
- Verify WebSocket hub is broadcasting

## Debug Commands

### Check Process Exit Codes
```batch
scraper.exe -mode=initial -from=2025-07-01 -to=2025-07-19
echo Exit Code: %ERRORLEVEL%
```

### Monitor WebSocket Messages
In browser console:
```javascript
// See all WebSocket messages
ws.addEventListener('message', (event) => {
    console.log('WS Message:', JSON.parse(event.data));
});
```

### Check File Permissions
```batch
icacls release\*.exe
icacls release\data
```

## Post-Test Cleanup
- [ ] Check logs for any warnings/errors
- [ ] Verify all data files generated correctly
- [ ] Document any issues found
- [ ] Update test results in issue tracker