@echo off
echo Installing GCC compiler for Go race detector support...
echo.
echo This script will install MinGW using Chocolatey.
echo Please make sure you run this as Administrator!
echo.
pause

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% == 0 (
    echo Running with Administrator privileges...
) else (
    echo ERROR: This script must be run as Administrator!
    echo Right-click on this file and select "Run as administrator"
    pause
    exit /b 1
)

REM Install MinGW
echo.
echo Installing MinGW...
choco install mingw -y

REM Set CGO_ENABLED
echo.
echo Setting CGO_ENABLED=1...
setx CGO_ENABLED 1

echo.
echo ========================================
echo Installation complete!
echo ========================================
echo.
echo Please close and reopen your Git Bash/terminal.
echo.
echo After restarting, verify installation with:
echo   gcc --version
echo   go env CGO_ENABLED
echo.
echo Then run tests with race detector:
echo   cd dev
echo   go test -race -v ./internal/websocket/...
echo.
pause