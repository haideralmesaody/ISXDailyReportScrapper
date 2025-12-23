# Fix for Frontend Embed Issue

## Problem
The Go embed fails with `all:frontend/*` because Next.js creates empty directories that Go cannot embed.

## Solution Options

### Option 1: Use Explicit Embed Patterns (Recommended)
Replace the wildcard embed with explicit patterns:

```go
//go:embed frontend/index.html
//go:embed frontend/404.html
//go:embed frontend/_next
//go:embed frontend/*.ico
//go:embed frontend/*.png
//go:embed frontend/*.svg
//go:embed frontend/site.webmanifest
var frontendFiles embed.FS
```

### Option 2: Create Required Files Manually
Create the minimal required files if they don't exist:

```bash
# Create required directories
mkdir -p api/cmd/web-licensed/frontend/_next
mkdir -p api/cmd/web-licensed/frontend/_next/static
mkdir -p api/cmd/web-licensed/frontend/_next/static/chunks
mkdir -p api/cmd/web-licensed/frontend/_next/static/webpack

# Create required files
touch api/cmd/web-licensed/frontend/_next/static/chunks/.gitkeep
touch api/cmd/web-licensed/frontend/_next/static/webpack/.gitkeep
echo '{"name": "main", "src": "", "dest": "static/index.js", "dynamic": true}' > api/cmd/web-licensed/frontend/_next/build-manifest.json
```

### Option 3: Build Without Static Generation (Quick Fix)
Disable static optimization in Next.js config:

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  // ... other config
  output: 'export',
  trailingSlash: true,
  // Add these to prevent embed issues
  generateEtags: false,
  generateBuildId: false,
}
```

## Steps to Fix

1. Revert main.go to use explicit embed patterns
2. Run Next.js build to generate required files
3. Create missing directories/files if needed
4. Build Go project

## Testing

After fixing, test with:
```bash
./build.bat -target=web
```

Should successfully create ISXPulse.exe in dist/