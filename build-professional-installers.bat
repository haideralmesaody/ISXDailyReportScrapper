@echo off
title Building Professional ISX Daily Reports Installers
color 0A

echo =====================================================
echo  Building Professional ISX Daily Reports Installers
echo  Using Inno Setup - Windows Standard Installer
echo =====================================================
echo.

REM Check if Inno Setup is installed
set "INNO_SETUP_COMPILER="
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
    set "INNO_SETUP_COMPILER=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
    set "INNO_SETUP_COMPILER=C:\Program Files\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files (x86)\Inno Setup 5\ISCC.exe" (
    set "INNO_SETUP_COMPILER=C:\Program Files (x86)\Inno Setup 5\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 5\ISCC.exe" (
    set "INNO_SETUP_COMPILER=C:\Program Files\Inno Setup 5\ISCC.exe"
)

if "%INNO_SETUP_COMPILER%"=="" (
    echo ERROR: Inno Setup compiler not found!
    echo.
    echo Please install Inno Setup from: https://jrsoftware.org/isinfo.php
    echo.
    echo After installation, run this script again to build professional installers.
    pause
    exit /b 1
)

echo Found Inno Setup: %INNO_SETUP_COMPILER%
echo.

REM Create output directory
if not exist "installer-output" mkdir "installer-output"

echo Building x64 Professional Installer...
echo ======================================
"%INNO_SETUP_COMPILER%" "installer\isx-professional-x64.iss"
if errorlevel 1 (
    echo ERROR: Failed to build x64 installer!
    pause
    exit /b 1
)
echo x64 installer built successfully!
echo.

echo Building ARM64 Professional Installer...
echo =========================================
"%INNO_SETUP_COMPILER%" "installer\isx-professional-arm64.iss"
if errorlevel 1 (
    echo ERROR: Failed to build ARM64 installer!
    pause
    exit /b 1
)
echo ARM64 installer built successfully!
echo.

echo =====================================================
echo  🎉 PROFESSIONAL INSTALLERS BUILT SUCCESSFULLY!
echo =====================================================
echo.
echo Output files:
dir "ISX-Daily-Reports-Professional-*.exe" /b 2>nul
echo.
echo Features of your new professional installers:
echo • Windows standard installer behavior
echo • Professional wizard-style interface  
echo • License agreement and readme pages
echo • Architecture detection (x64 vs ARM64)
echo • Desktop and Start Menu shortcuts with custom icons
echo • Automatic uninstaller creation
echo • Control Panel integration
echo • GitHub download with progress tracking
echo • Upgrade detection and handling
echo • Much smaller file sizes (1-2 MB vs 40+ MB)
echo • Professional Windows UI and progress bars
echo.
echo These installers are now ready for distribution!
echo.
pause 