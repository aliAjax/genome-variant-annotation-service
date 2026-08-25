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
	j.Results = cloneResults(j.Results)
	j.Failures = append([]Failure(nil), j.Failures...)
	return j
}
func cloneResults(in []annotation.Result) []annotation.Result {
	if len(in) == 0 {
		return nil
	}
	out := make([]annotation.Result, len(in))
	for i, r := range in {
		out[i] = cloneResult(r)
	}
	return out
}
func cloneResult(r annotation.Result) annotation.Result {
	r.Genes = append([]string(nil), r.Genes...)
	r.Consequences = append([]annotation.Consequence(nil), r.Consequences...)
	r.ClinicalLabels = append([]string(nil), r.ClinicalLabels...)
	r.Warnings = append([]string(nil), r.Warnings...)
	if len(r.Frequencies) != 0 {
		freqs := make(map[string]float64, len(r.Frequencies))
		for k, v := range r.Frequencies {
			freqs[k] = v
		}
		r.Frequencies = freqs
	}
	if len(r.Evidence) != 0 {
		ev := make([]annotation.Evidence, len(r.Evidence))
		for i, e := range r.Evidence {
			ev[i] = e
			if len(e.Fields) != 0 {
				fields := make(map[string]string, len(e.Fields))
				for k, v := range e.Fields {
					fields[k] = v
				}
				ev[i].Fields = fields
			}
		}
		r.Evidence = ev
	}
	return r
}
