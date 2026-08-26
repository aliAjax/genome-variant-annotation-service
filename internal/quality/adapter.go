package quality

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Checker interface {
	Check(context.Context, variant.Variant) []Finding
}
type QuarantineStore interface {
	Quarantine(context.Context, string, variant.Variant) (Quarantine, error)
	List(context.Context, string) []Quarantine
	Resolve(context.Context, string) error
}
