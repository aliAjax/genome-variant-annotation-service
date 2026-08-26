package storage

import (
	"context"
	"sync"
)

type ManifestRepository interface {
	Save(context.Context, *Manifest) error
	Get(context.Context, string) (*Manifest, error)
}

type MemoryManifestRepository struct {
	mu    sync.RWMutex
	items map[string]*Manifest
}

func NewMemoryManifestRepository() *MemoryManifestRepository {
	return &MemoryManifestRepository{items: map[string]*Manifest{}}
}

func (r *MemoryManifestRepository) Save(_ context.Context, manifest *Manifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[manifest.ID] = manifest
	return nil
}

func (r *MemoryManifestRepository) Get(_ context.Context, id string) (*Manifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	manifest, ok := r.items[id]
	if !ok {
		return nil, nil
	}
	return manifest, nil
}
