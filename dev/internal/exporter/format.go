package exporter

import (
	"fmt"
	"strings"
)

// formatFloat formats a float64 value for CSV output
func formatFloat(f float64) string {
	if f == 0 {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", f), "0"), ".")
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