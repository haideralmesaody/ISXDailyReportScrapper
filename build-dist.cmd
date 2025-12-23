@echo off
setlocal

rem One-command build wrapper for Windows CMD.
rem - Uses portable toolchains in .\tools\ when present.
rem - Falls back to system-installed Node/Go otherwise.
rem - Forces Go to use local toolchain (no auto-downloads).

set "ROOT=%~dp0"
set "NODE_DIR=%ROOT%tools\node"
set "GO_DIR=%ROOT%tools\go"

if exist "%GO_DIR%\bin\go.exe" (
  set "PATH=%GO_DIR%\bin;%PATH%"
)
if exist "%NODE_DIR%\node.exe" (
  set "PATH=%NODE_DIR%;%PATH%"
)

set "GOTOOLCHAIN=local"
set "NEXT_TELEMETRY_DISABLED=1"

call "%ROOT%build.bat" %*
