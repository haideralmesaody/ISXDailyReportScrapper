# ISX Daily Reports Scrapper

A comprehensive data collection and analysis system for Iraq Stock Exchange (ISX) daily trading reports. This application automates the process of downloading, parsing, processing, and visualizing ISX market data.

## Features

- **Automated Data Collection**: Downloads daily Excel reports from ISX website
- **Data Processing**: Parses Excel files and converts to structured CSV format
- **Historical Data Management**: Maintains complete trading history with forward-filling for non-trading days
- **Market Indices**: Extracts and tracks ISX60 and ISX15 index values
- **Web Interface**: Modern, responsive web dashboard for data visualization and management
- **Real-time Updates**: Automatic UI updates when data changes (no manual refresh needed)
- **Ticker Analytics**: Generates summary statistics for all traded securities
- **License Management**: Built-in licensing system for commercial deployment

## Project Structure

```
ISXDailyReportsScrapper/
├── dev/                        # Development source code
│   ├── cmd/                    # Command-line applications
│   │   ├── indexcsv/          # Index extractor 
│   │   ├── process/           # Data processor
│   │   └── web-licensed/      # Web application
│   ├── internal/              # Internal packages
│   │   ├── analytics/         # Data analysis functions
│   │   ├── common/            # Shared utilities (logger, paths, errors)
│   │   ├── exporter/          # CSV export utilities
│   │   ├── files/             # File management
│   │   ├── license/           # License management
│   │   ├── parser/            # Excel parsing
│   │   ├── pipeline/          # Pipeline manager and orchestration
│   │   ├── processor/         # Data processing
│   │   ├── progress/          # Progress tracking and WebSocket messages
│   │   ├── updater/           # Auto-update functionality
│   │   ├── watcher/           # File system monitoring
│   │   └── websocket/         # Real-time communication
│   └── web/                   # Web interface assets
├── release/                    # Production-ready binaries
│   ├── data/                  # Data storage
│   │   ├── downloads/         # Downloaded Excel files
│   │   └── reports/           # Generated CSV reports
│   └── web/                   # Web interface files
└── build.bat                  # Build script for Windows

```

## Installation

### Prerequisites

- Go 1.23+ (for building from source)
- Windows OS (primary support)
- Chrome browser (for web scraping)

### Quick Start

1. **Download the latest release** from the releases page
2. **Extract** to your desired location
3. **Run** `web-licensed.exe` to start the web interface
4. **Access** http://localhost:8080 in your browser

### Building from Source

```bash
# Clone the repository
git clone https://github.com/haideralmesaody/ISXDailyReportScrapper.git
cd ISXDailyReportScrapper

# Build all components
.\build.bat
```

The build script will create a `release/` folder with all executables and required files.

## Usage

### Web Interface (Recommended)

1. Start the web application:
   ```bash
   cd release
   web-licensed.exe
   ```

2. Open http://localhost:8080 in your browser

3. Use the interface to:
   - Run data scraping and processing
   - View ticker summaries
   - Browse historical data
   - Export reports

### Command Line Tools

#### Scraper
Downloads Excel files from ISX website:
```bash
scraper.exe --mode initial --from 2024-01-01 --to 2024-12-31
```

#### Data Processor
Processes Excel files and generates CSV reports:
```bash
process.exe
```

#### Index Extractor
Extracts market index values:
```bash
indexcsv.exe
```

## Data Formats

### Input
- Excel files from ISX website (`.xlsx` format)
- Contains daily trading data for all listed securities

### Output
All data follows strict specifications (see `DATA_SPECIFICATIONS.md`):

- **Daily Reports**: `isx_daily_YYYY_MM_DD.csv`
- **Combined Data**: `isx_combined_data.csv` 
- **Ticker Summaries**: `ticker_summary.json`
- **Individual Tickers**: `{SYMBOL}_trading_history.csv`
- **Market Indices**: `indexes.csv`

## Real-time Updates

The application includes a comprehensive real-time update system:

### Live Progress Tracking
- **Real-time Progress**: See live updates for each pipeline stage
- **Smart ETA Estimation**: 
  - Initial estimates based on historical data from previous runs
  - Estimates improve with each use of the application
  - Shows "(estimated)" label for predictions based on history
  - Switches to actual timing once processing begins
- **Detailed Status**: View current file being processed, records found, etc.
- **Error Recovery**: Clear error messages with suggested recovery actions

### Automatic UI Updates
- File changes are automatically detected
- WebSocket broadcasts updates to all connected clients
- UI components refresh without manual intervention
- No page refresh required

### Progress Accuracy
- First run shows "Calculating..." until timing data is available
- Subsequent runs show immediate time estimates
- Accuracy improves over time as more historical data is collected
- Metrics are stored locally for consistent predictions

## Configuration

### License Activation

On first run, you'll need to activate your license:
1. Start the web application
2. Enter your license key when prompted
3. License is validated and stored locally


## Data Specifications

See the following documents for detailed specifications:
- `DATA_SPECIFICATIONS.md` - Complete data format specifications
- `COLUMN_NAME_MAPPING.md` - Column name transformations
- `DATA_VALIDATION_CHECKLIST.md` - Data validation guidelines

## Development

### Project Dependencies

Key Go packages used:
- `github.com/xuri/excelize/v2` - Excel file parsing
- `github.com/chromedp/chromedp` - Web scraping
- `github.com/gorilla/mux` - HTTP routing
- `github.com/gorilla/websocket` - WebSocket support
- `github.com/fsnotify/fsnotify` - File system monitoring

### Contributing

1. Fork the repository
2. Create your feature branch
3. Ensure all data formats follow specifications
4. Test with `go test ./...`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, feature requests, or questions:
- Open an issue on GitHub
- Contact: [your-email@example.com]

## Acknowledgments

- Iraq Stock Exchange for providing public market data
- All contributors and users of this project