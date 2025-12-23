package operations

import (
	"fmt"
	"sync"
)

// Registry manages registered operation steps
type Registry struct {
	mu    sync.RWMutex
	steps map[string]Step
	order []string // Maintains registration order
}

// NewRegistry creates a new Step registry with dependency injection
func NewRegistry() *Registry {
	return &Registry{
		steps: make(map[string]Step),
		order: make([]string, 0),
	}
}

// Register adds a Step to the registry
func (r *Registry) Register(Step Step) error {
	if Step == nil {
		return fmt.Errorf("cannot register nil Step")
	}

	id := Step.ID()
	if id == "" {
		return fmt.Errorf("Step ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.steps[id]; exists {
		return fmt.Errorf("Step with ID %s already registered", id)
	}

	r.steps[id] = Step
	r.order = append(r.order, id)
	return nil
}

// Unregister removes a Step from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.steps[id]; !exists {
		return fmt.Errorf("Step with ID %s not found", id)
	}

	delete(r.steps, id)

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

// Get retrieves a Step by ID
func (r *Registry) Get(id string) (Step, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	Step, exists := r.steps[id]
	if !exists {
		return nil, fmt.Errorf("Step with ID %s not found", id)
	}

	return Step, nil
}

// Has checks if a Step is registered
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.steps[id]
	return exists
}

// List returns all registered steps in registration order
func (r *Registry) List() []Step {
	r.mu.RLock()
	defer r.mu.RUnlock()

	steps := make([]Step, 0, len(r.order))
	for _, id := range r.order {
		if Step, exists := r.steps[id]; exists {
			steps = append(steps, Step)
		}
	}

	return steps
}

// Size returns the number of registered steps
func (r *Registry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.steps)
}

// ListIDs returns all registered Step IDs in registration order
func (r *Registry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, len(r.order))
	copy(ids, r.order)
	return ids
}

// Count returns the number of registered steps
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.steps)
}

// Clear removes all registered steps
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.steps = make(map[string]Step)
	r.order = make([]string, 0)
}

// GetDependencyOrder returns steps ordered by dependencies
func (r *Registry) GetDependencyOrder() ([]Step, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for id := range r.steps {
		graph[id] = []string{}
		inDegree[id] = 0
	}

	// Build graph and calculate in-degrees
	for id, Step := range r.steps {
		deps := Step.GetDependencies()
		for _, dep := range deps {
			if _, exists := r.steps[dep]; !exists {
				return nil, fmt.Errorf("Step %s depends on non-existent Step %s", id, dep)
			}
			graph[dep] = append(graph[dep], id)
			inDegree[id]++
		}
	}

	// Topological sort using Kahn's algorithm
	// Use registration order for steps with same priority
	queue := make([]string, 0)
	for _, id := range r.order {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	ordered := make([]Step, 0, len(r.steps))
	processed := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		ordered = append(ordered, r.steps[current])
		processed++

		// Reduce in-degree for dependent steps
		// Collect newly available steps
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
	if processed != len(r.steps) {
		return nil, fmt.Errorf("dependency cycle detected")
	}

	return ordered, nil
}

// ValidateDependencies checks if all Step dependencies are satisfied
func (r *Registry) ValidateDependencies() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, Step := range r.steps {
		deps := Step.GetDependencies()
		for _, dep := range deps {
			if _, exists := r.steps[dep]; !exists {
				return fmt.Errorf("Step %s depends on non-existent Step %s", id, dep)
			}
		}
	}

	// Check for cycles
	_, err := r.GetDependencyOrder()
	return err
}

// GetDependents returns steps that depend on the given Step
func (r *Registry) GetDependents(stageID string) []Step {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dependents := make([]Step, 0)
	for _, Step := range r.steps {
		deps := Step.GetDependencies()
		for _, dep := range deps {
			if dep == stageID {
				dependents = append(dependents, Step)
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
		if Step, exists := r.steps[id]; exists {
			clone.steps[id] = Step
			clone.order = append(clone.order, id)
		}
	}

	return clone
}
