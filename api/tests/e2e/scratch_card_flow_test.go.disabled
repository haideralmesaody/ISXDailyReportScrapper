package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"isxcli/internal/license"
	"isxcli/internal/security"
	"isxcli/internal/services"
	"isxcli/internal/transport/http"
)

// ScratchCardFlowTestSuite tests the complete scratch card activation flow
type ScratchCardFlowTestSuite struct {
	suite.Suite
	tempDir         string
	licenseManager  *license.Manager
	licenseService  *services.LicenseService
	httpHandler     *http.LicenseHandler
	fingerprintMgr  *security.FingerprintManager
	testServer      *httptest.Server
	appsScriptURL   string
}

// SetupSuite initializes the test suite
func (s *ScratchCardFlowTestSuite) SetupSuite() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "scratch_card_e2e_*")
	s.Require().NoError(err)

	// Set up license manager
	licenseFile := filepath.Join(s.tempDir, "test_license.dat")
	s.licenseManager, err = license.NewManager(licenseFile)
	s.Require().NoError(err)

	// Set up fingerprint manager
	s.fingerprintMgr = security.NewFingerprintManager()

	// Set up license service
	s.licenseService = services.NewLicenseService(s.licenseManager)

	// Set up HTTP handler
	s.httpHandler = http.NewLicenseHandler(s.licenseService)

	// Set up mock Apps Script server
	s.setupMockAppsScriptServer()
}

// TearDownSuite cleans up after tests
func (s *ScratchCardFlowTestSuite) TearDownSuite() {
	if s.testServer != nil {
		s.testServer.Close()
	}
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// SetupTest prepares each individual test
func (s *ScratchCardFlowTestSuite) SetupTest() {
	// Clear any existing license
	licenseFile := filepath.Join(s.tempDir, "test_license.dat")
	if _, err := os.Stat(licenseFile); err == nil {
		os.Remove(licenseFile)
	}

	// Recreate license manager
	var err error
	s.licenseManager, err = license.NewManager(licenseFile)
	s.Require().NoError(err)
	
	s.licenseService = services.NewLicenseService(s.licenseManager)
	s.httpHandler = http.NewLicenseHandler(s.licenseService)
}

// setupMockAppsScriptServer creates a mock Google Apps Script server
func (s *ScratchCardFlowTestSuite) setupMockAppsScriptServer() {
	activatedLicenses := make(map[string]map[string]interface{})
	var mutex sync.RWMutex

	s.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		action, ok := requestBody["action"].(string)
		if !ok {
			action = "activateScratchCard" // Default for testing
		}

		switch action {
		case "activateScratchCard":
			s.handleScratchCardActivation(w, requestBody, activatedLicenses, &mutex)
		case "checkUniqueness":
			s.handleUniquenessCheck(w, requestBody, activatedLicenses, &mutex)
		case "deactivateLicense":
			s.handleDeactivation(w, requestBody, activatedLicenses, &mutex)
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Unknown action: " + action,
			})
		}
	}))

	s.appsScriptURL = s.testServer.URL
}

// handleScratchCardActivation handles scratch card activation requests
func (s *ScratchCardFlowTestSuite) handleScratchCardActivation(
	w http.ResponseWriter,
	requestBody map[string]interface{},
	activatedLicenses map[string]map[string]interface{},
	mutex *sync.RWMutex,
) {
	licenseKey, ok := requestBody["licenseKey"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Missing license key",
		})
		return
	}

	deviceFingerprint, ok := requestBody["deviceFingerprint"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Missing device fingerprint",
		})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Check if already activated
	if existing, exists := activatedLicenses[licenseKey]; exists {
		if existing["deviceFingerprint"] != deviceFingerprint {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "License already activated on another device",
			})
			return
		}

		// Same device re-activation
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"activationId": existing["activationId"],
			"message":      "License already activated on this device",
		})
		return
	}

	// Validate license format
	if !s.isValidScratchCardFormat(licenseKey) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid license key format",
		})
		return
	}

	// New activation
	activationID := fmt.Sprintf("act_%s_%d", licenseKey, time.Now().Unix())
	activatedLicenses[licenseKey] = map[string]interface{}{
		"activationId":      activationID,
		"deviceFingerprint": deviceFingerprint,
		"activatedAt":       time.Now().Unix(),
		"duration":          "1m", // Default duration
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"activationId": activationID,
		"message":      "License activated successfully",
	})
}

// handleUniquenessCheck handles license uniqueness verification
func (s *ScratchCardFlowTestSuite) handleUniquenessCheck(
	w http.ResponseWriter,
	requestBody map[string]interface{},
	activatedLicenses map[string]map[string]interface{},
	mutex *sync.RWMutex,
) {
	codes, ok := requestBody["codes"].([]interface{})
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Missing codes array",
		})
		return
	}

	mutex.RLock()
	defer mutex.RUnlock()

	var duplicates []string
	for _, code := range codes {
		if codeStr, ok := code.(string); ok {
			if _, exists := activatedLicenses[codeStr]; exists {
				duplicates = append(duplicates, codeStr)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"duplicates": duplicates,
	})
}

// handleDeactivation handles license deactivation requests
func (s *ScratchCardFlowTestSuite) handleDeactivation(
	w http.ResponseWriter,
	requestBody map[string]interface{},
	activatedLicenses map[string]map[string]interface{},
	mutex *sync.RWMutex,
) {
	licenseKey, ok := requestBody["licenseKey"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Missing license key",
		})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := activatedLicenses[licenseKey]; !exists {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "License not found or not activated",
		})
		return
	}

	delete(activatedLicenses, licenseKey)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "License deactivated successfully",
	})
}

// isValidScratchCardFormat validates scratch card format
func (s *ScratchCardFlowTestSuite) isValidScratchCardFormat(licenseKey string) bool {
	return license.ValidateScratchCardFormat(licenseKey) == nil
}

// TestCompleteActivationFlow tests the complete scratch card activation process
func (s *ScratchCardFlowTestSuite) TestCompleteActivationFlow() {
	ctx := context.Background()
	licenseKey := "ISX-1234-5678-90AB"

	// Get device fingerprint
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Step 1: Verify license is not activated
	info, err := s.licenseManager.GetLicenseInfo(ctx)
	s.Assert().Error(err) // Should not exist initially

	// Step 2: Activate scratch card
	activatedInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(activatedInfo.IsValid)
	s.Assert().NotEmpty(activatedInfo.ActivationID)
	s.Assert().Equal(fingerprint.Fingerprint, activatedInfo.DeviceFingerprint)

	// Step 3: Verify license is now activated and valid
	info, err = s.licenseManager.GetLicenseInfo(ctx)
	s.Require().NoError(err)
	s.Assert().True(info.IsValid)
	s.Assert().Equal(licenseKey, info.LicenseKey)
	s.Assert().Equal(activatedInfo.ActivationID, info.ActivationID)

	// Step 4: Validate license
	valid, err := s.licenseManager.ValidateLicenseWithContext(ctx)
	s.Require().NoError(err)
	s.Assert().True(valid)

	// Step 5: Try to activate again with same fingerprint (should succeed as existing)
	reactivateInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(reactivateInfo.IsValid)
	s.Assert().Equal(activatedInfo.ActivationID, reactivateInfo.ActivationID)

	// Step 6: Try to activate with different fingerprint (should fail)
	differentFingerprint := "different_device_fingerprint_hash"
	failedInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, differentFingerprint)
	s.Assert().Error(err)
	s.Assert().False(failedInfo.IsValid)
	s.Assert().Contains(err.Error(), "already activated")
}

// TestNetworkFailureRecovery tests handling of network failures during activation
func (s *ScratchCardFlowTestSuite) TestNetworkFailureRecovery() {
	ctx := context.Background()
	licenseKey := "ISX-FAIL-TEST-001"

	// Generate fingerprint
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Temporarily close the server to simulate network failure
	s.testServer.Close()

	// Try to activate - should fail due to network error
	_, err = s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "connection refused")

	// Restart the server
	s.setupMockAppsScriptServer()

	// Now activation should succeed
	activatedInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(activatedInfo.IsValid)
	s.Assert().NotEmpty(activatedInfo.ActivationID)
}

// TestDeviceChangeScenario tests what happens when device changes
func (s *ScratchCardFlowTestSuite) TestDeviceChangeScenario() {
	ctx := context.Background()
	licenseKey := "ISX-DEVICE-CHANGE-001"

	// Original device activation
	originalFingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Activate on original device
	originalInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, originalFingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(originalInfo.IsValid)

	// Simulate device change by creating different fingerprint
	// In real scenario, this would be automatically detected
	newDeviceFingerprint := "new_device_fingerprint_after_hardware_change"

	// Try to validate with new device fingerprint - should fail
	isValid := s.licenseManager.validateDeviceBinding(newDeviceFingerprint, originalInfo)
	s.Assert().False(isValid, "License should not be valid on different device")

	// Try to activate with new device - should fail
	newDeviceInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, newDeviceFingerprint)
	s.Assert().Error(err)
	s.Assert().False(newDeviceInfo.IsValid)
	s.Assert().Contains(err.Error(), "already activated")
}

// TestLicenseExpiryFlow tests license expiry behavior
func (s *ScratchCardFlowTestSuite) TestLicenseExpiryFlow() {
	ctx := context.Background()
	licenseKey := "ISX-EXPIRY-TEST-001"

	// Generate fingerprint
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Activate license
	activatedInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(activatedInfo.IsValid)

	// Check expiry date is set correctly (1 month from now by default)
	expectedExpiry := time.Now().AddDate(0, 1, 1) // +1 month +1 day
	timeDiff := activatedInfo.ExpiryDate.Sub(expectedExpiry)
	s.Assert().Less(timeDiff, time.Hour, "Expiry date should be approximately 1 month and 1 day from now")

	// Validate that license is currently valid
	valid, err := s.licenseManager.ValidateLicenseWithContext(ctx)
	s.Require().NoError(err)
	s.Assert().True(valid)

	// Test expiry calculation for different durations
	testCases := []struct {
		duration string
		months   int
		years    int
	}{
		{"1m", 1, 0},
		{"3m", 3, 0},
		{"6m", 6, 0},
		{"1y", 0, 1},
	}

	for _, tc := range testCases {
		s.T().Run(fmt.Sprintf("duration_%s", tc.duration), func(t *testing.T) {
			expiry := s.licenseManager.calculateExpiryDateFromDuration(tc.duration)
			expected := time.Now().AddDate(tc.years, tc.months, 1) // +1 day
			
			assert.Equal(t, expected.Year(), expiry.Year())
			assert.Equal(t, expected.Month(), expiry.Month())
			assert.Equal(t, expected.Day(), expiry.Day())
		})
	}
}

// TestConcurrentActivationAttempts tests concurrent activation attempts
func (s *ScratchCardFlowTestSuite) TestConcurrentActivationAttempts() {
	ctx := context.Background()
	licenseKey := "ISX-CONCURRENT-001"
	concurrentAttempts := 10

	var wg sync.WaitGroup
	var successCount, errorCount int32
	var mutex sync.Mutex
	results := make([]*license.LicenseInfo, concurrentAttempts)
	errors := make([]error, concurrentAttempts)

	// Generate different fingerprints for each attempt
	fingerprints := make([]string, concurrentAttempts)
	for i := 0; i < concurrentAttempts; i++ {
		fingerprints[i] = fmt.Sprintf("device_%d_fingerprint_hash", i)
	}

	// Launch concurrent activation attempts
	for i := 0; i < concurrentAttempts; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			result, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprints[index])
			
			mutex.Lock()
			results[index] = result
			errors[index] = err
			if err != nil {
				errorCount++
			} else if result.IsValid {
				successCount++
			}
			mutex.Unlock()
		}(i)
	}

	wg.Wait()

	// Only one activation should succeed (first one wins)
	s.Assert().Equal(int32(1), successCount, "Exactly one concurrent activation should succeed")
	s.Assert().Equal(int32(concurrentAttempts-1), errorCount, "All other attempts should fail")

	// Verify the successful activation
	var successfulResult *license.LicenseInfo
	for i, result := range results {
		if errors[i] == nil && result.IsValid {
			successfulResult = result
			break
		}
	}

	s.Require().NotNil(successfulResult, "Should have one successful activation")
	s.Assert().NotEmpty(successfulResult.ActivationID)
	s.Assert().Equal(licenseKey, successfulResult.LicenseKey)
}

// TestHTTPEndpointIntegration tests HTTP API integration
func (s *ScratchCardFlowTestSuite) TestHTTPEndpointIntegration() {
	licenseKey := "ISX-HTTP-TEST-001"

	// Generate fingerprint
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Create activation request
	activationReq := map[string]interface{}{
		"licenseKey":        licenseKey,
		"deviceFingerprint": fingerprint.Fingerprint,
	}

	requestBody, err := json.Marshal(activationReq)
	s.Require().NoError(err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/license/activate", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	s.httpHandler.ServeHTTP(rr, req)

	// Verify response
	s.Assert().Equal(http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	s.Require().NoError(err)

	s.Assert().True(response["success"].(bool))
	s.Assert().NotEmpty(response["activationId"])
	s.Assert().Equal(licenseKey, response["licenseKey"])

	// Test status endpoint
	statusReq := httptest.NewRequest("GET", "/api/v1/license/status", nil)
	statusRr := httptest.NewRecorder()

	s.httpHandler.ServeHTTP(statusRr, statusReq)
	s.Assert().Equal(http.StatusOK, statusRr.Code)

	var statusResponse map[string]interface{}
	err = json.NewDecoder(statusRr.Body).Decode(&statusResponse)
	s.Require().NoError(err)

	s.Assert().True(statusResponse["isValid"].(bool))
	s.Assert().Equal(licenseKey, statusResponse["licenseKey"])
}

// TestInvalidLicenseKeyFormats tests various invalid license key formats
func (s *ScratchCardFlowTestSuite) TestInvalidLicenseKeyFormats() {
	ctx := context.Background()
	
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	invalidFormats := []struct {
		key         string
		description string
	}{
		{"", "empty key"},
		{"ISX", "too short"},
		{"ABC-1234-5678-90AB", "wrong prefix"},
		{"ISX-12-34-56", "wrong segment length"},
		{"ISX-1!23-4567-890A", "invalid characters"},
		{"ISX-1234-5678-90AB-EXTRA", "too long"},
		{"isx-1234-5678-90ab", "lowercase (should be normalized first)"},
		{"ISX'; DROP TABLE licenses; --", "SQL injection"},
		{"ISX<script>alert(1)</script>", "XSS attempt"},
	}

	for _, test := range invalidFormats {
		s.T().Run(test.description, func(t *testing.T) {
			_, err := s.licenseManager.ActivateScratchCard(ctx, test.key, fingerprint.Fingerprint)
			assert.Error(t, err, "Invalid format should be rejected: %s", test.description)
		})
	}
}

// TestPersistenceAcrossRestarts tests that license data persists across manager restarts
func (s *ScratchCardFlowTestSuite) TestPersistenceAcrossRestarts() {
	ctx := context.Background()
	licenseKey := "ISX-PERSIST-001"

	// Generate fingerprint
	fingerprint, err := s.fingerprintMgr.GenerateFingerprint()
	s.Require().NoError(err)

	// Activate license
	originalInfo, err := s.licenseManager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	s.Require().NoError(err)
	s.Assert().True(originalInfo.IsValid)

	// Store the license file path
	licenseFile := filepath.Join(s.tempDir, "test_license.dat")

	// Create new manager instance (simulating restart)
	newManager, err := license.NewManager(licenseFile)
	s.Require().NoError(err)

	// Verify license is still valid after restart
	restoredInfo, err := newManager.GetLicenseInfo(ctx)
	s.Require().NoError(err)
	s.Assert().True(restoredInfo.IsValid)
	s.Assert().Equal(originalInfo.LicenseKey, restoredInfo.LicenseKey)
	s.Assert().Equal(originalInfo.ActivationID, restoredInfo.ActivationID)
	s.Assert().Equal(originalInfo.DeviceFingerprint, restoredInfo.DeviceFingerprint)

	// Verify license validation still works
	valid, err := newManager.ValidateLicenseWithContext(ctx)
	s.Require().NoError(err)
	s.Assert().True(valid)
}

// Run the test suite
func TestScratchCardFlowTestSuite(t *testing.T) {
	suite.Run(t, new(ScratchCardFlowTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkCompleteActivationFlow(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "bench_scratch_card_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	licenseFile := filepath.Join(tempDir, "bench_license.dat")
	manager, err := license.NewManager(licenseFile)
	require.NoError(b, err)

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		licenseKey := fmt.Sprintf("ISX-BENCH-%04d-%03d", i%10000, i%1000)
		
		_, err := manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Unexpected error in iteration %d: %v", i, err)
		}
		
		// Clean up for next iteration
		os.Remove(licenseFile)
		manager, _ = license.NewManager(licenseFile)
	}
}