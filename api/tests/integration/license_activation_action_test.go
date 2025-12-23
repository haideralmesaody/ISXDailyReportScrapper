package integration

import (
	"context"
	"strings"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/security"
)

// TestLicenseActivationActionName verifies that the correct action name is sent to Apps Script
func TestLicenseActivationActionName(t *testing.T) {
	// Skip in CI if no Apps Script URL is configured
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Get embedded credentials
	creds := config.GetCredentials()
	if creds.AppsScriptURL == "" {
		t.Skip("Apps Script URL not configured")
	}
	
	// Create security client
	securityConfig := security.DefaultAppsScriptSecurityConfig()
	client := security.NewSecureAppsScriptClient(securityConfig, nil)
	
	// Validate configuration
	err := client.ValidateConfiguration()
	require.NoError(t, err, "Security configuration should be valid")
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Test device fingerprint
	deviceFingerprint := "test-device-integration-" + time.Now().Format("20060102-150405")
	
	t.Run("ActivateAction_ShouldBeRecognized", func(t *testing.T) {
		// Create activation payload with CORRECT action name
		payload := map[string]interface{}{
			"action": license.ActionActivate,
			"code":   "ISX-TEST-INTG-TEST-" + time.Now().Format("1504"),
			"deviceInfo": map[string]interface{}{
				"fingerprint": deviceFingerprint,
				"hostname":    "test-host",
				"os":          "test-os",
			},
		}
		
		// Send request
		resp, err := client.SecureRequest(ctx, creds.AppsScriptURL, payload, deviceFingerprint)
		require.NoError(t, err, "Request should not fail at network level")
		
		// The action should be recognized (not "unknown action")
		if resp.Error != "" {
			assert.False(t, strings.Contains(strings.ToLower(resp.Error), "unknown action"),
				"Action 'activate' should be recognized by Apps Script, got error: %s", resp.Error)
		}
		
		// Log the response for debugging
		t.Logf("Activation response - Success: %v, Error: %s", resp.Success, resp.Error)
		if resp.Data != nil {
			t.Logf("Response data: %+v", resp.Data)
		}
	})
	
	t.Run("OldWrongAction_ShouldFail", func(t *testing.T) {
		// Test with the OLD WRONG action name that was causing the bug
		payload := map[string]interface{}{
			"action": "validateLicense", // The old wrong action
			"code":   "ISX-TEST-WRONG-TEST",
		}
		
		// Send request
		resp, err := client.SecureRequest(ctx, creds.AppsScriptURL, payload, deviceFingerprint)
		
		// Should either error or return unknown action
		if err == nil {
			assert.True(t, strings.Contains(strings.ToLower(resp.Error), "unknown action"),
				"Action 'validateLicense' should be unknown to Apps Script")
			t.Logf("Correctly rejected unknown action: %s", resp.Error)
		}
	})
	
	t.Run("ValidateAction_ShouldBeRecognized", func(t *testing.T) {
		// Test the correct validate action
		payload := map[string]interface{}{
			"action":       license.ActionValidate,
			"code":         "ISX-TEST-VALD-TEST",
			"activationId": "test-activation-id",
		}
		
		// Send request
		resp, err := client.SecureRequest(ctx, creds.AppsScriptURL, payload, deviceFingerprint)
		require.NoError(t, err, "Request should not fail at network level")
		
		// The action should be recognized (not "unknown action")
		if resp.Error != "" {
			// Validate might not be implemented, but it shouldn't be "unknown action"
			isUnknownAction := strings.Contains(strings.ToLower(resp.Error), "unknown action")
			isNotImplemented := strings.Contains(strings.ToLower(resp.Error), "not implemented")
			
			assert.True(t, !isUnknownAction || isNotImplemented,
				"Action 'validate' should either be recognized or explicitly not implemented, got: %s", resp.Error)
		}
	})
}

// TestLicenseManagerActivationAction tests the complete activation flow through the manager
func TestLicenseManagerActivationAction(t *testing.T) {
	// Skip in CI if no Apps Script URL is configured
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a temporary license file path
	tempLicenseFile := t.TempDir() + "/test_license.dat"
	
	// Create license manager
	manager, err := license.NewManager()
	require.NoError(t, err, "Should create license manager")
	
	// Test with a dummy license key
	testLicenseKey := "ISX-MNGR-TEST-ACTN-" + time.Now().Format("1504")
	
	// Attempt activation
	err = manager.ActivateLicense(testLicenseKey)
	
	// We expect it to fail (invalid license), but NOT with "unknown action"
	if err != nil {
		errorMsg := err.Error()
		assert.False(t, strings.Contains(strings.ToLower(errorMsg), "unknown action"),
			"Should not get 'unknown action' error, got: %s", errorMsg)
		
		// It should fail with "invalid license" or similar
		isInvalidLicense := strings.Contains(strings.ToLower(errorMsg), "invalid")
		isNotFound := strings.Contains(strings.ToLower(errorMsg), "not found")
		isServerError := strings.Contains(strings.ToLower(errorMsg), "server")
		
		assert.True(t, isInvalidLicense || isNotFound || isServerError,
			"Should get appropriate error message, got: %s", errorMsg)
		
		t.Logf("Activation correctly failed with: %s", errorMsg)
	}
}

// TestActionConstants verifies all action constants are defined correctly
func TestActionConstants(t *testing.T) {
	// Verify constants match expected values
	assert.Equal(t, "activate", license.ActionActivate)
	assert.Equal(t, "validate", license.ActionValidate)
	assert.Equal(t, "revoke", license.ActionRevoke)
	assert.Equal(t, "checkStatus", license.ActionCheckStatus)
	assert.Equal(t, "ping", license.ActionPing)
	
	// Verify status constants
	assert.Equal(t, "activated", license.StatusActivated)
	assert.Equal(t, "reactivated", license.StatusReactivated)
	assert.Equal(t, "valid", license.StatusValid)
	assert.Equal(t, "expired", license.StatusExpired)
}