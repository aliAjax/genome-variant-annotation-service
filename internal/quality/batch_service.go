package quality

import (
	"context"

	"github.com/example/genome-variant-annotation/internal/variant"
)

type BatchService struct {
	checker Checker
	scratch []Finding
}

func NewBatchService(checker Checker) *BatchService {
	return &BatchService{checker: checker}
}

func (s *BatchService) Evaluate(ctx context.Context, plan BatchPlan, values []variant.Variant) ([]BatchFinding, error) {
	stage := plan.Stages[0]
	chunks, err := plan.Split(values)
	if err != nil {
		return nil, err
	}
	results := make([]BatchFinding, 0, len(values))
	for _, chunk := range chunks {
		for _, value := range chunk {
			s.scratch = append(s.scratch[:0], s.checker.Check(ctx, value)...)
			results = append(results, BatchFinding{VariantKey: value.Key(), Stage: stage.Name, Findings: s.scratch})
		}
	}
	return results, nil
}
