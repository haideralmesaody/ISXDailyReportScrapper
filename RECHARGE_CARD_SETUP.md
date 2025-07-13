# ISX Recharge Card License System Setup

## 🎯 Overview
This system generates licenses like **recharge cards** - pre-generated licenses that can be distributed without user-specific information. Users can then activate these licenses on their machines.

## 📋 Step 1: Configure Your Google Sheet

### Required Columns in Row 1:
Add these **6 columns** to your Google Sheet at: 
https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit

| **A** | **B** | **C** | **D** | **E** | **F** |
|-------|-------|-------|-------|-------|-------|
| **LicenseKey** | **Duration** | **ExpiryDate** | **Status** | **MachineID** | **ActivatedDate** |

### Column Details:
- **LicenseKey**: Generated license key (e.g., ISX3M-ABC123-DEF456)
- **Duration**: License duration (1m, 3m, 6m, 1y)
- **ExpiryDate**: When license expires (set when activated)
- **Status**: Available/Activated/Expired/Revoked
- **MachineID**: Machine that activated it (empty until activated)
- **ActivatedDate**: When license was activated (empty until activated)

### Make Sheet Public:
1. Click **Share** → **Change to anyone with the link**
2. Set permissions to **Viewer**
3. Click **Done**

## 🔧 Step 2: Configure API Access

### Update license-config.json:
```json
{
  "sheet_id": "1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc",
  "api_key": "YOUR_GOOGLE_SHEETS_API_KEY_HERE",
  "sheet_name": "Licenses"
}
```

**Get your Google Sheets API Key:**
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Enable Google Sheets API
3. Create API Key
4. Replace `YOUR_GOOGLE_SHEETS_API_KEY_HERE` with your actual key

## 🎫 Step 3: Generate 100 Recharge Card Licenses

### Build the Tools:
```bash
# Build bulk license generator
go build -o bulk-license-generator.exe cmd/bulk-license-generator/main.go

# Build licensed web version
go build -o web-licensed.exe cmd/web-licensed/main.go
```

### Generate Licenses:
```bash
# Generate 100 random licenses
./bulk-license-generator.exe -total=100

# Or generate specific quantities:
./bulk-license-generator.exe -1m=25 -3m=50 -6m=20 -1y=5

# Or generate mixed:
./bulk-license-generator.exe -3m=30 -6m=30 -total=40
```

### Example License Types:
- **1-month**: `ISX1M-ABC123-DEF456789`
- **3-month**: `ISX3M-XYZ789-GHI012345`
- **6-month**: `ISX6M-JKL345-MNO678901`
- **1-year**: `ISX1Y-PQR567-STU234567`

## 🏪 Step 4: Business Model Usage

### Recharge Card Distribution:
1. **Generate in bulk**: Create 100+ licenses
2. **Distribute**: Sell/give licenses to users
3. **User activation**: Users activate licenses on their machines
4. **Track usage**: Monitor via Google Sheets

### Pricing Strategy Example:
- **1-month (ISX1M)**: $10 - Trial users
- **3-month (ISX3M)**: $25 - Regular users
- **6-month (ISX6M)**: $40 - Power users
- **1-year (ISX1Y)**: $100 - Enterprise users

## 🚀 Step 5: User Experience

### For End Users:
1. **Purchase**: Buy a license key (e.g., `ISX3M-ABC123-DEF456`)
2. **Install**: Download ISX application
3. **Activate**: Run `web-licensed.exe` and enter license key
4. **Use**: Application works for license duration

### License Status Flow:
```
Generated → Available → Activated → Expired/Revoked
```

## 📊 Step 6: Monitor and Manage

### Check License Status:
Monitor your Google Sheet to see:
- Total licenses generated
- Available licenses (not yet activated)
- Activated licenses (with machine IDs)
- Expired licenses
- Revenue tracking

### License Management:
- **Available**: Ready for sale/distribution
- **Activated**: In use on specific machine
- **Expired**: No longer valid
- **Revoked**: Manually disabled

## 🔒 Step 7: Security Features

### Machine Binding:
- Each license works on only ONE machine
- Machine ID generated from hostname + system hash
- Cannot transfer activated licenses

### Remote Validation:
- Daily checks with Google Sheets
- Automatic expiration enforcement
- Revocation support

## 📱 Step 8: Test the System

### Test License Generation:
```bash
# Generate test licenses
./bulk-license-generator.exe -1m=5

# Check your Google Sheet - should see 5 new licenses
```

### Test License Activation:
```bash
# Run licensed web version
./web-licensed.exe

# Follow prompts to activate a license
# Check sheet - license should show as "Activated"
```

## 🎯 Example Usage Scenarios

### Scenario 1: Software Vendor
- Generate 1000 licenses monthly
- Sell via online store
- Customers receive license keys via email
- Track sales and usage via Google Sheets

### Scenario 2: Trial Distribution
- Generate 100 1-month licenses
- Distribute at conferences/events
- Users activate for trial access
- Convert to paid licenses later

### Scenario 3: Subscription Service
- Generate licenses with different durations
- Provide renewal licenses before expiry
- Track customer usage patterns
- Automate billing integration

## 🛠️ Commands Reference

### Bulk License Generator:
```bash
# Generate specific quantities
./bulk-license-generator.exe -1m=25 -3m=50 -6m=20 -1y=5

# Generate random mix
./bulk-license-generator.exe -total=100

# Save to file
./bulk-license-generator.exe -total=50 -output=licenses.txt
```

### Web Licensed Version:
```bash
# Start licensed web server
./web-licensed.exe

# Will prompt for license activation if needed
# Access at: http://localhost:8080
```

## 🔧 Troubleshooting

### Common Issues:

1. **"API key invalid"**
   - Verify API key in license-config.json
   - Ensure Google Sheets API is enabled

2. **"Sheet not found"**
   - Check sheet name is exactly "Licenses"
   - Verify sheet is publicly accessible

3. **"License generation failed"**
   - Check Google Sheets API quotas
   - Verify network connectivity

### Support:
- Check Google Sheets for license status
- Review application logs
- Verify API key permissions

---

**Next Steps:**
1. ✅ Set up Google Sheets columns
2. ✅ Configure API access
3. ✅ Generate your first 100 licenses
4. ✅ Test license activation
5. ✅ Start distributing licenses! 