package normalization

import "context"

type SequenceProvider interface {
	Base(context.Context, string, int64) (byte, error)
}
type Cache interface {
	Get(string) (Result, bool)
	Put(string, Result)
}
