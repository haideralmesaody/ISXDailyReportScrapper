#!/bin/bash
# Install GCC using MSYS2 package manager

echo "Installing GCC for Go race detector using MSYS2..."

# Check if we're in MSYS2
if [ -z "$MSYSTEM" ]; then
    echo "ERROR: This script must be run in MSYS2 terminal!"
    echo "Please open MSYS2 MinGW 64-bit terminal and run this script."
    exit 1
fi

# Update package database
echo "Updating package database..."
pacman -Sy

# Install MinGW-w64 GCC toolchain
echo "Installing MinGW-w64 GCC toolchain..."
pacman -S --needed --noconfirm mingw-w64-x86_64-gcc mingw-w64-x86_64-toolchain

# Verify installation
echo ""
echo "Verifying installation..."
which gcc
gcc --version

# Set up environment
echo ""
echo "Setting up environment..."
echo 'export CGO_ENABLED=1' >> ~/.bashrc
echo 'export CC=gcc' >> ~/.bashrc

# Source for current session
export CGO_ENABLED=1
export CC=gcc

echo ""
echo "Testing Go race detector..."
cd "$(dirname "$0")/dev"

# Create test file
cat > test_race.go << 'EOF'
package main
import "testing"
func TestRace(t *testing.T) {}
EOF

if go test -race test_race.go; then
    echo "✓ Race detector is working!"
else
    echo "✗ Race detector test failed"
fi

rm -f test_race.go

echo ""
echo "Installation complete!"
echo ""
echo "You can now run tests with race detector:"
echo "  cd dev"
echo "  go test -race -v ./internal/websocket/..."