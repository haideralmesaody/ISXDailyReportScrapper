package operations

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultDeltaInterval    = 250 * time.Millisecond
	defaultSnapshotInterval = 30 * time.Second
	maxDeltaStringLength    = 256
)

type stageDeltaPayload struct {
	status    string
	progress  int
	message   string
	metadata  map[string]interface{}
	timestamp time.Time
	sequence  uint64
}

// StageBroadcaster interface for stage-specific broadcasting functionality
type StageBroadcaster interface {
	// UpdateProgress sends a basic progress update
	UpdateProgress(progress int, message string)

	// UpdateProgressWithMetadata sends a progress update with rich metadata
	UpdateProgressWithMetadata(progress int, message string, metadata map[string]interface{})

	// Complete marks the stage as completed
	Complete(message string)

	// Fail marks the stage as failed with an error
	Fail(err error)

	// Skip marks the stage as skipped with a reason
	Skip(reason string)

	// GetStageID returns the stage identifier
	GetStageID() string
}

// BaseStageBroadcaster provides common broadcasting functionality
type BaseStageBroadcaster struct {
	operationID string
	stageID     string
	hub         WebSocketHub
	logger      *slog.Logger
	enabled     bool
	mu          sync.RWMutex
	startedAt   time.Time

	// Delta throttling
	deltaInterval        time.Duration
	snapshotInterval     time.Duration
	deltaTimer           *time.Timer
	deltaSequence        uint64
	pendingDelta         *stageDeltaPayload
	lastSnapshotAt       time.Time
	lastSnapshotStatus   string
	lastSnapshotProgress int
}

// NewBaseStageBroadcaster creates a new base stage broadcaster
func NewBaseStageBroadcaster(operationID, stageID string, hub WebSocketHub, logger *slog.Logger, enabled bool) *BaseStageBroadcaster {
	return &BaseStageBroadcaster{
		operationID:      operationID,
		stageID:          stageID,
		hub:              hub,
		logger:           logger,
		enabled:          enabled,
		deltaInterval:    defaultDeltaInterval,
		snapshotInterval: defaultSnapshotInterval,
	}
}

// UpdateProgress sends a basic progress update
func (b *BaseStageBroadcaster) UpdateProgress(progress int, message string) {
	b.UpdateProgressWithMetadata(progress, message, nil)
}

// UpdateProgressWithMetadata sends a progress update with rich metadata
func (b *BaseStageBroadcaster) UpdateProgressWithMetadata(progress int, message string, metadata map[string]interface{}) {
	if !b.enabled {
		return
	}

	status := "running"

	// Combine progress with stage metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if metaStatus, ok := metadata["status"].(string); ok && metaStatus != "" {
		status = metaStatus
	} else if progress >= 100 {
		status = "completed"
	}
	// Keep metadata status aligned with the resolved status to avoid flicker
	metadata["status"] = status
	metadata["progress"] = progress
	metadata["message"] = message

	now := time.Now()
	b.queueDelta(status, progress, message, metadata, now)
	b.maybeBroadcastSnapshot(status, progress, message, metadata, now)

	if b.logger != nil {
		b.logger.Info("stage progress with metadata",
			"operation_id", b.operationID,
			"stage", b.stageID,
			"progress", progress,
			"message", message,
			"metadata_keys", len(metadata),
		)
	}
}

// Complete marks the stage as completed
func (b *BaseStageBroadcaster) Complete(message string) {
	b.CompleteWithMetadata(message, nil)
}

// CompleteWithMetadata broadcasts a completion update with custom metadata.
func (b *BaseStageBroadcaster) CompleteWithMetadata(message string, metadata map[string]interface{}) {
	if !b.enabled {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["progress"] = 100
	metadata["message"] = message
	metadata["status"] = "completed"

	b.flushPendingDelta()
	now := time.Now()
	b.queueDelta("completed", 100, message, metadata, now)
	b.flushPendingDelta()
	b.broadcastSnapshot("completed", 100, message, metadata)

	if b.logger != nil {
		b.logger.Info("stage completed",
			"operation_id", b.operationID,
			"stage", b.stageID,
			"message", message,
		)
	}
}

// Fail marks the stage as failed with an error
func (b *BaseStageBroadcaster) Fail(err error) {
	if !b.enabled {
		return
	}

	metadata := map[string]interface{}{
		"error":   err.Error(),
		"status":  "failed",
		"message": "Stage execution failed",
	}

	b.flushPendingDelta()
	now := time.Now()
	b.queueDelta("failed", 0, err.Error(), metadata, now)
	b.flushPendingDelta()
	b.broadcastSnapshot("failed", 0, err.Error(), metadata)

	if b.logger != nil {
		b.logger.Error("stage failed",
			"operation_id", b.operationID,
			"stage", b.stageID,
			"error", err,
		)
	}
}

// Skip marks the stage as skipped with a reason
func (b *BaseStageBroadcaster) Skip(reason string) {
	if !b.enabled {
		return
	}

	metadata := map[string]interface{}{
		"reason":  reason,
		"status":  "skipped",
		"message": "Stage execution skipped",
	}

	b.flushPendingDelta()
	now := time.Now()
	b.queueDelta("skipped", 0, reason, metadata, now)
	b.flushPendingDelta()
	b.broadcastSnapshot("skipped", 0, reason, metadata)

	if b.logger != nil {
		b.logger.Info("stage skipped",
			"operation_id", b.operationID,
			"stage", b.stageID,
			"reason", reason,
		)
	}
}

// GetStageID returns the stage identifier
func (b *BaseStageBroadcaster) GetStageID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stageID
}

func (b *BaseStageBroadcaster) broadcastSnapshot(status string, progress int, message string, metadata map[string]interface{}) {
	if !b.enabled || b.hub == nil {
		return
	}

	if b.operationID == "" || b.operationID == "temp" {
		if b.logger != nil {
			b.logger.Debug("skipping stage broadcast due to unresolved operation id",
				"stage_id", b.stageID,
				"status", status,
				"message", message,
			)
		}
		return
	}

	clampedProgress := clampProgress(progress)
	now := time.Now()
	startedAt := b.ensureStartedAt(now)
	b.mu.Lock()
	b.lastSnapshotAt = now
	b.lastSnapshotStatus = status
	b.lastSnapshotProgress = clampedProgress
	b.mu.Unlock()
	stageName := stageNameForID(b.stageID)

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if _, exists := metadata["stage_id"]; !exists {
		metadata["stage_id"] = b.stageID
	}
	metadata["progress_percent"] = clampedProgress
	metadata["status"] = status
	metadata["telemetry_missing"] = false
	if message != "" {
		metadata["message"] = message
	}
	metadata["updated_at"] = now.Format(time.RFC3339)

	step := map[string]interface{}{
		"id":         b.stageID,
		"stage_id":   b.stageID,
		"name":       stageName,
		"status":     status,
		"progress":   clampedProgress,
		"message":    message,
		"metadata":   metadata,
		"updated_at": now.Format(time.RFC3339),
	}

	snapshot := map[string]interface{}{
		"operation_id": b.operationID,
		"status":       status,
		"progress":     clampedProgress,
		"current_step": b.stageID,
		"stage_id":     b.stageID,
		"stage_name":   stageName,
		"message":      message,
		"started_at":   startedAt.Format(time.RFC3339),
		"updated_at":   now.Format(time.RFC3339),
		"metadata":     metadata,
		"steps":        []interface{}{step},
	}
	if status == "completed" {
		snapshot["completed_at"] = now.Format(time.RFC3339)
	}

	b.hub.BroadcastUpdate(EventTypeOperationSnapshot, b.stageID, status, snapshot)
}

func (b *BaseStageBroadcaster) maybeBroadcastSnapshot(status string, progress int, message string, metadata map[string]interface{}, ts time.Time) {
	if b.shouldSendSnapshot(status, progress, ts) {
		b.broadcastSnapshot(status, progress, message, metadata)
	}
}

func (b *BaseStageBroadcaster) shouldSendSnapshot(status string, progress int, ts time.Time) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.lastSnapshotAt.IsZero() {
		return true
	}
	if status != b.lastSnapshotStatus {
		return true
	}
	if progress >= 100 && b.lastSnapshotProgress < 100 {
		return true
	}
	if absInt(progress-b.lastSnapshotProgress) >= 5 {
		return true
	}
	if ts.Sub(b.lastSnapshotAt) >= b.snapshotInterval {
		return true
	}
	return false
}

func (b *BaseStageBroadcaster) queueDelta(status string, progress int, message string, metadata map[string]interface{}, ts time.Time) {
	if !b.enabled || b.hub == nil {
		return
	}

	payload := &stageDeltaPayload{
		status:    status,
		progress:  clampProgress(progress),
		message:   message,
		timestamp: ts,
		metadata:  sanitizeDeltaMetadata(metadata),
	}

	b.mu.Lock()
	b.deltaSequence++
	payload.sequence = b.deltaSequence
	b.pendingDelta = payload
	if b.deltaTimer != nil {
		b.deltaTimer.Stop()
	}
	interval := b.deltaInterval
	if interval <= 0 {
		interval = defaultDeltaInterval
	}
	b.deltaTimer = time.AfterFunc(interval, func() {
		b.flushPendingDelta()
	})
	b.mu.Unlock()
}

func (b *BaseStageBroadcaster) flushPendingDelta() {
	var payload *stageDeltaPayload

	b.mu.Lock()
	if b.pendingDelta != nil {
		payload = b.pendingDelta
		b.pendingDelta = nil
	}
	if b.deltaTimer != nil {
		b.deltaTimer.Stop()
		b.deltaTimer = nil
	}
	b.mu.Unlock()

	if payload != nil {
		b.broadcastDelta(payload)
	}
}

func (b *BaseStageBroadcaster) broadcastDelta(payload *stageDeltaPayload) {
	if payload == nil || !b.enabled || b.hub == nil {
		return
	}
	if b.operationID == "" || b.operationID == "temp" {
		return
	}

	data := map[string]interface{}{
		"operation_id": b.operationID,
		"stage_id":     b.stageID,
		"status":       payload.status,
		"progress":     payload.progress,
		"message":      payload.message,
		"updated_at":   payload.timestamp.Format(time.RFC3339),
		"sequence":     payload.sequence,
	}

	if len(payload.metadata) > 0 {
		data["metadata"] = payload.metadata
	}

	b.hub.BroadcastUpdate(EventTypeOperationDelta, b.stageID, payload.status, data)
}

func sanitizeDeltaMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	slim := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if sanitized, ok := convertLightweightValue(value); ok {
			slim[key] = sanitized
		}
	}

	if len(slim) == 0 {
		return nil
	}

	return slim
}

func convertLightweightValue(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		if len(v) > maxDeltaStringLength {
			return v[:maxDeltaStringLength], true
		}
		return v, true
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return v, true
	case time.Time:
		return v.Format(time.RFC3339), true
	case nil:
		return nil, false
	default:
		return nil, false
	}
}

func (b *BaseStageBroadcaster) ensureStartedAt(ts time.Time) time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.startedAt.IsZero() {
		b.startedAt = ts
	}
	return b.startedAt
}

// IsEnabled returns whether broadcasting is enabled
func (b *BaseStageBroadcaster) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// SetEnabled enables or disables broadcasting
func (b *BaseStageBroadcaster) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

// RichTelemetryBroadcaster extends the basic interface for complex stages
type RichTelemetryBroadcaster interface {
	StageBroadcaster

	// Advanced broadcasting methods
	UpdateProgressWithTelemetry(progress int, message string, telemetry interface{})
	UpdateFileProgress(fileName string, progress float64, status string)
	UpdateVelocityMetrics(velocity map[string]interface{})
	BroadcastMultiSourceUpdate(sources []MultiSourceProgress)
}

// MultiSourceProgress represents a source of progress updates for multi-source broadcasting
type MultiSourceProgress struct {
	SourceID string
	Progress float64
	Metadata map[string]interface{}
	Priority int
}

// Helper functions for decentralized broadcasting
// Consolidated here to avoid duplication across stage files

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampProgress(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// pickInt extracts the first non-zero integer from the given values
func pickInt(values ...interface{}) int {
	for _, v := range values {
		switch val := v.(type) {
		case int:
			if val != 0 {
				return val
			}
		case int32:
			if val != 0 {
				return int(val)
			}
		case int64:
			if val != 0 {
				return int(val)
			}
		case float64:
			if val != 0 {
				return int(val)
			}
		case float32:
			if val != 0 {
				return int(val)
			}
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && parsed != 0 {
				return parsed
			}
		case nil:
			continue
		}
	}
	return 0
}

// pickFloat extracts a float64 value from the given interface
func pickFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return parsed
		}
	}
	return 0
}

// FileProcessingStatus is defined in file_progress_tracker.go to avoid duplication

// FileProgressBroadcaster handles file-level progress tracking with atomic operations
type FileProgressBroadcaster struct {
	*BaseStageBroadcaster
	totalFiles     int64        // atomic counter
	completedFiles int64        // atomic counter
	failedFiles    int64        // atomic counter
	currentFile    atomic.Value // stores string
	// REMOVED: fileStatuses map and mu mutex to eliminate deadlock
}

// NewFileProgressBroadcaster creates a new file progress broadcaster
func NewFileProgressBroadcaster(operationID, stageID string, hub WebSocketHub, logger *slog.Logger, enabled bool) *FileProgressBroadcaster {
	fpb := &FileProgressBroadcaster{
		BaseStageBroadcaster: NewBaseStageBroadcaster(operationID, stageID, hub, logger, enabled),
	}
	fpb.currentFile.Store("") // Initialize atomic value
	return fpb
}

// UpdateFileStatus updates the status of a specific file using atomic operations
func (f *FileProgressBroadcaster) UpdateFileStatus(fileName string, status *FileProcessingStatus) {
	// Update atomic counters based on status
	if status.Status == "completed" {
		atomic.AddInt64(&f.completedFiles, 1)
	} else if status.Status == "failed" {
		atomic.AddInt64(&f.failedFiles, 1)
	}

	// Update current file
	f.currentFile.Store(fileName)

	// Broadcast individual file update (no mutex needed)
	totalFiles := atomic.LoadInt64(&f.totalFiles)
	completedFiles := atomic.LoadInt64(&f.completedFiles)

	metadata := map[string]interface{}{
		"file_status":      status,
		"total_files":      totalFiles,
		"processed_files":  completedFiles,
		"overall_progress": f.GetOverallProgress(),
	}

	f.BaseStageBroadcaster.UpdateProgressWithMetadata(int(f.GetOverallProgress()), "File status updated: "+fileName, metadata)
}

// GetOverallProgress calculates overall progress using atomic operations (deadlock-free)
func (f *FileProgressBroadcaster) GetOverallProgress() float64 {
	totalFiles := atomic.LoadInt64(&f.totalFiles)
	completedFiles := atomic.LoadInt64(&f.completedFiles)

	if totalFiles == 0 {
		return 0.0
	}

	// Calculate progress: completed files count as 100%
	progress := (float64(completedFiles) / float64(totalFiles)) * 100
	return progress
}

// BroadcastFileSummary broadcasts a summary of all file processing (deadlock-free)
func (f *FileProgressBroadcaster) BroadcastFileSummary() {
	// NO LOCKS - use atomic operations only
	totalFiles := atomic.LoadInt64(&f.totalFiles)
	completedFiles := atomic.LoadInt64(&f.completedFiles)
	failedFiles := atomic.LoadInt64(&f.failedFiles)
	currentFileInterface := f.currentFile.Load()
	currentFile := ""
	if currentFileInterface != nil {
		currentFile = currentFileInterface.(string)
	}

	// Generate file statuses for API compatibility
	fileStatuses := f.generateFileStatuses()

	metadata := map[string]interface{}{
		"total_files":      totalFiles,
		"completed_files":  completedFiles,
		"failed_files":     failedFiles,
		"file_statuses":    fileStatuses,
		"current_file":     currentFile,
		"overall_progress": f.GetOverallProgress(),
	}

	f.BaseStageBroadcaster.UpdateProgressWithMetadata(int(f.GetOverallProgress()), "File processing summary", metadata)
}

// generateFileStatuses converts atomic counters to FileProcessingStatus array for API compatibility
func (f *FileProgressBroadcaster) generateFileStatuses() []*FileProcessingStatus {
	totalFiles := atomic.LoadInt64(&f.totalFiles)
	completedFiles := atomic.LoadInt64(&f.completedFiles)
	failedFiles := atomic.LoadInt64(&f.failedFiles)
	currentFileInterface := f.currentFile.Load()
	currentFile := ""
	if currentFileInterface != nil {
		currentFile = currentFileInterface.(string)
	}

	var statuses []*FileProcessingStatus

	// Add completed files
	for i := int64(0); i < completedFiles; i++ {
		statuses = append(statuses, &FileProcessingStatus{
			FileName: fmt.Sprintf("completed_file_%d", i+1),
			Status:   "completed",
			Progress: 100,
		})
	}

	// Add failed files
	for i := int64(0); i < failedFiles; i++ {
		statuses = append(statuses, &FileProcessingStatus{
			FileName: fmt.Sprintf("failed_file_%d", i+1),
			Status:   "failed",
			Progress: 0,
		})
	}

	// Add current processing file
	if currentFile != "" {
		statuses = append(statuses, &FileProcessingStatus{
			FileName: currentFile,
			Status:   "processing",
			Progress: 0,
		})
	}

	// Add pending files
	processedCount := completedFiles + failedFiles
	if currentFile != "" {
		processedCount++ // Current file counted
	}
	pendingCount := totalFiles - processedCount
	for i := int64(0); i < pendingCount; i++ {
		statuses = append(statuses, &FileProcessingStatus{
			FileName: fmt.Sprintf("pending_file_%d", i+1),
			Status:   "pending",
			Progress: 0,
		})
	}

	return statuses
}

// SetTotalFiles sets the total number of files to process
func (f *FileProgressBroadcaster) SetTotalFiles(count int64) {
	atomic.StoreInt64(&f.totalFiles, count)
}

// SetCurrentFile updates the currently processing file
func (f *FileProgressBroadcaster) SetCurrentFile(fileName string) {
	f.currentFile.Store(fileName)
}

// Reset resets all counters for new operation
func (f *FileProgressBroadcaster) Reset() {
	atomic.StoreInt64(&f.totalFiles, 0)
	atomic.StoreInt64(&f.completedFiles, 0)
	atomic.StoreInt64(&f.failedFiles, 0)
	f.currentFile.Store("")
}

// GetCompletedFiles returns the count of completed files
func (f *FileProgressBroadcaster) GetCompletedFiles() int64 {
	return atomic.LoadInt64(&f.completedFiles)
}

// GetFailedFiles returns the count of failed files
func (f *FileProgressBroadcaster) GetFailedFiles() int64 {
	return atomic.LoadInt64(&f.failedFiles)
}

// GetTotalFiles returns the total number of files
func (f *FileProgressBroadcaster) GetTotalFiles() int64 {
	return atomic.LoadInt64(&f.totalFiles)
}

// VelocityTrackingBroadcaster handles velocity tracking for scraping operations
type VelocityTrackingBroadcaster struct {
	*BaseStageBroadcaster
	velocityMetrics map[string]float64
	lastUpdate      time.Time
	mu              sync.RWMutex
}

// NewVelocityTrackingBroadcaster creates a new velocity tracking broadcaster
func NewVelocityTrackingBroadcaster(operationID, stageID string, hub WebSocketHub, logger *slog.Logger, enabled bool) *VelocityTrackingBroadcaster {
	return &VelocityTrackingBroadcaster{
		BaseStageBroadcaster: NewBaseStageBroadcaster(operationID, stageID, hub, logger, enabled),
		velocityMetrics:      make(map[string]float64),
		lastUpdate:           time.Now(),
	}
}

// UpdateVelocity updates a specific velocity metric
func (v *VelocityTrackingBroadcaster) UpdateVelocity(metric string, value float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.velocityMetrics[metric] = value
	v.lastUpdate = time.Now()

	v.BroadcastVelocitySnapshot()
}

// GetDownloadVelocity returns the current download velocity
func (v *VelocityTrackingBroadcaster) GetDownloadVelocity() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.velocityMetrics["download_velocity"]
}

// BroadcastVelocitySnapshot broadcasts current velocity metrics
func (v *VelocityTrackingBroadcaster) BroadcastVelocitySnapshot() {
	v.mu.RLock()
	defer v.mu.RUnlock()

	metadata := map[string]interface{}{
		"velocity_metrics": v.velocityMetrics,
		"last_update":      v.lastUpdate,
	}

	// Use a reasonable progress estimate based on velocity
	progress := 50 // Use placeholder progress
	v.BaseStageBroadcaster.UpdateProgressWithMetadata(progress, "Velocity metrics updated", metadata)
}

// UpdateProgressWithTelemetry implements RichTelemetryBroadcaster interface
func (v *VelocityTrackingBroadcaster) UpdateProgressWithTelemetry(progress int, message string, telemetry interface{}) {
	v.BaseStageBroadcaster.UpdateProgressWithMetadata(progress, message, map[string]interface{}{
		"velocity_metrics": v.velocityMetrics,
	})
}

// UpdateFileProgress implements RichTelemetryBroadcaster interface
func (v *VelocityTrackingBroadcaster) UpdateFileProgress(fileName string, progress float64, status string) {
	v.BaseStageBroadcaster.UpdateProgressWithMetadata(int(progress), "File progress: "+fileName, map[string]interface{}{
		"file_name": fileName,
		"progress":  progress,
		"status":    status,
	})
}

// UpdateVelocityMetrics implements RichTelemetryBroadcaster interface
func (v *VelocityTrackingBroadcaster) UpdateVelocityMetrics(velocity map[string]interface{}) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Convert interface{} values to float64
	for key, val := range velocity {
		if floatVal, ok := val.(float64); ok {
			v.velocityMetrics[key] = floatVal
		}
	}

	v.lastUpdate = time.Now()
	v.BroadcastVelocitySnapshot()
}

// BroadcastMultiSourceUpdate implements RichTelemetryBroadcaster interface
func (v *VelocityTrackingBroadcaster) BroadcastMultiSourceUpdate(sources []MultiSourceProgress) {
	// Aggregate progress from sources
	totalProgress := 0.0
	for _, source := range sources {
		totalProgress += source.Progress
	}
	if len(sources) > 0 {
		totalProgress /= float64(len(sources))
	}

	metadata := map[string]interface{}{
		"progress":         int(totalProgress),
		"message":          "Multi-source velocity update",
		"sources_count":    len(sources),
		"velocity_metrics": v.velocityMetrics,
		"last_update":      v.lastUpdate,
	}

	v.BaseStageBroadcaster.UpdateProgressWithMetadata(int(totalProgress), "Multi-source velocity update", metadata)
}
