# Build Guide for ISX Daily Reports Scrapper

## Overview
This guide documents the various ways to build the ISX Daily Reports Scrapper project across different environments.

## Build Methods

### Method 1: Using build.bat (Recommended for Windows)

#### From PowerShell:
```powershell
.\build.bat
```

#### From Command Prompt (cmd):
```cmd
build.bat
```

#### From Windows Terminal:
```powershell
.\build.bat
# or
cmd /c build.bat
```

### Method 2: Manual Build from Git Bash

When using Git Bash or similar Unix-like environments on Windows, batch files don't execute natively. Use these commands:

```bash
# Navigate to dev directory where go.mod is located
cd dev

# Build each executable
go build -ldflags "-s -w" -o ../release/scraper.exe scraper.go
go build -ldflags "-s -w" -o ../release/process.exe cmd/process/data-processor.go
go build -ldflags "-s -w" -o ../release/indexcsv.exe cmd/indexcsv/index-extractor.go
go build -ldflags "-s -w" -o ../release/web-licensed.exe cmd/web-licensed/web-application.go

# Copy web assets
cp -r web/* ../release/web/
```

### Method 3: Running build.bat from Git Bash

If you need to run the batch file from Git Bash:

```bash
# Use cmd to execute the batch file
cmd //c build.bat

# Alternative syntax
cmd.exe /c build.bat
```

## Build Prerequisites

1. **Go Installation**: Go 1.21 or higher
2. **Directory Structure**: Must be in project root (where build.bat is located)
3. **Permissions**: Write access to create/modify release directory

## Common Issues and Solutions

### Issue 1: "build.bat: command not found" in Git Bash
**Solution**: Git Bash doesn't execute .bat files directly. Use `cmd //c build.bat` or build manually.

### Issue 2: "go.mod file not found"
**Solution**: Build commands must be run from the `dev` directory where go.mod is located.

### Issue 3: "package not found" errors
**Solution**: Ensure you're in the dev directory and go.mod exists:
```bash
cd dev
ls go.mod  # Should show the file
```

## Build Flags Explanation

- `-ldflags "-s -w"`: Strips debug information for smaller executables
  - `-s`: Omit symbol table and debug information
  - `-w`: Omit DWARF symbol table

## Quick Reference

### Environment Detection
```bash
# Check your current shell
echo $SHELL

# In Git Bash, this will show: /usr/bin/bash
# In PowerShell, this command won't work
```

### All-in-One Build Command for Git Bash
```bash
cd dev && \
go build -ldflags "-s -w" -o ../release/scraper.exe scraper.go && \
go build -ldflags "-s -w" -o ../release/process.exe cmd/process/data-processor.go && \
go build -ldflags "-s -w" -o ../release/indexcsv.exe cmd/indexcsv/index-extractor.go && \
go build -ldflags "-s -w" -o ../release/web-licensed.exe cmd/web-licensed/web-application.go && \
cp -r web/* ../release/web/ && \
echo "Build completed successfully!"
```

## Verification

After building, verify the executables exist:

```bash
# List executables
ls -la release/*.exe

# Expected output:
# release/scraper.exe
# release/process.exe
# release/indexcsv.exe
# release/web-licensed.exe
```

## Notes

- The build.bat script handles additional tasks like backing up data, creating directories, and copying documentation
- Manual builds only compile the executables; you may need to manually copy web assets and create directories
- Always build from a clean state when releasing to production