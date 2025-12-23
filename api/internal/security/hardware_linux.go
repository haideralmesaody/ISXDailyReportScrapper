//go:build linux

package security

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// getMotherboardSerial retrieves motherboard serial number on Linux
func (hf *HardwareFingerprinter) getMotherboardSerial() (string, error) {
	// Method 1: Try DMI board_serial
	if serial, err := hf.getDMIBoardSerial(); err == nil && serial != "" {
		return serial, nil
	}

	// Method 2: Try DMI product_uuid as fallback
	if uuid, err := hf.getDMIProductUUID(); err == nil && uuid != "" {
		slog.Debug("Using product UUID as motherboard identifier", slog.String("uuid", uuid))
		return uuid, nil
	}

	// Method 3: Try machine-id from /etc
	if machineID, err := hf.getMachineID(); err == nil && machineID != "" {
		slog.Debug("Using machine-id as motherboard identifier", slog.String("machine_id", machineID))
		return machineID, nil
	}

	return "", fmt.Errorf("failed to get motherboard serial from any source")
}

// getDMIBoardSerial reads from DMI/SMBIOS data
func (hf *HardwareFingerprinter) getDMIBoardSerial() (string, error) {
	// Try different DMI paths
	dmiPaths := []string{
		"/sys/class/dmi/id/board_serial",
		"/sys/class/dmi/id/board_asset_tag",
		"/sys/class/dmi/id/chassis_serial",
	}

	for _, path := range dmiPaths {
		if serial, err := os.ReadFile(path); err == nil {
			serialStr := strings.TrimSpace(string(serial))
			if serialStr != "" && !strings.Contains(strings.ToLower(serialStr), "none") &&
				!strings.Contains(strings.ToLower(serialStr), "unknown") &&
				!strings.Contains(strings.ToLower(serialStr), "to be filled") {
				slog.Debug("Retrieved motherboard serial from DMI",
					slog.String("path", path),
					slog.String("serial", serialStr),
				)
				return serialStr, nil
			}
		}
	}

	return "", fmt.Errorf("no valid board serial found in DMI data")
}

// getDMIProductUUID reads product UUID from DMI
func (hf *HardwareFingerprinter) getDMIProductUUID() (string, error) {
	if data, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		uuid := strings.TrimSpace(string(data))
		if uuid != "" && len(uuid) >= 32 {
			return strings.ToLower(uuid), nil
		}
	}

	return "", fmt.Errorf("no valid product UUID found in DMI data")
}

// getMachineID reads machine-id from system files
func (hf *HardwareFingerprinter) getMachineID() (string, error) {
	// Try different machine-id locations
	machineIDPaths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
		"/etc/hostid",
	}

	for _, path := range machineIDPaths {
		if data, err := os.ReadFile(path); err == nil {
			machineID := strings.TrimSpace(string(data))
			if machineID != "" {
				return strings.ToLower(machineID), nil
			}
		}
	}

	return "", fmt.Errorf("no valid machine-id found")
}

// getSystemUUID retrieves system UUID on Linux
func (hf *HardwareFingerprinter) getSystemUUID() (string, error) {
	// Method 1: Try DMI product_uuid
	if uuid, err := hf.getDMIProductUUID(); err == nil && uuid != "" {
		return uuid, nil
	}

	// Method 2: Try DMI system_uuid
	if uuid, err := hf.getDMISystemUUID(); err == nil && uuid != "" {
		return uuid, nil
	}

	// Method 3: Try machine-id as fallback
	if machineID, err := hf.getMachineID(); err == nil && machineID != "" {
		return machineID, nil
	}

	return "", fmt.Errorf("failed to get system UUID from any source")
}

// getDMISystemUUID reads system UUID from DMI
func (hf *HardwareFingerprinter) getDMISystemUUID() (string, error) {
	if data, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		uuid := strings.TrimSpace(string(data))
		if uuid != "" && len(uuid) >= 32 {
			return strings.ToLower(uuid), nil
		}
	}

	// Try alternative DMI paths
	dmiPaths := []string{
		"/sys/class/dmi/id/system_uuid",
		"/sys/class/dmi/id/chassis_asset_tag",
	}

	for _, path := range dmiPaths {
		if data, err := os.ReadFile(path); err == nil {
			uuid := strings.TrimSpace(string(data))
			if uuid != "" && len(uuid) >= 32 {
				slog.Debug("Retrieved system UUID from DMI",
					slog.String("path", path),
					slog.String("uuid", uuid),
				)
				return strings.ToLower(uuid), nil
			}
		}
	}

	return "", fmt.Errorf("no valid system UUID found in DMI data")
}

// getPrimaryDiskSerial retrieves primary disk serial number on Linux
func (hf *HardwareFingerprinter) getPrimaryDiskSerial() (string, error) {
	// Method 1: Try root filesystem disk serial
	if serial, err := hf.getRootDiskSerial(); err == nil && serial != "" {
		return serial, nil
	}

	// Method 2: Try first available disk serial
	if serial, err := hf.getFirstDiskSerial(); err == nil && serial != "" {
		return serial, nil
	}

	return "", fmt.Errorf("failed to get disk serial from any source")
}

// getRootDiskSerial gets the serial of the disk containing the root filesystem
func (hf *HardwareFingerprinter) getRootDiskSerial() (string, error) {
	// Find root device
	rootDevice, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return "", fmt.Errorf("failed to read mountinfo: %w", err)
	}

	lines := strings.Split(string(rootDevice), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[4] == "/" {
			// Found root mount, extract device
			device := fields[9]
			if device != "" && strings.HasPrefix(device, "/dev/") {
				// Get serial for this device
				if serial, err := hf.getDiskSerialByPath(device); err == nil && serial != "" {
					slog.Debug("Retrieved root disk serial",
						slog.String("device", device),
						slog.String("serial", serial),
					)
					return serial, nil
				}
			}
		}
	}

	return "", fmt.Errorf("root device not found in mountinfo")
}

// getFirstDiskSerial gets serial of the first available disk
func (hf *HardwareFingerprinter) getFirstDiskSerial() (string, error) {
	// Try to read from block devices
	blockPaths := []string{
		"/sys/block/sda/device/serial",
		"/sys/block/nvme0n1/device/serial",
		"/sys/block/vda/device/serial",
		"/sys/block/hda/device/serial",
	}

	for _, path := range blockPaths {
		if serial, err := os.ReadFile(path); err == nil {
			serialStr := strings.TrimSpace(string(serial))
			if serialStr != "" && !strings.Contains(strings.ToLower(serialStr), "unknown") {
				slog.Debug("Retrieved disk serial from block device",
					slog.String("path", path),
					slog.String("serial", serialStr),
				)
				return strings.ToLower(serialStr), nil
			}
		}
	}

	return "", fmt.Errorf("no valid disk serial found")
}

// getDiskSerialByPath gets serial for a specific disk path
func (hf *HardwareFingerprinter) getDiskSerialByPath(devicePath string) (string, error) {
	// Convert /dev/sda to sda, /dev/nvme0n1 to nvme0n1, etc.
	deviceName := strings.TrimPrefix(devicePath, "/dev/")

	// Try different serial paths
	serialPaths := []string{
		fmt.Sprintf("/sys/block/%s/device/serial", deviceName),
		fmt.Sprintf("/sys/block/%s/serial", deviceName),
	}

	// Handle NVMe devices differently
	if strings.HasPrefix(deviceName, "nvme") {
		serialPaths = []string{
			fmt.Sprintf("/sys/block/%s/device/serial", deviceName),
		}
	}

	for _, path := range serialPaths {
		if serial, err := os.ReadFile(path); err == nil {
			serialStr := strings.TrimSpace(string(serial))
			if serialStr != "" && !strings.Contains(strings.ToLower(serialStr), "unknown") {
				return strings.ToLower(serialStr), nil
			}
		}
	}

	return "", fmt.Errorf("serial not found for device %s", devicePath)
}

// getCPUModel retrieves CPU model information on Linux
func (hf *HardwareFingerprinter) getCPUModel() (string, error) {
	// Method 1: Try /proc/cpuinfo
	if model, err := hf.getCPUModelFromProc(); err == nil && model != "" {
		return model, nil
	}

	// Method 2: Try DMI processor information
	if model, err := hf.getDMIProcessorInfo(); err == nil && model != "" {
		return model, nil
	}

	return "", fmt.Errorf("failed to get CPU model from any source")
}

// getCPUModelFromProc reads CPU model from /proc/cpuinfo
func (hf *HardwareFingerprinter) getCPUModelFromProc() (string, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				model := strings.TrimSpace(parts[1])
				if model != "" {
					slog.Debug("Retrieved CPU model from /proc/cpuinfo", slog.String("model", model))
					return model, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no CPU model found in /proc/cpuinfo")
}

// getDMIProcessorInfo reads processor info from DMI
func (hf *HardwareFingerprinter) getDMIProcessorInfo() (string, error) {
	// Try DMI processor version
	dmiPaths := []string{
		"/sys/class/dmi/id/processor_version",
		"/sys/class/dmi/id/processor_family",
	}

	for _, path := range dmiPaths {
		if data, err := os.ReadFile(path); err == nil {
			info := strings.TrimSpace(string(data))
			if info != "" && !strings.Contains(strings.ToLower(info), "unknown") {
				slog.Debug("Retrieved processor info from DMI",
					slog.String("path", path),
					slog.String("info", info),
				)
				return info, nil
			}
		}
	}

	return "", fmt.Errorf("no processor info found in DMI data")
}

// getBIOSVersion retrieves BIOS version on Linux
func (hf *HardwareFingerprinter) getBIOSVersion() (string, error) {
	// Method 1: Try DMI bios_version
	if version, err := hf.getDMIBIOSVersion(); err == nil && version != "" {
		return version, nil
	}

	// Method 2: Try DMI bios_date as fallback
	if date, err := hf.getDMIBIOSDate(); err == nil && date != "" {
		slog.Debug("Using BIOS date as version identifier", slog.String("date", date))
		return date, nil
	}

	return "", fmt.Errorf("failed to get BIOS version from any source")
}

// getDMIBIOSVersion reads BIOS version from DMI
func (hf *HardwareFingerprinter) getDMIBIOSVersion() (string, error) {
	biosPaths := []string{
		"/sys/class/dmi-id/bios_version",
		"/sys/class/dmi/id/bios_release",
	}

	for _, path := range biosPaths {
		if data, err := os.ReadFile(path); err == nil {
			version := strings.TrimSpace(string(data))
			if version != "" && !strings.Contains(strings.ToLower(version), "unknown") {
				slog.Debug("Retrieved BIOS version from DMI",
					slog.String("path", path),
					slog.String("version", version),
				)
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("no BIOS version found in DMI data")
}

// getDMIBIOSDate reads BIOS date from DMI
func (hf *HardwareFingerprinter) getDMIBIOSDate() (string, error) {
	if data, err := os.ReadFile("/sys/class/dmi-id/bios_date"); err == nil {
		date := strings.TrimSpace(string(data))
		if date != "" {
			slog.Debug("Retrieved BIOS date from DMI", slog.String("date", date))
			return date, nil
		}
	}

	return "", fmt.Errorf("no BIOS date found in DMI data")
}