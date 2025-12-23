package services

// This file previously contained the deprecated Logger interface and related functions.
// All services now use slog directly via dependency injection.
// 
// For logging, use:
// - infrastructure.GetLogger() for basic logging
// - infrastructure.LoggerWithContext(ctx) for context-aware logging
// - Direct *slog.Logger injection in service constructors