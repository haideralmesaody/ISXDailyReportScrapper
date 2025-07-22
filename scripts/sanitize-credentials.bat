@echo off
echo ========================================
echo   Sanitizing Credentials for Commit
echo ========================================
echo.

REM Create backup directory
if not exist ".credentials-backup" mkdir ".credentials-backup"

REM Backup credential files
echo [1/4] Backing up credential files...
if exist "credentials.json" (
    copy /Y "credentials.json" ".credentials-backup\credentials.json.bak" >nul
    echo     - credentials.json backed up
)
if exist "sheets-config.json" (
    copy /Y "sheets-config.json" ".credentials-backup\sheets-config.json.bak" >nul
    echo     - sheets-config.json backed up
)
if exist "dev\internal\license\manager.go" (
    copy /Y "dev\internal\license\manager.go" ".credentials-backup\manager.go.bak" >nul
    echo     - manager.go backed up
)
if exist "internal\license\license.go" (
    copy /Y "internal\license\license.go" ".credentials-backup\license.go.bak" >nul
    echo     - license.go backed up
)

REM Create sanitized manager.go
echo.
echo [2/4] Sanitizing dev\internal\license\manager.go...
if exist "dev\internal\license\manager.go" (
    powershell -Command "& { $content = Get-Content 'dev\internal\license\manager.go' -Raw; $pattern = '(?s)(serviceAccountJSON\s*:=\s*`)([^`]+)(`)'; $replacement = '${1}{`n  \"type\": \"service_account\",`n  \"project_id\": \"YOUR_PROJECT_ID\",`n  \"private_key_id\": \"YOUR_PRIVATE_KEY_ID\",`n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\nYOUR_PRIVATE_KEY_HERE\\n-----END PRIVATE KEY-----\\n\",`n  \"client_email\": \"your-service@your-project.iam.gserviceaccount.com\",`n  \"client_id\": \"YOUR_CLIENT_ID\",`n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",`n  \"token_uri\": \"https://oauth2.googleapis.com/token\",`n  \"auth_provider_x509_cert_url\": \"https://www.googleapis.com/oauth2/v1/certs\",`n  \"client_x509_cert_url\": \"https://www.googleapis.com/robot/v1/metadata/x509/your-service%%40your-project.iam.gserviceaccount.com\",`n  \"universe_domain\": \"googleapis.com\"`n}${3}'; $content = $content -replace $pattern, $replacement; $content | Set-Content 'dev\internal\license\manager.go' -NoNewline }"
    echo     - Replaced embedded credentials with placeholders
    
    REM Also update to use environment variable
    powershell -Command "& { $content = Get-Content 'dev\internal\license\manager.go' -Raw; $pattern = '(?s)(// Placeholder for embedded credentials[^}]+})'; $replacement = '// Placeholder for embedded credentials - replaced during build`r`n`t// To use this package, create a service account credentials JSON file`r`n`t// and either:`r`n`t// 1. Set ISX_CREDENTIALS environment variable with the JSON content`r`n`t// 2. Place credentials.json in the same directory as the executable`r`n`t// 3. Replace this placeholder during build process`r`n`tserviceAccountJSON := os.Getenv(\"ISX_CREDENTIALS\")`r`n`tif serviceAccountJSON == \"\" {`r`n`t`t// Try to load from file if environment variable not set`r`n`t`tif credData, err := os.ReadFile(\"credentials.json\"); err == nil {`r`n`t`t`tserviceAccountJSON = string(credData)`r`n`t`t} else {`r`n`t`t`t// Use placeholder that will fail validation`r`n`t`t`tserviceAccountJSON = `{\"type\": \"service_account\", \"project_id\": \"PLACEHOLDER\"}`'`r`n`t`t}`r`n`t}'; $content = $content -replace $pattern, $replacement; $content | Set-Content 'dev\internal\license\manager.go' -NoNewline }"
    echo     - Updated to load from environment/file
)

REM Create sanitized license.go
echo.
echo [3/4] Sanitizing internal\license\license.go...
if exist "internal\license\license.go" (
    powershell -Command "& { $content = Get-Content 'internal\license\license.go' -Raw; $pattern = '(?s)(serviceAccountJSON\s*:=\s*`)([^`]+)(`)'; $replacement = 'serviceAccountJSON := os.Getenv(\"ISX_CREDENTIALS\")`r`n`tif serviceAccountJSON == \"\" {`r`n`t`t// Try to load from file if environment variable not set`r`n`t`tif credData, err := os.ReadFile(\"credentials.json\"); err == nil {`r`n`t`t`tserviceAccountJSON = string(credData)`r`n`t`t} else {`r`n`t`t`t// Use placeholder that will fail validation`r`n`t`t`tserviceAccountJSON = `{\"type\": \"service_account\", \"project_id\": \"PLACEHOLDER\"}`'`r`n`t`t}`r`n`t}'; $content = $content -replace $pattern, $replacement; $content | Set-Content 'internal\license\license.go' -NoNewline }"
    echo     - Replaced embedded credentials with environment loader
)

REM Stage sanitized files
echo.
echo [4/4] Staging sanitized files...
git add dev\internal\license\manager.go 2>nul
git add internal\license\license.go 2>nul
echo     - Files staged for commit

echo.
echo ========================================
echo   Sanitization Complete!
echo ========================================
echo.
echo IMPORTANT: Your credential files are backed up in .credentials-backup\
echo You can now safely commit your changes.
echo.
echo After pushing, run restore-credentials.bat to restore your local credentials.
echo.