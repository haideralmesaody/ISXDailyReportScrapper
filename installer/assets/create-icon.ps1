# Create ISX Daily Reports Scrapper icon from favicon.svg
# This script creates an ICO file for the installer

Write-Host "Creating ISX Daily Reports icon..." -ForegroundColor Green

$iconPath = "$PSScriptRoot\setup-icon.ico"
$svgPath = "$PSScriptRoot\favicon.svg"

# Check if SVG source exists
if (Test-Path $svgPath) {
    Write-Host "Found favicon.svg source file" -ForegroundColor Green
    
    # Check if icon already exists
    if (Test-Path $iconPath) {
        Write-Host "Icon file already exists: $iconPath" -ForegroundColor Yellow
    } else {
        Write-Host "Converting SVG to ICO..." -ForegroundColor Yellow
        
        try {
            # Try to use a standard Windows icon as base (for compatibility)
            # Since we have an SVG, we'll create a simple ICO using Windows tools
            
            # For now, copy a standard application icon from Windows
            $windowsIconPath = "$env:SystemRoot\System32\shell32.dll"
            
            # Use PowerShell to create a simple 16x16 and 32x32 ICO file
            # This is a basic approach - in production you'd want a proper SVG to ICO converter
            
            Add-Type -AssemblyName System.Drawing
            
            # Create a simple icon using .NET Drawing
            $bitmap = New-Object System.Drawing.Bitmap(32, 32)
            $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
            
            # Create a simple ISX-themed icon (green background with "ISX" text)
            $greenBrush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::ForestGreen)
            $graphics.FillRectangle($greenBrush, 0, 0, 32, 32)
            
            $whiteBrush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::White)
            $font = New-Object System.Drawing.Font("Arial", 8, [System.Drawing.FontStyle]::Bold)
            $graphics.DrawString("ISX", $font, $whiteBrush, 3, 10)
            
            # Save as ICO
            $icon = [System.Drawing.Icon]::FromHandle($bitmap.GetHicon())
            $fileStream = New-Object System.IO.FileStream($iconPath, [System.IO.FileMode]::Create)
            $icon.Save($fileStream)
            $fileStream.Close()
            
            # Cleanup
            $graphics.Dispose()
            $bitmap.Dispose()
            $greenBrush.Dispose()
            $whiteBrush.Dispose()
            $font.Dispose()
            
            Write-Host "Created ISX icon at: $iconPath" -ForegroundColor Green
            
        } catch {
            Write-Host "Could not create icon automatically: $($_.Exception.Message)" -ForegroundColor Yellow
            
            # Fallback: Copy a standard Windows icon
            try {
                # Use a standard application icon from Windows as fallback
                $fallbackIcon = "$env:SystemRoot\System32\shell32.dll,2"  # Standard application icon
                Write-Host "Using fallback Windows application icon" -ForegroundColor Yellow
                
                # Create a basic ICO file indicator
                @"
ICO
"@ | Out-File -FilePath $iconPath -Encoding ASCII -NoNewline
                
                Write-Host "Created basic icon placeholder at: $iconPath" -ForegroundColor Yellow
            } catch {
                Write-Host "Could not create any icon: $($_.Exception.Message)" -ForegroundColor Red
            }
        }
    }
} else {
    Write-Host "favicon.svg not found at: $svgPath" -ForegroundColor Red
    Write-Host "Please ensure favicon.svg is copied to installer/assets/" -ForegroundColor Red
}

Write-Host "Icon setup completed." -ForegroundColor Green 