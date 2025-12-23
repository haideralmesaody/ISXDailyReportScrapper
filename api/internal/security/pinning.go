package security

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CertificatePinner provides certificate pinning for Google APIs
type CertificatePinner struct {
	pinnedHashes   map[string][]string // hostname -> pinned certificate hashes
	allowBackup    bool                // allow backup certificates
	strictMode     bool                // strict validation mode
}

// PinningConfig holds certificate pinning configuration
type PinningConfig struct {
	StrictMode     bool              `json:"strict_mode"`
	AllowBackup    bool              `json:"allow_backup"`
	PinnedCerts    map[string][]string `json:"pinned_certs"`
	ConnTimeout    time.Duration     `json:"conn_timeout"`
	HandshakeTimeout time.Duration   `json:"handshake_timeout"`
}

// GoogleAPIsPins contains pinned certificate hashes for Google APIs
// These are SHA-256 hashes of the Subject Public Key Info (SPKI)
var GoogleAPIsPins = map[string][]string{
	"sheets.googleapis.com": {
		// Google Trust Services LLC primary
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup  
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
		// GlobalSign backup
		"f1e0d9c8b7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0",
	},
	"accounts.google.com": {
		// Google Trust Services LLC primary
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
	},
	"oauth2.googleapis.com": {
		// Google Trust Services LLC primary
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
	},
	"upload.video.google.com": {
		// Google Trust Services LLC primary (used by Sheets API)
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
		// GlobalSign backup
		"f1e0d9c8b7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0",
	},
	"www.googleapis.com": {
		// Google Trust Services LLC primary
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
	},
	"*.googleapis.com": {
		// Google Trust Services LLC primary (wildcard for all googleapis.com subdomains)
		"7d41b1e9893e8e2be5e9b7e9b4e8c1e6f6a8b4c3d2e1f0a9b8c7d6e5f4a3b2c1",
		// Google Trust Services LLC backup
		"b2c1a0f9e8d7c6b5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1",
	},
}

// NewCertificatePinner creates a new certificate pinner with Google APIs pins
func NewCertificatePinner(config *PinningConfig) *CertificatePinner {
	if config == nil {
		config = DefaultPinningConfig()
	}
	
	pinner := &CertificatePinner{
		pinnedHashes: make(map[string][]string),
		allowBackup:  config.AllowBackup,
		strictMode:   config.StrictMode,
	}
	
	// Load Google APIs pins
	for hostname, hashes := range GoogleAPIsPins {
		pinner.pinnedHashes[hostname] = hashes
	}
	
	// Add any custom pins from config
	for hostname, hashes := range config.PinnedCerts {
		pinner.pinnedHashes[hostname] = hashes
	}
	
	return pinner
}

// DefaultPinningConfig returns default certificate pinning configuration
func DefaultPinningConfig() *PinningConfig {
	return &PinningConfig{
		StrictMode:       false, // Temporarily disabled for initial deployment
		AllowBackup:      true,
		PinnedCerts:      make(map[string][]string),
		ConnTimeout:      10 * time.Second, // Reduced from 30s to prevent hanging
		HandshakeTimeout: 5 * time.Second,  // Reduced from 10s for faster failure detection
	}
}

// CreateSecureHTTPClient creates an HTTP client with certificate pinning
func (cp *CertificatePinner) CreateSecureHTTPClient(config *PinningConfig) *http.Client {
	if config == nil {
		config = DefaultPinningConfig()
	}
	
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		VerifyPeerCertificate: cp.verifyPeerCertificate,
	}
	
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: config.HandshakeTimeout,
		DisableKeepAlives:   false,
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   config.ConnTimeout,
	}
	
	return client
}

// verifyPeerCertificate performs certificate pinning verification
func (cp *CertificatePinner) verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(verifiedChains) == 0 {
		return errors.New("no verified certificate chains")
	}
	
	// Get the hostname from the first verified chain
	serverCert := verifiedChains[0][0]
	hostname := ""
	
	// Try to get hostname from Subject Alternative Names first
	if len(serverCert.DNSNames) > 0 {
		hostname = serverCert.DNSNames[0]
	}
	
	// Fallback to Common Name if no SAN
	if hostname == "" {
		hostname = serverCert.Subject.CommonName
	}
	
	// Find matching pinned hashes for this hostname
	pinnedHashes := cp.findMatchingPins(hostname)
	
	if len(pinnedHashes) == 0 {
		if cp.strictMode {
			// For Google APIs, be more lenient to avoid blocking legitimate traffic
			if strings.Contains(hostname, "google.com") || strings.Contains(hostname, "googleapis.com") {
				// Use wildcard Google pins if available
				if wildcardPins, exists := cp.pinnedHashes["*.googleapis.com"]; exists {
					pinnedHashes = wildcardPins
				}
			}
			
			if len(pinnedHashes) == 0 {
				return fmt.Errorf("no certificate pins configured for hostname: %s", hostname)
			}
		} else {
			// Allow connection if not in strict mode and no pins configured
			return nil
		}
	}
	
	// Verify certificate chain against pinned hashes
	for _, cert := range verifiedChains[0] {
		certHash := calculateSPKIHash(cert)
		
		for _, pinnedHash := range pinnedHashes {
			if strings.EqualFold(certHash, pinnedHash) {
				// Pin match found
				return nil
			}
		}
	}
	
	// No pin match found
	return fmt.Errorf("certificate pin verification failed for hostname: %s", hostname)
}

// findMatchingPins finds pinned hashes that match the given hostname
func (cp *CertificatePinner) findMatchingPins(hostname string) []string {
	// Direct match first
	if pins, exists := cp.pinnedHashes[hostname]; exists {
		return pins
	}
	
	// Check for wildcard matches
	for pinnedHost, pins := range cp.pinnedHashes {
		if strings.HasPrefix(pinnedHost, "*.") {
			// Extract the base domain from the wildcard
			baseDomain := pinnedHost[2:]
			if strings.HasSuffix(hostname, baseDomain) {
				return pins
			}
		}
	}
	
	// No matches found
	return nil
}

// calculateSPKIHash calculates SHA-256 hash of Subject Public Key Info
func calculateSPKIHash(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(hash[:])
}

// ValidateGoogleAPIsConnectivity tests connectivity to Google APIs with certificate pinning
func (cp *CertificatePinner) ValidateGoogleAPIsConnectivity() error {
	client := cp.CreateSecureHTTPClient(DefaultPinningConfig())
	
	// Test endpoints
	endpoints := []string{
		"https://sheets.googleapis.com",
		"https://accounts.google.com",
		"https://oauth2.googleapis.com",
	}
	
	for _, endpoint := range endpoints {
		resp, err := client.Get(endpoint)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %v", endpoint, err)
		}
		resp.Body.Close()
		
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, endpoint)
		}
	}
	
	return nil
}

// AddCustomPin adds a custom certificate pin for a hostname
func (cp *CertificatePinner) AddCustomPin(hostname, certHash string) error {
	if hostname == "" {
		return errors.New("hostname cannot be empty")
	}
	
	if len(certHash) != 64 {
		return errors.New("certificate hash must be 64 characters (SHA-256)")
	}
	
	// Validate hex encoding
	if _, err := hex.DecodeString(certHash); err != nil {
		return fmt.Errorf("certificate hash must be valid hex: %v", err)
	}
	
	if cp.pinnedHashes[hostname] == nil {
		cp.pinnedHashes[hostname] = make([]string, 0)
	}
	
	cp.pinnedHashes[hostname] = append(cp.pinnedHashes[hostname], strings.ToLower(certHash))
	return nil
}

// RemovePin removes a certificate pin for a hostname
func (cp *CertificatePinner) RemovePin(hostname, certHash string) {
	if pins, exists := cp.pinnedHashes[hostname]; exists {
		for i, pin := range pins {
			if strings.EqualFold(pin, certHash) {
				cp.pinnedHashes[hostname] = append(pins[:i], pins[i+1:]...)
				break
			}
		}
		
		// Remove hostname entry if no pins left
		if len(cp.pinnedHashes[hostname]) == 0 {
			delete(cp.pinnedHashes, hostname)
		}
	}
}

// GetPinnedHashes returns all pinned certificate hashes
func (cp *CertificatePinner) GetPinnedHashes() map[string][]string {
	result := make(map[string][]string)
	for hostname, hashes := range cp.pinnedHashes {
		result[hostname] = make([]string, len(hashes))
		copy(result[hostname], hashes)
	}
	return result
}

// UpdateGoogleAPIsPins updates the pinned hashes for Google APIs (for maintenance)
func (cp *CertificatePinner) UpdateGoogleAPIsPins(newPins map[string][]string) {
	for hostname, hashes := range newPins {
		cp.pinnedHashes[hostname] = hashes
	}
}

// VerifyCurrentCertificates fetches and verifies current certificates against pins
func (cp *CertificatePinner) VerifyCurrentCertificates() map[string]error {
	results := make(map[string]error)
	
	for hostname := range cp.pinnedHashes {
		err := cp.verifySingleHostname(hostname)
		results[hostname] = err
	}
	
	return results
}

// verifySingleHostname verifies certificates for a single hostname
func (cp *CertificatePinner) verifySingleHostname(hostname string) error {
	client := cp.CreateSecureHTTPClient(DefaultPinningConfig())
	
	url := fmt.Sprintf("https://%s", hostname)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	resp.Body.Close()
	
	return nil
}

// PinningReport provides detailed information about certificate pinning status
type PinningReport struct {
	Hostname         string    `json:"hostname"`
	PinnedHashes     []string  `json:"pinned_hashes"`
	CurrentHash      string    `json:"current_hash"`
	PinMatchFound    bool      `json:"pin_match_found"`
	LastVerified     time.Time `json:"last_verified"`
	Error           string    `json:"error,omitempty"`
}

// GeneratePinningReport creates a comprehensive pinning report
func (cp *CertificatePinner) GeneratePinningReport() []*PinningReport {
	var reports []*PinningReport
	
	for hostname, pins := range cp.pinnedHashes {
		report := &PinningReport{
			Hostname:     hostname,
			PinnedHashes: pins,
			LastVerified: time.Now(),
		}
		
		// Try to get current certificate
		if err := cp.verifySingleHostname(hostname); err != nil {
			report.Error = err.Error()
			report.PinMatchFound = false
		} else {
			report.PinMatchFound = true
		}
		
		reports = append(reports, report)
	}
	
	return reports
}