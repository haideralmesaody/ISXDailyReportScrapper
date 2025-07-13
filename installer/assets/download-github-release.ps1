# ISX Daily Reports Scraper - GitHub Release Downloader
# This script downloads the latest release from GitHub during installation

param(
    [Parameter(Mandatory=$true)]
    [string]$InstallPath
)

# GitHub repository information
$GitHubUser = "haideralmesaody"
$GitHubRepo = "ISXDailyReportScrapper"
$GitHubAPI = "https://api.github.com/repos/$GitHubUser/$GitHubRepo/releases/latest"

Write-Host "ISX Daily Reports Scraper - GitHub Downloader" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""

# Function to download file with progress
function Download-FileWithProgress {
    param(
        [string]$Url,
        [string]$OutputPath
    )
    
    try {
        $webClient = New-Object System.Net.WebClient
        
        # Add progress event handler
        $webClient.add_DownloadProgressChanged({
            param($sender, $e)
            Write-Progress -Activity "Downloading $(Split-Path $OutputPath -Leaf)" -Status "Progress: $($e.ProgressPercentage)%" -PercentComplete $e.ProgressPercentage
        })
        
        # Download the file
        $webClient.DownloadFileTaskAsync($Url, $OutputPath).Wait()
        $webClient.Dispose()
        
        Write-Progress -Activity "Downloading" -Completed
        return $true
    }
    catch {
        Write-Host "Error downloading file: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Function to extract ZIP file
function Extract-ZipFile {
    param(
        [string]$ZipPath,
        [string]$ExtractPath
    )
    
    try {
        Write-Host "Extracting files to: $ExtractPath" -ForegroundColor Yellow
        
        # Use .NET Framework method for compatibility
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::ExtractToDirectory($ZipPath, $ExtractPath)
        
        return $true
    }
    catch {
        Write-Host "Error extracting ZIP file: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

try {
    Write-Host "Checking for latest release..." -ForegroundColor Yellow
    
    # Get latest release information from GitHub API
    $releaseInfo = Invoke-RestMethod -Uri $GitHubAPI -Headers @{
        "User-Agent" = "ISX-Installer/1.0"
    }
    
    $latestVersion = $releaseInfo.tag_name
    $releaseAssets = $releaseInfo.assets
    
    Write-Host "Latest version: $latestVersion" -ForegroundColor Cyan
    Write-Host "Available assets: $($releaseAssets.Count)" -ForegroundColor Cyan
    
    # Look for the application ZIP file
    $appAsset = $releaseAssets | Where-Object { $_.name -like "*windows*" -or $_.name -like "*win64*" -or $_.name -like "*release*" }
    
    if (-not $appAsset) {
        # If no specific Windows asset, look for any ZIP file
        $appAsset = $releaseAssets | Where-Object { $_.name -like "*.zip" } | Select-Object -First 1
    }
    
    if (-not $appAsset) {
        throw "No suitable release asset found. Expected a ZIP file containing the application."
    }
    
    $downloadUrl = $appAsset.browser_download_url
    $fileName = $appAsset.name
    $fileSize = [math]::Round($appAsset.size / 1MB, 2)
    
    Write-Host ""
    Write-Host "Downloading: $fileName" -ForegroundColor Green
    Write-Host "Size: $fileSize MB" -ForegroundColor Green
    Write-Host "URL: $downloadUrl" -ForegroundColor Gray
    Write-Host ""
    
    # Create temporary download path
    $tempPath = Join-Path $env:TEMP "isx-download"
    $zipPath = Join-Path $tempPath $fileName
    
    if (-not (Test-Path $tempPath)) {
        New-Item -ItemType Directory -Path $tempPath -Force | Out-Null
    }
    
    # Download the release file
    if (-not (Download-FileWithProgress -Url $downloadUrl -OutputPath $zipPath)) {
        throw "Failed to download release file"
    }
    
    Write-Host "Download completed successfully!" -ForegroundColor Green
    Write-Host ""
    
    # Extract the ZIP file
    if (-not (Extract-ZipFile -ZipPath $zipPath -ExtractPath $tempPath)) {
        throw "Failed to extract release file"
    }
    
    # Find the extracted files and copy them to installation directory
    $extractedFiles = Get-ChildItem -Path $tempPath -Recurse -File
    $copiedCount = 0
    
    Write-Host "Installing files to: $InstallPath" -ForegroundColor Yellow
    
    foreach ($file in $extractedFiles) {
        if ($file.Name -eq $fileName) {
            continue # Skip the ZIP file itself
        }
        
        $relativePath = $file.FullName.Replace($tempPath, "").TrimStart("\")
        $destPath = Join-Path $InstallPath $relativePath
        $destDir = Split-Path $destPath -Parent
        
        # Create destination directory if it doesn't exist
        if (-not (Test-Path $destDir)) {
            New-Item -ItemType Directory -Path $destDir -Force | Out-Null
        }
        
        # Copy the file
        Copy-Item -Path $file.FullName -Destination $destPath -Force
        $copiedCount++
        
        Write-Host "  + $relativePath" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "Installation completed successfully!" -ForegroundColor Green
    Write-Host "Files installed: $copiedCount" -ForegroundColor Cyan
    
    # Clean up temporary files
    try {
        Remove-Item -Path $tempPath -Recurse -Force
        Write-Host "Temporary files cleaned up." -ForegroundColor Gray
    }
    catch {
        Write-Host "Warning: Could not clean up temporary files at $tempPath" -ForegroundColor Yellow
    }
    
    # Verify installation
    $mainExe = Join-Path $InstallPath "isxcli.exe"
    $webExe = Join-Path $InstallPath "web.exe"
    
    if ((Test-Path $mainExe) -and (Test-Path $webExe)) {
        Write-Host ""
        Write-Host "Installation verification: PASSED" -ForegroundColor Green
        Write-Host "Main executable: $mainExe" -ForegroundColor Green
        Write-Host "Web interface: $webExe" -ForegroundColor Green
    }
    else {
        Write-Host ""
        Write-Host "Installation verification: FAILED" -ForegroundColor Red
        Write-Host "Some required files are missing. Please check the GitHub release." -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "GitHub download and installation completed!" -ForegroundColor Green
    exit 0
}
catch {
    Write-Host ""
    Write-Host "ERROR: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Troubleshooting:" -ForegroundColor Yellow
    Write-Host "1. Check your internet connection" -ForegroundColor Yellow
    Write-Host "2. Verify the GitHub repository exists and has releases" -ForegroundColor Yellow
    Write-Host "3. Try running the installer as Administrator" -ForegroundColor Yellow
    Write-Host "4. Check Windows Firewall settings" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "For support, visit: https://github.com/$GitHubUser/$GitHubRepo/issues" -ForegroundColor Cyan
    
    exit 1
} 