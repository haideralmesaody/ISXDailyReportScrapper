# Google Sheets Setup Guide for ISX License System

## Overview
This guide will walk you through setting up Google Sheets API access and configuring your license tracking sheet.

## Step 1: Set up Google Sheets API

### 1.1 Create a Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Name it something like "ISX-License-System"

### 1.2 Enable Google Sheets API
1. In the Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for "Google Sheets API"
3. Click on it and press **Enable**

### 1.3 Create API Credentials
1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **API Key**
3. Copy the API key - you'll need this for the `license-config.json` file
4. (Optional) Click **Restrict Key** to limit it to Google Sheets API only for security

### 1.4 Update Configuration
1. Open `license-config.json` in your project
2. Replace `YOUR_GOOGLE_SHEETS_API_KEY_HERE` with your actual API key
3. The sheet_id is already configured for your sheet

## Step 2: Configure Your Google Sheet

### 2.1 Sheet Structure
Your Google Sheet at: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit

**Must have a sheet named "Licenses"** with these columns in Row 1:

| A | B | C | D | E | F | G | H |
|---|---|---|---|---|---|---|---|
| LicenseKey | UserEmail | ExpiryDate | Duration | MachineID | IssuedDate | Status | LastChecked |

### 2.2 Column Descriptions:
- **LicenseKey**: Generated license key (e.g., ISX3M-ABC123-DEF456)
- **UserEmail**: Customer's email address
- **ExpiryDate**: When the license expires (YYYY-MM-DD format)
- **Duration**: License duration (1m, 3m, 6m, 1y)
- **MachineID**: Unique machine identifier (auto-filled when activated)
- **IssuedDate**: When the license was created (YYYY-MM-DD format)
- **Status**: Active/Expired/Revoked
- **LastChecked**: Last time the license was validated

### 2.3 Make Sheet Public (Read-Only)
1. In your Google Sheet, click **Share** button
2. Click **Change to anyone with the link**
3. Set permissions to **Viewer** (read-only)
4. Click **Done**

## Step 3: Test the Setup

### 3.1 Generate Your First License
```bash
# Build the license generator
go build -o license-generator.exe cmd/license-generator/main.go

# Generate a test license
./license-generator.exe -email=test@example.com -duration=1m
```

### 3.2 Check the Sheet
After running the license generator, check your Google Sheet to see if the license was added successfully.

### 3.3 Test the Web Interface
```bash
# Build the licensed web version
go build -o web-licensed.exe cmd/web-licensed/main.go

# Run it
./web-licensed.exe
```

## Step 4: Production Deployment

### 4.1 Security Best Practices
1. **Restrict API Key**: In Google Cloud Console, restrict your API key to only Google Sheets API
2. **IP Restrictions**: If possible, restrict the API key to your server's IP address
3. **Environment Variables**: Store the API key in environment variables instead of the config file

### 4.2 Environment Variable Setup
Instead of storing the API key in the config file, you can use environment variables:

```bash
# Set environment variable
set GOOGLE_SHEETS_API_KEY=your_actual_api_key_here

# Or in PowerShell
$env:GOOGLE_SHEETS_API_KEY="your_actual_api_key_here"
```

### 4.3 Backup Strategy
1. Regularly export your Google Sheet as Excel/CSV
2. Keep backups of your license database
3. Monitor sheet access logs

## Troubleshooting

### Common Issues:

1. **"Sheet not found" error**
   - Ensure the sheet is named exactly "Licenses"
   - Check that the sheet is publicly accessible

2. **"API key invalid" error**
   - Verify the API key is correct
   - Ensure Google Sheets API is enabled
   - Check for any IP restrictions

3. **"Permission denied" error**
   - Make sure the sheet is shared with "Anyone with the link"
   - Verify the sheet ID is correct

4. **"Column not found" error**
   - Ensure all required columns are present in Row 1
   - Check spelling matches exactly

### Testing Commands:

```bash
# Test license generation
./license-generator.exe -email=your-email@example.com -duration=1m

# Test web interface with license
./web-licensed.exe

# Check if license file was created
dir *.license
```

## Support

If you encounter any issues:
1. Check the Google Cloud Console for API usage and errors
2. Verify your Google Sheet permissions and structure
3. Test with a simple 1-minute license first
4. Check the application logs for detailed error messages

---

**Next Steps:**
1. Set up your Google Sheets API key
2. Configure your sheet with the required columns
3. Test license generation
4. Deploy the licensed version of your application 