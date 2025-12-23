// Package app provides application initialization and lifecycle management for the ISX system.
// It handles the orchestration of all major components including configuration loading,
// service initialization, and graceful shutdown procedures.
//
// # Architecture
//
// The app package follows a dependency injection pattern where all components
// are wired together at startup. This ensures loose coupling and testability.
//
// # Initialization Flow
//
// The typical initialization sequence:
//
//	1. Load configuration from environment and files
//	2. Initialize logging and observability
//	3. Create data stores and repositories
//	4. Initialize services with their dependencies
//	5. Set up HTTP handlers and middleware
//	6. Configure and start the HTTP server
//	7. Set up graceful shutdown handlers
//
// # Usage
//
// The main entry point is typically:
//
//	app := app.New(config)
//	if err := app.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// # Graceful Shutdown
//
// The package handles SIGINT and SIGTERM signals to ensure:
//
//	- Active requests are completed
//	- WebSocket connections are closed cleanly
//	- Database connections are closed
//	- Temporary files are cleaned up
//	- Final metrics are flushed
//
// # Configuration
//
// The app package relies on the config package for all configuration
// needs. It supports both environment variables and configuration files.
//
// # Error Handling
//
// All initialization errors are returned to the caller for proper
// handling. The app does not call os.Exit() directly, allowing
// the main function to control the exit process.
package app