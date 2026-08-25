package storage

import (
	"context"
	"testing"

	"github.com/example/genome-variant-annotation/internal/platform"
)

func TestChunkStorePutCopiesInput(t *testing.T) {
	store := NewMemoryStore(platform.SystemClock{})
	data := []byte("first")
	chunk, err := store.Put(context.Background(), "job-1", 0, data)
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'X'
	got, err := store.Get(context.Background(), chunk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first" {
		t.Fatalf("stored chunk was aliased: %q", got)
	}
}

func TestChunkStoreGetReturnsCopy(t *testing.T) {
	store := NewMemoryStore(platform.SystemClock{})
	chunk, err := store.Put(context.Background(), "job-1", 0, []byte("first"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(context.Background(), chunk.ID)
	if err != nil {
		t.Fatal(err)
	}
	got[0] = 'X'
	again, err := store.Get(context.Background(), chunk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != "first" {
		t.Fatalf("stored chunk was mutated through getter: %q", again)
	}
}
