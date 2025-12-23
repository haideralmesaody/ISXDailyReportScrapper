//go:build windows

package security

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// getMotherboardSerial retrieves motherboard serial number on Windows
func (hf *HardwareFingerprinter) getMotherboardSerial() (string, error) {
	slog.Debug("Starting motherboard serial detection", "component", "motherboard_serial")

	// PowerShell is the primary method for motherboard serial detection
	slog.Debug("Trying PowerShell for motherboard serial", "method", "powershell")
	if serial, err := hf.getPowerShellBoardSerial(); err == nil && serial != "" {
		slog.Info("Motherboard serial found via PowerShell",
			"serial", serial,
			"method", "powershell")
		return serial, nil
	} else {
		slog.Info("PowerShell motherboard serial returned no data",
			"error", err.Error(),
			"method", "powershell")
	}

	// No WMIC fallback - WMIC is deprecated on modern Windows
	slog.Info("Motherboard serial detection failed - PowerShell method unavailable",
		"component", "motherboard_serial",
		"note", "motherboard_serial is critical for license validation")

	return "", fmt.Errorf("failed to get motherboard serial from PowerShell")
}

// getPowerShellBoardSerial uses PowerShell to get motherboard serial
func (hf *HardwareFingerprinter) getPowerShellBoardSerial() (string, error) {
	script := `Get-WmiObject -Class Win32_BaseBoard | Select-Object -ExpandProperty SerialNumber`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("powershell command failed: %w", err)
	}

	serial := strings.TrimSpace(string(output))
	if serial != "" && !strings.Contains(strings.ToLower(serial), "to be filled") {
		return serial, nil
	}

	return "", fmt.Errorf("no valid serial number found in powershell output")
}

// getSystemUUID retrieves system UUID on Windows
func (hf *HardwareFingerprinter) getSystemUUID() (string, error) {
	slog.Debug("Starting system UUID detection", "component", "system_uuid")

	// Method 1: Try PowerShell first (most reliable on modern Windows)
	slog.Debug("Trying PowerShell for system UUID", "method", "powershell")
	if uuid, err := hf.getPowerShellSystemUUID(); err == nil && uuid != "" {
		slog.Info("System UUID found via PowerShell",
			"uuid", uuid,
			"method", "powershell")
		return uuid, nil
	} else {
		slog.Info("PowerShell system UUID returned no data",
			"error", err.Error(),
			"method", "powershell")
	}

	// Method 2: Try registry as fallback
	slog.Debug("Trying Registry for system UUID", "method", "registry")
	if uuid, err := hf.getRegistrySystemUUID(); err == nil && uuid != "" {
		slog.Info("System UUID found via Registry",
			"uuid", uuid,
			"method", "registry")
		return uuid, nil
	} else {
		slog.Info("Registry system UUID returned no data",
			"error", err.Error(),
			"method", "registry")
	}

	// WMIC is no longer available - log at INFO level
	slog.Info("All system UUID methods returned no data - WMIC is deprecated",
		"component", "system_uuid",
		"note", "system_uuid is critical for license validation")

	return "", fmt.Errorf("failed to get system UUID from any source")
}


// getPowerShellSystemUUID uses PowerShell to get system UUID
func (hf *HardwareFingerprinter) getPowerShellSystemUUID() (string, error) {
	script := `Get-CimInstance -ClassName Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("powershell UUID command failed: %w", err)
	}

	uuid := strings.TrimSpace(string(output))
	if uuid != "" && len(uuid) >= 32 {
		uuid = strings.ReplaceAll(uuid, "{", "")
		uuid = strings.ReplaceAll(uuid, "}", "")
		return strings.ToLower(uuid), nil
	}

	return "", fmt.Errorf("no valid UUID found in powershell output")
}

// getRegistrySystemUUID reads UUID from Windows registry
func (hf *HardwareFingerprinter) getRegistrySystemUUID() (string, error) {
	cmd := exec.Command("reg", "query", `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("reg query failed: %w", err)
	}

	// Parse output like: MachineGuid    REG_SZ    xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	re := regexp.MustCompile(`MachineGuid\s+REG_SZ\s+([a-fA-F0-9-]{36})`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		uuid := matches[1]
		slog.Debug("Retrieved system UUID via registry", slog.String("uuid", uuid))
		return strings.ToLower(uuid), nil
	}

	return "", fmt.Errorf("no valid UUID found in registry output")
}

// getPrimaryDiskSerial retrieves primary disk serial number on Windows
func (hf *HardwareFingerprinter) getPrimaryDiskSerial() (string, error) {
	slog.Debug("Starting primary disk serial detection", "component", "disk_serial")

	// Method 1: Try PowerShell CIM for system drive serial (most reliable)
	slog.Debug("Trying PowerShell CIM for system drive serial", "method", "powershell_cim_system")
	if serial, err := hf.getPowerShellSystemDriveSerial(); err == nil && serial != "" {
		slog.Info("Disk serial found via PowerShell CIM system drive",
			"serial", serial,
			"method", "powershell_cim_system")
		return serial, nil
	} else {
		slog.Info("PowerShell CIM system drive serial returned no data",
			"error", err.Error(),
			"method", "powershell_cim_system")
	}

	// Method 2: Try PowerShell CIM for physical disk serial
	slog.Debug("Trying PowerShell CIM for physical disk serial", "method", "powershell_cim_physical")
	if serial, err := hf.getPowerShellPhysicalDiskSerial(); err == nil && serial != "" {
		slog.Info("Disk serial found via PowerShell CIM physical disk",
			"serial", serial,
			"method", "powershell_cim_physical")
		return serial, nil
	} else {
		slog.Info("PowerShell CIM physical disk serial returned no data",
			"error", err.Error(),
			"method", "powershell_cim_physical")
	}

	// Disk serial is considered optional - log at INFO level and continue
	slog.Info("All disk serial methods returned no data - disk serial will be omitted from fingerprint",
		"component", "disk_serial",
		"note", "disk_serial is optional for license validation")

	// Return empty string with a specific error that can be handled gracefully
	return "", fmt.Errorf("disk serial data unavailable")
}


// getCPUModel retrieves CPU model information on Windows
func (hf *HardwareFingerprinter) getCPUModel() (string, error) {
	slog.Debug("Starting CPU model detection", "component", "cpu_model")

	// Method 1: Try environment variable first (always available)
	slog.Debug("Trying environment variable for CPU model", "method", "env_var")
	if model := os.Getenv("PROCESSOR_IDENTIFIER"); model != "" {
		slog.Info("CPU model found via environment variable",
			"model", model,
			"method", "env_var")
		return model, nil
	} else {
		slog.Info("Environment variable CPU model returned no data",
			"error", "PROCESSOR_IDENTIFIER not set or empty",
			"method", "env_var")
	}

	// WMIC is no longer available - log at INFO level
	slog.Info("CPU model detection fallback to WMIC is deprecated",
		"component", "cpu_model",
		"note", "cpu_model is supplementary for license validation")

	return "", fmt.Errorf("failed to get CPU model from any source")
}


// getBIOSVersion retrieves BIOS version on Windows
func (hf *HardwareFingerprinter) getBIOSVersion() (string, error) {
	slog.Debug("Starting BIOS version detection", "component", "bios_version")

	// Method 1: Try PowerShell first (most reliable on modern Windows)
	slog.Debug("Trying PowerShell for BIOS version", "method", "powershell")
	if version, err := hf.getPowerShellBIOSVersion(); err == nil && version != "" {
		slog.Info("BIOS version found via PowerShell",
			"version", version,
			"method", "powershell")
		return version, nil
	} else {
		slog.Info("PowerShell BIOS version returned no data",
			"error", err.Error(),
			"method", "powershell")
	}

	// WMIC is no longer available - log at INFO level
	slog.Info("BIOS version detection fallback to WMIC is deprecated",
		"component", "bios_version",
		"note", "bios_version is supplementary for license validation")

	return "", fmt.Errorf("failed to get BIOS version from any source")
}


// getPowerShellBIOSVersion uses PowerShell to get BIOS version
func (hf *HardwareFingerprinter) getPowerShellBIOSVersion() (string, error) {
	script := `Get-WmiObject -Class Win32_BIOS | Select-Object -ExpandProperty SMBIOSBIOSVersion`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("powershell BIOS command failed: %w", err)
	}

	version := strings.TrimSpace(string(output))
	if version != "" {
		slog.Debug("Retrieved BIOS version via PowerShell", slog.String("version", version))
		return version, nil
	}

	return "", fmt.Errorf("no valid BIOS version found in powershell output")
}

// getPowerShellSystemDriveSerial uses PowerShell CIM to get system drive serial number
func (hf *HardwareFingerprinter) getPowerShellSystemDriveSerial() (string, error) {
	script := `Get-CimInstance -ClassName Win32_LogicalDisk | Where-Object {$_.DeviceID -eq 'C:'} | Select-Object -ExpandProperty VolumeSerialNumber`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("powershell CIM command failed: %w", err)
	}

	serial := strings.TrimSpace(string(output))
	if serial != "" {
		slog.Debug("Retrieved system drive serial via PowerShell CIM", slog.String("serial", serial))
		return strings.ToLower(serial), nil
	}

	return "", fmt.Errorf("no valid volume serial found in powershell CIM output")
}

// getPowerShellPhysicalDiskSerial uses PowerShell CIM to get physical disk serial number
func (hf *HardwareFingerprinter) getPowerShellPhysicalDiskSerial() (string, error) {
	script := `Get-CimInstance -ClassName Win32_DiskDrive | Select-Object -First 1 | Select-Object -ExpandProperty SerialNumber`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("powershell CIM command failed: %w", err)
	}

	serial := strings.TrimSpace(string(output))
	if serial != "" && !strings.Contains(strings.ToLower(serial), "to be filled") {
		slog.Debug("Retrieved physical disk serial via PowerShell CIM", slog.String("serial", serial))
		return serial, nil
	}

	return "", fmt.Errorf("no valid physical disk serial found in powershell CIM output")
}