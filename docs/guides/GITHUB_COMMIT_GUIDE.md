# GitHub Commit Strategy for ISX Daily Reports Scraper

This guide explains what files to commit to GitHub vs exclude for our small installer approach.

## 🎯 **Strategy Overview**

**Small Installer Approach:**
- **Installer**: 1-2 MB small installer that downloads from GitHub
- **Source Code**: All source files committed to GitHub
- **Releases**: Binary releases created via GitHub Actions or manual uploads
- **Distribution**: Users download small installer, which gets latest release

## ✅ **Files TO COMMIT (Include in Repository)**

### **Source Code Files**
```
✅ *.go                    # All Go source files
✅ go.mod                  # Go module definition  
✅ go.sum                  # Go module checksums
✅ main.go                 # Main application entry point
✅ cmd/**/*.go            # All command-line tools source
✅ internal/**/*.go       # Internal package source code
```

### **Documentation**
```
✅ README.md              # Main project documentation
✅ *.md                   # All markdown documentation files
✅ LICENSING_SETUP.md     # License setup guide
✅ EXPIRE_STATUS_SETUP.md # ExpireStatus guide
✅ GOOGLE_SHEETS_SETUP.md # Google Sheets setup
✅ WEB_INTERFACE_GUIDE.md # Web interface guide
✅ INSTALLER_README.md    # Installer documentation
✅ GITHUB_COMMIT_GUIDE.md # This file
```

### **Web Interface Files**
```
✅ web/                   # Complete web interface directory
✅ web/index.html         # Main web page
✅ web/license.html       # License activation page
✅ web/static/            # All static assets (CSS, JS, images)
✅ web/static/images/     # Icons and logos
```

### **Installer Assets**
```
✅ installer/             # Installer directory
✅ installer/*.iss        # Inno Setup scripts
✅ installer/assets/      # Installer assets directory
✅ installer/assets/*.txt # License, README, after-install text
✅ installer/assets/*.bat # Batch scripts for shortcuts
✅ installer/assets/*.json # Configuration templates
✅ installer/assets/*.ps1 # PowerShell download script
```

### **Build Scripts**
```
✅ build-web.bat          # Build script for web interface
✅ build-installer.bat    # Original full installer build
✅ build-github-installer.bat # Small GitHub installer build
```

### **Configuration Templates**
```
✅ installer/assets/license-config-template.json # License config template
✅ installer/assets/app-config.json              # Application config template
```

### **Source Control Files**
```
✅ .gitignore             # Git ignore rules
✅ .github/               # GitHub Actions workflows (if any)
```

## ❌ **Files NOT TO COMMIT (Exclude from Repository)**

### **Compiled Binaries**
```
❌ *.exe                  # All executable files
❌ *.dll                  # Dynamic link libraries
❌ *.bin                  # Binary files
❌ *.out                  # Go build output
❌ isxcli.exe            # Main application executable
❌ web.exe               # Web interface executable
❌ web-licensed.exe      # Licensed web interface
❌ cmd/**/*.exe          # All tool executables
```

### **Generated Data Files**
```
❌ downloads/             # Downloaded Excel reports (user data)
❌ reports/               # Generated CSV reports (user data)
❌ logs/                  # Application log files
❌ temp/                  # Temporary files
❌ *.csv                  # Generated CSV files
❌ *.xlsx                 # Excel files (downloaded reports)
❌ indexes.csv           # Generated index files
❌ formats.json          # Generated format files
```

### **License and Configuration Files with Real Data**
```
❌ license.dat            # Actual license file with real data
❌ license-config.json    # Real license configuration
❌ *.license             # Any license files with real keys
```

### **Build Artifacts and Dependencies**
```
❌ vendor/               # Go vendor directory
❌ bin/                  # Build output directory
❌ dist/                 # Distribution directory
❌ build/                # Build artifacts
❌ installer/output/     # Generated installer files
❌ installer/assets/vc_redist.x64.exe # Large dependency (downloaded)
```

### **IDE and OS Files**
```
❌ .vscode/              # VS Code settings
❌ .idea/                # IntelliJ IDEA settings
❌ *.swp, *.swo         # Vim swap files
❌ .DS_Store            # macOS system files
❌ Thumbs.db            # Windows system files
```

### **Environment and Debug Files**
```
❌ .env                  # Environment variables
❌ debug*.txt           # Debug output files
❌ test_*               # Test output files
❌ *.log                # Log files
❌ *.tmp                # Temporary files
```

## 📋 **Current .gitignore Configuration**

Our `.gitignore` file is already configured to exclude the right files:

```gitignore
# Compiled Binaries and Executables
*.exe
*.dll
*.so
*.dylib

# Generated Data Files
downloads/
reports/
logs/

# License Files with Real Data
license.dat
license-config.json

# Build Artifacts
bin/
dist/
build/
vendor/

# Installer Output
installer/output/

# Large Dependencies
installer/assets/vc_redist.x64.exe
```

## 🚀 **Commit Process**

### **1. Clean Up Existing Files**
```bash
# Remove any generated files that shouldn't be committed
git rm --cached *.exe
git rm --cached license.dat
git rm --cached license-config.json
git rm --cached installer/output/*
```

### **2. Stage Files for Commit**
```bash
# Add all source code and documentation
git add *.go
git add *.md
git add web/
git add installer/assets/
git add cmd/
git add internal/

# Add build scripts
git add build-*.bat
git add installer/*.iss
```

### **3. Commit to GitHub**
```bash
git commit -m "feat: Add licensing system and GitHub-based installer

- Complete licensing system with Google Sheets integration
- Small installer that downloads from GitHub releases
- Professional web interface with license management
- Comprehensive documentation and setup guides
- ExpireStatus tracking for license monitoring"

git push origin feature/add-licenses
```

## 📦 **GitHub Release Strategy**

### **Release Contents**
When creating GitHub releases, include:

```
📦 ISX-Daily-Reports-Scraper-v1.0.0-windows.zip
├── isxcli.exe                    # Main application
├── web.exe                       # Web interface
├── web-licensed.exe              # Licensed web interface
├── tools/                        # Tool executables
│   ├── process.exe
│   ├── indexcsv.exe
│   ├── license-generator.exe
│   └── bulk-license-generator.exe
├── web/                          # Web interface files
├── docs/                         # Documentation
├── license-generator/            # License generator app
└── *.bat                        # Batch scripts
```

### **Release Workflow**
1. **Build all executables**:
   ```bash
   go build -ldflags="-s -w" -o isxcli.exe .
   go build -ldflags="-s -w" -o web.exe ./cmd/web
   # ... build all other tools
   ```

2. **Create release ZIP**:
   ```bash
   # Create a release package with all executables and assets
   # Exclude source code, include only built binaries and web files
   ```

3. **Upload to GitHub Releases**:
   - Create new release on GitHub
   - Upload the ZIP file as a release asset
   - Use semantic versioning (v1.0.0, v1.0.1, etc.)

## 🔧 **Benefits of This Approach**

### **For Repository:**
- **Small size** - No large binaries in git history
- **Clean history** - Only source code changes tracked
- **Fast cloning** - Quick download for developers
- **Professional** - Industry standard approach

### **For Distribution:**
- **Small installer** - 1-2 MB vs 50+ MB
- **Always latest** - Automatically gets newest release
- **Bandwidth efficient** - Users only download when needed
- **Update friendly** - Easy to distribute updates

### **For Development:**
- **Faster CI/CD** - No large files in pipelines
- **Better collaboration** - Developers work with source only
- **Version control** - Clear separation of source vs builds
- **Professional workflow** - Standard software development practice

## 📞 **Next Steps**

1. **Review and clean** your current repository
2. **Commit source code** following this guide
3. **Create first GitHub release** with built binaries
4. **Test the small installer** to ensure it downloads correctly
5. **Distribute** the small installer to end users

---

**The Iraqi Investor Group**  
*Professional Iraqi Stock Exchange Data Solutions* 