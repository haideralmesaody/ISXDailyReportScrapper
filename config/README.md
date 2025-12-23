# Configuration Management

This directory contains all configuration templates and schemas for ISX Pulse.

## Directory Structure

```
config/
├── examples/                    # Configuration templates
│   ├── credentials.json.example    # Google Sheets API credentials
│   ├── sheets-config.json.example  # Sheet ID mappings
│   └── license.json.example        # License configuration (to be created)
├── schemas/                     # JSON schemas for validation
│   ├── credentials.schema.json     # Credentials validation schema
│   ├── sheets-config.schema.json   # Sheets config validation schema
│   └── license.schema.json         # License validation schema
└── README.md                    # This file
```

## Setup Instructions

### 1. Google Sheets API Credentials

1. Copy the template:
   ```bash
   cp config/examples/credentials.json.example credentials.json
   ```

2. Obtain credentials from Google Cloud Console:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create or select a project
   - Enable Google Sheets API
   - Create service account credentials
   - Download the JSON key file

3. Replace the content of `credentials.json` with your downloaded key file

### 2. Sheets Configuration

1. Copy the template:
   ```bash
   cp config/examples/sheets-config.json.example sheets-config.json
   ```

2. Update with your Google Sheets IDs:
   ```json
   {
     "daily_report": "YOUR_DAILY_REPORT_SHEET_ID",
     "weekly_report": "YOUR_WEEKLY_REPORT_SHEET_ID",
     "monthly_report": "YOUR_MONTHLY_REPORT_SHEET_ID"
   }
   ```

### 3. License Configuration

The license system uses hardware-locked encryption. License files are automatically generated during activation and should not be manually edited.

## Security Notes

⚠️ **IMPORTANT SECURITY PRACTICES**:

1. **Never commit real credentials** to version control
2. **Use environment variables** for sensitive data in production
3. **Rotate credentials regularly**
4. **Limit API key permissions** to minimum required
5. **Use encrypted storage** for production credentials

## File Permissions

Ensure proper file permissions for security:

```bash
# Unix/Linux/Mac
chmod 600 credentials.json
chmod 600 sheets-config.json
chmod 600 license.dat

# Windows (using icacls)
icacls credentials.json /inheritance:r /grant:r "%USERNAME%:F"
icacls sheets-config.json /inheritance:r /grant:r "%USERNAME%:F"
icacls license.dat /inheritance:r /grant:r "%USERNAME%:F"
```

## Environment Variables

For production deployments, use environment variables instead of files:

- `GOOGLE_APPLICATION_CREDENTIALS`: Path to credentials.json
- `ISX_SHEETS_CONFIG`: Path to sheets-config.json
- `ISX_LICENSE_FILE`: Path to license.dat

## Validation

Use the provided schemas to validate your configuration:

```bash
# Install ajv-cli globally
npm install -g ajv-cli

# Validate credentials
ajv validate -s config/schemas/credentials.schema.json -d credentials.json

# Validate sheets config
ajv validate -s config/schemas/sheets-config.schema.json -d sheets-config.json
```

## Troubleshooting

### Common Issues

1. **"credentials.json not found"**
   - Ensure the file exists in the project root
   - Check file permissions

2. **"Invalid credentials"**
   - Verify the service account has necessary permissions
   - Check if the API is enabled in Google Cloud Console

3. **"Sheet not found"**
   - Verify the sheet ID is correct
   - Ensure the service account has access to the sheet

## Support

For issues or questions, please refer to the main project documentation or open an issue on GitHub.