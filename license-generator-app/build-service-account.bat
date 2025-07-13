@echo off
echo 🚀 Building ISX License Generator (Service Account Version)...
echo.

REM Build the service account version
go build -o license-generator-sa.exe main-service-account.go

if %errorlevel% neq 0 (
    echo ❌ Build failed!
    pause
    exit /b %errorlevel%
)

echo ✅ Build successful!
echo 📁 Created: license-generator-sa.exe
echo.
echo 🎯 Usage: license-generator-sa.exe -total=100
echo.
echo 📋 Before running, make sure you have:
echo   ✅ service-account-credentials.json (from Google Cloud Console)
echo   ✅ Shared your Google Sheet with the service account email
echo   ✅ Added column headers to your sheet
echo.
pause 