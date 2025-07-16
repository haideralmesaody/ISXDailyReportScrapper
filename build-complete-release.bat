@echo off
echo ========================================
echo  ISX Daily Reports Scraper
echo  Complete Release Package Builder
echo ========================================
echo.

REM Clean up any existing release directory
echo [1/8] Cleaning up existing release directory...
if exist release rmdir /s /q release
echo ✓ Release directory cleaned

REM Create release directory structure
echo.
echo [2/8] Creating release directory structure...
mkdir release
mkdir release\bin
mkdir release\tools
mkdir release\docs
mkdir release\data
mkdir release\logs
echo ✓ Directory structure created

REM Build main executables
echo.
echo [3/8] Building main executables...
echo   Building ISX Web Scraper...
go build -o release\bin\isx-web-scraper.exe .
if errorlevel 1 (
    echo ❌ Failed to build web scraper
    pause
    exit /b 1
)

echo   Building Data Processor...
go build -o release\bin\process.exe .\cmd\process
if errorlevel 1 (
    echo ❌ Failed to build data processor
    pause
    exit /b 1
)

echo   Building Index Extractor...
go build -o release\bin\indexcsv.exe .\cmd\indexcsv
if errorlevel 1 (
    echo ❌ Failed to build index extractor
    pause
    exit /b 1
)

echo   Building Web Interface...
go build -o release\bin\isx-web-interface.exe .\cmd\web-licensed
if errorlevel 1 (
    echo ❌ Failed to build web interface
    pause
    exit /b 1
)
echo ✓ Main executables built successfully

REM Copy executables to release root for easy access
echo.
echo [4/8] Setting up user-friendly launchers...
copy release\bin\process.exe release\process.exe >nul
copy release\bin\indexcsv.exe release\indexcsv.exe >nul
copy release\bin\isx-web-interface.exe release\start-web-interface.exe >nul
echo ✓ User-friendly launchers created

REM Build additional tools
echo.
echo [5/8] Building additional tools...
echo   Building format identification tools...
go build -o release\tools\identifyformats.exe .\cmd\identifyformats
go build -o release\tools\sampleformats.exe .\cmd\sampleformats
go build -o release\tools\debugindices.exe .\cmd\debugindices
echo ✓ Additional tools built

REM Copy web interface files
echo.
echo [6/8] Copying web interface files...
xcopy web release\web /s /i /y >nul
if errorlevel 1 (
    echo ❌ Failed to copy web files
    pause
    exit /b 1
)
echo ✓ Web interface files copied

REM Copy documentation and configuration files
echo.
echo [7/8] Copying documentation and configuration...
if exist docs xcopy docs release\docs /s /i /y >nul
if exist release\docs\*.md copy release\docs\*.md release\ >nul
if exist formats.json copy formats.json release\ >nul
if exist license-config.json copy license-config.json release\ >nul

REM Create downloads directory (always start empty for web scraping)
echo   Creating downloads directory for fresh data...
mkdir release\downloads >nul 2>&1
echo   ✓ Downloads directory created (scraper will download fresh data automatically)

REM Create batch files for easy launching
echo.
echo [8/8] Creating launch scripts...

REM Create start-web-interface.bat
echo @echo off > release\start-web-interface.bat
echo echo Starting ISX Web Interface... >> release\start-web-interface.bat
echo echo. >> release\start-web-interface.bat
echo echo ^📱 Opening web browser to http://localhost:8080 >> release\start-web-interface.bat
echo echo ^🔑 If license activation is needed, please have your license key ready >> release\start-web-interface.bat
echo echo. >> release\start-web-interface.bat
echo start-web-interface.exe >> release\start-web-interface.bat

REM Create run-cli.bat
echo @echo off > release\run-cli.bat
echo echo ISX Daily Reports CLI Tools >> release\run-cli.bat
echo echo. >> release\run-cli.bat
echo echo Available commands: >> release\run-cli.bat
echo echo   process.exe - Process Excel files to CSV >> release\run-cli.bat
echo echo   indexcsv.exe - Extract market indices >> release\run-cli.bat
echo echo   isx-web-scraper.exe - Download from ISX website >> release\run-cli.bat
echo echo. >> release\run-cli.bat
echo echo Example: process.exe -in=downloads -out=reports >> release\run-cli.bat
echo echo. >> release\run-cli.bat
echo pause >> release\run-cli.bat

REM Create version file
echo ISX Daily Reports Scraper > release\VERSION.txt
echo Version: Enhanced v2.0.0 >> release\VERSION.txt
echo Build Date: %date% %time% >> release\VERSION.txt
echo. >> release\VERSION.txt
echo Components: >> release\VERSION.txt
echo - ISX Web Scraper (downloads Excel files from ISX website) >> release\VERSION.txt
echo - Data Processor (converts Excel files to CSV reports) >> release\VERSION.txt
echo - Index Extractor (extracts market indices from Excel files) >> release\VERSION.txt
echo - Web Interface (browser-based interface for all tools) >> release\VERSION.txt
echo. >> release\VERSION.txt
echo Usage: >> release\VERSION.txt
echo - Run start-web-interface.bat for web interface >> release\VERSION.txt
echo - Run individual .exe files for command-line usage >> release\VERSION.txt
echo - See docs/ directory for detailed documentation >> release\VERSION.txt

REM Create README file
echo # ISX Daily Reports Scraper > release\README.md
echo. >> release\README.md
echo ## Quick Start >> release\README.md
echo. >> release\README.md
echo **Web Interface (Recommended):** >> release\README.md
echo 1. Double-click `start-web-interface.bat` >> release\README.md
echo 2. Enter your license key when prompted >> release\README.md
echo 3. Use the web interface to scrape and process data >> release\README.md
echo. >> release\README.md
echo **Command Line:** >> release\README.md
echo - `process.exe -in=downloads -out=reports` - Process Excel files >> release\README.md
echo - `indexcsv.exe -dir=downloads -out=reports/indexes.csv` - Extract indices >> release\README.md
echo - `isx-web-scraper.exe -mode=initial` - Download from ISX website >> release\README.md
echo. >> release\README.md
echo ## License >> release\README.md
echo This software requires a valid license. Contact Iraqi Investor for licensing. >> release\README.md

echo.
echo ========================================
echo  🎉 RELEASE PACKAGE BUILD COMPLETE!
echo ========================================
echo.
echo 📦 Release package location: release\
echo.
echo 📋 Package contents:
echo   ✓ start-web-interface.exe - Main web interface
echo   ✓ process.exe - Data processor  
echo   ✓ indexcsv.exe - Index extractor
echo   ✓ bin\ - All executables with technical names
echo   ✓ tools\ - Additional development tools
echo   ✓ web\ - Web interface files
echo   ✓ docs\ - Documentation
echo   ✓ Launch scripts (.bat files)
echo.
echo 🚀 Ready for distribution!
echo.
pause 