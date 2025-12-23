package operations

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Stage constants and factory function
// All stage implementations have been moved to separate files for better maintainability

// extractDateFromFileName extracts date from ISX report filename
// Expected format: "2025 08 07 ISX Daily Report.xlsx" or similar
func extractDateFromFileName(fileName string) string {
	// Remove extension and split by spaces
	base := fileName
	if len(base) > 4 && base[len(base)-4:] == ".xlsx" {
		base = base[:len(base)-4]
	} else if len(base) > 4 && base[len(base)-4:] == ".xls" {
		base = base[:len(base)-4]
	}

	parts := strings.Fields(base)
	if len(parts) >= 3 {
		// Try to extract date in YYYY MM DD format
		if year, err := strconv.Atoi(parts[0]); err == nil && year >= 2000 && year <= 2030 {
			if month, err := strconv.Atoi(parts[1]); err == nil && month >= 1 && month <= 12 {
				if day, err := strconv.Atoi(parts[2]); err == nil && day >= 1 && day <= 31 {
					return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
				}
			}
		}
	}

	// If pattern doesn't match, return current date
	return time.Now().Format("2006-01-02")
}

// CreateStages creates operation steps with optional configuration
// Uses ProcessingStage for better progress tracking and file-level monitoring
func CreateStages(executableDir string, logger *slog.Logger, options *StageOptions) map[string]Step {
	return map[string]Step{
		StageIDScraping:   NewScrapingStage(executableDir, logger, options),
		StageIDProcessing: NewProcessingStage(executableDir, logger, options),
		StageIDIndices:    NewIndicesStage(executableDir, logger, options),
		StageIDLiquidity:  NewLiquidityStage(executableDir, logger, options),
	}
}
