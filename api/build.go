// build.go - ISX Pulse Build System
// Usage: go run build.go [-target=TARGET]
// Targets: all, web, scraper, processor, indexcsv, activate-license, frontend, clean, test, release, package

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	version = "0.1.0-alpha.4"
	module  = "isxcli"
)

// BuildContext holds configuration for the build process
type BuildContext struct {
	Verbose           bool
	EnableScratchCard bool
	AppsScriptURL     string
}

var (
	// Build directories - will be initialized in init()
	rootDir string
	apiDir  string
	webDir  string
	distDir string

	// Executable names (key = source dir name, value = output exe name)
	executables = map[string]string{
		"web-licensed":         "ISXPulse.exe",
		"scraper":              "scraper.exe",
		"processor":            "processor.exe",
		"indexcsv":             "indexcsv.exe",
		"activate-license":     "activate-license.exe",
		"fingerprint-analyzer": "fingerprint-analyzer.exe",
	}

	// Colors for Windows console
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func init() {
	// Detect where we're running from and set up paths accordingly
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current directory: %v", err))
	}

	// Check if we're in a subdirectory
	if filepath.Base(cwd) == "api" || filepath.Base(cwd) == "web" {
		// Running from subdirectory, go up one level
		rootDir = filepath.Dir(cwd)
	} else {
		// Running from root directory
		rootDir = cwd
	}

	// Set up paths for new structure ONLY
	apiDir = filepath.Join(rootDir, "api")
	webDir = filepath.Join(rootDir, "web")
	distDir = filepath.Join(rootDir, "dist")

	// Verify new structure exists
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("API directory not found at %s. Please ensure the project has been migrated to the new structure.", apiDir))
	}
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Web directory not found at %s. Please ensure the project has been migrated to the new structure.", webDir))
	}
}

func main() {
	// Parse command line flags
	target := flag.String("target", "all", "Build target")
	verbose := flag.Bool("v", false, "Verbose output")
	enableScratchCard := flag.Bool("scratch-card", false, "Enable scratch card license system")
	appsScriptURL := flag.String("apps-script-url", "", "Google Apps Script URL for license management")
	flag.Parse()

	// Enable color output on Windows
	if runtime.GOOS == "windows" {
		enableWindowsColors()
	}

	printHeader()

	// Execute the requested target
	startTime := time.Now()

	// Create build context
	buildCtx := &BuildContext{
		Verbose:           *verbose,
		EnableScratchCard: *enableScratchCard,
		AppsScriptURL:     *appsScriptURL,
	}

	switch *target {
	case "all":
		buildAll(buildCtx)
	case "web":
		buildWeb(buildCtx)
	case "web-licensed":
		buildWeb(buildCtx)
	case "scraper":
		buildExecutableWithContext("scraper", buildCtx)
	case "processor":
		buildExecutableWithContext("processor", buildCtx)
	case "indexcsv":
		buildExecutableWithContext("indexcsv", buildCtx)
	case "activate-license":
		buildExecutableWithContext("activate-license", buildCtx)
	case "fingerprint-analyzer":
		buildExecutableWithContext("fingerprint-analyzer", buildCtx)
	case "frontend":
		buildFrontend(buildCtx.Verbose)
	case "clean":
		clean(buildCtx.Verbose)
	case "test":
		runTests(buildCtx.Verbose)
	case "release":
		buildRelease(buildCtx)
	case "package":
		createPackage(buildCtx.Verbose)
	default:
		showHelp()
		os.Exit(1)
	}

	duration := time.Since(startTime)
	printSuccess(fmt.Sprintf("Build completed in %s", duration.Round(time.Millisecond)))
}

func printHeader() {
	fmt.Println(colorCyan + "===========================================" + colorReset)
	fmt.Println(colorCyan + "       ISX Pulse - Build System      " + colorReset)
	fmt.Println(colorCyan + "   The Heartbeat of Iraqi Markets    " + colorReset)
	fmt.Println(colorCyan + "===========================================" + colorReset)
	fmt.Println()
}

func printInfo(msg string) {
	fmt.Printf("%s[INFO]%s %s\n", colorBlue, colorReset, msg)
}

func printSuccess(msg string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", colorGreen, colorReset, msg)
}

func printError(msg string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, msg)
}

func printWarning(msg string) {
	fmt.Printf("%s[WARNING]%s %s\n", colorYellow, colorReset, msg)
}

func enableWindowsColors() {
	// Enable ANSI color codes on Windows
	cmd := exec.Command("cmd", "/c", "echo", "")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Run()
}

// Build all components
func buildAll(ctx *BuildContext) {
	printInfo("Building all components...")

	// Clear logs before building
	clearLogs(ctx.Verbose)

	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		printError(fmt.Sprintf("Prerequisites check failed: %v", err))
		os.Exit(1)
	}

	// Check Python environment for backtesting
	checkPythonEnvironment(ctx.Verbose)

	// Clean and create directories
	prepareDirectories(ctx.Verbose)

	// Build frontend first (needed for web executable)
	buildFrontend(ctx.Verbose)

	// Build all executables
	for name := range executables {
		buildExecutableWithContext(name, ctx)
	}

	// Copy configuration files
	copyConfigFiles(ctx.Verbose)

	printSuccess("All components built successfully!")
}

// Build web server only
func buildWeb(ctx *BuildContext) {
	printInfo("Building web server...")

	// Web server needs frontend to be built first
	if err := checkFrontendBuilt(); err != nil {
		printWarning("Frontend not built, building now...")
		buildFrontend(ctx.Verbose)
	}

	buildExecutableWithContext("web-licensed", ctx)
}

// Build a specific executable with build context
func buildExecutableWithContext(name string, ctx *BuildContext) {
	exeName, ok := executables[name]
	if !ok {
		printError(fmt.Sprintf("Unknown executable: %s", name))
		os.Exit(1)
	}

	printInfo(fmt.Sprintf("Building %s...", name))

	// Prepare build command
	outputPath := filepath.Join(distDir, exeName)
	sourcePath := "./cmd/" + name

	// Build flags for optimization and reproducible builds
	// Using time.Now() as fallback, but these should be overridden via build system for reproducibility
	buildTime := time.Now().Format(time.RFC3339)
	buildID := fmt.Sprintf("%x", time.Now().UnixNano())[:12] // Simple hash fallback

	ldflags := fmt.Sprintf("-s -w -X isxcli/internal/app.BuildTime=%s -X isxcli/internal/app.BuildID=%s -X main.Version=%s -X main.BuildTime=%s",
		buildTime, buildID, version, buildTime)

	// Add scratch card configuration to build flags
	if ctx.EnableScratchCard {
		ldflags += " -X main.EnableScratchCard=true"
		if ctx.AppsScriptURL != "" {
			ldflags += fmt.Sprintf(" -X main.AppsScriptURL=%s", ctx.AppsScriptURL)
		}
		printInfo("Building with scratch card license system enabled")
	}

	// Build command
	args := []string{
		"build",
		"-ldflags", ldflags,
		"-o", outputPath,
		sourcePath,
	}

	if ctx.Verbose {
		args = append([]string{"build", "-v"}, args[1:]...)
	}

	// Execute build
	cmd := exec.Command("go", args...)
	// Use api directory if it exists, otherwise fall back to dev
	if _, err := os.Stat(apiDir); err == nil {
		cmd.Dir = apiDir
	} else {
		cmd.Dir = apiDir
	}

	if ctx.Verbose {
		fmt.Printf("Running from %s: go %s\n", apiDir, strings.Join(args, " "))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		printError(fmt.Sprintf("Failed to build %s: %v", name, err))
		os.Exit(1)
	}

	// Report file size
	if info, err := os.Stat(outputPath); err == nil {
		sizeMB := float64(info.Size()) / 1024 / 1024
		printSuccess(fmt.Sprintf("Built %s (%.1f MB)", exeName, sizeMB))
	}
}

// Build a specific executable (backward compatibility)
func buildExecutable(name string, verbose bool) {
	ctx := &BuildContext{
		Verbose:           verbose,
		EnableScratchCard: false,
		AppsScriptURL:     "",
	}
	buildExecutableWithContext(name, ctx)
}

// Build frontend
func buildFrontend(verbose bool) {
	printInfo("Building frontend...")

	// Check if npm is available
	if err := checkNpm(); err != nil {
		printError("npm is not installed or not in PATH")
		printError("Please install Node.js from https://nodejs.org/")
		os.Exit(1)
	}

	// Install dependencies if needed
	nodeModules := filepath.Join(webDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		printInfo("Installing frontend dependencies...")
		cmd := exec.Command("npm", "ci")
		cmd.Dir = webDir
		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			printError(fmt.Sprintf("Failed to install dependencies: %v", err))
			os.Exit(1)
		}
	}

	// Build frontend
	printInfo("Running `npm run build` for Next.js (this may take several minutes)...")
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = webDir
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		printError(fmt.Sprintf("Failed to build frontend: %v", err))
		os.Exit(1)
	}
	printSuccess("Next.js build completed")

	// Copy to web-licensed command directory for embedding
	srcDir := filepath.Join(webDir, "out")
	// Try new structure first, then fall back to old
	destDir := filepath.Join(apiDir, "cmd", "web-licensed", "frontend")
	if _, err := os.Stat(filepath.Dir(destDir)); os.IsNotExist(err) {
		destDir = filepath.Join(apiDir, "cmd", "web-licensed", "frontend")
	}

	// Remove old frontend build
	os.RemoveAll(destDir)

	// Copy new build
	if err := copyDir(srcDir, destDir); err != nil {
		printError(fmt.Sprintf("Failed to copy frontend build: %v", err))
		os.Exit(1)
	}

	// Clean empty directories that break Go embedding
	printInfo("Optimizing frontend for Go embedding...")
	if err := cleanEmptyDirectories(destDir); err != nil {
		printWarning(fmt.Sprintf("Warning during optimization: %v", err))
	}

	// Validate frontend build meets requirements
	printInfo("Validating frontend build...")
	if err := validateFrontendBuild(destDir); err != nil {
		printError(fmt.Sprintf("Frontend validation failed: %v", err))
		os.Exit(1)
	}

	printSuccess("Frontend built and optimized for embedding")
}

// Clean build artifacts
func clean(verbose bool) {
	printInfo("Cleaning build artifacts and logs...")

	// Clear all log files first
	clearLogs(verbose)

	// Remove dist directory contents but keep the directory
	if err := cleanDir(distDir); err != nil {
		printError(fmt.Sprintf("Failed to clean dist directory: %v", err))
	}

	// Remove frontend build from cmd/web-licensed
	frontendBuild := filepath.Join(apiDir, "cmd", "web-licensed", "frontend")
	if err := os.RemoveAll(frontendBuild); err != nil && !os.IsNotExist(err) {
		printError(fmt.Sprintf("Failed to clean frontend build: %v", err))
	}

	// Remove Next.js build directories
	nextDirs := []string{
		filepath.Join(webDir, ".next"),
		filepath.Join(webDir, "out"),
	}

	for _, dir := range nextDirs {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			printError(fmt.Sprintf("Failed to clean %s: %v", dir, err))
		}
	}

	printSuccess("Build artifacts cleaned")
}

// Run tests
func runTests(verbose bool) {
	printInfo("Running tests...")

	// Go tests
	printInfo("Running Go tests...")
	args := []string{"test", "-race"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = apiDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		printError(fmt.Sprintf("Go tests failed: %v", err))
		os.Exit(1)
	}

	// Frontend tests
	printInfo("Running frontend tests...")
	cmd = exec.Command("npm", "test", "--", "--passWithNoTests")
	cmd.Dir = webDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		printError(fmt.Sprintf("Frontend tests failed: %v", err))
		os.Exit(1)
	}

	printSuccess("All tests passed")
}

// Build release version with optimizations
func buildRelease(ctx *BuildContext) {
	printInfo("Building release version...")

	// Clear logs first
	clearLogs(ctx.Verbose)

	// Clean artifacts
	clean(ctx.Verbose)

	// Set release environment
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", "windows")
	os.Setenv("GOARCH", "amd64")

	// Build all components
	buildAll(ctx)

	// Setup Python service files for distribution
	setupPythonService(ctx.Verbose)

	// Create version file
	versionFile := filepath.Join(distDir, "VERSION.txt")
	content := fmt.Sprintf("ISX Pulse v%s\nThe Heartbeat of Iraqi Markets\nBuilt: %s\n",
		version, time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(versionFile, []byte(content), 0644)

	printSuccess("Release build completed")
}

// Create distribution package
func createPackage(verbose bool) {
	printInfo("Creating distribution package...")

	// Create a build context for release build if needed
	ctx := &BuildContext{
		Verbose:           verbose,
		EnableScratchCard: os.Getenv("ENABLE_SCRATCH_CARD_MODE") == "true",
		AppsScriptURL:     os.Getenv("GOOGLE_APPS_SCRIPT_URL"),
	}

	// Ensure release build exists
	if _, err := os.Stat(filepath.Join(distDir, "ISXPulse.exe")); os.IsNotExist(err) {
		printWarning("No release build found, building now...")
		buildRelease(ctx)
	}

	// Package name with version
	packageName := fmt.Sprintf("ISXPulse-v%s", version)
	packageDir := filepath.Join(rootDir, packageName)

	// Remove old package
	os.RemoveAll(packageDir)

	// Copy dist files to package
	if err := copyDir(distDir, packageDir); err != nil {
		printError(fmt.Sprintf("Failed to create package: %v", err))
		os.Exit(1)
	}

	// Add documentation
	docs := map[string]string{
		"README.md":   filepath.Join(packageDir, "README.txt"),
		"LICENSE.txt": filepath.Join(packageDir, "LICENSE.txt"),
	}

	for src, dest := range docs {
		if _, err := os.Stat(src); err == nil {
			copyFile(src, dest)
		}
	}

	// Create zip file (requires external tool on Windows)
	zipName := packageName + ".zip"
	printInfo(fmt.Sprintf("Package created in %s", packageDir))
	printInfo(fmt.Sprintf("To create zip: powershell Compress-Archive -Path '%s' -DestinationPath '%s'", packageDir, zipName))

	printSuccess("Distribution package ready")
}

// Helper functions

func checkPrerequisites() error {
	// Check Go
	if err := exec.Command("go", "version").Run(); err != nil {
		return fmt.Errorf("Go is not installed or not in PATH")
	}

	// Check source directories
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		return fmt.Errorf("dev directory not found")
	}

	return nil
}

func checkNpm() error {
	return exec.Command("npm", "--version").Run()
}

func checkPythonEnvironment(verbose bool) {
	printInfo("Checking Python environment for backtesting support...")

	// Check if Python is installed
	pythonCmd := "python"
	if runtime.GOOS == "windows" {
		// Try python3 first on Windows
		if err := exec.Command("python3", "--version").Run(); err == nil {
			pythonCmd = "python3"
		}
	}

	if err := exec.Command(pythonCmd, "--version").Run(); err != nil {
		printWarning("Python not found - Backtesting features will be unavailable")
		printWarning("To enable backtesting, install Python 3.8+ and required packages")
		return
	}

	// Check Python version
	cmd := exec.Command(pythonCmd, "-c", "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		printInfo(fmt.Sprintf("Python %s detected", version))

		// Check for minimum version (3.8)
		if version < "3.8" {
			printWarning("Python 3.8+ required for backtesting (found " + version + ")")
			return
		}
	}

	// Check if Freqtrade service requirements are installed
	requirementsFile := filepath.Join(apiDir, "services", "freqtrade", "requirements.txt")
	if _, err := os.Stat(requirementsFile); err == nil {
		printInfo("Checking Freqtrade dependencies...")

		// Check if key packages are installed
		packages := []string{"fastapi", "uvicorn", "pandas", "numpy"}
		allInstalled := true

		for _, pkg := range packages {
			cmd := exec.Command(pythonCmd, "-c", fmt.Sprintf("import %s", pkg))
			if err := cmd.Run(); err != nil {
				allInstalled = false
				if verbose {
					printWarning(fmt.Sprintf("Package %s not installed", pkg))
				}
			}
		}

		if !allInstalled {
			printWarning("Some Python packages are missing")
			printInfo(fmt.Sprintf("To install: %s -m pip install -r %s", pythonCmd, requirementsFile))
		} else {
			printSuccess("Python environment ready for backtesting")
		}
	}
}

func setupPythonService(verbose bool) {
	printInfo("Setting up Python backtesting service...")

	pythonServiceDir := filepath.Join(apiDir, "services", "freqtrade")
	requirementsFile := filepath.Join(pythonServiceDir, "requirements.txt")

	// Create requirements.txt if it doesn't exist
	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		requirements := `fastapi==0.104.1
uvicorn==0.24.0
pandas==2.1.3
numpy==1.26.2
ta==0.11.0
scikit-learn==1.3.2
websockets==12.0
pydantic==2.5.0
python-multipart==0.0.6
`
		if err := os.WriteFile(requirementsFile, []byte(requirements), 0644); err != nil {
			printError(fmt.Sprintf("Failed to create requirements.txt: %v", err))
			return
		}
		printInfo("Created requirements.txt for Python service")
	}

	// Create startup script for Windows
	if runtime.GOOS == "windows" {
		startScript := filepath.Join(distDir, "start-backtesting-service.bat")
		scriptContent := `@echo off
echo Starting ISX Pulse Backtesting Service...
cd /d "%~dp0"
python -m uvicorn api.services.freqtrade.server:app --host 0.0.0.0 --port 8000 --reload
pause
`
		if err := os.WriteFile(startScript, []byte(scriptContent), 0755); err != nil {
			printWarning("Failed to create backtesting service startup script")
		} else {
			printInfo("Created backtesting service startup script")
		}
	}
}

func checkFrontendBuilt() error {
	frontendBuild := filepath.Join(apiDir, "cmd", "web-licensed", "frontend", "index.html")
	if _, err := os.Stat(frontendBuild); os.IsNotExist(err) {
		return fmt.Errorf("frontend not built")
	}
	return nil
}

func prepareDirectories(verbose bool) {
	// Create dist directory structure
	dirs := []string{
		distDir,
		filepath.Join(distDir, "data", "downloads"),
		filepath.Join(distDir, "data", "reports"),
		filepath.Join(distDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
		}
	}

	// Preserve license.dat if exists
	licenseSrc := filepath.Join(distDir, "license.dat")
	if _, err := os.Stat(licenseSrc); err == nil {
		printInfo("Preserving existing license.dat")
		licenseBackup := filepath.Join(rootDir, "license.dat.backup")
		copyFile(licenseSrc, licenseBackup)
		defer func() {
			copyFile(licenseBackup, licenseSrc)
			os.Remove(licenseBackup)
		}()
	}
}

func copyConfigFiles(verbose bool) {
	configs := map[string]string{
		"credentials.json.example":   filepath.Join(distDir, "credentials.json.example"),
		"sheets-config.json.example": filepath.Join(distDir, "sheets-config.json.example"),
		"start-server.bat":           filepath.Join(distDir, "start-server.bat"),
	}

	for src, dest := range configs {
		if _, err := os.Stat(src); err == nil {
			if err := copyFile(src, dest); err != nil {
				printWarning(fmt.Sprintf("Failed to copy %s: %v", src, err))
			}
		}
	}
}

func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(dest, input, 0644)
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return copyFile(path, destPath)
	})
}

// Frontend build validation following industry standards
type FrontendAssetSpec struct {
	RequiredFiles     []string
	RequiredDirs      []string
	ForbiddenPatterns []string
}

var productionAssets = FrontendAssetSpec{
	RequiredFiles: []string{
		"index.html",
		"404.html",
		"index.txt",   // Next.js 14+ React Server Components
		"favicon.ico",
		"site.webmanifest",
	},
	RequiredDirs: []string{
		"_next",
	},
	ForbiddenPatterns: []string{
		"*.map",        // No source maps in production
		".env*",        // No environment files
		"*.test.*",     // No test files
		"node_modules", // No dependencies
		".git*",        // No git files
	},
}

// validateFrontendBuild ensures frontend meets embedding requirements
func validateFrontendBuild(dir string) error {
	// Check required files
	for _, file := range productionAssets.RequiredFiles {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("missing required file: %s", file)
		}
	}

	// Check required directories exist and aren't empty
	for _, d := range productionAssets.RequiredDirs {
		path := filepath.Join(dir, d)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return fmt.Errorf("missing required directory: %s", d)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", d)
		}

		entries, _ := os.ReadDir(path)
		if len(entries) == 0 {
			return fmt.Errorf("required directory %s is empty", d)
		}
	}

	return nil
}

// cleanEmptyDirectories removes empty dirs that Next.js creates
func cleanEmptyDirectories(dir string) error {
	var emptyDirs []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			entries, _ := os.ReadDir(path)
			if len(entries) == 0 {
				emptyDirs = append(emptyDirs, path)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Remove from bottom up to handle nested empty dirs
	for i := len(emptyDirs) - 1; i >= 0; i-- {
		os.Remove(emptyDirs[i])
	}

	if len(emptyDirs) > 0 {
		printInfo(fmt.Sprintf("Removed %d empty directories", len(emptyDirs)))
	}
	return nil
}

// Clear all log files
func clearLogs(verbose bool) {
	printInfo("Clearing log files...")

	// Define log locations
	logDirs := []string{
		filepath.Join(distDir, "logs"),
		filepath.Join(apiDir, "logs"),
		filepath.Join(rootDir, "logs"),
	}

	// Clear log files in each directory
	for _, dir := range logDirs {
		if _, err := os.Stat(dir); err == nil {
			// Remove all .log files
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip errors
				}
				if !info.IsDir() && strings.HasSuffix(path, ".log") {
					if verbose {
						fmt.Printf("  Removing: %s\n", path)
					}
					os.Remove(path)
				}
				return nil
			})
		}
	}

	// Also clear root level log files
	rootLogs, _ := filepath.Glob(filepath.Join(rootDir, "*.log"))
	for _, logFile := range rootLogs {
		if verbose {
			fmt.Printf("  Removing: %s\n", logFile)
		}
		os.Remove(logFile)
	}

	printSuccess("Log files cleared")
}

func cleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == "license.dat" {
			continue // Preserve license
		}
		if name == "data" {
			continue // Preserve data folder (downloads, reports, cache)
		}
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func showHelp() {
	fmt.Println("Usage: go run build.go [-target=TARGET] [-v] [-scratch-card] [-apps-script-url=URL]")
	fmt.Println()
	fmt.Println("Targets:")
	fmt.Println("  all               Build all components (default)")
	fmt.Println("  web               Build ISX Pulse server")
	fmt.Println("  web-licensed      Build ISX Pulse server")
	fmt.Println("  scraper           Build scraper only")
	fmt.Println("  processor         Build processor only")
	fmt.Println("  indexcsv          Build indexcsv only")
	fmt.Println("  activate-license  Build license activation tool")
	fmt.Println("  frontend          Build frontend only")
	fmt.Println("  clean             Clean build artifacts")
	fmt.Println("  test              Run all tests")
	fmt.Println("  release           Build optimized release version")
	fmt.Println("  package           Create distribution package")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v                Verbose output")
	fmt.Println("  -scratch-card     Enable scratch card license system")
	fmt.Println("  -apps-script-url  Google Apps Script URL for license management")
	fmt.Println()
	fmt.Println("Scratch Card Examples:")
	fmt.Println("  go run build.go -scratch-card -apps-script-url=https://script.google.com/macros/s/YOUR_ID/exec")
	fmt.Println("  go run build.go -target=release -scratch-card")
}
