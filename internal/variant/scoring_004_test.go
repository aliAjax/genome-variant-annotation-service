package variant

import (
	"context"
	"testing"
)

func TestIndexPutNoPanic(t *testing.T) {
	index := NewIndex()
	if err := index.Put(context.Background(), Variant{Chromosome: "1", Position: 100, Reference: "A", Alternate: "G"}); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Get(context.Background(), "1:100:A:G"); err != nil {
		t.Fatal(err)
	}
}
