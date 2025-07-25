# Download correct MinGW-w64 binaries (not source code)

Write-Host "Downloading MinGW-w64 BINARIES (compiled version)..." -ForegroundColor Green
Write-Host ""

# Create directory
$installPath = "C:\mingw64"
if (Test-Path $installPath) {
    Write-Host "Removing old installation at $installPath..." -ForegroundColor Yellow
    Remove-Item -Path $installPath -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Creating directory: $installPath" -ForegroundColor Yellow
New-Item -ItemType Directory -Path $installPath -Force | Out-Null

# Download URL for pre-compiled binaries
$downloadUrl = "https://github.com/niXman/mingw-builds-binaries/releases/download/13.2.0-rt_v11-rev0/x86_64-13.2.0-release-posix-seh-ucrt-rt_v11-rev0.7z"
$tempFile = "$env:TEMP\mingw-binaries.7z"

Write-Host "Downloading MinGW-w64 binaries (about 50MB)..." -ForegroundColor Yellow
Write-Host "This may take a few minutes..." -ForegroundColor Gray

try {
    $ProgressPreference = 'SilentlyContinue'  # Faster download
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    Write-Host "✓ Download complete!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Download failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# Extract using 7-Zip or PowerShell
Write-Host "`nExtracting files..." -ForegroundColor Yellow

# Try 7-Zip first
$7zipPath = @(
    "${env:ProgramFiles}\7-Zip\7z.exe",
    "${env:ProgramFiles(x86)}\7-Zip\7z.exe",
    "$env:LOCALAPPDATA\7-Zip\7z.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1

if ($7zipPath) {
    Write-Host "Using 7-Zip to extract..." -ForegroundColor Gray
    & $7zipPath x $tempFile -o"C:\" -y | Out-Null
    Write-Host "✓ Extraction complete!" -ForegroundColor Green
} else {
    Write-Host "7-Zip not found. Please install 7-Zip from https://www.7-zip.org/" -ForegroundColor Red
    Write-Host "The downloaded file is at: $tempFile" -ForegroundColor Yellow
    Write-Host "Extract it manually to C:\ so you have C:\mingw64\bin\gcc.exe" -ForegroundColor Yellow
    exit 1
}

# Verify installation
if (Test-Path "C:\mingw64\bin\gcc.exe") {
    Write-Host "`n✓ MinGW-w64 binaries installed successfully!" -ForegroundColor Green
    Write-Host "  GCC location: C:\mingw64\bin\gcc.exe" -ForegroundColor Gray
    
    # Show version
    Write-Host "`nGCC Version:" -ForegroundColor Yellow
    & "C:\mingw64\bin\gcc.exe" --version | Select-Object -First 1
    
    Write-Host "`nNow run set-mingw-env.ps1 to set up environment variables!" -ForegroundColor Cyan
} else {
    Write-Host "ERROR: Installation failed. gcc.exe not found!" -ForegroundColor Red
}

# Clean up
Remove-Item $tempFile -ErrorAction SilentlyContinue