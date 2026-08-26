package reference

import (
	"context"
	"errors"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type PublicationCommitter interface {
	CommitPublication(context.Context, string, []byte) error
}

type PublicationService struct {
	source    *PublicationSource
	validator *PublicationValidator
	committer PublicationCommitter
}

func NewPublicationService(source *PublicationSource, validator *PublicationValidator, committer PublicationCommitter) *PublicationService {
	return &PublicationService{source: source, validator: validator, committer: committer}
}

func (s *PublicationService) Publish(ctx context.Context, key string) error {
	data, err := s.source.Load(ctx, key)
	if err != nil {
		return err
	}
	if err := s.validator.Validate(data); err != nil {
		if errors.Is(err, platform.ErrInvalid) {
			return err
		}
	}
	return s.committer.CommitPublication(ctx, key, data)
}
