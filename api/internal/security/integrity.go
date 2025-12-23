package security

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// IntegrityChecker provides binary integrity verification and anti-tampering detection
type IntegrityChecker struct {
	expectedHash   string
	binaryPath     string
	verificationResult *VerificationResult
}

// VerificationResult holds the result of integrity verification
type VerificationResult struct {
	IsValid       bool
	ActualHash    string
	ExpectedHash  string
	BinaryPath    string
	BinarySize    int64
	ErrorMessage  string
	TamperingDetected bool
}

// TamperingIndicators holds various tampering detection mechanisms
type TamperingIndicators struct {
	UnexpectedFileSize    bool
	ModifiedTimestamp     bool
	SuspiciousProcessName bool
	DebuggerDetected      bool
	VirtualMachineDetected bool
}

// NewIntegrityChecker creates a new integrity checker with expected binary hash
func NewIntegrityChecker(expectedHash string) *IntegrityChecker {
	return &IntegrityChecker{
		expectedHash: strings.ToLower(expectedHash),
	}
}

// VerifyBinaryIntegrity performs comprehensive binary integrity verification
func (ic *IntegrityChecker) VerifyBinaryIntegrity() (*VerificationResult, error) {
	// Get current executable path
	binaryPath, err := os.Executable()
	if err != nil {
		return &VerificationResult{
			IsValid:      false,
			ErrorMessage: fmt.Sprintf("failed to get executable path: %v", err),
			TamperingDetected: true,
		}, err
	}
	
	ic.binaryPath = binaryPath
	
	// Calculate current binary hash
	actualHash, fileSize, err := ic.calculateBinaryHash(binaryPath)
	if err != nil {
		return &VerificationResult{
			IsValid:      false,
			BinaryPath:   binaryPath,
			ErrorMessage: fmt.Sprintf("failed to calculate binary hash: %v", err),
			TamperingDetected: true,
		}, err
	}
	
	// Compare hashes
	isValid := strings.EqualFold(actualHash, ic.expectedHash)
	
	// Check for additional tampering indicators
	tampering := ic.detectTamperingIndicators()
	
	result := &VerificationResult{
		IsValid:       isValid && !tampering.hasIndicators(),
		ActualHash:    actualHash,
		ExpectedHash:  ic.expectedHash,
		BinaryPath:    binaryPath,
		BinarySize:    fileSize,
		TamperingDetected: !isValid || tampering.hasIndicators(),
	}
	
	if !isValid {
		result.ErrorMessage = "binary hash mismatch - possible tampering detected"
	} else if tampering.hasIndicators() {
		result.ErrorMessage = "tampering indicators detected"
	}
	
	ic.verificationResult = result
	return result, nil
}

// calculateBinaryHash computes SHA-256 hash of the binary file
func (ic *IntegrityChecker) calculateBinaryHash(filePath string) (string, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	
	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, err
	}
	fileSize := fileInfo.Size()
	
	// Calculate SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", 0, err
	}
	
	hash := hex.EncodeToString(hasher.Sum(nil))
	return strings.ToLower(hash), fileSize, nil
}

// detectTamperingIndicators checks for various tampering indicators
func (ic *IntegrityChecker) detectTamperingIndicators() *TamperingIndicators {
	indicators := &TamperingIndicators{}
	
	// Check for debugger presence
	indicators.DebuggerDetected = ic.isDebuggerPresent()
	
	// Check for virtual machine indicators
	indicators.VirtualMachineDetected = ic.isRunningInVM()
	
	// Check for suspicious process name
	indicators.SuspiciousProcessName = ic.hasSuspiciousProcessName()
	
	return indicators
}

// isDebuggerPresent detects if a debugger is attached (Windows/Go specific)
func (ic *IntegrityChecker) isDebuggerPresent() bool {
	// For Go applications, check for common debugging flags
	if runtime.GOOS == "windows" {
		// Check for delve debugger or common debugging environment variables
		debugVars := []string{
			"DELVE_PORT",
			"DEBUG_MODE", 
			"GO_DEBUG",
			"GODEBUG",
		}
		
		for _, debugVar := range debugVars {
			if os.Getenv(debugVar) != "" {
				return true
			}
		}
	}
	
	return false
}

// isRunningInVM detects if running in a virtual machine
func (ic *IntegrityChecker) isRunningInVM() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// Check for common VM indicators in environment
	vmIndicators := []string{
		"VBOX_",
		"VMWARE_",
		"VIRTUAL_",
	}
	
	for _, key := range os.Environ() {
		upperKey := strings.ToUpper(key)
		for _, indicator := range vmIndicators {
			if strings.Contains(upperKey, indicator) {
				return true
			}
		}
	}
	
	return false
}

// hasSuspiciousProcessName checks if the process has been renamed suspiciously
func (ic *IntegrityChecker) hasSuspiciousProcessName() bool {
	if ic.binaryPath == "" {
		return false
	}
	
	// Get the base name of the executable
	parts := strings.Split(ic.binaryPath, string(os.PathSeparator))
	if len(parts) == 0 {
		return false
	}
	
	baseName := strings.ToLower(parts[len(parts)-1])
	
	// Check for suspicious names that don't match expected ISX patterns
	expectedNames := []string{
		"web-licensed.exe",
		"web-licensed",
		"isx-daily-reports",
		"isxdailyreports",
	}
	
	for _, expected := range expectedNames {
		if strings.Contains(baseName, strings.ToLower(expected)) {
			return false
		}
	}
	
	// Check for obviously suspicious names
	suspiciousPatterns := []string{
		"debug",
		"test",
		"crack",
		"hack",
		"bypass",
		"temp",
		"copy",
	}
	
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(baseName, pattern) {
			return true
		}
	}
	
	return false
}

// hasIndicators returns true if any tampering indicators are present
func (ti *TamperingIndicators) hasIndicators() bool {
	return ti.UnexpectedFileSize ||
		   ti.ModifiedTimestamp ||
		   ti.SuspiciousProcessName ||
		   ti.DebuggerDetected ||
		   ti.VirtualMachineDetected
}

// GetDetailedTamperingReport returns a detailed report of tampering indicators
func (ti *TamperingIndicators) GetDetailedTamperingReport() string {
	var report strings.Builder
	
	if !ti.hasIndicators() {
		return "No tampering indicators detected"
	}
	
	report.WriteString("Tampering indicators detected:\n")
	
	if ti.UnexpectedFileSize {
		report.WriteString("- Unexpected file size\n")
	}
	if ti.ModifiedTimestamp {
		report.WriteString("- Modified timestamp\n")
	}
	if ti.SuspiciousProcessName {
		report.WriteString("- Suspicious process name\n")
	}
	if ti.DebuggerDetected {
		report.WriteString("- Debugger presence detected\n")
	}
	if ti.VirtualMachineDetected {
		report.WriteString("- Virtual machine environment detected\n")
	}
	
	return report.String()
}

// VerifyAndDecryptCredentials combines integrity verification with credential decryption
func VerifyAndDecryptCredentials(payload *EncryptedPayload, appSalt []byte, expectedBinaryHash string) (*SecureCredentials, error) {
	// First verify binary integrity
	checker := NewIntegrityChecker(expectedBinaryHash)
	result, err := checker.VerifyBinaryIntegrity()
	if err != nil {
		return nil, fmt.Errorf("integrity verification failed: %v", err)
	}
	
	if !result.IsValid {
		return nil, fmt.Errorf("binary integrity verification failed: %s", result.ErrorMessage)
	}
	
	// If integrity check passes, proceed with credential decryption
	credentials, err := DecryptCredentials(payload, appSalt, DefaultEncryptionConfig())
	if err != nil {
		return nil, fmt.Errorf("credential decryption failed: %v", err)
	}
	
	return credentials, nil
}

// GenerateBinaryHash generates the SHA-256 hash of the current binary for build-time embedding
func GenerateBinaryHash() (string, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	
	hash, _, err := (&IntegrityChecker{}).calculateBinaryHash(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate binary hash: %v", err)
	}
	
	return hash, nil
}

// ValidateIntegrityConfig validates integrity checker configuration
func ValidateIntegrityConfig(expectedHash string) error {
	if expectedHash == "" {
		return errors.New("expected hash cannot be empty")
	}
	
	if len(expectedHash) != 64 {
		return errors.New("expected hash must be 64 characters (SHA-256)")
	}
	
	// Validate hex encoding
	if _, err := hex.DecodeString(expectedHash); err != nil {
		return fmt.Errorf("expected hash must be valid hex: %v", err)
	}
	
	return nil
}