# Credentials Setup Guide

This repository has been sanitized to remove sensitive credentials. To run the application, you need to set up your own credentials.

## Files Needed

1. **credentials.json** - Google Service Account credentials
2. **sheets-config.json** - Google Sheets configuration

## Setup Instructions

### 1. Create Google Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google Sheets API
4. Create a Service Account:
   - Go to IAM & Admin > Service Accounts
   - Click "Create Service Account"
   - Give it a name and description
   - Grant it "Editor" role
   - Create and download JSON key

### 2. Set Up Credentials

Copy the example files and fill in your values:

```bash
# Copy example files
cp credentials.json.example credentials.json
cp sheets-config.json.example sheets-config.json
```

Edit `credentials.json` with your service account details from the downloaded JSON key.

Edit `sheets-config.json`:
- `spreadsheet_id`: Your Google Sheets ID (from the URL)
- `sheet_name`: Name of the sheet tab (default: "Licenses")

### 3. Alternative: Environment Variables

Instead of files, you can use environment variables:

```bash
# Set service account credentials
export ISX_CREDENTIALS='{"type": "service_account", ...}'

# Set sheet configuration
export ISX_SHEET_ID="your-sheet-id"
export ISX_SHEET_NAME="Licenses"
```

## Security Notes

- **NEVER** commit `credentials.json` or `sheets-config.json` to version control
- These files are already in `.gitignore` to prevent accidental commits
- Keep backup copies in a secure location
- The `.credentials-backup/` directory contains backups and should also not be committed

## Restoring Credentials

If you have backed up credentials:

```bash
# Restore from backup
cp .credentials-backup/credentials.json ./
cp .credentials-backup/sheets-config.json ./
```

## For Developers

The code now loads credentials in this order:
1. Environment variables (`ISX_CREDENTIALS`, `ISX_SHEET_ID`)
2. Local files (`credentials.json`, `sheets-config.json`)
3. Placeholder values (will fail validation)

This ensures the code can be safely shared without exposing sensitive data.