# Installation Guide

## System Requirements

- **Operating System**: Windows 10 or later (64-bit)
- **Browser**: Google Chrome (latest version)
- **Memory**: 4GB RAM minimum, 8GB recommended
- **Storage**: 1GB free space for application and data
- **Network**: Internet connection for downloading ISX data

## Quick Installation

### Option 1: Download Release (Recommended)

1. **Download the latest release**
   - Go to [Releases](https://github.com/haideralmesaody/ISXDailyReportScrapper/releases)
   - Download `ISXDailyReportsScrapper-v2.0.0.zip`

2. **Extract the archive**
   - Right-click the ZIP file
   - Select "Extract All..."
   - Choose a location (e.g., `C:\ISXScrapper`)

3. **Run the application**
   - Navigate to the extracted folder
   - Double-click `web-licensed.exe`
   - Your browser will open automatically

4. **Activate license** (first run only)
   - Enter your license key when prompted
   - License is saved for future use

### Option 2: Build from Source

#### Prerequisites
- [Go 1.23+](https://golang.org/dl/)
- [Git](https://git-scm.com/downloads)
- [Google Chrome](https://www.google.com/chrome/)

#### Build Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/haideralmesaody/ISXDailyReportScrapper.git
   cd ISXDailyReportScrapper
   ```

2. **Run the build script**
   ```bash
   .\build.bat
   ```

3. **Navigate to release folder**
   ```bash
   cd release
   ```

4. **Run the application**
   ```bash
   web-licensed.exe
   ```

## Directory Structure

After installation, your directory should look like:

```
ISXScrapper/
├── web-licensed.exe    # Main web application
├── scraper.exe         # Data scraper
├── process.exe         # Data processor  
├── indexcsv.exe        # Index extractor
├── license.dat         # License file (created on activation)
├── data/
│   ├── downloads/      # Downloaded Excel files
│   └── reports/        # Generated CSV reports
├── web/                # Web interface files
└── logs/               # Application logs
```

## First Run Setup

1. **Start the application**
   - Run `web-licensed.exe`
   - Wait for "Server started on :8080"

2. **Access the interface**
   - Browser opens automatically to http://localhost:8080
   - If not, open manually

3. **Activate license**
   - Enter your license key
   - Click "Activate"
   - You'll see "License activated successfully"

4. **Initial data download**
   - Click "Process Files" to start
   - First run downloads all available data
   - May take 10-30 minutes depending on connection

## Configuration

### Google Sheets Integration (Optional)

To enable Google Sheets export:

1. **Obtain credentials**
   - Create a Google Cloud project
   - Enable Google Sheets API
   - Download `credentials.json`

2. **Place credentials file**
   - Copy `credentials.json` to application folder
   - Ensure `sheets-config.json` has your sheet ID

### Customization

Edit `sheets-config.json`:
```json
{
  "spreadsheet_id": "your-sheet-id-here",
  "downloads_folder": "downloads",
  "reports_folder": "reports"
}
```

## Firewall and Antivirus

The application requires:
- Outbound HTTP/HTTPS to ISX website
- Local port 8080 for web interface
- Chrome automation (may trigger antivirus)

### Windows Defender
1. Add folder exclusion for installation directory
2. Allow `web-licensed.exe` through firewall

### Other Antivirus
- Add exception for all `.exe` files
- Whitelist Chrome automation

## Updating

### Automatic Updates
- Application checks for updates on startup
- Click "Update Available" when prompted

### Manual Update
1. Download new release
2. Stop running application
3. Extract new files (overwrites executables)
4. Keep `data/` and `license.dat`
5. Run new version

## Uninstallation

1. Stop the application
2. Delete the installation folder
3. No registry entries or system files are created

## Troubleshooting Installation

### "Port 8080 already in use"
- Another application is using the port
- Close other applications or change port

### "Chrome not found"
- Install Google Chrome
- Ensure Chrome is in default location

### "License activation failed"
- Check internet connection
- Verify license key is correct
- Contact support if issue persists

## Next Steps

After successful installation:
1. Read the [User Guide](README.md)
2. Learn about [Data Specifications](../specifications/DATA_SPECIFICATIONS.md)
3. Set up automated scheduling (Windows Task Scheduler)