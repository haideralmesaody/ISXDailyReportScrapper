@echo off
echo Direct MinGW Installation for Go Race Detector
echo ==============================================
echo.

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator!
    echo Right-click and select "Run as administrator"
    pause
    exit /b 1
)

echo Downloading MinGW-w64 directly...
echo.

REM Create tools directory
if not exist "C:\tools" mkdir C:\tools
cd /d C:\tools

echo Downloading MinGW-w64 (this may take a few minutes)...
powershell -Command "Invoke-WebRequest -Uri 'https://github.com/niXman/mingw-builds-binaries/releases/download/13.2.0-rt_v11-rev0/x86_64-13.2.0-release-posix-seh-ucrt-rt_v11-rev0.7z' -OutFile 'mingw64.7z'"

echo.
echo Extracting MinGW-w64...
REM Use PowerShell to extract
powershell -Command "Expand-Archive -Path 'mingw64.7z' -DestinationPath 'C:\tools' -Force" 2>nul
if errorlevel 1 (
    echo PowerShell extraction failed, trying 7-Zip...
    if exist "%ProgramFiles%\7-Zip\7z.exe" (
        "%ProgramFiles%\7-Zip\7z.exe" x mingw64.7z -oC:\tools -y
    ) else (
        echo.
        echo ERROR: Cannot extract archive. Please install 7-Zip from https://www.7-zip.org/
        echo Or manually extract mingw64.7z to C:\tools\
        pause
        exit /b 1
    )
)

echo.
echo Setting up environment variables...

REM Add to system PATH
setx /M PATH "%PATH%;C:\tools\mingw64\bin"

REM Set CGO_ENABLED globally
setx /M CGO_ENABLED "1"

echo.
echo Testing GCC installation...
C:\tools\mingw64\bin\gcc.exe --version

echo.
echo ========================================
echo Installation complete!
echo ========================================
echo.
echo IMPORTANT: Close ALL terminals and open new ones for changes to take effect!
echo.
echo After restarting terminal, verify with:
echo   gcc --version
echo   go env CGO_ENABLED
echo.
echo Then run Go tests with race detector:
echo   cd dev
echo   go test -race -v ./internal/websocket/...
echo.
pause