package progress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Metrics stores historical timing data for ETA predictions
type Metrics struct {
	mu      sync.RWMutex
	Stage   string                 `json:"stage"`
	History []ProcessingHistory    `json:"history"`
	Average map[string]float64     `json:"average_times"`
}

// ProcessingHistory records timing for a single processing run
type ProcessingHistory struct {
	Timestamp     time.Time `json:"timestamp"`
	TotalItems    int       `json:"total_items"`
	TotalDuration float64   `json:"total_duration_seconds"`
	AvgPerItem    float64   `json:"avg_per_item_seconds"`
}

// MetricsManager handles loading and saving metrics
type MetricsManager struct {
	metricsPath string
	metrics     map[string]*Metrics
	mu          sync.RWMutex
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(dataPath string) *MetricsManager {
	metricsPath := filepath.Join(dataPath, "metrics")
	os.MkdirAll(metricsPath, 0755)
	
	return &MetricsManager{
		metricsPath: metricsPath,
		metrics:     make(map[string]*Metrics),
	}
}

// LoadMetrics loads metrics for a specific stage
func (m *MetricsManager) LoadMetrics(stage string) (*Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if metrics, exists := m.metrics[stage]; exists {
		return metrics, nil
	}
	
	filename := filepath.Join(m.metricsPath, stage+"_metrics.json")
	metrics := &Metrics{
		Stage:   stage,
		History: make([]ProcessingHistory, 0),
		Average: make(map[string]float64),
	}
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// File doesn't exist, return empty metrics
		m.metrics[stage] = metrics
		return metrics, nil
	}
	
	if err := json.Unmarshal(data, metrics); err != nil {
		return nil, err
	}
	
	m.metrics[stage] = metrics
	return metrics, nil
}

// SaveMetrics saves metrics for a specific stage
func (m *MetricsManager) SaveMetrics(stage string) error {
	m.mu.RLock()
	metrics, exists := m.metrics[stage]
	m.mu.RUnlock()
	
	if !exists {
		return nil
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	return m.saveMetricsLocked(stage, metrics)
}

// saveMetricsLocked saves metrics when the mutex is already held
func (m *MetricsManager) saveMetricsLocked(stage string, metrics *Metrics) error {
	filename := filepath.Join(m.metricsPath, stage+"_metrics.json")
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(filename, data, 0644)
}

// RecordRun adds a new processing run to the metrics
func (m *MetricsManager) RecordRun(stage string, totalItems int, duration time.Duration) error {
	metrics, err := m.LoadMetrics(stage)
	if err != nil {
		return err
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	avgPerItem := duration.Seconds() / float64(totalItems)
	
	history := ProcessingHistory{
		Timestamp:     time.Now(),
		TotalItems:    totalItems,
		TotalDuration: duration.Seconds(),
		AvgPerItem:    avgPerItem,
	}
	
	metrics.History = append(metrics.History, history)
	
	// Keep only last 100 runs
	if len(metrics.History) > 100 {
		metrics.History = metrics.History[len(metrics.History)-100:]
	}
	
	// Update averages
	metrics.updateAverages()
	
	// Call the locked version since we already hold the mutex
	return m.saveMetricsLocked(stage, metrics)
}

// updateAverages calculates average times from history
func (m *Metrics) updateAverages() {
	if len(m.History) == 0 {
		return
	}
	
	var totalAvg float64
	for _, h := range m.History {
		totalAvg += h.AvgPerItem
	}
	
	m.Average["per_item"] = totalAvg / float64(len(m.History))
}

// GetAverageTimePerItem returns the average processing time per item
func (m *MetricsManager) GetAverageTimePerItem(stage string) (time.Duration, error) {
	metrics, err := m.LoadMetrics(stage)
	if err != nil {
		return 0, err
	}
	
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	
	if avgTime, exists := metrics.Average["per_item"]; exists && avgTime > 0 {
		return time.Duration(avgTime * float64(time.Second)), nil
	}
	
	// No historical data, return 0
	return 0, nil
}

// EnhancedCalculator extends Calculator with historical metrics
type EnhancedCalculator struct {
	*Calculator
	metricsManager *MetricsManager
	historicalAvg  time.Duration
}

// NewEnhancedCalculator creates a calculator with metrics support
func NewEnhancedCalculator(stage string, totalItems int, metricsManager *MetricsManager) *EnhancedCalculator {
	calc := NewCalculator(stage, totalItems)
	
	// Load historical average if available
	historicalAvg, _ := metricsManager.GetAverageTimePerItem(stage)
	
	return &EnhancedCalculator{
		Calculator:     calc,
		metricsManager: metricsManager,
		historicalAvg:  historicalAvg,
	}
}

// GetEnhancedETA provides ETA based on historical data
func (ec *EnhancedCalculator) GetEnhancedETA() string {
	// If we have processed items, use actual timing
	if ec.ProcessedItems > 0 {
		return ec.GetETA()
	}
	
	// Otherwise, use historical average if available
	if ec.historicalAvg > 0 {
		remainingTime := ec.historicalAvg * time.Duration(ec.TotalItems)
		
		if remainingTime < time.Minute {
			return fmt.Sprintf("%d seconds remaining (estimated)", int(remainingTime.Seconds()))
		} else if remainingTime < time.Hour {
			minutes := int(remainingTime.Minutes())
			return fmt.Sprintf("%d minutes remaining (estimated)", minutes)
		} else {
			hours := int(remainingTime.Hours())
			minutes := int(remainingTime.Minutes()) % 60
			return fmt.Sprintf("%dh %dm remaining (estimated)", hours, minutes)
		}
	}
	
	return "Calculating..."
}

// Complete records the completion of processing
func (ec *EnhancedCalculator) Complete() error {
	duration := time.Since(ec.StartTime)
	return ec.metricsManager.RecordRun(ec.Stage, ec.TotalItems, duration)
}