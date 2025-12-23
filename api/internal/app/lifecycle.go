package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/isxcli/isxcli/internal/config"

	"log/slog"
)

// checkPortAvailable checks if the specified port is available for binding
func (a *Application) checkPortAvailable(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is already in use - another ISXPulse instance may be running. "+
			"Please stop the existing instance or use a different port", port)
	}
	listener.Close()
	return nil
}

// createPIDFile creates a PID file to prevent multiple instances
func (a *Application) createPIDFile() error {
	// First, try to create Windows Mutex for additional protection
	if err := a.createSingleInstanceMutex(); err != nil {
		return err
	}

	paths, err := config.GetPaths()
	if err != nil {
		a.cleanupSingleInstanceMutex() // Clean up mutex on error
		return fmt.Errorf("failed to get paths for PID file: %w", err)
	}

	pidFile := filepath.Join(paths.ExecutableDir, "isxpulse.pid")

	// Check if PID file already exists
	if _, err := os.Stat(pidFile); err == nil {
		// Read existing PID file
		pidData, err := os.ReadFile(pidFile)
		if err == nil {
			existingPID := string(pidData)
			// Check if the process is still running using fixed Windows detection
			if a.isProcessRunning(existingPID) {
				a.cleanupSingleInstanceMutex() // Clean up mutex on error
				return fmt.Errorf("another ISXPulse instance is already running with PID %s. "+
					"Please stop the existing instance before starting a new one", existingPID)
			}
			// Process is not running, remove stale PID file
			os.Remove(pidFile)
			a.Logger.Info("Removed stale PID file", "existing_pid", existingPID)
		}
	}

	// Create new PID file with current process ID
	currentPID := strconv.Itoa(os.Getpid())
	err = os.WriteFile(pidFile, []byte(currentPID), 0644)
	if err != nil {
		a.cleanupSingleInstanceMutex() // Clean up mutex on error
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	a.Logger.Info("PID file created successfully", "pid", currentPID)
	return nil
}

// isProcessRunning checks if a process with the given PID is still running
func (a *Application) isProcessRunning(pidStr string) bool {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// Use platform-specific implementation
	if runtime.GOOS == "windows" {
		return isProcessRunningWindows(pid)
	}

	// For non-Windows systems, use the signal-based approach
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to signal the process (signal 0)
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// removePIDFile removes the PID file on shutdown
func (a *Application) removePIDFile() {
	paths, err := config.GetPaths()
	if err != nil {
		return
	}

	pidFile := filepath.Join(paths.ExecutableDir, "isxpulse.pid")
	os.Remove(pidFile) // Ignore errors during cleanup

	// Clean up Windows Mutex
	a.cleanupSingleInstanceMutex()
}

// createServer creates the HTTP server
func (a *Application) createServer() {
	a.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Config.Server.Port),
		Handler:      a.Router,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}
}

// Start starts the application
func (a *Application) Start(ctx context.Context, cancel context.CancelFunc) error {
	// Check if port is available
	if err := a.checkPortAvailable(a.Config.Server.Port); err != nil {
		a.Logger.ErrorContext(ctx, "Port conflict detected", slog.String("error", err.Error()))
		return err
	}

	// Create PID file to prevent multiple instances
	if err := a.createPIDFile(); err != nil {
		a.Logger.ErrorContext(ctx, "Failed to create PID file", slog.String("error", err.Error()))
		return err
	}

	a.Logger.InfoContext(ctx, "Starting application",
		slog.String("name", AppName),
		slog.String("version", VERSION),
		slog.Int("port", a.Config.Server.Port),
		slog.String("level", a.Config.Logging.Level))

	// Log important paths for debugging
	paths, _ := config.GetPaths()
	a.Logger.InfoContext(ctx, "Application paths",
		slog.String("executable_dir", paths.ExecutableDir),
		slog.String("data_dir", paths.DataDir),
		slog.String("web_dir", paths.WebDir),
		slog.String("logs_dir", paths.LogsDir),
		slog.String("license_file", paths.LicenseFile))

	// Start background services
	// Note: WebSocketHub already started in initializeServices() - no need to start again
	if a.JobQueue != nil {
		a.Logger.InfoContext(ctx, "Starting job queue with application context")
		a.JobQueue.Start(ctx)
	}

	// Start server
	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.ErrorContext(ctx, "Server error", slog.String("error", err.Error()))
			// Signal shutdown through context instead of os.Exit
			cancel()
		}
	}()

	// Perform health check on critical paths
	err := a.performStartupHealthCheck(ctx)
	if err != nil {
		a.Logger.WarnContext(ctx, "Startup health check warnings", slog.String("warnings", err.Error()))
	}

	a.Logger.InfoContext(ctx, "Application started successfully",
		slog.String("address", fmt.Sprintf("http://localhost:%d", a.Config.Server.Port)),
		slog.String("license_status", "checking..."))

	// Open browser after server is ready
	go func() {
		// Create independent context for browser opening
		browserCtx := context.Background()

		// Wait for server to be ready by checking health endpoint
		url := fmt.Sprintf("http://localhost:%d", a.Config.Server.Port)
		healthURL := fmt.Sprintf("%s/api/health", url)

		// Try to connect to health endpoint with retries
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			// Check if main context is cancelled
			select {
			case <-ctx.Done():
				a.Logger.InfoContext(ctx, "Browser opening cancelled - application shutting down")
				return
			default:
			}

			// Try health check
			resp, err := http.Get(healthURL)
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				a.Logger.InfoContext(browserCtx, "Server is ready, opening browser",
					slog.String("url", url),
					slog.Int("attempts", i+1))

				// Server is ready, open browser
				if err := openBrowser(url); err != nil {
					a.Logger.ErrorContext(browserCtx, "Failed to open browser after server ready",
						slog.String("error", err.Error()),
						slog.String("url", url))

					// Print manual instruction
					fmt.Printf("\n")
					fmt.Printf("========================================\n")
					fmt.Printf("ISX Pulse is running!\n")
					fmt.Printf("Please open your browser and navigate to:\n")
					fmt.Printf("  %s\n", url)
					fmt.Printf("========================================\n")
					fmt.Printf("\n")
				} else {
					a.Logger.InfoContext(browserCtx, "Browser opened successfully",
						slog.String("url", url))
				}
				return
			}

			if resp != nil {
				resp.Body.Close()
			}

			// Wait before retry
			time.Sleep(500 * time.Millisecond)
		}

		// If we get here, server never became ready
		a.Logger.ErrorContext(browserCtx, "Server did not become ready for browser opening",
			slog.String("url", url),
			slog.Int("max_retries", maxRetries))
	}()

	return nil
}

// Stop gracefully stops the application
func (a *Application) Stop(ctx context.Context) error {
	a.Logger.InfoContext(ctx, "Shutting down application")

	// Signal programmatic shutdown (close channel if not already closed)
	select {
	case <-a.shutdownChan:
		// Already closed
	default:
		close(a.shutdownChan)
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, a.Config.Server.ShutdownTimeout)
	defer cancel()

	// Stop server
	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Stop background services
	a.WebSocketHub.Stop()

	// Remove PID file
	a.removePIDFile()

	// Stop job queue with timeout
	if a.JobQueue != nil {
		a.Logger.InfoContext(ctx, "Stopping job queue")
		if err := a.JobQueue.Stop(30 * time.Second); err != nil {
			a.Logger.ErrorContext(ctx, "Failed to stop job queue gracefully", slog.String("error", err.Error()))
		}
	}

	// Cancel running operations
	if err := a.OperationService.CancelAll(ctx); err != nil {
		a.Logger.ErrorContext(ctx, "Error cancelling operations", slog.String("error", err.Error()))
	}

	// Shutdown OpenTelemetry providers
	if a.OTelProviders != nil {
		if err := a.OTelProviders.Shutdown(shutdownCtx); err != nil {
			a.Logger.ErrorContext(ctx, "Error shutting down OpenTelemetry", slog.String("error", err.Error()))
		}
	}

	a.Logger.InfoContext(ctx, "Application shutdown complete")
	return nil
}

// Run runs the application until interrupted
func (a *Application) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start application
	if err := a.Start(ctx, cancel); err != nil {
		return err
	}

	// Wait for interrupt (OS signal or programmatic shutdown)
	select {
	case <-sigChan:
		a.Logger.InfoContext(ctx, "Received interrupt signal")
	case <-a.shutdownChan:
		a.Logger.InfoContext(ctx, "Received programmatic shutdown")
	}

	// Graceful shutdown
	return a.Stop(ctx)
}

// performStartupHealthCheck performs health checks on critical paths and resources
func (a *Application) performStartupHealthCheck(ctx context.Context) error {
	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	var warnings []string

	// Check critical directories are writable
	directories := map[string]string{
		"Data":      paths.DataDir,
		"Downloads": paths.DownloadsDir,
		"Reports":   paths.ReportsDir,
		"Cache":     paths.CacheDir,
		"Logs":      paths.LogsDir,
	}

	for name, dir := range directories {
		// Try to create a test file to verify write access
		testFile := filepath.Join(dir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s directory not writable: %s", name, dir))
		} else {
			// Clean up test file
			os.Remove(testFile)
		}
	}

	// Check web directory exists and has content
	if !config.FileExists(paths.WebDir) {
		warnings = append(warnings, fmt.Sprintf("Web directory not found: %s", paths.WebDir))
	}

	// Check for critical configuration files (non-fatal)
	configFiles := map[string]string{
		"Credentials":   paths.CredentialsFile,
		"Sheets Config": paths.SheetsConfigFile,
	}

	for name, file := range configFiles {
		if !config.FileExists(file) {
			a.Logger.InfoContext(ctx, "Configuration file not found",
				slog.String("file", name),
				slog.String("path", file))
		}
	}

	if len(warnings) > 0 {
		return fmt.Errorf("startup health check warnings: %s", strings.Join(warnings, "; "))
	}

	a.Logger.InfoContext(ctx, "Startup health check passed")
	return nil
}

// openBrowser opens the default browser to the specified URL with retry logic
func openBrowser(url string) error {
	var lastErr error

	// Try multiple methods with retries
	methods := getBrowserOpenMethods(url)

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
			slog.Info("Retrying browser open",
				slog.Int("attempt", attempt+1),
				slog.String("url", url))
		}

		for _, method := range methods {
			slog.Info("Attempting to open browser",
				slog.String("method", method.name),
				slog.String("command", method.cmd),
				slog.Any("args", method.args),
				slog.String("url", url))

			// Set a timeout for the command - create context per attempt
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			cmd := exec.CommandContext(ctx, method.cmd, method.args...)

			if err := cmd.Start(); err != nil {
				cancel() // Cancel immediately since we're not deferring in a loop
				lastErr = err
				slog.Warn("Browser open method failed",
					slog.String("method", method.name),
					slog.String("error", err.Error()))
				continue
			}

			// Give the browser a moment to start
			time.Sleep(500 * time.Millisecond)
			cancel() // Cancel the context after successful start

			// Success
			slog.Info("Browser opened successfully",
				slog.String("method", method.name),
				slog.String("url", url))
			return nil
		}
	}

	return fmt.Errorf("failed to open browser after all attempts: %w", lastErr)
}

// browserMethod represents a method to open the browser
type browserMethod struct {
	name string
	cmd  string
	args []string
}

// getBrowserOpenMethods returns platform-specific browser opening methods
func getBrowserOpenMethods(url string) []browserMethod {
	switch runtime.GOOS {
	case "windows":
		return []browserMethod{
			{
				name: "start_command",
				cmd:  "cmd",
				args: []string{"/c", "start", "", url},
			},
			{
				name: "rundll32",
				cmd:  "rundll32",
				args: []string{"url.dll,FileProtocolHandler", url},
			},
			{
				name: "powershell",
				cmd:  "powershell",
				args: []string{"-Command", fmt.Sprintf("Start-Process '%s'", url)},
			},
			{
				name: "explorer",
				cmd:  "explorer",
				args: []string{url},
			},
		}
	case "darwin":
		return []browserMethod{
			{
				name: "open",
				cmd:  "open",
				args: []string{url},
			},
		}
	default: // Linux and others
		return []browserMethod{
			{
				name: "xdg-open",
				cmd:  "xdg-open",
				args: []string{url},
			},
			{
				name: "sensible-browser",
				cmd:  "sensible-browser",
				args: []string{url},
			},
			{
				name: "firefox",
				cmd:  "firefox",
				args: []string{url},
			},
			{
				name: "chromium",
				cmd:  "chromium",
				args: []string{url},
			},
		}
	}
}
