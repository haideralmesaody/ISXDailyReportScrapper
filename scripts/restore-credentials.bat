@echo off
echo ========================================
echo   Restoring Local Credentials
echo ========================================
echo.

REM Check if backup directory exists
if not exist ".credentials-backup" (
    echo ERROR: No backup directory found!
    echo Please ensure you ran sanitize-credentials.bat first.
    exit /b 1
)

REM Restore credential files
echo [1/2] Restoring credential files from backup...
set restored=0

if exist ".credentials-backup\credentials.json.bak" (
    copy /Y ".credentials-backup\credentials.json.bak" "credentials.json" >nul
    echo     - credentials.json restored
    set restored=1
)

if exist ".credentials-backup\sheets-config.json.bak" (
    copy /Y ".credentials-backup\sheets-config.json.bak" "sheets-config.json" >nul
    echo     - sheets-config.json restored
    set restored=1
)

if exist ".credentials-backup\manager.go.bak" (
    copy /Y ".credentials-backup\manager.go.bak" "dev\internal\license\manager.go" >nul
    echo     - manager.go restored
    set restored=1
)

if exist ".credentials-backup\license.go.bak" (
    copy /Y ".credentials-backup\license.go.bak" "internal\license\license.go" >nul
    echo     - license.go restored  
    set restored=1
)

if %restored%==0 (
    echo     - No files to restore
)

REM Clean up the changes from git
echo.
echo [2/2] Cleaning up git status...
git checkout -- dev\internal\license\manager.go 2>nul
git checkout -- internal\license\license.go 2>nul
echo     - Git status cleaned

echo.
echo ========================================
echo   Restoration Complete!
echo ========================================
echo.
echo Your local credentials have been restored.
echo The application will now work normally with your credentials.
echo.