# ISX Daily Reports Analytics - Alpha Package Verification
# Run this script to verify the Alpha release package is complete

Write-Host "ISX Daily Reports Analytics - Alpha Package Verification" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
Write-Host ""

$errors = 0
$warnings = 0

function Test-Component {
    param([string]$Path, [string]$Description, [bool]$Critical = $true)
    
    if (Test-Path $Path) {
        Write-Host "✅ $Description" -ForegroundColor Green
        return $true
    } else {
        if ($Critical) {
            Write-Host "❌ $Description - MISSING" -ForegroundColor Red
            $script:errors++
        } else {
            Write-Host "⚠️  $Description - Missing (Optional)" -ForegroundColor Yellow
            $script:warnings++
        }
        return $false
    }
}

Write-Host "Checking Core Executables..." -ForegroundColor Cyan
Test-Component "bin\isxcli.exe" "Main CLI Application"
Test-Component "bin\isx-web-interface.exe" "Web Interface Application"
Test-Component "tools\process.exe" "Data Processing Tool"
Test-Component "tools\indexcsv.exe" "Index Extraction Tool"

Write-Host ""
Write-Host "Checking Web Assets..." -ForegroundColor Cyan
Test-Component "web\index.html" "Main Web Interface"
Test-Component "web\license.html" "License Activation Page"
Test-Component "web\static\images\favicon.svg" "Favicon"
Test-Component "web\static\images\iraqi-investor-logo.svg" "Iraqi Investor Logo"

Write-Host ""
Write-Host "Checking Documentation..." -ForegroundColor Cyan
Test-Component "docs\ALPHA-USER-GUIDE.md" "Alpha User Guide"
Test-Component "docs\ALPHA-TESTING-GUIDE.md" "Alpha Testing Guide"
Test-Component "README-ALPHA.md" "Alpha README"

Write-Host ""
Write-Host "Checking Installation..." -ForegroundColor Cyan
Test-Component "install-alpha.ps1" "Alpha Installer Script"

Write-Host ""
Write-Host "Checking File Sizes..." -ForegroundColor Cyan

$files = @(
    @{Path="bin\isxcli.exe"; MinSize=20MB; Description="CLI Application"},
    @{Path="bin\isx-web-interface.exe"; MinSize=20MB; Description="Web Interface"},
    @{Path="tools\process.exe"; MinSize=5MB; Description="Process Tool"},
    @{Path="tools\indexcsv.exe"; MinSize=5MB; Description="Index Tool"}
)

foreach ($file in $files) {
    if (Test-Path $file.Path) {
        $size = (Get-Item $file.Path).Length
        if ($size -gt $file.MinSize) {
            Write-Host "✅ $($file.Description): $([math]::Round($size/1MB, 1))MB" -ForegroundColor Green
        } else {
            Write-Host "⚠️  $($file.Description): $([math]::Round($size/1MB, 1))MB (smaller than expected)" -ForegroundColor Yellow
            $warnings++
        }
    }
}

Write-Host ""
Write-Host "Testing Executable Dependencies..." -ForegroundColor Cyan

# Test if executables can start (basic dependency check)
try {
    $output = & "bin\isxcli.exe" --help 2>&1
    if ($LASTEXITCODE -eq 0 -or $output -match "ISX") {
        Write-Host "✅ CLI Application - Dependencies OK" -ForegroundColor Green
    } else {
        Write-Host "⚠️  CLI Application - May have dependency issues" -ForegroundColor Yellow
        $warnings++
    }
} catch {
    Write-Host "❌ CLI Application - Cannot execute" -ForegroundColor Red
    $errors++
}

Write-Host ""
Write-Host "Validation Summary:" -ForegroundColor Yellow
Write-Host "==================" -ForegroundColor Yellow

if ($errors -eq 0 -and $warnings -eq 0) {
    Write-Host "🎉 PERFECT! Alpha package is complete and ready for distribution." -ForegroundColor Green
} elseif ($errors -eq 0) {
    Write-Host "✅ GOOD! Alpha package is ready with $warnings minor warnings." -ForegroundColor Green
} else {
    Write-Host "❌ ISSUES! Found $errors critical errors and $warnings warnings." -ForegroundColor Red
    Write-Host "Please fix critical errors before distribution." -ForegroundColor Red
}

Write-Host ""
Write-Host "Package Statistics:" -ForegroundColor Cyan
$totalSize = (Get-ChildItem -Recurse | Measure-Object -Property Length -Sum).Sum
Write-Host "Total Package Size: $([math]::Round($totalSize/1MB, 1))MB"

$fileCount = (Get-ChildItem -Recurse -File | Measure-Object).Count
Write-Host "Total Files: $fileCount"

Write-Host ""
Write-Host "Ready for Alpha Testing Distribution: $(if ($errors -eq 0) {'YES ✅'} else {'NO ❌'})" -ForegroundColor $(if ($errors -eq 0) {'Green'} else {'Red'})

if ($errors -eq 0) {
    Write-Host ""
    Write-Host "📦 Next Steps:" -ForegroundColor Green
    Write-Host "1. Create ZIP package for distribution"
    Write-Host "2. Test installation on clean system"
    Write-Host "3. Send to alpha testers with documentation"
    Write-Host "4. Set up feedback collection process"
} 