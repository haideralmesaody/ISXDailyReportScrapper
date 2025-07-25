package handlers

import (
	"isxcli/internal/services"
)

// PipelineServiceAdapter adapts the services.PipelineService to PipelineServiceInterface
type PipelineServiceAdapter struct {
	service *services.PipelineService
}

// NewPipelineServiceAdapter creates a new adapter
func NewPipelineServiceAdapter(service *services.PipelineService) PipelineServiceInterface {
	return &PipelineServiceAdapter{
		service: service,
	}
}

// StartPipeline starts a new pipeline
func (a *PipelineServiceAdapter) StartPipeline(params map[string]interface{}) (string, error) {
	return a.service.StartPipeline(params)
}

// StopPipeline stops a pipeline
func (a *PipelineServiceAdapter) StopPipeline(pipelineID string) error {
	return a.service.StopPipeline(pipelineID)
}

// GetStatus gets the status of a pipeline
func (a *PipelineServiceAdapter) GetStatus(pipelineID string) (map[string]interface{}, error) {
	state, err := a.service.GetStatus(pipelineID)
	if err != nil {
		return nil, err
	}
	
	if state == nil {
		return nil, nil
	}
	
	// Convert PipelineState to map
	result := map[string]interface{}{
		"id":         state.ID,
		"status":     state.Status,
		"start_time": state.StartTime,
		"stages":     state.Stages,
		"context":    state.Context,
	}
	
	if state.EndTime != nil {
		result["end_time"] = *state.EndTime
	}
	
	// Calculate overall progress from stages
	if len(state.Stages) > 0 {
		totalProgress := 0.0
		for _, stage := range state.Stages {
			totalProgress += stage.Progress
		}
		result["progress"] = int(totalProgress / float64(len(state.Stages)))
	}
	
	return result, nil
}

// ListPipelines lists all pipelines
func (a *PipelineServiceAdapter) ListPipelines() []map[string]interface{} {
	pipelines := a.service.ListPipelines()
	result := make([]map[string]interface{}, 0, len(pipelines))
	
	for _, p := range pipelines {
		pipelineMap := map[string]interface{}{
			"id":         p.ID,
			"status":     p.Status,
			"start_time": p.StartTime,
			"stages":     p.Stages,
			"context":    p.Context,
		}
		
		if p.EndTime != nil {
			pipelineMap["end_time"] = *p.EndTime
		}
		
		// Calculate overall progress from stages
		if len(p.Stages) > 0 {
			totalProgress := 0.0
			for _, stage := range p.Stages {
				totalProgress += stage.Progress
			}
			pipelineMap["progress"] = int(totalProgress / float64(len(p.Stages)))
		}
		
		result = append(result, pipelineMap)
	}
	
	return result
}