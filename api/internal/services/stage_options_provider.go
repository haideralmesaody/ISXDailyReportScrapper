package services

import (
	"log/slog"

	"github.com/isxcli/isxcli/internal/operations"
)

// DefaultStageOptionsProvider provides dependencies for stage construction.
type DefaultStageOptionsProvider struct {
	executableDir     string
	logger            *slog.Logger
	stageOptions      *operations.StageOptions
	config            *operations.Config
	licenseChecker    operations.LicenseChecker
	webSocketManager  operations.WebSocketHub
	statusBroadcaster *operations.StatusBroadcaster
}

// NewDefaultStageOptionsProvider creates a new stage options provider.
func NewDefaultStageOptionsProvider(
	executableDir string,
	logger *slog.Logger,
	stageOptions *operations.StageOptions,
	licenseChecker operations.LicenseChecker,
	webSocketManager operations.WebSocketHub,
	statusBroadcaster *operations.StatusBroadcaster,
) *DefaultStageOptionsProvider {
	return &DefaultStageOptionsProvider{
		executableDir:     executableDir,
		logger:            logger,
		stageOptions:      stageOptions,
		config:            &operations.Config{},
		licenseChecker:    licenseChecker,
		webSocketManager:  webSocketManager,
		statusBroadcaster: statusBroadcaster,
	}
}

func (p *DefaultStageOptionsProvider) GetExecutableDir() string {
	return p.executableDir
}

func (p *DefaultStageOptionsProvider) GetLogger() *slog.Logger {
	return p.logger
}

func (p *DefaultStageOptionsProvider) GetStageOptions() *operations.StageOptions {
	return p.stageOptions
}

func (p *DefaultStageOptionsProvider) GetConfig() *operations.Config {
	return p.config
}

func (p *DefaultStageOptionsProvider) GetLicenseChecker() operations.LicenseChecker {
	return p.licenseChecker
}

func (p *DefaultStageOptionsProvider) GetWebSocketManager() operations.WebSocketHub {
	return p.webSocketManager
}

func (p *DefaultStageOptionsProvider) GetStatusBroadcaster() *operations.StatusBroadcaster {
	return p.statusBroadcaster
}
