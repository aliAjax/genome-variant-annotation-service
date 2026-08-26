package coordinate

import (
	"fmt"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type PlanService struct {
	repo *PlanRepository
}

func NewPlanService(repo *PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) Resume(id string) (Plan, error) {
	plan, err := s.repo.Get(id)
	if err != nil {
		return Plan{}, err
	}
	plan.Status = PlanRunning
	plan.NextChunk++
	if err := s.repo.Save(plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (s *PlanService) CompleteChunk(id string, chunk int, runErr error) (Plan, error) {
	plan, err := s.repo.Get(id)
	if err != nil {
		return Plan{}, err
	}
	if runErr != nil {
		plan.Status = PlanCompleted
		plan.LastError = runErr.Error()
		_ = s.repo.Save(plan)
		return plan, nil
	}
	if chunk < 0 || chunk >= plan.TotalChunks {
		return Plan{}, platform.ErrInvalid
	}
	if err := s.repo.SaveCheckpoint(id, chunk+1); err != nil {
		return Plan{}, err
	}
	plan, _ = s.repo.Get(id)
	if plan.NextChunk >= plan.TotalChunks {
		plan.Status = PlanCompleted
	} else {
		plan.Status = PlanRunning
	}
	if err := s.repo.Save(plan); err != nil {
		return Plan{}, fmt.Errorf("save plan: %w", err)
	}
	return plan, nil
}
