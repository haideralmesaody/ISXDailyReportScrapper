---
description: Start the ISXPulse server without rebuilding
---

Start the ISXPulse server for development or testing:

**Step 1: Check if server executable exists**
Run this command to verify the built executable:
```
if exist "dist\ISXPulse.exe" (
    echo ✅ ISXPulse.exe found in dist/ directory
) else (
    echo ❌ ISXPulse.exe not found. Please run ./build.bat first
    exit /b 1
)
```

**Step 2: Check if server is already running**
Run this command to check for existing ISXPulse processes:
```
powershell "Get-Process -Name ISXPulse -ErrorAction SilentlyContinue | Select-Object Id, ProcessName | Format-Table"
if %ERRORLEVEL% equ 0 (
    echo ⚠️  ISXPulse server is already running!
    echo Use /stop-server to stop it first, or /restart-server for a full restart.
    exit /b 1
)
```

**Step 3: Start the server**
Run this command to start the ISXPulse server:
```
cd dist && start "ISXPulse Server" cmd /k "ISXPulse.exe"
```

**Step 4: Wait for server to initialize**
The server will take a few moments to start. You should see:
- Server startup logs in the command window
- "Server starting on port 8080" message
- Automatic browser opening to http://localhost:8080

**Alternative: Start in background only**
If you don't want a command window visible, use this instead:
```
cd dist && start /B ISXPulse.exe
timeout /t 3 /nobreak >nul
start http://localhost:8080
```

## Server Status Check
To verify the server is running, you can:
1. Open http://localhost:8080 in your browser
2. Check for the ISXPulse process: `powershell "Get-Process -Name ISXPulse"`
3. Look for the "ISXPulse Server" command window

## Common Issues
- **"Executable not found"**: Run `./build.bat -target=web` first
- **"Port already in use"**: Use `/stop-server` to stop existing instances
- **"Access denied"**: Make sure you have permissions to run executables

## Related Commands
- `/stop-server` - Stop all running ISXPulse instances
- `/restart-server` - Full rebuild and restart
- `./build.bat -target=web` - Build only the web server