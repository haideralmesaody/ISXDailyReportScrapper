# ISX Scraper Licensing System Setup

This document explains how to set up and use the professional licensing system for the ISX Daily Reports Scraper.

## 🎯 **Overview**

The licensing system provides:
- **Time-based licenses** (1m, 3m, 6m, 1y)
- **Google Sheets integration** for license tracking
- **Machine-specific activation** (one license per machine)
- **Auto-update capabilities** from GitHub
- **Expiration enforcement** with renewal prompts

## 📋 **Prerequisites**

### 1. Google Sheets Setup
1. Create a new Google Sheet
2. Set up columns: `LicenseKey | UserEmail | ExpiryDate | Duration | MachineID | IssuedDate | Status | LastChecked`
3. Get your Google Sheets API key from [Google Cloud Console](https://console.cloud.google.com/)
4. Note your Sheet ID (from the URL)

### 2. Configuration Files
Create `license-config.json`:
```json
{
  "sheet_id": "1ABCDefGhijKLmnOpQrSTuvwxYZ123456789",
  "api_key": "AIzaSyD1234567890abcdefghijklmnopqrstuvw",
  "sheet_name": "Licenses"
}
```

## 🔧 **Build Instructions**

### 1. Build License Generator
```bash
go build -o license-generator.exe cmd/license-generator/main.go
```

### 2. Build Licensed Web Server
```bash
go build -o web-licensed.exe cmd/web-licensed/main.go
```

## 🎫 **License Generation**

### Generate Licenses for Users

```bash
# 1 month license
./license-generator.exe -email=user@example.com -duration=1m

# 3 month license
./license-generator.exe -email=user@example.com -duration=3m

# 6 month license
./license-generator.exe -email=user@example.com -duration=6m

# 1 year license
./license-generator.exe -email=user@example.com -duration=1y
```

### Sample Output:
```
🎫 LICENSE GENERATED SUCCESSFULLY!
═══════════════════════════════════
📧 Email:      user@example.com
⏱️  Duration:   3m
🔑 License:    ISX3M-AbCdEfGhIjKlMnOp
═══════════════════════════════════

📋 Instructions for user:
1. Run the ISX scraper application
2. When prompted, enter this license key: ISX3M-AbCdEfGhIjKlMnOp
3. The application will be activated for 3 months

💾 License has been saved to Google Sheets for tracking.
```

## 👤 **User Experience**

### First Run (No License)
```
❌ Invalid or Expired License
═══════════════════════════════════
Error: no local license found
Please enter your license key: ISX3M-AbCdEfGhIjKlMnOp
✅ License activated successfully!
ISX Web Interface (Licensed) starting on http://localhost:8080
```

### Subsequent Runs (Valid License)
```
✅ License Valid - 87 days remaining
ISX Web Interface (Licensed) starting on http://localhost:8080
```

### Expiring License Warning
```
✅ License Valid - 5 days remaining
⚠️  License expires soon: 2024-02-15
ISX Web Interface (Licensed) starting on http://localhost:8080
```

## 🔒 **License Enforcement**

### API Protection
All API endpoints require valid license:
```bash
# Without valid license
curl http://localhost:8080/api/tickers
# Returns: {"error":"License invalid or expired","message":"license expired","code":"LICENSE_REQUIRED"}

# With valid license
curl http://localhost:8080/api/tickers
# Returns: {"tickers":[...]}
```

### Web Interface Protection
- Invalid license blocks all functionality
- License status displayed in console
- Automatic expiration checking every 24 hours

## 🔄 **Auto-Update System**

### Features
- **Automatic checks** for GitHub releases
- **Background downloads** of updates
- **Safe replacement** with backup/restore
- **Platform-specific** binaries (Windows/Mac/Linux)

### API Endpoints
```bash
# Check for updates
GET /api/update/check

# Install update
POST /api/update/install
```

## 📊 **Google Sheets Tracking**

### License Data Structure
| Column | Description | Example |
|--------|-------------|---------|
| LicenseKey | Unique license identifier | ISX3M-AbCdEfGhIjKlMnOp |
| UserEmail | Customer email | user@example.com |
| ExpiryDate | When license expires | 2024-05-15 10:30:00 |
| Duration | License period | 3m |
| MachineID | Unique machine hash | a1b2c3d4e5f6g7h8 |
| IssuedDate | When license was created | 2024-02-15 10:30:00 |
| Status | Current status | active/expired/issued |
| LastChecked | Last validation time | 2024-03-01 14:22:33 |

### Status Values
- **issued**: License created but not activated
- **active**: License activated and valid
- **expired**: License past expiration date

## 🛡️ **Security Features**

### Machine Binding
- Licenses are bound to specific machines
- Uses hostname + user environment hash
- Prevents license sharing between machines

### Remote Validation
- Daily validation with Google Sheets
- Detects license revocation
- Handles offline scenarios gracefully

### Tamper Protection
- License file encrypted locally
- Machine ID verification
- Expiration date validation

## 🚀 **Deployment**

### For Distributors
1. Set up Google Sheets with API access
2. Build license-generator tool
3. Generate licenses as needed
4. Provide users with license keys

### For End Users
1. Download licensed application
2. Run application
3. Enter provided license key when prompted
4. Enjoy full functionality until expiration

## 🔧 **Troubleshooting**

### Common Issues

#### "License system not available"
- Check `license-config.json` exists
- Verify Google Sheets API key
- Ensure internet connectivity

#### "License already activated on another machine"
- Each license works on one machine only
- Contact distributor for additional licenses
- Check if license was transferred properly

#### "License validation failed"
- Verify license key format
- Check Google Sheets connectivity
- Ensure license hasn't been revoked

#### "Failed to save license locally"
- Check file permissions
- Ensure disk space available
- Run as administrator if needed

### Debug Mode
Set environment variable for verbose logging:
```bash
export ISX_DEBUG=true
./web-licensed.exe
```

## 📈 **Business Model Integration**

### Pricing Tiers
- **1 Month**: Trial/evaluation period
- **3 Months**: Standard users
- **6 Months**: Power users with discount
- **1 Year**: Enterprise users with maximum discount

### License Management
- Track usage via Google Sheets
- Monitor active/expired licenses
- Generate usage reports
- Handle renewals and upgrades

### Customer Support
- License activation assistance
- Machine transfer support
- Renewal notifications
- Technical support tiers

## 🎯 **Best Practices**

### For License Distributors
1. Keep Google Sheets secure and backed up
2. Monitor license usage patterns
3. Set up automated renewal reminders
4. Provide clear activation instructions

### For Application Distribution
1. Include clear licensing terms
2. Provide activation support
3. Test licensing system thoroughly
4. Monitor update deployment

## 📞 **Support**

For licensing system support:
- Check this documentation first
- Review Google Sheets data
- Test with debug mode enabled
- Contact technical support with specific error messages

---

**Note**: This licensing system is designed for professional distribution of the ISX scraper application. Ensure compliance with all applicable laws and ISX terms of service. 