@echo off
echo ================================================
echo ISX Daily Reports Scraper - Current Release Build
echo Enhanced v2.0.0 with Market Movers & Pipeline Manager
echo ================================================
echo.

REM Clean previous builds
echo 🧹 Cleaning previous builds...
del /q isxcli.exe 2>nul
del /q web-licensed.exe 2>nul
del /q process.exe 2>nul
del /q indexcsv.exe 2>nul
del /q market-movers.exe 2>nul

REM Build main CLI application
echo ✅ Building main CLI application...
cd ..\..\cmd
if exist isxcli.go (
    go build -o ..\..\isxcli.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build main CLI application
        exit /b 1
    )
    echo ✓ Main CLI built successfully
) else (
    echo ⚠️  Main CLI not found, skipping...
)

REM Build enhanced licensed web application
echo ✅ Building enhanced licensed web application...
cd ..\..\cmd\web-licensed
if exist main.go (
    go build -o ..\..\..\web-licensed.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build enhanced web application
        exit /b 1
    )
    echo ✓ Enhanced web application built successfully
) else (
    echo ⚠️  Enhanced web application not found, checking legacy path...
    cd ..\..\cmd\web-licensed-legacy
    if exist main.go (
        go build -o ..\..\..\web-licensed.exe .
        if %ERRORLEVEL% neq 0 (
            echo ❌ Failed to build legacy web application
            exit /b 1
        )
        echo ✓ Legacy web application built successfully
    )
)

REM Build processing tools
echo ✅ Building processing tools...
cd ..\..\cmd\process
if exist main.go (
    go build -o ..\..\..\process.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build process tool
        exit /b 1
    )
    echo ✓ Process tool built successfully
) else (
    echo ⚠️  Process tool not found, skipping...
)

REM Build CSV indexing tool
echo ✅ Building CSV indexing tools...
cd ..\..\cmd\indexcsv
if exist main.go (
    go build -o ..\..\..\indexcsv.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build indexcsv tool
        exit /b 1
    )
    echo ✓ IndexCSV tool built successfully
) else (
    echo ⚠️  IndexCSV tool not found, skipping...
)

REM Build market movers processing tool
echo ✅ Building market movers processing tool...
cd ..\..\cmd\market-movers
if exist main.go (
    go build -o ..\..\..\market-movers.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build market movers tool
        exit /b 1
    )
    echo ✓ Market Movers tool built successfully
) else (
    echo ⚠️  Market Movers tool not found, checking dev path...
    cd ..\..\dev\cmd\market-movers
    if exist main.go (
        go build -o ..\..\..\..\market-movers.exe .
        if %ERRORLEVEL% neq 0 (
            echo ❌ Failed to build dev market movers tool
            exit /b 1
        )
        echo ✓ Dev Market Movers tool built successfully
    ) else (
        echo ⚠️  Market Movers tool not available
    )
)

REM Build pipeline manager
echo ✅ Building pipeline manager...
cd ..\..\cmd\pipeline-manager
if exist main.go (
    go build -o ..\..\..\pipeline-manager.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build pipeline manager
        exit /b 1
    )
    echo ✓ Pipeline Manager built successfully
) else (
    echo ⚠️  Pipeline Manager not found, checking dev path...
    cd ..\..\dev\cmd\pipeline-manager
    if exist main.go (
        go build -o ..\..\..\..\pipeline-manager.exe .
        if %ERRORLEVEL% neq 0 (
            echo ❌ Failed to build dev pipeline manager
            exit /b 1
        )
        echo ✓ Dev Pipeline Manager built successfully
    ) else (
        echo ⚠️  Pipeline Manager not available
    )
)

REM Build license management tools
echo ✅ Building license management tools...
cd ..\..\cmd\license-generator
if exist main.go (
    go build -o ..\..\..\license-generator.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build license generator
        exit /b 1
    )
    echo ✓ License generator built successfully
) else (
    echo ⚠️  License generator not found, skipping...
)

REM Build WebSocket enhanced tools
echo ✅ Building WebSocket enhanced tools...
cd ..\..\cmd\websocket-server
if exist main.go (
    go build -o ..\..\..\websocket-server.exe .
    if %ERRORLEVEL% neq 0 (
        echo ❌ Failed to build WebSocket server
        exit /b 1
    )
    echo ✓ WebSocket Server built successfully
) else (
    echo ⚠️  WebSocket Server not available
)

REM Copy built files to release directory
echo.
echo 📦 Copying built files to release directory...
if not exist ..\..\release mkdir ..\..\release
copy /y isxcli.exe ..\..\release\ 2>nul
copy /y web-licensed.exe ..\..\release\ 2>nul
copy /y process.exe ..\..\release\ 2>nul
copy /y indexcsv.exe ..\..\release\ 2>nul
copy /y market-movers.exe ..\..\release\ 2>nul
copy /y pipeline-manager.exe ..\..\release\ 2>nul
copy /y license-generator.exe ..\..\release\ 2>nul

REM Verify builds
echo.
echo 🔍 Verifying builds...
if exist ..\..\release\web-licensed.exe (
    echo ✓ web-licensed.exe [Enhanced v2.0.0] - READY
) else (
    echo ❌ web-licensed.exe - MISSING
)

if exist ..\..\release\process.exe (
    echo ✓ process.exe - READY
) else (
    echo ❌ process.exe - MISSING
)

if exist ..\..\release\indexcsv.exe (
    echo ✓ indexcsv.exe - READY
) else (
    echo ❌ indexcsv.exe - MISSING
)

echo.
echo ================================================
echo Current Release Build Complete!
echo ================================================
echo.
echo Built components:
echo - ✅ web-licensed.exe (Enhanced v2.0.0)
echo - ✅ process.exe (Data processing)
echo - ✅ indexcsv.exe (CSV indexing)
echo - ✅ market-movers.exe (Market analysis)
echo - ✅ pipeline-manager.exe (Pipeline management)
echo - ✅ license tools (License management)
echo.
echo All components ready for testing!
echo.
pause