# Quick GCC setup for Windows
# This script uses winget (Windows Package Manager)

Write-Host "Quick GCC Setup for Go Race Detector" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green
Write-Host ""

# Check if winget is available
if (Get-Command winget -ErrorAction SilentlyContinue) {
    Write-Host "Installing MinGW using winget..." -ForegroundColor Yellow
    winget install -e --id MinGW.MinGW
    
    Write-Host ""
    Write-Host "Installation started. After completion:" -ForegroundColor Green
    Write-Host "1. Add C:\MinGW\bin to your PATH" -ForegroundColor White
    Write-Host "2. Set CGO_ENABLED=1" -ForegroundColor White
    Write-Host "3. Restart your terminal" -ForegroundColor White
} else {
    Write-Host "winget not found. Trying direct download..." -ForegroundColor Yellow
    
    # Direct download option
    $url = "https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z/download"
    $output = "$env:TEMP\mingw.7z"
    
    Write-Host "Downloading MinGW..."
    Invoke-WebRequest -Uri $url -OutFile $output -UseBasicParsing
    
    Write-Host "Download complete. Please:"
    Write-Host "1. Extract $output to C:\mingw64" -ForegroundColor Yellow
    Write-Host "2. Add C:\mingw64\bin to PATH" -ForegroundColor Yellow
    Write-Host "3. Set CGO_ENABLED=1" -ForegroundColor Yellow
    Write-Host "4. Restart terminal" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Manual PATH setup command:" -ForegroundColor Cyan
Write-Host '[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\mingw64\bin", [EnvironmentVariableTarget]::User)' -ForegroundColor White
Write-Host ""
Write-Host "Manual CGO setup command:" -ForegroundColor Cyan
Write-Host '[Environment]::SetEnvironmentVariable("CGO_ENABLED", "1", [EnvironmentVariableTarget]::User)' -ForegroundColor White