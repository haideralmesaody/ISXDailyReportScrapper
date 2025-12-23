package security

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestEncryptionDecryption(t *testing.T) {
	tests := []struct {
		name        string
		plaintext   []byte
		appSalt     []byte
		shouldError bool
	}{
		{
			name:        "Valid encryption/decryption",
			plaintext:   []byte(`{"type": "service_account", "project_id": "test"}`),
			appSalt:     []byte("test-application-salt-16-bytes"),
			shouldError: false,
		},
		{
			name:        "Empty plaintext",
			plaintext:   []byte{},
			appSalt:     []byte("test-application-salt-16-bytes"),
			shouldError: true,
		},
		{
			name:        "Short app salt",
			plaintext:   []byte("test data"),
			appSalt:     []byte("short"),
			shouldError: true,
		},
		{
			name:        "Large plaintext",
			plaintext:   make([]byte, 64*1024), // 64KB
			appSalt:     []byte("test-application-salt-16-bytes"),
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill large plaintext with test data
			if len(tt.plaintext) == 64*1024 {
				for i := range tt.plaintext {
					tt.plaintext[i] = byte(i % 256)
				}
			}

			config := DefaultEncryptionConfig()

			// Encrypt
			payload, err := EncryptCredentials(tt.plaintext, tt.appSalt, config)
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Validate payload structure
			if payload.Version != 1 {
				t.Errorf("Expected version 1, got %d", payload.Version)
			}
			if len(payload.Salt) != 32 {
				t.Errorf("Expected salt length 32, got %d", len(payload.Salt))
			}
			if len(payload.Nonce) != 12 {
				t.Errorf("Expected nonce length 12, got %d", len(payload.Nonce))
			}
			if len(payload.AuthTag) != 16 {
				t.Errorf("Expected auth tag length 16, got %d", len(payload.AuthTag))
			}

			// Decrypt
			credentials, err := DecryptCredentials(payload, tt.appSalt, config)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}
			defer credentials.Clear()

			// Verify plaintext
			if !bytes.Equal(credentials.Data(), tt.plaintext) {
				t.Error("Decrypted data does not match original")
			}
		})
	}
}

func TestSecureCredentials(t *testing.T) {
	plaintext := []byte("sensitive credential data")
	appSalt := []byte("test-application-salt-32-bytes!!")

	payload, err := EncryptCredentials(plaintext, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	credentials, err := DecryptCredentials(payload, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Test data access before clearing
	data := credentials.Data()
	if !bytes.Equal(data, plaintext) {
		t.Error("Data does not match original")
	}

	// Test clearing
	credentials.Clear()

	// Test data access after clearing (should return nil)
	clearedData := credentials.Data()
	if clearedData != nil {
		t.Error("Data should be nil after clearing")
	}

	// Test double clear (should not panic)
	credentials.Clear()
}

func TestEncryptionConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *EncryptionConfig
		shouldError bool
	}{
		{
			name:        "Valid config",
			config:      DefaultEncryptionConfig(),
			shouldError: false,
		},
		{
			name:        "Nil config",
			config:      nil,
			shouldError: true,
		},
		{
			name: "Weak SCRYPT N",
			config: &EncryptionConfig{
				SCryptN:      16384, // Below minimum
				SCryptR:      8,
				SCryptP:      1,
				SCryptKeyLen: 32,
				NonceSize:    12,
				TagSize:      16,
			},
			shouldError: true,
		},
		{
			name: "Invalid key length",
			config: &EncryptionConfig{
				SCryptN:      32768,
				SCryptR:      8,
				SCryptP:      1,
				SCryptKeyLen: 16, // Should be 32 for AES-256
				NonceSize:    12,
				TagSize:      16,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEncryptionConfig(tt.config)
			if tt.shouldError {
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

func TestIntegrityProtection(t *testing.T) {
	plaintext := []byte("test data")
	appSalt := []byte("test-application-salt-32-bytes!!")

	payload, err := EncryptCredentials(plaintext, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Test tampering with different parts of the payload
	tests := []struct {
		name   string
		tamper func(*EncryptedPayload)
	}{
		{
			name: "Tampered ciphertext",
			tamper: func(p *EncryptedPayload) {
				if len(p.Ciphertext) > 0 {
					p.Ciphertext[0] ^= 0x01
				}
			},
		},
		{
			name: "Tampered salt",
			tamper: func(p *EncryptedPayload) {
				if len(p.Salt) > 0 {
					p.Salt[0] ^= 0x01
				}
			},
		},
		{
			name: "Tampered nonce",
			tamper: func(p *EncryptedPayload) {
				if len(p.Nonce) > 0 {
					p.Nonce[0] ^= 0x01
				}
			},
		},
		{
			name: "Tampered auth tag",
			tamper: func(p *EncryptedPayload) {
				if len(p.AuthTag) > 0 {
					p.AuthTag[0] ^= 0x01
				}
			},
		},
		{
			name: "Tampered integrity hash",
			tamper: func(p *EncryptedPayload) {
				if len(p.Integrity) > 0 {
					p.Integrity[0] ^= 0x01
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the payload
			tamperedPayload := *payload
			tamperedPayload.Salt = make([]byte, len(payload.Salt))
			copy(tamperedPayload.Salt, payload.Salt)
			tamperedPayload.Nonce = make([]byte, len(payload.Nonce))
			copy(tamperedPayload.Nonce, payload.Nonce)
			tamperedPayload.Ciphertext = make([]byte, len(payload.Ciphertext))
			copy(tamperedPayload.Ciphertext, payload.Ciphertext)
			tamperedPayload.AuthTag = make([]byte, len(payload.AuthTag))
			copy(tamperedPayload.AuthTag, payload.AuthTag)
			tamperedPayload.Integrity = make([]byte, len(payload.Integrity))
			copy(tamperedPayload.Integrity, payload.Integrity)

			// Apply tampering
			tt.tamper(&tamperedPayload)

			// Attempt decryption (should fail)
			_, err := DecryptCredentials(&tamperedPayload, appSalt, DefaultEncryptionConfig())
			if err == nil {
				t.Error("Expected decryption to fail with tampered data")
			}
		})
	}
}

func TestKeyDerivationSecurity(t *testing.T) {
	plaintext := []byte("test data")
	appSalt1 := []byte("salt1-application-salt-32-bytes!!")
	appSalt2 := []byte("salt2-application-salt-32-bytes!!")

	// Encrypt with first salt
	payload1, err := EncryptCredentials(plaintext, appSalt1, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Encryption 1 failed: %v", err)
	}

	// Encrypt with second salt
	payload2, err := EncryptCredentials(plaintext, appSalt2, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Encryption 2 failed: %v", err)
	}

	// Ciphertexts should be different (different keys derived)
	if bytes.Equal(payload1.Ciphertext, payload2.Ciphertext) {
		t.Error("Ciphertexts should be different with different app salts")
	}

	// Cannot decrypt payload1 with appSalt2
	_, err = DecryptCredentials(payload1, appSalt2, DefaultEncryptionConfig())
	if err == nil {
		t.Error("Should not be able to decrypt with wrong app salt")
	}

	// Cannot decrypt payload2 with appSalt1
	_, err = DecryptCredentials(payload2, appSalt1, DefaultEncryptionConfig())
	if err == nil {
		t.Error("Should not be able to decrypt with wrong app salt")
	}
}

func TestMemoryProtection(t *testing.T) {
	plaintext := []byte("sensitive data that should be cleared")
	appSalt := []byte("test-application-salt-32-bytes!!")

	payload, err := EncryptCredentials(plaintext, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	credentials, err := DecryptCredentials(payload, appSalt, DefaultEncryptionConfig())
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Get reference to data before clearing
	data := credentials.Data()
	if len(data) == 0 {
		t.Fatal("Data should not be empty")
	}

	// Clear credentials
	credentials.Clear()

	// Verify data is cleared (this is a basic check - in real scenarios,
	// the underlying memory should also be overwritten)
	clearedData := credentials.Data()
	if clearedData != nil {
		t.Error("Data should be nil after clearing")
	}

	// Multiple clears should be safe
	credentials.Clear()
	credentials.Clear()
}

func TestConcurrentEncryption(t *testing.T) {
	plaintext := []byte("test data for concurrent encryption")
	appSalt := []byte("test-application-salt-32-bytes!!")
	config := DefaultEncryptionConfig()

	// Test concurrent encryptions with proper cleanup
	const numGoroutines = 10
	results := make(chan error, numGoroutines)
	
	// Use a timeout to prevent test hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() {
				// Ensure we always send to results channel to prevent goroutine leak
				if r := recover(); r != nil {
					results <- fmt.Errorf("panic: %v", r)
				}
			}()
			
			payload, err := EncryptCredentials(plaintext, appSalt, config)
			if err != nil {
				results <- err
				return
			}

			credentials, err := DecryptCredentials(payload, appSalt, config)
			if err != nil {
				results <- err
				return
			}
			defer credentials.Clear()

			if !bytes.Equal(credentials.Data(), plaintext) {
				results <- fmt.Errorf("decrypted data does not match original")
				return
			}

			results <- nil
		}()
	}

	// Wait for all goroutines with timeout
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent encryption failed: %v", err)
			}
		case <-ctx.Done():
			t.Fatalf("Test timeout after 10 seconds, %d goroutines may have leaked", numGoroutines-i)
		}
	}
	
	// Ensure all goroutines are done by closing the channel and verifying it's empty
	close(results)
	select {
	case <-results:
		// Channel should be empty now
	default:
		// This is expected - channel is empty
	}
}

func TestRandomnessQuality(t *testing.T) {
	// Test that encryption produces different outputs for same input
	plaintext := []byte("same input data")
	appSalt := []byte("test-application-salt-32-bytes!!")
	config := DefaultEncryptionConfig()

	const numTests = 10
	ciphertexts := make([][]byte, numTests)
	salts := make([][]byte, numTests)
	nonces := make([][]byte, numTests)

	for i := 0; i < numTests; i++ {
		payload, err := EncryptCredentials(plaintext, appSalt, config)
		if err != nil {
			t.Fatalf("Encryption %d failed: %v", i, err)
		}

		ciphertexts[i] = payload.Ciphertext
		salts[i] = payload.Salt
		nonces[i] = payload.Nonce
	}

	// All salts should be different
	for i := 0; i < numTests; i++ {
		for j := i + 1; j < numTests; j++ {
			if bytes.Equal(salts[i], salts[j]) {
				t.Error("Salt collision detected - randomness quality issue")
			}
			if bytes.Equal(nonces[i], nonces[j]) {
				t.Error("Nonce collision detected - randomness quality issue")
			}
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Error("Ciphertext collision detected - randomness quality issue")
			}
		}
	}
}

func BenchmarkEncryption(b *testing.B) {
	plaintext := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n"
	}`)
	appSalt := []byte("test-application-salt-32-bytes!!")
	config := DefaultEncryptionConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptCredentials(plaintext, appSalt, config)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkDecryption(b *testing.B) {
	plaintext := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n"
	}`)
	appSalt := []byte("test-application-salt-32-bytes!!")
	config := DefaultEncryptionConfig()

	// Pre-encrypt the data
	payload, err := EncryptCredentials(plaintext, appSalt, config)
	if err != nil {
		b.Fatalf("Setup encryption failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		credentials, err := DecryptCredentials(payload, appSalt, config)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
		credentials.Clear()
	}
}