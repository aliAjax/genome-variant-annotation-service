package review

import (
	"context"
	"fmt"

	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/quality"
)

type Service struct {
	repo  Repository
	clock platform.Clock
}

func NewService(repo Repository, clock platform.Clock) *Service {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) Open(ctx context.Context, item quality.Quarantine) (Review, error) {
	if item.ID == "" || item.DatasetID == "" {
		return Review{}, fmt.Errorf("open review: %w", platform.ErrInvalid)
	}
	now := s.clock.Now()
	review := Review{
		ID:           platform.NewID("review"),
		QuarantineID: item.ID,
		DatasetID:    item.DatasetID,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Save(ctx, review); err != nil {
		return Review{}, fmt.Errorf("save review: %w", err)
	}
	return review, nil
}

func (s *Service) Begin(ctx context.Context, id string) (Review, error) {
	return s.transition(ctx, id, StatusReviewing, "")
}

func (s *Service) Complete(ctx context.Context, id string, release bool, reason string) (Review, error) {
	next := StatusRejected
	if release {
		next = StatusReleased
	}
	return s.transition(ctx, id, next, reason)
}

func (s *Service) Retry(ctx context.Context, id string) (Review, error) {
	return s.transition(ctx, id, StatusPending, "")
}

func (s *Service) Get(ctx context.Context, id string) (Review, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) ListOpen(ctx context.Context, datasetID string) ([]Review, error) {
	return s.repo.ListOpen(ctx, datasetID)
}

func (s *Service) transition(ctx context.Context, id string, next Status, reason string) (Review, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return Review{}, err
	}
	if !item.CanTransition(next) {
		return Review{}, fmt.Errorf("review %s -> %s: %w", item.Status, next, platform.ErrInvalid)
	}
	item.Status = next
	item.Reason = reason
	item.UpdatedAt = s.clock.Now()
	if next == StatusReviewing {
		item.Attempts++
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return Review{}, fmt.Errorf("save review: %w", err)
	}
	return item, nil
}
