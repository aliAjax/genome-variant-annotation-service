package coordinate

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Converter interface {
	Convert(context.Context, string, string, variant.Variant) (Result, error)
}
type SegmentLoader interface {
	Load(context.Context, string, string) ([]Segment, error)
}
