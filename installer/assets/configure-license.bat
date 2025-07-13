@echo off
title ISX Daily Reports Scraper - License Configuration
color 0B

echo.
echo ================================================
echo  ISX Daily Reports Scraper - License Setup
echo  The Iraqi Investor Group
echo ================================================
echo.

cd /d "%~dp0"

echo This wizard will help you configure your license settings.
echo.

REM Check if license-config.json exists
if exist "license-config.json" (
    echo Current license configuration found.
    echo.
    choice /c YN /m "Do you want to reconfigure your license settings"
    if errorlevel 2 goto :check_license
    echo.
)

echo.
echo ==============================================
echo  STEP 1: Google Sheets Configuration
echo ==============================================
echo.
echo You need to set up Google Sheets integration for license management.
echo.
echo Required information:
echo - Google Sheets API Key
echo - Google Sheet ID (from the sheet URL)
echo - Sheet name (usually "Licenses")
echo.
echo For detailed setup instructions, see: docs\GOOGLE_SHEETS_SETUP.md
echo.
pause

echo.
set /p SHEET_ID=Enter your Google Sheet ID: 
set /p API_KEY=Enter your Google Sheets API Key: 
set /p SHEET_NAME=Enter sheet name (default: Licenses): 

if "%SHEET_NAME%"=="" set SHEET_NAME=Licenses

echo.
echo Creating license configuration file...

REM Create license-config.json
(
echo {
echo   "sheet_id": "%SHEET_ID%",
echo   "api_key": "%API_KEY%",
echo   "sheet_name": "%SHEET_NAME%"
echo }
) > license-config.json

echo Configuration file created successfully!
echo.

:check_license
echo.
echo ==============================================
echo  STEP 2: License Activation
echo ==============================================
echo.

REM Check if license is already activated
if exist "license.dat" (
    echo License file found. Checking license status...
    isxcli.exe -help >nul 2>&1
    if errorlevel 1 (
        echo License appears to be invalid or expired.
        echo.
        choice /c YN /m "Do you want to activate a new license"
        if errorlevel 2 goto :finish
    ) else (
        echo License appears to be valid.
        echo.
        choice /c YN /m "Do you want to activate a different license"
        if errorlevel 2 goto :finish
    )
)

echo.
echo Please enter your license key.
echo License key format: ISX1M-XXXXX, ISX3M-XXXXX, ISX6M-XXXXX, or ISX1Y-XXXXX
echo.
set /p LICENSE_KEY=Enter license key: 

if "%LICENSE_KEY%"=="" (
    echo No license key entered. Skipping license activation.
    goto :finish
)

echo.
echo Activating license...
echo.

REM Try to activate license using the CLI
isxcli.exe --activate-license "%LICENSE_KEY%"

if errorlevel 1 (
    echo.
    echo License activation failed. Please check:
    echo - Internet connection is working
    echo - License key is correct and valid
    echo - Google Sheets configuration is correct
    echo.
    echo You can try again later or contact support.
    pause
    goto :finish
)

echo.
echo License activated successfully!
echo.

:finish
echo.
echo ==============================================
echo  Configuration Complete!
echo ==============================================
echo.
echo Your ISX Daily Reports Scraper is now configured.
echo.
echo Next steps:
echo 1. Launch the web interface from the Start Menu
echo 2. Visit http://localhost:8080 in your browser
echo 3. Start scraping ISX data!
echo.
echo For help and documentation, see the docs folder.
echo.
pause 