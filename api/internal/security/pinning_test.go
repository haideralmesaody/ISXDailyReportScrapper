package security

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewCertificatePinner tests certificate pinner initialization
func TestNewCertificatePinner(t *testing.T) {
	tests := []struct {
		name   string
		config *PinningConfig
	}{
		{
			name:   "default_config",
			config: nil, // Should use default
		},
		{
			name:   "custom_config",
			config: &PinningConfig{
				StrictMode:  true,
				AllowBackup: false,
				PinnedCerts: map[string][]string{
					"custom.example.com": {"abcd1234"},
				},
				ConnTimeout:      5 * time.Second,
				HandshakeTimeout: 3 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pinner := NewCertificatePinner(tt.config)
			
			if pinner == nil {
				t.Fatal("Pinner should not be nil")
			}
			
			if pinner.pinnedHashes == nil {
				t.Error("PinnedHashes should not be nil")
			}
			
			// Should have Google APIs pins loaded
			if len(pinner.pinnedHashes) == 0 {
				t.Error("Should have default Google APIs pins loaded")
			}
			
			// Check that Google APIs pins are present
			if _, exists := pinner.pinnedHashes["sheets.googleapis.com"]; !exists {
				t.Error("Should have sheets.googleapis.com pins")
			}
			
			if tt.config != nil {
				// Check custom configuration was applied
				if pinner.strictMode != tt.config.StrictMode {
					t.Errorf("Expected strictMode %v, got %v", tt.config.StrictMode, pinner.strictMode)
				}
				
				if pinner.allowBackup != tt.config.AllowBackup {
					t.Errorf("Expected allowBackup %v, got %v", tt.config.AllowBackup, pinner.allowBackup)
				}
				
				// Check custom pins were added
				if customPins, exists := pinner.pinnedHashes["custom.example.com"]; !exists || len(customPins) == 0 {
					t.Error("Custom pins should be added")
				}
			}
		})
	}
}

// TestDefaultPinningConfig tests default configuration
func TestDefaultPinningConfig(t *testing.T) {
	config := DefaultPinningConfig()
	
	if config == nil {
		t.Fatal("Config should not be nil")
	}
	
	if config.ConnTimeout <= 0 {
		t.Error("ConnTimeout should be positive")
	}
	
	if config.HandshakeTimeout <= 0 {
		t.Error("HandshakeTimeout should be positive")
	}
	
	if config.PinnedCerts == nil {
		t.Error("PinnedCerts should not be nil")
	}
	
	// Check default values
	expectedDefaults := map[string]interface{}{
		"StrictMode":       false,
		"AllowBackup":      true,
		"ConnTimeout":      30 * time.Second,
		"HandshakeTimeout": 10 * time.Second,
	}
	
	if config.StrictMode != expectedDefaults["StrictMode"] {
		t.Errorf("Expected StrictMode %v, got %v", expectedDefaults["StrictMode"], config.StrictMode)
	}
	
	if config.AllowBackup != expectedDefaults["AllowBackup"] {
		t.Errorf("Expected AllowBackup %v, got %v", expectedDefaults["AllowBackup"], config.AllowBackup)
	}
	
	if config.ConnTimeout != expectedDefaults["ConnTimeout"] {
		t.Errorf("Expected ConnTimeout %v, got %v", expectedDefaults["ConnTimeout"], config.ConnTimeout)
	}
	
	if config.HandshakeTimeout != expectedDefaults["HandshakeTimeout"] {
		t.Errorf("Expected HandshakeTimeout %v, got %v", expectedDefaults["HandshakeTimeout"], config.HandshakeTimeout)
	}
}

// TestCreateSecureHTTPClient tests HTTP client creation with certificate pinning
func TestCreateSecureHTTPClient(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	tests := []struct {
		name   string
		config *PinningConfig
	}{
		{
			name:   "default_config",
			config: nil,
		},
		{
			name: "custom_config",
			config: &PinningConfig{
				ConnTimeout:      5 * time.Second,
				HandshakeTimeout: 3 * time.Second,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := pinner.CreateSecureHTTPClient(tt.config)
			
			if client == nil {
				t.Fatal("Client should not be nil")
			}
			
			if client.Transport == nil {
				t.Fatal("Transport should not be nil")
			}
			
			transport, ok := client.Transport.(*http.Transport)
			if !ok {
				t.Fatal("Transport should be *http.Transport")
			}
			
			if transport.TLSClientConfig == nil {
				t.Fatal("TLS config should not be nil")
			}
			
			// Check TLS configuration
			tlsConfig := transport.TLSClientConfig
			if tlsConfig.MinVersion != tls.VersionTLS12 {
				t.Errorf("Expected TLS 1.2 minimum, got %d", tlsConfig.MinVersion)
			}
			
			if !tlsConfig.PreferServerCipherSuites {
				t.Error("Should prefer server cipher suites")
			}
			
			if len(tlsConfig.CipherSuites) == 0 {
				t.Error("Should have cipher suites configured")
			}
			
			if tlsConfig.VerifyPeerCertificate == nil {
				t.Error("Should have certificate verification function")
			}
			
			// Check timeout configuration
			if tt.config != nil {
				if client.Timeout != tt.config.ConnTimeout {
					t.Errorf("Expected timeout %v, got %v", tt.config.ConnTimeout, client.Timeout)
				}
				
				if transport.TLSHandshakeTimeout != tt.config.HandshakeTimeout {
					t.Errorf("Expected handshake timeout %v, got %v", tt.config.HandshakeTimeout, transport.TLSHandshakeTimeout)
				}
			}
		})
	}
}

// TestFindMatchingPins tests hostname pin matching logic
func TestFindMatchingPins(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	// Add some test pins
	testPins := map[string][]string{
		"exact.example.com": {"pin1", "pin2"},
		"*.wildcard.com":    {"wildpin1", "wildpin2"},
		"sub.wildcard.com":  {"subpin1"},
	}
	
	for hostname, pins := range testPins {
		pinner.pinnedHashes[hostname] = pins
	}
	
	tests := []struct {
		name           string
		hostname       string
		expectedPins   []string
		shouldHavePins bool
	}{
		{
			name:           "exact_match",
			hostname:       "exact.example.com",
			expectedPins:   []string{"pin1", "pin2"},
			shouldHavePins: true,
		},
		{
			name:           "wildcard_match",
			hostname:       "test.wildcard.com",
			expectedPins:   []string{"wildpin1", "wildpin2"},
			shouldHavePins: true,
		},
		{
			name:           "subdomain_exact_match",
			hostname:       "sub.wildcard.com",
			expectedPins:   []string{"subpin1"},
			shouldHavePins: true,
		},
		{
			name:           "no_match",
			hostname:       "nomatch.example.org",
			expectedPins:   nil,
			shouldHavePins: false,
		},
		{
			name:           "partial_match_no_wildcard",
			hostname:       "notexact.example.com",
			expectedPins:   nil,
			shouldHavePins: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pins := pinner.findMatchingPins(tt.hostname)
			
			if tt.shouldHavePins {
				if len(pins) == 0 {
					t.Errorf("Expected pins for %s, got none", tt.hostname)
					return
				}
				
				if len(pins) != len(tt.expectedPins) {
					t.Errorf("Expected %d pins, got %d", len(tt.expectedPins), len(pins))
				}
				
				// Check that all expected pins are present
				for _, expectedPin := range tt.expectedPins {
					found := false
					for _, pin := range pins {
						if pin == expectedPin {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected pin %s not found", expectedPin)
					}
				}
			} else {
				if len(pins) > 0 {
					t.Errorf("Expected no pins for %s, got %v", tt.hostname, pins)
				}
			}
		})
	}
}

// TestCalculateSPKIHash tests SPKI hash calculation
func TestCalculateSPKIHash(t *testing.T) {
	// Create a test server with a self-signed certificate
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Get the server's certificate
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	resp.Body.Close()
	
	// Extract certificate from response
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		t.Fatal("No certificates in TLS connection")
	}
	
	cert := resp.TLS.PeerCertificates[0]
	hash := calculateSPKIHash(cert)
	
	// Validate hash format
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
	
	// Should be valid hex
	if err := ValidateIntegrityConfig(hash); err != nil {
		t.Errorf("Generated SPKI hash is not valid hex: %v", err)
	}
	
	// Should be consistent
	hash2 := calculateSPKIHash(cert)
	if hash != hash2 {
		t.Error("SPKI hash calculation should be consistent")
	}
	
	t.Logf("Generated SPKI hash: %s", hash)
}

// TestAddCustomPin tests adding custom certificate pins
func TestAddCustomPin(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	tests := []struct {
		name        string
		hostname    string
		certHash    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_pin",
			hostname:    "example.com",
			certHash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectError: false,
		},
		{
			name:        "empty_hostname",
			hostname:    "",
			certHash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectError: true,
			errorMsg:    "hostname cannot be empty",
		},
		{
			name:        "short_hash",
			hostname:    "example.com",
			certHash:    "short",
			expectError: true,
			errorMsg:    "certificate hash must be 64 characters",
		},
		{
			name:        "invalid_hex",
			hostname:    "example.com",
			certHash:    "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			expectError: true,
			errorMsg:    "certificate hash must be valid hex",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pinner.AddCustomPin(tt.hostname, tt.certHash)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			// Verify pin was added
			pins := pinner.pinnedHashes[tt.hostname]
			found := false
			for _, pin := range pins {
				if strings.EqualFold(pin, tt.certHash) {
					found = true
					break
				}
			}
			if !found {
				t.Error("Pin was not added to hostname")
			}
		})
	}
}

// TestRemovePin tests removing certificate pins
func TestRemovePin(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	// Add some test pins
	hostname := "test.example.com"
	pin1 := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	pin2 := "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
	
	pinner.AddCustomPin(hostname, pin1)
	pinner.AddCustomPin(hostname, pin2)
	
	// Verify pins were added
	pins := pinner.pinnedHashes[hostname]
	if len(pins) != 2 {
		t.Fatalf("Expected 2 pins, got %d", len(pins))
	}
	
	// Remove one pin
	pinner.RemovePin(hostname, pin1)
	pins = pinner.pinnedHashes[hostname]
	if len(pins) != 1 {
		t.Errorf("Expected 1 pin after removal, got %d", len(pins))
	}
	
	// Check the right pin was removed
	if !strings.EqualFold(pins[0], pin2) {
		t.Error("Wrong pin remained after removal")
	}
	
	// Remove the last pin
	pinner.RemovePin(hostname, pin2)
	_, exists := pinner.pinnedHashes[hostname]
	if exists {
		t.Error("Hostname should be removed when no pins remain")
	}
	
	// Remove from non-existent hostname (should not panic)
	pinner.RemovePin("nonexistent.com", pin1)
}

// TestGetPinnedHashes tests retrieving pinned hashes
func TestGetPinnedHashes(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	// Add custom pin
	customHostname := "custom.example.com"
	customPin := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	pinner.AddCustomPin(customHostname, customPin)
	
	hashes := pinner.GetPinnedHashes()
	
	if hashes == nil {
		t.Fatal("Hashes should not be nil")
	}
	
	// Should include Google APIs pins
	if _, exists := hashes["sheets.googleapis.com"]; !exists {
		t.Error("Should include default Google APIs pins")
	}
	
	// Should include custom pin
	if customPins, exists := hashes[customHostname]; !exists {
		t.Error("Should include custom pin")
	} else if len(customPins) == 0 || !strings.EqualFold(customPins[0], customPin) {
		t.Error("Custom pin not correctly included")
	}
	
	// Should be a copy (modifying returned map shouldn't affect original)
	hashes["test.modification.com"] = []string{"test"}
	originalHashes := pinner.GetPinnedHashes()
	if _, exists := originalHashes["test.modification.com"]; exists {
		t.Error("Returned hashes should be a copy")
	}
}

// TestUpdateGoogleAPIsPins tests updating Google APIs pins
func TestUpdateGoogleAPIsPins(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	newPins := map[string][]string{
		"new-api.googleapis.com": {"newpin1", "newpin2"},
		"sheets.googleapis.com":  {"updatedpin1"}, // Override existing
	}
	
	pinner.UpdateGoogleAPIsPins(newPins)
	
	// Check new pins were added
	if pins, exists := pinner.pinnedHashes["new-api.googleapis.com"]; !exists {
		t.Error("New API pins should be added")
	} else if len(pins) != 2 {
		t.Errorf("Expected 2 new pins, got %d", len(pins))
	}
	
	// Check existing pins were updated
	if pins, exists := pinner.pinnedHashes["sheets.googleapis.com"]; !exists {
		t.Error("Sheets API pins should exist")
	} else if len(pins) != 1 || pins[0] != "updatedpin1" {
		t.Error("Sheets API pins should be updated")
	}
}

// TestGeneratePinningReport tests pinning report generation
func TestGeneratePinningReport(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	// Add a custom pin for testing
	testHostname := "test.example.com"
	testPin := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	pinner.AddCustomPin(testHostname, testPin)
	
	reports := pinner.GeneratePinningReport()
	
	if len(reports) == 0 {
		t.Error("Should generate reports for pinned hostnames")
	}
	
	// Find our test hostname report
	var testReport *PinningReport
	for _, report := range reports {
		if report.Hostname == testHostname {
			testReport = report
			break
		}
	}
	
	if testReport == nil {
		t.Error("Should include report for test hostname")
		return
	}
	
	// Check report structure
	if testReport.Hostname != testHostname {
		t.Error("Report hostname should match")
	}
	
	if len(testReport.PinnedHashes) == 0 {
		t.Error("Report should include pinned hashes")
	}
	
	if testReport.LastVerified.IsZero() {
		t.Error("LastVerified should be set")
	}
	
	// Error is expected since we can't connect to test.example.com
	if testReport.Error == "" {
		t.Log("No error in report (unexpected for test.example.com)")
	}
	
	if testReport.PinMatchFound {
		t.Log("Pin match found (unexpected for test.example.com)")
	}
	
	t.Logf("Generated report for %s: %+v", testHostname, testReport)
}

// TestVerifySingleHostname tests single hostname verification
func TestVerifySingleHostname(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	tests := []struct {
		name        string
		hostname    string
		expectError bool
	}{
		{
			name:        "invalid_hostname",
			hostname:    "invalid.nonexistent.tld.that.should.not.exist",
			expectError: true,
		},
		{
			name:        "empty_hostname",
			hostname:    "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pinner.verifySingleHostname(tt.hostname)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestVerifyCurrentCertificates tests current certificate verification
func TestVerifyCurrentCertificates(t *testing.T) {
	pinner := NewCertificatePinner(nil)
	
	// Add some test hostnames
	testHostnames := []string{
		"invalid.test.hostname.that.should.fail",
		"another.invalid.hostname",
	}
	
	for _, hostname := range testHostnames {
		pinner.AddCustomPin(hostname, "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	}
	
	results := pinner.VerifyCurrentCertificates()
	
	if len(results) == 0 {
		t.Error("Should return results for pinned hostnames")
	}
	
	// Check that we have results for our test hostnames
	for _, hostname := range testHostnames {
		if _, exists := results[hostname]; !exists {
			t.Errorf("Should have result for %s", hostname)
		} else {
			// Should have error for invalid hostnames
			if results[hostname] == nil {
				t.Errorf("Expected error for invalid hostname %s", hostname)
			}
		}
	}
	
	t.Logf("Verification results: %v", results)
}

// TestPinningConfigStructure tests PinningConfig structure
func TestPinningConfigStructure(t *testing.T) {
	config := &PinningConfig{
		StrictMode:       true,
		AllowBackup:      false,
		PinnedCerts:      map[string][]string{"test.com": {"pin1"}},
		ConnTimeout:      15 * time.Second,
		HandshakeTimeout: 5 * time.Second,
	}
	
	// Test that all fields are accessible and correct type
	if config.StrictMode != true {
		t.Error("StrictMode should be bool and settable")
	}
	
	if config.AllowBackup != false {
		t.Error("AllowBackup should be bool and settable")
	}
	
	if config.PinnedCerts == nil || len(config.PinnedCerts) == 0 {
		t.Error("PinnedCerts should be map and settable")
	}
	
	if config.ConnTimeout != 15*time.Second {
		t.Error("ConnTimeout should be time.Duration and settable")
	}
	
	if config.HandshakeTimeout != 5*time.Second {
		t.Error("HandshakeTimeout should be time.Duration and settable")
	}
}

// TestGoogleAPIsPins tests the predefined Google APIs pins
func TestGoogleAPIsPins(t *testing.T) {
	expectedHosts := []string{
		"sheets.googleapis.com",
		"accounts.google.com",
		"oauth2.googleapis.com",
		"upload.video.google.com",
		"www.googleapis.com",
		"*.googleapis.com",
	}
	
	for _, host := range expectedHosts {
		if pins, exists := GoogleAPIsPins[host]; !exists {
			t.Errorf("Should have pins for %s", host)
		} else if len(pins) == 0 {
			t.Errorf("Should have at least one pin for %s", host)
		} else {
			// Validate pin format
			for _, pin := range pins {
				if len(pin) != 64 {
					t.Errorf("Pin for %s should be 64 characters, got %d", host, len(pin))
				}
				if err := ValidateIntegrityConfig(pin); err != nil {
					t.Errorf("Pin for %s is not valid hex: %v", host, err)
				}
			}
		}
	}
	
	t.Logf("Google APIs pins validated for %d hosts", len(expectedHosts))
}

// BenchmarkSPKIHashCalculation benchmarks SPKI hash calculation
func BenchmarkSPKIHashCalculation(b *testing.B) {
	// Create a test server to get a certificate
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Get certificate
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		b.Fatalf("Failed to get test certificate: %v", err)
	}
	resp.Body.Close()
	
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		b.Fatal("No certificates available")
	}
	
	cert := resp.TLS.PeerCertificates[0]
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSPKIHash(cert)
	}
}

// BenchmarkPinMatching benchmarks pin matching logic
func BenchmarkPinMatching(b *testing.B) {
	pinner := NewCertificatePinner(nil)
	
	// Add many pins for benchmarking
	for i := 0; i < 100; i++ {
		hostname := fmt.Sprintf("test%d.example.com", i)
		pin := fmt.Sprintf("%064d", i) // 64-character string
		pinner.AddCustomPin(hostname, pin)
	}
	
	testHostnames := []string{
		"test50.example.com",
		"sheets.googleapis.com",
		"nomatch.example.org",
		"test.wildcard.com",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hostname := testHostnames[i%len(testHostnames)]
		_ = pinner.findMatchingPins(hostname)
	}
}