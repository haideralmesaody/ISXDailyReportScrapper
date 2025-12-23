// Package handlers implements HTTP request handlers for the ISX web service.
// It provides a thin layer between HTTP transport and business logic, following
// the clean architecture principle of keeping handlers focused solely on HTTP concerns.
//
// # Architecture Principles
//
// Handlers in this package follow these principles:
//
//	1. Thin handlers - minimal logic, delegate to services
//	2. HTTP-only concerns - request parsing, response formatting
//	3. Error transformation - convert service errors to HTTP responses
//	4. No business logic - all logic belongs in the service layer
//	5. Consistent patterns - standardized request/response handling
//
// # Request Flow
//
// A typical request flows through these layers:
//
//	HTTP Request → Chi Router → Middleware → Handler → Service → Repository
//	                                              ↓
//	HTTP Response ← Handler ← Service Response ←─┘
//
// # Handler Structure
//
// Each handler follows this pattern:
//
//	func (h *Handler) HandleSomething(w http.ResponseWriter, r *http.Request) {
//	    // 1. Parse and validate request
//	    req, err := parseRequest(r)
//	    if err != nil {
//	        render.Render(w, r, errors.NewAPIError(...))
//	        return
//	    }
//	    
//	    // 2. Call service layer
//	    result, err := h.service.DoSomething(r.Context(), req)
//	    if err != nil {
//	        render.Render(w, r, transformError(err))
//	        return
//	    }
//	    
//	    // 3. Format and send response
//	    render.JSON(w, r, formatResponse(result))
//	}
//
// # Error Handling
//
// All errors follow RFC 7807 Problem Details specification:
//
//	{
//	    "type": "validation_error",
//	    "title": "Invalid request data",
//	    "status": 400,
//	    "detail": "The license key format is invalid",
//	    "instance": "/api/v1/license/activate"
//	}
//
// # WebSocket Support
//
// WebSocket handlers use Gorilla WebSocket and follow this pattern:
//
//	- Upgrade HTTP connection to WebSocket
//	- Register client with hub
//	- Handle messages in separate goroutines
//	- Clean up on disconnect
//
// # Middleware Integration
//
// Handlers work with these middleware:
//
//	- RequestID: Adds unique request ID for tracing
//	- Logger: Structured logging with slog
//	- License: Validates license before processing
//	- Recovery: Handles panics gracefully
//	- CORS: Configures cross-origin requests
//
// # Testing
//
// Handlers are tested using httptest:
//
//	- Mock service dependencies
//	- Test various HTTP scenarios
//	- Verify error responses
//	- Check middleware integration
package http