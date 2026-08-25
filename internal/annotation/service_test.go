package annotation

import (
	"context"
	"testing"

	"github.com/example/genome-variant-annotation/internal/normalization"
	"github.com/example/genome-variant-annotation/internal/reference"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type snapshotQuery struct {
	features []reference.Feature
}

func (s *snapshotQuery) Query(_ context.Context, _, _ string, _, _ int64) ([]reference.Feature, error) {
	return s.features, nil
}

func TestAnnotateUsesFeatureSnapshot(t *testing.T) {
	features := []reference.Feature{
		{Chromosome: "1", Start: 90, End: 150, Kind: "exon", ID: "e1", Gene: "GENE1"},
		{Chromosome: "1", Start: 10, End: 20, Kind: "exon", ID: "e2", Gene: "GENE2"},
	}
	query := &snapshotQuery{features: features}
	service := NewService(normalization.NewService(nil), query)
	result, err := service.Annotate(context.Background(), "d1", variant.Variant{Chromosome: "1", Position: 100, Reference: "A", Alternate: "G"})
	if err != nil {
		t.Fatal(err)
	}
	if features[0].Start != 90 || features[1].Start != 10 {
		t.Fatalf("annotate mutated the input feature slice: %+v", features)
	}
	if result.Evidence[0].RecordID != "e1" {
		t.Fatalf("unexpected evidence: %+v", result.Evidence)
	}
}
