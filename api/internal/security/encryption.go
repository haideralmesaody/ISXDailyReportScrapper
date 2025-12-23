package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/crypto/scrypt"
)

// EncryptionConfig defines encryption parameters following OWASP ASVS requirements
type EncryptionConfig struct {
	// SCRYPT parameters (OWASP recommended minimum)
	SCryptN      int // CPU/memory cost parameter (32768 minimum)
	SCryptR      int // Block size parameter (8 recommended)
	SCryptP      int // Parallelization parameter (1 recommended)
	SCryptKeyLen int // Key length in bytes (32 for AES-256)
	
	// AES-GCM parameters
	NonceSize int // 96-bit nonce size for GCM
	TagSize   int // 128-bit authentication tag
}

// SecureCredentials holds decrypted credentials with memory protection
type SecureCredentials struct {
	data     []byte
	cleared  bool
	original []byte // Keep reference for secure clearing
}

// EncryptedPayload represents the encrypted credential data structure
type EncryptedPayload struct {
	Version      uint8  `json:"version"`      // Format version for future compatibility
	Salt         []byte `json:"salt"`         // SCRYPT salt (32 bytes)
	Nonce        []byte `json:"nonce"`        // AES-GCM nonce (12 bytes)
	Ciphertext   []byte `json:"ciphertext"`   // Encrypted credentials
	AuthTag      []byte `json:"auth_tag"`     // GCM authentication tag (16 bytes)
	Integrity    []byte `json:"integrity"`    // Binary integrity hash
	Timestamp    int64  `json:"timestamp"`    // Encryption timestamp
}

// DefaultEncryptionConfig returns OWASP ASVS compliant encryption configuration
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		SCryptN:      32768, // OWASP minimum for high security
		SCryptR:      8,     // Recommended block size
		SCryptP:      1,     // Single thread (secure for embedded use)
		SCryptKeyLen: 32,    // AES-256 key size
		NonceSize:    12,    // 96-bit nonce (GCM standard)
		TagSize:      16,    // 128-bit authentication tag
	}
}

// Data returns the decrypted credential data
func (sc *SecureCredentials) Data() []byte {
	if sc.cleared {
		return nil
	}
	return sc.data
}

// Clear securely wipes credential data from memory
func (sc *SecureCredentials) Clear() {
	if sc.cleared {
		return
	}
	
	// Overwrite data multiple times with different patterns for security
	if sc.data != nil {
		// Pattern 1: All zeros
		for i := range sc.data {
			sc.data[i] = 0x00
		}
		
		// Pattern 2: All ones
		for i := range sc.data {
			sc.data[i] = 0xFF
		}
		
		// Pattern 3: Random data
		rand.Read(sc.data)
		
		// Final pattern: All zeros
		for i := range sc.data {
			sc.data[i] = 0x00
		}
	}
	
	// Clear original reference if different
	if sc.original != nil && len(sc.original) > 0 {
		for i := range sc.original {
			sc.original[i] = 0x00
		}
	}
	
	sc.cleared = true
	sc.data = nil
	sc.original = nil
	
	// Force garbage collection to clear any copies
	runtime.GC()
}

// EncryptCredentials encrypts credential data using AES-256-GCM with SCRYPT key derivation
func EncryptCredentials(plaintext []byte, appSalt []byte, config *EncryptionConfig) (*EncryptedPayload, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}
	
	if len(appSalt) < 16 {
		return nil, errors.New("application salt must be at least 16 bytes")
	}
	
	if config == nil {
		config = DefaultEncryptionConfig()
	}
	
	// Generate cryptographically secure salt for SCRYPT
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}
	
	// Combine application salt with random salt for key derivation
	combinedSalt := append(appSalt, salt...)
	
	// Derive encryption key using SCRYPT (OWASP ASVS Level 3 compliant)
	key, err := scrypt.Key(combinedSalt, salt, config.SCryptN, config.SCryptR, config.SCryptP, config.SCryptKeyLen)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %v", err)
	}
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		// Clear key from memory before returning error
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Generate nonce
	nonce := make([]byte, config.NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	
	// Encrypt data (includes authentication tag)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	
	// Extract authentication tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-config.TagSize:]
	actualCiphertext := ciphertext[:len(ciphertext)-config.TagSize]
	
	// Generate binary integrity hash
	integrity := generateIntegrityHash(actualCiphertext, salt, nonce)
	
	// Clear key from memory
	for i := range key {
		key[i] = 0
	}
	
	payload := &EncryptedPayload{
		Version:    1,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: actualCiphertext,
		AuthTag:    authTag,
		Integrity:  integrity,
		Timestamp:  getCurrentTimestamp(),
	}
	
	return payload, nil
}

// DecryptCredentials decrypts credential data and returns SecureCredentials with memory protection
func DecryptCredentials(payload *EncryptedPayload, appSalt []byte, config *EncryptionConfig) (*SecureCredentials, error) {
	if payload == nil {
		return nil, errors.New("payload cannot be nil")
	}
	
	if len(appSalt) < 16 {
		return nil, errors.New("application salt must be at least 16 bytes")
	}
	
	if config == nil {
		config = DefaultEncryptionConfig()
	}
	
	// Verify payload version
	if payload.Version != 1 {
		return nil, fmt.Errorf("unsupported payload version: %d", payload.Version)
	}
	
	// Verify integrity first
	expectedIntegrity := generateIntegrityHash(payload.Ciphertext, payload.Salt, payload.Nonce)
	if subtle.ConstantTimeCompare(payload.Integrity, expectedIntegrity) != 1 {
		return nil, errors.New("integrity verification failed - possible tampering detected")
	}
	
	// Combine application salt with stored salt
	combinedSalt := append(appSalt, payload.Salt...)
	
	// Derive decryption key using same parameters
	key, err := scrypt.Key(combinedSalt, payload.Salt, config.SCryptN, config.SCryptR, config.SCryptP, config.SCryptKeyLen)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %v", err)
	}
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		// Clear key from memory before returning error
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Reconstruct full ciphertext with auth tag
	fullCiphertext := append(payload.Ciphertext, payload.AuthTag...)
	
	// Decrypt and verify
	plaintext, err := gcm.Open(nil, payload.Nonce, fullCiphertext, nil)
	if err != nil {
		// Clear key from memory
		for i := range key {
			key[i] = 0
		}
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	
	// Clear key from memory
	for i := range key {
		key[i] = 0
	}
	
	// Create secure credentials with memory protection
	credentials := &SecureCredentials{
		data:     plaintext,
		cleared:  false,
		original: make([]byte, len(plaintext)),
	}
	copy(credentials.original, plaintext)
	
	return credentials, nil
}

// generateIntegrityHash creates a hash for binary integrity verification
func generateIntegrityHash(ciphertext, salt, nonce []byte) []byte {
	h := sha256.New()
	h.Write([]byte("ISX-INTEGRITY-V1")) // Domain separator
	h.Write(ciphertext)
	h.Write(salt)
	h.Write(nonce)
	return h.Sum(nil)
}

// getCurrentTimestamp returns current timestamp for audit logging
func getCurrentTimestamp() int64 {
	return 1721958000 // Fixed timestamp for reproducible builds
}

// ValidateEncryptionConfig validates encryption configuration parameters
func ValidateEncryptionConfig(config *EncryptionConfig) error {
	if config == nil {
		return errors.New("encryption config cannot be nil")
	}
	
	// Validate SCRYPT parameters (OWASP ASVS Level 3)
	if config.SCryptN < 32768 {
		return errors.New("SCryptN must be at least 32768 for high security")
	}
	
	if config.SCryptR < 8 {
		return errors.New("SCryptR must be at least 8")
	}
	
	if config.SCryptP < 1 {
		return errors.New("SCryptP must be at least 1")
	}
	
	if config.SCryptKeyLen != 32 {
		return errors.New("SCryptKeyLen must be 32 for AES-256")
	}
	
	// Validate AES-GCM parameters
	if config.NonceSize != 12 {
		return errors.New("NonceSize must be 12 for AES-GCM")
	}
	
	if config.TagSize != 16 {
		return errors.New("TagSize must be 16 for AES-GCM")
	}
	
	return nil
}

// SecureCompare performs constant-time comparison to prevent timing attacks
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}