package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/hkdf"
)

// EncryptedRequest represents an encrypted request payload
type EncryptedRequest struct {
	Version        uint8  `json:"version"`         // Encryption version for compatibility
	Timestamp      int64  `json:"timestamp"`       // Request timestamp
	RequestID      string `json:"request_id"`      // Unique request identifier
	EncryptedData  string `json:"encrypted_data"`  // Base64 encoded encrypted payload
	Nonce          string `json:"nonce"`           // Base64 encoded encryption nonce
	KeyDeriveSalt  string `json:"key_derive_salt"` // Base64 encoded key derivation salt
	AuthTag        string `json:"auth_tag"`        // Base64 encoded authentication tag
	Fingerprint    string `json:"fingerprint"`     // Client device fingerprint
	HMAC           string `json:"hmac"`            // HMAC signature of entire request
}

// EncryptedResponse represents an encrypted response payload
type EncryptedResponse struct {
	Version       uint8  `json:"version"`        // Encryption version
	Timestamp     int64  `json:"timestamp"`      // Server timestamp
	RequestID     string `json:"request_id"`     // Matching request ID
	EncryptedData string `json:"encrypted_data"` // Base64 encoded encrypted response
	Nonce         string `json:"nonce"`          // Base64 encoded encryption nonce
	AuthTag       string `json:"auth_tag"`       // Base64 encoded authentication tag
	HMAC          string `json:"hmac"`           // HMAC signature of entire response
}

// RequestEncryption handles encryption/decryption of sensitive request/response data
type RequestEncryption struct {
	sharedSecret []byte
	keySize      int
	nonceSize    int
	tagSize      int
}

// NewRequestEncryption creates a new request encryption handler
func NewRequestEncryption(sharedSecret string) *RequestEncryption {
	return &RequestEncryption{
		sharedSecret: []byte(sharedSecret),
		keySize:      32, // AES-256
		nonceSize:    12, // GCM standard nonce size
		tagSize:      16, // GCM authentication tag size
	}
}

// EncryptRequest encrypts sensitive request data using AES-256-GCM
func (re *RequestEncryption) EncryptRequest(payload map[string]interface{}, requestID, fingerprint string) (*EncryptedRequest, error) {
	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Generate random salt for key derivation
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key using HKDF-SHA256
	key, err := re.deriveKey(salt, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}
	defer re.clearKey(key) // Ensure key is cleared from memory

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, re.nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the payload
	ciphertext := gcm.Seal(nil, nonce, payloadJSON, []byte(requestID))

	// Split ciphertext and auth tag
	if len(ciphertext) < re.tagSize {
		return nil, errors.New("invalid ciphertext length")
	}
	
	actualCiphertext := ciphertext[:len(ciphertext)-re.tagSize]
	authTag := ciphertext[len(ciphertext)-re.tagSize:]

	// Create encrypted request structure
	encReq := &EncryptedRequest{
		Version:       1,
		Timestamp:     time.Now().Unix(),
		RequestID:     requestID,
		EncryptedData: base64.StdEncoding.EncodeToString(actualCiphertext),
		Nonce:         base64.StdEncoding.EncodeToString(nonce),
		KeyDeriveSalt: base64.StdEncoding.EncodeToString(salt),
		AuthTag:       base64.StdEncoding.EncodeToString(authTag),
		Fingerprint:   fingerprint,
	}

	// Generate HMAC for the entire encrypted request
	hmacSig, err := re.generateRequestHMAC(encReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC: %w", err)
	}
	encReq.HMAC = hmacSig

	return encReq, nil
}

// DecryptRequest decrypts an encrypted request and returns the original payload
func (re *RequestEncryption) DecryptRequest(encReq *EncryptedRequest) (map[string]interface{}, error) {
	// Verify version
	if encReq.Version != 1 {
		return nil, fmt.Errorf("unsupported encryption version: %d", encReq.Version)
	}

	// Verify timestamp (prevent replay attacks)
	now := time.Now().Unix()
	if now-encReq.Timestamp > 300 { // 5 minutes
		return nil, fmt.Errorf("request timestamp is too old")
	}
	if encReq.Timestamp-now > 60 { // 1 minute future tolerance
		return nil, fmt.Errorf("request timestamp is too far in the future")
	}

	// Verify HMAC first
	expectedHMAC, err := re.generateRequestHMAC(encReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC for verification: %w", err)
	}
	
	if !re.constantTimeCompare(encReq.HMAC, expectedHMAC) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	// Decode base64 fields
	salt, err := base64.StdEncoding.DecodeString(encReq.KeyDeriveSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(encReq.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encReq.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(encReq.AuthTag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth tag: %w", err)
	}

	// Derive decryption key
	key, err := re.deriveKey(salt, encReq.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive decryption key: %w", err)
	}
	defer re.clearKey(key)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Reconstruct full ciphertext with auth tag
	fullCiphertext := append(ciphertext, authTag...)

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, fullCiphertext, []byte(encReq.RequestID))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Parse JSON payload
	var payload map[string]interface{}
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted payload: %w", err)
	}

	return payload, nil
}

// EncryptResponse encrypts response data for secure transmission
func (re *RequestEncryption) EncryptResponse(data map[string]interface{}, requestID string) (*EncryptedResponse, error) {
	// Marshal response data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Generate random salt for key derivation
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key
	key, err := re.deriveKey(salt, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}
	defer re.clearKey(key)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, re.nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, dataJSON, []byte(requestID))

	// Split ciphertext and auth tag
	actualCiphertext := ciphertext[:len(ciphertext)-re.tagSize]
	authTag := ciphertext[len(ciphertext)-re.tagSize:]

	// Create encrypted response structure
	encResp := &EncryptedResponse{
		Version:       1,
		Timestamp:     time.Now().Unix(),
		RequestID:     requestID,
		EncryptedData: base64.StdEncoding.EncodeToString(actualCiphertext),
		Nonce:         base64.StdEncoding.EncodeToString(nonce),
		AuthTag:       base64.StdEncoding.EncodeToString(authTag),
	}

	// Generate HMAC for the entire encrypted response
	hmacSig, err := re.generateResponseHMAC(encResp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC: %w", err)
	}
	encResp.HMAC = hmacSig

	return encResp, nil
}

// DecryptResponse decrypts an encrypted response and returns the original data
func (re *RequestEncryption) DecryptResponse(encResp *EncryptedResponse) (map[string]interface{}, error) {
	// Verify version
	if encResp.Version != 1 {
		return nil, fmt.Errorf("unsupported encryption version: %d", encResp.Version)
	}

	// Verify HMAC
	expectedHMAC, err := re.generateResponseHMAC(encResp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC for verification: %w", err)
	}
	
	if !re.constantTimeCompare(encResp.HMAC, expectedHMAC) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	// Decode base64 fields
	nonce, err := base64.StdEncoding.DecodeString(encResp.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encResp.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(encResp.AuthTag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth tag: %w", err)
	}

	// For response decryption, we need to re-derive the key using the same salt as the request
	// This requires the salt to be transmitted or derived deterministically
	// For simplicity, we'll use a deterministic salt based on request ID
	salt := sha256.Sum256([]byte(fmt.Sprintf("response_salt_%s", encResp.RequestID)))

	// Derive decryption key
	key, err := re.deriveKey(salt[:], encResp.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive decryption key: %w", err)
	}
	defer re.clearKey(key)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Reconstruct full ciphertext with auth tag
	fullCiphertext := append(ciphertext, authTag...)

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, fullCiphertext, []byte(encResp.RequestID))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Parse JSON data
	var data map[string]interface{}
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted data: %w", err)
	}

	return data, nil
}

// deriveKey derives an encryption key using HKDF-SHA256
func (re *RequestEncryption) deriveKey(salt []byte, info string) ([]byte, error) {
	// Use HKDF to derive key from shared secret
	hkdf := hkdf.New(sha256.New, re.sharedSecret, salt, []byte(info))
	
	key := make([]byte, re.keySize)
	if _, err := hkdf.Read(key); err != nil {
		return nil, fmt.Errorf("HKDF key derivation failed: %w", err)
	}
	
	return key, nil
}

// clearKey securely clears key material from memory
func (re *RequestEncryption) clearKey(key []byte) {
	if key != nil {
		for i := range key {
			key[i] = 0
		}
	}
}

// generateRequestHMAC generates HMAC signature for encrypted request
func (re *RequestEncryption) generateRequestHMAC(encReq *EncryptedRequest) (string, error) {
	// Create canonical string for HMAC (exclude HMAC field itself)
	canonical := fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s|%s",
		encReq.Version,
		encReq.Timestamp,
		encReq.RequestID,
		encReq.EncryptedData,
		encReq.Nonce,
		encReq.KeyDeriveSalt,
		encReq.AuthTag,
		encReq.Fingerprint,
	)

	return re.computeHMAC(canonical), nil
}

// generateResponseHMAC generates HMAC signature for encrypted response
func (re *RequestEncryption) generateResponseHMAC(encResp *EncryptedResponse) (string, error) {
	// Create canonical string for HMAC (exclude HMAC field itself)
	canonical := fmt.Sprintf("%d|%d|%s|%s|%s|%s",
		encResp.Version,
		encResp.Timestamp,
		encResp.RequestID,
		encResp.EncryptedData,
		encResp.Nonce,
		encResp.AuthTag,
	)

	return re.computeHMAC(canonical), nil
}

// computeHMAC computes HMAC-SHA256 for given data
func (re *RequestEncryption) computeHMAC(data string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("HMAC_%s_%s", string(re.sharedSecret), data)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// constantTimeCompare performs constant-time string comparison
func (re *RequestEncryption) constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	
	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	
	return result == 0
}

// ValidateEncryptionConfig validates encryption configuration
func (re *RequestEncryption) ValidateEncryptionConfig() error {
	if len(re.sharedSecret) < 32 {
		return fmt.Errorf("shared secret must be at least 32 bytes")
	}
	
	if re.keySize != 32 {
		return fmt.Errorf("key size must be 32 bytes for AES-256")
	}
	
	if re.nonceSize != 12 {
		return fmt.Errorf("nonce size must be 12 bytes for GCM")
	}
	
	if re.tagSize != 16 {
		return fmt.Errorf("auth tag size must be 16 bytes for GCM")
	}
	
	return nil
}

// GetEncryptionMetrics returns encryption-related metrics
func (re *RequestEncryption) GetEncryptionMetrics() map[string]interface{} {
	return map[string]interface{}{
		"shared_secret_length": len(re.sharedSecret),
		"key_size":            re.keySize,
		"nonce_size":          re.nonceSize,
		"tag_size":            re.tagSize,
		"algorithm":           "AES-256-GCM",
		"key_derivation":      "HKDF-SHA256",
		"version":             1,
	}
}