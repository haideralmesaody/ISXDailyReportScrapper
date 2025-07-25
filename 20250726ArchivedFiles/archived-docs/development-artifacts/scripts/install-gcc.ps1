# PowerShell script to install GCC/MinGW for Go race detector support
# Run this script as Administrator

Write-Host "Installing MinGW GCC compiler for Go race detector support..." -ForegroundColor Green

# Check if running as administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))
{
    Write-Host "This script must be run as Administrator. Please run PowerShell as Administrator and try again." -ForegroundColor Red
    exit 1
}

# Install MinGW via Chocolatey
Write-Host "Installing MinGW..." -ForegroundColor Yellow
choco install mingw -y

# Refresh environment variables
Write-Host "Refreshing environment variables..." -ForegroundColor Yellow
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

# Set CGO_ENABLED permanently
Write-Host "Setting CGO_ENABLED=1 permanently..." -ForegroundColor Yellow
[Environment]::SetEnvironmentVariable("CGO_ENABLED", "1", [EnvironmentVariableTarget]::User)

Write-Host "`nInstallation complete!" -ForegroundColor Green
Write-Host "Please restart your Git Bash/terminal for changes to take effect." -ForegroundColor Yellow
Write-Host "`nAfter restarting, you can verify the installation with:" -ForegroundColor Cyan
Write-Host "  gcc --version" -ForegroundColor White
Write-Host "  go env CGO_ENABLED" -ForegroundColor White
Write-Host "`nThen run tests with race detector:" -ForegroundColor Cyan
Write-Host "  cd dev" -ForegroundColor White
Write-Host "  go test -race -v ./internal/websocket/..." -ForegroundColor White