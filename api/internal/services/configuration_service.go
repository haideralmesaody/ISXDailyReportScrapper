package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/isxcli/isxcli/internal/operations"
)

// ConfigurationService provides centralized configuration for all operations and stages
type ConfigurationService struct {
	logger *slog.Logger
}

// NewConfigurationService creates a new configuration service
func NewConfigurationService(logger *slog.Logger) *ConfigurationService {
	return &ConfigurationService{
		logger: logger,
	}
}

// GetOperationTypes returns all available operation types with their configurations
func (cs *ConfigurationService) GetOperationTypes(ctx context.Context) ([]operations.OperationType, error) {
	cs.logger.InfoContext(ctx, "getting_operation_types",
		slog.String("component", "configuration_service"),
		slog.String("method", "GetOperationTypes"))

	operationTypes := []operations.OperationType{
		{ID: operations.StageIDScraping, Name: "Scrape ISX Data", Description: "Download daily reports", CanRunAlone: true},
		{ID: operations.StageIDProcessing, Name: "Process Data", Description: "Convert Excel to CSV", CanRunAlone: true, Dependencies: []string{operations.StageIDScraping}},
		{ID: operations.StageIDIndices, Name: "Extract Indices", Description: "Extract index data", CanRunAlone: true, Dependencies: []string{operations.StageIDProcessing}},
		{ID: operations.StageIDLiquidity, Name: "Calculate Liquidity", Description: "Calculate liquidity metrics", CanRunAlone: true, Dependencies: []string{operations.StageIDProcessing}},
	}

	cs.logger.InfoContext(ctx, "operation_types_retrieved_successfully",
		slog.Int("total_operations", len(operationTypes)),
		slog.String("operations", "scraping, processing, indices, liquidity"))

	return operationTypes, nil
}
func (cs *ConfigurationService) GetStageParameters(ctx context.Context, stageID string) ([]operations.ParameterDefinition, error) {
	cs.logger.InfoContext(ctx, "getting_stage_parameters",
		slog.String("component", "configuration_service"),
		slog.String("method", "GetStageParameters"),
		slog.String("stage_id", stageID))

	switch stageID {
	case operations.StageIDScraping:
		today := time.Now().Format("2006-01-02")
		defaultFromDate := "2025-01-01"
		return []operations.ParameterDefinition{
			{
				Name:         "mode",
				Type:         "select",
				Description:  "Scraping mode",
				Required:     false,
				DefaultValue: "initial",
				Options:      []string{"initial", "incremental", "backfill"},
			},
			{
				Name:         "from",
				Type:         "date",
				Description:  "Start date (YYYY-MM-DD)",
				Required:     false,
				DefaultValue: defaultFromDate,
			},
			{
				Name:         "to",
				Type:         "date",
				Description:  "End date (YYYY-MM-DD)",
				Required:     false,
				DefaultValue: today,
			},
		}, nil

	case operations.StageIDProcessing, operations.StageIDIndices, operations.StageIDLiquidity:
		// These stages use default directories, no parameters needed
		return []operations.ParameterDefinition{}, nil

	default:
		cs.logger.WarnContext(ctx, "unknown_stage_type_for_parameters",
			slog.String("stage_id", stageID))
		return []operations.ParameterDefinition{}, nil
	}
}

// GetStageSummary returns summary information about available stages
func (cs *ConfigurationService) GetStageSummary() map[string]interface{} {
	return map[string]interface{}{
		"steps": []map[string]interface{}{
			{
				"id":          operations.StageIDScraping,
				"name":        "Scraping",
				"description": "Download daily reports from ISX website",
				"executable":  "scraper.exe",
			},
			{
				"id":          operations.StageIDProcessing,
				"name":        "Processing",
				"description": "Process Excel files into CSV format",
				"executable":  "processor.exe",
			},
			{
				"id":          operations.StageIDIndices,
				"name":        "Index Extraction",
				"description": "Extract market indices from processed data",
				"executable":  "indexcsv.exe",
			},
			{
				"id":          operations.StageIDLiquidity,
				"name":        "Liquidity Calculation",
				"description": "Calculate hybrid liquidity metrics and generate liquidity analysis reports",
				"executable":  "",
			},
		},
	}
}
