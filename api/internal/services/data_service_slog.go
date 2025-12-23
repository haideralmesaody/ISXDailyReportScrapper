package services

import (
	"context"
	"log/slog"

	"github.com/isxcli/isxcli/internal/infrastructure"
)

// Helper functions for data service logging using centralized infrastructure logger

// logDataError logs an error in data service operations
func logDataError(ctx context.Context, action, message string, attrs ...slog.Attr) {
	logger := infrastructure.LoggerWithContext(ctx)
	
	// Add standard attributes
	allAttrs := []slog.Attr{
		slog.String("component", "data_service"),
		slog.String("action", action),
	}
	
	allAttrs = append(allAttrs, attrs...)
	
	logger.LogAttrs(ctx, slog.LevelError, message, allAttrs...)
}