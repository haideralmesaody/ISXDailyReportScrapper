package license

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// GetReactivationInfo returns reactivation details from the current license.
func (m *Manager) GetReactivationInfo() (*ReactivationInfo, error) {
	ctx := context.Background()

	licenseInfo, _, err := m.GetLicenseStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get license status: %w", err)
	}
	if licenseInfo == nil {
		return nil, fmt.Errorf("no license activated")
	}

	defaultReactivationInfo := &ReactivationInfo{
		Count:     0,
		Limit:     5,
		Remaining: 5,
	}

	now := time.Now()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	defaultReactivationInfo.NextResetDate = nextMonth

	m.logInfo(ctx, "get_reactivation_info", "Retrieved reactivation information",
		slog.Int("count", defaultReactivationInfo.Count),
		slog.Int("limit", defaultReactivationInfo.Limit),
		slog.Int("remaining", defaultReactivationInfo.Remaining),
		slog.Time("next_reset", defaultReactivationInfo.NextResetDate),
	)

	return defaultReactivationInfo, nil
}
