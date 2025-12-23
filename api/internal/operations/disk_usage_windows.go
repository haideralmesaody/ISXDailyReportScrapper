//go:build windows

package operations

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func queryDiskUsage(path string) (*DiskSpaceInfo, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path for disk check: %w", err)
	}

	var freeBytes, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(ptr, &freeBytes, &totalBytes, &totalFreeBytes)
	if err != nil {
		return nil, fmt.Errorf("disk space query failed: %w", err)
	}

	info := &DiskSpaceInfo{
		TotalGB:     bytesToGB(totalBytes),
		AvailableGB: bytesToGB(freeBytes),
	}
	info.UsedGB = info.TotalGB - info.AvailableGB
	if info.TotalGB > 0 {
		info.UsedPercent = (info.UsedGB / info.TotalGB) * 100
	}

	return info, nil
}
