package delivery

import (
	"context"
	"fmt"
)

type Deliverer interface {
	Deliver(context.Context, Batch) error
}

type Worker struct {
	deliverer Deliverer
}

func NewWorker(deliverer Deliverer) *Worker {
	return &Worker{deliverer: deliverer}
}

func (w *Worker) Run(ctx context.Context, jobs <-chan Batch) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch, ok := <-jobs:
			if !ok {
				return nil
			}
			if err := w.deliverer.Deliver(ctx, batch); err != nil {
				return fmt.Errorf("deliver job: %v", err)
			}
		}
	}
}
