package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

)

func TestNewErrorMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new error middleware",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := newTestLogger(t)
			errorHandler := NewErrorHandler(logger, false)
			
			middleware := NewErrorMiddleware(errorHandler, logger)
			
			assert.NotNil(t, middleware)
			assert.Equal(t, errorHandler, middleware.handler)
			assert.NotNil(t, middleware.logger)
		})
	}
}

func TestErrorMiddleware_Handler(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		requestBody    string
		requestPath    string
		requestMethod  string
		wantStatus     int
		shouldPanic    bool
		wantLogLevel   slog.Level
		checkDuration  bool
	}{
		{
			name: "successful request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			requestPath:   "/api/v1/test",
			requestMethod: "GET",
			wantStatus:    http.StatusOK,
			wantLogLevel:  slog.LevelInfo,
			checkDuration: true,
		},
		{
			name: "client error request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request"))
			},
			requestPath:   "/api/v1/test",
			requestMethod: "POST",
			wantStatus:    http.StatusBadRequest,
			wantLogLevel:  slog.LevelWarn,
		},
		{
			name: "server error request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal error"))
			},
			requestPath:   "/api/v1/test",
			requestMethod: "PUT",
			wantStatus:    http.StatusInternalServerError,
			wantLogLevel:  slog.LevelError,
		},
		{
			name: "request with body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("validation error"))
			},
			requestBody:   `{"email": "invalid", "password": "short"}`,
			requestPath:   "/api/v1/users",
			requestMethod: "POST",
			wantStatus:    http.StatusBadRequest,
			wantLogLevel:  slog.LevelWarn,
		},
		{
			name: "request that panics",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			},
			requestPath:   "/api/v1/test",
			requestMethod: "GET",
			wantStatus:    http.StatusInternalServerError,
			shouldPanic:   true,
			wantLogLevel:  slog.LevelError,
		},
		{
			name: "request with query parameters",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad query"))
			},
			requestPath:   "/api/v1/test?limit=10&offset=0",
			requestMethod: "GET",
			wantStatus:    http.StatusBadRequest,
			wantLogLevel:  slog.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logHandler := newTestLogger(t)
			errorHandler := NewErrorHandler(logger, true)
			errorMiddleware := NewErrorMiddleware(errorHandler, logger)
			
			middleware := errorMiddleware.Handler(tt.handler)
			
			var body io.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			}
			
			w := httptest.NewRecorder() 
			r := httptest.NewRequest(tt.requestMethod, tt.requestPath, body)
			if tt.requestBody != "" {
				r.Header.Set("Content-Type", "application/json")
			}
			
			// Add request ID for tracing using Chi middleware key
			ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id")
			r = r.WithContext(ctx)

			// Set User-Agent header
			r.Header.Set("User-Agent", "test-client/1.0")

			middleware.ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			// Check that request was logged
			assert.True(t, logHandler.ContainsMessage("http request"))
			
			// Check log level based on status
			records := logHandler.GetRecordsByLevel(tt.wantLogLevel)
			assert.Greater(t, len(records), 0, "Expected log record at level %s", tt.wantLogLevel)

			// Verify log attributes
			logRecords := logHandler.GetRecords()
			require.Greater(t, len(logRecords), 0, "Should have at least one log record")
			
			var httpLogRecord *LogRecord
			for _, record := range logRecords {
				if strings.Contains(record.Message, "http request") {
					httpLogRecord = &record
					break
				}
			}
			require.NotNil(t, httpLogRecord, "Should have HTTP request log record")

			// Check log attributes
			assert.Equal(t, tt.requestMethod, httpLogRecord.Attrs["method"])
			
			if strings.Contains(tt.requestPath, "?") {
				pathParts := strings.Split(tt.requestPath, "?")
				assert.Equal(t, pathParts[0], httpLogRecord.Attrs["path"])
				assert.Equal(t, pathParts[1], httpLogRecord.Attrs["query"])
			} else {
				assert.Equal(t, tt.requestPath, httpLogRecord.Attrs["path"])
			}
			
			// Status is logged as int64, not int
			assert.Equal(t, int64(tt.wantStatus), httpLogRecord.Attrs["status"])
			assert.Equal(t, "test-request-id", httpLogRecord.Attrs["request_id"])
			assert.Equal(t, "test-client/1.0", httpLogRecord.Attrs["user_agent"])
			
			if tt.checkDuration {
				assert.Contains(t, httpLogRecord.Attrs, "duration")
				duration, ok := httpLogRecord.Attrs["duration"].(time.Duration)
				assert.True(t, ok, "Duration should be time.Duration type")
				assert.GreaterOrEqual(t, duration, time.Duration(0))
			}

			// For error responses with body, check that request body was logged
			if tt.wantStatus >= 400 && tt.requestBody != "" {
				assert.Contains(t, httpLogRecord.Attrs, "request_body")
			}

			if tt.shouldPanic {
				// Should also have a panic recovery log
				assert.True(t, logHandler.ContainsMessage("panic recovered"))
			}
		})
	}
}

func TestErrorMiddleware_RequestBodyCapture(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		contentSize int64
		wantCaptured bool
		expectTruncation bool
	}{
		{
			name:         "small JSON body",
			requestBody:  `{"email":"test@example.com","name":"John Doe"}`, // Compact JSON (no spaces)
			contentSize:  0, // Will be calculated
			wantCaptured: true,
		},
		{
			name:         "empty body",
			requestBody:  "",
			contentSize:  0,
			wantCaptured: false,
		},
		{
			name:        "large body exceeds limit",
			requestBody: strings.Repeat("a", 1024*1024+1), // > 1MB
			contentSize: 1024*1024 + 1,
			wantCaptured: false,
		},
		{
			name:             "body requiring truncation",
			requestBody:      strings.Repeat("a", 600), // > 500 chars for truncation
			contentSize:      0, // Will be calculated
			wantCaptured:     true,
			expectTruncation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logHandler := newTestLogger(t)
			errorHandler := NewErrorHandler(logger, false)
			errorMiddleware := NewErrorMiddleware(errorHandler, logger)

			handler := errorMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Return error status to trigger request body logging
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("error"))
			}))

			var body io.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/test", body)
			if tt.contentSize > 0 {
				r.ContentLength = tt.contentSize
			} else if tt.requestBody != "" {
				r.ContentLength = int64(len(tt.requestBody))
			}

			// Add request ID for tracing using Chi middleware key
			ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id")
			r = r.WithContext(ctx)

			handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			// Check if request body was captured in logs
			logRecords := logHandler.GetRecords()
			var httpLogRecord *LogRecord
			for _, record := range logRecords {
				if strings.Contains(record.Message, "http request") {
					httpLogRecord = &record
					break
				}
			}

			if tt.wantCaptured {
				require.NotNil(t, httpLogRecord)
				assert.Contains(t, httpLogRecord.Attrs, "request_body")
				
				loggedBody := httpLogRecord.Attrs["request_body"].(string)
				if tt.expectTruncation {
					assert.True(t, strings.HasSuffix(loggedBody, "..."))
					assert.Equal(t, 503, len(loggedBody)) // 500 chars + "..."
				} else {
					assert.Equal(t, tt.requestBody, loggedBody)
				}
			} else {
				if httpLogRecord != nil {
					assert.NotContains(t, httpLogRecord.Attrs, "request_body")
				}
			}
		})
	}
}

func TestSanitizeRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "sanitize password field",
			input:    `{"username": "john", "password": "secret123"}`,
			expected: `{"password":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "sanitize multiple sensitive fields",
			input:    `{"email": "test@example.com", "password": "secret", "api_key": "abc123", "name": "John"}`,
			expected: `{"api_key":"[REDACTED]","email":"test@example.com","name":"John","password":"[REDACTED]"}`,
		},
		{
			name:     "sanitize license_key field",
			input:    `{"license_key": "ISX1Y-12345-67890", "user": "john"}`,
			expected: `{"license_key":"[REDACTED]","user":"john"}`,
		},
		{
			name:     "sanitize token field",
			input:    `{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "data": "value"}`,
			expected: `{"data":"value","token":"[REDACTED]"}`,
		},
		{
			name:     "no sensitive fields",
			input:    `{"name": "John", "email": "john@example.com", "age": 30}`,
			expected: `{"age":30,"email":"john@example.com","name":"John"}`,
		},
		{
			name:     "invalid JSON",
			input:    `not a json string`,
			expected: `not a json string`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "sanitize credit_card field",
			input:    `{"credit_card": "1234-5678-9012-3456", "amount": 100}`,
			expected: `{"amount":100,"credit_card":"[REDACTED]"}`,
		},
		{
			name:     "sanitize ssn field",
			input:    `{"ssn": "123-45-6789", "name": "John"}`,
			expected: `{"name":"John","ssn":"[REDACTED]"}`,
		},
		{
			name:     "sanitize secret field",
			input:    `{"secret": "top-secret-value", "public": "public-value"}`,
			expected: `{"public":"public-value","secret":"[REDACTED]"}`,
		},
		{
			name:     "sanitize apiKey camelCase field",
			input:    `{"apiKey": "secret-api-key", "userId": 123}`,
			expected: `{"apiKey":"[REDACTED]","userId":123}`,
		},
		{
			name:     "sanitize licenseKey camelCase field",
			input:    `{"licenseKey": "ISX1Y-ABCDE-FGHIJ", "version": "1.0"}`,
			expected: `{"licenseKey":"[REDACTED]","version":"1.0"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeRequestBody(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		shouldPanic bool
		wantStatus  int
	}{
		{
			name: "normal request without panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			shouldPanic: false,
			wantStatus:  http.StatusOK,
		},
		{
			name: "request that panics with string",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			},
			shouldPanic: true,
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name: "request that panics with error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(assert.AnError)
			},
			shouldPanic: true,
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name: "request that panics with integer",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(42)
			},
			shouldPanic: true,
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logHandler := newTestLogger(t)
			errorHandler := NewErrorHandler(logger, true)
			
			middleware := RecoveryMiddleware(errorHandler)(tt.handler)
			
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			// Add request ID for tracing using Chi middleware key
			ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id")
			r = r.WithContext(ctx)

			middleware.ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.shouldPanic {
				// Should have logged the panic
				assert.True(t, logHandler.ContainsMessage("panic recovered"))
				
				// Response should be JSON problem details
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				
				var problem ProblemDetails
				err := json.NewDecoder(w.Body).Decode(&problem)
				require.NoError(t, err)
				
				assert.Equal(t, TypeInternal, problem.Type)
				assert.Equal(t, "Internal Server Error", problem.Title)
				assert.Equal(t, http.StatusInternalServerError, problem.Status)
				assert.Equal(t, "An unexpected error occurred", problem.Detail)
				// trace_id might be nil if not set by chi middleware
			if traceID, exists := problem.Extensions["trace_id"]; exists {
				assert.NotNil(t, traceID)
			}
			}
		})
	}
}

func TestErrorMiddleware_LargeRequestBodyHandling(t *testing.T) {
	t.Run("large request body not captured", func(t *testing.T) {
		logger, logHandler := newTestLogger(t)
		errorHandler := NewErrorHandler(logger, false)
		errorMiddleware := NewErrorMiddleware(errorHandler, logger)

		// Create a large body (> 1MB)
		largeBody := strings.Repeat("a", 1024*1024+1)
		
		handler := errorMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request body can still be read by the handler
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, largeBody, string(body))
			
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error"))
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
		r.ContentLength = int64(len(largeBody))
		
		// Add request ID for tracing
		ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id")
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Large body should not be captured in logs
		logRecords := logHandler.GetRecords()
		for _, record := range logRecords {
			if strings.Contains(record.Message, "http request") {
				assert.NotContains(t, record.Attrs, "request_body")
				break
			}
		}
	})
}

func TestErrorMiddleware_NilRequestBody(t *testing.T) {
	t.Run("nil request body", func(t *testing.T) {
		logger, logHandler := newTestLogger(t)
		errorHandler := NewErrorHandler(logger, false)
		errorMiddleware := NewErrorMiddleware(errorHandler, logger)

		handler := errorMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error"))
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil) // nil body
		
		// Add request ID for tracing
		ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id")
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Should not try to capture nil request body
		logRecords := logHandler.GetRecords()
		for _, record := range logRecords {
			if strings.Contains(record.Message, "http request") {
				assert.NotContains(t, record.Attrs, "request_body")
				break
			}
		}
	})
}

func TestErrorMiddleware_LoggingAttributes(t *testing.T) {
	t.Run("comprehensive logging attributes", func(t *testing.T) {
		logger, logHandler := newTestLogger(t)
		errorHandler := NewErrorHandler(logger, false)
		errorMiddleware := NewErrorMiddleware(errorHandler, logger)

		handler := errorMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write some data to test bytes written
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/v1/users?limit=10&offset=20", strings.NewReader(`{"name": "test"}`))
		r.RemoteAddr = "192.168.1.1:12345"
		r.Header.Set("User-Agent", "TestClient/1.0")
		
		// Add request ID for tracing
		ctx := context.WithValue(r.Context(), "RequestID", "test-req-123")
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify all expected log attributes
		logRecords := logHandler.GetRecords()
		var httpLogRecord *LogRecord
		for _, record := range logRecords {
			if strings.Contains(record.Message, "http request") {
				httpLogRecord = &record
				break
			}
		}

		require.NotNil(t, httpLogRecord)
		
		assert.Equal(t, "POST", httpLogRecord.Attrs["method"])
		assert.Equal(t, "/api/v1/users", httpLogRecord.Attrs["path"])
		assert.Equal(t, "limit=10&offset=20", httpLogRecord.Attrs["query"])
		// Status might be int or int64
		status := httpLogRecord.Attrs["status"]
		switch v := status.(type) {
		case int:
			assert.Equal(t, http.StatusOK, v)
		case int64:
			assert.Equal(t, int64(http.StatusOK), v)
		default:
			t.Errorf("Unexpected type for status: %T", v)
		}
		assert.Equal(t, "192.168.1.1:12345", httpLogRecord.Attrs["remote_addr"])
		assert.Equal(t, "TestClient/1.0", httpLogRecord.Attrs["user_agent"])
		// Request ID might be empty if not set by chi middleware
		assert.Contains(t, httpLogRecord.Attrs, "request_id")
		
		assert.Contains(t, httpLogRecord.Attrs, "duration")
		assert.Contains(t, httpLogRecord.Attrs, "bytes")
		
		// Chi middleware might return int64 for bytes written
		bytesWritten := httpLogRecord.Attrs["bytes"]
		switch v := bytesWritten.(type) {
		case int:
			assert.Equal(t, len("Hello, World!"), v)
		case int64:
			assert.Equal(t, int64(len("Hello, World!")), v)
		default:
			t.Errorf("Unexpected type for bytes: %T", v)
		}
	})
}

func TestErrorMiddleware_ConcurrentRequests(t *testing.T) {
	t.Run("concurrent request handling", func(t *testing.T) {
		logger, _ := newTestLogger(t)
		errorHandler := NewErrorHandler(logger, false)
		errorMiddleware := NewErrorMiddleware(errorHandler, logger)

		handler := errorMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		const numRequests = 10
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(r.Context(), "RequestID", fmt.Sprintf("req-%d", i))
				r = r.WithContext(ctx)

				handler.ServeHTTP(w, r)
				results <- w.Code
			}(i)
		}

		// Collect all results
		for i := 0; i < numRequests; i++ {
			select {
			case statusCode := <-results:
				assert.Equal(t, http.StatusOK, statusCode)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}
	})
}

func TestErrorMiddleware_Integration(t *testing.T) {
	t.Run("integration with chi middleware", func(t *testing.T) {
		logger, logHandler := newTestLogger(t)
		errorHandler := NewErrorHandler(logger, false)
		errorMiddleware := NewErrorMiddleware(errorHandler, logger)

		// Stack middleware like in real application
		handler := chimiddleware.RequestID(
			errorMiddleware.Handler(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot) // Unusual status for testing
					w.Write([]byte("I'm a teapot"))
				}),
			),
		)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusTeapot, w.Code)
		assert.Equal(t, "I'm a teapot", w.Body.String())

		// Verify request was logged with proper request ID
		assert.True(t, logHandler.ContainsMessage("http request"))
		
		logRecords := logHandler.GetRecords()
		for _, record := range logRecords {
			if strings.Contains(record.Message, "http request") {
				// Should have request ID from chi middleware
				assert.Contains(t, record.Attrs, "request_id")
				requestID := record.Attrs["request_id"].(string)
				assert.NotEmpty(t, requestID)
				break
			}
		}
	})
}