package exporter

import (
	"fmt"
)

// formatFloat formats a float64 value for CSV output with exactly 2 decimal places
func formatFloat(f float64) string {
	// Always format with exactly 2 decimal places for consistency
	// This ensures values like 13.4 appear as 13.40 in CSV
	return fmt.Sprintf("%.2f", f)
}

// formatInt formats an int64 value for CSV output
func formatInt(i int64) string {
	return fmt.Sprintf("%d", i)
}

// formatBool formats a boolean value for CSV output
func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}