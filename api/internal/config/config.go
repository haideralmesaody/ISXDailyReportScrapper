package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server" envconfig:"SERVER"`
	Security SecurityConfig `yaml:"security" envconfig:"SECURITY"`
	Logging  LoggingConfig  `yaml:"logging" envconfig:"LOGGING"`
	Paths    PathsConfig    `yaml:"paths" envconfig:"PATHS"`
	WebSocket WebSocketConfig `yaml:"websocket" envconfig:"WEBSOCKET"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port             int           `yaml:"port" envconfig:"PORT" default:"8080"`
	ReadTimeout      time.Duration `yaml:"read_timeout" envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout     time.Duration `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT" default:"15s"`
	IdleTimeout      time.Duration `yaml:"idle_timeout" envconfig:"IDLE_TIMEOUT" default:"60s"`
	MaxHeaderBytes   int           `yaml:"max_header_bytes" envconfig:"MAX_HEADER_BYTES" default:"1048576"`
	ShutdownTimeout  time.Duration `yaml:"shutdown_timeout" envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	OperationTimeout time.Duration `yaml:"operation_timeout" envconfig:"OPERATION_TIMEOUT" default:"2h"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" envconfig:"ALLOWED_ORIGINS" default:"http://localhost:8080"`
	EnableCORS     bool     `yaml:"enable_cors" envconfig:"ENABLE_CORS" default:"true"`
	EnableCSRF     bool     `yaml:"enable_csrf" envconfig:"ENABLE_CSRF" default:"false"`
	RateLimit      RateLimitConfig `yaml:"rate_limit" envconfig:"RATE_LIMIT"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled bool    `yaml:"enabled" envconfig:"ENABLED" default:"true"`
	RPS     float64 `yaml:"rps" envconfig:"RPS" default:"100"`
	Burst   int     `yaml:"burst" envconfig:"BURST" default:"50"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level       string `yaml:"level" envconfig:"LEVEL" default:"info"`
	Format      string `yaml:"format" envconfig:"FORMAT" default:"json"`
	Output      string `yaml:"output" envconfig:"OUTPUT" default:"console"`
	FilePath    string `yaml:"file_path" envconfig:"FILE_PATH" default:"logs/app.log"`
	Development bool   `yaml:"development" envconfig:"DEVELOPMENT" default:"true"`
}

// PathsConfig contains file system paths configuration
type PathsConfig struct {
	ExecutableDir string `yaml:"executable_dir" envconfig:"EXECUTABLE_DIR"`
	LicenseFile   string `yaml:"license_file" envconfig:"LICENSE_FILE" default:"license.dat"`
	DataDir       string `yaml:"data_dir" envconfig:"DATA_DIR" default:"data"`
	WebDir        string `yaml:"web_dir" envconfig:"WEB_DIR" default:"web"`
	LogsDir       string `yaml:"logs_dir" envconfig:"LOGS_DIR" default:"logs"`
}

// WebSocketConfig contains WebSocket configuration
type WebSocketConfig struct {
	ReadBufferSize  int           `yaml:"read_buffer_size" envconfig:"READ_BUFFER_SIZE" default:"1024"`
	WriteBufferSize int           `yaml:"write_buffer_size" envconfig:"WRITE_BUFFER_SIZE" default:"1024"`
	PingPeriod      time.Duration `yaml:"ping_period" envconfig:"PING_PERIOD" default:"30s"`
	PongWait        time.Duration `yaml:"pong_wait" envconfig:"PONG_WAIT" default:"60s"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	var cfg Config

	// Load from environment variables first
	if err := envconfig.Process("ISX", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from env: %w", err)
	}

	// Load from config file if exists
	configFile := getConfigFilePath()
	if _, err := os.Stat(configFile); err == nil {
		fileConfig, err := loadFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
		cfg = mergeConfigs(*fileConfig, cfg)
	}

	// Resolve relative paths
	if err := cfg.resolvePaths(); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	// Validate paths
	if err := cfg.ValidatePaths(); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	return &cfg, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// mergeConfigs merges file config with env config (env takes precedence)
func mergeConfigs(fileConfig, envConfig Config) Config {
	// Server config
	if envConfig.Server.Port == 0 {
		envConfig.Server.Port = fileConfig.Server.Port
	}
	if envConfig.Server.ReadTimeout == 0 {
		envConfig.Server.ReadTimeout = fileConfig.Server.ReadTimeout
	}
	// ... continue for other fields

	return envConfig
}

// resolvePaths sets up the executable directory and validates paths
func (c *Config) resolvePaths() error {
	// Use centralized paths system to get all paths
	paths, err := GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Update config with resolved paths from centralized system
	c.Paths.ExecutableDir = paths.ExecutableDir
	
	// Keep the configured relative paths for backward compatibility
	// The Get* methods will use the centralized paths system
	
	return nil
}

// ValidatePaths validates that critical paths exist or can be created
func (c *Config) ValidatePaths() error {
	paths, err := GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}
	
	// Ensure all required directories exist
	if err := paths.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}
	
	// Log path resolution for debugging
	paths.LogPathResolution()
	
	return nil
}

// GetDataDir returns the resolved data directory path
func (c *Config) GetDataDir() string {
	// Use centralized paths system as the single source of truth
	paths, err := GetPaths()
	if err != nil {
		// Fallback to config-based resolution if paths system fails
		if filepath.IsAbs(c.Paths.DataDir) {
			return c.Paths.DataDir
		}
		return filepath.Join(c.Paths.ExecutableDir, c.Paths.DataDir)
	}
	return paths.DataDir
}

// GetWebDir returns the resolved web directory path
func (c *Config) GetWebDir() string {
	// Use centralized paths system as the single source of truth
	paths, err := GetPaths()
	if err != nil {
		// Fallback to config-based resolution if paths system fails
		if filepath.IsAbs(c.Paths.WebDir) {
			return c.Paths.WebDir
		}
		return filepath.Join(c.Paths.ExecutableDir, c.Paths.WebDir)
	}
	return paths.WebDir
}

// GetLogsDir returns the resolved logs directory path
func (c *Config) GetLogsDir() string {
	// Use centralized paths system as the single source of truth
	paths, err := GetPaths()
	if err != nil {
		// Fallback to config-based resolution if paths system fails
		if filepath.IsAbs(c.Paths.LogsDir) {
			return c.Paths.LogsDir
		}
		return filepath.Join(c.Paths.ExecutableDir, c.Paths.LogsDir)
	}
	return paths.LogsDir
}

// GetLicenseFile returns the resolved license file path
func (c *Config) GetLicenseFile() string {
	// Use GetLicensePath as the single source of truth
	// This ensures consistency across the application and handles edge cases
	path, err := GetLicensePath()
	if err != nil {
		// Fallback to standard executable directory resolution
		return filepath.Join(c.Paths.ExecutableDir, c.Paths.LicenseFile)
	}
	
	return path
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be positive")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be positive")
	}

	if len(c.Security.AllowedOrigins) == 0 {
		return fmt.Errorf("at least one allowed origin must be specified")
	}

	// Validate logging configuration per CLAUDE.md
	if c.Logging.Format != "json" {
		// Per CLAUDE.md: Always use JSON format
		c.Logging.Format = "json"
	}

	if c.Logging.Output != "both" && c.Logging.Output != "file" {
		// Per CLAUDE.md: Always dual output (stdout + file)
		c.Logging.Output = "both"
	}

	if c.Logging.FilePath == "" {
		c.Logging.FilePath = "logs/app.log"
	}

	return nil
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() string {
	// Check for config file in common locations
	locations := []string{
		"config.yaml",
		"configs/config.yaml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	return "" // No config file found, use env vars only
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            8080,
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1MB
			ShutdownTimeout: 30 * time.Second,
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{"http://localhost:8080"},
			EnableCORS:     true,
			EnableCSRF:     false,
			RateLimit: RateLimitConfig{
				Enabled: true,
				RPS:     100,
				Burst:   50,
			},
		},
		Logging: LoggingConfig{
			Level:       "info",
			Format:      "json",
			Output:      "both",
			FilePath:    "logs/app.log",
			Development: true,
		},
		Paths: PathsConfig{
			LicenseFile: "license.dat",
			DataDir:     "data",
			WebDir:      "web",
			LogsDir:     "logs",
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			PingPeriod:      30 * time.Second,
			PongWait:        60 * time.Second,
		},
	}
}