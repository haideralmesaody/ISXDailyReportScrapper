# Project Structure Reference

## 🎯 Quick Navigation

### Core Development
- **Source Code**: `/dev/` - Go source files for development
- **Commands**: `/cmd/` - Entry points for various executables
- **Internal Packages**: `/internal/` - Shared internal packages

### Building & Scripts
- **Build Scripts**: `/scripts/build/` - All build-related batch files
- **Credential Scripts**: `/scripts/` - Credential sanitization tools
- **Utilities**: `/scripts/utils/` - Helper scripts

### Release & Distribution
- **Release**: `/release/` - Production-ready executables and assets
- **Installer**: `/installer/` - Inno Setup scripts and assets

### Documentation
- **Main Docs**: `/docs/` - All project documentation
  - `/docs/developer/` - Developer guides
  - `/docs/setup/` - Setup and configuration guides
  - `/docs/installer/` - Installer documentation
  - `/docs/guides/` - General guides
  - `/docs/alpha-testing/` - Alpha testing materials
  - `/docs/specifications/` - Technical specifications
  - `/docs/user/` - End-user documentation

### Testing
- **Tests**: `/tests/` - Test files organized by type
  - `/tests/unit/` - Unit tests
  - `/tests/integration/` - Integration tests
  - `/tests/e2e/` - End-to-end tests

### Web Interface
- **Web Assets**: `/web/` - Web interface HTML/CSS/JS files

## 📁 Key Files in Root

- **README.md** - Main project documentation
- **go.mod** / **go.sum** - Go module files
- **main.go** - Main entry point
- **.gitignore** - Git ignore rules
- **SECURE_WORKFLOW.md** - Security workflow guide
- **credentials.json** - Local credentials (git-ignored)

## 🛠️ Common Tasks

### Building the Project
```bash
cd scripts/build
build-release.bat
```

### Running Tests
```bash
go test ./...
```

### Starting Development
```bash
scripts\new-feature.bat my-feature
```

### Before Committing
```bash
scripts\sanitize-credentials.bat
```

## 🗑️ Removed Clutter (120MB+ saved!)
- Duplicate `isx-scrapper-v0.4.0/` directory
- Embedded git repositories
- Backup files (*.backup, *.bak)
- Test artifacts (coverage.out, etc.)
- Malformed path directories

## 📋 Clean Structure Benefits
- ✅ Clear separation of concerns
- ✅ Easy to navigate
- ✅ No duplicate files
- ✅ Organized documentation
- ✅ Proper script organization
- ✅ Professional layout