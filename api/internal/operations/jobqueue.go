package operations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type stageExecutionFailedError struct {
	stageID string
	err     error
}

func (e stageExecutionFailedError) Error() string {
	if e.stageID == "" {
		return fmt.Sprintf("stage failed: %v", e.err)
	}
	return fmt.Sprintf("stage %s failed: %v", e.stageID, e.err)
}

func (e stageExecutionFailedError) Unwrap() error { return e.err }

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents an async operation job
type Job struct {
	ID          string                 `json:"id"`
	OperationID string                 `json:"operation_id"`
	StageID     string                 `json:"stage_id"`
	StageName   string                 `json:"stage_name"`
	Status      JobStatus              `json:"status"`
	Progress    int                    `json:"progress"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Request     *OperationRequest      `json:"request,omitempty"`
}

// JobStore interface for job persistence
type JobStore interface {
	// Job operations
	CreateJob(job *Job) error
	GetJob(id string) (*Job, error)
	UpdateJob(job *Job) error
	ListJobs(filter JobFilter) ([]*Job, error)
	DeleteJob(id string) error
}

// JobFilter for querying jobs
type JobFilter struct {
	Status      JobStatus
	OperationID string
	StageID     string
	Since       time.Time
	Limit       int
}

// JobQueue manages async job execution
type JobQueue struct {
	mu       sync.RWMutex
	jobs     chan *Job
	workers  int
	wg       sync.WaitGroup
	store    JobStore
	manager  *Manager
	logger   *slog.Logger
	shutdown chan struct{}
	active   map[string]*Job // Currently executing jobs
}

// broadcastJobQueueStageTerminalSnapshot emits a best-effort terminal snapshot for jobqueue-level failures
// (e.g. CanRun=false or panic) without relying on deprecated legacy progress events.
func broadcastJobQueueStageTerminalSnapshot(manager *Manager, logger *slog.Logger, operationID, stageID, status, message string, err error) {
	if manager == nil || operationID == "" || operationID == "temp" || stageID == "" {
		return
	}
	hub := manager.GetHub()
	if hub == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}

	phase := StagePhaseRunning
	progress := 0.0
	switch status {
	case "failed":
		phase = StagePhaseFailed
	case "skipped":
		phase = StagePhaseCompleted
	}

	telemetry := NewStageTelemetry(stageID, nil, phase, message, progress, 0, 0, "", nil)
	metadata := make(map[string]interface{}, 8)
	telemetry.ApplyToMetadata(metadata)
	metadata["status"] = status
	if status == "skipped" {
		metadata["stage_skipped"] = true
		metadata["reason"] = message
	}
	if err != nil {
		metadata["error"] = err.Error()
	}

	b := NewBaseStageBroadcaster(operationID, stageID, hub, logger, true)
	b.UpdateProgressWithMetadata(int(progress), message, metadata)
}

// NewJobQueue creates a new job queue
func NewJobQueue(workers int, store JobStore, manager *Manager, logger *slog.Logger) *JobQueue {
	if workers <= 0 {
		workers = 4 // Default number of workers
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &JobQueue{
		jobs:     make(chan *Job, workers*10), // Buffer size = 10x workers (was 2x workers)
		workers:  workers,
		store:    store,
		manager:  manager,
		logger:   logger.With(slog.String("component", "jobqueue")),
		shutdown: make(chan struct{}),
		active:   make(map[string]*Job),
	}
}

// Start begins processing jobs
func (q *JobQueue) Start(ctx context.Context) {
	q.logger.Info("starting job queue", slog.Int("workers", q.workers))

	// Start worker goroutines
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}

	// Start job recovery (for jobs that were running when system stopped)
	go q.recoverJobs(ctx)
}

// Stop gracefully shuts down the job queue
func (q *JobQueue) Stop(timeout time.Duration) error {
	q.logger.Info("stopping job queue")

	// Signal shutdown
	close(q.shutdown)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		q.logger.Info("job queue stopped gracefully")
		return nil
	case <-time.After(timeout):
		q.logger.Warn("job queue stop timeout exceeded")
		return fmt.Errorf("timeout waiting for workers to finish")
	}
}

// ExecuteSingleStage executes a single stage job (public method)
func (q *JobQueue) ExecuteSingleStage(ctx context.Context, job *Job, logger *slog.Logger) error {
	return q.executeSingleStage(ctx, job, logger)
}

// EnqueueOperation builds and enqueues a job from an OperationRequest
func (q *JobQueue) EnqueueOperation(req OperationRequest) (*Job, error) {
	if req.ID == "" {
		req.ID = fmt.Sprintf("operation-%d", time.Now().Unix())
	}

	stageID := req.Type
	if stageID == "" {
		if step, ok := req.Parameters["step"].(string); ok && step != "" {
			stageID = step
		}
	}
	if stageID == "" {
		return nil, fmt.Errorf("stage id required in operation request")
	}

	st, err := q.manager.GetRegistry().Get(stageID)
	if err != nil {
		return nil, fmt.Errorf("stage not found: %w", err)
	}
	stageName := st.Name()

	job := &Job{
		ID:          fmt.Sprintf("%s-%d", req.ID, time.Now().UnixNano()),
		OperationID: req.ID,
		StageID:     stageID,
		StageName:   stageName,
		Status:      JobStatusPending,
		CreatedAt:   time.Now(),
		Request:     &req,
	}

	if err := q.Enqueue(job); err != nil {
		return nil, err
	}

	return job, nil
}

// Enqueue adds a job to the queue
func (q *JobQueue) Enqueue(job *Job) error {
	// Check for existing job to prevent duplicates
	if existingJob, err := q.store.GetJob(job.ID); err == nil && existingJob != nil {
		q.logger.Info("job already exists, checking status",
			slog.String("job_id", job.ID),
			slog.String("existing_status", string(existingJob.Status)),
			slog.Time("created_at", existingJob.CreatedAt),
			slog.Duration("age", time.Since(existingJob.CreatedAt)))

		// Check if existing job is stale (stuck in pending for > 5 minutes)
		if existingJob.Status == JobStatusPending && time.Since(existingJob.CreatedAt) > 5*time.Minute {
			q.logger.Warn("detected stale pending job, replacing it",
				slog.String("job_id", job.ID),
				slog.Duration("stale_duration", time.Since(existingJob.CreatedAt)))

			// Mark stale job as failed and replace it
			existingJob.Status = JobStatusFailed
			existingJob.Error = "Job replaced due to stale timeout"
			completedAt := time.Now()
			existingJob.CompletedAt = &completedAt
			if err := q.store.UpdateJob(existingJob); err != nil {
				q.logger.Warn("failed to update stale job",
					slog.String("job_id", job.ID),
					slog.Any("error", err))
			}
		} else {
			// Return existing job if it's not stale
			return nil
		}
	}

	// Set initial status
	job.Status = JobStatusPending
	job.CreatedAt = time.Now()

	// Save to store
	if err := q.store.CreateJob(job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	// Add to queue
	q.logger.Info("attempting to enqueue job",
		slog.String("job_id", job.ID),
		slog.String("stage_id", job.StageID),
		slog.String("operation_id", job.OperationID),
		slog.Int("current_queue_depth", len(q.jobs)),
		slog.Int("queue_capacity", cap(q.jobs)),
		slog.Int("active_workers", q.workers))

	select {
	case q.jobs <- job:
		q.logger.Info("job enqueued successfully",
			slog.String("job_id", job.ID),
			slog.String("stage_id", job.StageID),
			slog.String("operation_id", job.OperationID),
			slog.Int("new_queue_depth", len(q.jobs)),
			slog.Int("active_workers", q.workers))
		return nil
	default:
		// Queue is full, mark as failed
		q.logger.Error("queue is full, job rejected",
			slog.String("job_id", job.ID),
			slog.String("stage_id", job.StageID),
			slog.String("operation_id", job.OperationID),
			slog.Int("queue_capacity", cap(q.jobs)),
			slog.Int("queue_depth", len(q.jobs)),
			slog.Int("active_workers", q.workers))
		job.Status = JobStatusFailed
		job.Error = "job queue is full"
		q.store.UpdateJob(job)
		return fmt.Errorf("job queue is full")
	}
}

// GetJob retrieves a job by ID
func (q *JobQueue) GetJob(id string) (*Job, error) {
	// Check if job is currently active
	q.mu.RLock()
	if activeJob, ok := q.active[id]; ok {
		q.mu.RUnlock()
		return activeJob, nil
	}
	q.mu.RUnlock()

	// Otherwise get from store
	return q.store.GetJob(id)
}

// CancelJob cancels a running job
func (q *JobQueue) CancelJob(id string) error {
	job, err := q.GetJob(id)
	if err != nil {
		return err
	}

	if job.Status != JobStatusRunning && job.Status != JobStatusPending {
		return fmt.Errorf("job %s cannot be cancelled (status: %s)", id, job.Status)
	}

	// Update status
	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	return q.store.UpdateJob(job)
}

// ListJobs returns jobs matching the filter
func (q *JobQueue) ListJobs(filter JobFilter) ([]*Job, error) {
	return q.store.ListJobs(filter)
}

// worker processes jobs from the queue
func (q *JobQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	logger := q.logger.With(slog.Int("worker_id", workerID))
	logger.Info("worker started",
		slog.Int("total_workers", q.workers),
		slog.Int("queue_capacity", cap(q.jobs)),
		slog.Int("current_queue_depth", len(q.jobs)))

	// DEBUG: Log worker state before entering loop
	logger.Debug("worker entering main loop",
		slog.Int("worker_id", workerID),
		slog.Bool("context_done", ctx.Done() != nil),
		slog.Bool("shutdown_channel_closed", q.shutdown != nil))

	for {
		// DEBUG: Log current queue state
		q.mu.RLock()
		queueDepth := len(q.jobs)
		q.mu.RUnlock()

		if queueDepth > 0 {
			logger.Debug("worker waiting for job",
				slog.Int("worker_id", workerID),
				slog.Int("queue_depth", queueDepth),
				slog.Int("queue_capacity", cap(q.jobs)))
		}

		select {
		case <-ctx.Done():
			logger.Warn("worker stopped by context cancellation",
				slog.Int("worker_id", workerID),
				slog.String("context_error", ctx.Err().Error()))
			return
		case <-q.shutdown:
			logger.Info("worker stopped by shutdown signal",
				slog.Int("worker_id", workerID))
			return
		case job := <-q.jobs:
			logger.Info("🎯 WORKER RECEIVED JOB - CRITICAL BREAKTHROUGH!",
				slog.String("job_id", job.ID),
				slog.String("stage_id", job.StageID),
				slog.String("operation_id", job.OperationID),
				slog.Int("queue_depth_after_receiving", len(q.jobs)),
				slog.Time("received_at", time.Now()))

			// Verify job is not nil
			if job == nil {
				logger.Error("worker received nil job - this should never happen!",
					slog.Int("worker_id", workerID))
				continue
			}

			q.processJob(ctx, job, logger)
		}
	}
}

// processJob executes a single job
func (q *JobQueue) processJob(ctx context.Context, job *Job, logger *slog.Logger) {
	// Add trace ID to context
	if job.Metadata != nil {
		if traceID, ok := job.Metadata["trace_id"].(string); ok {
			ctx = context.WithValue(ctx, middleware.RequestIDKey, traceID)
		}
	}

	logger = logger.With(
		slog.String("job_id", job.ID),
		slog.String("operation_id", job.OperationID),
		slog.String("stage_id", job.StageID),
	)

	logger.Info("processing job started")

	// Mark job as active
	q.mu.Lock()
	q.active[job.ID] = job
	q.mu.Unlock()

	// Track current stage for panic recovery (single stage mode)
	var currentStageID string

	defer func() {
		// Recover from any panics to prevent server crash
		if r := recover(); r != nil {
			logger.Error("job processing panicked",
				slog.Any("panic", r),
				slog.String("job_id", job.ID),
				slog.String("current_stage", currentStageID))

			// ✅ FIX B4: Mark the failed stage in broadcaster (QAQC #6)
				// If we know which stage panicked, mark it as failed
				if currentStageID != "" {
					panicErr := fmt.Errorf("stage panicked: %v", r)
					broadcastJobQueueStageTerminalSnapshot(
						q.manager,
						logger,
						job.OperationID,
						currentStageID,
						"failed",
						"Stage panicked",
						panicErr,
					)
				}

			// Mark job as failed
			job.Status = JobStatusFailed
			job.Error = fmt.Sprintf("job processing panicked: %v", r)
			if currentStageID != "" {
				job.Message = fmt.Sprintf("Internal error in stage %s", currentStageID)
			} else {
				job.Message = "Internal error occurred"
			}
			completedAt := time.Now()
			job.CompletedAt = &completedAt

			if err := q.store.UpdateJob(job); err != nil {
				logger.Error("failed to update job after panic", slog.String("error", err.Error()))
			}

				// No stage code to emit; snapshot above is best-effort.
			}

		// Remove from active jobs
		q.mu.Lock()
		delete(q.active, job.ID)
		q.mu.Unlock()
	}()

	// Update job status to running
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	job.Progress = 0
	job.Message = "Job started"

	if err := q.store.UpdateJob(job); err != nil {
		logger.Error("failed to update job status", slog.String("error", err.Error()))
	}

		// Single stage execution only
		if job.StageID == "" {
			q.handleJobError(job, fmt.Errorf("stage id is required for job"), logger)
			return
	}

	currentStageID = job.StageID // Track for panic recovery
	if err := q.executeSingleStage(ctx, job, logger); err != nil {
		q.handleJobError(job, err, logger)
		return
	}

		// If executeSingleStage already marked completion (e.g. skipped), don't overwrite.
		if job.Status != JobStatusCompleted {
			job.Status = JobStatusCompleted
			job.Progress = 100
			job.Message = "Job completed successfully"
			completedAt := time.Now()
			job.CompletedAt = &completedAt

			if err := q.store.UpdateJob(job); err != nil {
				logger.Error("failed to update job completion", slog.String("error", err.Error()))
			}
		}

		logger.Info("processing job completed")
	}

// executeSingleStage runs a single stage
func (q *JobQueue) executeSingleStage(ctx context.Context, job *Job, logger *slog.Logger) error {
	// Get the stage from registry using the exported method
	stage, err := q.manager.GetRegistry().Get(job.StageID)
	if err != nil {
		return fmt.Errorf("stage not found: %w", err)
	}

	logger.Debug("Checking if stage can run",
		slog.String("stage_id", job.StageID),
		slog.String("operation_id", job.OperationID),
		slog.String("request_id", middleware.GetReqID(ctx)))

	canRun := stage.CanRun()

	logger.Info("Stage CanRun check completed",
		slog.String("stage_id", job.StageID),
		slog.Bool("can_run", canRun),
		slog.String("request_id", middleware.GetReqID(ctx)))

	if !canRun {
		logger.Warn("stage requirements not met",
			slog.String("stage_id", job.StageID),
			slog.String("operation_id", job.OperationID))

		// Treat as a skip: stage cannot execute so it cannot emit its own snapshot.
		job.Status = JobStatusCompleted
		job.Message = fmt.Sprintf("Skipped %s: requirements not met", stage.Name())
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		_ = q.store.UpdateJob(job)

		broadcastJobQueueStageTerminalSnapshot(
			q.manager,
			logger,
			job.OperationID,
			stage.ID(),
			"skipped",
			"Requirements not met - required input data not available",
			nil,
		)

		return nil
	}

		job.Message = fmt.Sprintf("Executing %s", stage.Name())
		_ = q.store.UpdateJob(job)

	// Create operation state for the stage
	state := NewOperationState(job.OperationID)
	state.SetConfig(ContextKeyFromDate, job.Request.FromDate)
	state.SetConfig(ContextKeyToDate, job.Request.ToDate)

	// Initialize the stage state to prevent nil pointer dereference
	stepState := NewStepState(stage.ID(), stage.Name())
	state.SetStage(stage.ID(), stepState)

	// ✅ FIX A3: Attach broadcaster for consistent real-time progress updates
	// This ensures stepState's internal broadcaster is properly wired
	// (Note: broadcaster already declared on line 369, reusing same instance)
		// Deprecated: stage implementations broadcast via their own stage broadcasters.

	// ✅ HARDENING: Call stepState.Start() before invoking the processor so StartTime and status are always initialized
	stepState.Start()

	// Execute the stage
	logger.Info("executing stage", slog.String("stage", stage.ID()))

		if err := stage.Execute(ctx, state); err != nil {
			return stageExecutionFailedError{stageID: stage.ID(), err: err}
		}

	// Check if the stage was skipped or already handled its own completion
	stepState = state.GetStage(stage.ID())
	if stepState != nil && (stepState.Status == StepStatusSkipped || stepState.Status == StepStatusCompleted) {
		// If skipped, update job status but don't force 100% progress if it wasn't set
		if stepState.Status == StepStatusSkipped {
			job.Message = fmt.Sprintf("Skipped %s: %s", stage.Name(), stepState.Message)
			// Don't overwrite progress with 100 if skipped
		} else {
			job.Progress = 100
			job.Message = fmt.Sprintf("Completed %s", stage.Name())
		}
		// Explicitly mark as completed so processJob doesn't overwrite
		job.Status = JobStatusCompleted
		completedAt := time.Now()
		job.CompletedAt = &completedAt

		q.store.UpdateJob(job)
		return nil
	}

		job.Progress = 100
		job.Message = fmt.Sprintf("Completed %s", stage.Name())
		_ = q.store.UpdateJob(job)

	return nil
}

// handleJobError handles job execution errors
func (q *JobQueue) handleJobError(job *Job, err error, logger *slog.Logger) {
	logger.Error("job failed", slog.String("error", err.Error()))

	job.Status = JobStatusFailed
	job.Error = err.Error()
	job.Message = "Job failed"
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	if err := q.store.UpdateJob(job); err != nil {
		logger.Error("failed to update job error", slog.String("error", err.Error()))
	}

	// Only broadcast failures that occur before stage code can emit snapshots.
	var stageErr stageExecutionFailedError
	if !errors.As(err, &stageErr) {
		broadcastJobQueueStageTerminalSnapshot(
			q.manager,
			logger,
			job.OperationID,
			job.StageID,
			"failed",
			"Job failed",
			err,
		)
	}
}

// recoverJobs recovers jobs that were running when the system stopped
func (q *JobQueue) recoverJobs(ctx context.Context) {
	q.logger.Info("recovering pending and running jobs")

	// Find jobs that were running or pending
	jobs, err := q.store.ListJobs(JobFilter{
		Status: JobStatusRunning,
	})
	if err != nil {
		q.logger.Error("failed to recover running jobs", slog.String("error", err.Error()))
		return
	}

	pendingJobs, err := q.store.ListJobs(JobFilter{
		Status: JobStatusPending,
	})
	if err != nil {
		q.logger.Error("failed to recover pending jobs", slog.String("error", err.Error()))
	} else {
		jobs = append(jobs, pendingJobs...)
	}

	// Re-queue recovered jobs
	for _, job := range jobs {
		// Reset running jobs to pending
		if job.Status == JobStatusRunning {
			job.Status = JobStatusPending
			job.StartedAt = nil
			job.Progress = 0
			q.store.UpdateJob(job)
		}

		// Re-enqueue
		select {
		case q.jobs <- job:
			q.logger.Info("recovered job",
				slog.String("job_id", job.ID),
				slog.String("status", string(job.Status)))
		default:
			q.logger.Warn("could not recover job - queue full",
				slog.String("job_id", job.ID))
		}
	}
}

// GetQueueStats returns queue statistics
func (q *JobQueue) GetQueueStats() map[string]interface{} {
	q.mu.RLock()
	activeCount := len(q.active)
	q.mu.RUnlock()

	return map[string]interface{}{
		"workers":     q.workers,
		"queue_size":  len(q.jobs),
		"queue_cap":   cap(q.jobs),
		"active_jobs": activeCount,
	}
}

// UpdateJob updates a job in the store (public method for external access)
func (q *JobQueue) UpdateJob(job *Job) error {
	return q.store.UpdateJob(job)
}
