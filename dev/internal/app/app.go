package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"isxcli/internal/config"
	"isxcli/internal/errors"
	"isxcli/internal/handlers"
	"isxcli/internal/license"
	customMiddleware "isxcli/internal/middleware"
	"isxcli/internal/services"
	"isxcli/internal/updater"
	ws "isxcli/internal/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

const (
	VERSION    = "enhanced-v3.0.0"
	REPO_URL   = "https://github.com/haideralmesaody/ISXDailyReportScrapper"
	AppName    = "ISX Daily Reports Scrapper"
	Executable = "web-licensed.exe"
)

// Application represents the main application container
type Application struct {
	Config          *config.Config
	Router          *chi.Mux
	Server          *http.Server
	LicenseManager  *license.Manager
	WebSocketHub    *ws.Hub
	PipelineService *services.PipelineService
	DataService     *services.DataService
	HealthService   *services.HealthService
	UpdateChecker   *updater.AutoUpdateChecker
	Logger          services.Logger
	SLogger         *slog.Logger
	Services        *ServiceContainer
}

// ServiceContainer holds all application services
type ServiceContainer struct {
	License  *license.Manager
	LicenseService services.LicenseService
	Pipeline *services.PipelineService
	Data     *services.DataService
	Health   *services.HealthService
	WebSocket *ws.Hub
}

// NewApplication creates a new application instance with dependency injection
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger, err := services.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	// Initialize slog logger
	var slogLevel slog.Level
	switch cfg.Logging.Level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}
	slogHandler := slog.NewJSONHandler(os.Stdout, opts)
	slogger := slog.New(slogHandler)

	// Create application
	app := &Application{
		Config:  cfg,
		Logger:  logger,
		SLogger: slogger,
	}

	// Initialize services in order
	if err := app.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Setup router
	app.setupRouter()

	// Create HTTP server
	app.createServer()

	return app, nil
}

// initializeServices initializes all application services
func (a *Application) initializeServices() error {
	// Initialize license manager
	licensePath := a.Config.GetLicenseFile()
	licenseManager, err := license.NewManager(licensePath)
	if err != nil {
		return fmt.Errorf("failed to initialize license manager: %w", err)
	}
	a.LicenseManager = licenseManager

	// Initialize WebSocket hub
	hub := ws.NewHub()
	a.WebSocketHub = hub

	// Initialize pipeline service
	pipelineAdapter := services.NewWebSocketPipelineAdapter(hub)
	pipelineLogger := services.NewPipelineLogger("Pipeline", hub)
	pipelineService, err := services.NewPipelineService(pipelineAdapter, pipelineLogger, a.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize pipeline service: %w", err)
	}
	a.PipelineService = pipelineService

	// Initialize data service
	dataService, err := services.NewDataService(a.Config, a.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize data service: %w", err)
	}
	a.DataService = dataService

	// Initialize health service
	healthService := services.NewHealthService(
		VERSION,
		REPO_URL,
		a.Config.Paths,
		a.LicenseManager,
		a.PipelineService.GetManager(),
		a.WebSocketHub,
		a.Logger,
	)
	a.HealthService = healthService

	// Initialize update checker
	upd, err := updater.NewUpdater(VERSION, REPO_URL)
	if err != nil {
		return fmt.Errorf("failed to initialize updater: %w", err)
	}
	
	updateChecker := updater.NewAutoUpdateChecker(upd, 24*time.Hour, func(info *updater.UpdateInfo) bool {
		a.Logger.Info("Update available", "current", info.CurrentVersion, "latest", info.LatestVersion)
		return false // Don't auto-install
	})
	a.UpdateChecker = updateChecker

	// Initialize license service
	licenseService := services.NewLicenseService(licenseManager, a.SLogger)

	// Create service container
	a.Services = &ServiceContainer{
		License:   licenseManager,
		LicenseService: licenseService,
		Pipeline:  pipelineService,
		Data:      dataService,
		Health:    healthService,
		WebSocket: hub,
	}

	return nil
}

// setupRouter configures the HTTP router with all routes
func (a *Application) setupRouter() {
	r := chi.NewRouter()

	// Apply MINIMAL middleware that won't interfere with WebSocket
	// These are safe because they don't wrap the ResponseWriter
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// WebSocket route with minimal middleware
	// MUST be registered after minimal middleware but before the group
	r.HandleFunc("/ws", a.handleWebSocket)

	// Create a route group for everything else with FULL middleware
	r.Group(func(r chi.Router) {
		// Apply remaining middleware only to this group
		r.Use(middleware.Recoverer)
		r.Use(middleware.Timeout(a.Config.Server.ReadTimeout))
		r.Use(customMiddleware.NewStructuredLogger(a.Logger))
		r.Use(customMiddleware.SecurityHeaders)
		
		// CORS middleware
		if a.Config.Security.EnableCORS {
			r.Use(customMiddleware.CORS(
				a.Config.Security.AllowedOrigins,
				[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				[]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			))
		}
		
		// Rate limiting
		if a.Config.Security.RateLimit.Enabled {
			r.Use(customMiddleware.NewRateLimiter(
				a.Config.Security.RateLimit.RPS,
				a.Config.Security.RateLimit.Burst,
			).Handler)
		}
		
		// License validation
		licenseValidator := customMiddleware.NewLicenseValidator(a.LicenseManager, a.SLogger)
		r.Use(licenseValidator.Handler)
		
		// Now register all other routes within this group
		a.setupAPIRoutes(r)
		a.setupStaticRoutes(r)
		a.setupHTMLRoutes(r)
	})

	a.Router = r
}

// setupMiddleware is no longer used - middleware is now applied in setupRouter using route groups
// Keeping this comment for reference to the middleware that was moved

// setupAPIRoutes configures API endpoints
func (a *Application) setupAPIRoutes(r chi.Router) {
	// API routes with common middleware
	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		// Health handler
		healthHandler := handlers.NewHealthHandler(a.HealthService, a.Logger)
		r.Get("/health", healthHandler.HealthCheck)
		r.Get("/health/ready", healthHandler.ReadinessCheck)
		r.Get("/health/live", healthHandler.LivenessCheck)
		r.Get("/version", healthHandler.Version)

		// License endpoints
		licenseHandler := handlers.NewLicenseHandler(a.Services.LicenseService, a.SLogger)
		r.Mount("/license", licenseHandler.Routes())

		// Create error handler
		errorHandler := errors.NewErrorHandler(a.SLogger, false)

		// Data handler
		dataHandler := handlers.NewDataHandler(a.DataService, a.SLogger, errorHandler)
		r.Mount("/data", dataHandler.Routes())

		// Pipeline handler
		pipelineServiceAdapter := handlers.NewPipelineServiceAdapter(a.PipelineService)
		pipelineHandler := handlers.NewPipelineHandler(pipelineServiceAdapter, a.SLogger, errorHandler)
		r.Mount("/pipeline", pipelineHandler.Routes())

		// Client logging endpoint
		r.Post("/logs", handlers.NewClientLogHandler(a.Logger).Handle)

		// Operation shortcuts
		r.Post("/scrape", func(w http.ResponseWriter, r *http.Request) {
			var params map[string]interface{}
			if err := render.DecodeJSON(r.Body, &params); err != nil {
				render.JSON(w, r, map[string]interface{}{"error": "Invalid request"})
				return
			}
			pipelineID, err := a.PipelineService.StartScraping(params)
			if err != nil {
				render.JSON(w, r, map[string]interface{}{"error": err.Error()})
				return
			}
			render.JSON(w, r, map[string]interface{}{"pipeline_id": pipelineID, "status": "started"})
		})
		r.Post("/process", func(w http.ResponseWriter, r *http.Request) {
			var params map[string]interface{}
			if err := render.DecodeJSON(r.Body, &params); err != nil {
				render.JSON(w, r, map[string]interface{}{"error": "Invalid request"})
				return
			}
			pipelineID, err := a.PipelineService.StartProcessing(params)
			if err != nil {
				render.JSON(w, r, map[string]interface{}{"error": err.Error()})
				return
			}
			render.JSON(w, r, map[string]interface{}{"pipeline_id": pipelineID, "status": "started"})
		})
		r.Post("/indexcsv", func(w http.ResponseWriter, r *http.Request) {
			var params map[string]interface{}
			if err := render.DecodeJSON(r.Body, &params); err != nil {
				render.JSON(w, r, map[string]interface{}{"error": "Invalid request"})
				return
			}
			pipelineID, err := a.PipelineService.StartIndexExtraction(params)
			if err != nil {
				render.JSON(w, r, map[string]interface{}{"error": err.Error()})
				return
			}
			render.JSON(w, r, map[string]interface{}{"pipeline_id": pipelineID, "status": "started"})
		})
	})
}

// setupStaticRoutes configures static file serving
func (a *Application) setupStaticRoutes(r chi.Router) {
	staticDir := filepath.Join(a.Config.GetWebDir(), "static")
	templatesDir := filepath.Join(a.Config.GetWebDir(), "templates")

	// Serve static files
	r.Route("/static", func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Handle("/*", http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))
	})

	// Serve templates
	r.Route("/templates", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/templates", http.FileServer(http.Dir(templatesDir))))
	})
}

// setupHTMLRoutes configures HTML page routes
func (a *Application) setupHTMLRoutes(r chi.Router) {
	// License page (default)
	r.Get("/", handlers.RedirectToLicense)
	r.Get("/license", handlers.ServeLicensePage(a.Config.GetWebDir()))
	r.Get("/app", handlers.ServeMainApp(a.Config.GetWebDir()))
	r.Get("/test", handlers.ServeTestPage())
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
func (a *Application) Start(ctx context.Context) error {
	a.Logger.Info("Starting application",
		"name", AppName,
		"version", VERSION,
		"port", a.Config.Server.Port,
		"level", a.Config.Logging.Level)

	// Start background services
	go a.WebSocketHub.Run()
	go a.UpdateChecker.Start()

	// Start server
	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	a.Logger.Info("Application started successfully",
		"address", fmt.Sprintf("http://localhost:%d", a.Config.Server.Port),
		"license_status", "checking...")

	// Open browser after a short delay to ensure server is ready
	go func() {
		time.Sleep(1 * time.Second)
		if err := openBrowser(fmt.Sprintf("http://localhost:%d", a.Config.Server.Port)); err != nil {
			a.Logger.Warn("Failed to open browser", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the application
func (a *Application) Stop(ctx context.Context) error {
	a.Logger.Info("Shutting down application")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, a.Config.Server.ShutdownTimeout)
	defer cancel()

	// Stop server
	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Stop background services
	a.UpdateChecker.Stop()
	a.WebSocketHub.Stop()

	// Cancel running pipelines
	if err := a.PipelineService.CancelAll(); err != nil {
		a.Logger.Error("Error cancelling pipelines", "error", err)
	}

	a.Logger.Info("Application shutdown complete")
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
	if err := a.Start(ctx); err != nil {
		return err
	}

	// Wait for interrupt
	<-sigChan
	a.Logger.Info("Received interrupt signal")

	// Graceful shutdown
	return a.Stop(ctx)
}

// handleWebSocket handles WebSocket connections
func (a *Application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract any available request ID (might not have middleware)
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("ws-%d", time.Now().UnixNano())
	}
	
	// Structured logging per CLAUDE.md
	ctx := context.WithValue(r.Context(), "request_id", reqID)
	a.SLogger.InfoContext(ctx, "WebSocket upgrade request",
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("origin", r.Header.Get("Origin")))
	
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Implement proper origin validation per CLAUDE.md security
			origin := r.Header.Get("Origin")
			a.SLogger.DebugContext(ctx, "WebSocket origin check",
				slog.String("origin", origin))
			return true // Development mode
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		a.SLogger.ErrorContext(ctx, "WebSocket upgrade failed",
			slog.String("error", err.Error()),
			slog.String("details", fmt.Sprintf("%+v", err)))
		return
	}

	// Create a new client and register with hub
	client := ws.NewClient(a.WebSocketHub, conn)
	a.WebSocketHub.Register(client)
	
	a.SLogger.InfoContext(ctx, "WebSocket client connected",
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("request_id", reqID))

	// Start client goroutines with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.SLogger.ErrorContext(ctx, "WebSocket write pump panic",
					slog.Any("panic", r),
					slog.String("request_id", reqID))
			}
		}()
		client.WritePump()
	}()
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.SLogger.ErrorContext(ctx, "WebSocket read pump panic",
					slog.Any("panic", r),
					slog.String("request_id", reqID))
			}
		}()
		client.ReadPump()
	}()
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}