package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type StrategyRunInfo struct {
	RunID       string    `json:"run_id"`
	StrategyID  string    `json:"strategy_id"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	Total     int `json:"total"`
	BuyCount  int `json:"buy_count"`
	SellCount int `json:"sell_count"`
	HoldCount int `json:"hold_count"`

	ErrorCount  int  `json:"error_count"`
	HasBacktest bool `json:"has_backtest"`
}

type persistedRunSummary struct {
	RunID       string              `json:"run_id"`
	StrategyID  string              `json:"strategy_id"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at"`
	Total       int                 `json:"total"`
	BuyCount    int                 `json:"buy_count"`
	SellCount   int                 `json:"sell_count"`
	HoldCount   int                 `json:"hold_count"`
	Errors      []ExecuteBatchError `json:"errors,omitempty"`
	Backtest    json.RawMessage     `json:"backtest,omitempty"`
}

// ListStrategyRuns loads persisted batch strategy runs from:
// <data_dir>/strategies/<strategy_id>/<run_id>/summary.json
func (s *StrategyService) ListStrategyRuns(ctx context.Context, strategyID string, limit int) ([]StrategyRunInfo, error) {
	if s.dataService == nil || s.dataService.paths == nil {
		return nil, fmt.Errorf("data service paths not configured")
	}

	strategyID = strings.TrimSpace(strategyID)
	if strategyID == "" {
		return nil, fmt.Errorf("strategy_id is required")
	}

	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	baseDir := filepath.Join(s.dataService.paths.DataDir, "strategies", strategyID)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []StrategyRunInfo{}, nil
		}
		return nil, fmt.Errorf("read runs dir: %w", err)
	}

	runs := make([]StrategyRunInfo, 0, len(entries))
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !entry.IsDir() {
			continue
		}

		runID := strings.TrimSpace(entry.Name())
		if runID == "" {
			continue
		}

		summaryPath := filepath.Join(baseDir, runID, "summary.json")
		bytes, err := os.ReadFile(summaryPath)
		if err != nil {
			continue
		}

		var persisted persistedRunSummary
		if err := json.Unmarshal(bytes, &persisted); err != nil {
			continue
		}

		hasBacktest := len(persisted.Backtest) > 0 && string(persisted.Backtest) != "null"

		runs = append(runs, StrategyRunInfo{
			RunID:       persisted.RunID,
			StrategyID:  persisted.StrategyID,
			StartedAt:   persisted.StartedAt,
			CompletedAt: persisted.CompletedAt,
			Total:       persisted.Total,
			BuyCount:    persisted.BuyCount,
			SellCount:   persisted.SellCount,
			HoldCount:   persisted.HoldCount,
			ErrorCount:  len(persisted.Errors),
			HasBacktest: hasBacktest,
		})
	}

	sort.Slice(runs, func(i, j int) bool {
		if runs[i].StartedAt.Equal(runs[j].StartedAt) {
			return runs[i].RunID > runs[j].RunID
		}
		return runs[i].StartedAt.After(runs[j].StartedAt)
	})

	if len(runs) > limit {
		runs = runs[:limit]
	}

	return runs, nil
}

// GetStrategyRun loads a single persisted batch strategy run summary from:
// <data_dir>/strategies/<strategy_id>/<run_id>/summary.json
func (s *StrategyService) GetStrategyRun(ctx context.Context, strategyID, runID string) (ExecuteBatchResponse, error) {
	if s.dataService == nil || s.dataService.paths == nil {
		return ExecuteBatchResponse{}, fmt.Errorf("data service paths not configured")
	}

	strategyID = strings.TrimSpace(strategyID)
	runID = strings.TrimSpace(runID)
	if strategyID == "" {
		return ExecuteBatchResponse{}, fmt.Errorf("strategy_id is required")
	}
	if runID == "" {
		return ExecuteBatchResponse{}, fmt.Errorf("run_id is required")
	}

	select {
	case <-ctx.Done():
		return ExecuteBatchResponse{}, ctx.Err()
	default:
	}

	path := filepath.Join(s.dataService.paths.DataDir, "strategies", strategyID, runID, "summary.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return ExecuteBatchResponse{}, fmt.Errorf("read run summary (%s): %w", path, err)
	}

	var resp ExecuteBatchResponse
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return ExecuteBatchResponse{}, fmt.Errorf("parse run summary: %w", err)
	}
	return resp, nil
}
