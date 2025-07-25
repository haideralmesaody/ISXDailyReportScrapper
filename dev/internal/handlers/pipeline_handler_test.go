package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"

	apierrors "isxcli/internal/errors"
	"isxcli/internal/services"
)

// MockPipelineService is a mock implementation of PipelineServiceInterface
type MockPipelineService struct {
	mock.Mock
}

func (m *MockPipelineService) StartPipeline(params map[string]interface{}) (string, error) {
	args := m.Called(params)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineService) StopPipeline(pipelineID string) error {
	args := m.Called(pipelineID)
	return args.Error(0)
}

func (m *MockPipelineService) GetStatus(pipelineID string) (map[string]interface{}, error) {
	args := m.Called(pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockPipelineService) ListPipelines() []map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return []map[string]interface{}{}
	}
	return args.Get(0).([]map[string]interface{})
}

func TestPipelineHandler_StartPipeline(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockPipelineService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful start pipeline",
			requestBody: map[string]interface{}{
				"type":     "scraping",
				"priority": "high",
				"parameters": map[string]interface{}{
					"date": "2024-01-20",
				},
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StartPipeline", mock.MatchedBy(func(params map[string]interface{}) bool {
					return params["type"] == "scraping" && params["priority"] == "high"
				})).Return("pipeline-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"pipeline_id":"pipeline-123"`,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMock: func(m *MockPipelineService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid request format"`,
		},
		{
			name: "missing type field",
			requestBody: map[string]interface{}{
				"priority": "high",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StartPipeline", mock.Anything).Return("", services.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"VALIDATION_FAILED"`,
		},
		{
			name: "pipeline already running",
			requestBody: map[string]interface{}{
				"type": "processing",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StartPipeline", mock.Anything).Return("", services.ErrPipelineRunning)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `"PIPELINE_ALREADY_RUNNING"`,
		},
		{
			name: "internal error",
			requestBody: map[string]interface{}{
				"type": "full",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StartPipeline", mock.Anything).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"Internal Server Error"`,
		},
		{
			name: "default priority",
			requestBody: map[string]interface{}{
				"type": "scraping",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StartPipeline", mock.MatchedBy(func(params map[string]interface{}) bool {
					return params["priority"] == "normal"
				})).Return("pipeline-456", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"priority":"normal"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPipelineService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(mockService, logger, errorHandler)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/pipeline/start", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			handler.StartPipeline(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPipelineHandler_StopPipeline(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockPipelineService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful stop pipeline",
			requestBody: map[string]interface{}{
				"pipeline_id": "pipeline-123",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StopPipeline", "pipeline-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"Pipeline stopped successfully"`,
		},
		{
			name: "pipeline not found",
			requestBody: map[string]interface{}{
				"pipeline_id": "invalid-id",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StopPipeline", "invalid-id").Return(services.ErrPipelineNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NOT_FOUND"`,
		},
		{
			name: "pipeline not running",
			requestBody: map[string]interface{}{
				"pipeline_id": "pipeline-456",
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StopPipeline", "pipeline-456").Return(services.ErrPipelineNotRunning)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `"PIPELINE_NOT_RUNNING"`,
		},
		{
			name: "missing pipeline_id",
			requestBody: map[string]interface{}{},
			setupMock: func(m *MockPipelineService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Pipeline ID is required"`,
		},
		{
			name: "force stop",
			requestBody: map[string]interface{}{
				"pipeline_id": "pipeline-789",
				"force": true,
			},
			setupMock: func(m *MockPipelineService) {
				m.On("StopPipeline", "pipeline-789").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"Pipeline stopped successfully"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPipelineService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(mockService, logger, errorHandler)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/pipeline/stop", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			handler.StopPipeline(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPipelineHandler_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockPipelineService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "get specific pipeline status",
			queryParams: "?pipeline_id=pipeline-123",
			setupMock: func(m *MockPipelineService) {
				status := map[string]interface{}{
					"id":       "pipeline-123",
					"status":   "running",
					"progress": 45.5,
				}
				m.On("GetStatus", "pipeline-123").Return(status, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"progress":45.5`,
		},
		{
			name:        "pipeline not found",
			queryParams: "?pipeline_id=invalid",
			setupMock: func(m *MockPipelineService) {
				m.On("GetStatus", "invalid").Return(nil, services.ErrPipelineNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NOT_FOUND"`,
		},
		{
			name:        "overall status - no pipeline ID",
			queryParams: "",
			setupMock: func(m *MockPipelineService) {
				pipelines := []map[string]interface{}{
					{"id": "p1", "status": "running"},
					{"id": "p2", "status": "completed"},
					{"id": "p3", "status": "processing"},
				}
				m.On("ListPipelines").Return(pipelines)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"active_pipelines":2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPipelineService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(mockService, logger, errorHandler)

			// Create request
			req := httptest.NewRequest("GET", "/api/pipeline/status"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.GetStatus(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPipelineHandler_ListPipelines(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockPipelineService)
		expectedStatus int
		expectedBody   string
		expectedCount  int
	}{
		{
			name:        "list all pipelines",
			queryParams: "",
			setupMock: func(m *MockPipelineService) {
				pipelines := []map[string]interface{}{
					{"id": "p1", "status": "running", "type": "scraping"},
					{"id": "p2", "status": "completed", "type": "processing"},
					{"id": "p3", "status": "error", "type": "full"},
				}
				m.On("ListPipelines").Return(pipelines)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":3`,
			expectedCount:  3,
		},
		{
			name:        "filter by status",
			queryParams: "?status=running",
			setupMock: func(m *MockPipelineService) {
				pipelines := []map[string]interface{}{
					{"id": "p1", "status": "running", "type": "scraping"},
					{"id": "p2", "status": "completed", "type": "processing"},
					{"id": "p3", "status": "running", "type": "full"},
				}
				m.On("ListPipelines").Return(pipelines)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":2`,
			expectedCount:  2,
		},
		{
			name:        "filter by type",
			queryParams: "?type=processing",
			setupMock: func(m *MockPipelineService) {
				pipelines := []map[string]interface{}{
					{"id": "p1", "status": "running", "type": "scraping"},
					{"id": "p2", "status": "completed", "type": "processing"},
					{"id": "p3", "status": "running", "type": "processing"},
				}
				m.On("ListPipelines").Return(pipelines)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":2`,
			expectedCount:  2,
		},
		{
			name:        "empty list",
			queryParams: "",
			setupMock: func(m *MockPipelineService) {
				m.On("ListPipelines").Return([]map[string]interface{}{})
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"count":0`,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPipelineService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(mockService, logger, errorHandler)

			// Create request
			req := httptest.NewRequest("GET", "/api/pipeline/pipelines"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.ListPipelines(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			
			// Check the actual count in response
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			assert.Equal(t, float64(tt.expectedCount), response["count"])
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestPipelineHandler_PipelineCtx(t *testing.T) {
	tests := []struct {
		name           string
		pipelineID     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid pipeline ID",
			pipelineID:     "pipeline-123456789",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "empty pipeline ID",
			pipelineID:     "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Pipeline ID is required",
		},
		{
			name:           "short pipeline ID",
			pipelineID:     "short",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid pipeline ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(nil, logger, errorHandler)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Create router with middleware
			r := chi.NewRouter()
			r.Route("/{pipelineID}", func(r chi.Router) {
				r.Use(handler.PipelineCtx)
				r.Get("/", testHandler)
			})

			// Create request
			path := "/" + tt.pipelineID + "/"
			if tt.pipelineID == "" {
				path = "//"
			}
			req := httptest.NewRequest("GET", path, nil)
			rec := httptest.NewRecorder()

			// Execute
			r.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestPipelineHandler_GetPipeline(t *testing.T) {
	tests := []struct {
		name           string
		pipelineID     string
		setupMock      func(*MockPipelineService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful get pipeline",
			pipelineID: "pipeline-123",
			setupMock: func(m *MockPipelineService) {
				status := map[string]interface{}{
					"id":         "pipeline-123",
					"status":     "running",
					"type":       "scraping",
					"started_at": "2024-01-20T10:00:00Z",
				}
				m.On("GetStatus", "pipeline-123").Return(status, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"type":"scraping"`,
		},
		{
			name:       "pipeline not found",
			pipelineID: "nonexistent",
			setupMock: func(m *MockPipelineService) {
				m.On("GetStatus", "nonexistent").Return(nil, services.ErrPipelineNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NOT_FOUND"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockPipelineService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewPipelineHandler(mockService, logger, errorHandler)

			// Create router with context
			r := chi.NewRouter()
			r.Route("/{pipelineID}", func(r chi.Router) {
				r.Get("/", handler.GetPipeline)
			})

			// Create request
			req := httptest.NewRequest("GET", "/"+tt.pipelineID+"/", nil)
			rec := httptest.NewRecorder()

			// Execute
			r.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}