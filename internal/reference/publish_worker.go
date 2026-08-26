package reference

import (
	"context"
	"errors"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type PublicationWorker struct {
	service     *PublicationService
	maxAttempts int
}

func NewPublicationWorker(service *PublicationService, maxAttempts int) *PublicationWorker {
	return &PublicationWorker{service: service, maxAttempts: maxAttempts}
}

func (w *PublicationWorker) Run(ctx context.Context, key string) error {
	var last error
	for attempt := 0; attempt < w.maxAttempts; attempt++ {
		last = w.service.Publish(ctx, key)
		if last == nil {
			return nil
		}
		if errors.Is(last, platform.ErrInvalid) {
			continue
		}
		return last
	}
	return last
}
