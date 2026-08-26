package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"sync"
	"time"
)

type Chunk struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Index     int       `json:"index"`
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
type Store interface {
	Put(context.Context, string, int, []byte) (Chunk, error)
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}
type MemoryStore struct {
	mu       sync.RWMutex
	metadata map[string]Chunk
	data     map[string][]byte
	clock    platform.Clock
}

func NewMemoryStore(clock platform.Clock) *MemoryStore {
	return &MemoryStore{metadata: map[string]Chunk{}, data: map[string][]byte{}, clock: clock}
}
func (s *MemoryStore) Put(_ context.Context, job string, index int, data []byte) (Chunk, error) {
	if job == "" || index < 0 || len(data) == 0 {
		return Chunk{}, platform.ErrInvalid
	}
	digest := sha256.Sum256(data)
	chunk := Chunk{ID: platform.NewID("chunk"), JobID: job, Index: index, Digest: fmt.Sprintf("sha256:%x", digest), Size: int64(len(data)), CreatedAt: s.clock.Now()}
	s.mu.Lock()
	s.metadata[chunk.ID] = chunk
	s.data[chunk.ID] = append([]byte(nil), data...)
	s.mu.Unlock()
	return chunk, nil
}
func (s *MemoryStore) Get(_ context.Context, id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.data[id]
	if !ok {
		return nil, platform.ErrNotFound
	}
	return append([]byte(nil), data...), nil
}
func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return platform.ErrNotFound
	}
	delete(s.data, id)
	delete(s.metadata, id)
	return nil
}
