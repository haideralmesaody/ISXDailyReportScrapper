@echo off
title ISX Daily Reports Scraper - Web Interface
color 0A

echo.
echo ================================================
echo  ISX Daily Reports Scraper - Web Interface
echo  The Iraqi Investor Group
echo ================================================
echo.

cd /d "%~dp0"

echo Starting web interface...
echo.
echo Web interface will be available at: http://localhost:8080
echo.
echo Press Ctrl+C to stop the server
echo.

REM Try to start the licensed version first, fallback to regular version
if exist "web-licensed.exe" (
    echo Starting licensed web interface...
    web-licensed.exe
) else if exist "web.exe" (
    echo Starting web interface...
    web.exe
) else (
    echo ERROR: Web interface executable not found!
    echo Please reinstall the application.
    pause
    exit /b 1
)

echo.
echo Web interface has stopped.
pause 