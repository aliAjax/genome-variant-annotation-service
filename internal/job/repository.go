package job

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/annotation"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"sync"
)

type Repository interface {
	Save(context.Context, Job) error
	Get(context.Context, string) (Job, error)
	List(context.Context) ([]Job, error)
}

type MemoryRepository struct {
	mu     sync.RWMutex
	values map[string]Job
}

func NewMemoryRepository() *MemoryRepository { return &MemoryRepository{values: map[string]Job{}} }
func (r *MemoryRepository) Save(_ context.Context, j Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[j.ID] = cloneJob(j)
	return nil
}
func (r *MemoryRepository) Get(_ context.Context, id string) (Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.values[id]
	if !ok {
		return Job{}, fmt.Errorf("job %s: %w", id, platform.ErrNotFound)
	}
	return cloneJob(j), nil
}
func (r *MemoryRepository) List(_ context.Context) ([]Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Job, 0, len(r.values))
	for _, j := range r.values {
		out = append(out, cloneJob(j))
	}
	return out, nil
}
func cloneJob(j Job) Job {
	j.Input = append([]variant.Variant(nil), j.Input...)
	j.Results = append([]annotation.Result(nil), j.Results...)
	j.Failures = append([]Failure(nil), j.Failures...)
	return j
}
