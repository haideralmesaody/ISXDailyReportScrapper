# 🚀 ISX Alpha Release - Deployment Guide

## 📦 **Quick Start - Recommended Installation**

### **Step 1: Download & Extract**
1. Download the `ISX-Alpha-Release.zip` package
2. Extract to a temporary folder (e.g., `C:\Temp\ISX-Alpha\`)
3. **Do NOT manually copy files** - use the installer instead!

### **Step 2: Run Professional Installer**

**Option A: Executable Installer (Recommended)**
```
1. Double-click "ISX-Alpha-Installer.exe"
2. Follow the setup wizard
3. Installer requires Administrator privileges (auto-prompt)
4. Select installation options and proceed
```

**Option B: PowerShell Installer**
```powershell
# Right-click PowerShell → "Run as Administrator"
cd C:\Temp\ISX-Alpha
.\install-alpha.ps1
```

**The installer will:**
- ✅ Check system requirements (Windows 10+, Chrome, etc.)
- ✅ Create proper directory structure in `C:\Program Files\ISX\`
- ✅ Copy all files to correct locations
- ✅ Set up Windows PATH environment
- ✅ Create desktop shortcuts
- ✅ Configure Windows Firewall rules
- ✅ Verify installation integrity

---

## 🎯 **After Installation - What To Run**

### **Option 1: Desktop Shortcuts (Recommended)**
The installer creates these shortcuts on your desktop:
- **`ISX Web Interface`** - Main application (double-click to start)
- **`ISX Command Line`** - Advanced CLI operations

### **Option 2: Start Menu**
- Search for "ISX" in Windows Start Menu
- Click "ISX Web Interface" or "ISX CLI"

### **Option 3: Command Line**
```powershell
# From anywhere in PowerShell/CMD:
isx-web-interface    # Starts web application
isxcli --help        # Shows CLI commands
```

---

## 📁 **Installation Directory Structure**

### **Default Installation Location:**
```
C:\Program Files\ISX\
├── bin\
│   ├── isx-web-interface.exe  ← Main web application
│   └── isxcli.exe            ← Command-line tool
├── web\
│   ├── index.html            ← Web interface files
│   ├── license.html
│   └── static\               ← CSS, JS, images
├── docs\
│   ├── ALPHA-USER-GUIDE.md   ← Complete user manual
│   └── ALPHA-TESTING-GUIDE.md ← Testing instructions
├── data\
│   ├── downloads\            ← Downloaded ISX data
│   └── reports\              ← Generated reports
└── tools\
    └── process.exe           ← Data processing utilities
```

### **User Data Location:**
```
C:\Users\[Username]\AppData\Local\ISX\
├── license.dat               ← License file
├── config\                   ← User settings
└── logs\                     ← Application logs
```

---

## ⚡ **First Run Instructions**

### **1. Start the Application:**
- Double-click **"ISX Web Interface"** desktop shortcut
- OR run `isx-web-interface` from command line

### **2. License Activation:**
- Browser opens automatically to `http://localhost:8080`
- Click **"License"** tab
- Enter your alpha testing license key
- Complete activation process

### **3. Begin Testing:**
- Navigate to **"Data Collection"** tab
- Click **"Start Scraping"** to test data collection
- Monitor progress and report any issues

---

## 🔧 **Manual Installation (Advanced Users)**

If you prefer manual installation or the automatic installer fails:

### **Step 1: Create Directory Structure**
```powershell
# Create main directory
mkdir "C:\Program Files\ISX"
cd "C:\Program Files\ISX"

# Create subdirectories
mkdir bin, web, docs, data, tools
mkdir data\downloads, data\reports
```

### **Step 2: Copy Files Manually**
```powershell
# Copy executables
copy release\bin\*.exe "C:\Program Files\ISX\bin\"

# Copy web assets
xcopy release\web "C:\Program Files\ISX\web\" /E /I

# Copy documentation
xcopy release\docs "C:\Program Files\ISX\docs\" /E /I

# Copy tools
copy release\tools\*.exe "C:\Program Files\ISX\tools\"
```

### **Step 3: Add to PATH**
```powershell
# Add ISX to system PATH
$oldPath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
$newPath = $oldPath + ';C:\Program Files\ISX\bin'
[Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine')
```

### **Step 4: Create Shortcuts**
```powershell
# Desktop shortcut for web interface
$WshShell = New-Object -comObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$Home\Desktop\ISX Web Interface.lnk")
$Shortcut.TargetPath = "C:\Program Files\ISX\bin\isx-web-interface.exe"
$Shortcut.Save()
```

---

## 🔍 **Verification & Troubleshooting**

### **Verify Installation:**
```powershell
# Check if executables are accessible
isx-web-interface --version
isxcli --version

# Verify file structure
dir "C:\Program Files\ISX" /s
```

### **Common Issues:**

#### **"Command not found" Error**
- **Cause**: PATH not updated
- **Solution**: Restart PowerShell/CMD or reboot computer
- **Alternative**: Use full path: `"C:\Program Files\ISX\bin\isx-web-interface.exe"`

#### **"Access Denied" Error**
- **Cause**: Insufficient permissions
- **Solution**: Run PowerShell as Administrator
- **Alternative**: Install to user directory instead of Program Files

#### **"Port 8080 in use" Error**
- **Cause**: Another application using port 8080
- **Solution**: Close other applications or restart computer
- **Alternative**: Kill process using port: `netstat -ano | findstr :8080`

#### **License Activation Fails**
- **Cause**: Network connectivity or invalid license
- **Solution**: Check internet connection and verify license key
- **Support**: Contact alpha testing coordinator

---

## 📞 **Alpha Testing Support**

### **During Alpha Testing:**
- **Documentation**: Check `docs\ALPHA-USER-GUIDE.md` for detailed instructions
- **Testing Guide**: Follow `docs\ALPHA-TESTING-GUIDE.md` for structured testing
- **Bug Reports**: Use the bug report template in testing guide
- **Direct Support**: Contact alpha testing coordinator

### **Log Files Location:**
```
C:\Users\[Username]\AppData\Local\ISX\logs\
├── application.log           ← Main application logs
├── license.log              ← License system logs
└── error.log                ← Error details
```

---

## 🎯 **Summary - What Users Should Do**

### **For Most Users:**
1. **Extract** alpha package to temporary folder
2. **Run** `ISX-Alpha-Installer.exe` (or `install-alpha.ps1` as Administrator)
3. **Double-click** "ISX Web Interface" desktop shortcut
4. **Activate** license in browser interface
5. **Begin** testing following the alpha testing guide

### **Key Files Users Interact With:**
- **`install-alpha.ps1`** - Run this FIRST (one-time setup)
- **`ISX Web Interface` shortcut** - Run this to use the application
- **`docs\ALPHA-USER-GUIDE.md`** - Read for detailed instructions
- **`docs\ALPHA-TESTING-GUIDE.md`** - Follow for structured testing

### **Users Should NOT manually run:**
- Individual `.exe` files from the package (use installer instead)
- Files from temporary extraction folder (use installed versions)
- Configuration scripts (installer handles these)

---

**🚀 Ready to revolutionize ISX data analysis! Follow this guide for a smooth alpha testing experience.** 