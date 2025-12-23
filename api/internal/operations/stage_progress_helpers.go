package operations

import "time"

// updateStageProgress normalizes metadata and broadcasts progress for non-scraping stages.
func updateStageProgress(
	stageID string,
	stepState *StepState,
	broadcaster StageBroadcaster,
	progress int,
	message string,
	status string,
	filesProcessed int,
	totalFiles int,
	failedFiles int,
	currentFile string,
) {
	if stepState == nil {
		return
	}

	clamped := clampProgress(progress)
	if status == "" {
		status = "running"
	}

	phase := derivePhaseByProgress(stageID, clamped, status)
	now := time.Now()

	stepState.mu.Lock()
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	md := stepState.Metadata

	md["stage_id"] = stageID
	md["target_stage_id"] = stageID
	md["operation_type"] = stageID
	md["progress_percent"] = clamped
	md["stage_message"] = message
	md["phase_message"] = message
	md["phase"] = string(phase)
	md["current_phase"] = string(phase)
	md["status"] = status
	md["files_processed"] = filesProcessed
	md["total_files"] = totalFiles
	md["failed_files"] = failedFiles
	md["current_file"] = currentFile
	md["telemetry_missing"] = false

	if stepState.StartTime != nil {
		md["stage_started_at"] = stepState.StartTime.Format(time.RFC3339)
	}
	md["stage_updated_at"] = now.Format(time.RFC3339)
	stepState.mu.Unlock()

	stepState.UpdateProgress(float64(clamped), message)

	if status == "completed" {
		stepState.Status = StepStatusCompleted
		stepState.EndTime = &now
	} else if status == "failed" {
		stepState.Status = StepStatusFailed
		stepState.EndTime = &now
	}

	if broadcaster != nil {
		broadcaster.UpdateProgressWithMetadata(clamped, message, cloneMetadata(md))
	}
}

// finalizeStageSuccess broadcasts a completion with final metadata.
func finalizeStageSuccess(
	stageID string,
	stepState *StepState,
	broadcaster StageBroadcaster,
	message string,
) {
	if stepState == nil || broadcaster == nil {
		return
	}
	stepState.mu.RLock()
	md := cloneMetadata(stepState.Metadata)
	stepState.mu.RUnlock()
	broadcaster.UpdateProgressWithMetadata(100, message, md)
	broadcaster.Complete(message)
}

// finalizeStageFailure broadcasts a failure with error metadata.
func finalizeStageFailure(
	stageID string,
	stepState *StepState,
	broadcaster StageBroadcaster,
	message string,
	err error,
) {
	if stepState != nil {
		stepState.mu.Lock()
		if stepState.Metadata == nil {
			stepState.Metadata = make(map[string]interface{})
		}
		stepState.Metadata["error"] = err.Error()
		stepState.Metadata["status"] = "failed"
		stepState.Metadata["stage_message"] = message
		stepState.Metadata["phase_message"] = message
		stepState.Metadata["phase"] = string(StagePhaseFailed)
		stepState.Metadata["current_phase"] = string(StagePhaseFailed)
		stepState.mu.Unlock()
	}
	if stepState != nil {
		stepState.UpdateProgress(0, message)
	}
	if broadcaster != nil {
		var md map[string]interface{}
		if stepState != nil {
			md = cloneMetadata(stepState.Metadata)
		}
		broadcaster.UpdateProgressWithMetadata(0, message, md)
		broadcaster.Fail(err)
	}
}

// derivePhaseByProgress maps progress/status to a generic phase for non-scraping stages.
func derivePhaseByProgress(stageID string, progress int, status string) StagePhase {
	if status == "failed" {
		return StagePhaseFailed
	}
	switch {
	case progress < 20:
		return StagePhaseStarting
	case progress < 85:
		return StagePhaseRunning
	case progress < 100:
		return StagePhaseCompleting
	default:
		return StagePhaseCompleted
	}
}
