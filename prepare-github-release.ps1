# Prepare GitHub Release Assets for Smart Installer
# This script creates the ZIP files needed for GitHub releases

param(
    [string]$Version = "v1.0-alpha"
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   Preparing GitHub Release Assets" -ForegroundColor Cyan
Write-Host "   Version: $Version" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host

# Check if release directory exists
if (-not (Test-Path "release")) {
    Write-Host "❌ ERROR: release directory not found!" -ForegroundColor Red
    Write-Host "Please build the release package first." -ForegroundColor Yellow
    exit 1
}

# Create output directory for GitHub assets
$assetsDir = "github-release-assets"
if (Test-Path $assetsDir) {
    Remove-Item $assetsDir -Recurse -Force
}
New-Item -ItemType Directory -Path $assetsDir | Out-Null

Write-Host "📁 Created assets directory: $assetsDir" -ForegroundColor Green

# 1. Copy main executables
Write-Host "📋 Copying main executables..." -ForegroundColor Yellow

if (Test-Path "release\bin\isx-web-interface.exe") {
    Copy-Item "release\bin\isx-web-interface.exe" "$assetsDir\"
    $size = (Get-Item "$assetsDir\isx-web-interface.exe").Length / 1MB
    Write-Host "  ✅ isx-web-interface.exe ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
} else {
    Write-Host "  ❌ isx-web-interface.exe not found!" -ForegroundColor Red
}

if (Test-Path "release\bin\isxcli.exe") {
    Copy-Item "release\bin\isxcli.exe" "$assetsDir\"
    $size = (Get-Item "$assetsDir\isxcli.exe").Length / 1MB
    Write-Host "  ✅ isxcli.exe ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
} else {
    Write-Host "  ❌ isxcli.exe not found!" -ForegroundColor Red
}

# 2. Create web-assets.zip
Write-Host "📦 Creating web-assets.zip..." -ForegroundColor Yellow

if (Test-Path "release\web") {
    $webZip = "$assetsDir\web-assets.zip"
    Compress-Archive -Path "release\web\*" -DestinationPath $webZip -Force
    $size = (Get-Item $webZip).Length / 1KB
    Write-Host "  ✅ web-assets.zip ($([math]::Round($size, 1)) KB)" -ForegroundColor Green
} else {
    Write-Host "  ❌ release\web directory not found!" -ForegroundColor Red
}

# 3. Create docs.zip
Write-Host "📚 Creating docs.zip..." -ForegroundColor Yellow

if (Test-Path "release\docs") {
    $docsZip = "$assetsDir\docs.zip"
    Compress-Archive -Path "release\docs\*" -DestinationPath $docsZip -Force
    $size = (Get-Item $docsZip).Length / 1KB
    Write-Host "  ✅ docs.zip ($([math]::Round($size, 1)) KB)" -ForegroundColor Green
} else {
    Write-Host "  ❌ release\docs directory not found!" -ForegroundColor Red
}

# 4. Create tools.zip (optional)
Write-Host "🔧 Creating tools.zip..." -ForegroundColor Yellow

if (Test-Path "release\tools") {
    $toolsZip = "$assetsDir\tools.zip"
    Compress-Archive -Path "release\tools\*" -DestinationPath $toolsZip -Force
    $size = (Get-Item $toolsZip).Length / 1KB
    Write-Host "  ✅ tools.zip ($([math]::Round($size, 1)) KB)" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  release\tools directory not found (optional)" -ForegroundColor Yellow
}

# 5. Generate release notes
Write-Host "📝 Generating release notes..." -ForegroundColor Yellow

$releaseNotes = @"
# 🚀 ISX Daily Reports Scraper - Alpha Release $Version

## 🎯 **What's New**
- Complete automation pipeline (scraping → processing → visualization)
- Professional web interface with 4-tab design  
- Enterprise license system with Google Sheets integration
- Real-time WebSocket updates and responsive design
- Professional installer with comprehensive documentation

## 📦 **Installation**
Download and run the **ISX-Smart-Installer.exe** for automatic installation.

### Smart Installer Features:
- ⚡ Downloads latest components automatically
- 🌐 Installs Chrome browser if needed
- 🔧 Sets up PATH environment and shortcuts
- 📊 Professional setup wizard interface
- ✅ Always gets the most current version

## 🧪 **For Alpha Testers**
1. Download **ISX-Smart-Installer.exe** (small download ~2-5MB)
2. Run as Administrator
3. Follow the setup wizard
4. Use desktop shortcut to start ISX
5. Follow testing guide in docs/

## 📋 **Package Contents**
- **isx-web-interface.exe** - Main web application
- **isxcli.exe** - Command-line interface
- **web-assets.zip** - Web interface files
- **docs.zip** - Complete documentation and testing guides

## 🔄 **System Requirements**
- Windows 10 or newer
- Internet connection (for installation)
- Chrome browser (auto-installed if needed)
- 500MB free disk space

## 📞 **Support**
- 📖 User Guide: Check docs/ after installation
- 🧪 Testing Guide: Follow structured testing plan
- 🐛 Issues: Report on GitHub Issues
- 📧 Contact: Alpha testing coordinator

---
**🎯 Ready to revolutionize ISX data analysis!**
"@

$releaseNotes | Out-File "$assetsDir\RELEASE_NOTES.md" -Encoding UTF8
Write-Host "  ✅ RELEASE_NOTES.md created" -ForegroundColor Green

# 6. Create GitHub CLI command
Write-Host "🌐 Generating GitHub CLI command..." -ForegroundColor Yellow

$ghCommand = @"
# GitHub CLI command to create release
gh release create $Version \
  $assetsDir\isx-web-interface.exe \
  $assetsDir\isxcli.exe \
  $assetsDir\web-assets.zip \
  $assetsDir\docs.zip \
  --title "ISX Alpha Release $Version" \
  --notes-file $assetsDir\RELEASE_NOTES.md \
  --prerelease
"@

$ghCommand | Out-File "$assetsDir\github-release-command.txt" -Encoding UTF8
Write-Host "  ✅ github-release-command.txt created" -ForegroundColor Green

# 7. Summary
Write-Host
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   📦 GITHUB RELEASE ASSETS READY!" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

Write-Host
Write-Host "📁 Assets created in: $assetsDir" -ForegroundColor Green
Write-Host

# List all created files with sizes
Get-ChildItem $assetsDir | ForEach-Object {
    $size = if ($_.Length -gt 1MB) { 
        "$([math]::Round($_.Length / 1MB, 1)) MB" 
    } else { 
        "$([math]::Round($_.Length / 1KB, 1)) KB" 
    }
    Write-Host "  📄 $($_.Name) - $size" -ForegroundColor White
}

$totalSize = (Get-ChildItem $assetsDir | Measure-Object -Property Length -Sum).Sum
$totalSizeMB = [math]::Round($totalSize / 1MB, 1)
Write-Host
Write-Host "📊 Total size: $totalSizeMB MB" -ForegroundColor Cyan

Write-Host
Write-Host "🚀 Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Review files in '$assetsDir' directory" -ForegroundColor White
Write-Host "  2. Go to: https://github.com/haideralmesaody/ISXDailyReportScrapper/releases" -ForegroundColor White
Write-Host "  3. Click 'Create a new release'" -ForegroundColor White
Write-Host "  4. Upload the 4 main files (.exe and .zip files)" -ForegroundColor White
Write-Host "  5. Use RELEASE_NOTES.md as release description" -ForegroundColor White
Write-Host "  6. Check 'This is a pre-release'" -ForegroundColor White
Write-Host "  7. Publish release" -ForegroundColor White
Write-Host
Write-Host "OR use GitHub CLI:" -ForegroundColor Yellow
Write-Host "  See: $assetsDir\github-release-command.txt" -ForegroundColor White
Write-Host
Write-Host "🎯 Then build ISX-Smart-Installer.exe and distribute!" -ForegroundColor Green 