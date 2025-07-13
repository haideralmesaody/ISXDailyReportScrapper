# 🚀 **Simple Service Account Setup** (Recommended)

Since OAuth2 verification is complex, here's a **much simpler approach** using Service Account:

## 📋 **Step 1: Create Service Account (5 minutes)**

### 1.1 Go to Google Cloud Console
1. Visit [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project (or create new one)

### 1.2 Create Service Account
1. Go to **IAM & Admin** → **Service Accounts**
2. Click **+ CREATE SERVICE ACCOUNT**
3. **Service account name**: `isx-license-generator`
4. **Service account ID**: `isx-license-generator` (auto-filled)
5. Click **CREATE AND CONTINUE**
6. **Role**: Select **Editor** (or **Project Editor**)
7. Click **CONTINUE** → **DONE**

### 1.3 Generate JSON Key
1. Find your new service account in the list
2. Click on the **service account email**
3. Go to **Keys** tab
4. Click **ADD KEY** → **Create new key**
5. Choose **JSON** format
6. Click **CREATE**
7. **Save the downloaded JSON file** as `service-account-credentials.json` in your license-generator-app folder

## 📊 **Step 2: Share Google Sheet with Service Account**

### 2.1 Get Service Account Email
From the JSON file you downloaded, find the `client_email`:
```json
{
  "client_email": "isx-license-generator@your-project.iam.gserviceaccount.com"
}
```

### 2.2 Share Your Google Sheet
1. Open your Google Sheet: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit
2. Click **Share** button
3. Add the service account email with **Editor** permissions
4. Click **Send**

### 2.3 Set up Sheet Columns
Make sure your sheet has these columns in **Row 1**:
| **A** | **B** | **C** | **D** | **E** | **F** |
|-------|-------|-------|-------|-------|-------|
| **LicenseKey** | **Duration** | **ExpiryDate** | **Status** | **MachineID** | **ActivatedDate** |

## 🎯 **Step 3: Update the License Generator**

### 3.1 Modify main.go
Replace the OAuth2 authentication with service account authentication:

```go
// At the top, replace oauth2 imports with:
import (
    "context"
    "google.golang.org/api/option"
    "google.golang.org/api/sheets/v4"
)

// In NewLicenseGenerator function, replace OAuth2 setup with:
func NewLicenseGenerator(configFile string) (*LicenseGenerator, error) {
    config, err := loadServiceAccountConfig(configFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %v", err)
    }

    // Create Sheets service with service account
    ctx := context.Background()
    sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile("service-account-credentials.json"))
    if err != nil {
        return nil, fmt.Errorf("failed to create sheets service: %v", err)
    }

    return &LicenseGenerator{
        config:        config,
        sheetsService: sheetsService,
    }, nil
}
```

## 🚀 **Step 4: Generate 100 Licenses**

```bash
# Build the updated application
go build -o license-generator-sa.exe main.go

# Generate 100 licenses (no authentication prompts!)
./license-generator-sa.exe -total=100
```

## ✅ **Why Service Account is Better:**

- **✅ No OAuth2 verification** required
- **✅ No browser authentication** needed  
- **✅ No test user setup** required
- **✅ Works immediately** after setup
- **✅ More secure** for automation
- **✅ No token expiration** issues

## 🔧 **Files You Need:**

1. `service-account-credentials.json` (downloaded from Google Cloud)
2. `service-account-config.json` (already created)
3. Updated `main.go` with service account authentication

## 📞 **Quick Test:**

After setup, test with:
```bash
./license-generator-sa.exe -total=1
```

Should show:
```
🎫 ISX License Generator v3.0 (Service Account)
✅ Generated 1/1 random licenses
🎉 License Generation Complete!
```

**No prompts, no browser, just works!** 🎯 