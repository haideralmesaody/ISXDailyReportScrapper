//go:build darwin

package security

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// getMotherboardSerial retrieves motherboard serial number on macOS
func (hf *HardwareFingerprinter) getMotherboardSerial() (string, error) {
	// Method 1: Try system_profiler for hardware serial
	if serial, err := hf.getSystemProfilerSerial(); err == nil && serial != "" {
		return serial, nil
	}

	// Method 2: Try ioreg for platform serial
	if serial, err := hf.getIOPlatformSerial(); err == nil && serial != "" {
		return serial, nil
	}

	return "", fmt.Errorf("failed to get motherboard serial from any source")
}

// getSystemProfilerSerial uses system_profiler to get hardware serial
func (hf *HardwareFingerprinter) getSystemProfilerSerial() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler command failed: %w", err)
	}

	// Parse output for serial number
	re := regexp.MustCompile(`Serial Number \(system\):\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		serial := strings.TrimSpace(matches[1])
		if serial != "" {
			slog.Debug("Retrieved hardware serial via system_profiler", slog.String("serial", serial))
			return serial, nil
		}
	}

	return "", fmt.Errorf("no serial number found in system_profiler output")
}

// getIOPlatformSerial uses ioreg to get platform serial
func (hf *HardwareFingerprinter) getIOPlatformSerial() (string, error) {
	cmd := exec.Command("ioreg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ioreg command failed: %w", err)
	}

	// Parse output for IOPlatformSerialNumber
	re := regexp.MustCompile(`"IOPlatformSerialNumber"\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		serial := strings.TrimSpace(matches[1])
		if serial != "" {
			slog.Debug("Retrieved platform serial via ioreg", slog.String("serial", serial))
			return serial, nil
		}
	}

	return "", fmt.Errorf("no platform serial found in ioreg output")
}

// getSystemUUID retrieves system UUID on macOS
func (hf *HardwareFingerprinter) getSystemUUID() (string, error) {
	// Method 1: Try system_profiler for hardware UUID
	if uuid, err := hf.getSystemProfilerUUID(); err == nil && uuid != "" {
		return uuid, nil
	}

	// Method 2: Try ioreg for platform UUID
	if uuid, err := hf.getIOPlatformUUID(); err == nil && uuid != "" {
		return uuid, nil
	}

	return "", fmt.Errorf("failed to get system UUID from any source")
}

// getSystemProfilerUUID uses system_profiler to get hardware UUID
func (hf *HardwareFingerprinter) getSystemProfilerUUID() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler command failed: %w", err)
	}

	// Parse output for UUID
	re := regexp.MustCompile(`Hardware UUID:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		uuid := strings.TrimSpace(matches[1])
		if uuid != "" {
			slog.Debug("Retrieved hardware UUID via system_profiler", slog.String("uuid", uuid))
			return strings.ToLower(uuid), nil
		}
	}

	return "", fmt.Errorf("no UUID found in system_profiler output")
}

// getIOPlatformUUID uses ioreg to get platform UUID
func (hf *HardwareFingerprinter) getIOPlatformUUID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ioreg command failed: %w", err)
	}

	// Parse output for IOPlatformUUID
	re := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		uuid := strings.TrimSpace(matches[1])
		if uuid != "" {
			slog.Debug("Retrieved platform UUID via ioreg", slog.String("uuid", uuid))
			return strings.ToLower(uuid), nil
		}
	}

	return "", fmt.Errorf("no platform UUID found in ioreg output")
}

// getPrimaryDiskSerial retrieves primary disk serial number on macOS
func (hf *HardwareFingerprinter) getPrimaryDiskSerial() (string, error) {
	// Method 1: Try diskutil for system disk
	if serial, err := hf.getDiskutilSystemSerial(); err == nil && serial != "" {
		return serial, nil
	}

	// Method 2: Try system_profiler for storage
	if serial, err := hf.getSystemProfilerStorageSerial(); err == nil && serial != "" {
		return serial, nil
	}

	return "", fmt.Errorf("failed to get disk serial from any source")
}

// getDiskutilSystemSerial uses diskutil to get system disk serial
func (hf *HardwareFingerprinter) getDiskutilSystemSerial() (string, error) {
	// Get the system disk identifier
	cmd := exec.Command("diskutil", "info", "/")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("diskutil info command failed: %w", err)
	}

	// Parse output for device identifier
	re := regexp.MustCompile(`Device Identifier:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		device := strings.TrimSpace(matches[1])
		if device != "" {
			// Get serial for this device
			if serial, err := hf.getDiskutilDeviceSerial(device); err == nil && serial != "" {
				slog.Debug("Retrieved system disk serial via diskutil",
					slog.String("device", device),
					slog.String("serial", serial),
				)
				return serial, nil
			}
		}
	}

	return "", fmt.Errorf("no device identifier found in diskutil output")
}

// getDiskutilDeviceSerial gets serial for a specific disk device
func (hf *HardwareFingerprinter) getDiskutilDeviceSerial(device string) (string, error) {
	cmd := exec.Command("diskutil", "info", device)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("diskutil info command failed for %s: %w", device, err)
	}

	// Parse output for volume UUID or serial
	uuidRe := regexp.MustCompile(`Volume UUID:\s*(.+)`)
	serialRe := regexp.MustCompile(`Device / Media Name:\s*(.+)`)

	// Try UUID first
	if matches := uuidRe.FindStringSubmatch(string(output)); len(matches) >= 2 {
		uuid := strings.TrimSpace(matches[1])
		if uuid != "" {
			return strings.ToLower(uuid), nil
		}
	}

	// Try device/media name as fallback
	if matches := serialRe.FindStringSubmatch(string(output)); len(matches) >= 2 {
		serial := strings.TrimSpace(matches[1])
		if serial != "" {
			return strings.ToLower(serial), nil
		}
	}

	return "", fmt.Errorf("no serial found for device %s", device)
}

// getSystemProfilerStorageSerial uses system_profiler to get storage serial
func (hf *HardwareFingerprinter) getSystemProfilerStorageSerial() (string, error) {
	cmd := exec.Command("system_profiler", "SPStorageDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler storage command failed: %w", err)
	}

	// Parse output for BSD Name and then get serial
	lines := strings.Split(string(output), "\n")
	var currentDevice string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "BSD Name:") {
			if matches := regexp.MustCompile(`BSD Name:\s*(.+)`).FindStringSubmatch(line); len(matches) >= 2 {
				currentDevice = strings.TrimSpace(matches[1])
			}
		}

		// Look for serial or UUID for current device
		if currentDevice != "" && (strings.Contains(line, "Volume UUID:") || strings.Contains(line, "Serial Number:")) {
			if matches := regexp.MustCompile(`(Volume UUID|Serial Number):\s*(.+)`).FindStringSubmatch(line); len(matches) >= 3 {
				serial := strings.TrimSpace(matches[2])
				if serial != "" {
					slog.Debug("Retrieved storage serial via system_profiler",
						slog.String("device", currentDevice),
						slog.String("serial", serial),
					)
					return strings.ToLower(serial), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no storage serial found in system_profiler output")
}

// getCPUModel retrieves CPU model information on macOS
func (hf *HardwareFingerprinter) getCPUModel() (string, error) {
	// Method 1: Try system_profiler
	if model, err := hf.getSystemProfilerCPUModel(); err == nil && model != "" {
		return model, nil
	}

	// Method 2: Try sysctl
	if model, err := hf.getSysctlCPUModel(); err == nil && model != "" {
		return model, nil
	}

	return "", fmt.Errorf("failed to get CPU model from any source")
}

// getSystemProfilerCPUModel uses system_profiler to get CPU model
func (hf *HardwareFingerprinter) getSystemProfilerCPUModel() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler command failed: %w", err)
	}

	// Parse output for processor info
	re := regexp.MustCompile(`Processor Name:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		model := strings.TrimSpace(matches[1])
		if model != "" {
			slog.Debug("Retrieved CPU model via system_profiler", slog.String("model", model))
			return model, nil
		}
	}

	return "", fmt.Errorf("no processor name found in system_profiler output")
}

// getSysctlCPUModel uses sysctl to get CPU model
func (hf *HardwareFingerprinter) getSysctlCPUModel() (string, error) {
	// Try different sysctl parameters
	sysctlParams := []string{
		"machdep.cpu.brand_string",
		"hw.machine",
		"hw.model",
	}

	for _, param := range sysctlParams {
		cmd := exec.Command("sysctl", "-n", param)
		output, err := cmd.Output()
		if err == nil {
			model := strings.TrimSpace(string(output))
			if model != "" && !strings.Contains(strings.ToLower(model), "unknown") {
				slog.Debug("Retrieved CPU model via sysctl",
					slog.String("param", param),
					slog.String("model", model),
				)
				return model, nil
			}
		}
	}

	return "", fmt.Errorf("no CPU model found via sysctl")
}

// getBIOSVersion retrieves BIOS/firmware version on macOS
func (hf *HardwareFingerprinter) getBIOSVersion() (string, error) {
	// Method 1: Try system_profiler for boot ROM version
	if version, err := hf.getSystemProfilerBootROM(); err == nil && version != "" {
		return version, nil
	}

	// Method 2: Try system_profiler for secure virtual memory info
	if version, err := hf.getSystemProfilerSecureVM(); err == nil && version != "" {
		return version, nil
	}

	return "", fmt.Errorf("failed to get BIOS/firmware version from any source")
}

// getSystemProfilerBootROM uses system_profiler to get boot ROM version
func (hf *HardwareFingerprinter) getSystemProfilerBootROM() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler command failed: %w", err)
	}

	// Parse output for Boot ROM Version
	re := regexp.MustCompile(`Boot ROM Version:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		version := strings.TrimSpace(matches[1])
		if version != "" {
			slog.Debug("Retrieved Boot ROM version via system_profiler", slog.String("version", version))
			return version, nil
		}
	}

	return "", fmt.Errorf("no Boot ROM version found in system_profiler output")
}

// getSystemProfilerSecureVM uses system_profiler for system firmware info
func (hf *HardwareFingerprinter) getSystemProfilerSecureVM() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler command failed: %w", err)
	}

	// Look for System Firmware Version or similar
	re := regexp.MustCompile(`System Firmware Version:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		version := strings.TrimSpace(matches[1])
		if version != "" {
			slog.Debug("Retrieved system firmware version via system_profiler", slog.String("version", version))
			return version, nil
		}
	}

	// Try to find any version-related info as fallback
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Version:") && (strings.Contains(line, "Boot") ||
			strings.Contains(line, "ROM") || strings.Contains(line, "Firmware")) {
			if matches := regexp.MustCompile(`(.+?)Version:\s*(.+)`).FindStringSubmatch(line); len(matches) >= 3 {
				version := strings.TrimSpace(matches[2])
				if version != "" {
					slog.Debug("Retrieved firmware version (fallback) via system_profiler", slog.String("version", version))
					return version, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no firmware version found in system_profiler output")
}