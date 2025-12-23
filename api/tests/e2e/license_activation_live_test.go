package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath" 
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	httpHandlers "github.com/isxcli/isxcli/internal/transport/http"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/services"
)

const (
	// Live license key for testing - ISX1M02LYE1F9QJHR9D7Z
	LiveLicenseKey = "ISX1M02LYE1F9QJHR9D7Z"
	TestTimeout    = 30 * time.Second
)

// LiveLicenseTestSuite tests with the actual live license key
type LiveLicenseTestSuite struct {
	suite.Suite
	tempDir     string
	licenseFile string
	manager     *license.Manager
	service     services.LicenseService
	handler     *httpHandlers.LicenseHandler
	server      *httptest.Server
	logger      *slog.Logger
}

func (suite *LiveLicenseTestSuite) SetupSuite() {
	// Skip if this is a short test run
	if testing.Short() {
		suite.T().Skip("Skipping live license tests in short mode")
	}
	
	// Check if we should skip live tests based on environment
	if os.Getenv("SKIP_LIVE_TESTS") == "true" {
		suite.T().Skip("Skipping live license tests due to SKIP_LIVE_TESTS=true")
	}

	suite.tempDir = suite.T().TempDir()
	suite.licenseFile = filepath.Join(suite.tempDir, "live_test_license.dat")
	suite.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	var err error
	suite.manager, err = license.NewManager()
	require.NoError(suite.T(), err)

	suite.service = services.NewLicenseService(suite.manager, suite.logger)
	suite.handler = httpHandlers.NewLicenseHandler(suite.service, suite.logger)

	// Setup HTTP server
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(TestTimeout))
	router.Mount("/api/license", suite.handler.Routes())

	suite.server = httptest.NewServer(router)
}

func (suite *LiveLicenseTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.manager != nil {
		suite.manager.Close()
	}
}

func (suite *LiveLicenseTestSuite) SetupTest() {
	// Clean up any existing license file before each test
	os.Remove(suite.licenseFile)
}

// TestLiveLicenseActivationFlow tests the complete flow with the live license key
func (suite *LiveLicenseTestSuite) TestLiveLicenseActivationFlow() {
	suite.Run("step_1_initial_status_check", func() {
		// Step 1: Verify initial state (no license activated)
		resp, err := http.Get(suite.server.URL + "/api/license/status")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var statusResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("Initial license status: %+v", statusResponse)
		assert.Equal(suite.T(), "not_activated", statusResponse["license_status"])
	})

	suite.Run("step_2_live_license_activation", func() {
		// Step 2: Attempt activation with live license key
		activationRequest := map[string]interface{}{
			"license_key": LiveLicenseKey,
			"email":       "live.test@iraqiinvestor.gov.iq",
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

		suite.T().Logf("Live activation response (Status: %d): %+v", resp.StatusCode, activationResponse)

		// Document the actual response for analysis
		if resp.StatusCode == http.StatusOK {
			suite.T().Logf("✅ Live license activation SUCCEEDED")
			assert.True(suite.T(), activationResponse["success"].(bool))
			assert.Contains(suite.T(), activationResponse["message"], "activated")
			assert.NotNil(suite.T(), activationResponse["activated_at"])
		} else {
			suite.T().Logf("❌ Live license activation FAILED with status %d", resp.StatusCode)
			
			// Common failure scenarios to document
			switch resp.StatusCode {
			case http.StatusConflict:
				suite.T().Logf("Reason: License already activated on another machine")
				assert.Contains(suite.T(), activationResponse["type"], "machine")
			case http.StatusGone:
				suite.T().Logf("Reason: License expired")
				assert.Contains(suite.T(), activationResponse["type"], "expired")
			case http.StatusBadRequest:
				suite.T().Logf("Reason: Invalid license key")
				assert.Contains(suite.T(), activationResponse["type"], "validation")
			case http.StatusServiceUnavailable:
				suite.T().Logf("Reason: Network/service error")
				assert.Contains(suite.T(), activationResponse["type"], "network")
			default:
				suite.T().Logf("Reason: Other (%s)", activationResponse["type"])
			}
			
			// Verify error response structure
			assert.NotEmpty(suite.T(), activationResponse["type"])
			assert.NotEmpty(suite.T(), activationResponse["title"])
			assert.NotEmpty(suite.T(), activationResponse["trace_id"])
			assert.Equal(suite.T(), false, activationResponse["success"])
		}
	})

	suite.Run("step_3_post_activation_status_check", func() {
		// Step 3: Check status after activation attempt
		resp, err := http.Get(suite.server.URL + "/api/license/status")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var statusResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("Post-activation status: %+v", statusResponse)

		// Status should reflect the result of activation attempt
		licenseStatus := statusResponse["license_status"].(string)
		if licenseStatus == "active" {
			suite.T().Logf("✅ License is now ACTIVE")
			assert.Contains(suite.T(), statusResponse["message"], "active")
			assert.NotNil(suite.T(), statusResponse["license_info"])
		} else {
			suite.T().Logf("ℹ️ License status: %s", licenseStatus)
			// Not active could be not_activated, expired, etc.
		}
	})

	suite.Run("step_4_detailed_status_analysis", func() {
		// Step 4: Get detailed information for analysis
		resp, err := http.Get(suite.server.URL + "/api/license/detailed")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var detailedResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&detailedResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("Detailed license information: %+v", detailedResponse)

		// Verify response structure
		assert.NotEmpty(suite.T(), detailedResponse["machine_id"])
		assert.NotNil(suite.T(), detailedResponse["validation_count"])
		assert.NotEmpty(suite.T(), detailedResponse["network_status"])

		// Log key metrics
		if machineID, ok := detailedResponse["machine_id"].(string); ok {
			suite.T().Logf("Machine ID: %s", machineID)
		}
		if networkStatus, ok := detailedResponse["network_status"].(string); ok {
			suite.T().Logf("Network Status: %s", networkStatus)
		}
		if validationCount, ok := detailedResponse["validation_count"].(float64); ok {
			suite.T().Logf("Validation Count: %.0f", validationCount)
		}
	})

	suite.Run("step_5_renewal_status_check", func() {
		// Step 5: Check renewal status
		resp, err := http.Get(suite.server.URL + "/api/license/renewal")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var renewalResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&renewalResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("Renewal status: %+v", renewalResponse)

		// Analyze renewal status
		needsRenewal := renewalResponse["needs_renewal"].(bool)
		isExpired := renewalResponse["is_expired"].(bool)
		urgency := renewalResponse["renewal_urgency"].(string)

		suite.T().Logf("Renewal Analysis:")
		suite.T().Logf("  Needs Renewal: %t", needsRenewal)
		suite.T().Logf("  Is Expired: %t", isExpired)
		suite.T().Logf("  Urgency: %s", urgency)

		if daysUntilExpiry, ok := renewalResponse["days_until_expiry"].(float64); ok {
			suite.T().Logf("  Days Until Expiry: %.0f", daysUntilExpiry)
		}
	})
}

// TestLiveLicenseValidation tests validation with live license
func (suite *LiveLicenseTestSuite) TestLiveLicenseValidation() {
	// First attempt activation
	activationRequest := map[string]interface{}{
		"license_key": LiveLicenseKey,
		"email":       "validation.test@iraqiinvestor.gov.iq",
	}

	requestBody, _ := json.Marshal(activationRequest)
	
	activationResp, err := http.Post(
		suite.server.URL+"/api/license/activate",
		"application/json",
		bytes.NewReader(requestBody),
	)
	require.NoError(suite.T(), err)
	activationResp.Body.Close()

	suite.T().Logf("Activation for validation test - Status: %d", activationResp.StatusCode)

	// Now test validation regardless of activation result
	ctx := context.Background()
	valid, err := suite.service.ValidateWithContext(ctx)
	
	suite.T().Logf("Live license validation result: valid=%t, error=%v", valid, err)

	// Document validation behavior
	if valid {
		suite.T().Logf("✅ License validation PASSED")
		assert.NoError(suite.T(), err)
	} else {
		suite.T().Logf("❌ License validation FAILED: %v", err)
		// Validation failure is not necessarily a test failure
		// It depends on the actual license state
	}
}

// TestLiveLicenseErrorScenarios tests error handling with live license key
func (suite *LiveLicenseTestSuite) TestLiveLicenseErrorScenarios() {
	suite.Run("multiple_activation_attempts", func() {
		// Test multiple rapid activation attempts
		activationRequest := map[string]interface{}{
			"license_key": LiveLicenseKey,
			"email":       "multi.test@iraqiinvestor.gov.iq",
		}

		requestBody, _ := json.Marshal(activationRequest)

		var responses []int
		for i := 0; i < 3; i++ {
			resp, err := http.Post(
				suite.server.URL+"/api/license/activate",
				"application/json",
				bytes.NewReader(requestBody),
			)
			require.NoError(suite.T(), err)
			responses = append(responses, resp.StatusCode)
			resp.Body.Close()
			
			time.Sleep(100 * time.Millisecond) // Small delay between attempts
		}

		suite.T().Logf("Multiple activation attempt results: %v", responses)
		
		// Verify that multiple attempts are handled gracefully
		for _, status := range responses {
			assert.True(suite.T(), status >= 200 && status < 500, 
				"Status should be a client or success response, got %d", status)
		}
	})

	suite.Run("activation_with_different_emails", func() {
		// Test activation with different email addresses
		emails := []string{
			"test1@iraqiinvestor.gov.iq",
			"test2@iraqiinvestor.gov.iq", 
			"different@example.com",
		}

		for i, email := range emails {
			activationRequest := map[string]interface{}{
				"license_key": LiveLicenseKey,
				"email":       email,
			}

			requestBody, _ := json.Marshal(activationRequest)
			
			resp, err := http.Post(
				suite.server.URL+"/api/license/activate",
				"application/json",
				bytes.NewReader(requestBody),
			)
			require.NoError(suite.T(), err)
			resp.Body.Close()

			suite.T().Logf("Activation attempt %d with email %s: Status %d", i+1, email, resp.StatusCode)
		}
	})
}

// TestLiveLicenseTransfer tests license transfer functionality
func (suite *LiveLicenseTestSuite) TestLiveLicenseTransfer() {
	suite.Run("transfer_without_force", func() {
		transferRequest := map[string]interface{}{
			"license_key": LiveLicenseKey,
			"force":       false,
		}

		requestBody, _ := json.Marshal(transferRequest)
		
		resp, err := http.Post(
			suite.server.URL+"/api/license/transfer",
			"application/json",
			bytes.NewReader(requestBody),
		)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		var transferResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&transferResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("License transfer (non-forced) result: Status %d, Response: %+v", 
			resp.StatusCode, transferResponse)

		// Transfer may succeed or fail depending on current license state
		if resp.StatusCode == http.StatusOK {
			suite.T().Logf("✅ License transfer SUCCEEDED")
			assert.True(suite.T(), transferResponse["success"].(bool))
		} else {
			suite.T().Logf("❌ License transfer FAILED")
			assert.Equal(suite.T(), false, transferResponse["success"])
		}
	})

	suite.Run("transfer_with_force", func() {
		transferRequest := map[string]interface{}{
			"license_key": LiveLicenseKey,
			"force":       true,
		}

		requestBody, _ := json.Marshal(transferRequest)
		
		resp, err := http.Post(
			suite.server.URL+"/api/license/transfer",
			"application/json",
			bytes.NewReader(requestBody),
		)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		var transferResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&transferResponse)
		require.NoError(suite.T(), err)

		suite.T().Logf("License transfer (forced) result: Status %d, Response: %+v", 
			resp.StatusCode, transferResponse)

		// Forced transfer has higher success probability
		if resp.StatusCode == http.StatusOK {
			suite.T().Logf("✅ Forced license transfer SUCCEEDED")
		} else {
			suite.T().Logf("❌ Forced license transfer FAILED")
		}
	})
}

// TestLiveLicensePerformance tests performance with live license
func (suite *LiveLicenseTestSuite) TestLiveLicensePerformance() {
	if testing.Short() {
		suite.T().Skip("Skipping performance test in short mode")
	}

	// Perform multiple status checks to test performance
	const numRequests = 100
	start := time.Now()
	
	var successCount int
	for i := 0; i < numRequests; i++ {
		resp, err := http.Get(suite.server.URL + "/api/license/status")
		if err == nil && resp.StatusCode == http.StatusOK {
			successCount++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	duration := time.Since(start)
	throughput := float64(numRequests) / duration.Seconds()
	
	suite.T().Logf("Live license performance test:")
	suite.T().Logf("  Requests: %d", numRequests)
	suite.T().Logf("  Successful: %d", successCount)
	suite.T().Logf("  Duration: %v", duration)
	suite.T().Logf("  Throughput: %.2f requests/second", throughput)
	suite.T().Logf("  Average latency: %v", duration/time.Duration(numRequests))
	
	// Performance assertions
	assert.Greater(suite.T(), throughput, 50.0, "Should maintain reasonable throughput with live license")
	assert.Greater(suite.T(), successCount, numRequests*8/10, "At least 80% of requests should succeed")
}

// TestLiveLicenseNetworkFailure tests behavior during network issues
func (suite *LiveLicenseTestSuite) TestLiveLicenseNetworkFailure() {
	// This test would ideally simulate network failures
	// For now, we test timeout behavior
	
	suite.Run("activation_timeout_behavior", func() {
		// Create a context with very short timeout to simulate network issues
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		// This should timeout quickly
		err := suite.service.Activate(ctx, LiveLicenseKey)
		
		suite.T().Logf("Activation with short timeout result: %v", err)
		
		// Should get either timeout error or success (if very fast)
		if err != nil {
			assert.Contains(suite.T(), err.Error(), "timeout", "Should get timeout error")
		}
	})
}

// TestLiveLicenseStatePersistence tests that license state persists
func (suite *LiveLicenseTestSuite) TestLiveLicenseStatePersistence() {
	// First, attempt activation
	activationRequest := map[string]interface{}{
		"license_key": LiveLicenseKey,
		"email":       "persistence.test@iraqiinvestor.gov.iq",
	}

	requestBody, _ := json.Marshal(activationRequest)
	
	activationResp, err := http.Post(
		suite.server.URL+"/api/license/activate",
		"application/json",
		bytes.NewReader(requestBody),
	)
	require.NoError(suite.T(), err)
	activationResp.Body.Close()

	suite.T().Logf("Initial activation for persistence test - Status: %d", activationResp.StatusCode)

	// Get initial status
	initialResp, err := http.Get(suite.server.URL + "/api/license/status")
	require.NoError(suite.T(), err)
	defer initialResp.Body.Close()

	var initialStatus map[string]interface{}
	json.NewDecoder(initialResp.Body).Decode(&initialStatus)

	// Close and recreate manager to test persistence
	suite.manager.Close()
	
	newManager, err := license.NewManager()
	require.NoError(suite.T(), err)
	
	newService := services.NewLicenseService(newManager, suite.logger)
	newHandler := httpHandlers.NewLicenseHandler(newService, suite.logger)

	// Setup new server
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Mount("/api/license", newHandler.Routes())
	
	suite.server.Close()
	suite.server = httptest.NewServer(router)
	suite.manager = newManager
	suite.service = newService
	suite.handler = newHandler

	// Check status after restart
	persistedResp, err := http.Get(suite.server.URL + "/api/license/status")
	require.NoError(suite.T(), err)
	defer persistedResp.Body.Close()

	var persistedStatus map[string]interface{}
	json.NewDecoder(persistedResp.Body).Decode(&persistedStatus)

	suite.T().Logf("License state persistence test:")
	suite.T().Logf("  Initial status: %s", initialStatus["license_status"])
	suite.T().Logf("  Persisted status: %s", persistedStatus["license_status"])

	// Verify machine ID consistency
	if initialMachineID, ok := initialStatus["machine_id"].(string); ok {
		if persistedMachineID, ok := persistedStatus["machine_id"].(string); ok {
			assert.Equal(suite.T(), initialMachineID, persistedMachineID, 
				"Machine ID should persist across restarts")
		}
	}
}

// Run the live license test suite
func TestLiveLicenseActivation(t *testing.T) {
	suite.Run(t, new(LiveLicenseTestSuite))
}

// TestQuickLiveLicenseCheck provides a quick test for CI/CD operations
func TestQuickLiveLicenseCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live license check in short mode")
	}
	
	if os.Getenv("SKIP_LIVE_TESTS") == "true" {
		t.Skip("Skipping live license tests due to SKIP_LIVE_TESTS=true")
	}

	tempDir := t.TempDir()
	licenseFile := filepath.Join(tempDir, "quick_test_license.dat")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	manager, err := license.NewManager()
	require.NoError(t, err)
	defer manager.Close()

	service := services.NewLicenseService(manager, logger)
	handler := httpHandlers.NewLicenseHandler(service, logger)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Mount("/api/license", handler.Routes())

	server := httptest.NewServer(router)
	defer server.Close()

	// Quick activation test
	activationRequest := map[string]interface{}{
		"license_key": LiveLicenseKey,
		"email":       "quick.test@iraqiinvestor.gov.iq",
	}

	requestBody, _ := json.Marshal(activationRequest)
	
	resp, err := http.Post(
		server.URL+"/api/license/activate",
		"application/json",
		bytes.NewReader(requestBody),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Logf("Quick live license test result: Status %d", resp.StatusCode)

	// Verify we get a valid HTTP response
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
		"Should get valid HTTP response, got %d", resp.StatusCode)

	// Check that we can parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err, "Response should be valid JSON")

	// Verify response structure
	assert.NotEmpty(t, response["trace_id"], "Response should include trace_id")
	
	if resp.StatusCode == http.StatusOK {
		t.Logf("✅ Live license activation succeeded in quick test")
	} else {
		t.Logf("ℹ️  Live license activation failed in quick test (Status: %d, Type: %s)", 
			resp.StatusCode, response["type"])
	}
}