package artifact

import (
	"context"
	"errors"
)

type Lease interface {
	Acquire(context.Context, string) error
	Release(context.Context, string) error
}

type Worker struct {
	lease Lease
}

func NewWorker(lease Lease) *Worker {
	return &Worker{lease: lease}
}

func (w *Worker) WithLease(ctx context.Context, id string, run func(context.Context) error) (err error) {
	if err := w.lease.Acquire(ctx, id); err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.lease.Release(context.WithoutCancel(ctx), id))
	}()
	return run(ctx)
}
