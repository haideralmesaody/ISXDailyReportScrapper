@echo off
echo ================================================
echo ISX Daily Reports Scraper - Clean Release Build
echo ================================================
echo.

echo 🧹 Cleaning up old files...
if exist release rmdir /s /q release
if exist release-package rmdir /s /q release-package
if exist *.zip del /q *.zip
if exist web.exe del /q web.exe
if exist web-licensed.exe del /q web-licensed.exe
if exist isx-scraper.exe del /q isx-scraper.exe
if exist license-generator.exe del /q license-generator.exe
if exist bulk-license-generator.exe del /q bulk-license-generator.exe
if exist process.exe del /q process.exe
if exist indexcsv.exe del /q indexcsv.exe
if exist sampleformats.exe del /q sampleformats.exe
if exist identifyformats.exe del /q identifyformats.exe
if exist debugindices.exe del /q debugindices.exe
if exist isx-scraper-licensed.exe del /q isx-scraper-licensed.exe

echo ✅ Old files cleaned up

echo.
echo 🔨 Building all components with latest fixes...

echo   • Building main CLI application...
go build -o isxcli.exe .
if %ERRORLEVEL% neq 0 (
    echo ❌ Failed to build main CLI application
    exit /b 1
)

echo   • Building licensed web application...
cd cmd\web-licensed
go build -o web-licensed.exe .
if %ERRORLEVEL% neq 0 (
    echo ❌ Failed to build licensed web application
    exit /b 1
)
cd ..\..

echo   • Building processing tools...
cd cmd\process
go build -o process.exe .
cd ..\indexcsv
go build -o indexcsv.exe .
cd ..\identifyformats
go build -o identifyformats.exe .
cd ..\sampleformats
go build -o sampleformats.exe .
cd ..\debugindices
go build -o debugindices.exe .
cd ..\..

echo   • Building license management tools...
cd cmd\license-generator
go build -o license-generator.exe .
cd ..\bulk-license-generator
go build -o bulk-license-generator.exe .
cd ..\..

echo ✅ All components built successfully

echo.
echo 📦 Creating clean release package...
mkdir release

echo   • Copying main executables...
copy isxcli.exe release\
copy cmd\web-licensed\web-licensed.exe release\web.exe

echo   • Copying processing tools...
mkdir release\tools
copy cmd\process\process.exe release\tools\
copy cmd\indexcsv\indexcsv.exe release\tools\
copy cmd\identifyformats\identifyformats.exe release\tools\
copy cmd\sampleformats\sampleformats.exe release\tools\
copy cmd\debugindices\debugindices.exe release\tools\
copy cmd\license-generator\license-generator.exe release\tools\
copy cmd\bulk-license-generator\bulk-license-generator.exe release\tools\

echo   • Copying web interface...
xcopy web release\web /s /i /q

echo   • Copying documentation...
mkdir release\docs
copy README.md release\docs\
copy WEB_README.md release\docs\
copy WEB_INTERFACE_GUIDE.md release\docs\
copy BUILTIN_LICENSING_SETUP.md release\docs\
copy EXPIRE_STATUS_SETUP.md release\docs\
copy GOOGLE_SHEETS_SETUP.md release\docs\
copy RECHARGE_CARD_SETUP.md release\docs\

echo   • Copying configuration files...
copy go.mod release\
copy go.sum release\

echo   • Creating quick start batch files...
echo @echo off > release\start-web-interface.bat
echo echo Starting ISX Daily Reports Scraper Web Interface... >> release\start-web-interface.bat
echo web.exe >> release\start-web-interface.bat

echo @echo off > release\run-cli.bat
echo echo ISX Daily Reports Scraper CLI >> release\run-cli.bat
echo isxcli.exe %%* >> release\run-cli.bat

echo ✅ Release package created successfully

echo.
echo 📊 Release package contents:
dir release /s /b

echo.
echo ================================================
echo Clean Release Build Complete!
echo ================================================
echo.
echo 📁 Release folder: .\release\
echo 🌐 Main web app: .\release\web.exe (licensed version with auto-browser opening)
echo 💻 CLI tool: .\release\isxcli.exe  
echo 🔧 Additional tools: .\release\tools\
echo 📖 Documentation: .\release\docs\
echo.
echo ✅ All files are freshly compiled and ready for distribution
echo.
pause 