@echo off
echo ================================================
echo ISX Daily Reports Scraper - Fixed Release Build
echo ================================================
echo.

echo ✅ Building enhanced licensed web application with auto-browser opening...
cd cmd\web-licensed
go build -o web-licensed.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build enhanced web application
    exit /b 1
)
cd ..\..
echo ✓ Enhanced web application built successfully

echo.
echo ✅ Building main CLI application...
go build -o isxcli.exe .
if %ERRORLEVEL% neq 0 (
    echo Failed to build main CLI application
    exit /b 1
)
echo ✓ Main CLI built successfully

echo.
echo ================================================
echo Fixed Release Build Complete!
echo ================================================
echo.
echo Enhanced Features:
echo - ✅ Automatic browser opening
echo - ✅ Improved error handling with detailed messages
echo - ✅ Network connectivity test functionality
echo - ✅ Troubleshooting tips for users
echo - ✅ Better license validation feedback
echo.
echo Ready to test and build installer!
echo.
pause 