# ISX Daily Reports Analytics - Alpha User Guide

**Version:** Alpha-1.0.0  
**Date:** July 15, 2025  
**Status:** Alpha Release - For Testing Only

---

## 🎯 Welcome to Alpha Testing

Thank you for participating in the Alpha testing of ISX Daily Reports Analytics! This guide will help you understand and test all the features of our Iraq Stock Exchange data processing and analysis platform.

> ⚠️ **Alpha Notice**: This is pre-release software. Expect bugs, changes, and potential data loss. Always backup important data.

---

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [System Overview](#system-overview)
3. [Getting Started](#getting-started)
4. [Core Features](#core-features)
5. [Web Interface Guide](#web-interface-guide)
6. [Command Line Interface](#command-line-interface)
7. [License Management](#license-management)
8. [Troubleshooting](#troubleshooting)
9. [Feedback & Bug Reporting](#feedback--bug-reporting)

---

## 🚀 Quick Start

### Installation
1. **Download** the Alpha release package
2. **Extract** to a temporary folder
3. **Run as Administrator**: `install-alpha.ps1`
4. **Follow** the installation prompts
5. **Launch** from desktop shortcut: "ISX Analytics (Alpha)"

### First Run
1. **License Activation**: Enter your alpha testing license key
2. **System Check**: Verify all components are working
3. **Test Scraping**: Try downloading sample data
4. **Explore Interface**: Navigate through all tabs

---

## 🏗️ System Overview

### What Does This System Do?
ISX Daily Reports Analytics is a comprehensive platform for:

- **📊 Data Collection**: Automatically scrape daily reports from ISX portal
- **⚡ Data Processing**: Convert Excel files to structured CSV data
- **📈 Market Analysis**: Extract and visualize market indices (ISX60, ISX15)
- **🎯 Ticker Analysis**: Individual stock analysis with charts
- **📁 File Management**: Organized data storage and retrieval
- **🌐 Web Interface**: Professional dashboard for all operations

### Architecture
```
ISX Portal → Excel Files → CSV Processing → Market Indices → Web Dashboard
     ↓              ↓              ↓              ↓              ↓
  Scraping    File Storage   Data Analysis   Visualization   User Interface
```

---

## 🎬 Getting Started

### 1. Launch the Application
**Desktop Shortcut:**
- Double-click "ISX Analytics (Alpha)"

**Manual Launch:**
- Navigate to installation folder
- Run `Start-Web-Interface.bat`

**Command Line:**
```powershell
# If added to PATH during installation
isxcli

# Or directly
"C:\Program Files\ISX Daily Reports\bin\isx-web-interface.exe"
```

### 2. License Activation (First Time Only)
1. **Browser Opens**: System opens http://localhost:8080
2. **License Screen**: Enter your alpha testing license key
3. **Activation**: Click "Activate License"
4. **Verification**: Wait for "License Valid" confirmation

### 3. Initial Setup Verification
Check that all components show "Ready":
- ✅ License Status: Valid
- ✅ Web Interface: Running
- ✅ File System: Accessible
- ✅ Network: Connected

---

## 🎯 Core Features

### 1. **Automated Data Pipeline**
The system runs a complete pipeline with one click:

```
Scrape → Process → Extract Indices → Update UI
```

**What happens automatically:**
- Downloads Excel files from ISX portal
- Converts to structured CSV data
- Extracts market indices (ISX60, ISX15)
- Updates all charts and data views
- Refreshes file archive

### 2. **Professional Web Interface**
- **4 Main Tabs**: Data Collection, Dashboard, Ticker Charts, File Archive
- **Real-time Updates**: WebSocket-powered live console
- **Responsive Design**: Works on desktop, tablet, mobile
- **Professional Styling**: Iraqi Investor branding

### 3. **Advanced File Management**
- **Organized Categories**: Downloads, Ticker Reports, Daily Reports, System Files
- **Smart Sorting**: Alphabetical for tickers, chronological for daily reports
- **One-click Downloads**: Direct access to all generated files
- **Visual Indicators**: File type badges and icons

### 4. **Interactive Charts**
- **Market Indices**: ISX60 and ISX15 time series
- **Individual Tickers**: Candlestick charts with full interactivity
- **Zoom & Pan**: Mouse wheel zoom, drag to pan
- **Hover Details**: Real-time data on mouseover

---

## 🌐 Web Interface Guide

### Tab 1: Data Collection
**Purpose**: Scrape and process ISX data

**Key Features:**
- **From Date**: Start date for scraping (default: 2025-01-01)
- **To Date**: End date for scraping (default: today)
- **Mode Selection**: 
  - `initial`: Download missing files only
  - `update`: Download latest files
  - `full`: Re-download everything
- **Auto-Processing**: Automatically runs complete pipeline

**Usage:**
1. Set date range (if different from defaults)
2. Select mode (`initial` recommended for first run)
3. Click "Start Scraping"
4. Watch console for real-time progress
5. All data updates automatically

### Tab 2: Dashboard
**Purpose**: View market overview and indices

**Key Features:**
- **Market Overview Cards**: Summary statistics
- **ISX Index Chart**: Interactive time series chart
- **Performance Metrics**: Daily/weekly/monthly changes
- **Data Freshness**: Shows last update time

**Usage:**
- Charts load automatically
- Hover for detailed data points
- Zoom with mouse wheel
- Pan by dragging

### Tab 3: Ticker Charts
**Purpose**: Analyze individual stocks

**Key Features:**
- **Ticker List**: All available stocks (A-Z)
- **Search Functionality**: Quick ticker lookup
- **Interactive Charts**: Candlestick visualization
- **Data Table**: Detailed trading information

**Usage:**
1. **Search**: Type ticker symbol (e.g., "IBSD")
2. **Select**: Click ticker from list
3. **Analyze**: View chart and data
4. **Interact**: Zoom, pan, hover for details

### Tab 4: File Archive
**Purpose**: Browse and download all generated files

**Categories:**
- **Downloaded Files**: Original Excel reports from ISX
- **Ticker Reports**: Individual stock CSV files (A-Z)
- **Daily Reports**: Daily processing results (Recent → Old)
- **System Files**: Summary and index files

**Usage:**
- **Browse**: Scroll through categorized files
- **Select**: Click files to highlight
- **Download**: Click download button for any file
- **Refresh**: Use refresh button to update lists

---

## 💻 Command Line Interface

For advanced users, the CLI provides direct access to all functions:

### Basic Commands
```powershell
# Show help
isxcli --help

# Scrape data
isxcli --mode=initial --from=2025-01-01 --to=2025-07-15

# Process data
isxcli process -in=downloads

# Extract indices
isxcli indexcsv -dir=reports
```

### Advanced Usage
```powershell
# Full pipeline (scrape + process + extract)
isxcli --mode=full --from=2025-01-01 --headless=true
```

---

## 🔐 License Management

### Alpha Testing License
- **Duration**: Typically 30-90 days for alpha testing
- **Features**: Full access to all functionality
- **Expiry**: System shows warnings before expiration
- **Renewal**: Contact alpha testing coordinator

### License Status
- **Valid**: Green status, full functionality
- **Warning**: Yellow status, expiring soon
- **Expired**: Red status, limited functionality

### Troubleshooting License Issues
1. **Check Internet**: License validation requires connectivity
2. **Verify Key**: Ensure correct alpha testing key
3. **Contact Support**: If persistent issues occur

---

## 🛠️ Troubleshooting

### Common Issues

#### 1. **Application Won't Start**
```
Solution:
• Check if port 8080 is available
• Restart as Administrator
• Check Windows Firewall settings
• Verify installation integrity
```

#### 2. **Scraping Fails**
```
Solution:
• Verify internet connectivity to isx-iq.net
• Check Chrome browser installation
• Ensure ISX portal is accessible
• Try different date ranges
```

#### 3. **Charts Don't Load**
```
Solution:
• Ensure data processing completed successfully
• Check if CSV files exist in reports/ folder
• Refresh browser (Ctrl+F5)
• Check console for JavaScript errors
```

#### 4. **License Activation Problems**
```
Solution:
• Verify internet connectivity
• Check license key format
• Ensure alpha testing period is valid
• Contact alpha testing coordinator
```

### Log Files
- **Installation**: `%TEMP%\ISX-Alpha-Install.log`
- **Application**: `C:\Program Files\ISX Daily Reports\logs\`
- **Console Output**: Real-time in web interface

### Performance Tips
- **Memory**: 8GB+ RAM recommended for large datasets
- **Storage**: SSD recommended for faster processing
- **Network**: Stable internet for reliable scraping
- **Browser**: Chrome recommended for best experience

---

## 📝 Feedback & Bug Reporting

### What to Test
1. **Installation Process**: Document any issues
2. **License Activation**: Test activation flow
3. **Data Scraping**: Try different date ranges and modes
4. **Web Interface**: Test all tabs and features
5. **File Operations**: Download and verify files
6. **Charts**: Test interactivity and performance
7. **Error Handling**: Note how system handles errors

### How to Report Issues
**Include the following information:**

1. **System Information**:
   - Windows version
   - RAM/Storage available
   - Chrome version

2. **Steps to Reproduce**:
   - Exact steps taken
   - Expected vs actual behavior
   - Screenshots if applicable

3. **Error Details**:
   - Error messages
   - Log file excerpts
   - Console output

4. **Environment**:
   - License status
   - Data size being processed
   - Network conditions

### Contact Information
- **Alpha Testing Coordinator**: [Contact details]
- **Technical Support**: [Contact details]
- **Feedback Email**: [Contact details]

---

## 📊 Alpha Testing Goals

Help us verify:
- ✅ **Installation Process**: Smooth and error-free
- ✅ **User Experience**: Intuitive and professional
- ✅ **Data Accuracy**: Correct processing and calculations
- ✅ **Performance**: Acceptable speed with real data
- ✅ **Stability**: No crashes or data corruption
- ✅ **Documentation**: Clear and helpful guides

---

## 🎯 Next Steps After Alpha

Based on your feedback, we'll:
1. **Fix Critical Bugs**: Address any blocking issues
2. **Improve UX**: Enhance user interface based on feedback
3. **Optimize Performance**: Address speed and memory issues
4. **Add Features**: Implement requested functionality
5. **Beta Release**: Prepare for wider testing

---

**Thank you for helping us build a better ISX analytics platform!** 🚀

---

*This guide is updated regularly. Check for the latest version in your installation directory.* 