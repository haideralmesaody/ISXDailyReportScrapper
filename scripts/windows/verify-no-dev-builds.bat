@echo off
:: Build Verification Script - Ensures no builds in dev directory
:: This script checks for forbidden build artifacts and removes them

echo ===============================================
echo    BUILD VERIFICATION - NO DEV BUILDS
echo ===============================================
echo.

set VIOLATIONS=0

:: Check for .next directory in dev/frontend
if exist "dev\frontend\.next" (
    echo [ERROR] Found forbidden .next directory in dev/frontend!
    echo         Removing...
    rmdir /S /Q "dev\frontend\.next" 2>nul
    set /A VIOLATIONS+=1
)

:: Check for out directory in dev/frontend
if exist "dev\frontend\out" (
    echo [ERROR] Found forbidden out directory in dev/frontend!
    echo         Removing...
    rmdir /S /Q "dev\frontend\out" 2>nul
    set /A VIOLATIONS+=1
)

:: Check for .exe files in dev directory
for /R dev %%f in (*.exe) do (
    echo [ERROR] Found forbidden .exe file: %%f
    echo         Removing...
    del "%%f" 2>nul
    set /A VIOLATIONS+=1
)

:: Check for node_modules/.cache in dev/frontend
if exist "dev\frontend\node_modules\.cache" (
    echo [WARNING] Found node_modules cache in dev/frontend
    echo           Clearing...
    rmdir /S /Q "dev\frontend\node_modules\.cache" 2>nul
)

:: Check for tsconfig.tsbuildinfo in dev/frontend
if exist "dev\frontend\tsconfig.tsbuildinfo" (
    echo [WARNING] Found tsconfig build info in dev/frontend
    echo           Removing...
    del "dev\frontend\tsconfig.tsbuildinfo" 2>nul
)

:: Check for web-licensed/frontend directory (old location)
if exist "dev\cmd\web-licensed\frontend" (
    echo [INFO] Found frontend build in web-licensed directory
    echo        This is allowed for embedding but will be cleaned
)

echo.
if %VIOLATIONS% GTR 0 (
    echo [FAIL] Found %VIOLATIONS% build violations in dev directory!
    echo.
    echo REMEMBER: 
    echo   - NEVER run 'npm run build' in dev/frontend
    echo   - NEVER run 'go build' in dev/
    echo   - ALWAYS use ./build.bat from project root
    echo   - ALL builds output to dist/ directory
    echo.
    exit /b 1
) else (
    echo [PASS] No build violations found in dev directory
    echo.
    echo Good job! Following the build rules correctly.
    echo Remember to always use ./build.bat for builds.
    echo.
    exit /b 0
)