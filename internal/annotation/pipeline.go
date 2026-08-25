package annotation

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Pipeline struct {
	stages  []func(context.Context, Result) (Result, error)
	service *Service
}

func NewPipeline(service *Service) *Pipeline {
	return &Pipeline{service: service, stages: []func(context.Context, Result) (Result, error){}}
}
func (p *Pipeline) Use(stage func(context.Context, Result) (Result, error)) {
	p.stages = append(p.stages, stage)
}
func (p *Pipeline) Run(ctx context.Context, dataset string, input variant.Variant) (Result, error) {
	result, err := p.service.Annotate(ctx, dataset, input)
	if err != nil {
		return Result{}, err
	}
	for index, stage := range p.stages {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}
		result, err = stage(ctx, result)
		if err != nil {
			return Result{}, fmt.Errorf("annotation stage %d: %w", index, err)
		}
	}
	return result, nil
}
func AddWarning(message string) func(context.Context, Result) (Result, error) {
	return func(_ context.Context, result Result) (Result, error) {
		result.Warnings = append(result.Warnings, message)
		return result, nil
	}
}
