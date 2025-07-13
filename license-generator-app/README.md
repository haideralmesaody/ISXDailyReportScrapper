# ISX License Generator

A standalone application for generating ISX recharge card licenses that can be distributed and activated by end users.

## 🎯 Features

- **Standalone Application**: Works independently from the main ISX software
- **Recharge Card Model**: Generate licenses like prepaid cards
- **Google Sheets Integration**: Automatic license tracking and management
- **Bulk Generation**: Generate hundreds of licenses at once
- **Multiple Durations**: 1-month, 3-month, 6-month, and 1-year licenses
- **Export Options**: Save licenses to text files
- **Professional UI**: Clean command-line interface with progress indicators

## 📋 Requirements

- Go 1.19 or higher
- Google Sheets API access
- Internet connection for API calls

## 🚀 Quick Start

### 1. Setup Google Sheets API

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google Sheets API
4. Create API credentials (API Key)
5. Update `license-config.json` with your API key

### 2. Prepare Your Google Sheet

Your Google Sheet needs these columns in Row 1:

| **A** | **B** | **C** | **D** | **E** | **F** |
|-------|-------|-------|-------|-------|-------|
| **LicenseKey** | **Duration** | **ExpiryDate** | **Status** | **MachineID** | **ActivatedDate** |

Make sure your sheet is named "Licenses" and is publicly accessible (read-only).

### 3. Configure the Application

Edit `license-config.json`:
```json
{
  "sheet_id": "1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc",
  "api_key": "YOUR_ACTUAL_API_KEY_HERE",
  "sheet_name": "Licenses"
}
```

### 4. Build and Run

```bash
# Build the application
go build -o license-generator.exe

# Generate 100 random licenses
./license-generator.exe -total=100

# Generate specific quantities
./license-generator.exe -1m=25 -3m=50 -6m=20 -1y=5

# Save to file
./license-generator.exe -total=50 -output=licenses.txt
```

## 💡 Usage Examples

### Generate Mixed Licenses
```bash
# Generate 30 3-month and 20 6-month licenses
./license-generator.exe -3m=30 -6m=20
```

### Generate for Distribution
```bash
# Generate 100 licenses and save to file
./license-generator.exe -total=100 -output=my-licenses.txt
```

### Generate Specific Batches
```bash
# Generate different durations
./license-generator.exe -1m=10 -3m=20 -6m=15 -1y=5
```

### Custom Configuration
```bash
# Use different config file
./license-generator.exe -config=my-config.json -total=50
```

## 🎫 License Types

The generator creates licenses with these formats:

- **1-month**: `ISX1M-ABC123-456789`
- **3-month**: `ISX3M-DEF456-789012`
- **6-month**: `ISX6M-GHI789-012345`
- **1-year**: `ISX1Y-JKL012-345678`

## 📊 Command Line Options

| Option | Description | Example |
|--------|-------------|---------|
| `-1m` | Number of 1-month licenses | `-1m=25` |
| `-3m` | Number of 3-month licenses | `-3m=50` |
| `-6m` | Number of 6-month licenses | `-6m=20` |
| `-1y` | Number of 1-year licenses | `-1y=5` |
| `-total` | Total random licenses | `-total=100` |
| `-output` | Output file for licenses | `-output=licenses.txt` |
| `-config` | Configuration file path | `-config=my-config.json` |

## 🔧 Configuration File

The `license-config.json` file contains:

```json
{
  "sheet_id": "YOUR_GOOGLE_SHEET_ID",
  "api_key": "YOUR_GOOGLE_SHEETS_API_KEY",
  "sheet_name": "Licenses"
}
```

### Getting Your Sheet ID
From your Google Sheet URL:
`https://docs.google.com/spreadsheets/d/SHEET_ID_HERE/edit`

### Getting Your API Key
1. Google Cloud Console → APIs & Services → Credentials
2. Create Credentials → API Key
3. (Optional) Restrict key to Google Sheets API

## 📱 Output Examples

### Console Output
```
🎫 ISX License Generator v1.0
═══════════════════════════════

🔄 Generating 50 licenses for 3m duration...
   ✅ Generated 10/50 licenses for 3m
   ✅ Generated 20/50 licenses for 3m
   ✅ Generated 30/50 licenses for 3m
   ✅ Generated 40/50 licenses for 3m
   ✅ Generated 50/50 licenses for 3m

🎉 License Generation Complete!
═══════════════════════════════
✅ Total licenses generated: 50
🔗 Google Sheet: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit

📋 Sample licenses:
   • ISX3M-ABC123-456789 (3m)
   • ISX3M-DEF456-789012 (3m)
   • ISX3M-GHI789-012345 (3m)
   • ISX3M-JKL012-345678 (3m)
   • ISX3M-MNO345-678901 (3m)
   ... and 45 more
```

### File Output
```
ISX License Keys
================

ISX3M-ABC123-456789 (3m)
ISX3M-DEF456-789012 (3m)
ISX3M-GHI789-012345 (3m)
ISX1Y-JKL012-345678 (1y)
ISX6M-MNO345-678901 (6m)
```

## 🏪 Business Model Integration

### Recharge Card Distribution
1. **Generate**: Create licenses in bulk
2. **Distribute**: Sell/give license keys to customers
3. **Track**: Monitor usage via Google Sheets
4. **Support**: Help customers activate licenses

### Pricing Strategy
- **ISX1M**: $10 - Trial/Demo licenses
- **ISX3M**: $25 - Standard licenses
- **ISX6M**: $40 - Power user licenses
- **ISX1Y**: $100 - Enterprise licenses

## 🛡️ Security Features

- **Cryptographically Secure**: Uses `crypto/rand` for key generation
- **Unique Keys**: Timestamp-based uniqueness guarantee
- **Machine Binding**: Licenses bind to specific machines when activated
- **Remote Tracking**: Real-time status updates in Google Sheets

## 🔧 Troubleshooting

### Common Issues

1. **"API key invalid"**
   - Verify API key in `license-config.json`
   - Check Google Sheets API is enabled
   - Ensure no spaces/extra characters

2. **"Sheet not found"**
   - Verify sheet name is exactly "Licenses"
   - Check sheet is publicly accessible
   - Confirm sheet ID is correct

3. **"Request failed"**
   - Check internet connection
   - Verify Google Sheets API quotas
   - Try reducing batch size

### Debug Steps

1. Test with small batch first: `-total=5`
2. Check Google Sheets manually
3. Verify API key permissions
4. Review error messages

## 📈 Performance

- **Generation Speed**: ~5 licenses per second
- **API Rate Limits**: Built-in delays to prevent quota issues
- **Memory Usage**: Minimal memory footprint
- **File Size**: Small executable (~8MB)

## 🔄 Updates

To update the license generator:
1. Download new version
2. Replace executable
3. Keep existing `license-config.json`
4. Run as usual

## 📞 Support

For technical support:
1. Check Google Sheets for license status
2. Review error messages carefully
3. Verify API configuration
4. Test with minimal examples

---

**License Generator v1.0** - Standalone ISX License Management Tool 