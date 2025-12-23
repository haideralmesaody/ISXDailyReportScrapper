// Package config provides centralized configuration management for the ISX system.
// It handles loading configuration from multiple sources, validation, and provides
// a type-safe API for accessing configuration values throughout the application.
//
// # Configuration Sources
//
// Configuration is loaded from the following sources in order of precedence:
//
//	1. Environment variables (highest priority)
//	2. Configuration files (JSON/YAML)
//	3. Default values (lowest priority)
//
// # Environment Variables
//
// All environment variables follow the pattern ISX_* for namespacing:
//
//	ISX_PORT=8080
//	ISX_DATABASE_URL=postgres://...
//	ISX_LICENSE_SERVER=https://license.iraqiinvestor.com
//	ISX_LOG_LEVEL=info
//	ISX_ENABLE_METRICS=true
//
// # Configuration Structure
//
// The main configuration struct:
//
//	type Config struct {
//	    Port           int    `envconfig:"PORT" default:"8080"`
//	    DatabaseURL    string `envconfig:"DATABASE_URL" required:"true"`
//	    LicenseServer  string `envconfig:"LICENSE_SERVER"`
//	    LogLevel       string `envconfig:"LOG_LEVEL" default:"info"`
//	    EnableMetrics  bool   `envconfig:"ENABLE_METRICS" default:"true"`
//	}
//
// # Path Management
//
// The package provides centralized path management through the Paths type,
// which handles all file system paths relative to the executable location:
//
//	paths := config.GetPaths()
//	downloadPath := paths.GetDownloadPath("report.xlsx")
//	reportPath := paths.GetReportPath("summary.csv")
//
// # Validation
//
// All configuration is validated at load time to ensure:
//
//	- Required fields are present
//	- Values are within acceptable ranges
//	- File paths are accessible
//	- URLs are properly formatted
//
// # Usage
//
// Load configuration at application startup:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Testing
//
// For testing, use the config.WithDefaults() function to create
// a configuration with sensible test defaults that don't require
// environment variables or external resources.
package config