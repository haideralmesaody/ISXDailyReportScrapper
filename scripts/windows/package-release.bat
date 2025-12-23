@echo off
:: ISX Pulse - Package Release
:: The Heartbeat of Iraqi Markets
:: Creates a distributable ZIP file

set VERSION=0.0.1-alpha
set RELEASE_NAME=ISXPulse-v%VERSION%-win64

echo ==========================================
echo         Packaging ISX Pulse v%VERSION%
echo     The Heartbeat of Iraqi Markets
echo ==========================================
echo.

:: Check if dist directory exists
if not exist dist (
    echo [ERROR] Dist directory not found. Run build first.
    pause
    exit /b 1
)

:: Check if PowerShell is available for zipping
where powershell >nul 2>nul
if errorlevel 1 (
    echo [ERROR] PowerShell not found. Cannot create ZIP file.
    echo Please create ZIP manually from the dist directory.
    pause
    exit /b 1
)

echo Creating %RELEASE_NAME%.zip...

:: Remove old zip if exists
if exist %RELEASE_NAME%.zip del %RELEASE_NAME%.zip

:: Create zip using PowerShell
powershell -Command "Compress-Archive -Path 'dist\*' -DestinationPath '%RELEASE_NAME%.zip' -Force"

if errorlevel 1 (
    echo [ERROR] Failed to create ZIP file
    pause
    exit /b 1
)

:: Get file size
for %%A in (%RELEASE_NAME%.zip) do set SIZE=%%~zA
set /a SIZE_MB=%SIZE% / 1048576

echo.
echo ==========================================
echo Release package created successfully!
echo ==========================================
echo.
echo File: %RELEASE_NAME%.zip
echo Size: ~%SIZE_MB% MB
echo.
echo The package contains:
echo - 4 executables (web, scraper, processor, indexcsv)
echo - Configuration templates
echo - Directory structure
echo - README and documentation
echo.
echo Ready for distribution!
echo.
pause