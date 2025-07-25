# PowerShell script to add MinGW to PATH and set CGO_ENABLED

Write-Host "Setting up MinGW environment variables..." -ForegroundColor Green

# Common MinGW installation paths - update if yours is different
$mingwPaths = @(
    "C:\mingw64\bin",
    "C:\tools\mingw64\bin",
    "C:\MinGW\bin",
    "C:\Program Files\mingw64\bin",
    "C:\Program Files (x86)\mingw64\bin"
)

# Find where MinGW is installed
$mingwPath = ""
foreach ($path in $mingwPaths) {
    if (Test-Path "$path\gcc.exe") {
        $mingwPath = $path
        Write-Host "Found MinGW at: $mingwPath" -ForegroundColor Green
        break
    }
}

if ($mingwPath -eq "") {
    Write-Host "MinGW not found in standard locations!" -ForegroundColor Red
    Write-Host "Please enter the path to your MinGW bin directory:" -ForegroundColor Yellow
    Write-Host "Example: C:\mingw64\bin" -ForegroundColor Gray
    $mingwPath = Read-Host "MinGW bin path"
    
    if (-not (Test-Path "$mingwPath\gcc.exe")) {
        Write-Host "ERROR: gcc.exe not found at $mingwPath" -ForegroundColor Red
        exit 1
    }
}

# Get current PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)

# Check if MinGW is already in PATH
if ($currentPath -like "*$mingwPath*") {
    Write-Host "MinGW is already in PATH" -ForegroundColor Yellow
} else {
    # Add MinGW to PATH
    Write-Host "Adding $mingwPath to User PATH..." -ForegroundColor Yellow
    $newPath = "$currentPath;$mingwPath"
    [Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::User)
    Write-Host "✓ Added to PATH" -ForegroundColor Green
    
    # Also update current session
    $env:Path = "$env:Path;$mingwPath"
}

# Set CGO_ENABLED
Write-Host "Setting CGO_ENABLED=1..." -ForegroundColor Yellow
[Environment]::SetEnvironmentVariable("CGO_ENABLED", "1", [EnvironmentVariableTarget]::User)
$env:CGO_ENABLED = "1"
Write-Host "✓ CGO_ENABLED set" -ForegroundColor Green

# Verify installation
Write-Host "`nVerifying installation..." -ForegroundColor Yellow
Write-Host "GCC location: " -NoNewline
& "$mingwPath\gcc.exe" --version | Select-Object -First 1

Write-Host "`nGo environment:" -ForegroundColor Yellow
go env CGO_ENABLED
go env CC

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nIMPORTANT: You must restart your terminal for changes to take full effect!" -ForegroundColor Yellow
Write-Host "`nTo test race detector now in PowerShell:" -ForegroundColor Cyan
Write-Host "  cd dev" -ForegroundColor White
Write-Host "  go test -race -v ./internal/websocket/... -run TestHubCreation" -ForegroundColor White