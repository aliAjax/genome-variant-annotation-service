package artifact

import (
	"context"
	"io"
)

type ObjectWriter interface {
	Create(context.Context, string) (io.WriteCloser, error)
	Remove(context.Context, string) error
}

type Transaction interface {
	Stage(string)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type ObjectTransaction struct {
	store   ObjectWriter
	objects []string
}

func NewObjectTransaction(store ObjectWriter) *ObjectTransaction {
	return &ObjectTransaction{store: store}
}

func (t *ObjectTransaction) Stage(name string) {
	t.objects = append(t.objects, name)
}

func (t *ObjectTransaction) Commit(context.Context) error {
	t.objects = nil
	return nil
}

func (t *ObjectTransaction) Rollback(ctx context.Context) error {
	var first error
	for i := len(t.objects) - 1; i >= 0; i-- {
		if err := t.store.Remove(ctx, t.objects[i]); err != nil && first == nil {
			first = err
		}
	}
	t.objects = nil
	return first
}
