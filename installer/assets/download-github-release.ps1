# ISX Daily Reports Scrapper - Professional GitHub Release Downloader
# Enhanced version for use with professional Inno Setup installers

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\ISXDailyReports",
    [string]$AppName = "ISX Daily Reports",
    [string]$Architecture = "",
    [string]$RepoOwner = "haideralmesaody",
    [string]$RepoName = "ISXDailyReportScrapper",
    [switch]$Silent = $false
)

# Function to write output (respects silent mode)
function Write-InstallerOutput {
    param($Message, $Color = "White")
    if (-not $Silent) {
        Write-Host $Message -ForegroundColor $Color
    }
}

# Function to write progress (always visible for installer)
function Write-InstallerProgress {
    param($Message)
    Write-Host $Message -ForegroundColor Yellow
}

Write-InstallerOutput "Downloading $AppName..." "Green"

try {
    # Create installation directory
    Write-InstallerProgress "Preparing installation directory..."
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }

    # GitHub API URL to get latest release
    $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    
    Write-InstallerProgress "Connecting to GitHub..."
    
    # Get latest release info with timeout
    $releaseInfo = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing -TimeoutSec 30
    $version = $releaseInfo.tag_name
    
    # Architecture-specific download logic
    $downloadUrl = $null
    
    if ($Architecture -eq "ARM64") {
        # Try to find ARM64-specific release first
        $downloadUrl = $releaseInfo.assets | Where-Object { 
            $_.name -match "arm64.*\.zip$" -or $_.name -match ".*arm64.*\.zip$" 
        } | Select-Object -First 1 | Select-Object -ExpandProperty browser_download_url
        
        Write-InstallerOutput "Looking for ARM64-specific release..." "Cyan"
    }
    
    # Fallback to general release if no architecture-specific found
    if (-not $downloadUrl) {
        $downloadUrl = $releaseInfo.assets | Where-Object { 
            $_.name -match "\.zip$" 
        } | Select-Object -First 1 | Select-Object -ExpandProperty browser_download_url
        
        if ($Architecture -eq "ARM64") {
            Write-InstallerOutput "No ARM64-specific release found, using general release" "Yellow"
        }
    }
    
    if (-not $downloadUrl) {
        throw "No compatible release package found for your system"
    }
    
    Write-InstallerOutput "Found version: $version" "Green"
    
    # Download with progress
    $fileName = [System.IO.Path]::GetFileName($downloadUrl)
    $zipPath = Join-Path $env:TEMP $fileName
    
    Write-InstallerProgress "Downloading $fileName..."
    
    # Create WebClient for progress tracking
    $webClient = New-Object System.Net.WebClient
    
    # Download with progress (silent for installer compatibility)
    try {
        $progressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing -TimeoutSec 300
        $progressPreference = 'Continue'
    } catch {
        $webClient.Dispose()
        throw "Download failed: $($_.Exception.Message)"
    }
    
    $webClient.Dispose()
    
    # Verify download
    if (-not (Test-Path $zipPath)) {
        throw "Download verification failed - file not found"
    }
    
    $fileSize = (Get-Item $zipPath).Length / 1MB
    Write-InstallerOutput "Download completed ($([math]::Round($fileSize, 1)) MB)" "Green"
    
    # Extract files
    Write-InstallerProgress "Installing application files..."
    
    try {
        # Remove existing files if they exist (upgrade scenario)
        $existingFiles = @("web", "bin", "tools", "start-web-interface.exe", "run-cli.exe")
        foreach ($file in $existingFiles) {
            $fullPath = Join-Path $InstallPath $file
            if (Test-Path $fullPath) {
                Remove-Item $fullPath -Recurse -Force -ErrorAction SilentlyContinue
            }
        }
        
        # Extract new files
        Expand-Archive -Path $zipPath -DestinationPath $InstallPath -Force
        
    } catch {
        throw "File extraction failed: $($_.Exception.Message)"
    }
    
    # Clean up downloaded zip
    Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
    
    # Verify installation
    $exePath = Join-Path $InstallPath "start-web-interface.exe"
    if (-not (Test-Path $exePath)) {
        # Try to find the executable in subdirectories
        $found = Get-ChildItem -Path $InstallPath -Name "start-web-interface.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) {
            $exePath = Join-Path $InstallPath $found
        } else {
            Write-InstallerOutput "Warning: Main executable not found at expected location" "Yellow"
        }
    }
    
    Write-InstallerProgress "Installation completed successfully!"
    Write-InstallerOutput "Installed to: $InstallPath" "Cyan"
    
    # Return success
    exit 0
    
} catch {
    Write-InstallerOutput "Installation failed: $($_.Exception.Message)" "Red"
    
    # Clean up on failure
    if ($zipPath -and (Test-Path $zipPath)) {
        Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
    }
    
    # Return error code
    exit 1
} 