# ISX Daily Reports Analytics - Alpha Release Installer
# Version: Alpha-1.0.0
# Date: 2025-07-15

param(
    [string]$InstallPath = "$env:ProgramFiles\ISX Daily Reports",
    [switch]$SkipChecks = $false,
    [switch]$Verbose = $false
)

# Set up logging
$LogFile = "$env:TEMP\ISX-Alpha-Install.log"
$ErrorCount = 0

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $LogEntry = "[$Timestamp] [$Level] $Message"
    Write-Host $LogEntry
    Add-Content -Path $LogFile -Value $LogEntry
}

function Test-AdminRights {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-SystemRequirements {
    Write-Log "Checking system requirements..." "INFO"
    
    # Check Windows version
    $osVersion = [System.Environment]::OSVersion.Version
    if ($osVersion.Major -lt 10) {
        Write-Log "ERROR: Windows 10 or newer required. Found: $($osVersion)" "ERROR"
        return $false
    }
    Write-Log "✓ Windows version: $($osVersion)" "INFO"
    
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Log "ERROR: PowerShell 5.0 or newer required. Found: $($PSVersionTable.PSVersion)" "ERROR"
        return $false
    }
    Write-Log "✓ PowerShell version: $($PSVersionTable.PSVersion)" "INFO"
    
    # Check available disk space (minimum 500MB)
    $drive = (Get-Item $InstallPath -ErrorAction SilentlyContinue)?.Root?.Name ?? "C:\"
    $freeSpace = (Get-WmiObject -Class Win32_LogicalDisk -Filter "DeviceID='$($drive.TrimEnd('\'))'").FreeSpace
    $requiredSpace = 500MB
    
    if ($freeSpace -lt $requiredSpace) {
        Write-Log "ERROR: Insufficient disk space. Required: 500MB, Available: $([math]::Round($freeSpace/1MB))MB" "ERROR"
        return $false
    }
    Write-Log "✓ Disk space: $([math]::Round($freeSpace/1MB))MB available" "INFO"
    
    # Check network connectivity
    try {
        $null = Test-NetConnection -ComputerName "isx-iq.net" -Port 80 -InformationLevel Quiet -ErrorAction Stop
        Write-Log "✓ Network connectivity to ISX portal verified" "INFO"
    } catch {
        Write-Log "WARNING: Cannot reach ISX portal (isx-iq.net). Internet connection may be required." "WARN"
    }
    
    # Check for .NET Framework (required for some operations)
    $dotNetVersion = Get-ItemProperty "HKLM:SOFTWARE\Microsoft\NET Framework Setup\NDP\v4\Full\" -Name Release -ErrorAction SilentlyContinue
    if ($dotNetVersion.Release -lt 461808) {
        Write-Log "WARNING: .NET Framework 4.7.2 or newer recommended for optimal performance." "WARN"
    } else {
        Write-Log "✓ .NET Framework version is sufficient" "INFO"
    }
    
    return $true
}

function Install-Dependencies {
    Write-Log "Installing/checking dependencies..." "INFO"
    
    # Check for Chrome/Chromium for web scraping
    $chromeInstalled = $false
    $chromePaths = @(
        "${env:ProgramFiles}\Google\Chrome\Application\chrome.exe",
        "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe",
        "${env:LOCALAPPDATA}\Google\Chrome\Application\chrome.exe"
    )
    
    foreach ($path in $chromePaths) {
        if (Test-Path $path) {
            Write-Log "✓ Google Chrome found at: $path" "INFO"
            $chromeInstalled = $true
            break
        }
    }
    
    if (-not $chromeInstalled) {
        Write-Log "WARNING: Google Chrome not found. Web scraping may require manual browser setup." "WARN"
        Write-Log "Please install Google Chrome from: https://www.google.com/chrome/" "INFO"
    }
    
    # Create Windows Firewall rules for the web interface
    try {
        $existingRule = Get-NetFirewallRule -DisplayName "ISX Web Interface" -ErrorAction SilentlyContinue
        if (-not $existingRule) {
            New-NetFirewallRule -DisplayName "ISX Web Interface" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow -Profile Any
            Write-Log "✓ Windows Firewall rule created for port 8080" "INFO"
        } else {
            Write-Log "✓ Windows Firewall rule already exists" "INFO"
        }
    } catch {
        Write-Log "WARNING: Could not create Windows Firewall rule. Manual configuration may be needed." "WARN"
    }
}

function Install-Application {
    Write-Log "Installing ISX Daily Reports Analytics Alpha..." "INFO"
    
    try {
        # Create installation directory
        if (Test-Path $InstallPath) {
            Write-Log "Installation directory exists. Backing up existing installation..." "INFO"
            $backupPath = "$InstallPath.backup.$(Get-Date -Format 'yyyyMMdd-HHmmss')"
            Move-Item -Path $InstallPath -Destination $backupPath
            Write-Log "Existing installation backed up to: $backupPath" "INFO"
        }
        
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
        Write-Log "✓ Created installation directory: $InstallPath" "INFO"
        
        # Create subdirectories
        $subDirs = @("bin", "tools", "web", "data", "downloads", "reports", "logs")
        foreach ($dir in $subDirs) {
            New-Item -ItemType Directory -Path "$InstallPath\$dir" -Force | Out-Null
        }
        Write-Log "✓ Created application directories" "INFO"
        
        # Copy application files
        Copy-Item -Path "bin\*" -Destination "$InstallPath\bin\" -Recurse -Force
        Copy-Item -Path "tools\*" -Destination "$InstallPath\tools\" -Recurse -Force
        Copy-Item -Path "web\*" -Destination "$InstallPath\web\" -Recurse -Force
        Write-Log "✓ Application files copied successfully" "INFO"
        
        # Create batch files for easy launching
        $webBatch = @"
@echo off
echo Starting ISX Daily Reports Analytics (Alpha)...
echo Web interface will open automatically at http://localhost:8080
echo.
echo To stop the application, close this window or press Ctrl+C
echo.
cd /d "$InstallPath"
bin\isx-web-interface.exe
pause
"@
        $webBatch | Out-File -FilePath "$InstallPath\Start-Web-Interface.bat" -Encoding ASCII
        
        $cliBatch = @"
@echo off
echo ISX Daily Reports Analytics - Command Line Interface (Alpha)
echo.
cd /d "$InstallPath"
bin\isxcli.exe %*
pause
"@
        $cliBatch | Out-File -FilePath "$InstallPath\Start-CLI.bat" -Encoding ASCII
        
        Write-Log "✓ Launcher scripts created" "INFO"
        
        # Create desktop shortcuts
        $shell = New-Object -ComObject WScript.Shell
        
        $webShortcut = $shell.CreateShortcut("$env:PUBLIC\Desktop\ISX Analytics (Alpha).lnk")
        $webShortcut.TargetPath = "$InstallPath\Start-Web-Interface.bat"
        $webShortcut.WorkingDirectory = $InstallPath
        $webShortcut.Description = "ISX Daily Reports Analytics - Web Interface (Alpha)"
        $webShortcut.Save()
        
        Write-Log "✓ Desktop shortcut created" "INFO"
        
        # Add to PATH environment variable
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($currentPath -notlike "*$InstallPath\bin*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallPath\bin", "Machine")
            Write-Log "✓ Added to system PATH" "INFO"
        }
        
        return $true
    }
    catch {
        Write-Log "ERROR: Installation failed: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

function Show-PostInstallInstructions {
    Write-Log "Installation completed successfully!" "INFO"
    Write-Host ""
    Write-Host "🎉 ISX Daily Reports Analytics Alpha Installation Complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📁 Installation Location: $InstallPath" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "🚀 Getting Started:" -ForegroundColor Yellow
    Write-Host "   • Double-click 'ISX Analytics (Alpha)' on your desktop"
    Write-Host "   • Or run: $InstallPath\Start-Web-Interface.bat"
    Write-Host "   • Web interface will open at: http://localhost:8080"
    Write-Host ""
    Write-Host "📋 Alpha Testing Notes:" -ForegroundColor Yellow
    Write-Host "   • This is an Alpha version - expect bugs and changes"
    Write-Host "   • Please report issues and feedback"
    Write-Host "   • License activation required on first run"
    Write-Host "   • Chrome browser recommended for best experience"
    Write-Host ""
    Write-Host "📖 Documentation:" -ForegroundColor Cyan
    Write-Host "   • Installation log: $LogFile"
    Write-Host "   • User guide: $InstallPath\docs\ALPHA-USER-GUIDE.md"
    Write-Host "   • Testing guide: $InstallPath\docs\ALPHA-TESTING-GUIDE.md"
    Write-Host ""
    Write-Host "🔧 Command Line Access:" -ForegroundColor Cyan
    Write-Host "   • Open new Command Prompt/PowerShell and type: isxcli"
    Write-Host "   • Or run: $InstallPath\Start-CLI.bat"
    Write-Host ""
    
    if ($ErrorCount -gt 0) {
        Write-Host "⚠️  Installation completed with $ErrorCount warnings. Check the log file for details." -ForegroundColor Yellow
    }
}

# Main installation process
Write-Host "ISX Daily Reports Analytics - Alpha Release Installer" -ForegroundColor Green
Write-Host "===================================================" -ForegroundColor Green
Write-Host ""

# Check if running as administrator
if (-not (Test-AdminRights)) {
    Write-Host "ERROR: This installer must be run as Administrator." -ForegroundColor Red
    Write-Host "Right-click on PowerShell and select 'Run as Administrator', then try again." -ForegroundColor Yellow
    exit 1
}

Write-Log "Starting Alpha installation process..." "INFO"
Write-Log "Install path: $InstallPath" "INFO"
Write-Log "Log file: $LogFile" "INFO"

# System requirements check
if (-not $SkipChecks) {
    if (-not (Test-SystemRequirements)) {
        Write-Host "ERROR: System requirements not met. Installation aborted." -ForegroundColor Red
        exit 1
    }
}

# Install dependencies
Install-Dependencies

# Install the application
if (Install-Application) {
    Show-PostInstallInstructions
    Write-Log "Alpha installation completed successfully!" "INFO"
    exit 0
} else {
    Write-Host "ERROR: Installation failed. Check the log file: $LogFile" -ForegroundColor Red
    exit 1
} 