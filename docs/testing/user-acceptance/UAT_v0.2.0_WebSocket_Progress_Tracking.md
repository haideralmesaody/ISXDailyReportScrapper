# User Acceptance Test: Enhanced Progress Tracking and ETA Estimation
**Version**: v0.2.0-alpha  
**Feature**: WebSocket Message Standardization & Historical ETA  
**Date**: January 2025  
**Document Type**: User Acceptance Test (UAT)

---

## Executive Summary

This document provides step-by-step test scenarios for end users to validate the new progress tracking and ETA estimation features in ISX Daily Reports Scrapper v0.2.0. These improvements provide more accurate time estimates and better progress visibility during data processing.

## What's New in This Version

1. **Smart ETA Estimation**: The application now learns from previous runs to provide immediate time estimates
2. **Enhanced Progress Tracking**: More detailed progress information for each processing stage
3. **Better Error Messages**: Clear, actionable error messages when issues occur
4. **Historical Metrics**: The system remembers processing times to improve future predictions

## Prerequisites

Before starting these tests, ensure you have:
- [ ] ISX Daily Reports Scrapper v0.2.0-alpha installed
- [ ] A valid license key activated
- [ ] Google Chrome browser (recommended) or Firefox/Edge
- [ ] Stable internet connection
- [ ] At least 5GB free disk space
- [ ] 30 minutes available for complete testing

## Test Environment Setup

1. **Verify Installation**
   - Navigate to your ISX Daily Reports Scrapper folder
   - Check that `web-licensed.exe` exists
   - Verify the version by checking the About section in the web interface

2. **Prepare Clean Test Environment** (Optional - for first-run experience)
   - Navigate to `data` folder
   - Rename or delete `metrics` folder if it exists
   - This simulates a fresh installation

## Understanding the Interface

When you open the application, you'll see:
- **Left Sidebar**: Navigation menu with different sections
- **Main Area**: Content area that changes based on selected tab
- **Pipeline Visualization**: Four connected stages showing processing flow
- **Console Output**: Black terminal-style area at bottom showing detailed logs

---

## Test Scenario 1: First-Time User Experience
**Objective**: Verify the system behavior for new users with no historical data  
**Expected Duration**: 10-15 minutes

### Steps:

1. **Start the Application**
   ```
   1. Double-click web-licensed.exe
   2. Wait for "Server started" message
   3. Note: A browser should open automatically
   ```

2. **Access the Web Interface**
   ```
   - If browser didn't open, go to: http://localhost:8080
   - You should see the ISX Daily Reports dashboard
   ```

3. **Initiate Data Download**
   ```
   1. Ensure "ISX Data Collection" tab is selected (should be default)
   2. In the form, set the following:
      - Mode: Select "Initial (Fresh start)"
      - Headless Browser: Select "Yes (Headless)"
      - From Date: Click calendar and select date 5 days ago
      - To Date: Click calendar and select today's date
   3. Click the blue "Start Scraping" button
   ```

4. **Observe Initial Progress**
   
   **What to Look For:**
   - [ ] The "Scraping" stage in the pipeline should activate (turn blue/green)
   - [ ] A progress section appears below the form showing:
     - "ISX Data Download Progress" header
     - Progress bar (striped and animated)
     - Downloaded and Existing file counts
   - [ ] Initial status should show "Initializing..." then "Calculating..."
   - [ ] After the first file downloads, time remaining updates (e.g., "2 minutes")

   **Expected Behavior:**
   ```
   Initial state: "Calculating..."
   After ~30 seconds: "3 minutes remaining" (actual time will vary)
   ```
   
   **Progress Display Details:**
   - Progress shows as percentage (e.g., "25%")
   - Downloaded/Existing shown as numbers
   - Time remaining shown in minutes or seconds
   - Status text updates with current activity

5. **Monitor Progress Updates**
   
   **Check These Elements:**
   - [ ] Progress percentage increases smoothly (shown as "X%")
   - [ ] Downloaded count increases (under "Downloaded" label)
   - [ ] Existing count shows skipped files (under "Existing" label)
   - [ ] Time remaining updates continuously
   - [ ] Status message shows current activity
   - [ ] Console output below shows detailed file information

6. **Verify Stage Transitions**
   
   **Expected Flow:**
   ```
   Scraping (blue) → Scraping (green ✓) → 
   Processing (blue) → Processing (green ✓) → 
   Indices (blue) → Indices (green ✓) → 
   Analysis (blue) → Analysis (green ✓)
   ```

### Success Criteria:
- All stages complete successfully
- No "undefined" or "null" values in progress display
- ETA changes from "Calculating..." to actual time
- All stages show green checkmarks when complete

### Screenshot Requirements:
Please capture:
1. Initial "Calculating..." state
2. Progress with actual ETA
3. Completed pipeline

---

## Test Scenario 2: Returning User Experience
**Objective**: Verify improved ETA accuracy for users with historical data  
**Expected Duration**: 10-15 minutes  
**Prerequisite**: Complete Test Scenario 1 first

### Steps:

1. **Close and Restart Application**
   ```
   1. Close the browser tab
   2. Press Ctrl+C in the command window
   3. Wait 5 seconds
   4. Double-click web-licensed.exe again
   ```

2. **Run Same Date Range Again**
   ```
   1. Navigate to "ISX Data Collection" tab
   2. In the form, set:
      - Mode: "Initial (Fresh start)"  
      - Headless Browser: "Yes (Headless)"
      - From Date: Same as Test 1
      - To Date: Same as Test 1
   3. Click "Start Scraping" button
   ```

3. **Observe Immediate ETA**
   
   **What to Look For:**
   - [ ] ETA shows immediately (not "Calculating...")
   - [ ] ETA includes "(estimated)" label
   - [ ] Estimate is reasonable based on previous run
   
   **Example Display:**
   ```
   "3 minutes remaining (estimated)"
   ```

4. **Verify Estimate Accuracy**
   
   **Track These Times:**
   ```
   Initial Estimate: _______ minutes
   Actual Time Taken: _______ minutes
   Difference: _______ minutes
   ```
   
   **Success Criteria:**
   - Estimate within ±30% of actual time
   - "(estimated)" label disappears once processing starts

5. **Check All Stages**
   
   Each stage should show immediate estimates:
   - [ ] Scraping: Shows estimate immediately
   - [ ] Processing: Shows estimate when stage starts
   - [ ] Indices: Shows estimate when stage starts

### Success Criteria:
- Immediate ETA display (no "Calculating...")
- Reasonable accuracy compared to first run
- Smooth transition from estimated to actual time

---

## Test Scenario 3: Error Handling and Recovery
**Objective**: Verify user-friendly error messages and recovery options  
**Expected Duration**: 5-10 minutes

### Steps:

1. **Simulate Network Error**
   ```
   1. Start a new download (any date range)
   2. After 2-3 files download:
      - Disconnect your internet (Wi-Fi off or unplug cable)
   3. Wait for error message
   ```

2. **Verify Error Message Quality**
   
   **Check for These Elements:**
   - [ ] Error appears in console output (bottom section)
   - [ ] Clear error message with [ERROR] prefix
   - [ ] Specific file that failed
   - [ ] Helpful recovery hint
   - [ ] Pipeline stage may show error state

   **Good Error Example in Console:**
   ```
   [ERROR] Failed to download 2024 01 15 ISX Daily Report.xlsx: network timeout
   [WEBSOCKET_ERROR] Check network connection or file permissions
   ```

3. **Test Recovery**
   ```
   1. Reconnect your internet
   2. Click "Start Scraping" button again
   3. Verify process resumes successfully
   ```

### Success Criteria:
- Error messages are clear and helpful
- No technical jargon in user-facing messages
- Can recover and continue after fixing issue

---

## Test Scenario 4: Data Validation
**Objective**: Ensure data accuracy is maintained with new features  
**Expected Duration**: 5 minutes

### Steps:

1. **After Successful Pipeline Run**
   ```
   1. Click on "Ticker Summary" tab
   2. Verify data loads correctly
   3. Click on any ticker symbol
   ```

2. **Validate Data Display**
   
   **Check:**
   - [ ] All numbers display properly (no "NaN" or "undefined")
   - [ ] Dates are formatted correctly
   - [ ] Charts render without errors
   - [ ] Can download CSV files

### Success Criteria:
- All data displays correctly
- No formatting errors
- Features work as before

---

## Test Scenario 5: Performance Validation
**Objective**: Ensure acceptable performance with progress tracking  
**Expected Duration**: 5 minutes

### Steps:

1. **Run Larger Date Range**
   ```
   1. Navigate to "ISX Data Collection" tab
   2. In the form, set:
      - Mode: "Initial (Fresh start)"
      - From Date: 30 days ago
      - To Date: Today
   3. Click "Start Scraping" button
   ```

2. **Monitor Performance**
   
   **Check:**
   - [ ] Browser remains responsive
   - [ ] Can switch tabs while processing
   - [ ] Progress updates don't freeze
   - [ ] No browser warnings about slow scripts

### Success Criteria:
- UI remains responsive
- No performance degradation
- Smooth progress updates

---

## Test Scenario 6: Browser Compatibility
**Objective**: Verify features work across different browsers  
**Expected Duration**: 15 minutes (5 per browser)

### Test in Each Browser:
1. **Google Chrome** (Recommended)
2. **Mozilla Firefox**
3. **Microsoft Edge**

### Quick Test per Browser:
1. Start application
2. Run small date range (3 days)
3. Verify:
   - [ ] Progress displays correctly
   - [ ] ETA shows and updates
   - [ ] No console errors (F12 → Console)

---

## Post-Test Validation

### Check Metrics Storage
1. Navigate to your installation folder
2. Open `data\metrics\` folder
3. Verify these files exist:
   - `scraping_metrics.json`
   - `processing_metrics.json`
   - `indices_metrics.json`

### Final Checklist
Please confirm:
- [ ] All test scenarios completed
- [ ] No critical issues found
- [ ] Application is faster/same speed as before
- [ ] User experience is improved
- [ ] Would recommend upgrade to other users

---

## Feedback Form

### Overall Experience
**Rate the new progress tracking feature:**
- [ ] Excellent - Significant improvement
- [ ] Good - Noticeable improvement
- [ ] Fair - Some improvement
- [ ] Poor - No improvement or worse

### Specific Feedback

1. **ETA Accuracy**
   - First run: How helpful was seeing "Calculating..."?
   - Second run: How accurate was the estimate?
   
2. **Progress Information**
   - Was the progress information sufficient?
   - Any additional information you'd like to see?

3. **Error Messages**
   - Were error messages clear and helpful?
   - Could you resolve issues based on the hints?

4. **Performance**
   - Did the application feel faster/slower/same?
   - Any lag or freezing noticed?

### Issues Found
Please list any issues encountered:

| Issue Description | Severity (High/Medium/Low) | Steps to Reproduce |
|------------------|---------------------------|-------------------|
| | | |
| | | |

### Suggestions for Improvement
```
[Your suggestions here]
```

---

## Submission

### Test Completion Details
- **Tester Name**: _______________________
- **Test Date**: ________________________
- **Version Tested**: v0.2.0-alpha
- **Total Test Duration**: _____ minutes
- **Overall Result**: ⬜ PASS / ⬜ FAIL

### How to Submit Results
1. Save this document with your results
2. Include any screenshots taken
3. Email to: [support email]
4. Subject: "UAT Results - v0.2.0 Progress Tracking"

---

## Thank You!
Your testing helps us ensure the ISX Daily Reports Scrapper provides the best possible experience for all users. Your feedback is valuable and will be incorporated into future improvements.

## Support
If you encounter any issues during testing:
- Email: [support email]
- Include screenshots and this completed test document
- We'll respond within 24 hours

---
*This document is part of the ISX Daily Reports Scrapper User Acceptance Testing suite.*