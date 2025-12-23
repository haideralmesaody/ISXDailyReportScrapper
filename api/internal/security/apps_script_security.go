package security

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	
	"github.com/isxcli/isxcli/internal/config"
)

// AppsScriptSecurityConfig holds configuration for secure Apps Script communication
type AppsScriptSecurityConfig struct {
	SharedSecret         string        `json:"shared_secret"`          // HMAC signing key
	RequestTimeout       time.Duration `json:"request_timeout"`        // Maximum request timeout
	MaxRetries          int           `json:"max_retries"`            // Maximum retry attempts
	RetryBaseDelay      time.Duration `json:"retry_base_delay"`       // Base delay for exponential backoff
	TimestampWindow     time.Duration `json:"timestamp_window"`       // Maximum age of request timestamp
	EnableEncryption    bool          `json:"enable_encryption"`      // Enable request/response encryption
	RequireSignature    bool          `json:"require_signature"`      // Require HMAC signatures
	UserAgent          string        `json:"user_agent"`             // Custom user agent string
	MaxRequestSize      int64         `json:"max_request_size"`       // Maximum request body size
	AllowInsecure       bool          `json:"allow_insecure"`         // Allow HTTP (for testing only)
}

// SecureAppsScriptClient provides secure communication with Google Apps Script
type SecureAppsScriptClient struct {
	config      *AppsScriptSecurityConfig
	httpClient  *http.Client
	certPinner  *CertificatePinner
	logger      *slog.Logger
	lastRequest time.Time
	requestID   uint64
}

// SignedRequest represents a request with HMAC signature
type SignedRequest struct {
	Timestamp   int64                  `json:"timestamp"`    // Unix timestamp
	Nonce       string                 `json:"nonce"`        // Random nonce for replay protection
	RequestID   string                 `json:"request_id"`   // Unique request identifier
	Payload     map[string]interface{} `json:"payload"`      // Original request data
	Signature   string                 `json:"signature"`    // HMAC-SHA256 signature
	Fingerprint string                 `json:"fingerprint"`  // Client fingerprint
}

// SignedResponse represents a response with signature verification
type SignedResponse struct {
	Timestamp int64                  `json:"timestamp"`  // Server timestamp
	RequestID string                 `json:"request_id"` // Matching request ID
	Success   bool                   `json:"success"`    // Operation success status
	Data      map[string]interface{} `json:"data"`       // Response data
	Error     string                 `json:"error"`      // Error message if any
	Signature string                 `json:"signature"`  // Server HMAC signature
}

// SecurityEvent represents a security-related event for audit logging
type SecurityEvent struct {
	Timestamp   time.Time   `json:"timestamp"`
	EventType   string      `json:"event_type"`
	RequestID   string      `json:"request_id"`
	Success     bool        `json:"success"`
	Error       string      `json:"error,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	ClientIP    string      `json:"client_ip,omitempty"`
	UserAgent   string      `json:"user_agent,omitempty"`
	Fingerprint string      `json:"fingerprint,omitempty"`
	Details     interface{} `json:"details,omitempty"`
}

// DefaultAppsScriptSecurityConfig returns secure default configuration
func DefaultAppsScriptSecurityConfig() *AppsScriptSecurityConfig {
	// Use embedded credentials from config package
	creds := config.GetCredentials()
	
	return &AppsScriptSecurityConfig{
		SharedSecret:        creds.AppsScriptSecret,  // Use embedded shared secret
		RequestTimeout:      30 * time.Second,
		MaxRetries:         3,
		RetryBaseDelay:     1 * time.Second,
		TimestampWindow:    5 * time.Minute,
		EnableEncryption:   creds.EnableEncryption,
		RequireSignature:   creds.RequireSignature,
		UserAgent:         "ISX-Pulse-SecureClient/1.0",
		MaxRequestSize:    1024 * 1024, // 1MB limit
		AllowInsecure:     creds.AllowInsecureForTest,
	}
}

// NewSecureAppsScriptClient creates a new secure Apps Script client
func NewSecureAppsScriptClient(config *AppsScriptSecurityConfig, certPinner *CertificatePinner) *SecureAppsScriptClient {
	if config == nil {
		config = DefaultAppsScriptSecurityConfig()
	}

	// Create secure HTTP client with certificate pinning
	var httpClient *http.Client
	if certPinner != nil {
		httpClient = certPinner.CreateSecureHTTPClient(DefaultPinningConfig())
	} else {
		httpClient = &http.Client{
			Timeout: config.RequestTimeout,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
			},
		}
	}

	// Override timeout from config
	httpClient.Timeout = config.RequestTimeout

	return &SecureAppsScriptClient{
		config:     config,
		httpClient: httpClient,
		certPinner: certPinner,
		logger:     slog.Default(),
	}
}

// SetLogger sets a custom logger for the client
func (c *SecureAppsScriptClient) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

// SecureRequest makes a secure request to Apps Script with HMAC signature and optional encryption
func (c *SecureAppsScriptClient) SecureRequest(ctx context.Context, endpoint string, payload map[string]interface{}, fingerprint string) (*SignedResponse, error) {
	start := time.Now()
	requestID := c.generateRequestID()

	// Log request start with full details
	c.logger.Info("Apps Script request starting",
		slog.String("endpoint", endpoint),
		slog.String("shared_secret_length", fmt.Sprintf("%d chars", len(c.config.SharedSecret))),
		slog.Bool("signature_required", c.config.RequireSignature),
		slog.String("fingerprint", fingerprint),
		slog.Any("payload", payload),
	)
	
	c.logSecurityEvent("secure_request_start", requestID, true, "", start, "", "", fingerprint, map[string]interface{}{
		"endpoint": endpoint,
		"action":   payload["action"],
	})

	// Validate endpoint
	if err := c.validateEndpoint(endpoint); err != nil {
		c.logSecurityEvent("endpoint_validation_failed", requestID, false, err.Error(), start, "", "", fingerprint, nil)
		return nil, fmt.Errorf("endpoint validation failed: %w", err)
	}

	// Create signed request
	signedReq, err := c.createSignedRequest(payload, fingerprint, requestID)
	if err != nil {
		c.logSecurityEvent("request_signing_failed", requestID, false, err.Error(), start, "", "", fingerprint, nil)
		return nil, fmt.Errorf("failed to create signed request: %w", err)
	}

	// Send request with retry logic
	response, err := c.sendRequestWithRetry(ctx, endpoint, signedReq, requestID, fingerprint)
	if err != nil {
		c.logSecurityEvent("request_failed", requestID, false, err.Error(), start, "", "", fingerprint, nil)
		return nil, err
	}

	// Verify response signature
	if c.config.RequireSignature {
		if err := c.verifyResponseSignature(response, requestID); err != nil {
			c.logSecurityEvent("response_verification_failed", requestID, false, err.Error(), start, "", "", fingerprint, nil)
			return nil, fmt.Errorf("response signature verification failed: %w", err)
		}
	}

	// Log successful request
	c.logSecurityEvent("secure_request_success", requestID, true, "", start, "", "", fingerprint, map[string]interface{}{
		"response_size": len(fmt.Sprintf("%v", response.Data)),
		"server_time":   time.Unix(response.Timestamp, 0),
	})

	return response, nil
}

// createSignedRequest creates a request with HMAC signature
func (c *SecureAppsScriptClient) createSignedRequest(payload map[string]interface{}, fingerprint, requestID string) (*SignedRequest, error) {
	now := time.Now()
	nonce, err := c.generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	signedReq := &SignedRequest{
		Timestamp:   now.Unix(),
		Nonce:       nonce,
		RequestID:   requestID,
		Payload:     payload,
		Fingerprint: fingerprint,
	}

	// Create signature if required
	if c.config.RequireSignature {
		signature, err := c.signRequest(signedReq)
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}
		signedReq.Signature = signature
	}

	return signedReq, nil
}

// signRequest creates HMAC-SHA256 signature for the request
func (c *SecureAppsScriptClient) signRequest(req *SignedRequest) (string, error) {
	// Create canonical string to sign
	canonical := fmt.Sprintf("%d|%s|%s|%s", req.Timestamp, req.Nonce, req.RequestID, req.Fingerprint)
	
	// Add payload data in sorted order for consistency
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	canonical += "|" + string(payloadJSON)

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(c.config.SharedSecret))
	h.Write([]byte(canonical))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}

// verifyResponseSignature verifies the HMAC signature of the response
func (c *SecureAppsScriptClient) verifyResponseSignature(resp *SignedResponse, expectedRequestID string) error {
	if resp.Signature == "" {
		return fmt.Errorf("response signature is missing")
	}

	if resp.RequestID != expectedRequestID {
		return fmt.Errorf("request ID mismatch: expected %s, got %s", expectedRequestID, resp.RequestID)
	}

	// Check timestamp freshness (allow up to 10 minutes for server clock skew)
	now := time.Now()
	serverTime := time.Unix(resp.Timestamp, 0)
	if now.Sub(serverTime) > 10*time.Minute || serverTime.Sub(now) > 10*time.Minute {
		return fmt.Errorf("response timestamp is too old or too far in the future")
	}

	// Create canonical string to verify
	canonical := fmt.Sprintf("%d|%s|%t", resp.Timestamp, resp.RequestID, resp.Success)
	
	// Add response data
	if resp.Data != nil {
		dataJSON, err := json.Marshal(resp.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal response data: %w", err)
		}
		canonical += "|" + string(dataJSON)
	}
	
	if resp.Error != "" {
		canonical += "|" + resp.Error
	}

	// Verify HMAC signature
	h := hmac.New(sha256.New, []byte(c.config.SharedSecret))
	h.Write([]byte(canonical))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Debug logging for HMAC verification
	c.logger.Info("HMAC Verification Debug",
		slog.String("canonical_string", canonical),
		slog.Int64("timestamp", resp.Timestamp),
		slog.String("request_id", resp.RequestID),
		slog.Bool("success", resp.Success),
		slog.Any("data", resp.Data),
		slog.String("error", resp.Error),
		slog.String("received_signature", resp.Signature),
		slog.String("expected_signature", expectedSignature),
		slog.String("shared_secret_length", fmt.Sprintf("%d chars", len(c.config.SharedSecret))),
		slog.Bool("signatures_match", hmac.Equal([]byte(resp.Signature), []byte(expectedSignature))),
	)

	if !hmac.Equal([]byte(resp.Signature), []byte(expectedSignature)) {
		return fmt.Errorf("HMAC signature verification failed")
	}

	return nil
}

// sendRequestWithRetry sends request with exponential backoff retry logic
func (c *SecureAppsScriptClient) sendRequestWithRetry(ctx context.Context, endpoint string, signedReq *SignedRequest, requestID, fingerprint string) (*SignedResponse, error) {
	var lastErr error
	backoff := c.config.RetryBaseDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
				if backoff > 30*time.Second {
					backoff = 30 * time.Second // Cap at 30 seconds
				}
			}

			c.logSecurityEvent("request_retry", requestID, false, fmt.Sprintf("attempt %d", attempt), time.Now(), "", "", fingerprint, map[string]interface{}{
				"backoff_delay": backoff.String(),
			})
		}

		response, err := c.sendSingleRequest(ctx, endpoint, signedReq)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry certain errors
		if isNonRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// sendSingleRequest sends a single HTTP request
func (c *SecureAppsScriptClient) sendSingleRequest(ctx context.Context, endpoint string, signedReq *SignedRequest) (*SignedResponse, error) {
	// Marshal request
	requestData, err := json.Marshal(signedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Check request size limit
	if int64(len(requestData)) > c.config.MaxRequestSize {
		return nil, fmt.Errorf("request size %d exceeds limit %d", len(requestData), c.config.MaxRequestSize)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set security headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("X-Request-ID", signedReq.RequestID)
	req.Header.Set("X-Timestamp", strconv.FormatInt(signedReq.Timestamp, 10))
	req.Header.Set("X-Security-Level", "enhanced")
	
	// Anti-DDoS headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("DNT", "1") // Do Not Track

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is gzipped
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Verify response status
	if resp.StatusCode != http.StatusOK {
		// Try to parse as JSON error response first
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			// If it's valid JSON, format it nicely
			if msg, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
			}
			// Otherwise show the whole JSON
			return nil, fmt.Errorf("HTTP %d: %v", resp.StatusCode, errorResp)
		}
		// If not JSON, show as string (truncate if too long for HTML pages)
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "... (truncated)"
		}
		c.logger.Error("Apps Script request failed",
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", bodyStr),
			slog.String("endpoint", endpoint),
		)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)
	}

	// Parse response
	var signedResp SignedResponse
	if err := json.Unmarshal(body, &signedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Debug log the raw response
	c.logger.Info("Apps Script Raw Response Debug",
		slog.String("raw_response", string(body)),
		slog.Int64("parsed_timestamp", signedResp.Timestamp),
		slog.String("parsed_request_id", signedResp.RequestID),
		slog.Bool("parsed_success", signedResp.Success),
		slog.String("parsed_error", signedResp.Error),
		slog.String("parsed_signature", signedResp.Signature),
	)

	return &signedResp, nil
}

// validateEndpoint validates that the endpoint is secure and allowed
func (c *SecureAppsScriptClient) validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	// Must be HTTPS unless explicitly allowing insecure (testing only)
	if !c.config.AllowInsecure && !strings.HasPrefix(endpoint, "https://") {
		return fmt.Errorf("endpoint must use HTTPS")
	}

	// Must be Google Apps Script domain
	if !strings.Contains(endpoint, "script.google.com") {
		return fmt.Errorf("endpoint must be on script.google.com domain")
	}

	// Parse URL to validate path portion only
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check for path traversal attempts
	if strings.Contains(parsedURL.Path, "..") {
		return fmt.Errorf("endpoint path contains invalid characters (path traversal)")
	}

	// Check for double slashes in path (not in scheme)
	// Skip the leading slash if present
	pathToCheck := parsedURL.Path
	if len(pathToCheck) > 1 {
		pathToCheck = pathToCheck[1:] // Remove leading slash
	}
	if strings.Contains(pathToCheck, "//") {
		return fmt.Errorf("endpoint path contains double slashes")
	}

	return nil
}

// generateNonce generates a cryptographically secure random nonce
func (c *SecureAppsScriptClient) generateNonce() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce), nil
}

// generateRequestID generates a unique request identifier
func (c *SecureAppsScriptClient) generateRequestID() string {
	c.requestID++
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req_%d_%d", timestamp, c.requestID)
}

// generateSecureSecret generates a cryptographically secure shared secret
func generateSecureSecret() string {
	secret := make([]byte, 32)
	rand.Read(secret)
	return base64.StdEncoding.EncodeToString(secret)
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	errStr := err.Error()
	
	// Don't retry authentication/authorization errors
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") {
		return true
	}
	
	// Don't retry bad requests
	if strings.Contains(errStr, "400") {
		return true
	}
	
	// Don't retry signature verification failures
	if strings.Contains(errStr, "signature") || strings.Contains(errStr, "HMAC") {
		return true
	}
	
	// Don't retry endpoint validation failures
	if strings.Contains(errStr, "endpoint validation") {
		return true
	}
	
	return false
}

// logSecurityEvent logs security events for audit trail
func (c *SecureAppsScriptClient) logSecurityEvent(eventType, requestID string, success bool, errorMsg string, startTime time.Time, clientIP, userAgent, fingerprint string, details interface{}) {
	duration := time.Since(startTime)
	
	event := SecurityEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		RequestID:   requestID,
		Success:     success,
		Error:       errorMsg,
		Duration:    duration.String(),
		ClientIP:    clientIP,
		UserAgent:   userAgent,
		Fingerprint: fingerprint,
		Details:     details,
	}

	// Determine log level
	level := slog.LevelInfo
	if !success {
		level = slog.LevelError
	}

	// Log with structured fields
	c.logger.Log(context.Background(), level, "Apps Script security event",
		slog.String("event_type", event.EventType),
		slog.String("request_id", event.RequestID),
		slog.Bool("success", event.Success),
		slog.String("error", event.Error),
		slog.String("duration", event.Duration),
		slog.String("client_ip", event.ClientIP),
		slog.String("user_agent", event.UserAgent),
		slog.String("fingerprint", event.Fingerprint),
		slog.Any("details", event.Details),
		slog.Time("timestamp", event.Timestamp),
	)
}

// GetSecurityMetrics returns security metrics for monitoring
func (c *SecureAppsScriptClient) GetSecurityMetrics() map[string]interface{} {
	return map[string]interface{}{
		"shared_secret_length":   len(c.config.SharedSecret),
		"request_timeout":        c.config.RequestTimeout.String(),
		"max_retries":           c.config.MaxRetries,
		"retry_base_delay":      c.config.RetryBaseDelay.String(),
		"timestamp_window":      c.config.TimestampWindow.String(),
		"encryption_enabled":    c.config.EnableEncryption,
		"signature_required":    c.config.RequireSignature,
		"max_request_size":      c.config.MaxRequestSize,
		"certificate_pinning":   c.certPinner != nil,
		"last_request":          c.lastRequest,
		"total_requests":        c.requestID,
	}
}

// ValidateConfiguration validates the security configuration
func (c *SecureAppsScriptClient) ValidateConfiguration() error {
	config := c.config

	if len(config.SharedSecret) < 32 {
		return fmt.Errorf("shared secret must be at least 32 characters long")
	}

	if config.RequestTimeout < 5*time.Second {
		return fmt.Errorf("request timeout must be at least 5 seconds")
	}

	if config.RequestTimeout > 120*time.Second {
		return fmt.Errorf("request timeout must not exceed 120 seconds")
	}

	if config.MaxRetries < 0 || config.MaxRetries > 10 {
		return fmt.Errorf("max retries must be between 0 and 10")
	}

	if config.TimestampWindow < 1*time.Minute {
		return fmt.Errorf("timestamp window must be at least 1 minute")
	}

	if config.TimestampWindow > 30*time.Minute {
		return fmt.Errorf("timestamp window must not exceed 30 minutes")
	}

	if config.MaxRequestSize < 1024 {
		return fmt.Errorf("max request size must be at least 1024 bytes")
	}

	if config.MaxRequestSize > 10*1024*1024 {
		return fmt.Errorf("max request size must not exceed 10MB")
	}

	return nil
}

// Close performs cleanup and logs final security event
func (c *SecureAppsScriptClient) Close() {
	c.logSecurityEvent("client_shutdown", "system", true, "", time.Now(), "", "", "", map[string]interface{}{
		"total_requests": c.requestID,
		"uptime":        time.Since(c.lastRequest).String(),
	})
}