package operations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfigDefaults tests that NewConfig returns proper defaults
func TestConfigDefaults(t *testing.T) {
	config := NewConfig()

	assert.NotNil(t, config)
	assert.Equal(t, ExecutionModeSequential, config.ExecutionMode)
	assert.Equal(t, false, config.ContinueOnError)
	assert.Equal(t, 1, config.MaxConcurrency)
	assert.Equal(t, false, config.EnableCheckpoints)
	assert.Equal(t, "data/checkpoints", config.CheckpointDir)

	// Test stage timeouts
	assert.Equal(t, DefaultScrapingTimeout, config.StageTimeouts[StageIDScraping])
	assert.Equal(t, DefaultProcessingTimeout, config.StageTimeouts[StageIDProcessing])
	assert.Equal(t, DefaultIndicesTimeout, config.StageTimeouts[StageIDIndices])
	assert.Equal(t, DefaultLiquidityTimeout, config.StageTimeouts[StageIDLiquidity])

	// Test retry config defaults
	assert.Equal(t, 3, config.RetryConfig.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.RetryConfig.InitialDelay)
	assert.Equal(t, 30*time.Second, config.RetryConfig.MaxDelay)
	assert.Equal(t, 2.0, config.RetryConfig.Multiplier)

	// Test maps are initialized
	assert.NotNil(t, config.StageTimeouts)
	assert.NotNil(t, config.StepConfigs)
}

// TestGetStageTimeout tests stage timeout retrieval
func TestGetStageTimeout(t *testing.T) {
	config := NewConfig()

	tests := []struct {
		name            string
		stageID         string
		expectedTimeout time.Duration
	}{
		{
			name:            "existing scraping stage timeout",
			stageID:         StageIDScraping,
			expectedTimeout: DefaultScrapingTimeout,
		},
		{
			name:            "existing processing stage timeout",
			stageID:         StageIDProcessing,
			expectedTimeout: DefaultProcessingTimeout,
		},
		{
			name:            "existing indices stage timeout",
			stageID:         StageIDIndices,
			expectedTimeout: DefaultIndicesTimeout,
		},
		{
			name:            "existing analysis stage timeout",
			stageID:         StageIDLiquidity,
			expectedTimeout: DefaultLiquidityTimeout,
		},
		{
			name:            "non-existing stage returns default",
			stageID:         "non-existing-stage",
			expectedTimeout: DefaultStageTimeout,
		},
		{
			name:            "empty stage ID returns default",
			stageID:         "",
			expectedTimeout: DefaultStageTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := config.GetStageTimeout(tt.stageID)
			assert.Equal(t, tt.expectedTimeout, timeout)
		})
	}
}

// TestSetStageTimeout tests stage timeout setting
func TestSetStageTimeout(t *testing.T) {
	tests := []struct {
		name            string
		setupConfig     func() *Config
		stageID         string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name: "set timeout on existing stage",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID:         StageIDScraping,
			timeout:         2 * time.Minute,
			expectedTimeout: 2 * time.Minute,
		},
		{
			name: "set timeout on new stage",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID:         "new-stage",
			timeout:         5 * time.Minute,
			expectedTimeout: 5 * time.Minute,
		},
		{
			name: "set timeout with nil map",
			setupConfig: func() *Config {
				config := NewConfig()
				config.StageTimeouts = nil
				return config
			},
			stageID:         "test-stage",
			timeout:         3 * time.Minute,
			expectedTimeout: 3 * time.Minute,
		},
		{
			name: "set zero timeout",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID:         "zero-stage",
			timeout:         0,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			config.SetStageTimeout(tt.stageID, tt.timeout)

			assert.NotNil(t, config.StageTimeouts)
			assert.Equal(t, tt.expectedTimeout, config.GetStageTimeout(tt.stageID))
		})
	}
}

// TestGetStepConfig tests step configuration retrieval
func TestGetStepConfig(t *testing.T) {
	config := NewConfig()

	t.Run("get non-existing config", func(t *testing.T) {
		value, exists := config.GetStepConfig("non-existing")
		assert.Nil(t, value)
		assert.False(t, exists)
	})

	t.Run("get config with nil map", func(t *testing.T) {
		config.StepConfigs = nil
		value, exists := config.GetStepConfig("any-stage")
		assert.Nil(t, value)
		assert.False(t, exists)
	})

	t.Run("get existing config", func(t *testing.T) {
		config = NewConfig()
		expectedConfig := map[string]interface{}{
			"key": "value",
			"num": 42,
		}
		config.StepConfigs["test-stage"] = expectedConfig

		value, exists := config.GetStepConfig("test-stage")
		assert.True(t, exists)
		assert.Equal(t, expectedConfig, value)
	})
}

// TestSetStepConfig tests step configuration setting
func TestSetStepConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *Config
		stageID     string
		config      interface{}
	}{
		{
			name: "set config on initialized map",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID: "test-stage",
			config: map[string]interface{}{
				"setting1": "value1",
				"setting2": 123,
			},
		},
		{
			name: "set config with nil map",
			setupConfig: func() *Config {
				config := NewConfig()
				config.StepConfigs = nil
				return config
			},
			stageID: "test-stage",
			config:  "simple string config",
		},
		{
			name: "set nil config",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID: "nil-stage",
			config:  nil,
		},
		{
			name: "set complex config",
			setupConfig: func() *Config {
				return NewConfig()
			},
			stageID: "complex-stage",
			config: struct {
				Name     string
				Value    int
				Enabled  bool
				Settings map[string]string
			}{
				Name:    "complex",
				Value:   42,
				Enabled: true,
				Settings: map[string]string{
					"debug": "true",
					"level": "info",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			config.SetStepConfig(tt.stageID, tt.config)

			assert.NotNil(t, config.StepConfigs)

			value, exists := config.GetStepConfig(tt.stageID)
			assert.True(t, exists)
			assert.Equal(t, tt.config, value)
		})
	}
}

// TestConfigBuilder tests the fluent configuration builder
func TestConfigBuilder(t *testing.T) {
	t.Run("build default config", func(t *testing.T) {
		builder := NewConfigBuilder()
		assert.NotNil(t, builder)
		assert.NotNil(t, builder.config)

		config := builder.Build()
		assert.NotNil(t, config)

		// Should have same defaults as NewConfig()
		defaultConfig := NewConfig()
		assert.Equal(t, defaultConfig.ExecutionMode, config.ExecutionMode)
		assert.Equal(t, defaultConfig.ContinueOnError, config.ContinueOnError)
		assert.Equal(t, defaultConfig.MaxConcurrency, config.MaxConcurrency)
	})

	t.Run("build with execution mode", func(t *testing.T) {
		config := NewConfigBuilder().
			WithExecutionMode(ExecutionModeParallel).
			Build()

		assert.Equal(t, ExecutionModeParallel, config.ExecutionMode)
	})

	t.Run("build with stage timeout", func(t *testing.T) {
		customTimeout := 45 * time.Minute
		config := NewConfigBuilder().
			WithStageTimeout("custom-stage", customTimeout).
			Build()

		assert.Equal(t, customTimeout, config.GetStageTimeout("custom-stage"))
	})

	t.Run("build with retry config", func(t *testing.T) {
		customRetry := RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 2 * time.Second,
			MaxDelay:     60 * time.Second,
			Multiplier:   1.5,
		}

		config := NewConfigBuilder().
			WithRetryConfig(customRetry).
			Build()

		assert.Equal(t, customRetry, config.RetryConfig)
	})

	t.Run("build with continue on error", func(t *testing.T) {
		config := NewConfigBuilder().
			WithContinueOnError(true).
			Build()

		assert.True(t, config.ContinueOnError)
	})

	t.Run("build with max concurrency", func(t *testing.T) {
		config := NewConfigBuilder().
			WithMaxConcurrency(8).
			Build()

		assert.Equal(t, 8, config.MaxConcurrency)
	})

	t.Run("build with checkpoints", func(t *testing.T) {
		config := NewConfigBuilder().
			WithCheckpoints(true, "/custom/checkpoint/dir").
			Build()

		assert.True(t, config.EnableCheckpoints)
		assert.Equal(t, "/custom/checkpoint/dir", config.CheckpointDir)
	})

	t.Run("build with checkpoints enabled but no directory", func(t *testing.T) {
		config := NewConfigBuilder().
			WithCheckpoints(true, "").
			Build()

		assert.True(t, config.EnableCheckpoints)
		assert.Equal(t, "data/checkpoints", config.CheckpointDir) // Should keep default
	})

	t.Run("build with step config", func(t *testing.T) {
		stepConfig := map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		}

		config := NewConfigBuilder().
			WithStepConfig("test-stage", stepConfig).
			Build()

		value, exists := config.GetStepConfig("test-stage")
		assert.True(t, exists)
		assert.Equal(t, stepConfig, value)
	})

	t.Run("build with chained methods", func(t *testing.T) {
		config := NewConfigBuilder().
			WithExecutionMode(ExecutionModeParallel).
			WithMaxConcurrency(4).
			WithContinueOnError(true).
			WithStageTimeout(StageIDScraping, 2*time.Hour).
			WithRetryConfig(RetryConfig{
				MaxAttempts:  10,
				InitialDelay: 500 * time.Millisecond,
				MaxDelay:     2 * time.Minute,
				Multiplier:   1.2,
			}).
			WithCheckpoints(true, "/tmp/checkpoints").
			WithStepConfig("scraping", ScrapingStepConfig{
				Mode:     "full",
				FromDate: "2024-01-01",
				ToDate:   "2024-12-31",
			}).
			Build()

		assert.Equal(t, ExecutionModeParallel, config.ExecutionMode)
		assert.Equal(t, 4, config.MaxConcurrency)
		assert.True(t, config.ContinueOnError)
		assert.Equal(t, 2*time.Hour, config.GetStageTimeout(StageIDScraping))
		assert.Equal(t, 10, config.RetryConfig.MaxAttempts)
		assert.True(t, config.EnableCheckpoints)
		assert.Equal(t, "/tmp/checkpoints", config.CheckpointDir)

		value, exists := config.GetStepConfig("scraping")
		assert.True(t, exists)
		assert.IsType(t, ScrapingStepConfig{}, value)
	})
}

// TestNewRetryConfig tests the default retry configuration
func TestNewRetryConfig(t *testing.T) {
	config := NewRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

// TestStepConfigTypes tests various step configuration types
func TestStepConfigTypes(t *testing.T) {
	t.Run("scraping step config", func(t *testing.T) {
		config := ScrapingStepConfig{
			StepConfig: StepConfig{
				ID:      "scraping",
				Type:    "data_collection",
				Enabled: true,
				Retries: 3,
				Timeout: 60 * time.Minute,
			},
			Mode:     "full",
			FromDate: "2024-01-01",
			ToDate:   "2024-12-31",
			OutDir:   "/data/downloads",
		}

		assert.Equal(t, "scraping", config.ID)
		assert.Equal(t, "data_collection", config.Type)
		assert.True(t, config.Enabled)
		assert.Equal(t, "full", config.Mode)
		assert.Equal(t, "2024-01-01", config.FromDate)
		assert.Equal(t, "2024-12-31", config.ToDate)
		assert.Equal(t, "/data/downloads", config.OutDir)
	})

	t.Run("processing step config", func(t *testing.T) {
		config := ProcessingStepConfig{
			StepConfig: StepConfig{
				ID:            "processing",
				Type:          "data_processing",
				Enabled:       true,
				SkipOnFailure: false,
			},
			InDir:      "/data/downloads",
			OutDir:     "/data/reports",
			FullRework: true,
		}

		assert.Equal(t, "processing", config.ID)
		assert.Equal(t, "data_processing", config.Type)
		assert.Equal(t, "/data/downloads", config.InDir)
		assert.Equal(t, "/data/reports", config.OutDir)
		assert.True(t, config.FullRework)
	})

	t.Run("indices step config", func(t *testing.T) {
		config := IndicesStepConfig{
			StepConfig: StepConfig{
				ID:      "indices",
				Type:    "index_extraction",
				Enabled: true,
			},
			InputDir:   "/data/reports",
			OutputFile: "/data/indices.csv",
		}

		assert.Equal(t, "indices", config.ID)
		assert.Equal(t, "index_extraction", config.Type)
		assert.Equal(t, "/data/reports", config.InputDir)
		assert.Equal(t, "/data/indices.csv", config.OutputFile)
	})

	t.Run("analysis step config", func(t *testing.T) {
		config := LiquidityStepConfig{
			StepConfig: StepConfig{
				ID:      "analysis",
				Type:    "data_analysis",
				Enabled: true,
				Parameters: map[string]interface{}{
					"threshold": 0.05,
					"window":    30,
				},
			},
			InputFile:  "/data/indices.csv",
			OutputFile: "/data/analysis.json",
		}

		assert.Equal(t, "analysis", config.ID)
		assert.Equal(t, "data_analysis", config.Type)
		assert.Equal(t, "/data/indices.csv", config.InputFile)
		assert.Equal(t, "/data/analysis.json", config.OutputFile)
		assert.NotNil(t, config.Parameters)
		assert.Equal(t, 0.05, config.Parameters["threshold"])
		assert.Equal(t, 30, config.Parameters["window"])
	})
}

// TestStepConfigWithRetryOverride tests step-specific retry configuration
func TestStepConfigWithRetryOverride(t *testing.T) {
	customRetry := &RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 2 * time.Second,
		MaxDelay:     60 * time.Second,
		Multiplier:   1.5,
	}

	config := StepConfig{
		ID:          "custom-retry-stage",
		Type:        "test",
		Enabled:     true,
		Retries:     5,
		RetryConfig: customRetry,
	}

	assert.Equal(t, "custom-retry-stage", config.ID)
	assert.Equal(t, 5, config.Retries)
	assert.NotNil(t, config.RetryConfig)
	assert.Equal(t, 5, config.RetryConfig.MaxAttempts)
	assert.Equal(t, 2*time.Second, config.RetryConfig.InitialDelay)
}

// TestConfigEdgeCases tests edge cases and boundary conditions
func TestConfigEdgeCases(t *testing.T) {
	t.Run("negative values", func(t *testing.T) {
		config := NewConfigBuilder().
			WithMaxConcurrency(-1).
			WithRetryConfig(RetryConfig{
				MaxAttempts:  -1,
				InitialDelay: -1 * time.Second,
				MaxDelay:     -1 * time.Second,
				Multiplier:   -1.0,
			}).
			Build()

		// Should accept negative values (validation is elsewhere)
		assert.Equal(t, -1, config.MaxConcurrency)
		assert.Equal(t, -1, config.RetryConfig.MaxAttempts)
		assert.Equal(t, -1*time.Second, config.RetryConfig.InitialDelay)
	})

	t.Run("zero values", func(t *testing.T) {
		config := NewConfigBuilder().
			WithMaxConcurrency(0).
			WithStageTimeout("zero-stage", 0).
			Build()

		assert.Equal(t, 0, config.MaxConcurrency)
		assert.Equal(t, time.Duration(0), config.GetStageTimeout("zero-stage"))
	})

	t.Run("very large values", func(t *testing.T) {
		largeTimeout := 365 * 24 * time.Hour // 1 year
		config := NewConfigBuilder().
			WithMaxConcurrency(1000000).
			WithStageTimeout("large-stage", largeTimeout).
			Build()

		assert.Equal(t, 1000000, config.MaxConcurrency)
		assert.Equal(t, largeTimeout, config.GetStageTimeout("large-stage"))
	})

	t.Run("empty string parameters", func(t *testing.T) {
		config := NewConfigBuilder().
			WithStageTimeout("", 5*time.Minute).
			WithStepConfig("", "empty-stage-config").
			WithCheckpoints(true, "").
			Build()

		// Should handle empty strings gracefully
		assert.Equal(t, 5*time.Minute, config.GetStageTimeout(""))

		value, exists := config.GetStepConfig("")
		assert.True(t, exists)
		assert.Equal(t, "empty-stage-config", value)

		assert.Equal(t, "data/checkpoints", config.CheckpointDir) // Should keep default
	})
}
