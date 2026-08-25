package normalization

import (
	"context"
	"testing"

	"github.com/example/genome-variant-annotation/internal/variant"
)

func TestNormalizeMissingReferenceNoPanic(t *testing.T) {
	var ref *MemoryReference
	service := NewService(ref)
	_, err := service.Normalize(context.Background(), variant.Variant{Chromosome: "1", Position: 100, Reference: "A", Alternate: "AT"})
	if err != nil {
		t.Fatalf("expected typed-nil reference to be skipped without panic, got %v", err)
	}
}

func TestCachePutInitializesValue(t *testing.T) {
	cache := NewMemoryCache(2)
	cache.Put("k", Result{})
	if cache.Size() != 1 {
		t.Fatalf("expected cache size 1, got %d", cache.Size())
	}
}
