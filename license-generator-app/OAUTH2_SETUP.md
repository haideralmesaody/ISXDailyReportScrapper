# OAuth2 Setup Guide for ISX License Generator

## 🎯 Overview
This guide will help you set up OAuth2 authentication for the ISX License Generator to securely access Google Sheets.

## 📋 Step 1: Google Cloud Console Setup

### 1.1 Create/Select a Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Name it something like "ISX-License-Generator"

### 1.2 Enable Google Sheets API
1. In the Google Cloud Console, navigate to **APIs & Services** > **Library**
2. Search for "Google Sheets API"
3. Click on it and press **Enable**

### 1.3 Create OAuth2 Credentials
1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. If prompted, configure the **OAuth consent screen**:
   - Choose **External** user type
   - Fill in required fields (App name, User support email, etc.)
   - Add your email to test users
   - Save and continue through all steps

4. When creating OAuth client ID:
   - **Application type**: Choose **Desktop application**
   - **Name**: Enter "ISX License Generator"
   - Click **Create**

5. **Download the credentials**:
   - Click the download button (⬇️) next to your newly created OAuth client
   - Save the JSON file as `credentials.json`

### 1.4 Extract Client ID and Secret
From the downloaded `credentials.json` file, find:
- `client_id`: The OAuth2 client ID
- `client_secret`: The OAuth2 client secret

## 🔧 Step 2: Configure the License Generator

### 2.1 Update oauth-config.json
Edit the `oauth-config.json` file:

```json
{
  "sheet_id": "1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc",
  "sheet_name": "Licenses",
  "client_id": "YOUR_ACTUAL_CLIENT_ID_HERE",
  "client_secret": "YOUR_ACTUAL_CLIENT_SECRET_HERE",
  "redirect_url": "urn:ietf:wg:oauth:2.0:oob"
}
```

**Replace**:
- `YOUR_ACTUAL_CLIENT_ID_HERE` with your actual client ID
- `YOUR_ACTUAL_CLIENT_SECRET_HERE` with your actual client secret

### 2.2 Prepare Your Google Sheet
Make sure your Google Sheet has these columns in **Row 1**:

| **A** | **B** | **C** | **D** | **E** | **F** |
|-------|-------|-------|-------|-------|-------|
| **LicenseKey** | **Duration** | **ExpiryDate** | **Status** | **MachineID** | **ActivatedDate** |

**Important**: The sheet must be named exactly **"Licenses"**

## 🚀 Step 3: Build and Run

### 3.1 Build the Application
```bash
go build -o license-generator.exe main.go
```

### 3.2 First-Time Authentication
When you run the license generator for the first time:

```bash
./license-generator.exe -total=5
```

You'll see:
```
🔐 OAuth2 Authentication Required
═══════════════════════════════════
📋 Please visit this URL to authorize the application:
🔗 https://accounts.google.com/o/oauth2/auth?...

✏️  Enter the authorization code: 
```

### 3.3 Complete Authentication
1. **Copy the URL** and open it in your browser
2. **Sign in** with your Google account
3. **Grant permissions** to the application
4. **Copy the authorization code** from the browser
5. **Paste the code** into the terminal
6. Press **Enter**

### 3.4 Token Storage
After successful authentication:
- A `token.json` file will be created
- Future runs won't require re-authentication
- The token will be automatically refreshed

## 🎫 Step 4: Generate Licenses

### 4.1 Test Generation
```bash
# Generate 5 test licenses
./license-generator.exe -total=5
```

### 4.2 Generate 100 Licenses
```bash
# Generate 100 random licenses
./license-generator.exe -total=100

# Generate specific quantities
./license-generator.exe -1m=25 -3m=50 -6m=20 -1y=5

# Save to file
./license-generator.exe -total=100 -output=licenses.txt
```

## 🔒 Security Features

### OAuth2 Benefits:
- **More Secure**: Uses proper OAuth2 flow
- **Token Refresh**: Automatic token renewal
- **Granular Permissions**: Only requests Sheets access
- **Revocable**: Can revoke access anytime

### Files Created:
- `token.json`: OAuth2 access token (keep secure)
- `credentials.json`: OAuth2 credentials (keep secure)
- `oauth-config.json`: Configuration file

## 🛠️ Troubleshooting

### Common Issues:

#### 1. "OAuth consent screen not configured"
- Go to **APIs & Services** > **OAuth consent screen**
- Fill in required information
- Add your email to test users

#### 2. "Invalid client ID or secret"
- Verify client ID and secret in `oauth-config.json`
- Make sure they match your Google Cloud credentials

#### 3. "Sheet not found"
- Verify sheet name is exactly "Licenses"
- Check sheet ID in configuration

#### 4. "Access denied"
- Make sure to grant all requested permissions
- Check if your account has access to the sheet

#### 5. "Token expired"
- Delete `token.json` and re-authenticate
- The app will automatically refresh tokens

### Debug Steps:

1. **Verify Configuration**:
   ```bash
   # Check if config file exists and is valid
   cat oauth-config.json
   ```

2. **Test Authentication**:
   ```bash
   # Delete token and re-authenticate
   rm token.json
   ./license-generator.exe -total=1
   ```

3. **Check Permissions**:
   - Visit [Google Account Permissions](https://myaccount.google.com/permissions)
   - Verify "ISX License Generator" is listed
   - Check granted permissions

## 📱 Example Run Output

```
🎫 ISX License Generator v2.0 (OAuth2)
═══════════════════════════════════════

🔄 Generating 5 random duration licenses...
   ✅ Generated 5/5 random licenses

🎉 License Generation Complete!
═══════════════════════════════════════
✅ Total licenses generated: 5
🔗 Google Sheet: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit

📋 Sample licenses:
   • ISX3M-ABC123-456789 (3m)
   • ISX6M-DEF456-789012 (6m)
   • ISX1Y-GHI789-012345 (1y)
   • ISX1M-JKL012-345678 (1m)
   • ISX3M-MNO345-678901 (3m)
```

## 🔄 Token Management

### Automatic Token Refresh:
- OAuth2 tokens are automatically refreshed
- No manual intervention required
- Tokens are stored securely in `token.json`

### Manual Token Reset:
If you need to reset authentication:
```bash
rm token.json
./license-generator.exe -total=1
```

## 📞 Support

For issues with OAuth2 setup:
1. Check Google Cloud Console for API quotas
2. Verify OAuth consent screen configuration
3. Ensure proper permissions are granted
4. Check network connectivity

---

**ISX License Generator v2.0** - Secure OAuth2 Authentication 