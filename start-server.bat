@echo off
setlocal enabledelayedexpansion
echo ========================================
echo ISX Daily Reports Scrapper - Server Manager
echo ========================================

echo.
echo [1/4] Checking for existing server instances...

REM Check if server is running
set SERVER_RUNNING=0
tasklist /FI "IMAGENAME eq web-licensed.exe" 2>NUL | find /I /N "web-licensed.exe">NUL
if "%ERRORLEVEL%"=="0" (
    set SERVER_RUNNING=1
    echo ⚠ Found running web-licensed.exe process
)

netstat -ano | findstr :8080 | findstr LISTENING >nul 2>&1
if %errorlevel% == 0 (
    set SERVER_RUNNING=1
    echo ⚠ Found process using port 8080
)

if !SERVER_RUNNING! == 0 (
    echo ✓ No existing server instances found
    goto :check_requirements
)

echo.
echo Stopping existing server instances...

REM Kill any existing web-licensed.exe processes
taskkill /IM "web-licensed.exe" /F >nul 2>&1
if %errorlevel% == 0 (
    echo ✓ Stopped web-licensed.exe processes
)

REM Kill any processes using port 8080
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
    taskkill /PID %%a /F >nul 2>&1
    if !errorlevel! == 0 (
        echo ✓ Killed process using port 8080 (PID: %%a)
    )
)

echo.
echo [2/4] Waiting for processes to terminate...
timeout /t 3 >nul

REM Verify port is actually free
netstat -ano | findstr :8080 | findstr LISTENING >nul 2>&1
if %errorlevel% == 0 (
    echo ⚠ Warning: Port 8080 may still be in use, proceeding anyway...
) else (
    echo ✓ Port 8080 is now free
)

:check_requirements

echo.
echo [3/4] Checking server requirements...

REM Check if release directory exists
if not exist "release\" (
    echo ✗ Release directory not found! Run build.bat first.
    pause
    exit /b 1
)

REM Check if web-licensed.exe exists
if not exist "release\web-licensed.exe" (
    echo ✗ web-licensed.exe not found in release directory! Run build.bat first.
    pause
    exit /b 1
)

REM Check if license.dat exists
if not exist "release\license.dat" (
    echo ⚠ Warning: license.dat not found. Server will prompt for license activation.
)

REM Check if web assets exist
if not exist "release\web\index.html" (
    echo ✗ Web assets not found! Run build.bat first.
    pause
    exit /b 1
)

if not exist "release\web\static\js\main.js" (
    echo ✗ Static assets not found! Run build.bat first.
    pause
    exit /b 1
)

echo ✓ All requirements met

echo.
echo [4/4] Starting fresh server instance...
echo.
echo Server will start on http://localhost:8080
echo Press Ctrl+C to stop the server
echo.
echo ========================================

REM Change to release directory and start server
cd release
web-licensed.exe

echo.
echo ========================================
echo Server stopped.
echo ========================================
pause