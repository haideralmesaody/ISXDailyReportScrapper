@echo off
REM ============================================
REM ISX Pulse - Executable Location Verifier
REM ============================================
REM 
REM This script verifies that all executables are in the correct location (dist/)
REM and that no executables exist in the root or api directories.
REM 
REM Exit codes:
REM   0 - All checks passed
REM   1 - Executables found in wrong location
REM   2 - Missing executables in dist/
REM
REM Usage:
REM   verify-exe-location.bat [--verbose]
REM
REM Can be integrated into CI/CD pipelines for automated checking.

setlocal enabledelayedexpansion

REM Color codes
set "RED=[31m"
set "GREEN=[32m"
set "YELLOW=[33m"
set "BLUE=[34m"
set "RESET=[0m"

REM Parse arguments
set "VERBOSE=0"
if "%1"=="--verbose" set "VERBOSE=1"

echo %BLUE%============================================%RESET%
echo %BLUE%    ISX Pulse - Executable Verifier    %RESET%
echo %BLUE%============================================%RESET%
echo.

REM Initialize counters
set "ERRORS=0"
set "WARNINGS=0"

REM Check for executables in root directory
echo %BLUE%[CHECK]%RESET% Verifying root directory...
set "ROOT_EXES=0"
for %%F in (*.exe) do (
    if exist "%%F" (
        echo %RED%[ERROR]%RESET% Found executable in root: %%F
        set /a ROOT_EXES+=1
        set /a ERRORS+=1
    )
)
if !ROOT_EXES! equ 0 (
    echo %GREEN%[PASS]%RESET% No executables in root directory
)

REM Check for executables in api directory
echo.
echo %BLUE%[CHECK]%RESET% Verifying api directory...
set "API_EXES=0"
if exist "api\" (
    for %%F in (api\*.exe) do (
        if exist "%%F" (
            echo %RED%[ERROR]%RESET% Found executable in api: %%F
            set /a API_EXES+=1
            set /a ERRORS+=1
        )
    )
    REM Check subdirectories in api
    for /r "api" %%F in (*.exe) do (
        echo %RED%[ERROR]%RESET% Found executable: %%F
        set /a API_EXES+=1
        set /a ERRORS+=1
    )
)
if !API_EXES! equ 0 (
    echo %GREEN%[PASS]%RESET% No executables in api directory
)

REM Check for executables in web directory
echo.
echo %BLUE%[CHECK]%RESET% Verifying web directory...
set "WEB_EXES=0"
if exist "web\" (
    for %%F in (web\*.exe) do (
        if exist "%%F" (
            echo %RED%[ERROR]%RESET% Found executable in web: %%F
            set /a WEB_EXES+=1
            set /a ERRORS+=1
        )
    )
)
if !WEB_EXES! equ 0 (
    echo %GREEN%[PASS]%RESET% No executables in web directory
)

REM Check required executables exist in dist
echo.
echo %BLUE%[CHECK]%RESET% Verifying dist directory...
set "MISSING=0"
set "REQUIRED_EXES=ISXPulse.exe scraper.exe processor.exe indexcsv.exe activate-license.exe"

for %%E in (%REQUIRED_EXES%) do (
    if exist "dist\%%E" (
        if "%VERBOSE%"=="1" (
            for %%F in ("dist\%%E") do (
                set SIZE=%%~zF
                set /a SIZE_MB=!SIZE! / 1048576
                echo %GREEN%[OK]%RESET%    dist\%%E ^(!SIZE_MB! MB^)
            )
        )
    ) else (
        echo %YELLOW%[MISSING]%RESET% dist\%%E not found
        set /a MISSING+=1
        set /a WARNINGS+=1
    )
)

if !MISSING! equ 0 (
    echo %GREEN%[PASS]%RESET% All required executables present in dist/
) else (
    echo %YELLOW%[WARNING]%RESET% !MISSING! executable^(s^) missing from dist/
)

REM Check for any unexpected executables in dist
echo.
echo %BLUE%[CHECK]%RESET% Checking for unexpected executables...
set "UNEXPECTED=0"
for %%F in (dist\*.exe) do (
    set "FOUND=0"
    for %%E in (%REQUIRED_EXES%) do (
        if "%%~nxF"=="%%E" set "FOUND=1"
    )
    if "!FOUND!"=="0" (
        echo %YELLOW%[WARNING]%RESET% Unexpected executable: %%F
        set /a UNEXPECTED+=1
        set /a WARNINGS+=1
    )
)
if !UNEXPECTED! equ 0 (
    echo %GREEN%[PASS]%RESET% No unexpected executables in dist/
)

REM Check license.dat location
echo.
echo %BLUE%[CHECK]%RESET% Verifying license file location...
if exist "dist\license.dat" (
    echo %GREEN%[PASS]%RESET% license.dat found in dist/
) else (
    if exist "license.dat" (
        echo %YELLOW%[WARNING]%RESET% license.dat found in root instead of dist/
        set /a WARNINGS+=1
    ) else (
        echo %BLUE%[INFO]%RESET% license.dat not present ^(OK if not activated^)
    )
)

REM Final report
echo.
echo %BLUE%============================================%RESET%
echo %BLUE%              VERIFICATION SUMMARY          %RESET%
echo %BLUE%============================================%RESET%

if !ERRORS! gtr 0 (
    echo %RED%[FAILED]%RESET% !ERRORS! error^(s^) found
    echo.
    echo %RED%Action Required:%RESET%
    echo   1. Remove all .exe files from root and api directories
    echo   2. Use ./build.bat to build properly to dist/
    echo   3. Never use 'go build' directly
    echo.
    exit /b 1
) else if !WARNINGS! gtr 0 (
    echo %YELLOW%[WARNING]%RESET% !WARNINGS! warning^(s^) found
    echo.
    echo %YELLOW%Recommendations:%RESET%
    if !MISSING! gtr 0 (
        echo   - Run './build.bat -target=all' to build missing executables
    )
    if !UNEXPECTED! gtr 0 (
        echo   - Review unexpected executables in dist/
    )
    echo.
    exit /b 0
) else (
    echo %GREEN%[SUCCESS]%RESET% All checks passed!
    echo.
    echo %GREEN%Status:%RESET%
    echo   - No executables in root directory ✓
    echo   - No executables in api directory ✓
    echo   - No executables in web directory ✓
    echo   - All required executables in dist/ ✓
    echo   - No unexpected executables ✓
    echo.
    exit /b 0
)

endlocal