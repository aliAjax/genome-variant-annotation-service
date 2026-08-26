package job

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/annotation"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"sync"
)

type Annotator interface {
	Annotate(context.Context, string, variant.Variant) (annotation.Result, error)
}
type Service struct {
	repo      Repository
	annotator Annotator
	clock     platform.Clock
	queue     chan string
	mu        sync.Mutex
}

func NewService(repo Repository, annotator Annotator, clock platform.Clock, queueSize int) *Service {
	return &Service{repo: repo, annotator: annotator, clock: clock, queue: make(chan string, queueSize)}
}
func (s *Service) Submit(ctx context.Context, datasetID string, input []variant.Variant) (Job, error) {
	if datasetID == "" || len(input) == 0 {
		return Job{}, platform.ErrInvalid
	}
	if len(input) > 100000 {
		return Job{}, platform.ErrBudget
	}
	j := Job{ID: platform.NewID("job"), DatasetID: datasetID, Status: StatusQueued, Input: append([]variant.Variant(nil), input...), Results: []annotation.Result{}, Failures: []Failure{}, CreatedAt: s.clock.Now()}
	if err := s.repo.Save(ctx, j); err != nil {
		return Job{}, fmt.Errorf("save job: %w", err)
	}
	select {
	case s.queue <- j.ID:
		return j, nil
	case <-ctx.Done():
		return Job{}, ctx.Err()
	}
}
func (s *Service) Get(ctx context.Context, id string) (Job, error) { return s.repo.Get(ctx, id) }
func (s *Service) Cancel(ctx context.Context, id string) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, err := s.repo.Get(ctx, id)
	if err != nil {
		return Job{}, err
	}
	if j.Status == StatusCompleted || j.Status == StatusFailed || j.Status == StatusCancelled {
		return j, nil
	}
	j.CancelRequested = true
	if j.Status == StatusQueued {
		j.Status = StatusCancelled
		now := s.clock.Now()
		j.FinishedAt = &now
	}
	if err := s.repo.Save(ctx, j); err != nil {
		return Job{}, err
	}
	return j, nil
}
func (s *Service) Run(ctx context.Context, workers int) {
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go s.worker(ctx)
	}
}
func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-s.queue:
			s.process(ctx, id)
		}
	}
}
func (s *Service) process(ctx context.Context, id string) {
	s.mu.Lock()
	j, err := s.repo.Get(ctx, id)
	if err != nil || j.Status != StatusQueued {
		s.mu.Unlock()
		return
	}
	now := s.clock.Now()
	j.Status = StatusRunning
	j.StartedAt = &now
	_ = s.repo.Save(ctx, j)
	s.mu.Unlock()
	for i, v := range j.Input {
		s.mu.Lock()
		current, _ := s.repo.Get(ctx, id)
		if current.CancelRequested {
			current.Status = StatusCancelled
			finished := s.clock.Now()
			current.FinishedAt = &finished
			_ = s.repo.Save(ctx, current)
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		result, err := s.annotator.Annotate(ctx, j.DatasetID, v)
		s.mu.Lock()
		current, _ = s.repo.Get(ctx, id)
		if err != nil {
			current.Failures = append(current.Failures, Failure{Index: i, VariantKey: v.Key(), Error: err.Error()})
		} else {
			current.Results = append(current.Results, result)
		}
		current.Progress = i + 1
		_ = s.repo.Save(ctx, current)
		s.mu.Unlock()
	}
	s.mu.Lock()
	current, _ := s.repo.Get(ctx, id)
	finished := s.clock.Now()
	current.FinishedAt = &finished
	switch {
	case len(current.Failures) == 0:
		current.Status = StatusCompleted
	case len(current.Results) == 0:
		current.Status = StatusFailed
	default:
		current.Status = StatusPartial
	}
	_ = s.repo.Save(ctx, current)
	s.mu.Unlock()
}
