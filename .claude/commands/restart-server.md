---
description: Stop all servers, rebuild dist, and restart ISXPulse
---

Execute the complete server restart workflow:

**Step 1: Stop all running ISXPulse server instances**
Run this command:
```
powershell "Stop-Process -Name ISXPulse -Force -ErrorAction SilentlyContinue; Start-Sleep -Seconds 2"
```

**Step 2: Full rebuild to dist folder**
Run this command:
```
./build.bat -target=all
```

**Step 3: Start the server in background**
Run this command:
```
cd dist && ./ISXPulse.exe
```

The server will automatically open your browser to http://localhost:8080 when ready.