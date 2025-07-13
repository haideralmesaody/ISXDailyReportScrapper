@echo off
title ISX Daily Reports Scraper - Command Line Interface
color 0E

echo.
echo ================================================
echo  ISX Daily Reports Scraper - CLI Mode
echo  The Iraqi Investor Group
echo ================================================
echo.

cd /d "%~dp0"

echo Welcome to the ISX Daily Reports Scraper CLI!
echo.
echo Available options:
echo 1. Run scraper with default settings (accumulative mode)
echo 2. Run scraper for specific date range
echo 3. Process existing Excel files
echo 4. Extract market indices
echo 5. Show help
echo 6. Exit
echo.

:menu
set /p choice=Enter your choice (1-6): 

if "%choice%"=="1" goto :default_scrape
if "%choice%"=="2" goto :date_range
if "%choice%"=="3" goto :process_files
if "%choice%"=="4" goto :extract_indices
if "%choice%"=="5" goto :show_help
if "%choice%"=="6" goto :exit

echo Invalid choice. Please enter 1-6.
goto :menu

:default_scrape
echo.
echo Running scraper in accumulative mode...
echo This will download any new reports since the last run.
echo.
isxcli.exe --mode=accumulative
goto :finished

:date_range
echo.
echo Running scraper for specific date range...
echo Date format: YYYY-MM-DD (e.g., 2024-01-15)
echo.
set /p start_date=Enter start date: 
set /p end_date=Enter end date (or press Enter for today): 

if "%end_date%"=="" (
    isxcli.exe --mode=initial --from=%start_date%
) else (
    isxcli.exe --mode=initial --from=%start_date% --to=%end_date%
)
goto :finished

:process_files
echo.
echo Processing existing Excel files...
echo This will convert Excel files to CSV format.
echo.
.\tools\process.exe
goto :finished

:extract_indices
echo.
echo Extracting market indices (ISX60, ISX15)...
echo.
.\tools\indexcsv.exe
goto :finished

:show_help
echo.
echo ISX Daily Reports Scraper - Help
echo ================================
echo.
echo Command line options:
echo   --mode=initial         Download all reports from start date
echo   --mode=accumulative    Download only new reports (default)
echo   --from=YYYY-MM-DD      Start date for scraping
echo   --to=YYYY-MM-DD        End date for scraping
echo   --headless=true        Run browser in headless mode (default)
echo   --headless=false       Show browser window during scraping
echo   --out=directory        Output directory for downloads
echo.
echo Examples:
echo   isxcli.exe --mode=accumulative
echo   isxcli.exe --mode=initial --from=2024-01-01 --to=2024-01-31
echo   isxcli.exe --from=2024-01-15 --headless=false
echo.
echo Processing tools:
echo   .\tools\process.exe     Convert Excel files to CSV
echo   .\tools\indexcsv.exe    Extract market indices
echo.
echo For web interface, run: web.exe or web-licensed.exe
echo Then visit: http://localhost:8080
echo.
pause
goto :menu

:finished
echo.
echo Operation completed.
echo.
choice /c YN /m "Do you want to run another command"
if errorlevel 2 goto :exit
goto :menu

:exit
echo.
echo Thank you for using ISX Daily Reports Scraper!
echo.
pause 