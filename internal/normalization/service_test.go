package normalization

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/variant"
	"testing"
)

func TestMinimalRepresentation(t *testing.T) {
	service := NewService(nil)
	result, err := service.Normalize(context.Background(), variant.Variant{Chromosome: "chr1", Position: 100, Reference: "ATC", Alternate: "AGC"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Normalized.Key() != "1:101:T:G" {
		t.Fatalf("unexpected normalized variant: %s", result.Normalized.Key())
	}
}

func TestSplitMultiAllelic(t *testing.T) {
	values, err := SplitMultiAllelic(variant.Variant{Chromosome: "1", Position: 100, Reference: "A", Alternate: "G,T", Info: map[string][]string{"AF": {"0.1", "0.2"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[1].Info["AF"][0] != "0.2" {
		t.Fatalf("unexpected split: %+v", values)
	}
}
