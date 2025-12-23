#!/bin/bash

# ISX Pulse Alpha Release Preparation Script
# Usage: ./scripts/prepare-alpha-release.sh 5
# Creates v0.1.0-alpha.5

set -e  # Exit on error

# Check if alpha number is provided
if [ -z "$1" ]; then
    echo "❌ Error: Alpha number required"
    echo "Usage: ./scripts/prepare-alpha-release.sh 5"
    exit 1
fi

ALPHA_NUM=$1
VERSION="0.1.0-alpha.$ALPHA_NUM"

echo "🚀 Preparing Alpha Release: v$VERSION"
echo "========================================="

# 1. Build all components
echo ""
echo "📦 Building all components..."
./build.bat -target=all
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

# 2. Build updater
echo ""
echo "🔄 Building updater..."
cd api/cmd/updater
go build -ldflags "-s -w" -o ../../../dist/updater.exe
if [ $? -ne 0 ]; then
    echo "❌ Updater build failed"
    exit 1
fi
cd ../../..

# 3. Create version.txt
echo ""
echo "📝 Creating version.txt..."
echo "$VERSION" > dist/version.txt

# 4. Verify icon exists
echo ""
echo "🎨 Verifying icon..."
if [ ! -f "installer/assets/ISXPulse.ico" ]; then
    echo "⚠️  Icon not found, copying from web/public..."
    mkdir -p installer/assets
    cp web/public/favicon.ico installer/assets/ISXPulse.ico
fi

# 5. Update Inno Setup script version
echo ""
echo "📝 Updating installer version..."
if [ -f "installer/ISXPulse.iss" ]; then
    sed -i "s/AppVersion=.*/AppVersion=$VERSION/" installer/ISXPulse.iss
    # Update numeric version (0.1.0.5)
    NUMERIC_VERSION="0.1.0.$ALPHA_NUM"
    sed -i "s/VersionInfoVersion=.*/VersionInfoVersion=$NUMERIC_VERSION/" installer/ISXPulse.iss
    echo "✅ Updated ISXPulse.iss to version $VERSION"
else
    echo "❌ installer/ISXPulse.iss not found"
    exit 1
fi

# 6. Display dist folder contents
echo ""
echo "📂 Dist folder contents:"
ls -lh dist/*.exe 2>/dev/null || echo "No executables found"
ls -lh dist/version.txt 2>/dev/null || echo "No version.txt found"

# 7. Check if Inno Setup is installed
echo ""
echo "🛠️  Looking for Inno Setup..."
INNO_SETUP_PATH="C:/Program Files (x86)/Inno Setup 6/ISCC.exe"

if [ -f "$INNO_SETUP_PATH" ]; then
    # 8. Compile installer
    echo ""
    echo "🔨 Compiling installer..."
    "$INNO_SETUP_PATH" installer/ISXPulse.iss

    if [ $? -eq 0 ]; then
        # 9. Rename output with alpha number
        if [ -f "installer/Output/ISXPulseSetup-Alpha.exe" ]; then
            mv "installer/Output/ISXPulseSetup-Alpha.exe" "installer/Output/ISXPulseSetup-Alpha-$ALPHA_NUM.exe"

            echo ""
            echo "========================================="
            echo "✅ Alpha Release Preparation Complete!"
            echo "========================================="
            echo ""
            echo "📦 Installer: installer/Output/ISXPulseSetup-Alpha-$ALPHA_NUM.exe"

            # Display file size
            if [ -f "installer/Output/ISXPulseSetup-Alpha-$ALPHA_NUM.exe" ]; then
                SIZE=$(ls -lh "installer/Output/ISXPulseSetup-Alpha-$ALPHA_NUM.exe" | awk '{print $5}')
                echo "📊 Size: $SIZE"
            fi

            echo ""
            echo "📤 Next Steps:"
            echo "1. Test the installer locally"
            echo "2. Create GitHub Release:"
            echo "   git tag v$VERSION"
            echo "   git push origin v$VERSION"
            echo "3. Upload installer to GitHub Releases"
            echo "4. Mark release as 'pre-release' (Alpha)"
            echo ""
        else
            echo "❌ Installer output not found"
            exit 1
        fi
    else
        echo "❌ Installer compilation failed"
        exit 1
    fi
else
    echo ""
    echo "⚠️  Inno Setup not found at: $INNO_SETUP_PATH"
    echo ""
    echo "📦 Dist folder is ready for manual installer build"
    echo ""
    echo "📥 Download Inno Setup from:"
    echo "   https://jrsoftware.org/isdl.php"
    echo ""
    echo "After installing Inno Setup, run this script again."
    echo ""
fi

echo "✨ Done!"
