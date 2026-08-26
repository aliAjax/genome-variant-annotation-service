package coordinate

import (
	"sync"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type PlanRepository struct {
	mu    sync.RWMutex
	plans map[string]Plan
}

func NewPlanRepository() *PlanRepository {
	return &PlanRepository{plans: map[string]Plan{}}
}

func (r *PlanRepository) Save(plan Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plans[plan.ID] = plan
	return nil
}

func (r *PlanRepository) Get(id string) (Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plan, ok := r.plans[id]
	if !ok {
		return Plan{}, platform.ErrNotFound
	}
	return plan, nil
}

func (r *PlanRepository) SaveCheckpoint(id string, next int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	plan, ok := r.plans[id]
	if !ok {
		return platform.ErrNotFound
	}
	plan.NextChunk = next + 1
	r.plans[id] = plan
	return nil
}
