# Convert SVG to ICO for installer
# This script converts the Iraqi Investor logo SVG to ICO format

param(
    [string]$SvgPath = "..\..\web\static\images\iraqi-investor-logo.svg",
    [string]$OutputPath = "setup-icon.ico"
)

Write-Host "Converting SVG to ICO format..." -ForegroundColor Green

# Check if ImageMagick is available
$magickPath = Get-Command "magick" -ErrorAction SilentlyContinue
if (-not $magickPath) {
    Write-Host "ImageMagick not found. Attempting to download and use online converter..." -ForegroundColor Yellow
    
    # Alternative: Use .NET System.Drawing to create a basic ICO
    Add-Type -AssemblyName System.Drawing
    
    # Create a simple bitmap version (basic fallback)
    $bitmap = New-Object System.Drawing.Bitmap(48, 48)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.Clear([System.Drawing.Color]::White)
    
    # Draw a simple representation
    $brush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::FromArgb(45, 90, 61))
    $graphics.FillEllipse($brush, 10, 10, 28, 28)
    
    $font = New-Object System.Drawing.Font("Arial", 8, [System.Drawing.FontStyle]::Bold)
    $graphics.DrawString("TII", $font, $brush, 12, 32)
    
    # Save as ICO
    $icon = [System.Drawing.Icon]::FromHandle($bitmap.GetHicon())
    $fileStream = New-Object System.IO.FileStream($OutputPath, [System.IO.FileMode]::Create)
    $icon.Save($fileStream)
    $fileStream.Close()
    
    $graphics.Dispose()
    $bitmap.Dispose()
    $brush.Dispose()
    $font.Dispose()
    
    Write-Host "Created basic ICO file: $OutputPath" -ForegroundColor Green
} else {
    # Use ImageMagick if available
    try {
        & magick "$SvgPath" -resize 48x48 "$OutputPath"
        Write-Host "Successfully converted SVG to ICO: $OutputPath" -ForegroundColor Green
    } catch {
        Write-Host "Error converting with ImageMagick: $_" -ForegroundColor Red
    }
}

Write-Host "Icon conversion complete!" -ForegroundColor Green 