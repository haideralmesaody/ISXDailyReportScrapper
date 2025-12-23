@echo off
:: ISX Pulse - Server Starter
:: The Heartbeat of Iraqi Markets
:: This script starts the ISX Pulse server from the dist directory

if not exist "dist\ISXPulse.exe" (
    echo ERROR: dist\ISXPulse.exe not found!
    echo Please run build.bat first to create the distribution.
    pause
    exit /b 1
)

cd dist
ISXPulse.exe