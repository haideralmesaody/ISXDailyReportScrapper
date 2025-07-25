# Go Race Detector Setup for Windows

## Quick Start

1. **Run as Administrator**: `install-gcc.bat`
2. **Restart your terminal**
3. **Enable in current session**: 
   ```bash
   cd dev
   source enable-race-detector.sh
   ```

## What is the Race Detector?

The Go race detector is a tool that helps find data races in concurrent programs. Per CLAUDE.md requirements, all tests must be run with the race detector enabled.

## Installation Methods

### Method 1: Chocolatey (Recommended)

1. **Run PowerShell as Administrator**
2. **Execute**:
   ```powershell
   .\install-gcc.ps1
   ```
   OR right-click `install-gcc.bat` and select "Run as administrator"

3. **Restart your terminal**

### Method 2: Manual MSYS2 Installation

If Chocolatey method fails:

1. **Download MSYS2** from https://www.msys2.org/
2. **Install and open MSYS2 terminal**
3. **Update packages**:
   ```bash
   pacman -Syu
   ```
4. **Install GCC**:
   ```bash
   pacman -S mingw-w64-x86_64-gcc
   pacman -S mingw-w64-x86_64-toolchain
   ```

### Method 3: Visual Studio Build Tools

1. Download from: https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022
2. Install "Desktop development with C++"
3. Set environment:
   ```bash
   export CC="cl.exe"
   ```

## Verification

After installation:

```bash
# Check GCC
gcc --version

# Check CGO is enabled
go env CGO_ENABLED

# Test race detector
cd dev
go test -race -v ./internal/websocket/... -run TestHubCreation
```

## Troubleshooting

### "gcc: command not found"
- Ensure you restarted your terminal after installation
- Check PATH includes MinGW: `echo $PATH | grep -i mingw`

### "cgo: C compiler not found"
- Run `source enable-race-detector.sh` in dev directory
- Manually set: `export CGO_ENABLED=1 && export CC=gcc`

### Still not working?
- Use WSL2 for development (race detector works out of the box)
- Ensure CI/CD runs tests with race detector on Linux

## Environment Variables

Add to your `.bashrc` or `.bash_profile`:

```bash
export CGO_ENABLED=1
export CC=gcc
export PATH="/c/tools/mingw64/bin:$PATH"  # Adjust path as needed
```

## Running Tests with Race Detector

Once configured:

```bash
# Run all tests with race detector
cd dev
go test -race ./...

# Run specific package
go test -race -v ./internal/websocket/...

# Run with coverage
go test -race -cover ./...
```

## CI/CD Configuration

For GitHub Actions, race detector works by default:

```yaml
- name: Test with race detector
  run: |
    cd dev
    go test -race -v ./...
```

---

Remember: Even if race detector doesn't work locally, ensure your code is race-free by:
- Using proper synchronization (mutexes, channels)
- Following Go concurrency patterns
- Testing thoroughly in CI/CD with race detector enabled