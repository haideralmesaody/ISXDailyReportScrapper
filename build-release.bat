@echo off
echo ================================================
echo ISX Daily Reports Scraper - Release Build
echo ================================================
echo.

echo Building main CLI application...
go build -o isxcli.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build main CLI application
    exit /b 1
)
echo ✓ Main CLI built successfully

echo.
echo Building licensed web application...
cd cmd\web-licensed
go build -o web-licensed.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build licensed web application
    exit /b 1
)
cd ..\..
echo ✓ Licensed web application built successfully

echo.
echo Building processing tools...
cd cmd\process
go build -o process.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build process tool
    exit /b 1
)
cd ..\..
echo ✓ Process tool built successfully

cd cmd\indexcsv
go build -o indexcsv.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build indexcsv tool
    exit /b 1
)
cd ..\..
echo ✓ IndexCSV tool built successfully

cd cmd\identifyformats
go build -o identifyformats.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build identifyformats tool
    exit /b 1
)
cd ..\..
echo ✓ IdentifyFormats tool built successfully

cd cmd\sampleformats
go build -o sampleformats.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build sampleformats tool
    exit /b 1
)
cd ..\..
echo ✓ SampleFormats tool built successfully

cd cmd\debugindices
go build -o debugindices.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build debugindices tool
    exit /b 1
)
cd ..\..
echo ✓ DebugIndices tool built successfully

echo.
echo Building license management tools...
cd cmd\license-generator
go build -o license-generator.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build license generator
    exit /b 1
)
cd ..\..
echo ✓ License generator built successfully

cd cmd\bulk-license-generator
go build -o bulk-license-generator.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build bulk license generator
    exit /b 1
)
cd ..\..
echo ✓ Bulk license generator built successfully

echo.
echo ================================================
echo All components built successfully!
echo ================================================
echo.
echo Built components:
echo - isxcli.exe (Main CLI application)
echo - cmd\web-licensed\web-licensed.exe (Licensed web interface)
echo - cmd\process\process.exe (Data processing tool)
echo - cmd\indexcsv\indexcsv.exe (CSV indexing tool)
echo - cmd\identifyformats\identifyformats.exe (Format identification)
echo - cmd\sampleformats\sampleformats.exe (Format sampling)
echo - cmd\debugindices\debugindices.exe (Index debugging)
echo - cmd\license-generator\license-generator.exe (License generator)
echo - cmd\bulk-license-generator\bulk-license-generator.exe (Bulk license generator)
echo.
echo Ready to build installer package!
echo.
pause 