@echo off
echo Building ISX License Generator...
echo ================================

REM Build for Windows
echo Building for Windows...
go build -o license-generator.exe main.go

if exist license-generator.exe (
    echo ✅ Successfully built license-generator.exe
    echo.
    echo 📋 Usage examples:
    echo   license-generator.exe -total=100
    echo   license-generator.exe -3m=50 -6m=30 -1y=10
    echo   license-generator.exe -total=50 -output=licenses.txt
    echo.
    echo 🔧 Don't forget to update license-config.json with your API key!
) else (
    echo ❌ Build failed
    exit /b 1
)

pause 