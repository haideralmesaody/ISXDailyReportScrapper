// Package services implements the business logic layer of the ISX application.
// It provides a clean separation between HTTP handlers and data access, ensuring
// that business rules are centralized and testable.
//
// # Architecture
//
// Services follow these architectural principles:
//
//	1. Interface-driven design for testability
//	2. Context propagation for cancellation and tracing
//	3. Dependency injection for loose coupling
//	4. Transaction management for data consistency
//	5. Domain-focused methods that encapsulate business rules
//
// # Service Layer Responsibilities
//
// The service layer is responsible for:
//
//	- Business logic and validation
//	- Transaction coordination
//	- Cross-cutting concerns (logging, metrics)
//	- Error handling and transformation
//	- Caching strategies
//	- External API integration
//
// # Common Service Pattern
//
// Services typically follow this structure:
//
//	type ServiceName struct {
//	    repo   Repository
//	    logger *slog.Logger
//	    cache  Cache
//	}
//	
//	func NewServiceName(repo Repository, logger *slog.Logger) *ServiceName {
//	    return &ServiceName{
//	        repo:   repo,
//	        logger: logger,
//	    }
//	}
//	
//	func (s *ServiceName) BusinessOperation(ctx context.Context, input Input) (*Output, error) {
//	    // Validate input
//	    if err := input.Validate(); err != nil {
//	        return nil, fmt.Errorf("validation failed: %w", err)
//	    }
//	    
//	    // Execute business logic
//	    result, err := s.repo.Operation(ctx, input)
//	    if err != nil {
//	        s.logger.ErrorContext(ctx, "operation failed",
//	            "error", err,
//	            "input", input,
//	        )
//	        return nil, fmt.Errorf("operation failed: %w", err)
//	    }
//	    
//	    return result, nil
//	}
//
// # Available Services
//
// The package provides these core services:
//
//	- DataService: Handles report data operations
//	- LicenseService: Manages license validation and activation
//	- HealthService: Provides system health checks
//	- OperationsService: Orchestrates multi-step data operations
//
// # Error Handling
//
// Services return domain-specific errors that handlers can transform:
//
//	- Validation errors for invalid input
//	- Not found errors for missing resources
//	- Conflict errors for duplicate operations
//	- Internal errors for unexpected failures
//
// # Testing
//
// Services are tested by mocking dependencies:
//
//	mockRepo := mocks.NewRepository(t)
//	service := NewService(mockRepo, logger)
//	
//	mockRepo.On("Find", id).Return(entity, nil)
//	result, err := service.Get(ctx, id)
//
// # Performance Considerations
//
// Services implement various performance optimizations:
//
//	- Connection pooling for database access
//	- Caching for frequently accessed data
//	- Batch operations where applicable
//	- Concurrent processing with proper synchronization
package services