@echo off
echo ==========================================
echo    ISX Pulse Server Stop Script
echo ==========================================
echo.

echo [1/3] Stopping ISXPulse processes...
tasklist | findstr /i "ISXPulse.exe" >nul
if %errorlevel% == 0 (
    echo     Found ISXPulse processes, stopping them...
    taskkill /F /IM ISXPulse.exe
    timeout /t 2 /nobreak >nul
    echo     ISXPulse processes stopped
) else (
    echo     No ISXPulse processes found running
)

echo [2/3] Removing PID file...
if exist "dist\isxpulse.pid" (
    del "dist\isxpulse.pid" >nul 2>&1
    echo     PID file removed
) else (
    echo     No PID file found
)

echo [3/3] Verifying shutdown...
tasklist | findstr /i "ISXPulse.exe" >nul
if %errorlevel% == 1 (
    echo     SUCCESS: All ISXPulse processes stopped
) else (
    echo     WARNING: Some ISXPulse processes may still be running
)

echo.
echo ==========================================
echo Server shutdown complete!
echo ==========================================