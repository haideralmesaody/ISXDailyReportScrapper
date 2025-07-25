package handlers

import (
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

// MockDataService is a mock implementation of DataService
type MockDataService struct {
	mock.Mock
}

func (m *MockDataService) GetReports() ([]map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDataService) GetTickers() (interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockDataService) GetIndices() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataService) GetFiles() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataService) GetMarketMovers(period, limit, minVolume string) (map[string]interface{}, error) {
	args := m.Called(period, limit, minVolume)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataService) GetTickerChart(ticker string) (map[string]interface{}, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDataService) DownloadFile(w http.ResponseWriter, r *http.Request, fileType, filename string) error {
	args := m.Called(w, r, fileType, filename)
	return args.Error(0)
}

func TestDataHandler_GetReports(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDataService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful get reports",
			setupMock: func(m *MockDataService) {
				reports := []map[string]interface{}{
					{"id": 1, "name": "Report 1"},
					{"id": 2, "name": "Report 2"},
				}
				m.On("GetReports").Return(reports, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success","data":[{"id":1,"name":"Report 1"},{"id":2,"name":"Report 2"}],"count":2}`,
		},
		{
			name: "no reports found",
			setupMock: func(m *MockDataService) {
				m.On("GetReports").Return(nil, services.ErrNoReportsFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NO_REPORTS_FOUND"`,
		},
		{
			name: "internal error",
			setupMock: func(m *MockDataService) {
				m.On("GetReports").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"Internal Server Error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDataService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(mockService, logger, errorHandler)

			// Create request
			req := httptest.NewRequest("GET", "/api/data/reports", nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.GetReports(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataHandler_GetTickers(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDataService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful get tickers",
			setupMock: func(m *MockDataService) {
				tickers := []interface{}{
					map[string]interface{}{"symbol": "AAPL", "price": 150.0},
					map[string]interface{}{"symbol": "GOOGL", "price": 2800.0},
				}
				m.On("GetTickers").Return(tickers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success","data":[{"symbol":"AAPL","price":150},{"symbol":"GOOGL","price":2800}],"count":2}`,
		},
		{
			name: "no tickers found",
			setupMock: func(m *MockDataService) {
				m.On("GetTickers").Return(nil, services.ErrNoTickersFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NO_TICKERS_FOUND"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDataService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(mockService, logger, errorHandler)

			// Create request
			req := httptest.NewRequest("GET", "/api/data/tickers", nil)
			rec := httptest.NewRecorder()

			// Execute
			handler.GetTickers(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataHandler_GetMarketMovers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMock      func(*MockDataService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful with default params",
			queryParams: map[string]string{},
			setupMock: func(m *MockDataService) {
				movers := map[string]interface{}{
					"gainers": []interface{}{
						map[string]interface{}{"symbol": "AAPL", "change": 5.2},
					},
					"losers":     []interface{}{},
					"mostActive": []interface{}{},
				}
				m.On("GetMarketMovers", "daily", "10", "0").Return(movers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name: "invalid period",
			queryParams: map[string]string{
				"period": "hourly",
			},
			setupMock:      func(m *MockDataService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid period. Must be one of: daily, weekly, monthly"`,
		},
		{
			name: "invalid limit",
			queryParams: map[string]string{
				"limit": "abc",
			},
			setupMock:      func(m *MockDataService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Limit must be a number between 1 and 100"`,
		},
		{
			name: "limit too high",
			queryParams: map[string]string{
				"limit": "200",
			},
			setupMock:      func(m *MockDataService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Limit must be a number between 1 and 100"`,
		},
		{
			name: "no market movers found",
			queryParams: map[string]string{
				"period": "weekly",
				"limit":  "20",
			},
			setupMock: func(m *MockDataService) {
				m.On("GetMarketMovers", "weekly", "20", "0").Return(nil, services.ErrNoMarketMovers)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NO_MARKET_MOVERS"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDataService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(mockService, logger, errorHandler)

			// Create request with query params
			req := httptest.NewRequest("GET", "/api/data/market-movers", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()

			// Execute
			handler.GetMarketMovers(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataHandler_GetTickerChart(t *testing.T) {
	tests := []struct {
		name           string
		ticker         string
		setupMock      func(*MockDataService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful get chart",
			ticker: "AAPL",
			setupMock: func(m *MockDataService) {
				chart := map[string]interface{}{
					"data": []float64{150, 152, 148, 155},
				}
				m.On("GetTickerChart", "AAPL").Return(chart, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "ticker not found",
			ticker: "INVALID",
			setupMock: func(m *MockDataService) {
				m.On("GetTickerChart", "INVALID").Return(nil, services.ErrTickerNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"TICKER_NOT_FOUND"`,
		},
		{
			name:   "no chart data",
			ticker: "XYZ",
			setupMock: func(m *MockDataService) {
				m.On("GetTickerChart", "XYZ").Return(nil, services.ErrNoChartData)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"NO_CHART_DATA"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockDataService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(mockService, logger, errorHandler)

			// Create router with context
			r := chi.NewRouter()
			r.Route("/ticker/{ticker}", func(r chi.Router) {
				r.Get("/chart", handler.GetTickerChart)
			})

			// Create request
			req := httptest.NewRequest("GET", "/ticker/"+tt.ticker+"/chart", nil)
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

func TestDataHandler_TickerCtx(t *testing.T) {
	tests := []struct {
		name           string
		ticker         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid ticker",
			ticker:         "AAPL",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "empty ticker",
			ticker:         "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Ticker symbol is required",
		},
		{
			name:           "ticker too short",
			ticker:         "A",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid ticker symbol format",
		},
		{
			name:           "ticker too long",
			ticker:         "VERYLONGTICKER",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid ticker symbol format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(&services.DataService{}, logger, errorHandler)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Create router with middleware
			r := chi.NewRouter()
			r.Route("/ticker/{ticker}", func(r chi.Router) {
				r.Use(handler.TickerCtx)
				r.Get("/", testHandler)
			})

			// Create request
			req := httptest.NewRequest("GET", "/ticker/"+tt.ticker+"/", nil)
			rec := httptest.NewRecorder()

			// Execute
			r.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestDataHandler_DownloadCtx(t *testing.T) {
	tests := []struct {
		name           string
		fileType       string
		filename       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid download",
			fileType:       "report",
			filename:       "daily-report.pdf",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "invalid file type",
			fileType:       "invalid",
			filename:       "test.txt",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid file type: invalid",
		},
		{
			name:           "empty filename",
			fileType:       "excel",
			filename:       "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Filename is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			errorHandler := apierrors.NewErrorHandler(logger, false)
			handler := NewDataHandler(&services.DataService{}, logger, errorHandler)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Create router with middleware
			r := chi.NewRouter()
			r.Route("/download/{type}/{filename}", func(r chi.Router) {
				r.Use(handler.DownloadCtx)
				r.Get("/", testHandler)
			})

			// Create request
			path := "/download/" + tt.fileType + "/" + tt.filename
			if tt.filename == "" {
				path = "/download/" + tt.fileType + "/"
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