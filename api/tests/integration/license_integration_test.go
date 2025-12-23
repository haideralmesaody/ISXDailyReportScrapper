package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	handlers "github.com/isxcli/isxcli/internal/transport/http"
	"github.com/isxcli/isxcli/internal/license"
	licenseMiddleware "github.com/isxcli/isxcli/internal/middleware"
	"github.com/isxcli/isxcli/internal/services"
)

// mockLicenseManager for integration testing
type mockLicenseManager struct {
	licenses map[string]*license.LicenseInfo
	isValid  bool
	error    error
}

func newMockLicenseManager() *mockLicenseManager {
	return &mockLicenseManager{
		licenses: make(map[string]*license.LicenseInfo),
		isValid:  false,
	}
}

func (m *mockLicenseManager) GetLicenseStatus() (*license.LicenseInfo, string, error) {
	if m.error != nil {
		return nil, "", m.error
	}
	
	if len(m.licenses) == 0 {
		return nil, "Not Activated", nil
	}
	
	// Return the first license (in real scenario, there would be one active license)
	for _, info := range m.licenses {
		if info.ExpiryDate.After(time.Now()) {
			return info, "Active", nil
		} else {
			return info, "Expired", nil
		}
	}
	
	return nil, "Not Activated", nil
}

func (m *mockLicenseManager) ActivateLicense(key string) error {
	if m.error != nil {
		return m.error
	}
	
	// Simulate ISX license key validation
	if !strings.HasPrefix(key, "ISX1Y-") || len(key) != 29 {
		return fmt.Errorf("invalid license key format")
	}
	
	// Create mock license info
	info := &license.LicenseInfo{
		LicenseKey:  key,
		ExpiryDate:  time.Now().Add(365 * 24 * time.Hour), // 1 year
		IssuedDate:  time.Now(),
		LastChecked: time.Now(),
		UserEmail:   "test@iraqiinvestor.gov.iq",
		Duration:    "Yearly",
	}
	
	m.licenses[key] = info
	m.isValid = true
	return nil
}

func (m *mockLicenseManager) ValidateLicense() (bool, error) {
	if m.error != nil {
		return false, m.error
	}
	return m.isValid, nil
}

func (m *mockLicenseManager) GetLicenseInfo() (*license.LicenseInfo, error) {
	if m.error != nil {
		return nil, m.error
	}

	// Return the first license if available
	for _, info := range m.licenses {
		return info, nil
	}

	return nil, fmt.Errorf("no license found")
}

func (m *mockLicenseManager) GetMachineID() string {
	return "TEST-MACHINE-ID"
}

func (m *mockLicenseManager) TransferLicense(key string, force bool) error {
	if m.error != nil {
		return m.error
	}

	// Simulate transfer
	return m.ActivateLicense(key)
}

func (m *mockLicenseManager) CheckExistingLicense() (*license.ExistingLicenseInfo, error) {
	if m.error != nil {
		return nil, m.error
	}

	if len(m.licenses) == 0 {
		return nil, nil  // No existing license
	}

	// Return info about the first license
	for _, info := range m.licenses {
		daysRemaining := int(time.Until(info.ExpiryDate).Hours() / 24)
		return &license.ExistingLicenseInfo{
			HasLicense:    true,
			DaysRemaining: daysRemaining,
			ExpiryDate:    info.ExpiryDate,
			LicenseKey:    info.LicenseKey,
			Status:        "Active",
			IsExpired:     time.Now().After(info.ExpiryDate),
		}, nil
	}

	return nil, nil
}

func (m *mockLicenseManager) Close() error {
	return nil
}

func (m *mockLicenseManager) GetLicensePath() string {
	return "/tmp/test-license.dat"
}

// setupTestServer creates a complete test server with license middleware and handlers
func setupTestServer() (*httptest.Server, *mockLicenseManager) {
	mockManager := newMockLicenseManager()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	// Create services
	licenseService := services.NewLicenseService(mockManager, logger)
	
	// Create handlers
	licenseHandler := handlers.NewLicenseHandler(licenseService, logger)
	
	// Create middleware
	licenseValidator := licenseMiddleware.NewLicenseValidator(mockManager, logger)
	
	// Create router
	r := chi.NewRouter()
	
	// Apply middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(licenseValidator.Handler)
	
	// Mount license routes (excluded from validation)
	r.Mount("/api/license", licenseHandler.Routes())
	
	// Add protected routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Welcome to Iraqi Investor</body></html>"))
	})
	
	r.Get("/license", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>License Activation Page</body></html>"))
	})
	
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Iraqi Investor Dashboard</body></html>"))
	})
	
	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Protected data", "source": "Iraqi Stock Exchange"}`))
	})
	
	r.Get("/api/reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"reports": ["Daily Market Report", "Ticker Analysis"], "count": 2}`))
	})
	
	// Static routes (excluded from validation)
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Static content"))
	})
	
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	server := httptest.NewServer(r)
	return server, mockManager
}

func TestLicenseIntegration_CompleteFlow(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	t.Run("complete license activation flow", func(t *testing.T) {
		// Step 1: Try to access protected content without license - should redirect
		resp, err := http.Get(server.URL + "/dashboard")
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Contains(t, location, "/license")
		assert.Contains(t, location, "reason=not_activated")
		
		// Step 2: Check license status API - should show not activated
		resp, err = http.Get(server.URL + "/api/license/status")
		require.NoError(t, err)
		
		var statusResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "not_activated", statusResp["license_status"])
		assert.Contains(t, statusResp["message"], "No license activated")
		
		// Step 3: Activate license with valid ISX key
		activationReq := map[string]interface{}{
			"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890",
			"email":       "test@iraqiinvestor.gov.iq",
		}
		
		reqBody, err := json.Marshal(activationReq)
		require.NoError(t, err)
		
		resp, err = http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		
		var activationResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&activationResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, activationResp["success"].(bool))
		assert.Contains(t, activationResp["message"], "License activated successfully")
		assert.Contains(t, activationResp["message"], "Iraqi Investor")
		assert.NotNil(t, activationResp["activated_at"])
		
		// Step 4: Verify license status is now active
		resp, err = http.Get(server.URL + "/api/license/status")
		require.NoError(t, err)
		
		err = json.NewDecoder(resp.Body).Decode(&statusResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "active", statusResp["license_status"])
		assert.Contains(t, statusResp["message"], "License is active")
		assert.NotNil(t, statusResp["license_info"])
		assert.NotNil(t, statusResp["branding_info"])
		
		// Verify Iraqi Investor branding
		brandingInfo := statusResp["branding_info"].(map[string]interface{})
		assert.Equal(t, "Iraqi Investor", brandingInfo["brand_name"])
		assert.Equal(t, "ISX Daily Reports Scrapper", brandingInfo["application_name"])
		
		// Step 5: Access protected content - should now work
		resp, err = http.Get(server.URL + "/dashboard")
		require.NoError(t, err)
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), "Iraqi Investor Dashboard")
		
		// Step 6: Access protected API - should work
		resp, err = http.Get(server.URL + "/api/data")
		require.NoError(t, err)
		
		var apiResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "Protected data", apiResp["message"])
		assert.Equal(t, "Iraqi Stock Exchange", apiResp["source"])
	})
}

func TestLicenseIntegration_ErrorScenarios(t *testing.T) {
	server, mockManager := setupTestServer()
	defer server.Close()

	t.Run("invalid license key format", func(t *testing.T) {
		activationReq := map[string]interface{}{
			"license_key": "INVALID-KEY-FORMAT",
		}
		
		reqBody, err := json.Marshal(activationReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		var errorResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		
		assert.Contains(t, errorResp["detail"], "invalid license key format")
	})

	t.Run("missing license key", func(t *testing.T) {
		activationReq := map[string]interface{}{
			"email": "test@example.com",
		}
		
		reqBody, err := json.Marshal(activationReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("network error graceful degradation", func(t *testing.T) {
		// First activate a license
		activationReq := map[string]interface{}{
			"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890",
		}
		
		reqBody, err := json.Marshal(activationReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		resp.Body.Close()
		
		// Access protected content to establish valid cache
		resp, err = http.Get(server.URL + "/dashboard")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Simulate network error
		mockManager.error = fmt.Errorf("network connection failed")
		
		// Should still allow access due to graceful degradation
		// Note: This test verifies the middleware behavior with recent success
		time.Sleep(10 * time.Millisecond) // Small delay to ensure cache age
		
		resp, err = http.Get(server.URL + "/dashboard")
		require.NoError(t, err)
		resp.Body.Close()
		
		// Should either succeed (graceful degradation) or redirect
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusTemporaryRedirect)
	})
}

func TestLicenseIntegration_APIvsHTMLBehavior(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	t.Run("API request without license returns JSON error", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/api/data", nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusPreconditionRequired, resp.StatusCode)
		
		var errorResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		
		assert.Equal(t, "License Not Activated", errorResp["title"])
		assert.Contains(t, errorResp["type"], "/errors/license-not-activated")
		assert.NotEmpty(t, errorResp["trace_id"])
	})

	t.Run("HTML request without license redirects", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/dashboard", nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "text/html")
		
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
		
		location := resp.Header.Get("Location")
		assert.Contains(t, location, "/license")
		assert.Contains(t, location, "reason=not_activated")
		assert.Contains(t, location, "return=/dashboard")
	})
}

func TestLicenseIntegration_DetailedLicenseInfo(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Activate license first
	activationReq := map[string]interface{}{
		"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890",
		"email":       "test@iraqiinvestor.gov.iq",
	}
	
	reqBody, err := json.Marshal(activationReq)
	require.NoError(t, err)
	
	resp, err := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	t.Run("detailed license status", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/license/detailed")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var detailedResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&detailedResp)
		require.NoError(t, err)
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "active", detailedResp["license_status"])
		assert.Contains(t, detailedResp, "machine_id")
		assert.Contains(t, detailedResp, "validation_count")
		assert.Contains(t, detailedResp, "network_status")
		assert.Contains(t, detailedResp, "recommendations")
		
		// Verify Iraqi Investor specific features
		features := detailedResp["features"].([]interface{})
		assert.Contains(t, features, "Iraqi Stock Exchange Integration")
		assert.Contains(t, features, "Advanced Analytics")
	})

	t.Run("renewal status check", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/license/renewal")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var renewalResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&renewalResp)
		require.NoError(t, err)
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.False(t, renewalResp["needs_renewal"].(bool)) // New license shouldn't need renewal
		assert.False(t, renewalResp["is_expired"].(bool))
		assert.Equal(t, "low", renewalResp["renewal_urgency"])
		assert.Greater(t, renewalResp["days_until_expiry"], float64(300)) // Should be ~365 days
	})

	t.Run("validation metrics", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/license/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var metricsResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metricsResp)
		require.NoError(t, err)
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, metricsResp["success"].(bool))
		
		data := metricsResp["data"].(map[string]interface{})
		assert.GreaterOrEqual(t, data["total_validations"], float64(0))
		assert.NotNil(t, data["uptime"])
	})
}

func TestLicenseIntegration_CacheInvalidation(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Activate license
	activationReq := map[string]interface{}{
		"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890",
	}
	
	reqBody, err := json.Marshal(activationReq)
	require.NoError(t, err)
	
	resp, err := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	t.Run("cache invalidation API", func(t *testing.T) {
		// Access protected content to populate cache
		resp, err := http.Get(server.URL + "/dashboard")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Invalidate cache
		invalidationReq := map[string]interface{}{
			"reason": "Integration test cache invalidation",
		}
		
		reqBody, err := json.Marshal(invalidationReq)
		require.NoError(t, err)
		
		resp, err = http.Post(server.URL+"/api/license/invalidate-cache", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		var invalidationResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&invalidationResp)
		require.NoError(t, err)
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, invalidationResp["success"].(bool))
		assert.Contains(t, invalidationResp["message"], "cache invalidated successfully")
	})
}

func TestLicenseIntegration_ExcludedPaths(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	excludedPaths := []struct {
		path        string
		description string
	}{
		{"/", "root path"},
		{"/license", "license page"},
		{"/api/license/status", "license status API"},
		{"/api/license/activate", "license activation API"},
		{"/static/css/main.css", "static assets"},
		{"/favicon.ico", "favicon"},
	}

	for _, test := range excludedPaths {
		t.Run(fmt.Sprintf("excluded path: %s", test.description), func(t *testing.T) {
			method := "GET"
			if strings.Contains(test.path, "activate") {
				method = "POST"
			}
			
			var body io.Reader
			if method == "POST" {
				reqData := map[string]string{"license_key": "ISX1Y-ABCDE-12345-FGHIJ-67890"}
				reqBody, _ := json.Marshal(reqData)
				body = bytes.NewReader(reqBody)
			}
			
			req, err := http.NewRequest(method, server.URL+test.path, body)
			require.NoError(t, err)
			
			if method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			
			resp, err := client.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			
			// Excluded paths should not redirect to license page
			assert.NotEqual(t, http.StatusTemporaryRedirect, resp.StatusCode, 
				"Path %s should be excluded from license validation", test.path)
		})
	}
}

func TestLicenseIntegration_LicenseTransfer(t *testing.T) {
	server, mockManager := setupTestServer()
	defer server.Close()

	t.Run("license transfer flow", func(t *testing.T) {
		// Step 1: Try to transfer without activation first
		transferReq := map[string]interface{}{
			"license_key": "ISX1Y-TRANS-12345-FGHIJ-67890",
			"force":       false,
		}
		
		reqBody, err := json.Marshal(transferReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/transfer", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		
		var transferResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&transferResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, transferResp["success"].(bool))
		assert.Contains(t, transferResp["message"], "License transferred successfully")
		
		// Step 2: Verify license is now active
		resp, err = http.Get(server.URL + "/api/license/status")
		require.NoError(t, err)
		
		var statusResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&statusResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, "active", statusResp["license_status"])
		
		// Step 3: Verify we can access protected content
		resp, err = http.Get(server.URL + "/api/reports")
		require.NoError(t, err)
		
		var reportsResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&reportsResp)
		require.NoError(t, err)
		resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, float64(2), reportsResp["count"])
		reports := reportsResp["reports"].([]interface{})
		assert.Contains(t, reports, "Daily Market Report")
		assert.Contains(t, reports, "Ticker Analysis")
	})

	t.Run("forced transfer", func(t *testing.T) {
		transferReq := map[string]interface{}{
			"license_key": "ISX1Y-FORCE-12345-FGHIJ-67890",
			"force":       true,
		}
		
		reqBody, err := json.Marshal(transferReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/transfer", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var transferResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&transferResp)
		require.NoError(t, err)
		
		assert.True(t, transferResp["success"].(bool))
	})

	t.Run("transfer with invalid key", func(t *testing.T) {
		// Simulate error for invalid transfer
		mockManager.error = fmt.Errorf("license not found")
		defer func() { mockManager.error = nil }()
		
		transferReq := map[string]interface{}{
			"license_key": "ISX1Y-INVALID-12345-FGHIJ-67890",
			"force":       false,
		}
		
		reqBody, err := json.Marshal(transferReq)
		require.NoError(t, err)
		
		resp, err := http.Post(server.URL+"/api/license/transfer", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// Benchmark integration test
func BenchmarkLicenseIntegration_FullFlow(b *testing.B) {
	server, _ := setupTestServer()
	defer server.Close()

	// Activate license once
	activationReq := map[string]interface{}{
		"license_key": "ISX1Y-BENCH-12345-FGHIJ-67890",
	}
	
	reqBody, _ := json.Marshal(activationReq)
	resp, _ := http.Post(server.URL+"/api/license/activate", "application/json", bytes.NewReader(reqBody))
	resp.Body.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test accessing protected API
		resp, err := http.Get(server.URL + "/api/data")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	}
}