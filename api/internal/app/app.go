package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/operations"
	"github.com/isxcli/isxcli/internal/services"
	ws "github.com/isxcli/isxcli/internal/websocket"
	"net/http"
)

const (
	VERSION    = "enhanced-v3.0.0"
	REPO_URL   = "https://github.com/haideralmesaody/ISXDailyReportScrapper"
	AppName    = "ISX Pulse - The Heartbeat of Iraqi Markets"
	Executable = "ISXPulse.exe"
)

var (
	// BuildTime is set at compile time via ldflags
	BuildTime = "unknown" // Default value, overridden at build time
	// BuildID is a unique identifier for this build, set via ldflags
	BuildID = "unknown" // Default value, overridden at build time
)

// Application represents the main application container
type Application struct {
	Config           *config.Config
	Router           *chi.Mux
	Server           *http.Server
	LicenseManager   *license.Manager
	WebSocketHub     *ws.Hub
	OperationService *services.OperationService
	DataService      *services.DataService
	HealthService    *services.HealthService
	Logger           *slog.Logger // Single slog instance per CLAUDE.md
	Services         *ServiceContainer
	OTelProviders    *infrastructure.OTelProviders // OpenTelemetry providers
	FrontendFS       fs.FS                         // Embedded frontend filesystem
	JobQueue         *operations.JobQueue          // Async job queue for operations
	shutdownChan     chan struct{}                 // Channel for programmatic shutdown (used by tests)
	// Platform-specific single instance enforcement (reused for POSIX compatibility)
	windowsMutex uintptr // Platform-specific mutex handle
}

// ServiceContainer holds all application services
type ServiceContainer struct {
	License        *license.Manager
	LicenseService services.LicenseService
	Data           *services.DataService
	Health         *services.HealthService
	WebSocket      *ws.Hub
	Liquidity      *services.LiquidityService
	Strategy       *services.StrategyService
}

// NewApplication creates a new application instance with dependency injection
func NewApplication(frontendFS fs.FS) (*Application, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize single infrastructure logger per CLAUDE.md
	logger, err := infrastructure.InitializeLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log startup information
	logger.Info("Application starting",
		slog.String("name", AppName),
		slog.String("version", VERSION),
		slog.String("executable", Executable))

	// Validate and log all paths at startup for debugging
	paths, err := config.GetPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure all required directories exist
	logger.Info("Ensuring required directories exist")
	if err := paths.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Log all resolved paths at startup for debugging
	paths.LogPathResolution()

	// Create application instance with minimal setup
	app := &Application{
		Config:       cfg,
		Logger:       logger,
		FrontendFS:   frontendFS,
		shutdownChan: make(chan struct{}), // Initialize shutdown channel for programmatic shutdown
	}

	// NOTE: License validation runs on startup but doesn't fail-fast to allow web-based license activation
	// Server will start regardless of license status and handle licensing via middleware
	logger.Info("License validation running in non-fail-fast mode - allowing server startup")

	// We validate license on startup but continue even if invalid (middleware will handle access control)
	if err := app.ValidateLicenseOnStartup(context.Background()); err != nil {
		logger.Warn("License validation failed on startup - server will continue in unlicensed mode",
			slog.String("error", err.Error()),
			slog.String("phase", "startup_validation"))
	} else {
		logger.Info("License validation successful - proceeding in licensed mode")
	}

	// Initialize OpenTelemetry (only after license validation succeeds)
	otelProviders, err := infrastructure.InitializeOTel(infrastructure.DefaultOTelConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
	}
	app.OTelProviders = otelProviders

	// Initialize global operation tracer
	if err := operations.InitGlobalOperationTracer(otelProviders); err != nil {
		return nil, fmt.Errorf("failed to initialize operation tracer: %w", err)
	}

	// Initialize services in order (server starts regardless of license status)
	if err := app.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Setup router
	app.setupRouter()

	// Create HTTP server
	app.createServer()

	return app, nil
}
