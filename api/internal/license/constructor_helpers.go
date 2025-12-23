package license

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/security"
)

func logInitDiagnostics(logger *slog.Logger, opts managerInitOptions) {
	if logger == nil {
		return
	}

	absPath, _ := filepath.Abs(opts.licenseFile)
	workingDir, _ := os.Getwd()
	execPath, _ := os.Executable()

	diagnosticAttrs := []any{
		slog.String("provided_path", opts.licenseFile),
		slog.String("absolute_path", absPath),
		slog.String("working_directory", workingDir),
		slog.String("executable_path", execPath),
		slog.String("path_source", opts.pathSource),
		slog.Bool("file_exists", config.FileExists(opts.licenseFile)),
		slog.Bool("file_readable", checkFileReadable(opts.licenseFile)),
		slog.String("file_size", getFileSizeString(opts.licenseFile)),
		slog.Time("initialization_time", time.Now()),
	}
	if opts.cacheInvalidator != nil {
		diagnosticAttrs = append(diagnosticAttrs, slog.Bool("has_cache_invalidator", true))
	}
	if opts.pathSource == "config_system" {
		diagnosticAttrs = append(diagnosticAttrs, slog.String("config_method", "config.GetLicensePath()"))
	}

	logger.Info("License manager initialization - Comprehensive Diagnostics", diagnosticAttrs...)
}

// resolveLicensePath fills defaults and validates the license path.
func resolveLicensePath(opts managerInitOptions, logger *slog.Logger) (managerInitOptions, error) {
	if strings.TrimSpace(opts.licenseFile) == "" {
		defaultPath, err := config.GetLicensePath()
		if err != nil {
			return opts, fmt.Errorf("failed to resolve default license path: %v", err)
		}
		opts.licenseFile = defaultPath
		if opts.pathSource == "" {
			opts.pathSource = "config_system"
		}
	}

	if err := validateLicensePath(opts.licenseFile); err != nil {
		if logger != nil {
			logger.Error("License manager initialization failed - Invalid path",
				slog.String("provided_path", opts.licenseFile),
				slog.String("path_source", opts.pathSource),
				slog.String("error", err.Error()),
			)
		}
		return opts, fmt.Errorf("invalid license path: %v", err)
	}

	return opts, nil
}

// initManagerComponents initializes core manager dependencies.
func initManagerComponents(opts managerInitOptions) (*Manager, error) {
	sheetsConfig := getBuiltInConfig()
	cache := NewLicenseCache(5*time.Minute, 1000)
	securityMgr := NewSecurityManager(5, 15*time.Minute, 5*time.Minute)

	credentialsManager, err := security.NewSecureCredentialsManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secure credentials manager: %v", err)
	}

	return &Manager{
		config:                   sheetsConfig,
		licenseFile:              opts.licenseFile,
		cache:                    cache,
		security:                 securityMgr,
		performanceTracker:       NewPerformanceTracker(),
		validationCache:          NewValidationCache(),
		credentialsManager:       credentialsManager,
		secureMode:               true,
		fingerprintManager:       security.NewFingerprintManager(),
		hardwareFingerprinter:    security.NewHardwareFingerprinter(),
		hybridFingerprintManager: security.NewHybridFingerprintManager(true),
		licenseValidator:         opts.cacheInvalidator,
	}, nil
}
