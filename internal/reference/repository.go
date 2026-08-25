package reference

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"sync"
)

type Repository interface {
	Save(context.Context, Dataset) error
	Get(context.Context, string) (Dataset, error)
	List(context.Context) ([]Dataset, error)
	AddFeatures(context.Context, string, []Feature) error
	Query(context.Context, string, string, int64, int64) ([]Feature, error)
}
type MemoryRepository struct {
	mu       sync.RWMutex
	datasets map[string]Dataset
	features map[string][]Feature
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{datasets: map[string]Dataset{}, features: map[string][]Feature{}}
}
func (r *MemoryRepository) Save(_ context.Context, d Dataset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.datasets[d.ID] = d
	return nil
}
func (r *MemoryRepository) Get(_ context.Context, id string) (Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.datasets[id]
	if !ok {
		return Dataset{}, fmt.Errorf("dataset %s: %w", id, platform.ErrNotFound)
	}
	return d, nil
}
func (r *MemoryRepository) List(_ context.Context) ([]Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Dataset, 0, len(r.datasets))
	for _, d := range r.datasets {
		out = append(out, d)
	}
	return out, nil
}
func (r *MemoryRepository) AddFeatures(_ context.Context, id string, features []Feature) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.datasets[id]; !ok {
		return platform.ErrNotFound
	}
	for _, f := range features {
		f.DatasetID = id
		f.Attributes = cloneStringMap(f.Attributes)
		r.features[id] = append(r.features[id], f)
	}
	d := r.datasets[id]
	d.Records = len(r.features[id])
	r.datasets[id] = d
	return nil
}
func (r *MemoryRepository) Query(_ context.Context, id, chromosome string, start, end int64) ([]Feature, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.datasets[id]; !ok {
		return nil, platform.ErrNotFound
	}
	out := []Feature{}
	for _, f := range r.features[id] {
		if f.Chromosome == chromosome && f.Overlaps(start, end) {
			f.Attributes = cloneStringMap(f.Attributes)
			out = append(out, f)
		}
	}
	return out, nil
}
func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
