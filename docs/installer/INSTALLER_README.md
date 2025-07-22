# ISX Daily Reports Scraper - Windows Installer

This document provides instructions for building and distributing the Windows installer for the ISX Daily Reports Scraper.

## 🎯 **Overview**

The installer packages the complete ISX Daily Reports Scraper application into a professional Windows installer (.exe) that provides:

- **One-click installation** of all components
- **Automatic dependency management** (Visual C++ Redistributable)
- **License configuration wizard**
- **Windows Firewall configuration**
- **Start Menu shortcuts and desktop icons**
- **Professional uninstallation support**

## 📋 **Prerequisites**

### Required Software:
1. **Inno Setup 6.0+** - Download from [jrsoftware.org](https://www.jrsoftware.org/isinfo.php)
2. **Go 1.19+** - For building the application components
3. **Git** - For version control
4. **Internet connection** - For downloading Visual C++ Redistributable

### System Requirements:
- Windows 10 or Windows 11
- 4 GB RAM minimum
- 2 GB free disk space
- Administrator privileges

## 🚀 **Quick Start**

### 1. Install Prerequisites
```bash
# Install Inno Setup from https://www.jrsoftware.org/isinfo.php
# Make sure to add Inno Setup to your PATH

# Verify installation
iscc /?
```

### 2. Build the Installer
```bash
# Clone the repository
git clone https://github.com/haideralmesaody/ISXDailyReportScrapper.git
cd ISXDailyReportScrapper

# Run the build script
build-installer.bat
```

### 3. Distribute the Installer
The installer will be created in:
```
installer/output/ISX-Daily-Reports-Scraper-Setup-1.0.0.exe
```

## 📁 **Installer Structure**

```
installer/
├── isx-scraper-installer.iss      # Main Inno Setup script
├── assets/                        # Installer assets
│   ├── LICENSE.txt               # End-user license agreement
│   ├── README.txt                # Pre-installation information
│   ├── AFTER_INSTALL.txt         # Post-installation instructions
│   ├── setup-icon.ico            # Installer icon
│   ├── vc_redist.x64.exe         # Visual C++ Redistributable
│   ├── license-config-template.json # License configuration template
│   ├── app-config.json           # Application configuration
│   ├── start-web-interface.bat   # Web interface launcher
│   ├── configure-license.bat     # License configuration wizard
│   └── run-scraper.bat          # CLI launcher
└── output/                       # Generated installer files
    └── ISX-Daily-Reports-Scraper-Setup-1.0.0.exe
```

## 🔧 **Build Process**

The `build-installer.bat` script performs the following steps:

1. **Builds all Go applications**:
   - `isxcli.exe` - Main scraper
   - `web.exe` - Web interface
   - `web-licensed.exe` - Licensed web interface
   - All tool executables in `cmd/` subdirectories

2. **Prepares installer assets**:
   - Downloads Visual C++ Redistributable
   - Creates installer icon from favicon
   - Validates all required files

3. **Compiles the installer**:
   - Runs Inno Setup compiler
   - Creates signed installer package
   - Validates installer integrity

## 📦 **Installer Features**

### Installation Options:
- **Desktop icon** - Optional desktop shortcut
- **Quick Launch** - Quick launch toolbar icon
- **File associations** - Associate .xlsx files with ISX processor
- **PATH integration** - Add application to system PATH
- **Auto-start** - Start with Windows (optional)

### Installed Components:
- **Main application** - All executables and tools
- **Web interface** - Complete web dashboard
- **Documentation** - User guides and setup instructions
- **License generator** - License management tools
- **Configuration files** - Templates and defaults
- **Batch scripts** - Convenience launchers

### System Integration:
- **Windows Firewall** - Configures port 8080 access
- **Registry entries** - Application settings and file associations
- **Start Menu** - Professional shortcuts and organization
- **Uninstaller** - Clean removal of all components

## 🛠️ **Customization**

### Modify Installer Script:
Edit `installer/isx-scraper-installer.iss` to customize:

```ini
#define AppName "ISX Daily Reports Scraper"
#define AppVersion "1.0.0"
#define AppPublisher "The Iraqi Investor Group"
```

### Add Custom Files:
Add files to the `[Files]` section:
```ini
Source: "your-file.exe"; DestDir: "{app}"; Flags: ignoreversion
```

### Custom Install Options:
Add tasks to the `[Tasks]` section:
```ini
Name: "customtask"; Description: "Custom installation option"
```

### Registry Customization:
Add registry entries in the `[Registry]` section:
```ini
Root: HKCU; Subkey: "SOFTWARE\YourCompany\YourApp"; ValueType: string; ValueName: "Setting"; ValueData: "Value"
```

## 🔐 **License Integration**

The installer includes comprehensive license management:

### License Configuration:
- **Interactive wizard** during installation
- **Google Sheets integration** setup
- **License key validation**
- **Machine ID binding**

### License Files:
- `license-config.json` - Google Sheets configuration
- `license.dat` - Local license data
- `configure-license.bat` - License setup wizard

## 📋 **Distribution**

### For End Users:
1. Download the installer from your distribution channel
2. Run as Administrator
3. Follow the installation wizard
4. Configure license when prompted
5. Launch the web interface

### For Developers:
1. Build the installer using `build-installer.bat`
2. Test installation on clean Windows systems
3. Distribute through your chosen channels
4. Provide license keys to customers

## 🐛 **Troubleshooting**

### Build Issues:

#### "iscc.exe not found"
- Install Inno Setup from the official website
- Add Inno Setup to your system PATH
- Restart command prompt and try again

#### "Go build failed"
- Ensure Go 1.19+ is installed
- Run `go mod tidy` to resolve dependencies
- Check for compilation errors in the output

#### "Visual C++ download failed"
- Download manually from [Microsoft](https://aka.ms/vs/17/release/vc_redist.x64.exe)
- Place in `installer/assets/vc_redist.x64.exe`
- Re-run the build script

### Installation Issues:

#### "Administrator privileges required"
- Right-click installer and select "Run as administrator"
- Ensure Windows UAC is properly configured

#### "Port 8080 already in use"
- Check for other applications using port 8080
- Use `netstat -an | findstr :8080` to identify conflicts
- Stop conflicting services or change the port

#### "License activation failed"
- Verify internet connection
- Check Google Sheets configuration
- Ensure license key format is correct

## 🔄 **Updates**

### Version Management:
1. Update version in `installer/isx-scraper-installer.iss`
2. Update version in `go.mod` and application code
3. Rebuild all components
4. Test the new installer
5. Distribute updated installer

### Upgrade Support:
The installer supports:
- **In-place upgrades** - Preserves user data and settings
- **Settings migration** - Maintains configuration across versions
- **Rollback capability** - Uninstall and reinstall previous version

## 📞 **Support**

For installer-related issues:
- **GitHub Issues**: [ISXDailyReportScrapper Issues](https://github.com/haideralmesaody/ISXDailyReportScrapper/issues)
- **Email Support**: support@iraqiinvestor.com
- **Documentation**: See `docs/` folder in installation

## 📄 **License**

This installer and the ISX Daily Reports Scraper are licensed under the terms specified in the End User License Agreement (EULA) included with the software.

---

**The Iraqi Investor Group**  
*Professional Iraqi Stock Exchange Data Solutions* 