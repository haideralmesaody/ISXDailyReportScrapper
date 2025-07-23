# Regression Test Plan: Pipeline Orchestration

**Version**: v0.3.0-alpha  
**Component**: Pipeline Orchestration (PIPE-002 Fix)  
**Date**: January 2025  
**Document Type**: Regression Test Plan (REG)

## Executive Summary

This regression test plan ensures that the Pipeline Manager implementation (PIPE-002) fixes the critical bug where the pipeline would stop after the scraping stage, while maintaining all existing functionality. The tests verify that no new issues are introduced and existing features continue to work as expected.

## Test Objectives

1. **Verify Bug Fix**: Confirm pipeline no longer stops after scraping
2. **Backward Compatibility**: Ensure existing features still work
3. **API Compatibility**: Verify HTTP endpoints maintain same interface
4. **WebSocket Compatibility**: Confirm message format unchanged
5. **Performance**: No significant performance degradation

## Bug Reproduction Test

### Test R1: Original Bug Reproduction

**Purpose**: Reproduce the original bug to confirm it's fixed

**Steps to Reproduce Original Bug** (v0.2.0):
1. Click "Update Data" button with date range
2. Observe scraping stage completes successfully
3. Note that processing stage never starts
4. Check logs showing "Scraper completed with success=true"
5. Pipeline remains stuck, no further progress

**Original Bug Symptoms**:
- Last log: "[SUMMARY] ====== Download Complete ======"
- No processing stage activation
- WebSocket stops sending updates
- UI shows scraping complete but nothing else

**Test with v0.3.0**:
1. Deploy v0.3.0 with Pipeline Manager
2. Follow same reproduction steps
3. **Expected Result**: All 4 stages execute automatically

**Verification Checklist**:
- ✓ Scraping completes
- ✓ Processing starts automatically
- ✓ Indices extraction follows
- ✓ Analysis stage completes
- ✓ "Pipeline complete" message received

## Backward Compatibility Tests

### Test R2: HTTP API Compatibility

**Purpose**: Ensure API endpoints work identically

#### Test R2.1: /api/scrape Endpoint

**Request Format** (must be unchanged):
```json
POST /api/scrape
{
  "args": {
    "from": "2025-01-01",
    "to": "2025-01-20",
    "mode": "initial"
  }
}
```

**Response Format** (must be unchanged):
```json
{
  "success": true,
  "output": "Pipeline started successfully",
  "error": ""
}
```

**Test Steps**:
1. Send request with curl/Postman
2. Verify response format
3. Check backward compatibility

#### Test R2.2: /api/process Endpoint

**Verify**:
- Endpoint still exists
- Accepts same parameters
- Returns same response format

#### Test R2.3: /api/system/stats

**Verify**:
- Stats include pipeline information
- Format unchanged from v0.2.0

### Test R3: WebSocket Message Compatibility

**Purpose**: Ensure frontend doesn't break

#### Test R3.1: Progress Message Format

**v0.2.0 Format** (must be maintained):
```json
{
  "type": "pipeline_progress",
  "stage": "scraping",
  "status": "active",
  "progress": 45.5,
  "message": "Downloading file 5 of 10",
  "metadata": {
    "current_file": "2025-01-05.xlsx",
    "total_files": 10
  }
}
```

**Test Verification**:
1. Connect WebSocket client
2. Start pipeline
3. Capture all messages
4. Verify format matches exactly

#### Test R3.2: Status Message Compatibility

**Required Messages**:
- pipeline_status (stage activation)
- pipeline_progress (progress updates)
- pipeline_complete (completion)
- pipeline_error (on failure)

**Test Each Message Type**:
```javascript
// Frontend code that must continue working
socket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    switch(data.type) {
        case 'pipeline_status':
            updateStageStatus(data.stage, data.status);
            break;
        case 'pipeline_progress':
            updateProgress(data.stage, data.progress, data.message);
            break;
        // ... etc
    }
}
```

### Test R4: Data Output Compatibility

**Purpose**: Ensure output files remain identical

#### Test R4.1: CSV Format Verification

**Files to Check**:
1. `data/reports/isx_combined_data.csv`
   - Headers unchanged
   - Column order same
   - Data types consistent

2. `data/reports/indexes.csv`
   - Date,ISX60,ISX15 format
   - No additional columns

3. `data/reports/ticker_summary.json`
   - JSON structure unchanged
   - Field names still snake_case
   - No missing fields

**Test Method**:
1. Run pipeline with v0.2.0 (save outputs)
2. Run pipeline with v0.3.0 (same data)
3. Diff the output files
4. Should be identical (except timestamps)

### Test R5: Configuration Compatibility

**Purpose**: Verify existing configs still work

**Test Cases**:
1. Command line arguments
2. Environment variables
3. License file location
4. Directory structures

**Verification**:
- No new required configuration
- Defaults unchanged
- Paths remain relative

## Performance Regression Tests

### Test R6: Execution Time Comparison

**Purpose**: Ensure no significant slowdown

**Test Setup**:
- Same dataset (30 days of data)
- Same machine
- Average of 3 runs each

**Metrics to Compare**:
| Stage | v0.2.0 Time | v0.3.0 Time | Acceptable Delta |
|-------|-------------|-------------|------------------|
| Scraping | ~5 min | Should be ≤5.5 min | +10% |
| Processing | ~2 min | Should be ≤2.2 min | +10% |
| Indices | ~30 sec | Should be ≤33 sec | +10% |
| Analysis | ~1 min | Should be ≤1.1 min | +10% |
| **Total** | ~8.5 min | Should be ≤9.5 min | +12% |

### Test R7: Memory Usage

**Purpose**: Check for memory leaks

**Test Method**:
1. Monitor memory before pipeline
2. Run pipeline 5 times consecutively
3. Check memory after each run
4. Should return to baseline

**Acceptable Criteria**:
- Peak memory < 2GB
- No cumulative growth
- Cleanup after completion

### Test R8: Resource Usage

**Purpose**: Verify resource efficiency

**Monitor**:
- CPU usage patterns
- Disk I/O
- Network usage (scraping)
- File handles

**Expected**:
- Similar patterns to v0.2.0
- No resource leaks
- Proper cleanup

## Frontend Compatibility Tests

### Test R9: UI Functionality

**Purpose**: Ensure UI works without changes

**Test Areas**:
1. **Pipeline Control**
   - Start button works
   - Date pickers function
   - Mode selection works

2. **Progress Display**
   - Progress bars update
   - Stage indicators change color
   - Messages display correctly

3. **Data Display**
   - Charts load after completion
   - Ticker list updates
   - File list refreshes

**Test Method**:
- Use existing frontend (no changes)
- Should work identically with v0.3.0 backend

### Test R10: Real-time Updates

**Purpose**: Verify WebSocket updates work

**Test Scenarios**:
1. Open multiple browser tabs
2. Start pipeline in one tab
3. All tabs should update simultaneously
4. Progress should be smooth

## Error Handling Regression

### Test R11: Error Scenarios

**Purpose**: Ensure error handling unchanged

**Test Cases**:

1. **Network Error During Scraping**
   - Disconnect network mid-download
   - Should show clear error message
   - Should not crash

2. **Invalid Date Range**
   - Use future dates
   - Should reject with message
   - Same as v0.2.0

3. **Missing License**
   - Remove license.dat
   - Should show license error
   - Same behavior as before

4. **Corrupted Excel File**
   - Place corrupted file in downloads
   - Processing should handle gracefully
   - Error message should be clear

## Integration Tests

### Test R12: External Process Integration

**Purpose**: Verify executables called correctly

**Test Points**:
1. scraper.exe receives correct arguments
2. process.exe finds input files
3. indexcsv.exe creates output
4. Working directory set correctly

**Verification Method**:
- Add logging to executables
- Verify arguments match v0.2.0
- Check working directory

### Test R13: File System Operations

**Purpose**: Ensure file operations unchanged

**Test Cases**:
1. Directory creation
2. File permissions
3. Path resolution
4. Cleanup operations

## Stress Testing

### Test R14: Concurrent Requests

**Purpose**: Verify stability under load

**Test Scenario**:
1. Send 3 pipeline requests rapidly
2. First should execute
3. Others should queue or reject
4. No crashes or corruption

### Test R15: Large Dataset

**Purpose**: Test with maximum data

**Test Setup**:
- 2 years of historical data
- ~500 Excel files
- Run complete pipeline

**Success Criteria**:
- Completes successfully
- Memory stays under control
- All files processed

## Test Execution Checklist

### Pre-Test Setup
- [ ] Backup v0.2.0 installation
- [ ] Document current behavior
- [ ] Prepare test datasets
- [ ] Setup monitoring tools

### Critical Path Tests (Must Pass)
- [ ] R1: Bug fix verification
- [ ] R2: API compatibility
- [ ] R3: WebSocket compatibility
- [ ] R4: Data output compatibility
- [ ] R9: UI functionality

### Important Tests
- [ ] R6: Performance comparison
- [ ] R7: Memory usage
- [ ] R11: Error handling
- [ ] R13: File operations

### Nice to Have
- [ ] R14: Concurrent requests
- [ ] R15: Large dataset

## Rollback Plan

If critical regressions found:

1. **Immediate Rollback**
   - Restore v0.2.0 executables
   - Document issues found
   - No data migration needed

2. **Partial Rollback**
   - Keep bug fix
   - Revert other changes
   - Patch specific issues

3. **Forward Fix**
   - Address issues in v0.3.1
   - Maintain compatibility
   - Quick turnaround

## Test Results Template

```
Test ID: R1
Test Name: Original Bug Reproduction
Date: [Date]
Tester: [Name]
Result: PASS/FAIL
Notes: [Observations]
Evidence: [Screenshots/Logs]
```

## Success Criteria

### Must Have (v0.3.0 Release)
- Pipeline completes all stages (bug fixed)
- No API breaking changes
- No WebSocket format changes
- No data format changes
- No performance regression >20%

### Should Have
- Memory usage improved or same
- Error messages unchanged
- UI works without modification
- All edge cases handled

### Could Have
- Performance improvements
- Better error messages
- Enhanced logging

## Appendix: Quick Regression Script

```bash
#!/bin/bash
# Quick regression test script

echo "=== Pipeline Orchestration Regression Test ==="

# Test 1: API Compatibility
echo "Test R2: Testing API compatibility..."
curl -X POST http://localhost:8080/api/scrape \
  -H "Content-Type: application/json" \
  -d '{"args":{"from":"2025-01-01","to":"2025-01-10"}}' \
  > api_response.json

# Verify response format
if grep -q '"success":true' api_response.json; then
  echo "✓ API response format correct"
else
  echo "✗ API response format changed!"
fi

# Test 2: Watch WebSocket messages
echo "Test R3: Monitoring WebSocket messages..."
# (WebSocket monitoring script would go here)

# Test 3: Check output files
echo "Test R4: Verifying output files..."
for file in "isx_combined_data.csv" "indexes.csv" "ticker_summary.json"; do
  if [ -f "data/reports/$file" ]; then
    echo "✓ $file exists"
  else
    echo "✗ $file missing!"
  fi
done

echo "=== Regression Test Complete ==="
```

---

*This regression test plan ensures the Pipeline Manager implementation fixes the critical bug while maintaining full backward compatibility with existing systems and interfaces.*