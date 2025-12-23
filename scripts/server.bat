@echo off
setlocal enabledelayedexpansion

echo ==========================================
echo    ISX Pulse Server Management Script
echo ==========================================
echo.

:: Check for existing ISXPulse instances
echo [1/4] Checking for existing ISXPulse instances...
tasklist | findstr /i "ISXPulse.exe" >nul
if %errorlevel% == 0 (
    echo     WARNING: Found existing ISXPulse processes
    echo     Stopping existing instances...
    taskkill /F /IM ISXPulse.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
    echo     Existing instances stopped
) else (
    echo     No existing instances found
)

:: Check for PID file
echo [2/4] Checking for PID file...
if exist "dist\isxpulse.pid" (
    echo     Found existing PID file, removing...
    del "dist\isxpulse.pid" >nul 2>&1
    echo     PID file removed
) else (
    echo     No existing PID file found
)

:: Verify port 8080 is available
echo [3/4] Checking port availability...
netstat -an | findstr ":8080" >nul
if %errorlevel% == 0 (
    echo     WARNING: Port 8080 appears to be in use
    echo     If another ISXPulse instance is running, please stop it manually
    choice /c YN /m "Continue anyway? (Y/N) "
    if !errorlevel! == 2 (
        echo     Operation cancelled by user
        pause
        exit /b 1
    )
) else (
    echo     Port 8080 is available
)

:: Start the server
echo [4/4] Starting ISXPulse server...
echo     Starting single ISXPulse instance...
cd dist
start "ISXPulse Server" /D "%CD%" ISXPulse.exe

echo.
echo ==========================================
echo ISXPulse server started successfully!
echo ==========================================
echo.
echo Server URL: http://localhost:8080
echo Press any key to open in browser...
pause >nul
start http://localhost:8080
echo.
echo Server management complete.