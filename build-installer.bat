@echo off
title ISX Daily Reports Scraper - Installer Build Script
color 0C

echo.
echo ================================================
echo  ISX Daily Reports Scraper - Installer Builder
echo  The Iraqi Investor Group
echo ================================================
echo.

REM Check if Inno Setup is installed
where iscc >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Inno Setup Compiler (iscc.exe) not found!
    echo.
    echo Please install Inno Setup from: https://www.jrsoftware.org/isinfo.php
    echo Make sure to add Inno Setup to your PATH or run this script from the Inno Setup directory.
    echo.
    pause
    exit /b 1
)

echo Step 1: Building all application components...
echo.

REM Build main applications
echo Building main scraper...
go build -ldflags="-s -w" -o isxcli.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build main scraper
    pause
    exit /b 1
)

echo Building web interface...
go build -ldflags="-s -w" -o web.exe ./cmd/web
if %errorlevel% neq 0 (
    echo ERROR: Failed to build web interface
    pause
    exit /b 1
)

echo Building licensed web interface...
go build -ldflags="-s -w" -o web-licensed.exe ./cmd/web-licensed
if %errorlevel% neq 0 (
    echo ERROR: Failed to build licensed web interface
    pause
    exit /b 1
)

REM Build tools
echo Building tools...
cd cmd

echo - Building process tool...
cd process
go build -ldflags="-s -w" -o process.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build process tool
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building indexcsv tool...
cd indexcsv
go build -ldflags="-s -w" -o indexcsv.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build indexcsv tool
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building license generator...
cd license-generator
go build -ldflags="-s -w" -o license-generator.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build license generator
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building bulk license generator...
cd bulk-license-generator
go build -ldflags="-s -w" -o bulk-license-generator.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build bulk license generator
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building identifyformats tool...
cd identifyformats
go build -ldflags="-s -w" -o identifyformats.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build identifyformats tool
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building sampleformats tool...
cd sampleformats
go build -ldflags="-s -w" -o sampleformats.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build sampleformats tool
    cd ..\..
    pause
    exit /b 1
)
cd ..

echo - Building debugindices tool...
cd debugindices
go build -ldflags="-s -w" -o debugindices.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build debugindices tool
    cd ..\..
    pause
    exit /b 1
)
cd ..

cd ..

echo.
echo Step 2: Building license generator app...
cd license-generator-app
go build -ldflags="-s -w" -o license-generator.exe .
if %errorlevel% neq 0 (
    echo ERROR: Failed to build license generator app
    cd ..
    pause
    exit /b 1
)
cd ..

echo.
echo Step 3: Creating installer directories...
mkdir installer\output 2>nul
mkdir installer\assets 2>nul

echo.
echo Step 4: Checking installer assets...
if not exist "installer\assets\LICENSE.txt" (
    echo ERROR: LICENSE.txt not found in installer\assets\
    echo Please create the required installer assets first.
    pause
    exit /b 1
)

if not exist "installer\assets\README.txt" (
    echo ERROR: README.txt not found in installer\assets\
    echo Please create the required installer assets first.
    pause
    exit /b 1
)

echo.
echo Step 5: Downloading Visual C++ Redistributable...
if not exist "installer\assets\vc_redist.x64.exe" (
    echo Downloading Visual C++ Redistributable...
    powershell -Command "Invoke-WebRequest -Uri 'https://aka.ms/vs/17/release/vc_redist.x64.exe' -OutFile 'installer\assets\vc_redist.x64.exe'"
    if %errorlevel% neq 0 (
        echo WARNING: Failed to download Visual C++ Redistributable
        echo You may need to download it manually from: https://aka.ms/vs/17/release/vc_redist.x64.exe
        echo.
    )
)

echo.
echo Step 6: Creating installer icon...
if not exist "installer\assets\setup-icon.ico" (
    echo Creating default installer icon...
    copy "web\static\images\favicon.ico" "installer\assets\setup-icon.ico" >nul 2>&1
    if %errorlevel% neq 0 (
        echo WARNING: Could not create installer icon
        echo You may need to create installer\assets\setup-icon.ico manually
        echo.
    )
)

echo.
echo Step 7: Building installer...
echo.
echo Running Inno Setup Compiler...
iscc installer\isx-scraper-installer.iss

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Installer build failed!
    echo Please check the Inno Setup compiler output for errors.
    pause
    exit /b 1
)

echo.
echo ================================================
echo  INSTALLER BUILD COMPLETED SUCCESSFULLY!
echo ================================================
echo.
echo The installer has been created in: installer\output\
echo.
dir installer\output\*.exe
echo.
echo You can now distribute the installer to end users.
echo.
echo Installation features:
echo - Installs all application components
echo - Creates Start Menu shortcuts
echo - Configures Windows Firewall
echo - Installs Visual C++ Redistributable if needed
echo - License configuration wizard
echo - Automatic uninstallation support
echo.
pause 