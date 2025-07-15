@echo off
title ISX Daily Reports Scraper - GitHub Installer Builder
color 0C

echo.
echo ================================================
echo  ISX Daily Reports Scraper - GitHub Installer
echo  The Iraqi Investor Group
echo ================================================
echo.

echo Checking for Inno Setup...
iscc /? >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Inno Setup Compiler (iscc.exe) not found!
    echo.
    echo Please install Inno Setup from: https://www.jrsoftware.org/isinfo.php
    echo Make sure to add Inno Setup to your PATH.
    echo.
    pause
    exit /b 1
)

echo Inno Setup found!
echo.

echo Creating installer directories...
if not exist "installer\output" mkdir "installer\output"

echo.
echo Building GitHub installer...
echo.
cd installer
echo Running: iscc github-installer.iss
iscc github-installer.iss

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Installer build failed!
    echo Please check the Inno Setup compiler output for errors.
    cd ..
    pause
    exit /b 1
)

cd ..

echo.
echo ================================================
echo  GITHUB INSTALLER BUILD COMPLETED!
echo ================================================
echo.
echo The installer has been created in: installer\output\
echo.
dir installer\output\*.exe
echo.
echo INSTALLER FEATURES:
echo - Small size (1-2 MB) - downloads from GitHub
echo - Always gets the latest release automatically
echo - Includes license configuration wizard
echo - Professional Windows integration
echo.
echo NEXT STEPS:
echo 1. Create a GitHub release with your application binaries
echo 2. Distribute this small installer to end users
echo 3. Users will automatically get the latest version
echo.
pause 