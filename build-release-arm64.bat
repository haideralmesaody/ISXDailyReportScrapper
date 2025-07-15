@echo off
echo ================================================
echo ISX Daily Reports Scraper - ARM64 Release Build
echo ================================================
echo.

REM Set environment variables for ARM64 Windows builds
set GOOS=windows
set GOARCH=arm64
set CGO_ENABLED=0

echo Building ARM64 executables for Windows ARM laptops...
echo Architecture: %GOARCH%
echo Operating System: %GOOS%
echo.

echo Building main CLI application...
go build -ldflags="-s -w" -o isxcli-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build main CLI application
    exit /b 1
)
echo ✓ Main CLI built successfully (ARM64)

echo.
echo Building licensed web application...
cd cmd\web-licensed
go build -ldflags="-s -w" -o web-licensed-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build licensed web application
    exit /b 1
)
cd ..\..
echo ✓ Licensed web application built successfully (ARM64)

echo.
echo Building processing tools...
cd cmd\process
go build -ldflags="-s -w" -o process-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build process tool
    exit /b 1
)
cd ..\..
echo ✓ Process tool built successfully (ARM64)

cd cmd\indexcsv
go build -ldflags="-s -w" -o indexcsv-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build indexcsv tool
    exit /b 1
)
cd ..\..
echo ✓ IndexCSV tool built successfully (ARM64)

cd cmd\identifyformats
go build -ldflags="-s -w" -o identifyformats-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build identifyformats tool
    exit /b 1
)
cd ..\..
echo ✓ IdentifyFormats tool built successfully (ARM64)

cd cmd\sampleformats
go build -ldflags="-s -w" -o sampleformats-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build sampleformats tool
    exit /b 1
)
cd ..\..
echo ✓ SampleFormats tool built successfully (ARM64)

cd cmd\debugindices
go build -ldflags="-s -w" -o debugindices-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build debugindices tool
    exit /b 1
)
cd ..\..
echo ✓ DebugIndices tool built successfully (ARM64)

echo.
echo Building license management tools...
cd cmd\license-generator
go build -ldflags="-s -w" -o license-generator-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build license generator
    exit /b 1
)
cd ..\..
echo ✓ License generator built successfully (ARM64)

cd cmd\bulk-license-generator
go build -ldflags="-s -w" -o bulk-license-generator-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build bulk license generator
    exit /b 1
)
cd ..\..
echo ✓ Bulk license generator built successfully (ARM64)

echo.
echo Building web interface...
cd cmd\web
go build -ldflags="-s -w" -o web-arm64.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build web interface
    exit /b 1
)
cd ..\..
echo ✓ Web interface built successfully (ARM64)

echo.
echo Creating ARM64 release directory...
if exist release-arm64 rmdir /s /q release-arm64
mkdir release-arm64
mkdir release-arm64\bin
mkdir release-arm64\tools
mkdir release-arm64\web
mkdir release-arm64\docs

echo.
echo Copying ARM64 executables to release directory...
copy isxcli-arm64.exe release-arm64\bin\isxcli.exe
copy cmd\web-licensed\web-licensed-arm64.exe release-arm64\bin\start-web-interface.exe
copy cmd\web\web-arm64.exe release-arm64\bin\web.exe
copy cmd\process\process-arm64.exe release-arm64\tools\
copy cmd\indexcsv\indexcsv-arm64.exe release-arm64\tools\
copy cmd\identifyformats\identifyformats-arm64.exe release-arm64\tools\
copy cmd\sampleformats\sampleformats-arm64.exe release-arm64\tools\
copy cmd\debugindices\debugindices-arm64.exe release-arm64\tools\
copy cmd\license-generator\license-generator-arm64.exe release-arm64\tools\
copy cmd\bulk-license-generator\bulk-license-generator-arm64.exe release-arm64\tools\

echo.
echo Copying web interface files...
xcopy web release-arm64\web /s /i /y

echo.
echo Copying documentation...
copy README.md release-arm64\docs\
copy *.md release-arm64\docs\ 2>nul

echo.
echo Creating launch scripts...
echo @echo off > release-arm64\start-web-interface.bat
echo echo Starting ISX Daily Reports Scraper Web Interface (ARM64)... >> release-arm64\start-web-interface.bat
echo echo. >> release-arm64\start-web-interface.bat
echo echo ^📱 Opening web browser to http://localhost:8080 >> release-arm64\start-web-interface.bat
echo echo ^🔑 If license activation is needed, please have your license key ready >> release-arm64\start-web-interface.bat
echo echo. >> release-arm64\start-web-interface.bat
echo bin\start-web-interface.exe >> release-arm64\start-web-interface.bat

echo @echo off > release-arm64\run-cli.bat
echo echo ISX Daily Reports Scraper CLI (ARM64) >> release-arm64\run-cli.bat
echo echo. >> release-arm64\run-cli.bat
echo bin\isxcli.exe %%* >> release-arm64\run-cli.bat

echo.
echo Creating VERSION.txt file...
echo v0.1.0-arm64 > release-arm64\VERSION.txt

echo.
echo ================================================
echo ARM64 Release Build Complete!
echo ================================================
echo.
echo 📁 Release folder: .\release-arm64\
echo 🌐 Main web app: .\release-arm64\bin\start-web-interface.exe
echo 💻 CLI tool: .\release-arm64\bin\isxcli.exe  
echo 🔧 Additional tools: .\release-arm64\tools\
echo 🚀 Quick start: .\release-arm64\start-web-interface.bat
echo.
echo Note: These executables are compiled for Windows ARM64 architecture
echo and will run natively on ARM laptops like yours!

REM Reset environment variables
set GOOS=
set GOARCH=
set CGO_ENABLED= 