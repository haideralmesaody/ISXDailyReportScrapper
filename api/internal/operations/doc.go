// Package operation provides a flexible and extensible operation execution framework
// for orchestrating multi-Step data processing workflows.
//
// The operation package is designed to replace the monolithic handleScrape function
// with a modular, maintainable architecture that supports:
//
//   - Step-based execution with dependency management
//   - Configurable retry logic and error handling
//   - Real-time progress tracking via WebSocket
//   - Parallel and sequential execution modes
//   - State persistence and recovery
//   - Extensible Step implementations
//
// Core Components:
//
// Manager: The main orchestrator that manages operation execution, Step registration,
// and state management. It coordinates the execution of steps based on their
// dependencies and configured execution mode.
//
// Step: An interface that defines a single unit of work in the operation. steps
// can have dependencies on other steps and are executed in the correct order.
//
// Registry: Manages the registration and retrieval of steps. It validates
// dependencies and provides topological sorting for execution order.
//
// State: Tracks the runtime state of both the operation and individual steps,
// including progress, errors, and metadata.
//
// Config: Provides configuration options for operation execution, including
// timeouts, retry policies, and execution modes.
//
// Example usage:
//
//	// Create a new operation manager
//	manager := operation.NewManager(wsHub)
//
//	// Register steps
//	manager.RegisterStage(NewScrapingStage())
//	manager.RegisterStage(NewProcessingStage())
//	manager.RegisterStage(NewIndicesStage())
//	manager.RegisterStage(NewLiquidityStage())
//	manager.RegisterStage(NewAnalysisStage())
//	manager.RegisterStage(NewIndicatorsStage())
//
//	// Configure operation
//	config := operation.NewConfigBuilder().
//		WithExecutionMode(operation.ExecutionModeSequential).
//		WithRetryConfig(operation.DefaultRetryConfig()).
//		Build()
//	manager.SetConfig(config)
//
//	// Execute operation
//	req := operation.OperationRequest{
//		Mode:     "initial",
//		FromDate: "2024-01-01",
//		ToDate:   "2024-01-31",
//	}
//	resp, err := manager.Execute(ctx, req)
//
// The operation package integrates with the existing WebSocket infrastructure
// to provide real-time updates on operation progress and Step status changes.
package operations
