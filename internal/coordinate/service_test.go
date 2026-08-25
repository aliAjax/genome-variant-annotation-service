package coordinate

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/variant"
	"testing"
)

func TestCoordinateConversion(t *testing.T) {
	service := NewService()
	if err := service.AddSegments([]Segment{{SourceBuild: "GRCh37", TargetBuild: "GRCh38", Chromosome: "1", SourceStart: 100, SourceEnd: 200, TargetStart: 120}}); err != nil {
		t.Fatal(err)
	}
	result, err := service.Convert(context.Background(), "GRCh37", "GRCh38", variant.Variant{Chromosome: "1", Position: 150, Reference: "A", Alternate: "G"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Converted.Position != 170 {
		t.Fatalf("got %d", result.Converted.Position)
	}
}
