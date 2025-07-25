@echo off
setlocal enabledelayedexpansion

:: Set build timestamp
set BUILD_TIME=%date% %time%
echo ==========================================
echo ISX Daily Reports Scrapper - Build Script
echo Build started at: %BUILD_TIME%
echo ==========================================
echo.

:: Check prerequisites
echo [CHECKING] Go installation...
go version >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Go is not installed or not in PATH
    echo         Please install Go from https://golang.org/
    exit /b 1
)
for /f "tokens=3" %%v in ('go version') do set GO_VERSION=%%v
echo [SUCCESS] Found Go %GO_VERSION%
echo.

:: Preserve existing data
set PRESERVE_LICENSE=0
if exist release\license.dat (
    echo [INFO] Found existing license.dat - preserving it
    copy release\license.dat license_backup.dat >nul 2>&1
    set PRESERVE_LICENSE=1
)

:: Clean release directory
echo [CLEANING] Removing old build artifacts...
if exist release\web (
    rmdir /s /q release\web
    echo [CLEANED] release\web
)
if exist release\data (
    rmdir /s /q release\data
    echo [CLEANED] release\data
)
if exist release\logs (
    rmdir /s /q release\logs
    echo [CLEANED] release\logs
)
echo.

:: Create directory structure
echo [CREATING] Directory structure...
mkdir release\web\static\js\core 2>nul
mkdir release\web\static\js\services 2>nul
mkdir release\web\static\js\components 2>nul
mkdir release\web\static\css 2>nul
mkdir release\web\static\images 2>nul
mkdir release\web\templates\layout 2>nul
mkdir release\data\downloads 2>nul
mkdir release\data\reports 2>nul
mkdir release\logs 2>nul
mkdir release\config 2>nul
echo [SUCCESS] Directory structure created
echo.

:: Restore license if it existed
if %PRESERVE_LICENSE%==1 (
    copy license_backup.dat release\license.dat >nul 2>&1
    del license_backup.dat >nul 2>&1
    echo [RESTORED] license.dat
    echo.
)

:: Update dependencies
echo [UPDATING] Go module dependencies...
cd dev
echo [DEBUG] Running go mod tidy...
go mod tidy
if errorlevel 1 (
    echo [WARNING] Could not update dependencies - continuing anyway
    echo.
) else (
    echo [SUCCESS] Dependencies updated
)
echo.

:: Build executables
echo ======== BUILDING EXECUTABLES ========
set BUILD_FAILED=0

:: Build flags for Windows - simple and effective
set GOFLAGS=-mod=readonly
set CGO_ENABLED=0

:: Build scraper
echo.
echo [BUILDING] scraper.exe...
go build -ldflags "-s -w" -o ..\release\scraper.exe ./cmd/scraper
if errorlevel 1 (
    echo [FAILED] scraper.exe build failed!
    set BUILD_FAILED=1
) else (
    if exist ..\release\scraper.exe (
        for %%A in (..\release\scraper.exe) do set SIZE=%%~zA
        set /a SIZE_KB=!SIZE!/1024
        echo [SUCCESS] scraper.exe built successfully (!SIZE_KB! KB^)
    ) else (
        echo [FAILED] scraper.exe not found after build
        set BUILD_FAILED=1
    )
)

:: Build process
echo.
echo [BUILDING] process.exe...
go build -ldflags "-s -w" -o ..\release\process.exe ./cmd/process
if errorlevel 1 (
    echo [FAILED] process.exe build failed!
    set BUILD_FAILED=1
) else (
    if exist ..\release\process.exe (
        for %%A in (..\release\process.exe) do set SIZE=%%~zA
        set /a SIZE_KB=!SIZE!/1024
        echo [SUCCESS] process.exe built successfully (!SIZE_KB! KB^)
    ) else (
        echo [FAILED] process.exe not found after build
        set BUILD_FAILED=1
    )
)

:: Build indexcsv
echo.
echo [BUILDING] indexcsv.exe...
go build -ldflags "-s -w" -o ..\release\indexcsv.exe ./cmd/indexcsv
if errorlevel 1 (
    echo [FAILED] indexcsv.exe build failed!
    set BUILD_FAILED=1
) else (
    if exist ..\release\indexcsv.exe (
        for %%A in (..\release\indexcsv.exe) do set SIZE=%%~zA
        set /a SIZE_KB=!SIZE!/1024
        echo [SUCCESS] indexcsv.exe built successfully (!SIZE_KB! KB^)
    ) else (
        echo [FAILED] indexcsv.exe not found after build
        set BUILD_FAILED=1
    )
)

:: Build web-licensed (with more detailed error reporting)
echo.
echo [BUILDING] web-licensed.exe...
echo [INFO] This may take longer as it includes the web server components...
go build -ldflags "-s -w" -o ..\release\web-licensed.exe ./cmd/web-licensed 2>build_web.log
if errorlevel 1 (
    echo [FAILED] web-licensed.exe build failed!
    echo.
    echo [ERROR] Build output:
    type build_web.log
    set BUILD_FAILED=1
    del build_web.log 2>nul
) else (
    if exist ..\release\web-licensed.exe (
        for %%A in (..\release\web-licensed.exe) do set SIZE=%%~zA
        set /a SIZE_KB=!SIZE!/1024
        echo [SUCCESS] web-licensed.exe built successfully (!SIZE_KB! KB^)
        del build_web.log 2>nul
    ) else (
        echo [FAILED] web-licensed.exe not found after build
        set BUILD_FAILED=1
    )
)

cd ..

:: Check if any builds failed
if %BUILD_FAILED%==1 (
    echo.
    echo [ERROR] One or more executables failed to build
    echo        Check the error messages above
    exit /b 1
)

:: Copy configuration and documentation
echo.
echo ======== COPYING CONFIGURATION ========

:: Copy configuration files
echo [COPYING] Configuration files...
if exist dev\config.yaml (
    copy dev\config.yaml release\config\config.yaml.example >nul 2>&1
    echo [SUCCESS] config.yaml.example
)

if exist credentials.json.example (
    copy credentials.json.example release\credentials.json.example >nul 2>&1
    echo [SUCCESS] credentials.json.example
)

if exist sheets-config.json.example (
    copy sheets-config.json.example release\sheets-config.json.example >nul 2>&1
    echo [SUCCESS] sheets-config.json.example
)

:: Copy web assets
echo.
echo ======== COPYING WEB ASSETS ========

:: Copy HTML files
echo.
echo [COPYING] HTML files...
set HTML_COUNT=0
for %%f in (dev\web\*.html) do (
    copy "%%f" "release\web\" >nul 2>&1
    if errorlevel 1 (
        echo [FAILED] Could not copy %%~nxf
    ) else (
        echo [SUCCESS] Copied %%~nxf
        set /a HTML_COUNT+=1
    )
)
echo [INFO] Copied %HTML_COUNT% HTML files

:: Copy static directory
echo.
echo [COPYING] Static assets...
xcopy /E /I /Y /Q dev\web\static release\web\static >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Could not copy static directory
    set BUILD_FAILED=1
) else (
    :: Count files in static
    set STATIC_COUNT=0
    for /r release\web\static %%f in (*) do set /a STATIC_COUNT+=1
    echo [SUCCESS] Copied static directory (%STATIC_COUNT% files^)
)

:: Copy templates
echo.
echo [COPYING] Templates...
xcopy /E /I /Y /Q dev\web\templates release\web\templates >nul 2>&1
if errorlevel 1 (
    echo [FAILED] Could not copy templates directory
    set BUILD_FAILED=1
) else (
    :: Count template files
    set TEMPLATE_COUNT=0
    for /r release\web\templates %%f in (*.html) do set /a TEMPLATE_COUNT+=1
    echo [SUCCESS] Copied templates directory (%TEMPLATE_COUNT% files^)
)

:: Copy batch scripts
echo.
echo [COPYING] Batch scripts...
if exist start-server.bat (
    copy start-server.bat release\start-server.bat >nul 2>&1
    echo [SUCCESS] start-server.bat
)

:: Final verification
echo.
echo ======== BUILD VERIFICATION ========
echo.
echo === Executables ===
set VERIFY_FAILED=0

for %%f in (scraper.exe process.exe indexcsv.exe web-licensed.exe) do (
    if exist release\%%f (
        for %%A in (release\%%f) do set SIZE=%%~zA
        set /a SIZE_KB=!SIZE!/1024
        echo [OK] %%f ^(!SIZE_KB! KB^)
    ) else (
        echo [MISSING] %%f
        set VERIFY_FAILED=1
    )
)

echo.
echo === Web Assets ===
set WEB_ASSETS_OK=0

if exist release\web\index.html (
    echo [OK] index.html
    set /a WEB_ASSETS_OK+=1
) else (
    echo [MISSING] index.html
    set VERIFY_FAILED=1
)

if exist release\web\license.html (
    echo [OK] license.html
    set /a WEB_ASSETS_OK+=1
) else (
    echo [MISSING] license.html
    set VERIFY_FAILED=1
)

:: Check static directory structure
set STATIC_COUNT=0
for /r release\web\static %%f in (*) do set /a STATIC_COUNT+=1
if !STATIC_COUNT! GTR 0 (
    echo [OK] Static assets (!STATIC_COUNT! files^)
    set /a WEB_ASSETS_OK+=1
) else (
    echo [MISSING] Static assets
    set VERIFY_FAILED=1
)

:: Check templates directory structure
set TEMPLATE_COUNT=0
for /r release\web\templates %%f in (*.html) do set /a TEMPLATE_COUNT+=1
if !TEMPLATE_COUNT! GTR 0 (
    echo [OK] Templates (!TEMPLATE_COUNT! files^)
    set /a WEB_ASSETS_OK+=1
) else (
    echo [MISSING] Templates
    set VERIFY_FAILED=1
)

:: Check configuration files
echo.
echo === Configuration ===
if exist release\credentials.json.example (
    echo [OK] credentials.json.example
) else (
    echo [MISSING] credentials.json.example
)

if exist release\sheets-config.json.example (
    echo [OK] sheets-config.json.example
) else (
    echo [MISSING] sheets-config.json.example
)

:: Check directories
echo.
echo === Directories ===
if exist release\data\downloads (
    echo [OK] Downloads directory
) else (
    echo [MISSING] Downloads directory
    set VERIFY_FAILED=1
)

if exist release\data\reports (
    echo [OK] Reports directory
) else (
    echo [MISSING] Reports directory
    set VERIFY_FAILED=1
)

if exist release\logs (
    echo [OK] Logs directory
) else (
    echo [MISSING] Logs directory
    set VERIFY_FAILED=1
)

:: Build summary
echo.
echo ==========================================
set BUILD_STATUS=SUCCESS

if %VERIFY_FAILED%==1 (
    set BUILD_STATUS=FAILED
    echo BUILD STATUS: !BUILD_STATUS!
    echo.
    echo Some files are missing from the build.
    echo Please check the errors above.
    exit /b 1
) else if %BUILD_FAILED%==1 (
    set BUILD_STATUS=PARTIAL_SUCCESS
    echo BUILD STATUS: !BUILD_STATUS!
    echo.
    echo Build completed with some errors.
    echo Check the warnings above.
) else (
    echo BUILD STATUS: !BUILD_STATUS!
    echo.
    echo All components built successfully!
    echo.
    echo QUICK START:
    echo   cd release
    echo   start-server.bat
    echo.
    echo Or run directly:
    echo   cd release
    echo   web-licensed.exe
    echo.
    echo WEB INTERFACE:
    echo   http://localhost:8080
    echo.
    echo CONFIGURATION:
    echo   1. Copy credentials.json.example to credentials.json
    echo   2. Configure your Google Sheets API credentials
    echo   3. Set up sheets-config.json for sheet mappings
    echo.
    echo DIRECTORY STRUCTURE:
    echo   release\
    echo     web-licensed.exe    (Main web server)
    echo     scraper.exe         (Data scraper)
    echo     process.exe         (Data processor)
    echo     indexcsv.exe        (Index extractor)
    echo     web\                (Web interface files)
    echo     data\               (Data files)
    echo       downloads\        (Downloaded Excel files)
    echo       reports\          (Generated CSV reports)
    echo     logs\               (Application logs)
    echo     config\             (Configuration files)
    echo.
    echo DEFAULT PATHS:
    echo   - Scraper saves to: data\downloads
    echo   - Processor reads from: data\downloads
    echo   - Processor saves to: data\reports
    echo   - Index extractor saves to: data\reports\indexes.csv
)
echo.
echo Build completed at: %date% %time%
echo ==========================================

endlocal