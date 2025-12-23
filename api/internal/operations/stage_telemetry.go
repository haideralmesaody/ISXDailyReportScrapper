package operations

import (
	"math"
	"strings"
	"time"
)

// StagePhase represents the canonical set of stage phases shared across the UI and backend
type StagePhase string

const (
	StagePhasePending    StagePhase = "pending"
	StagePhaseStarting   StagePhase = "starting"
	StagePhasePreparing  StagePhase = "preparing"
	StagePhaseRunning    StagePhase = "running"
	StagePhaseValidating StagePhase = "validating"
	StagePhaseCompleting StagePhase = "completing"
	StagePhaseCompleted  StagePhase = "completed"
	StagePhaseComplete   StagePhase = "complete"
	StagePhaseFailed     StagePhase = "failed"
)

// StageTelemetry contains the SSOT fields that every stage update must emit
type StageTelemetry struct {
	StageID         string                  `json:"stage_id"`
	StageAlias      string                  `json:"stage_alias,omitempty"`
	Phase           StagePhase              `json:"phase"`
	PhaseMessage    string                  `json:"phase_message"`
	ProgressPercent float64                 `json:"progress_percent"`
	FilesProcessed  int                     `json:"files_processed"`
	TotalFiles      int                     `json:"total_files"`
	CurrentFile     string                  `json:"current_file,omitempty"`
	StageStartedAt  time.Time               `json:"stage_started_at"`
	StageUpdatedAt  time.Time               `json:"stage_updated_at"`
	FileStatuses    []*FileProcessingStatus `json:"file_statuses,omitempty"`
}

// StageTelemetryProvider is implemented by stage metric structs that can emit canonical telemetry
type StageTelemetryProvider interface {
	GetStageTelemetry() *StageTelemetry
}

// NewStageTelemetry constructs a telemetry payload and keeps metadata in sync with canonical keys
func NewStageTelemetry(stageID string, stepState *StepState, phase StagePhase, phaseMessage string, progress float64, filesProcessed, totalFiles int, currentFile string, fileStatuses []*FileProcessingStatus) *StageTelemetry {
	now := time.Now()

	metadata := ensureStepMetadata(stepState)
	stageAlias := ""
	if metadata != nil {
		if alias, ok := metadata["stage_alias"].(string); ok && alias != "" {
			stageAlias = alias
		} else if alias, ok := metadata["target_stage_id"].(string); ok && alias != "" {
			stageAlias = alias
		}
	}

	stageStartedAt := now
	if stepState != nil && stepState.StartTime != nil {
		stageStartedAt = *stepState.StartTime
	}

	return &StageTelemetry{
		StageID:         stageID,
		StageAlias:      stageAlias,
		Phase:           phase,
		PhaseMessage:    phaseMessage,
		ProgressPercent: clampProgressPercent(progress),
		FilesProcessed:  maxIntZero(filesProcessed),
		TotalFiles:      maxIntZero(totalFiles),
		CurrentFile:     currentFile,
		StageStartedAt:  stageStartedAt,
		StageUpdatedAt:  now,
		FileStatuses:    fileStatuses,
	}
}

// ApplyToMetadata copies telemetry fields into the metadata map that is broadcast to clients
func (t *StageTelemetry) ApplyToMetadata(metadata map[string]interface{}) {
	if t == nil || metadata == nil {
		return
	}

	metadata["stage_id"] = t.StageID
	if t.StageAlias != "" {
		metadata["stage_alias"] = t.StageAlias
	}

	if t.Phase != "" {
		metadata["phase"] = string(t.Phase)
		metadata["current_phase"] = string(t.Phase)
	}
	if t.PhaseMessage != "" {
		metadata["phase_message"] = t.PhaseMessage
		metadata["stage_message"] = t.PhaseMessage
	}

	metadata["progress_percent"] = t.ProgressPercent
	metadata["files_processed"] = t.FilesProcessed
	metadata["total_files"] = t.TotalFiles
	if t.CurrentFile != "" {
		metadata["current_file"] = t.CurrentFile
	}

	if !t.StageStartedAt.IsZero() {
		metadata["stage_started_at"] = t.StageStartedAt.Format(time.RFC3339)
	}
	if !t.StageUpdatedAt.IsZero() {
		metadata["stage_updated_at"] = t.StageUpdatedAt.Format(time.RFC3339)
	}

	if len(t.FileStatuses) > 0 {
		metadata["file_statuses"] = t.FileStatuses
	}
}

// ensureStepMetadata guarantees the StepState metadata map exists before we mutate it
func ensureStepMetadata(stepState *StepState) map[string]interface{} {
	if stepState == nil {
		return nil
	}
	if stepState.Metadata == nil {
		stepState.Metadata = make(map[string]interface{})
	}
	return stepState.Metadata
}

// setPhaseOnStep updates all canonical + legacy keys for a given StepState
func setPhaseOnStep(stepState *StepState, phase StagePhase, message string) {
	metadata := ensureStepMetadata(stepState)
	applyPhaseMetadata(metadata, phase, message)
}

// applyPhaseMetadata applies canonical/legacy phase keys on a metadata map
func applyPhaseMetadata(metadata map[string]interface{}, phase StagePhase, message string) {
	if metadata == nil {
		return
	}

	if phase != "" {
		metadata["phase"] = string(phase)
		metadata["current_phase"] = string(phase)
	}
	if message != "" {
		metadata["phase_message"] = message
		metadata["stage_message"] = message
	}
}

// derivePhaseFromMetadata pulls the current phase + message from the StepState map and ensures defaults
func derivePhaseFromMetadata(stepState *StepState, fallbackPhase StagePhase, fallbackMessage string) (StagePhase, string) {
	metadata := ensureStepMetadata(stepState)
	phase := fallbackPhase
	if metadata != nil {
		if raw, ok := metadata["phase"].(string); ok && raw != "" {
			phase = StagePhase(strings.ToLower(raw))
		} else if raw, ok := metadata["current_phase"].(string); ok && raw != "" {
			phase = StagePhase(strings.ToLower(raw))
			metadata["phase"] = string(phase)
		} else if fallbackPhase != "" {
			metadata["phase"] = string(fallbackPhase)
			metadata["current_phase"] = string(fallbackPhase)
			phase = fallbackPhase
		}

		if msg, ok := metadata["phase_message"].(string); ok && msg != "" {
			fallbackMessage = msg
		} else if fallbackMessage != "" {
			metadata["phase_message"] = fallbackMessage
		}

		if fallbackMessage != "" {
			metadata["stage_message"] = fallbackMessage
		}
	}

	if phase == "" {
		phase = StagePhaseRunning
		if metadata != nil {
			metadata["phase"] = string(phase)
			metadata["current_phase"] = string(phase)
		}
	}

	if fallbackMessage == "" && metadata != nil {
		if msg, ok := metadata["stage_message"].(string); ok && msg != "" {
			fallbackMessage = msg
		}
	}

	return phase, fallbackMessage
}

func clampProgressPercent(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return math.Round(value*100) / 100
}

func maxIntZero(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
