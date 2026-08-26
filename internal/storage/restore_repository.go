package storage

import (
	"context"
	"sync"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type RestoreRepository interface {
	Save(context.Context, *RestoreSession) error
	Get(context.Context, string) (*RestoreSession, error)
}

type MemoryRestoreRepository struct {
	mu    sync.RWMutex
	items map[string]*RestoreSession
}

func NewMemoryRestoreRepository() *MemoryRestoreRepository {
	return &MemoryRestoreRepository{items: map[string]*RestoreSession{}}
}

func (r *MemoryRestoreRepository) Save(_ context.Context, session *RestoreSession) error {
	if session == nil || session.ID == "" {
		return platform.ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[session.ID] = cloneRestoreSession(session)
	return nil
}

func (r *MemoryRestoreRepository) Get(_ context.Context, id string) (*RestoreSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.items[id]
	if !ok {
		return nil, platform.ErrNotFound
	}
	return cloneRestoreSession(session), nil
}
