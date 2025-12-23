package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DeviceFingerprint represents device identification information
type DeviceFingerprint struct {
	Fingerprint string `json:"fingerprint"`
	Hostname    string `json:"hostname"`
	MACAddress  string `json:"mac_address"`
	CPUID       string `json:"cpu_id"`
	OS          string `json:"os"`
	Platform    string `json:"platform"`
	GeneratedAt time.Time `json:"generated_at"`
}

// FingerprintManager handles device fingerprinting operations
type FingerprintManager struct {
	cache          *DeviceFingerprint
	cacheMutex     sync.RWMutex
	cacheExpiry    time.Time
	cacheDuration  time.Duration
}

// NewFingerprintManager creates a new fingerprint manager with caching
func NewFingerprintManager() *FingerprintManager {
	return &FingerprintManager{
		cacheDuration: 1 * time.Hour, // Cache fingerprint for 1 hour
	}
}

// GetMACAddress retrieves the primary network interface MAC address
func (fm *FingerprintManager) GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Look for the first non-loopback, up interface with a MAC address
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Check if interface has a valid MAC address
		if len(iface.HardwareAddr) > 0 {
			mac := iface.HardwareAddr.String()
			if mac != "" && mac != "00:00:00:00:00:00" {
				slog.Debug("MAC address found",
					slog.String("interface", iface.Name),
					slog.String("mac", mac),
					slog.String("flags", iface.Flags.String()),
				)
				return mac, nil
			}
		}
	}

	// Fallback: use any interface with MAC address
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 {
			mac := iface.HardwareAddr.String()
			if mac != "" && mac != "00:00:00:00:00:00" {
				slog.Warn("Using fallback MAC address",
					slog.String("interface", iface.Name),
					slog.String("mac", mac),
				)
				return mac, nil
			}
		}
	}

	return "", fmt.Errorf("no valid MAC address found")
}

// GetHostname retrieves the machine hostname
func (fm *FingerprintManager) GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Normalize hostname (lowercase, trim spaces)
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	if hostname == "" {
		return "", fmt.Errorf("hostname is empty")
	}

	return hostname, nil
}

// GetCPUID retrieves CPU identification information (OS-specific)
func (fm *FingerprintManager) GetCPUID() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return fm.getCPUIDWindows()
	case "linux":
		return fm.getCPUIDLinux()
	case "darwin":
		return fm.getCPUIDDarwin()
	default:
		// Fallback: use runtime architecture and OS
		cpuInfo := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		slog.Warn("Using fallback CPU ID for unsupported OS",
			slog.String("os", runtime.GOOS),
			slog.String("arch", runtime.GOARCH),
			slog.String("cpu_id", cpuInfo),
		)
		return cpuInfo, nil
	}
}

// getCPUIDWindows gets CPU information on Windows systems
func (fm *FingerprintManager) getCPUIDWindows() (string, error) {
	// Try to read processor identifier from environment
	if procId := os.Getenv("PROCESSOR_IDENTIFIER"); procId != "" {
		// Hash the processor identifier to normalize length
		hash := sha256.Sum256([]byte(procId))
		cpuId := hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
		
		slog.Debug("Windows CPU ID from PROCESSOR_IDENTIFIER",
			slog.String("raw", procId),
			slog.String("cpu_id", cpuId),
		)
		return cpuId, nil
	}

	// Fallback to architecture info
	cpuInfo := fmt.Sprintf("windows-%s-%s", runtime.GOARCH, os.Getenv("PROCESSOR_ARCHITECTURE"))
	hash := sha256.Sum256([]byte(cpuInfo))
	cpuId := hex.EncodeToString(hash[:8])
	
	slog.Debug("Windows CPU ID fallback",
		slog.String("raw", cpuInfo),
		slog.String("cpu_id", cpuId),
	)
	return cpuId, nil
}

// getCPUIDLinux gets CPU information on Linux systems
func (fm *FingerprintManager) getCPUIDLinux() (string, error) {
	// Try to read from /proc/cpuinfo
	cpuData, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		lines := strings.Split(string(cpuData), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "processor") || 
			   strings.HasPrefix(line, "model name") ||
			   strings.HasPrefix(line, "cpu family") {
				// Hash the CPU info to normalize length
				hash := sha256.Sum256([]byte(line))
				cpuId := hex.EncodeToString(hash[:8])
				
				slog.Debug("Linux CPU ID from /proc/cpuinfo",
					slog.String("raw", line),
					slog.String("cpu_id", cpuId),
				)
				return cpuId, nil
			}
		}
	}

	// Fallback to architecture info
	cpuInfo := fmt.Sprintf("linux-%s", runtime.GOARCH)
	hash := sha256.Sum256([]byte(cpuInfo))
	cpuId := hex.EncodeToString(hash[:8])
	
	slog.Debug("Linux CPU ID fallback",
		slog.String("raw", cpuInfo),
		slog.String("cpu_id", cpuId),
	)
	return cpuId, nil
}

// getCPUIDDarwin gets CPU information on macOS systems
func (fm *FingerprintManager) getCPUIDDarwin() (string, error) {
	// For macOS, use system info that's relatively stable
	cpuInfo := fmt.Sprintf("darwin-%s", runtime.GOARCH)
	
	// Try to get more specific info if available
	if procType := os.Getenv("HOSTTYPE"); procType != "" {
		cpuInfo = fmt.Sprintf("darwin-%s-%s", runtime.GOARCH, procType)
	}
	
	hash := sha256.Sum256([]byte(cpuInfo))
	cpuId := hex.EncodeToString(hash[:8])
	
	slog.Debug("macOS CPU ID",
		slog.String("raw", cpuInfo),
		slog.String("cpu_id", cpuId),
	)
	return cpuId, nil
}

// GenerateFingerprint creates a device fingerprint by combining hardware factors
func (fm *FingerprintManager) GenerateFingerprint() (*DeviceFingerprint, error) {
	// Check cache first
	fm.cacheMutex.RLock()
	if fm.cache != nil && time.Now().Before(fm.cacheExpiry) {
		cachedFingerprint := *fm.cache // Create a copy
		fm.cacheMutex.RUnlock()
		
		slog.Debug("Using cached device fingerprint",
			slog.String("fingerprint", cachedFingerprint.Fingerprint),
			slog.Time("cached_at", cachedFingerprint.GeneratedAt),
		)
		return &cachedFingerprint, nil
	}
	fm.cacheMutex.RUnlock()

	start := time.Now()
	slog.Debug("Generating new device fingerprint")

	// Get MAC address
	macAddr, err := fm.GetMACAddress()
	if err != nil {
		macAddr = "unknown-mac"
		slog.Warn("Failed to get MAC address, using fallback",
			slog.String("error", err.Error()),
		)
	}

	// Get hostname
	hostname, err := fm.GetHostname()
	if err != nil {
		hostname = "unknown-host"
		slog.Warn("Failed to get hostname, using fallback",
			slog.String("error", err.Error()),
		)
	}

	// Get CPU ID
	cpuID, err := fm.GetCPUID()
	if err != nil {
		cpuID = "unknown-cpu"
		slog.Warn("Failed to get CPU ID, using fallback",
			slog.String("error", err.Error()),
		)
	}

	// Combine factors
	factors := []string{
		macAddr,
		hostname,
		cpuID,
		runtime.GOOS,
		runtime.GOARCH,
	}

	// Create fingerprint string
	combinedData := strings.Join(factors, "|")
	hash := sha256.Sum256([]byte(combinedData))
	fingerprint := hex.EncodeToString(hash[:])

	// Create device fingerprint object
	deviceFingerprint := &DeviceFingerprint{
		Fingerprint: fingerprint,
		Hostname:    hostname,
		MACAddress:  macAddr,
		CPUID:       cpuID,
		OS:          runtime.GOOS,
		Platform:    runtime.GOARCH,
		GeneratedAt: time.Now(),
	}

	// Cache the result
	fm.cacheMutex.Lock()
	fm.cache = deviceFingerprint
	fm.cacheExpiry = time.Now().Add(fm.cacheDuration)
	fm.cacheMutex.Unlock()

	duration := time.Since(start)
	slog.Info("Device fingerprint generated successfully",
		slog.String("fingerprint", fingerprint),
		slog.String("hostname", hostname),
		slog.String("mac_address", macAddr),
		slog.String("cpu_id", cpuID),
		slog.String("os", runtime.GOOS),
		slog.String("platform", runtime.GOARCH),
		slog.Duration("generation_time", duration),
	)

	return deviceFingerprint, nil
}

// ValidateFingerprint compares current device fingerprint with stored one
func (fm *FingerprintManager) ValidateFingerprint(storedFingerprint string) (bool, error) {
	current, err := fm.GenerateFingerprint()
	if err != nil {
		return false, fmt.Errorf("failed to generate current fingerprint: %w", err)
	}

	matches := current.Fingerprint == storedFingerprint
	
	slog.Debug("Device fingerprint validation",
		slog.String("stored", storedFingerprint),
		slog.String("current", current.Fingerprint),
		slog.Bool("matches", matches),
	)

	return matches, nil
}

// GetFingerprintComponents returns individual components for debugging
func (fm *FingerprintManager) GetFingerprintComponents() (map[string]string, error) {
	macAddr, _ := fm.GetMACAddress()
	hostname, _ := fm.GetHostname()
	cpuID, _ := fm.GetCPUID()

	components := map[string]string{
		"mac_address": macAddr,
		"hostname":    hostname,
		"cpu_id":      cpuID,
		"os":          runtime.GOOS,
		"platform":    runtime.GOARCH,
	}

	return components, nil
}

// ClearCache clears the cached fingerprint
func (fm *FingerprintManager) ClearCache() {
	fm.cacheMutex.Lock()
	defer fm.cacheMutex.Unlock()
	
	fm.cache = nil
	fm.cacheExpiry = time.Time{}
	
	slog.Debug("Device fingerprint cache cleared")
}