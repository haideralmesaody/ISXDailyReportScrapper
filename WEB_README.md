# ISX Daily Reports Scraper - Web Interface

A modern, single-page web application that provides a graphical interface for all the ISX Daily Reports Scraper CLI tools.

## Features

- **Modern Web UI**: Clean, responsive interface built with Bootstrap 5
- **Real-time Output**: WebSocket-powered live command output and progress updates
- **File Management**: Browse, download, and manage scraped files and generated reports
- **All CLI Tools**: Web forms for every command with proper validation
- **Connection Status**: Live connection indicator and automatic reconnection
- **Cross-platform**: Works on Windows, macOS, and Linux

## Quick Start

### 1. Build Everything
```bash
# On Windows
build-web.bat

# On macOS/Linux
chmod +x build-web.sh
./build-web.sh
```

### 2. Start the Web Server
```bash
./web.exe
```

### 3. Open Your Browser
Navigate to: `http://localhost:8080`

## Available Commands in Web Interface

### 📥 Data Scraping
- **Mode**: Initial (fresh start) or Accumulative (incremental updates)
- **Date Range**: Specify from/to dates
- **Browser**: Choose headless or visible browser mode
- **Real-time Progress**: Watch scraping progress live

### ⚙️ Process Data
- **Parse Excel Files**: Extract trading records from downloaded reports
- **Input Directory**: Specify which directory to process
- **Console Output**: View parsed records and statistics

### 📊 Extract Indices
- **Market Indices**: Extract ISX60 and ISX15 index values
- **Mode**: Initial or accumulative processing
- **CSV Output**: Generate time-series data for analysis

### 🔍 Market Scan
- **Market Statistics**: Extract comprehensive market metrics
- **Sampling**: Configure date ranges and sampling intervals
- **Comprehensive Data**: Listed companies, trading volumes, suspended stocks, etc.

### 📋 Combine Data
- **Merge Datasets**: Combine indices with market summary data
- **Single Dataset**: Create unified CSV for analysis
- **Custom Files**: Specify input and output file names

### 👁️ Inspect Files
- **File Analysis**: Examine Excel file structure and contents
- **Sheet Information**: View sheet names, headers, and sample data
- **Debug Tool**: Useful for troubleshooting parsing issues

### 📁 File Manager
- **Browse Files**: View all downloaded and generated files
- **Download Files**: One-click download of any file
- **File Organization**: Separate views for raw data and processed results

## API Endpoints

The web interface exposes REST API endpoints that can be used programmatically:

```bash
# Scraping
POST /api/scrape
{
  "args": {
    "mode": "initial",
    "from": "2024-01-01",
    "headless": "true"
  }
}

# Processing
POST /api/process
{
  "args": {
    "in": "downloads"
  }
}

# File listing
GET /api/files

# File download
GET /api/download/{filename}

# Server status
GET /api/status
```

## WebSocket Connection

Real-time updates are delivered via WebSocket at `/ws`:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log(message.type, message.message);
};
```

Message types:
- `info`: General information
- `success`: Command completed successfully
- `error`: Command failed or error occurred
- `output`: Raw command output

## Configuration

### Port Configuration
The web server runs on port 8080 by default. To change this, modify the `main.go` file in `cmd/web/`:

```go
log.Fatal(http.ListenAndServe(":8080", r))
```

### File Paths
The web interface expects CLI executables in specific locations:
- `./isxcli.exe` - Main scraper
- `./cmd/process/process.exe` - Process tool
- `./cmd/indexcsv/indexcsv.exe` - Index extraction
- `./cmd/marketscan/marketscan.exe` - Market scanning
- `./cmd/combine/combine.exe` - Data combining
- `./cmd/inspect/inspect.exe` - File inspection

## Security Considerations

⚠️ **Important**: This web interface is designed for local development use.

For production deployment:
1. Add authentication and authorization
2. Implement HTTPS/TLS
3. Add input validation and sanitization
4. Restrict file access and downloads
5. Add rate limiting
6. Configure proper CORS policies

## Troubleshooting

### Connection Issues
- Ensure the web server is running on port 8080
- Check that no firewall is blocking the connection
- Verify WebSocket connection in browser developer tools

### Command Failures
- Check that all CLI executables are built and in the correct locations
- Verify file permissions on executables
- Review console output for specific error messages

### File Access Issues
- Ensure the `downloads` directory exists and is writable
- Check file permissions for generated CSV/JSON files
- Verify paths are correct for your operating system

## Development

### Adding New Commands
1. Add the command handler in `cmd/web/main.go`
2. Create the API endpoint
3. Add the form section to `web/index.html`
4. Update the JavaScript to handle the new command

### Customizing the UI
- Edit `web/index.html` for layout and styling
- Modify the CSS in the `<style>` section
- Update JavaScript for new functionality

### Building for Different Platforms
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o web.exe ./cmd/web

# macOS
GOOS=darwin GOARCH=amd64 go build -o web ./cmd/web

# Linux
GOOS=linux GOARCH=amd64 go build -o web ./cmd/web
```

## License

Same as the main project (MIT). 