package operations

import (
	"time"
)

// Config represents the operation execution configuration
type Config struct {
	// Execution mode (sequential or parallel)
	ExecutionMode ExecutionMode `json:"execution_mode"`

	// Step-specific timeouts
	StageTimeouts map[string]time.Duration `json:"stage_timeouts"`

	// Retry configuration for steps
	RetryConfig RetryConfig `json:"retry_config"`

	// Whether to continue on Step failures
	ContinueOnError bool `json:"continue_on_error"`

	// Maximum concurrent steps (for parallel execution)
	MaxConcurrency int `json:"max_concurrency"`

	// Whether to enable checkpointing
	EnableCheckpoints bool `json:"enable_checkpoints"`

	// Checkpoint directory
	CheckpointDir string `json:"checkpoint_dir"`

	// Custom Step configurations
	StepConfigs map[string]interface{} `json:"stage_configs"`
}

// NewConfig returns the default operation configuration
func NewConfig() *Config {
	return &Config{
		ExecutionMode: ExecutionModeSequential,
		StageTimeouts: map[string]time.Duration{
			StageIDScraping:   DefaultScrapingTimeout,
			StageIDProcessing: DefaultProcessingTimeout,
			StageIDIndices:    DefaultIndicesTimeout,
			StageIDLiquidity:  DefaultLiquidityTimeout,
		},
		RetryConfig:       *NewRetryConfig(),
		ContinueOnError:   false,
		MaxConcurrency:    1,
		EnableCheckpoints: false,
		CheckpointDir:     "data/checkpoints",
		StepConfigs:       make(map[string]interface{}),
	}
}

// GetStageTimeout returns the timeout for a specific Step
func (c *Config) GetStageTimeout(stageID string) time.Duration {
	if timeout, ok := c.StageTimeouts[stageID]; ok {
		return timeout
	}
	return DefaultStageTimeout
}

// SetStageTimeout sets the timeout for a specific Step
func (c *Config) SetStageTimeout(stageID string, timeout time.Duration) {
	if c.StageTimeouts == nil {
		c.StageTimeouts = make(map[string]time.Duration)
	}
	c.StageTimeouts[stageID] = timeout
}

// GetStepConfig returns the configuration for a specific Step
func (c *Config) GetStepConfig(stageID string) (interface{}, bool) {
	if c.StepConfigs == nil {
		return nil, false
	}
	config, ok := c.StepConfigs[stageID]
	return config, ok
}

// SetStepConfig sets the configuration for a specific Step
func (c *Config) SetStepConfig(stageID string, config interface{}) {
	if c.StepConfigs == nil {
		c.StepConfigs = make(map[string]interface{})
	}
	c.StepConfigs[stageID] = config
}

// StepConfig represents configuration for individual steps
type StepConfig struct {
	// Step identification
	ID string `json:"id"`

	// Step type
	Type string `json:"type"`

	// Step dependencies
	Dependencies []string `json:"dependencies,omitempty"`

	// Number of retries
	Retries int `json:"retries,omitempty"`

	// Whether this Step is enabled
	Enabled bool `json:"enabled"`

	// Whether to skip this Step on failure
	SkipOnFailure bool `json:"skip_on_failure"`

	// Custom timeout for this Step
	Timeout time.Duration `json:"timeout"`

	// Retry configuration override
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// Step-specific parameters
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ScrapingConfig represents configuration for scraping operations
type ScrapingConfig struct {
	BaseURL     string        `json:"base_url"`
	DownloadDir string        `json:"download_dir"`
	Timeout     time.Duration `json:"timeout"`
	UserAgent   string        `json:"user_agent"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
}

// ScrapingStepConfig represents configuration for the scraping Step
type ScrapingStepConfig struct {
	StepConfig
	Mode     string `json:"mode"` // initial or accumulative
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	OutDir   string `json:"out_dir"`
}

// ProcessingStepConfig represents configuration for the processing Step
type ProcessingStepConfig struct {
	StepConfig
	InDir      string `json:"in_dir"`
	OutDir     string `json:"out_dir"`
	FullRework bool   `json:"full_rework"`
}

// IndicesStepConfig represents configuration for the indices extraction Step
type IndicesStepConfig struct {
	StepConfig
	InputDir   string `json:"input_dir"`
	OutputFile string `json:"output_file"`
}

// LiquidityStepConfig represents configuration for the liquidity calculation step
type LiquidityStepConfig struct {
	StepConfig
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
}

// Builder provides a fluent interface for building operation configurations
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// WithExecutionMode sets the execution mode
func (b *ConfigBuilder) WithExecutionMode(mode ExecutionMode) *ConfigBuilder {
	b.config.ExecutionMode = mode
	return b
}

// WithStageTimeout sets the timeout for a Step
func (b *ConfigBuilder) WithStageTimeout(stageID string, timeout time.Duration) *ConfigBuilder {
	b.config.SetStageTimeout(stageID, timeout)
	return b
}

// WithRetryConfig sets the retry configuration
func (b *ConfigBuilder) WithRetryConfig(config RetryConfig) *ConfigBuilder {
	b.config.RetryConfig = config
	return b
}

// WithContinueOnError sets whether to continue on errors
func (b *ConfigBuilder) WithContinueOnError(continueOnError bool) *ConfigBuilder {
	b.config.ContinueOnError = continueOnError
	return b
}

// WithMaxConcurrency sets the maximum concurrency
func (b *ConfigBuilder) WithMaxConcurrency(maxConcurrency int) *ConfigBuilder {
	b.config.MaxConcurrency = maxConcurrency
	return b
}

// WithCheckpoints enables checkpointing
func (b *ConfigBuilder) WithCheckpoints(enabled bool, dir string) *ConfigBuilder {
	b.config.EnableCheckpoints = enabled
	if dir != "" {
		b.config.CheckpointDir = dir
	}
	return b
}

// WithStepConfig sets the configuration for a Step
func (b *ConfigBuilder) WithStepConfig(stageID string, config interface{}) *ConfigBuilder {
	b.config.SetStepConfig(stageID, config)
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() *Config {
	return b.config
}
