@echo off
if "%1"=="" (
    echo Usage: new-feature.bat feature-name
    echo Example: new-feature.bat add-export-button
    exit /b 1
)

echo ========================================
echo   Creating New Feature Branch
echo ========================================
echo.

REM Ensure we're on main
echo [1/4] Switching to main branch...
git checkout main >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Could not switch to main branch
    exit /b 1
)

REM Pull latest changes
echo [2/4] Pulling latest changes...
git pull origin main
if %errorlevel% neq 0 (
    echo ERROR: Could not pull latest changes
    exit /b 1
)

REM Create feature branch
echo [3/4] Creating feature branch...
git checkout -b feature/%1
if %errorlevel% neq 0 (
    echo ERROR: Could not create feature branch
    exit /b 1
)

REM Verify credentials are working
echo [4/4] Verifying local setup...
if exist "credentials.json" (
    echo     - credentials.json found
) else (
    echo     WARNING: credentials.json not found!
    echo     The application may not work properly.
)

echo.
echo ========================================
echo   Feature Branch Ready!
echo ========================================
echo.
echo You are now on branch: feature/%1
echo.
echo REMEMBER:
echo 1. Develop and test your feature
echo 2. Run sanitize-credentials.bat before committing
echo 3. Push to origin when ready
echo 4. Create a Pull Request on GitHub
echo.
echo Happy coding!
echo.