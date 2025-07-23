// Package pipeline provides a flexible and extensible pipeline execution framework
// for orchestrating multi-stage data processing workflows.
//
// The pipeline package is designed to replace the monolithic handleScrape function
// with a modular, maintainable architecture that supports:
//
//   - Stage-based execution with dependency management
//   - Configurable retry logic and error handling
//   - Real-time progress tracking via WebSocket
//   - Parallel and sequential execution modes
//   - State persistence and recovery
//   - Extensible stage implementations
//
// Core Components:
//
// Manager: The main orchestrator that manages pipeline execution, stage registration,
// and state management. It coordinates the execution of stages based on their
// dependencies and configured execution mode.
//
// Stage: An interface that defines a single unit of work in the pipeline. Stages
// can have dependencies on other stages and are executed in the correct order.
//
// Registry: Manages the registration and retrieval of stages. It validates
// dependencies and provides topological sorting for execution order.
//
// State: Tracks the runtime state of both the pipeline and individual stages,
// including progress, errors, and metadata.
//
// Config: Provides configuration options for pipeline execution, including
// timeouts, retry policies, and execution modes.
//
// Example usage:
//
//	// Create a new pipeline manager
//	manager := pipeline.NewManager(wsHub, logger)
//
//	// Register stages
//	manager.RegisterStage(NewScrapingStage())
//	manager.RegisterStage(NewProcessingStage())
//	manager.RegisterStage(NewIndicesStage())
//	manager.RegisterStage(NewAnalysisStage())
//
//	// Configure pipeline
//	config := pipeline.NewConfigBuilder().
//		WithExecutionMode(pipeline.ExecutionModeSequential).
//		WithRetryConfig(pipeline.DefaultRetryConfig()).
//		Build()
//	manager.SetConfig(config)
//
//	// Execute pipeline
//	req := pipeline.PipelineRequest{
//		Mode:     "initial",
//		FromDate: "2024-01-01",
//		ToDate:   "2024-01-31",
//	}
//	resp, err := manager.Execute(ctx, req)
//
// The pipeline package integrates with the existing WebSocket infrastructure
// to provide real-time updates on pipeline progress and stage status changes.
package pipeline