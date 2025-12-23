---
description: Stop all running ISXPulse server instances
---

Stop all running ISXPulse server instances gracefully:

**Step 1: Check for running ISXPulse processes**
Run this command to see what ISXPulse processes are currently running:
```
powershell "Get-Process -Name ISXPulse -ErrorAction SilentlyContinue | Select-Object Id, StartTime, ProcessName | Format-Table"
if %ERRORLEVEL% neq 0 (
    echo ℹ️  No ISXPulse processes are currently running
    exit /b 0
)
```

**Step 2: Stop all ISXPulse processes**
Run this command to stop all ISXPulse server instances:
```
powershell "Stop-Process -Name ISXPulse -Force -ErrorAction SilentlyContinue; echo 'Stopped ISXPulse processes'"
```

**Step 3: Wait for graceful shutdown**
Run this command to allow processes to terminate:
```
echo Waiting for processes to terminate...
timeout /t 3 /nobreak >nul
```

**Step 4: Verify all processes are stopped**
Run this command to confirm no ISXPulse processes remain:
```
powershell "Get-Process -Name ISXPulse -ErrorAction SilentlyContinue; if ($?) { echo '❌ Some ISXPulse processes are still running' } else { echo '✅ All ISXPulse processes stopped successfully' }"
```

**Alternative: Force stop if graceful shutdown fails**
If some processes don't stop gracefully, use this more forceful approach:
```
echo Force stopping any remaining ISXPulse processes...
powershell "Get-Process -Name ISXPulse -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue"
taskkill /F /IM ISXPulse.exe 2>nul
timeout /t 2 /nobreak >nul
```

## Port Cleanup
Sometimes ports can remain bound after process termination. To clean up port 8080:
```
echo Checking for processes using port 8080...
powershell "Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | Select-Object LocalPort, State, OwningProcess"
```

If you see processes still using port 8080, you may need to restart your terminal or use:
```
netstat -ano | findstr :8080
```

## Server Status Verification
To confirm the server is fully stopped:
1. Try accessing http://localhost:8080 - should show "connection refused"
2. Check for ISXPulse processes: `powershell "Get-Process -Name ISXPulse"`
3. Verify no windows titled "ISXPulse Server" remain open

## Common Issues
- **"Access denied"**: Some processes may require administrator privileges to stop
- **"Process still running"**: Use the force stop method or restart your terminal
- **"Port still in use"**: Restart your terminal or wait a few minutes for port cleanup

## Related Commands
- `/start-server` - Start the ISXPulse server
- `/restart-server` - Full rebuild and restart
- `./build.bat -target=clean` - Clean build artifacts

## Development Workflow
Typical development workflow:
1. `/stop-server` - Stop current instance
2. Make code changes
3. `./build.bat -target=web` - Quick rebuild (if needed)
4. `/start-server` - Start updated server