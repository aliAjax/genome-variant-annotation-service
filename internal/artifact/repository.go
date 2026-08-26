package artifact

import (
	"context"
	"sync"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type Repository interface {
	Save(context.Context, Bundle) error
	Get(context.Context, string) (Bundle, error)
}

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]Bundle
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]Bundle{}}
}

func (r *MemoryRepository) Save(_ context.Context, bundle Bundle) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bundle.Objects = append([]string(nil), bundle.Objects...)
	r.items[bundle.ID] = bundle
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (Bundle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bundle, ok := r.items[id]
	if !ok {
		return Bundle{}, platform.ErrNotFound
	}
	bundle.Objects = append([]string(nil), bundle.Objects...)
	return bundle, nil
}
