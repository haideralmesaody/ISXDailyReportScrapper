package app

import (
	"context"
	"fmt"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/operations"
	"github.com/isxcli/isxcli/internal/services"
	"github.com/isxcli/isxcli/internal/strategy"
	"github.com/isxcli/isxcli/internal/strategy/strategies"
	ws "github.com/isxcli/isxcli/internal/websocket"

	"log/slog"
)

// initializeServices initializes all application services
func (a *Application) initializeServices() error {
	// License manager is already initialized in ValidateLicenseOnStartup
	// Skip re-initialization to avoid duplicate operations
	if a.LicenseManager == nil {
		return fmt.Errorf("license manager not initialized - this should never happen")
	}

	// Note: License expiry warnings are already shown in ValidateLicenseOnStartup if needed

	// Initialize WebSocket hub with timing diagnostics
	a.Logger.Info("🔧 [WS-SETUP] Initializing WebSocket hub")
	hubStart := time.Now()

	hub := ws.NewHub(a.Logger)
	hubInitTime := time.Since(hubStart)

	a.Logger.Info("🔧 [WS-SETUP] WebSocket hub created",
		slog.Duration("hub_init_time", hubInitTime),
		slog.Time("hub_init_at", time.Now()))

	// Start the hub with timing
	hubStartStart := time.Now()
	hub.Start()
	hubStartTime := time.Since(hubStartStart)

	// Verify hub is running
	if !hub.IsRunning() {
		return fmt.Errorf("WebSocket hub failed to start")
	}

	a.Logger.Info("🔧 [WS-SETUP] WebSocket hub started successfully",
		slog.Duration("hub_start_time", hubStartTime),
		slog.Time("hub_started_at", time.Now()),
		slog.Bool("hub_running", hub.IsRunning()))

	a.WebSocketHub = hub

	// Initialize operations components in correct order
	// 1. Create registry first
	registry := operations.NewRegistry()

	// 2. Create manager with registry
	opsManager := operations.NewManager(hub, registry, &operations.Config{
		ExecutionMode: operations.ExecutionModeSequential,
	})

	// 3. Get application paths
	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get application paths: %w", err)
	}

	// 6. Create operation service
	OperationService, err := services.NewOperationService(
		context.Background(),
		opsManager,
		a.LicenseManager,
		a.Logger,
		paths,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize operation service: %w", err)
	}
	a.OperationService = OperationService

	// 7. Initialize job queue for async operations
	jobStore := operations.NewMemoryJobStore()
	jobQueue := operations.NewJobQueue(4, jobStore, opsManager, a.Logger) // Use opsManager directly
	a.JobQueue = jobQueue

	// CRITICAL: Connect job queue to operation service for queue-driven standardization
	OperationService.SetJobQueue(jobQueue)

	// CRITICAL: Connect jobqueue back to manager for unified execution path
	// This ensures manager.executeStageViaJobQueue() can route execution through jobqueue
	opsManager.SetJobQueue(a.JobQueue)

	a.Logger.Info("jobqueue_connected_to_manager",
		"component", "app",
		"jobqueue_workers", 4,
		"unified_execution_enabled", true)

	// Initialize data service with injected logger
	dataService, err := services.NewDataServiceWithLogger(a.Config, a.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize data service: %w", err)
	}
	a.DataService = dataService

	// Initialize health service with injected logger
	healthService := services.NewHealthServiceWithBuildInfo(
		VERSION,
		REPO_URL,
		BuildTime,
		BuildID,
		a.Config.Paths,
		a.LicenseManager,
		a.OperationService.GetManager(),
		a.WebSocketHub,
		a.Logger,
	)
	a.HealthService = healthService

	// Initialize license service
	licenseService := services.NewLicenseService(a.LicenseManager, a.Logger)

	// Get paths for liquidity service (reuse existing paths)
	// paths already available from earlier initialization

	// Initialize liquidity service
	liquidityService := services.NewLiquidityService(paths.ReportsDir, a.Logger)

	// Initialize strategy framework
	strategyManager := strategy.NewManager()

	// RSI(14) Mean Reversion (EOD) - fixed parameters
	if err := strategyManager.RegisterStrategy(strategies.NewRSI14MeanReversionEODStrategy(a.Logger)); err != nil {
		return fmt.Errorf("register RSI14 mean reversion strategy: %w", err)
	}

	// Initialize strategy service
	strategyService := services.NewStrategyService(strategyManager, dataService, a.Logger)

	// Create service container
	a.Services = &ServiceContainer{
		License:        a.LicenseManager,
		LicenseService: licenseService,
		Data:           dataService,
		Health:         healthService,
		WebSocket:      hub,
		Liquidity:      liquidityService,
		Strategy:       strategyService,
	}

	return nil
}
