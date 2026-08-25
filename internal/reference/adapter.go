package reference

import (
	"context"
	"io"
)

type ObjectStore interface {
	Open(context.Context, string) (io.ReadCloser, error)
	Put(context.Context, string, io.Reader) (string, error)
}
type Publisher interface {
	Publish(context.Context, string) (Dataset, error)
}
type FeatureQuery interface {
	Query(context.Context, string, string, int64, int64) ([]Feature, error)
}
