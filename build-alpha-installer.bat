@echo off
:: Build ISX Alpha Release Installer
:: Requires Inno Setup 6.0+ to be installed

echo ============================================
echo   Building ISX Alpha Release Installer
echo ============================================
echo.

:: Check if Inno Setup is installed
set INNO_PATH=C:\Program Files (x86)\Inno Setup 6\ISCC.exe
set INNO_PATH_ALT=C:\Program Files\Inno Setup 6\ISCC.exe

if exist "%INNO_PATH%" goto :found_inno
if exist "%INNO_PATH_ALT%" goto :found_inno_alt

echo ERROR: Inno Setup Compiler not found!
echo.
echo Please install Inno Setup 6.0+ from: https://jrsoftware.org/isinfo.php
echo.
echo Expected locations:
echo   - %INNO_PATH%
echo   - %INNO_PATH_ALT%
echo.
pause
exit /b 1

:found_inno
set COMPILER=%INNO_PATH%
goto :compiler_found

:found_inno_alt
set COMPILER=%INNO_PATH_ALT%
goto :compiler_found

:compiler_found

echo Found Inno Setup Compiler: %COMPILER%
echo.

:: Check if release directory exists
if not exist "release\bin\" (
    echo ERROR: Release directory not found!
    echo.
    echo Please ensure the release package is built first:
    echo   - release\bin\isx-web-interface.exe
    echo   - release\bin\isxcli.exe
    echo   - release\web\
    echo   - release\docs\
    echo.
    pause
    exit /b 1
)

:: Check if required files exist
if not exist "release\bin\isx-web-interface.exe" (
    echo ERROR: Main executable not found: release\bin\isx-web-interface.exe
    pause
    exit /b 1
)

if not exist "release\bin\isxcli.exe" (
    echo ERROR: CLI executable not found: release\bin\isxcli.exe
    pause
    exit /b 1
)

echo All required files found.
echo.

:: Create output directory if it doesn't exist
if not exist "release\" mkdir "release"

:: Compile the installer
echo Compiling installer...
echo.

"%COMPILER%" "installer\isx-alpha-installer.iss"

if %ERRORLEVEL% equ 0 (
    echo.
    echo ============================================
    echo    ✅ INSTALLER BUILD SUCCESSFUL!
    echo ============================================
    echo.
    echo Output: release\ISX-Alpha-Installer.exe
    
    if exist "release\ISX-Alpha-Installer.exe" (
        for %%I in ("release\ISX-Alpha-Installer.exe") do (
            echo Size: %%~zI bytes
        )
        echo.
        echo 🎯 Ready for distribution!
        echo.
        echo Users can now:
        echo   1. Download ISX-Alpha-Installer.exe
        echo   2. Run as Administrator
        echo   3. Follow the setup wizard
        echo   4. Use desktop shortcut to start ISX
        echo.
    )
) else (
    echo.
    echo ============================================
    echo    ❌ INSTALLER BUILD FAILED!
    echo ============================================
    echo.
    echo Check the error messages above for details.
    echo Common issues:
    echo   - Missing source files in release directory
    echo   - Invalid file paths in installer script
    echo   - Inno Setup syntax errors
    echo.
)

pause 