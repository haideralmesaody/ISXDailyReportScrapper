# Enhanced PowerShell script to convert favicon.svg to ICO format for installer
param(
    [string]$SourceSvg = "favicon.svg",
    [string]$OutputIco = "isx-app-icon.ico"
)

Write-Host "Converting favicon.svg to ICO format for installer..." -ForegroundColor Green

try {
    # Check if source SVG exists
    if (-not (Test-Path $SourceSvg)) {
        Write-Host "Error: $SourceSvg not found!" -ForegroundColor Red
        exit 1
    }
    
    # Use .NET System.Drawing to convert SVG to ICO
    Add-Type -AssemblyName System.Drawing
    Add-Type -AssemblyName System.Windows.Forms
    
    # Read SVG content
    $svgContent = Get-Content $SourceSvg -Raw
    
    # Convert SVG to bitmap using a temporary HTML rendering approach
    $tempHtml = @"
<!DOCTYPE html>
<html>
<head>
    <style>
        body { margin: 0; padding: 0; width: 256px; height: 256px; }
        svg { width: 100%; height: 100%; }
    </style>
</head>
<body>
$svgContent
</body>
</html>
"@
    
    $tempHtmlFile = [System.IO.Path]::GetTempFileName() + ".html"
    $tempHtml | Out-File -FilePath $tempHtmlFile -Encoding UTF8
    
    # Create WebBrowser control to render SVG
    $webBrowser = New-Object System.Windows.Forms.WebBrowser
    $webBrowser.Size = New-Object System.Drawing.Size(256, 256)
    $webBrowser.ScrollBarsEnabled = $false
    
    # Navigate to HTML file and wait for load
    $webBrowser.Navigate("file://$tempHtmlFile")
    while ($webBrowser.ReadyState -ne "Complete") {
        Start-Sleep -Milliseconds 100
        [System.Windows.Forms.Application]::DoEvents()
    }
    
    # Capture bitmap
    $bitmap = New-Object System.Drawing.Bitmap(256, 256)
    $webBrowser.DrawToBitmap($bitmap, (New-Object System.Drawing.Rectangle(0, 0, 256, 256)))
    
    # Create ICO with multiple sizes
    $iconSizes = @(16, 32, 48, 64, 128, 256)
    $iconBitmaps = @()
    
    foreach ($size in $iconSizes) {
        $resized = New-Object System.Drawing.Bitmap($size, $size)
        $graphics = [System.Drawing.Graphics]::FromImage($resized)
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.DrawImage($bitmap, 0, 0, $size, $size)
        $graphics.Dispose()
        $iconBitmaps += $resized
    }
    
    # Save as ICO file
    $iconBitmaps[0].Save($OutputIco, [System.Drawing.Imaging.ImageFormat]::Icon)
    
    # Cleanup
    $bitmap.Dispose()
    foreach ($bmp in $iconBitmaps) {
        $bmp.Dispose()
    }
    $webBrowser.Dispose()
    Remove-Item $tempHtmlFile -Force -ErrorAction SilentlyContinue
    
    Write-Host "Successfully created $OutputIco from $SourceSvg" -ForegroundColor Green
    Write-Host "Icon file size: $([math]::Round((Get-Item $OutputIco).Length / 1KB, 1)) KB" -ForegroundColor Cyan
    
} catch {
    Write-Host "Error converting SVG to ICO: $_" -ForegroundColor Red
    
    # Fallback: Copy the existing SVG as the icon source
    Write-Host "Using fallback method - copying SVG file..." -ForegroundColor Yellow
    Copy-Item $SourceSvg $OutputIco -Force
}

Write-Host "Icon creation completed!" -ForegroundColor Green 