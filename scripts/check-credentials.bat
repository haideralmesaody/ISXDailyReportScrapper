@echo off
echo ========================================
echo   Checking for Exposed Credentials
echo ========================================
echo.

set found_credentials=0

echo Scanning for credential patterns...
echo.

REM Check for common credential patterns
echo [1/4] Checking for private keys...
findstr /I /C:"BEGIN PRIVATE KEY" /C:"BEGIN RSA PRIVATE KEY" *.go *.json *.js *.html 2>nul >nul
if %errorlevel%==0 (
    echo     WARNING: Found private key patterns!
    set found_credentials=1
) else (
    echo     - No private keys found
)

echo [2/4] Checking for client secrets...
findstr /I /C:"client_secret" /C:"private_key_id" *.go *.json *.js 2>nul >nul
if %errorlevel%==0 (
    echo     WARNING: Found client secret patterns!
    set found_credentials=1
) else (
    echo     - No client secrets found
)

echo [3/4] Checking for specific project IDs...
findstr /C:"isxportfolio" /C:"@isxportfolio" *.go *.json 2>nul >nul
if %errorlevel%==0 (
    echo     WARNING: Found real project identifiers!
    set found_credentials=1
) else (
    echo     - No real project IDs found
)

echo [4/4] Checking staged files...
git diff --cached --name-only | findstr /I "credentials\|secret\|key" >nul 2>nul
if %errorlevel%==0 (
    echo     WARNING: Credential files are staged for commit!
    set found_credentials=1
) else (
    echo     - No credential files staged
)

echo.
echo ========================================
if %found_credentials%==1 (
    echo   RESULT: CREDENTIALS DETECTED!
    echo ========================================
    echo.
    echo STOP! Do not commit!
    echo Run sanitize-credentials.bat before committing.
    exit /b 1
) else (
    echo   RESULT: All Clear
    echo ========================================
    echo.
    echo No credentials detected. Safe to commit.
    exit /b 0
)