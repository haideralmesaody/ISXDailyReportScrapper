@echo off
echo Testing Chi Middleware Fix...
echo.

echo Starting server with fixed middleware order...
start /B test-websocket-fixed.exe

echo Waiting for server to start...
timeout /t 3 /nobreak > nul

echo.
echo Server should be running at http://localhost:8080
echo Check for:
echo   1. No panic errors
echo   2. WebSocket connects successfully
echo   3. License page loads
echo.
echo Opening test page...
start test-websocket.html

echo.
echo Press Ctrl+C to stop the server.
pause