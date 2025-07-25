@echo off
echo Searching for process using port 8080...

:: Find the PID of the process using port 8080
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
    set PID=%%a
    goto :found
)

echo No process found listening on port 8080.
goto :end

:found
echo Found process with PID: %PID%

:: Get the process name for confirmation
for /f "tokens=1" %%b in ('tasklist /FI "PID eq %PID%" ^| findstr %PID%') do (
    set PROCESS_NAME=%%b
)

echo Process name: %PROCESS_NAME%
echo Killing process...

:: Kill the process
taskkill //PID %PID% //F

if %ERRORLEVEL% == 0 (
    echo Successfully killed %PROCESS_NAME% (PID: %PID%)
) else (
    echo Failed to kill process. It may require administrator privileges.
)

:end
pause