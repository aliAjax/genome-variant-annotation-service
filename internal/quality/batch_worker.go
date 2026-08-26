package quality

import (
	"context"

	"github.com/example/genome-variant-annotation/internal/variant"
)

type BatchWorker struct {
	service *BatchService
	buffer  []BatchFinding
}

func NewBatchWorker(service *BatchService) *BatchWorker {
	return &BatchWorker{service: service}
}

func (w *BatchWorker) Run(ctx context.Context, plan BatchPlan, values []variant.Variant) ([]BatchFinding, error) {
	results, err := w.service.Evaluate(ctx, plan, values)
	if err != nil {
		return nil, err
	}
	w.buffer = append(w.buffer, results...)
	return w.buffer, nil
}
