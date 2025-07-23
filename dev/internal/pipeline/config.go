package pipeline

import (
	"time"
)

// Config represents the pipeline configuration
type Config struct {
	// Execution mode (sequential or parallel)
	ExecutionMode ExecutionMode `json:"execution_mode"`

	// Stage-specific timeouts
	StageTimeouts map[string]time.Duration `json:"stage_timeouts"`

	// Retry configuration for stages
	RetryConfig RetryConfig `json:"retry_config"`

	// Whether to continue on stage failures
	ContinueOnError bool `json:"continue_on_error"`

	// Maximum concurrent stages (for parallel execution)
	MaxConcurrency int `json:"max_concurrency"`

	// Whether to enable checkpointing
	EnableCheckpoints bool `json:"enable_checkpoints"`

	// Checkpoint directory
	CheckpointDir string `json:"checkpoint_dir"`

	// Custom stage configurations
	StageConfigs map[string]interface{} `json:"stage_configs"`
}

// DefaultConfig returns the default pipeline configuration
func DefaultConfig() *Config {
	return &Config{
		ExecutionMode: ExecutionModeSequential,
		StageTimeouts: map[string]time.Duration{
			StageIDScraping:  DefaultScrapingTimeout,
			StageIDProcessing: DefaultProcessingTimeout,
			StageIDIndices:   DefaultIndicesTimeout,
			StageIDAnalysis:  DefaultAnalysisTimeout,
		},
		RetryConfig:       DefaultRetryConfig(),
		ContinueOnError:   false,
		MaxConcurrency:    1,
		EnableCheckpoints: false,
		CheckpointDir:     "data/checkpoints",
		StageConfigs:      make(map[string]interface{}),
	}
}

// GetStageTimeout returns the timeout for a specific stage
func (c *Config) GetStageTimeout(stageID string) time.Duration {
	if timeout, ok := c.StageTimeouts[stageID]; ok {
		return timeout
	}
	return DefaultStageTimeout
}

// SetStageTimeout sets the timeout for a specific stage
func (c *Config) SetStageTimeout(stageID string, timeout time.Duration) {
	if c.StageTimeouts == nil {
		c.StageTimeouts = make(map[string]time.Duration)
	}
	c.StageTimeouts[stageID] = timeout
}

// GetStageConfig returns the configuration for a specific stage
func (c *Config) GetStageConfig(stageID string) (interface{}, bool) {
	if c.StageConfigs == nil {
		return nil, false
	}
	config, ok := c.StageConfigs[stageID]
	return config, ok
}

// SetStageConfig sets the configuration for a specific stage
func (c *Config) SetStageConfig(stageID string, config interface{}) {
	if c.StageConfigs == nil {
		c.StageConfigs = make(map[string]interface{})
	}
	c.StageConfigs[stageID] = config
}

// StageConfig represents configuration for individual stages
type StageConfig struct {
	// Whether this stage is enabled
	Enabled bool `json:"enabled"`

	// Whether to skip this stage on failure
	SkipOnFailure bool `json:"skip_on_failure"`

	// Custom timeout for this stage
	Timeout time.Duration `json:"timeout"`

	// Retry configuration override
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// Stage-specific parameters
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ScrapingStageConfig represents configuration for the scraping stage
type ScrapingStageConfig struct {
	StageConfig
	Mode     string `json:"mode"`     // initial or accumulative
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	OutDir   string `json:"out_dir"`
}

// ProcessingStageConfig represents configuration for the processing stage
type ProcessingStageConfig struct {
	StageConfig
	InDir      string `json:"in_dir"`
	OutDir     string `json:"out_dir"`
	FullRework bool   `json:"full_rework"`
}

// IndicesStageConfig represents configuration for the indices extraction stage
type IndicesStageConfig struct {
	StageConfig
	InputDir   string `json:"input_dir"`
	OutputFile string `json:"output_file"`
}

// AnalysisStageConfig represents configuration for the analysis stage
type AnalysisStageConfig struct {
	StageConfig
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
}

// Builder provides a fluent interface for building pipeline configurations
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// WithExecutionMode sets the execution mode
func (b *ConfigBuilder) WithExecutionMode(mode ExecutionMode) *ConfigBuilder {
	b.config.ExecutionMode = mode
	return b
}

// WithStageTimeout sets the timeout for a stage
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

// WithStageConfig sets the configuration for a stage
func (b *ConfigBuilder) WithStageConfig(stageID string, config interface{}) *ConfigBuilder {
	b.config.SetStageConfig(stageID, config)
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() *Config {
	return b.config
}