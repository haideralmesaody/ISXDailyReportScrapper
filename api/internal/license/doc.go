// Package license implements comprehensive license management for the ISX system.
// It provides hardware-based license validation, activation workflows, and
// integration with the central license server following enterprise security standards.
//
// # Architecture Overview
//
// The license system consists of several components:
//
//   - Manager: Core license validation and activation logic
//   - State: Persistent storage of license information
//   - Security: Hardware fingerprinting and encryption
//   - Cache: In-memory caching for performance
//   - Health: License system health monitoring
//
// # License Validation Flow
//
// The validation process follows these steps:
//
//  1. Load stored license from encrypted file
//  2. Validate license format and signature
//  3. Check hardware fingerprint match
//  4. Verify expiration date
//  5. Cache result for performance
//
// # Hardware Fingerprinting
//
// The system generates a unique hardware fingerprint using:
//
//   - MAC addresses (primary network interfaces)
//   - CPU information (model, cores)
//   - System UUID (where available)
//   - Disk serial numbers
//
// This ensures licenses are bound to specific hardware and cannot
// be copied between machines.
//
// # Activation Process
//
// License activation follows this workflow:
//
//	key := "ISX1M02LYE1F9QJHR9D7Z"
//	license, err := manager.ActivateLicense(ctx, key)
//
// The activation process:
//  1. Validates key format (20 chars, alphanumeric)
//  2. Generates hardware fingerprint
//  3. Contacts activation server
//  4. Stores encrypted license data
//  5. Returns activation status
//
// # Security Measures
//
// The package implements multiple security layers:
//
//   - AES-256-GCM encryption for stored licenses
//   - HMAC signature verification
//   - Certificate pinning for server communication
//   - Rate limiting on activation attempts
//   - Secure key derivation (PBKDF2)
//
// # Offline Support
//
// For environments without internet access:
//
//  1. Generate activation request with hardware info
//  2. Transfer request to online system
//  3. Receive activation response
//  4. Apply response to activate offline
//
// # Integration
//
// The license package integrates with:
//
//   - HTTP middleware for request validation
//   - Health checks for monitoring
//   - Metrics for usage tracking
//   - WebSocket for real-time updates
//
// # Error Handling
//
// The package defines specific error types:
//
//   - ErrInvalidLicense: License format or signature invalid
//   - ErrExpiredLicense: License has expired
//   - ErrHardwareMismatch: Hardware fingerprint doesn't match
//   - ErrActivationFailed: Server communication failed
//
// # Performance
//
// Performance optimizations include:
//
//   - In-memory caching of validation results
//   - Lazy loading of license data
//   - Minimal allocations in hot paths
//   - Concurrent-safe operations
//
// # Testing
//
// The package provides test utilities:
//
//   - Mock license manager for unit tests
//   - Test fixtures for various scenarios
//   - Hardware fingerprint simulation
//   - Time-based testing helpers
package license
