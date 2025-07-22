# ISX Daily Reports Scraper

A comprehensive web scraping and analytics platform for the Iraqi Stock Exchange (ISX) daily trading reports. This application automates the download, processing, and visualization of stock market data from the ISX portal.

## Features

### 🚀 Core Functionality
- **Automated Scraping**: Downloads daily reports from ISX portal using ChromeDP
- **Data Processing**: Processes Excel files and converts to CSV format
- **Web Interface**: Modern web dashboard with real-time updates via WebSocket
- **Ticker Analytics**: Individual stock analysis with candlestick charts
- **Search & Filtering**: Sortable ticker tables with search functionality

### 📊 Visualization
- **Professional Charts**: Highcharts-powered candlestick charts with volume
- **Interactive Tables**: Sortable columns for all ticker data
- **Real-time Updates**: WebSocket-based live data updates
- **Responsive Design**: Mobile-friendly interface

### 🛠 Technical Features
- **Go Backend**: High-performance web server with concurrent processing
- **Chrome Automation**: Headless browser automation for reliable scraping
- **Excel Processing**: Advanced Excel file parsing with Excelize library
- **Trading Days Filter**: Excludes non-trading days from analysis
- **Individual Ticker Data**: Accurate last trading dates per ticker

## Installation

### Prerequisites
- Go 1.19 or later
- Chrome/Chromium browser (for scraping)
- Git

### Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ISXDailyReportsScrapper.git
   cd ISXDailyReportsScrapper
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the applications:
   ```bash
   # Build CLI scraper
   go build -o scraper.exe cmd/scraper/main.go
   
   # Build web interface
   go build -o web.exe cmd/web/main.go
   
   # Build processing tools
   go build -o process.exe cmd/process/main.go
   go build -o indexcsv.exe cmd/indexcsv/main.go
   ```

## Usage

### Command Line Interface

#### Scraper
```bash
# Download latest daily report
./scraper.exe

# Download specific date
./scraper.exe -date=2024-01-15

# Download date range
./scraper.exe -start=2024-01-01 -end=2024-01-31
```

#### Data Processing
```bash
# Process downloaded Excel files to CSV
./process.exe

# Create ticker summaries and indices
./indexcsv.exe
```

### Web Interface

1. Start the web server:
   ```bash
   ./web.exe
   ```

2. Open browser to: `http://localhost:8080`

3. Features available:
   - **Dashboard**: Overview of market data
   - **Ticker Charts**: Individual stock analysis
   - **File Archive**: Downloaded and processed files
   - **Real-time Console**: Live operation updates

## Project Structure

```
ISXDailyReportsScrapper/
├── cmd/                    # Application entry points
│   ├── scraper/           # CLI scraper application
│   ├── web/               # Web server application
│   ├── process/           # Data processing tool
│   └── indexcsv/          # CSV indexing tool
├── internal/              # Internal packages
│   ├── scraper/          # Scraping logic
│   ├── processor/        # Data processing
│   └── utils/            # Utility functions
├── web/                   # Web interface files
│   ├── index.html        # Main web interface
│   └── static/           # Static assets
├── reports/              # Generated CSV files (gitignored)
├── downloads/            # Downloaded files (gitignored)
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # This file
```

## Configuration

### Environment Variables
- `ISX_HEADLESS`: Set to `false` to run Chrome with GUI (default: `true`)
- `ISX_PORT`: Web server port (default: `8080`)
- `ISX_TIMEOUT`: Scraping timeout in seconds (default: `60`)

### Advanced Options
- Modify `internal/scraper/config.go` for scraping parameters
- Adjust chart styling in `web/index.html` CSS variables
- Configure data processing rules in `internal/processor/`

## API Endpoints

### Web Interface
- `GET /` - Main dashboard
- `GET /api/tickers` - Get ticker summaries
- `GET /api/ticker/{symbol}` - Get individual ticker data
- `WebSocket /ws` - Real-time updates

## Development

### Adding New Features
1. Core logic goes in `internal/` packages
2. CLI commands in `cmd/` directories
3. Web interface updates in `web/index.html`
4. Follow Go conventions and add tests

### Building for Production
```bash
# Build all components
go build -ldflags="-s -w" -o scraper.exe cmd/scraper/main.go
go build -ldflags="-s -w" -o web.exe cmd/web/main.go
go build -ldflags="-s -w" -o process.exe cmd/process/main.go
go build -ldflags="-s -w" -o indexcsv.exe cmd/indexcsv/main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/new-feature`
5. Submit a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [ChromeDP](https://github.com/chromedp/chromedp) for browser automation
- [Excelize](https://github.com/qax-os/excelize) for Excel file processing
- [Highcharts](https://www.highcharts.com/) for professional charting
- [Bootstrap](https://getbootstrap.com/) for responsive UI components

## Disclaimer

This tool is for educational and analytical purposes only. Users are responsible for complying with the Iraqi Stock Exchange's terms of service and applicable regulations when scraping data.

---

**The Iraqi Investor** - Professional ISX Analytics Platform 