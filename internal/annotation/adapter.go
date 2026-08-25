package annotation

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Annotator interface {
	Annotate(context.Context, string, variant.Variant) (Result, error)
}
type ResultSink interface {
	Write(context.Context, string, []Result) error
}
type Notifier interface {
	Notify(context.Context, string, string) error
}
