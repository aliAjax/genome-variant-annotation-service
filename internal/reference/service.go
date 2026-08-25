package reference

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"strings"
)

type Service struct {
	repo  Repository
	clock platform.Clock
}

func NewService(repo Repository, clock platform.Clock) *Service {
	return &Service{repo: repo, clock: clock}
}
func (s *Service) Create(ctx context.Context, name, build string) (Dataset, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(build) == "" {
		return Dataset{}, platform.ErrInvalid
	}
	d := Dataset{ID: platform.NewID("dataset"), Name: name, Build: build, Version: 1, Status: StatusDraft, CreatedAt: s.clock.Now()}
	d.Digest = fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(name+"\x00"+build)))
	if err := s.repo.Save(ctx, d); err != nil {
		return Dataset{}, fmt.Errorf("save dataset: %w", err)
	}
	return d, nil
}
func (s *Service) AddFeatures(ctx context.Context, id string, features []Feature) error {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if d.Status != StatusDraft {
		return platform.ErrConflict
	}
	for i, f := range features {
		if f.Chromosome == "" || f.Start < 1 || f.End < f.Start {
			return fmt.Errorf("feature %d: %w", i, platform.ErrInvalid)
		}
	}
	return s.repo.AddFeatures(ctx, id, features)
}
func (s *Service) Publish(ctx context.Context, id string) (Dataset, error) {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		return Dataset{}, err
	}
	if d.Status != StatusDraft {
		return Dataset{}, platform.ErrConflict
	}
	if d.Records == 0 {
		return Dataset{}, fmt.Errorf("empty dataset: %w", platform.ErrInvalid)
	}
	now := s.clock.Now()
	d.Status = StatusPublished
	d.PublishedAt = &now
	d.Version++
	if err := s.repo.Save(ctx, d); err != nil {
		return Dataset{}, err
	}
	return d, nil
}
func (s *Service) Get(ctx context.Context, id string) (Dataset, error) { return s.repo.Get(ctx, id) }
func (s *Service) Query(ctx context.Context, id, chromosome string, start, end int64) ([]Feature, error) {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Status != StatusPublished {
		return nil, fmt.Errorf("dataset %s not published: %w", id, platform.ErrConflict)
	}
	return s.repo.Query(ctx, id, chromosome, start, end)
}
