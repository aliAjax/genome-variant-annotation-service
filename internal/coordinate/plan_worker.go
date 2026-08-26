package coordinate

import "context"

type ChunkRunner interface {
	RunChunk(context.Context, Plan, int) error
}

type PlanWorker struct {
	service *PlanService
	repo    *PlanRepository
	runner  ChunkRunner
}

func NewPlanWorker(service *PlanService, repo *PlanRepository, runner ChunkRunner) *PlanWorker {
	return &PlanWorker{service: service, repo: repo, runner: runner}
}

func (w *PlanWorker) Run(ctx context.Context, id string) (Plan, error) {
	plan, err := w.repo.Get(id)
	if err != nil {
		return Plan{}, err
	}
	if plan.Status != PlanRunning {
		plan, err = w.service.Resume(id)
		if err != nil {
			return Plan{}, err
		}
	}
	for chunk := plan.NextChunk; chunk < plan.TotalChunks; chunk++ {
		runErr := w.runner.RunChunk(ctx, plan, chunk)
		plan, err = w.service.CompleteChunk(id, chunk, runErr)
		if err != nil {
			return Plan{}, err
		}
	}
	return plan, nil
}
