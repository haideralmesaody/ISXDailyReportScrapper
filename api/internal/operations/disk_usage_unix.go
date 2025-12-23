//go:build !windows

package operations

import (
	"fmt"
	"syscall"
)

func queryDiskUsage(path string) (*DiskSpaceInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("disk space query failed: %w", err)
	}

	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)

	info := &DiskSpaceInfo{
		TotalGB:     bytesToGB(total),
		AvailableGB: bytesToGB(free),
	}
	info.UsedGB = info.TotalGB - info.AvailableGB
	if info.TotalGB > 0 {
		info.UsedPercent = (info.UsedGB / info.TotalGB) * 100
	}

	return info, nil
}
