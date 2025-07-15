# ISX Daily Reports Analytics - Alpha Testing Guide

**Version:** Alpha-1.0.0  
**Testing Period:** July 15 - August 15, 2025  
**Target Testers:** 5-10 Alpha Users

---

## 🎯 Alpha Testing Objectives

We need your help to validate:
1. **💾 Installation Process** - Smooth, error-free setup
2. **🔐 License System** - Activation, validation, expiry handling
3. **📊 Data Pipeline** - Scraping, processing, accuracy
4. **🌐 User Interface** - Usability, performance, responsiveness
5. **📁 File Management** - Organization, downloads, integrity
6. **📈 Charts & Visualization** - Accuracy, interactivity, performance
7. **🛠️ Error Handling** - Graceful failures, helpful messages
8. **📖 Documentation** - Clarity, completeness, usefulness

---

## 📋 Pre-Testing Setup

### System Requirements Verification
- [ ] **Windows 10/11** (version check)
- [ ] **8GB+ RAM** (recommended for large datasets)
- [ ] **2GB+ Free Disk Space** 
- [ ] **Google Chrome** (latest version)
- [ ] **Stable Internet** (for ISX portal access)
- [ ] **Administrator Rights** (for installation)

### Alpha Testing License
- [ ] **Received alpha license key** from coordinator
- [ ] **Noted expiry date** (usually 30-90 days)
- [ ] **Confirmed testing period** dates

---

## 🚀 Phase 1: Installation Testing (30 minutes)

### 1.1 Download & Extract
**Test Steps:**
1. Download Alpha release package
2. Extract to temporary folder
3. Verify all files present

**Expected Results:**
- [ ] All files extracted successfully
- [ ] No corruption warnings
- [ ] Folder structure is complete

**Report if:**
- ❌ Download fails or corrupts
- ❌ Missing files in package
- ❌ Extraction errors

### 1.2 Installation Process
**Test Steps:**
1. Right-click PowerShell → "Run as Administrator"
2. Navigate to extracted folder
3. Run: `.\install-alpha.ps1`
4. Follow all prompts

**Expected Results:**
- [ ] ✅ System requirements check passes
- [ ] ✅ Dependencies detected/installed
- [ ] ✅ Installation completes without errors
- [ ] ✅ Desktop shortcut created
- [ ] ✅ PATH variable updated

**Report if:**
- ❌ System requirements check fails incorrectly
- ❌ Installation fails with errors
- ❌ Missing shortcuts or PATH issues
- ❌ Firewall rules not created

### 1.3 First Launch
**Test Steps:**
1. Double-click desktop shortcut "ISX Analytics (Alpha)"
2. Wait for browser to open
3. Observe startup sequence

**Expected Results:**
- [ ] ✅ Application starts without errors
- [ ] ✅ Browser opens to http://localhost:8080
- [ ] ✅ License activation screen appears
- [ ] ✅ No console errors visible

**Report if:**
- ❌ Application fails to start
- ❌ Browser doesn't open automatically
- ❌ Port 8080 conflicts
- ❌ License screen doesn't appear

---

## 🔐 Phase 2: License System Testing (20 minutes)

### 2.1 License Activation
**Test Steps:**
1. Enter your alpha testing license key
2. Click "Activate License"
3. Wait for validation

**Expected Results:**
- [ ] ✅ License validates successfully
- [ ] ✅ "License Valid" status appears
- [ ] ✅ Main interface loads
- [ ] ✅ Expiry date shows correctly

**Report if:**
- ❌ Valid license rejected
- ❌ Activation fails with error
- ❌ Incorrect expiry date shown
- ❌ Interface doesn't load after activation

### 2.2 License Status Display
**Test Steps:**
1. Check license status in footer
2. Navigate between tabs
3. Note renewal warnings (if applicable)

**Expected Results:**
- [ ] ✅ Status consistently displayed
- [ ] ✅ Days remaining accurate
- [ ] ✅ Appropriate warning levels (30/7 days)

**Report if:**
- ❌ Inconsistent status display
- ❌ Incorrect days remaining
- ❌ Missing or excessive warnings

### 2.3 License Expiry Handling
**Test Steps:**
*(If your license is near expiry)*
1. Note behavior as expiry approaches
2. Test functionality with expired license

**Expected Results:**
- [ ] ✅ Clear expiry warnings
- [ ] ✅ Graceful degradation when expired
- [ ] ✅ Clear renewal instructions

---

## 📊 Phase 3: Data Pipeline Testing (45 minutes)

### 3.1 Basic Scraping Test
**Test Steps:**
1. Go to "Data Collection" tab
2. Set date range: **Last 7 days** (smaller test)
3. Mode: **initial**
4. Click "Start Scraping"
5. Monitor console output

**Expected Results:**
- [ ] ✅ Scraping starts without errors
- [ ] ✅ Progress messages appear in console
- [ ] ✅ Excel files downloaded to downloads/
- [ ] ✅ Processing automatically triggered
- [ ] ✅ Index extraction completes
- [ ] ✅ Ticker summary regenerated
- [ ] ✅ UI components refresh automatically

**Report if:**
- ❌ Scraping fails to start
- ❌ Network errors to ISX portal
- ❌ Downloaded files corrupted
- ❌ Processing steps fail
- ❌ UI doesn't auto-refresh

### 3.2 Large Dataset Test
**Test Steps:**
1. Set date range: **Last 30 days**
2. Mode: **initial**
3. Monitor performance and memory usage

**Expected Results:**
- [ ] ✅ Handles large datasets without crashes
- [ ] ✅ Reasonable processing time
- [ ] ✅ Memory usage stays manageable
- [ ] ✅ All files processed successfully

**Report if:**
- ❌ Application crashes with large data
- ❌ Excessive memory usage (>4GB)
- ❌ Processing takes hours (should be minutes)
- ❌ Incomplete or corrupted output

### 3.3 Data Accuracy Verification
**Test Steps:**
1. Download a recent Excel file manually from ISX portal
2. Compare with scraped version
3. Check processed CSV data accuracy
4. Verify ticker summary data

**Expected Results:**
- [ ] ✅ Scraped files identical to manual downloads
- [ ] ✅ CSV data accurate vs Excel source
- [ ] ✅ Ticker summary reflects latest data
- [ ] ✅ Index calculations correct

**Report if:**
- ❌ Data differences between scraped/manual
- ❌ Processing introduces errors
- ❌ Missing or incorrect calculations
- ❌ Date/time discrepancies

---

## 🌐 Phase 4: User Interface Testing (40 minutes)

### 4.1 Tab Navigation & Functionality
**Test Steps:**
1. **Data Collection Tab**:
   - [ ] Form inputs work correctly
   - [ ] Date picker functions
   - [ ] Mode selection works
   - [ ] Console shows real-time updates

2. **Dashboard Tab**:
   - [ ] Index chart loads and displays data
   - [ ] Chart is interactive (zoom, pan, hover)
   - [ ] Data appears current and accurate

3. **Ticker Charts Tab**:
   - [ ] Ticker list loads (should show 82+ tickers)
   - [ ] Search functionality works
   - [ ] Clicking ticker loads chart
   - [ ] Chart interactivity functions

4. **File Archive Tab**:
   - [ ] Files organized in 4 categories
   - [ ] Lists are scrollable
   - [ ] Download buttons work
   - [ ] Refresh updates lists

**Report if:**
- ❌ Any tab fails to load
- ❌ Navigation issues between tabs
- ❌ Missing or broken functionality
- ❌ UI elements don't respond

### 4.2 Responsive Design Testing
**Test Steps:**
1. **Browser Resize**: Drag browser window smaller/larger
2. **Mobile View**: F12 → Device Toolbar → Test mobile sizes
3. **Tablet View**: Test tablet breakpoints

**Expected Results:**
- [ ] ✅ Layout adapts to screen size
- [ ] ✅ File Archive switches to stacked layout
- [ ] ✅ Charts remain functional
- [ ] ✅ All text remains readable

**Report if:**
- ❌ Layout breaks at certain sizes
- ❌ Text becomes unreadable
- ❌ Functionality lost on mobile
- ❌ Horizontal scrollbars appear

### 4.3 Performance & Usability
**Test Steps:**
1. **Load Times**: Note how quickly tabs load
2. **Chart Performance**: Test with large datasets
3. **File Operations**: Download multiple files
4. **Multi-tasking**: Run multiple operations

**Expected Results:**
- [ ] ✅ Tabs load within 2-3 seconds
- [ ] ✅ Charts responsive with 100+ data points
- [ ] ✅ Downloads start immediately
- [ ] ✅ Interface remains responsive during operations

**Report if:**
- ❌ Slow load times (>10 seconds)
- ❌ Charts lag or freeze
- ❌ Download delays or failures
- ❌ Interface becomes unresponsive

---

## 📁 Phase 5: File Management Testing (20 minutes)

### 5.1 File Archive Organization
**Test Steps:**
1. Navigate to File Archive tab
2. Verify file categorization:
   - **Downloaded Files**: Excel reports (.xlsx)
   - **Ticker Reports**: Individual CSV files (A-Z sorted)
   - **Daily Reports**: Processing results (newest first)
   - **System Files**: Summary and index files

**Expected Results:**
- [ ] ✅ Files properly categorized
- [ ] ✅ Sorting follows specified rules
- [ ] ✅ File counts make sense
- [ ] ✅ Icons and badges display correctly

### 5.2 Download Functionality
**Test Steps:**
1. **Download Excel File**: From Downloaded Files section
2. **Download Ticker CSV**: From Ticker Reports section
3. **Download Daily Report**: From Daily Reports section
4. **Download System File**: From System Files section

**Expected Results:**
- [ ] ✅ All downloads start immediately
- [ ] ✅ Files download to browser's download folder
- [ ] ✅ Downloaded files open correctly in appropriate applications
- [ ] ✅ File sizes are reasonable (not 0 bytes or corrupted)

### 5.3 File Integrity
**Test Steps:**
1. Open downloaded Excel files in Microsoft Excel
2. Open CSV files in Excel or text editor
3. Verify data completeness and formatting

**Expected Results:**
- [ ] ✅ Excel files open without errors
- [ ] ✅ CSV files are properly formatted
- [ ] ✅ Data appears complete and accurate
- [ ] ✅ No corruption or encoding issues

---

## 📈 Phase 6: Charts & Visualization Testing (25 minutes)

### 6.1 Index Chart Testing
**Test Steps:**
1. Go to Dashboard tab
2. Wait for index chart to load
3. Test chart interactions:
   - **Zoom**: Mouse wheel
   - **Pan**: Click and drag
   - **Hover**: Mouse over data points
   - **Reset**: Double-click to reset zoom

**Expected Results:**
- [ ] ✅ Chart loads with real data
- [ ] ✅ All interactions work smoothly
- [ ] ✅ Data tooltips are accurate
- [ ] ✅ Chart remains responsive

### 6.2 Ticker Chart Testing
**Test Steps:**
1. Go to Ticker Charts tab
2. Search for "IBSD" (or another active ticker)
3. Click on ticker to load chart
4. Test same interactions as index chart

**Expected Results:**
- [ ] ✅ Ticker chart loads successfully
- [ ] ✅ Candlestick data displays correctly
- [ ] ✅ Volume bars show if available
- [ ] ✅ Date range selector works

### 6.3 Chart Data Accuracy
**Test Steps:**
1. Compare chart data with known market data
2. Verify recent trading days appear
3. Check that weekends/holidays are handled correctly

**Expected Results:**
- [ ] ✅ Data matches external sources
- [ ] ✅ Latest trading day included
- [ ] ✅ Non-trading days excluded appropriately

---

## 🛠️ Phase 7: Error Handling & Edge Cases (30 minutes)

### 7.1 Network Interruption Testing
**Test Steps:**
1. **Start scraping operation**
2. **Disconnect internet** during process
3. **Reconnect** and observe behavior
4. **Retry** operation

**Expected Results:**
- [ ] ✅ Clear error messages about network issues
- [ ] ✅ Graceful handling of interruptions
- [ ] ✅ Successful retry after reconnection
- [ ] ✅ No data corruption from interruptions

### 7.2 Invalid Input Testing
**Test Steps:**
1. **Invalid Date Ranges**: Future dates, invalid formats
2. **Bad License Keys**: Wrong format, expired keys
3. **File System Issues**: Full disk, permission errors

**Expected Results:**
- [ ] ✅ Input validation prevents invalid operations
- [ ] ✅ Clear error messages guide user to fix issues
- [ ] ✅ Application doesn't crash on bad input

### 7.3 System Resource Testing
**Test Steps:**
1. **Low Memory**: Run with limited RAM
2. **Full Disk**: Test with low disk space
3. **Multiple Instances**: Try running twice

**Expected Results:**
- [ ] ✅ Graceful degradation with limited resources
- [ ] ✅ Clear warnings about resource issues
- [ ] ✅ Prevents multiple instances or handles cleanly

---

## 📊 Testing Checklist Summary

### ✅ Must Pass (Critical)
- [ ] Installation completes successfully
- [ ] License activation works
- [ ] Basic scraping and processing functions
- [ ] Web interface loads and is navigable
- [ ] File downloads work
- [ ] No data corruption

### ⚠️ Should Pass (Important)
- [ ] Responsive design works on mobile
- [ ] Chart interactions are smooth
- [ ] Large datasets process reasonably quickly
- [ ] Error messages are helpful
- [ ] Auto-refresh works correctly

### 💡 Nice to Have (Enhancement)
- [ ] Installation is very fast
- [ ] Charts are highly interactive
- [ ] UI is visually appealing
- [ ] Documentation is comprehensive

---

## 📝 Bug Reporting Template

When reporting issues, please use this template:

```markdown
## Bug Report

**Severity:** [Critical/High/Medium/Low]
**Area:** [Installation/License/Scraping/UI/Charts/Files/Other]

### Environment
- Windows Version: 
- RAM: 
- Chrome Version: 
- License Status: 

### Steps to Reproduce
1. 
2. 
3. 

### Expected Behavior
[What should happen]

### Actual Behavior
[What actually happened]

### Screenshots/Logs
[Attach if available]

### Additional Notes
[Any other relevant information]
```

---

## 🎯 Success Criteria

We consider Alpha testing successful if:

### 🟢 **Pass Criteria**
- **95%+ testers** can install successfully
- **90%+ testers** can activate license and use basic features
- **85%+ testers** can complete full data pipeline
- **No critical bugs** that prevent basic usage
- **Performance acceptable** for typical datasets

### 🔴 **Fail Criteria**
- **>20% installation failures** 
- **License system fundamentally broken**
- **Data corruption or loss**
- **Frequent application crashes**
- **Unusable performance** (>10 minutes for basic operations)

---

## 📞 Support & Communication

### 📧 **Primary Contact**
- **Alpha Testing Coordinator**: [Email]
- **Response Time**: Within 24 hours during testing period

### 💬 **Communication Channels**
- **Email**: For detailed bug reports
- **Phone/Chat**: For urgent installation issues
- **Weekly Check-ins**: Progress updates

### 📅 **Timeline**
- **Week 1**: Installation and basic functionality
- **Week 2**: Full feature testing
- **Week 3**: Edge cases and performance
- **Week 4**: Final feedback and wrap-up

---

## 🚀 After Alpha Testing

Based on your feedback, we'll:
1. **Fix critical bugs** identified during testing
2. **Improve documentation** based on user confusion
3. **Enhance UI/UX** based on usability feedback
4. **Optimize performance** for identified bottlenecks
5. **Prepare Beta release** with expanded user base

---

**Thank you for your time and effort in making this system better!** 

Your feedback will directly impact the quality of the final product. 🎯

---

*This testing guide will be updated based on initial feedback. Please check for updates regularly.* 