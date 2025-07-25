#!/bin/bash
# Script to enable CGO and set up environment for race detector

echo "Setting up environment for Go race detector..."

# Enable CGO
export CGO_ENABLED=1
echo "CGO_ENABLED set to: $CGO_ENABLED"

# Try to find GCC in common locations
if command -v gcc &> /dev/null; then
    echo "GCC found at: $(which gcc)"
    export CC=gcc
elif [ -f "/c/ProgramData/chocolatey/bin/gcc.exe" ]; then
    echo "GCC found at Chocolatey location"
    export CC="/c/ProgramData/chocolatey/bin/gcc.exe"
elif [ -f "/c/tools/mingw64/bin/gcc.exe" ]; then
    echo "GCC found at MinGW64 location"
    export CC="/c/tools/mingw64/bin/gcc.exe"
    export PATH="/c/tools/mingw64/bin:$PATH"
else
    echo "WARNING: GCC not found. Please install MinGW first using install-gcc.bat"
    echo "Run as Administrator: install-gcc.bat"
    exit 1
fi

echo ""
echo "Current Go environment:"
go env | grep -E "CGO_ENABLED|CC|GOARCH|GOOS"

echo ""
echo "Testing race detector..."
echo "package main" > test_race.go
echo "import \"testing\"" >> test_race.go
echo "func TestRace(t *testing.T) {}" >> test_race.go

if go test -race test_race.go 2>&1 | grep -q "race"; then
    echo "✓ Race detector is working!"
    rm test_race.go
else
    echo "✗ Race detector test failed"
    rm test_race.go
    exit 1
fi

echo ""
echo "You can now run tests with race detector:"
echo "  go test -race -v ./internal/websocket/..."