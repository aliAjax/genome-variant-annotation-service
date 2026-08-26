package review

import (
	"context"
	"sync"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type Repository interface {
	Save(context.Context, Review) error
	Get(context.Context, string) (Review, error)
	ListOpen(context.Context, string) ([]Review, error)
}

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]Review
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]Review{}}
}

func (r *MemoryRepository) Save(_ context.Context, item Review) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return Review{}, platform.ErrNotFound
	}
	return item, nil
}

func (r *MemoryRepository) ListOpen(_ context.Context, datasetID string) ([]Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Review, 0)
	for _, item := range r.items {
		if datasetID != "" && item.DatasetID != datasetID {
			continue
		}
		if item.Status == StatusPending || item.Status == StatusReviewing {
			items = append(items, item)
		}
	}
	return items, nil
}
