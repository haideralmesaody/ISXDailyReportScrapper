@echo off
echo Building ISX Daily Reports Scraper Web Interface...
echo.

echo Building main scraper...
go build -o isxcli.exe .
if %errorlevel% neq 0 (
    echo Failed to build main scraper
    pause
    exit /b 1
)

echo Building web server...
go build -o web.exe ./cmd/web
if %errorlevel% neq 0 (
    echo Failed to build web server
    pause
    exit /b 1
)

echo Building CLI tools...
cd cmd

echo - Building process tool...
cd process
go build -o process.exe .
if %errorlevel% neq 0 (
    echo Failed to build process tool
    pause
    exit /b 1
)
cd ..

echo - Building indexcsv tool...
cd indexcsv
go build -o indexcsv.exe .
if %errorlevel% neq 0 (
    echo Failed to build indexcsv tool
    pause
    exit /b 1
)
cd ..

echo - Building marketscan tool...
cd marketscan
go build -o marketscan.exe .
if %errorlevel% neq 0 (
    echo Failed to build marketscan tool
    pause
    exit /b 1
)
cd ..

echo - Building combine tool...
cd combine
go build -o combine.exe .
if %errorlevel% neq 0 (
    echo Failed to build combine tool
    pause
    exit /b 1
)
cd ..

echo - Building inspect tool...
cd inspect
go build -o inspect.exe .
if %errorlevel% neq 0 (
    echo Failed to build inspect tool
    pause
    exit /b 1
)
cd ..

echo - Building identifyformats tool...
cd identifyformats
go build -o identifyformats.exe .
if %errorlevel% neq 0 (
    echo Failed to build identifyformats tool
    pause
    exit /b 1
)
cd ..

echo - Building sampleformats tool...
cd sampleformats
go build -o sampleformats.exe .
if %errorlevel% neq 0 (
    echo Failed to build sampleformats tool
    pause
    exit /b 1
)
cd ..

echo - Building debugindices tool...
cd debugindices
go build -o debugindices.exe .
if %errorlevel% neq 0 (
    echo Failed to build debugindices tool
    pause
    exit /b 1
)
cd ..

cd ..

echo.
echo Creating downloads directory...
if not exist "downloads" mkdir downloads

echo.
echo All tools built successfully!
echo.
echo To start the web interface:
echo   1. Run: web.exe
echo   2. Open browser to: http://localhost:8080
echo.
echo Available executables:
echo   - isxcli.exe          (Main scraper)
echo   - web.exe             (Web interface)
echo   - cmd/process/process.exe
echo   - cmd/indexcsv/indexcsv.exe
echo   - cmd/marketscan/marketscan.exe
echo   - cmd/combine/combine.exe
echo   - cmd/inspect/inspect.exe
echo   - cmd/identifyformats/identifyformats.exe
echo   - cmd/sampleformats/sampleformats.exe
echo   - cmd/debugindices/debugindices.exe
echo.
pause 