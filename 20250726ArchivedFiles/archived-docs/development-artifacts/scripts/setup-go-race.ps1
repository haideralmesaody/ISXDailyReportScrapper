# PowerShell script to setup Go race detector environment

Write-Host "Setting up Go race detector environment..." -ForegroundColor Green

# Find GCC installation
$gccPaths = @(
    "C:\ProgramData\chocolatey\bin\gcc.exe",
    "C:\tools\mingw64\bin\gcc.exe",
    "C:\tools\mingw32\bin\gcc.exe",
    "C:\MinGW\bin\gcc.exe",
    "C:\msys64\mingw64\bin\gcc.exe"
)

$gccFound = $false
$gccPath = ""

foreach ($path in $gccPaths) {
    if (Test-Path $path) {
        $gccFound = $true
        $gccPath = Split-Path $path -Parent
        Write-Host "Found GCC at: $path" -ForegroundColor Green
        break
    }
}

if (-not $gccFound) {
    # Try to find it dynamically
    Write-Host "Searching for GCC installation..." -ForegroundColor Yellow
    $searchPaths = @("C:\ProgramData\chocolatey", "C:\tools", "C:\")
    
    foreach ($searchPath in $searchPaths) {
        if (Test-Path $searchPath) {
            $found = Get-ChildItem -Path $searchPath -Filter "gcc.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($found) {
                $gccFound = $true
                $gccPath = $found.DirectoryName
                Write-Host "Found GCC at: $($found.FullName)" -ForegroundColor Green
                break
            }
        }
    }
}

if (-not $gccFound) {
    Write-Host "ERROR: GCC not found!" -ForegroundColor Red
    Write-Host "Please ensure MinGW was installed correctly." -ForegroundColor Red
    Write-Host "Try running: choco install mingw -y" -ForegroundColor Yellow
    exit 1
}

# Add to PATH for current session
Write-Host "`nAdding GCC to PATH..." -ForegroundColor Yellow
$env:Path = "$gccPath;$env:Path"

# Set CGO_ENABLED
Write-Host "Setting CGO_ENABLED=1..." -ForegroundColor Yellow
$env:CGO_ENABLED = "1"

# Verify
Write-Host "`nVerifying installation..." -ForegroundColor Yellow
Write-Host "GCC Version:" -ForegroundColor Cyan
& "$gccPath\gcc.exe" --version | Select-Object -First 1

Write-Host "`nGo CGO settings:" -ForegroundColor Cyan
go env CGO_ENABLED
go env CC

# Test race detector
Write-Host "`nTesting race detector..." -ForegroundColor Yellow
Set-Location "dev"

$testFile = "test_race.go"
@"
package main
import "testing"
func TestRace(t *testing.T) {}
"@ | Out-File -FilePath $testFile -Encoding UTF8

$output = go test -race $testFile 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Race detector is working!" -ForegroundColor Green
} else {
    Write-Host "✗ Race detector test failed" -ForegroundColor Red
    Write-Host $output
}
Remove-Item $testFile -ErrorAction SilentlyContinue

Write-Host "`nTo make these changes permanent:" -ForegroundColor Yellow
Write-Host "1. Add to System PATH: $gccPath" -ForegroundColor White
Write-Host "2. Set System Environment Variable: CGO_ENABLED=1" -ForegroundColor White

Write-Host "`nYou can now run tests with:" -ForegroundColor Green
Write-Host "  go test -race -v ./internal/websocket/..." -ForegroundColor White