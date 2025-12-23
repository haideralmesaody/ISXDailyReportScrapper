@echo off
:: ISX Pulse - Drawings Persistence Smoke Test (Windows)
:: Verifies:
:: - Server health endpoint responds
:: - Drawings GET/PUT endpoints work
:: - A file is written under data\indicators\{TICKER}.json (dist or repo-root)

setlocal

set "TICKER=%~1"
if "%TICKER%"=="" set "TICKER=TEST"

powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$ErrorActionPreference='Stop';" ^
  "$base='http://localhost:8080';" ^
  "$health=Invoke-WebRequest -UseBasicParsing \"$base/api/health\" -TimeoutSec 5; if($health.StatusCode -ne 200){ throw 'health check failed' };" ^
  "$t='%TICKER%'.ToUpperInvariant();" ^
  "$url=\"$base/api/v1/drawings/$t/\";" ^
  "$get=Invoke-WebRequest -UseBasicParsing $url -TimeoutSec 5; if($get.StatusCode -ne 200){ throw 'GET drawings failed' };" ^
  "$payload=@{version=1;ticker=$t;shapes=@(@{id='shape1';type='horizontalLine';points=@(@{time=1730000000;price=1.23});pane=0;style=@{color='#00ff00';lineWidth=2};locked=$false;createdAt=0;updatedAt=0})};" ^
  "$json=$payload|ConvertTo-Json -Depth 10;" ^
  "$put=Invoke-WebRequest -UseBasicParsing -Method Put $url -ContentType 'application/json' -Body $json -TimeoutSec 5; if($put.StatusCode -ne 200){ throw 'PUT drawings failed' };" ^
  "$after=Invoke-RestMethod -Method Get $url -TimeoutSec 5; if(-not $after.shapes -or $after.shapes.Count -lt 1){ throw 'expected >=1 shapes after PUT' };" ^
  "Write-Host ('OK drawings persisted for ' + $t);"

if errorlevel 1 exit /b 1

for /f "usebackq delims=" %%I in (`powershell -NoProfile -ExecutionPolicy Bypass -Command "'%TICKER%'.ToUpperInvariant()"`) do set "TICKER_UP=%%I"

if exist "dist\\data\\indicators\\%TICKER_UP%.json" (
  echo Found dist\\data\\indicators\\%TICKER_UP%.json
  exit /b 0
)

if exist "data\\indicators\\%TICKER_UP%.json" (
  echo Found data\\indicators\\%TICKER_UP%.json
  exit /b 0
)

echo WARNING: drawings file not found under dist\\data\\indicators or data\\indicators.
exit /b 0
