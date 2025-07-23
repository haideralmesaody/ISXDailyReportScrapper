package pipeline

import (
	"fmt"
	"sync"
)

// Registry manages registered pipeline stages
type Registry struct {
	mu     sync.RWMutex
	stages map[string]Stage
	order  []string // Maintains registration order
}

// NewRegistry creates a new stage registry
func NewRegistry() *Registry {
	return &Registry{
		stages: make(map[string]Stage),
		order:  make([]string, 0),
	}
}

// Register adds a stage to the registry
func (r *Registry) Register(stage Stage) error {
	if stage == nil {
		return fmt.Errorf("cannot register nil stage")
	}

	id := stage.ID()
	if id == "" {
		return fmt.Errorf("stage ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stages[id]; exists {
		return fmt.Errorf("stage with ID %s already registered", id)
	}

	r.stages[id] = stage
	r.order = append(r.order, id)
	return nil
}

// Unregister removes a stage from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.stages[id]; !exists {
		return fmt.Errorf("stage with ID %s not found", id)
	}

	delete(r.stages, id)

	// Remove from order slice
	newOrder := make([]string, 0, len(r.order)-1)
	for _, stageID := range r.order {
		if stageID != id {
			newOrder = append(newOrder, stageID)
		}
	}
	r.order = newOrder

	return nil
}

// Get retrieves a stage by ID
func (r *Registry) Get(id string) (Stage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stage, exists := r.stages[id]
	if !exists {
		return nil, fmt.Errorf("stage with ID %s not found", id)
	}

	return stage, nil
}

// Has checks if a stage is registered
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.stages[id]
	return exists
}

// List returns all registered stages in registration order
func (r *Registry) List() []Stage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stages := make([]Stage, 0, len(r.order))
	for _, id := range r.order {
		if stage, exists := r.stages[id]; exists {
			stages = append(stages, stage)
		}
	}

	return stages
}

// ListIDs returns all registered stage IDs in registration order
func (r *Registry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, len(r.order))
	copy(ids, r.order)
	return ids
}

// Count returns the number of registered stages
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.stages)
}

// Clear removes all registered stages
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stages = make(map[string]Stage)
	r.order = make([]string, 0)
}

// GetDependencyOrder returns stages ordered by dependencies
func (r *Registry) GetDependencyOrder() ([]Stage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	
	// Initialize
	for id := range r.stages {
		graph[id] = []string{}
		inDegree[id] = 0
	}
	
	// Build graph and calculate in-degrees
	for id, stage := range r.stages {
		deps := stage.GetDependencies()
		for _, dep := range deps {
			if _, exists := r.stages[dep]; !exists {
				return nil, fmt.Errorf("stage %s depends on non-existent stage %s", id, dep)
			}
			graph[dep] = append(graph[dep], id)
			inDegree[id]++
		}
	}
	
	// Topological sort using Kahn's algorithm
	// Use registration order for stages with same priority
	queue := make([]string, 0)
	for _, id := range r.order {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	
	ordered := make([]Stage, 0, len(r.stages))
	processed := 0
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		ordered = append(ordered, r.stages[current])
		processed++
		
		// Reduce in-degree for dependent stages
		// Collect newly available stages
		newAvailable := make([]string, 0)
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				newAvailable = append(newAvailable, dependent)
			}
		}
		
		// Sort newly available by registration order
		for _, id := range r.order {
			for _, available := range newAvailable {
				if id == available {
					queue = append(queue, id)
					break
				}
			}
		}
	}
	
	// Check for cycles
	if processed != len(r.stages) {
		return nil, fmt.Errorf("dependency cycle detected")
	}
	
	return ordered, nil
}

// ValidateDependencies checks if all stage dependencies are satisfied
func (r *Registry) ValidateDependencies() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, stage := range r.stages {
		deps := stage.GetDependencies()
		for _, dep := range deps {
			if _, exists := r.stages[dep]; !exists {
				return fmt.Errorf("stage %s depends on non-existent stage %s", id, dep)
			}
		}
	}

	// Check for cycles
	_, err := r.GetDependencyOrder()
	return err
}

// GetDependents returns stages that depend on the given stage
func (r *Registry) GetDependents(stageID string) []Stage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dependents := make([]Stage, 0)
	for _, stage := range r.stages {
		deps := stage.GetDependencies()
		for _, dep := range deps {
			if dep == stageID {
				dependents = append(dependents, stage)
				break
			}
		}
	}

	return dependents
}

// Clone creates a copy of the registry
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := NewRegistry()
	for _, id := range r.order {
		if stage, exists := r.stages[id]; exists {
			clone.stages[id] = stage
			clone.order = append(clone.order, id)
		}
	}

	return clone
}