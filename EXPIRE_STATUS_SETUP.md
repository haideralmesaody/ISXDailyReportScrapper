# 📊 ExpireStatus Column Setup Guide

## 🎯 Overview
The ISX license system now includes an **ExpireStatus** column in Google Sheets for better license management and monitoring. This column automatically tracks the expiration status of each license based on days remaining.

## 📋 Updated Google Sheets Structure

### New Column Layout (8 columns total):
| **A** | **B** | **C** | **D** | **E** | **F** | **G** | **H** |
|-------|-------|-------|-------|-------|-------|-------|-------|
| **LicenseKey** | **Duration** | **ExpiryDate** | **Status** | **MachineID** | **ActivatedDate** | **LastConnected** | **ExpireStatus** |

### ExpireStatus Values:
- **Available** - License not yet activated (no expiry date set)
- **Active** - More than 30 days remaining (Green status)
- **Warning** - 8-30 days remaining (Yellow status)  
- **Critical** - 7 or fewer days remaining (Red status)
- **Expired** - License has expired (Red status)

## 🔧 How to Update Existing Google Sheets

### Option 1: Add Column Header Only (Recommended)
1. Open your Google Sheet
2. Click on column **H** (the first empty column after LastConnected)
3. Right-click and select "Insert 1 left"
4. In cell **H1**, type: `ExpireStatus`
5. The system will automatically populate this column during the next heartbeat

### Option 2: Add Column with Formula (Advanced)
1. Follow steps 1-4 above
2. In cell **H2**, add this formula:
   ```
   =IF(C2="","Available",IF(C2<TODAY(),"Expired",IF(C2-TODAY()<=7,"Critical",IF(C2-TODAY()<=30,"Warning","Active"))))
   ```
3. Copy this formula down to all existing rows
4. The system will automatically overwrite these values with live calculations

## 📊 Benefits of ExpireStatus Column

### For License Management:
- **Quick Status Overview**: See all license statuses at a glance
- **Proactive Monitoring**: Identify licenses approaching expiration
- **Automated Tracking**: No manual calculation needed
- **Historical Data**: Track status changes over time

### For Business Analytics:
- **Renewal Opportunities**: Identify customers needing renewals
- **Usage Patterns**: See which license types expire most frequently
- **Support Prioritization**: Focus on critical/warning status licenses
- **Revenue Forecasting**: Predict renewal revenue based on expiry status

## 🚀 Automatic Updates

### When ExpireStatus Updates:
- **Every 30 minutes**: During license heartbeat calls
- **On license activation**: When user activates a license
- **During validation**: When system validates license remotely

### Real-time Calculation:
The system automatically calculates ExpireStatus based on:
```
Days Remaining = (ExpiryDate - Current Date)

If Days Remaining:
- No expiry date set → "Available"
- <= 0 days → "Expired"
- <= 7 days → "Critical"
- <= 30 days → "Warning"
- > 30 days → "Active"
```

## 📈 Monitoring Dashboard Ideas

### Filter Views You Can Create:
1. **Critical Licenses** - Filter by ExpireStatus = "Critical"
2. **Expiring Soon** - Filter by ExpireStatus = "Warning" or "Critical"
3. **Active Licenses** - Filter by ExpireStatus = "Active"
4. **Available Licenses** - Filter by ExpireStatus = "Available"

### Conditional Formatting:
1. Select column H (ExpireStatus)
2. Format → Conditional formatting
3. Add rules:
   - **"Critical"** → Red background
   - **"Warning"** → Yellow background
   - **"Active"** → Green background
   - **"Available"** → Blue background
   - **"Expired"** → Dark red background

## 🔍 Troubleshooting

### Common Issues:

#### Column Not Updating
- **Check**: Ensure column H exists and is named "ExpireStatus"
- **Solution**: Wait for next heartbeat (30 minutes) or restart application

#### Wrong Status Values
- **Check**: Verify ExpiryDate format is YYYY-MM-DD
- **Solution**: Correct date format in column C

#### Missing ExpireStatus
- **Check**: Ensure Google Sheet has 8 columns (A-H)
- **Solution**: Add ExpireStatus column as described above

### Backwards Compatibility:
- **Existing sheets**: Will continue to work with 7 columns
- **New functionality**: ExpireStatus will be added automatically
- **No data loss**: All existing data remains intact

## 📞 Support

For issues with ExpireStatus column:
1. Verify Google Sheet structure matches the 8-column format
2. Check that the application has write permissions to the sheet
3. Wait for automatic updates (30-minute intervals)
4. Contact support if issues persist

---
**Note**: The ExpireStatus column enhances license management without affecting existing functionality. All licenses will continue to work normally during the transition. 