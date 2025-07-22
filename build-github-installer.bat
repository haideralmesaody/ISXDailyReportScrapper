@echo off
title ISX Daily Reports Scraper - GitHub Installer Builder
color 0A

echo.
echo ================================================
echo  ISX Daily Reports Scraper - GitHub Installer
echo  The Iraqi Investor Group
echo ================================================
echo.

echo This script builds a SMALL installer that downloads the latest release from GitHub.
echo The installer will be approximately 1-2 MB and will download the application
echo files during installation from your GitHub releases.
echo.

REM Check if Inno Setup is installed
where iscc >nul 2>&1
if errorlevel 1 (
    echo ERROR: Inno Setup Compiler (iscc.exe) not found!
    echo.
    echo Please install Inno Setup from: https://www.jrsoftware.org/isinfo.php
    echo Make sure to add Inno Setup to your PATH or run this script from the Inno Setup directory.
    echo.
    pause
    exit /b 1
)

echo Step 1: Creating installer directories...
mkdir installer\output 2>nul

echo.
echo Step 2: Creating icon placeholder (replace with actual .ico file)...
cd installer\assets
powershell.exe -ExecutionPolicy Bypass -File "create-icon.ps1"
cd ..\..

echo.
echo Step 3: Building installer with Inno Setup...
echo Compiling: installer\isx-github-installer.iss
echo.

iscc "installer\isx-github-installer.iss"

if errorlevel 1 (
    echo.
    echo ERROR: Installer compilation failed!
    echo Please check the Inno Setup script for errors.
    pause
    exit /b 1
)

echo.
echo ================================================
echo  SUCCESS! GitHub Installer Created
echo ================================================
echo.
echo Installer location: installer\output\
echo Installer name: ISXDailyReportsInstaller-v0.1.0.exe
echo.
echo IMPORTANT: Before distributing this installer:
echo 1. Make sure you have published a release on GitHub with the zip file
echo 2. The installer will automatically download from: https://github.com/haideralmesaody/ISXDailyReportScrapper
echo 3. Icon will be generated automatically from favicon.svg
echo.
echo The installer will:
echo - Download the latest release from your GitHub repository
echo - Install the application to Program Files
echo - Create desktop shortcut with icon
echo - Add Start Menu entries
echo.

dir "installer\output\*.exe" /B 2>nul
if errorlevel 1 (
    echo No installer found in output directory!
) else (
    echo.
    echo Installer size:
    for %%f in ("installer\output\*.exe") do echo %%~nxf - %%~zf bytes
)

echo.
echo Ready for distribution!
echo.
pause 