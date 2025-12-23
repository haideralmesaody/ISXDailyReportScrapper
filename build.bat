@echo off
:: ISX Pulse - Build Wrapper
:: The Heartbeat of Iraqi Markets
:: This is a simple wrapper for the Go build system

:: Check for scratch card environment variables and pass them to build.go
set SCRATCH_CARD_ARGS=
if defined GOOGLE_APPS_SCRIPT_URL (
    set SCRATCH_CARD_ARGS=%SCRATCH_CARD_ARGS% -apps-script-url=%GOOGLE_APPS_SCRIPT_URL%
)
if defined ENABLE_SCRATCH_CARD_MODE (
    if /i "%ENABLE_SCRATCH_CARD_MODE%"=="true" (
        set SCRATCH_CARD_ARGS=%SCRATCH_CARD_ARGS% -scratch-card
    )
)

:: Run the build with all arguments
go run build.go %* %SCRATCH_CARD_ARGS%