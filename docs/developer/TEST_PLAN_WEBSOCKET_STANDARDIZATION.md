# Test Plan: WebSocket Message Standardization

## Objective
Verify that all executables (scraper, processor, indexcsv) properly send standardized WebSocket messages with accurate progress tracking and historical ETA estimation.

## Test Environment
- Windows 10/11
- ISX Daily Reports Scrapper v0.2.0-alpha
- Chrome browser (latest version)
- Valid license file

## Prerequisites
1. Fresh build of all executables using `build.bat`
2. Valid license.dat file in release folder
3. Chrome browser installed
4. At least 10GB free disk space
5. Stable internet connection

## Test Data Setup

### Clean Environment Test
1. Delete `release/data/metrics/` folder (to test first-run scenario)
2. Keep existing downloaded files in `release/data/downloads/` (optional)
3. Clear browser cache and cookies

### Historical Data Test
1. Ensure previous test run completed successfully
2. Verify metrics files exist in `release/data/metrics/`

## Test Cases

### Test Case 1: First-Run Experience (No Historical Data)
**Objective**: Verify behavior when no metrics exist

**Steps**:
1. Start web application:
   ```cmd
   cd release
   web-licensed.exe
   ```
2. Open browser to http://localhost:8080
3. Open Browser Developer Console (F12)
4. Navigate to Network → WS tab to monitor WebSocket messages
5. Click "Download Fresh Data"
6. Set date range: Last 7 days from today
7. Click "Start Download"

**Expected Results**:
- [ ] License validation shows days remaining
- [ ] Pipeline stages show correct initial states
- [ ] Scraping stage activates with "Calculating..." ETA
- [ ] After first file downloads, ETA updates to actual time (e.g., "5 minutes remaining")
- [ ] Progress percentage increases smoothly (not jumpy)
- [ ] WebSocket messages in console show proper format:
  ```json
  {
    "stage": "scraping",
    "current": 1,
    "total": 5,
    "percentage": 20.0,
    "message": "Downloading Excel reports (1 new, 0 existing)",
    "eta": "2 minutes remaining",
    "details": {
      "downloaded": 1,
      "existing": 0,
      "current_page": 1,
      "elapsed": "30s"
    }
  }
  ```
- [ ] All fields are present (no undefined or null in critical fields)
- [ ] Details object contains relevant metadata

**Screenshots Required**:
- Initial "Calculating..." state
- Progress with actual ETA
- WebSocket message format

### Test Case 2: Second-Run Experience (With Historical Data)
**Objective**: Verify historical metrics improve ETA estimation

**Steps**:
1. Complete Test Case 1 successfully
2. Note the actual time taken for each stage
3. Stop web-licensed.exe (Ctrl+C)
4. Restart web-licensed.exe
5. Open browser to http://localhost:8080
6. Run pipeline with same date range as Test Case 1

**Expected Results**:
- [ ] Scraping shows estimated time immediately: "15 minutes remaining (estimated)"
- [ ] Estimate should be within ±20% of actual time from first run
- [ ] "(estimated)" label is visible in ETA
- [ ] Once processing starts, "(estimated)" is removed
- [ ] Final actual time should be close to initial estimate
- [ ] Progress messages show enhanced ETA

**Verification**:
- Compare estimated vs actual times
- Check that estimates improve accuracy

### Test Case 3: Error Handling and Recovery
**Objective**: Verify structured error messages work correctly

**Steps**:
1. Start pipeline normally
2. After 2-3 files download, disconnect network adapter
3. Wait for error to occur
4. Reconnect network
5. Observe error handling

**Expected Results**:
- [ ] Error message appears with proper structure:
  ```json
  {
    "code": "DOWNLOAD_ERROR",
    "message": "Failed to download 2024 01 15 ISX Daily Report.xlsx",
    "details": "Get http://...: dial tcp: lookup ...",
    "stage": "scraping",
    "recoverable": true,
    "hint": "Check network connection or file permissions"
  }
  ```
- [ ] Pipeline stage shows error state (red)
- [ ] Error code is specific (not generic)
- [ ] Hint provides actionable advice
- [ ] Stage is correctly identified
- [ ] Can retry after reconnecting

**Additional Error Tests**:
1. Corrupt Excel file (processor error)
2. Missing index data (indexcsv error)
3. Disk full scenario

### Test Case 4: Full Pipeline Execution
**Objective**: Verify all stages work with standardized messages

**Steps**:
1. Run complete pipeline from fresh start
2. Monitor each stage transition
3. Record timing for each stage
4. Watch for smooth transitions

**Expected Results**:
- [ ] Each stage activates in sequence
- [ ] Progress is accurate for each stage:
  - Scraping: based on file count
  - Processing: based on Excel files
  - Indices: based on files to extract
- [ ] Status messages follow format:
  ```json
  {
    "stage": "processing",
    "status": "completed",
    "message": "Processing completed: 5 files processed in 2.5 minutes"
  }
  ```
- [ ] All stages complete with green checkmarks
- [ ] Final summary shows all metrics
- [ ] No stage gets stuck or shows incorrect status

**Timing Log**:
```
Stage      | Estimated | Actual | Accuracy
-----------|-----------|--------|----------
Scraping   | N/A       | ___min | N/A
Processing | N/A       | ___min | N/A
Indices    | N/A       | ___min | N/A
```

### Test Case 5: Progress Message Metadata
**Objective**: Verify all progress messages include proper metadata

**Steps**:
1. Run pipeline and capture all progress messages
2. Verify each message type has correct metadata

**Expected Results**:

**Scraping Progress**:
- [ ] Contains: downloaded, existing, current_page, elapsed
- [ ] Shows current file being downloaded
- [ ] Page number increments correctly

**Processing Progress**:
- [ ] Contains: current_file, file_date, processed_files, total_files
- [ ] Shows file name being processed
- [ ] Forward-fill phase shows separate progress

**Indices Progress**:
- [ ] Contains: current_file, file_date, indices_extracted, total_files
- [ ] Shows which index values were extracted

### Test Case 6: Metrics Persistence
**Objective**: Verify metrics are saved and loaded correctly

**Steps**:
1. After completing full pipeline run
2. Navigate to `release/data/metrics/`
3. Open each JSON file
4. Verify structure and content

**Expected Results**:
- [ ] Three files exist:
  - `scraping_metrics.json`
  - `processing_metrics.json`
  - `indices_metrics.json`
- [ ] Each file contains valid JSON
- [ ] Structure matches specification:
  ```json
  {
    "stage": "scraping",
    "history": [{
      "timestamp": "2025-01-19T10:30:00Z",
      "total_items": 20,
      "total_duration_seconds": 300.5,
      "avg_per_item_seconds": 15.025
    }],
    "average_times": {
      "per_item": 14.875
    }
  }
  ```
- [ ] Timestamps are valid ISO 8601 format
- [ ] Averages are calculated correctly
- [ ] History array has entries (max 100)

### Test Case 7: Performance and Scalability
**Objective**: Test with larger date ranges

**Steps**:
1. Run pipeline with 30-day date range
2. Monitor performance and accuracy

**Expected Results**:
- [ ] Progress calculations remain accurate
- [ ] ETA adjusts appropriately for larger datasets
- [ ] No memory leaks or crashes
- [ ] UI remains responsive
- [ ] Messages don't flood the WebSocket

### Test Case 8: Browser Compatibility
**Objective**: Verify WebSocket messages work across browsers

**Test Browsers**:
- [ ] Chrome (primary)
- [ ] Firefox
- [ ] Edge

**Expected Results**:
- [ ] Messages display correctly in all browsers
- [ ] Progress bars update smoothly
- [ ] No console errors related to message parsing

## Regression Tests

### Previous Functionality
Verify these still work after changes:
- [ ] File download works correctly
- [ ] Data processing accurate
- [ ] CSV files generated properly
- [ ] UI updates without refresh
- [ ] License validation works

## Performance Benchmarks

Record these metrics for future comparison:

| Metric | Value | Notes |
|--------|-------|-------|
| Scraper startup time | ___s | Time to first progress message |
| Message frequency | ___/s | Messages per second during active processing |
| Memory usage | ___MB | Peak memory during full pipeline |
| CPU usage | ___% | Average during processing |

## Test Summary

### Pass/Fail Criteria
- All test cases must pass
- No regression in existing functionality
- Performance within acceptable limits
- No console errors or warnings

### Known Issues
Document any issues found during testing:

| Issue | Severity | Workaround |
|-------|----------|------------|
| | | |

### Sign-off
- [ ] All tests completed
- [ ] Documentation updated
- [ ] Code reviewed
- [ ] Ready for release

**Tested By**: ________________
**Date**: ____________________
**Version**: v0.2.0-alpha
**Result**: PASS / FAIL

## Appendix: Troubleshooting

### Common Issues During Testing

1. **"Calculating..." never updates**
   - Check if executables are sending messages
   - Verify WebSocket connection is established
   - Check browser console for errors

2. **Estimates wildly inaccurate**
   - Delete metrics folder and rebuild history
   - Check if date ranges are consistent
   - Verify no system time issues

3. **Messages missing fields**
   - Rebuild executables
   - Check progress package is imported correctly
   - Verify JSON marshaling works

4. **WebSocket connection drops**
   - Check firewall settings
   - Verify port 8080 is not blocked
   - Try different browser