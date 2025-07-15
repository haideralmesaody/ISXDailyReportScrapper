@echo off
title ISX Daily Reports Scraper - GitHub Installer Builder
color 0C

echo.
echo ================================================
echo  ISX Daily Reports Scraper - GitHub Installer
echo  The Iraqi Investor Group
echo ================================================
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

echo This script builds a SMALL installer that downloads the latest release from GitHub.
echo.
echo The installer will be approximately 1-2 MB and will download the application
echo files during installation from your GitHub releases.
echo.

echo Step 1: Creating installer directories...
mkdir installer\output 2>nul

echo.
echo Step 2: Checking installer assets...
if not exist "installer\assets\LICENSE.txt" (
    echo ERROR: LICENSE.txt not found in installer\assets\
    echo Please make sure all installer assets are created.
    pause
    exit /b 1
)

if not exist "installer\assets\README.txt" (
    echo ERROR: README.txt not found in installer\assets\
    echo Please make sure all installer assets are created.
    pause
    exit /b 1
)

if not exist "installer\assets\download-github-release.ps1" (
    echo ERROR: download-github-release.ps1 not found in installer\assets\
    echo Please make sure all installer assets are created.
    pause
    exit /b 1
)

echo.
echo Step 3: Creating installer icon...
if not exist "installer\assets\setup-icon.ico" (
    echo Creating default installer icon...
    copy "web\static\images\favicon.ico" "installer\assets\setup-icon.ico" >nul 2>&1
    if errorlevel 1 (
        echo WARNING: Could not create installer icon
        echo You may need to create installer\assets\setup-icon.ico manually
        echo.
    )
)

echo.
echo Step 4: Downloading Visual C++ Redistributable (if needed)...
if not exist "installer\assets\vc_redist.x64.exe" (
    echo Downloading Visual C++ Redistributable...
    powershell -Command "try { Invoke-WebRequest -Uri 'https://aka.ms/vs/17/release/vc_redist.x64.exe' -OutFile 'installer\assets\vc_redist.x64.exe' -UseBasicParsing } catch { exit 1 }"
    if errorlevel 1 (
        echo WARNING: Failed to download Visual C++ Redistributable
        echo The installer will still work, but may require manual installation of VC++ Redist
        echo You can download it manually from: https://aka.ms/vs/17/release/vc_redist.x64.exe
        echo.
    ) else (
        echo Visual C++ Redistributable downloaded successfully.
    )
)

echo.
echo Step 5: Building GitHub installer...
echo.
echo Running Inno Setup Compiler...
cd installer
iscc github-installer.iss
cd ..

if errorlevel 1 (
    echo.
    echo ERROR: Installer build failed!
    echo Please check the Inno Setup compiler output for errors.
    pause
    exit /b 1
)

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
echo - Configures Windows Firewall automatically
echo - Creates Start Menu shortcuts
echo - Professional uninstallation support
echo.
echo NEXT STEPS:
echo 1. Commit your source code to GitHub
echo 2. Create a GitHub release with application binaries
echo 3. Distribute this small installer to end users
echo 4. Users will automatically get the latest version
echo.
echo For GitHub release creation, see: GITHUB_COMMIT_GUIDE.md
echo.
pause 