@echo off
echo Starting WebSocket Test...
echo.
echo 1. Starting the web server...
start /B test-websocket.exe > server.log 2>&1

echo 2. Waiting for server to start...
timeout /t 3 /nobreak > nul

echo 3. Opening test page in browser...
start test-websocket.html

echo.
echo Server is running. Check the browser for WebSocket test results.
echo Press Ctrl+C to stop the server.
echo.

:: Keep the window open
pause