package artifact

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/example/genome-variant-annotation/internal/variant"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type memoryObjectStore struct {
	mu      sync.Mutex
	objects map[string]*bytes.Buffer
}

func newMemoryObjectStore() *memoryObjectStore {
	return &memoryObjectStore{objects: map[string]*bytes.Buffer{}}
}

func (s *memoryObjectStore) Create(_ context.Context, name string) (io.WriteCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	buffer := &bytes.Buffer{}
	s.objects[name] = buffer
	return nopWriteCloser{Writer: buffer}, nil
}

func (s *memoryObjectStore) Remove(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, name)
	return nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

type lineEncoder struct{}

func (lineEncoder) Encode(w io.Writer, item variant.Variant) error {
	_, err := io.WriteString(w, item.Key()+"\n")
	return err
}

func TestArtifactExport(t *testing.T) {
	repo := NewMemoryRepository()
	store := newMemoryObjectStore()
	service := NewService(repo, store, lineEncoder{}, fixedClock{now: time.Unix(1, 0)})
	bundle, err := service.Create(context.Background(), "dataset-1")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err = service.Export(context.Background(), bundle.ID, [][]variant.Variant{{{
		Chromosome: "1", Position: 1, Reference: "A", Alternate: "G",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Status != StatusReady || len(bundle.Objects) != 1 {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}
}
