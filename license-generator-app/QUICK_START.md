# 🚀 Quick Start Guide - Service Account (Simplest Method)

## ⚡ 3 Simple Steps to Generate 100 Licenses

### Step 1: Get Service Account Credentials (5 minutes)
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing project
3. Go to **IAM & Admin** → **Service Accounts**
4. Click **+ CREATE SERVICE ACCOUNT**
5. Name: `isx-license-generator`, Role: **Editor**
6. Click **Keys** tab → **ADD KEY** → **Create new key** → **JSON**
7. **Save the downloaded file as `service-account-credentials.json`** in this folder

### Step 2: Share Your Google Sheet (1 minute)
1. Open the downloaded JSON file
2. Copy the `client_email` (looks like: `isx-license-generator@project.iam.gserviceaccount.com`)
3. Go to your Google Sheet: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit
4. Click **Share** → Add the service account email → **Editor** permission → **Send**

### Step 3: Add Column Headers (30 seconds)
Make sure **Row 1** of your sheet has these headers:
```
LicenseKey | Duration | ExpiryDate | Status | MachineID | ActivatedDate
```

### Step 4: Generate Licenses (30 seconds)
```cmd
# Build the service account version
build-service-account.bat

# Generate 100 licenses
license-generator-sa.exe -total=100
```

## ✅ That's It!
- **No OAuth2 verification needed**
- **No browser authentication**
- **No Google app approval required**
- **Works immediately**

## 🔧 Files You Need:
- ✅ `service-account-credentials.json` (from Google Cloud Console)
 - ✅ `service-account-config.json` (copy from `service-account-config-template.json`)
- ✅ `main-service-account.go` (already created)

## 🎯 Test with 1 License First:
```cmd
license-generator-sa.exe -total=1
```

## 🏆 Why This Is Better:
- **OAuth2**: Complex, requires verification, browser auth, test users
- **Service Account**: Simple, works immediately, no verification required

**Choose Service Account - it's the simplest approach!** 🎯 