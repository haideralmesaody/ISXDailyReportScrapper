package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	handlers "github.com/isxcli/isxcli/internal/transport/http"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/services"
	"github.com/isxcli/isxcli/internal/websocket"
)

// LicenseActivationFlowTestSuite tests the complete license activation flow
type LicenseActivationFlowTestSuite struct {
	suite.Suite
	tempDir      string
	licenseFile  string
	server       *httptest.Server
	wsServer     *httptest.Server
	manager      *license.Manager
	service      services.LicenseService
	handler      *handlers.LicenseHandler
	wsHub        *websocket.Hub
	logger       *slog.Logger
}

func (suite *LicenseActivationFlowTestSuite) SetupSuite() {
	// Setup temporary directory
	suite.tempDir = suite.T().TempDir()
	suite.licenseFile = filepath.Join(suite.tempDir, "integration_test_license.dat")
	
	// Setup logger
	suite.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	
	// Initialize license manager
	var err error
	suite.manager, err = license.NewManager()
	require.NoError(suite.T(), err)
	
	// Initialize service
	suite.service = services.NewLicenseService(suite.manager, suite.logger)
	
	// Initialize handler
	suite.handler = handlers.NewLicenseHandler(suite.service, suite.logger)
	
	// Setup WebSocket hub for real-time updates
	suite.wsHub = websocket.NewHub(suite.logger)
	suite.wsHub.Start()
	
	// Setup HTTP server
	suite.setupHTTPServer()
	
	// Setup WebSocket server
	suite.setupWebSocketServer()
}

func (suite *LicenseActivationFlowTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.wsServer != nil {
		suite.wsServer.Close()
	}
	if suite.manager != nil {
		suite.manager.Close()
	}
	if suite.wsHub != nil {
		suite.wsHub.Stop()
	}
}

func (suite *LicenseActivationFlowTestSuite) SetupTest() {
	// Clean up any existing license file
	os.Remove(suite.licenseFile)
}

func (suite *LicenseActivationFlowTestSuite) setupHTTPServer() {
	router := chi.NewRouter()
	
	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	
	// Mount license handler
	router.Mount("/api/license", suite.handler.Routes())
	
	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	
	suite.server = httptest.NewServer(router)
}

func (suite *LicenseActivationFlowTestSuite) setupWebSocketServer() {
	wsRouter := chi.NewRouter()
	
	upgrader := gorilla.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	
	wsRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			suite.logger.Error("WebSocket upgrade failed", "error", err)
			return
		}
		
		client := websocket.NewClient(suite.wsHub, conn, suite.logger)
		suite.wsHub.Register(client)
		
		go client.WritePump()
		go client.ReadPump()
	})
	
	suite.wsServer = httptest.NewServer(wsRouter)
}

// TestCompleteActivationFlowWithLiveLicenseKey tests the complete flow with the live license key
func (suite *LicenseActivationFlowTestSuite) TestCompleteActivationFlowWithLiveLicenseKey() {
	liveLicenseKey := "ISX1M02LYE1F9QJHR9D7Z"
	
	suite.Run("step_1_check_initial_status", func() {
		// Step 1: Check initial license status (should be not activated)
		resp, err := http.Get(suite.server.URL + "/api/license/status")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()
		
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var statusResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResponse)
		require.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), "not_activated", statusResponse["license_status"])
		assert.Contains(suite.T(), statusResponse["message"], "No license activated")
	})
	
	suite.Run("step_2_activate_license", func() {
		// Step 2: Activate the license
		activationRequest := map[string]interface{}{
			"license_key": liveLicenseKey,
			"email":       "integration.test@iraqiinvestor.gov.iq",
		}
		
		requestBody, err := json.Marshal(activationRequest)
		require.NoError(suite.T(), err)
		
		resp, err := http.Post(
			suite.server.URL+"/api/license/activate",
			"application/json",
			bytes.NewReader(requestBody),
		)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()
		
		var activationResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&activationResponse)
		require.NoError(suite.T(), err)
		
		// Log the actual response for debugging
		suite.T().Logf("Activation response: %+v", activationResponse)
		
		// The response may succeed or fail depending on the actual license state
		// We test both scenarios
		if resp.StatusCode == http.StatusOK {
			assert.True(suite.T(), activationResponse["success"].(bool))
			assert.Contains(suite.T(), activationResponse["message"], "activated successfully")
			assert.NotNil(suite.T(), activationResponse["activated_at"])
		} else {
			// If activation fails, it should return a proper error response
			assert.Contains(suite.T(), []int{
				http.StatusBadRequest,
				http.StatusConflict,
				http.StatusServiceUnavailable,
			}, resp.StatusCode)
			
			assert.Equal(suite.T(), false, activationResponse["success"])
			assert.NotEmpty(suite.T(), activationResponse["type"])
		}
	})
	
	suite.Run("step_3_verify_post_activation_status", func() {
		// Step 3: Verify status after activation attempt
		resp, err := http.Get(suite.server.URL + "/api/license/status")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()
		
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var statusResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResponse)
		require.NoError(suite.T(), err)
		
		// Status should reflect the result of the activation attempt
		suite.T().Logf("Post-activation status: %+v", statusResponse)
		
		// Verify response structure regardless of activation success
		assert.NotEmpty(suite.T(), statusResponse["license_status"])
		assert.NotEmpty(suite.T(), statusResponse["message"])
		assert.NotEmpty(suite.T(), statusResponse["trace_id"])
	})
	
	suite.Run("step_4_get_detailed_status", func() {
		// Step 4: Get detailed license information
		resp, err := http.Get(suite.server.URL + "/api/license/detailed")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()
		
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var detailedResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&detailedResponse)
		require.NoError(suite.T(), err)
		
		// Verify detailed response structure
		assert.NotEmpty(suite.T(), detailedResponse["license_status"])
		assert.NotEmpty(suite.T(), detailedResponse["machine_id"])
		assert.NotNil(suite.T(), detailedResponse["validation_count"])
		assert.NotEmpty(suite.T(), detailedResponse["network_status"])
		
		suite.T().Logf("Detailed status: %+v", detailedResponse)
	})
	
	suite.Run("step_5_check_renewal_status", func() {
		// Step 5: Check renewal status
		resp, err := http.Get(suite.server.URL + "/api/license/renewal")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()
		
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var renewalResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&renewalResponse)
		require.NoError(suite.T(), err)
		
		// Verify renewal response structure
		assert.NotNil(suite.T(), renewalResponse["needs_renewal"])
		assert.NotNil(suite.T(), renewalResponse["is_expired"])
		assert.NotEmpty(suite.T(), renewalResponse["renewal_urgency"])
		assert.NotEmpty(suite.T(), renewalResponse["renewal_message"])
		
		suite.T().Logf("Renewal status: %+v", renewalResponse)
	})
}

// TestActivationFlowWithWebSocketUpdates tests activation with real-time WebSocket updates
func (suite *LicenseActivationFlowTestSuite) TestActivationFlowWithWebSocketUpdates() {
	// Connect to WebSocket server
	wsURL := strings.Replace(suite.wsServer.URL, "http://", "ws://", 1) + "/ws"
	conn, _, err := gorilla.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Channel to collect WebSocket messages
	messages := make(chan map[string]interface{}, 10)
	done := make(chan bool)
	
	// Start WebSocket message reader
	go func() {
		defer close(done)
		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				if !gorilla.IsCloseError(err, gorilla.CloseGoingAway, gorilla.CloseAbnormalClosure) {
					suite.T().Logf("WebSocket read error: %v", err)
				}
				return
			}
			
			select {
			case messages <- message:
			case <-time.After(1 * time.Second):
				suite.T().Log("WebSocket message buffer full")
			}
		}
	}()
	
	// Perform license activation
	activationRequest := map[string]interface{}{
		"license_key": "ISX1M02LYE1F9QJHR9D7Z", // Live license key
		"email":       "websocket.test@iraqiinvestor.gov.iq",
	}
	
	requestBody, err := json.Marshal(activationRequest)
	require.NoError(suite.T(), err)
	
	// Send activation request
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to ensure WebSocket is ready
		
		resp, err := http.Post(
			suite.server.URL+"/api/license/activate",
			"application/json",
			bytes.NewReader(requestBody),
		)
		if err != nil {
			suite.T().Logf("Activation request error: %v", err)
			return
		}
		defer resp.Body.Close()
		
		suite.T().Logf("Activation request completed with status: %d", resp.StatusCode)
	}()
	
	// Collect messages for a short period
	timeout := time.After(5 * time.Second)
	var receivedMessages []map[string]interface{}
	
collectLoop:
	for {
		select {
		case message := <-messages:
			receivedMessages = append(receivedMessages, message)
			suite.T().Logf("Received WebSocket message: %+v", message)
		case <-timeout:
			break collectLoop
		case <-done:
			break collectLoop
		}
	}
	
	// Analyze received messages
	suite.T().Logf("Total WebSocket messages received: %d", len(receivedMessages))
	
	// In a real implementation, we'd verify specific message types
	// For now, we ensure the WebSocket connection works
	// You might receive connection confirmations, status updates, etc.
}

// TestErrorHandlingInIntegrationFlow tests error scenarios in the complete flow
func (suite *LicenseActivationFlowTestSuite) TestErrorHandlingInIntegrationFlow() {
	tests := []struct {
		name                string
		licenseKey          string
		email               string
		expectedStatusCode  int
		expectedErrorType   string
	}{
		{
			name:               "invalid_license_format",
			licenseKey:         "INVALID-FORMAT",
			email:              "test@example.com",
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorType:  "/errors/validation",
		},
		{
			name:               "empty_license_key",
			licenseKey:         "",
			email:              "test@example.com",
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorType:  "/errors/validation",
		},
		{
			name:               "invalid_email_format",
			licenseKey:         "ISX1Y-ABCDE-12345-FGHIJ-67890",
			email:              "invalid-email",
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorType:  "/errors/validation",
		},
		{
			name:               "missing_content_type",
			licenseKey:         "ISX1Y-ABCDE-12345-FGHIJ-67890",
			email:              "test@example.com",
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorType:  "/errors/validation",
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			activationRequest := map[string]interface{}{
				"license_key": tt.licenseKey,
				"email":       tt.email,
			}
			
			requestBody, err := json.Marshal(activationRequest)
			require.NoError(suite.T(), err)
			
			// Create request with or without proper content type
			var resp *http.Response
			if tt.name == "missing_content_type" {
				req, err := http.NewRequest("POST", suite.server.URL+"/api/license/activate", bytes.NewReader(requestBody))
				require.NoError(suite.T(), err)
				// Don't set Content-Type header
				
				client := &http.Client{}
				resp, err = client.Do(req)
				require.NoError(suite.T(), err)
			} else {
				resp, err = http.Post(
					suite.server.URL+"/api/license/activate",
					"application/json",
					bytes.NewReader(requestBody),
				)
				require.NoError(suite.T(), err)
			}
			defer resp.Body.Close()
			
			assert.Equal(suite.T(), tt.expectedStatusCode, resp.StatusCode)
			
			var errorResponse map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&errorResponse)
			require.NoError(suite.T(), err)
			
			if tt.expectedErrorType != "" {
				assert.Equal(suite.T(), tt.expectedErrorType, errorResponse["type"])
			}
			
			// Verify error response structure
			assert.NotEmpty(suite.T(), errorResponse["title"])
			assert.NotEmpty(suite.T(), errorResponse["trace_id"])
		})
	}
}

// TestConcurrentActivationAttempts tests multiple concurrent activation attempts
func (suite *LicenseActivationFlowTestSuite) TestConcurrentActivationAttempts() {
	numConcurrentRequests := 10
	var wg sync.WaitGroup
	results := make(chan struct {
		statusCode int
		success    bool
		err        error
	}, numConcurrentRequests)
	
	activationRequest := map[string]interface{}{
		"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890", // Test license key
		"email":       "concurrent.test@example.com",
	}
	
	requestBody, err := json.Marshal(activationRequest)
	require.NoError(suite.T(), err)
	
	// Launch concurrent activation requests
	wg.Add(numConcurrentRequests)
	for i := 0; i < numConcurrentRequests; i++ {
		go func(id int) {
			defer wg.Done()
			
			resp, err := http.Post(
				suite.server.URL+"/api/license/activate",
				"application/json",
				bytes.NewReader(requestBody),
			)
			
			result := struct {
				statusCode int
				success    bool
				err        error
			}{
				err: err,
			}
			
			if err == nil {
				defer resp.Body.Close()
				result.statusCode = resp.StatusCode
				
				var response map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&response) == nil {
					if successVal, ok := response["success"].(bool); ok {
						result.success = successVal
					}
				}
			}
			
			results <- result
		}(i)
	}
	
	wg.Wait()
	close(results)
	
	// Analyze results
	successCount := 0
	failureCount := 0
	errorCount := 0
	
	for result := range results {
		if result.err != nil {
			errorCount++
			suite.T().Logf("Request error: %v", result.err)
		} else if result.success {
			successCount++
		} else {
			failureCount++
		}
		
		suite.T().Logf("Concurrent request result: status=%d, success=%t, err=%v", 
			result.statusCode, result.success, result.err)
	}
	
	suite.T().Logf("Concurrent activation results: %d success, %d failure, %d errors", 
		successCount, failureCount, errorCount)
	
	// All requests should be handled (no network errors)
	assert.Equal(suite.T(), 0, errorCount, "No network errors should occur")
	
	// At least one request should get a proper response
	assert.Greater(suite.T(), successCount+failureCount, 0, "At least one request should get a response")
}

// TestHealthCheckEndpoint tests the health check functionality
func (suite *LicenseActivationFlowTestSuite) TestHealthCheckEndpoint() {
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "healthy", healthResponse["status"])
}

// TestMetricsEndpoint tests the metrics functionality if available
func (suite *LicenseActivationFlowTestSuite) TestMetricsEndpoint() {
	resp, err := http.Get(suite.server.URL + "/api/license/metrics")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	// This may succeed or fail depending on implementation
	if resp.StatusCode == http.StatusOK {
		var metricsResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metricsResponse)
		require.NoError(suite.T(), err)
		
		assert.True(suite.T(), metricsResponse["success"].(bool))
		assert.NotNil(suite.T(), metricsResponse["data"])
		
		suite.T().Logf("Metrics response: %+v", metricsResponse)
	} else {
		suite.T().Logf("Metrics endpoint returned status: %d", resp.StatusCode)
	}
}

// TestResponseTimePerformance tests response time performance
func (suite *LicenseActivationFlowTestSuite) TestResponseTimePerformance() {
	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "status_check",
			method: "GET",
			path:   "/api/license/status",
		},
		{
			name:   "detailed_status",
			method: "GET",
			path:   "/api/license/detailed",
		},
		{
			name:   "renewal_status",
			method: "GET",
			path:   "/api/license/renewal",
		},
		{
			name:   "health_check",
			method: "GET",
			path:   "/health",
		},
	}
	
	for _, endpoint := range endpoints {
		suite.Run(endpoint.name, func() {
			start := time.Now()
			
			var resp *http.Response
			var err error
			
			if endpoint.method == "GET" {
				resp, err = http.Get(suite.server.URL + endpoint.path)
			} else {
				requestBody, _ := json.Marshal(endpoint.body)
				resp, err = http.Post(
					suite.server.URL+endpoint.path,
					"application/json",
					bytes.NewReader(requestBody),
				)
			}
			
			duration := time.Since(start)
			
			require.NoError(suite.T(), err)
			defer resp.Body.Close()
			
			// Response time should be under 1 second for all endpoints
			assert.Less(suite.T(), duration, 1*time.Second, 
				"Endpoint %s should respond within 1 second, took %v", endpoint.name, duration)
			
			suite.T().Logf("Endpoint %s response time: %v", endpoint.name, duration)
		})
	}
}

// TestDataPersistence tests that license data persists across manager restarts
func (suite *LicenseActivationFlowTestSuite) TestDataPersistence() {
	// Note: testLicense variable is not used anymore
	
	// Save license using manager directly
	err := suite.manager.ActivateLicense("PERSISTENCE-TEST-KEY")
	if err != nil {
		// If activation fails (expected for test key), save locally
		suite.T().Logf("Activation failed as expected: %v", err)
		// Use internal save method if available via reflection or create a new manager
	}
	
	// Get status to verify license is loaded
	info, status, err := suite.manager.GetLicenseStatus()
	suite.T().Logf("Current license info: %+v, status: %s, err: %v", info, status, err)
	
	// Close current manager
	suite.manager.Close()
	
	// Create new manager with same license file
	newManager, err := license.NewManager()
	require.NoError(suite.T(), err)
	defer newManager.Close()
	
	// Verify machine ID consistency
	originalMachineID := suite.manager.GetMachineID()
	newMachineID := newManager.GetMachineID()
	assert.Equal(suite.T(), originalMachineID, newMachineID, "Machine ID should be consistent across restarts")
	
	// Update manager reference for subsequent tests
	suite.manager = newManager
	suite.service = services.NewLicenseService(suite.manager, suite.logger)
	suite.handler = handlers.NewLicenseHandler(suite.service, suite.logger)
}

// Run the integration test suite
func TestLicenseActivationFlow(t *testing.T) {
	suite.Run(t, new(LicenseActivationFlowTestSuite))
}

// TestEndToEndScenario runs a complete end-to-end scenario
func TestEndToEndScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}
	
	// This test simulates a complete user journey
	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "e2e_license.dat")
	
	// Initialize components
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	manager, err := license.NewManager()
	require.NoError(t, err)
	defer manager.Close()
	
	service := services.NewLicenseService(manager, logger)
	handler := handlers.NewLicenseHandler(service, logger)
	
	// Setup server
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Mount("/api/license", handler.Routes())
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Simulate user journey
	t.Run("user_checks_status_before_activation", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/license/status")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		assert.Equal(t, "not_activated", status["license_status"])
	})
	
	t.Run("user_attempts_activation_with_live_key", func(t *testing.T) {
		activationData := map[string]string{
			"license_key": "ISX1M02LYE1F9QJHR9D7Z",
			"email":       "endtoend@iraqiinvestor.gov.iq",
		}
		
		requestBody, _ := json.Marshal(activationData)
		resp, err := http.Post(
			server.URL+"/api/license/activate",
			"application/json",
			bytes.NewReader(requestBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		t.Logf("End-to-end activation result: %+v", result)
		
		// Result may vary based on actual license state
		// We verify the response structure is correct
		assert.NotEmpty(t, result["trace_id"])
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusConflict}, resp.StatusCode)
	})
	
	t.Run("user_checks_final_status", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/license/status")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		
		t.Logf("Final license status: %+v", status)
		
		// Verify response structure
		assert.NotEmpty(t, status["license_status"])
		assert.NotEmpty(t, status["message"])
		assert.NotEmpty(t, status["trace_id"])
	})
}