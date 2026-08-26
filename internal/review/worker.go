package review

import (
	"context"
	"fmt"

	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/quality"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type QuarantineSource interface {
	GetQuarantine(context.Context, string) (quality.Quarantine, error)
}

type Revalidator interface {
	Check(context.Context, variant.Variant) []quality.Finding
}

type Worker struct {
	reviews    *Service
	quarantine QuarantineSource
	checker    Revalidator
}

func NewWorker(reviews *Service, quarantine QuarantineSource, checker Revalidator) *Worker {
	return &Worker{reviews: reviews, quarantine: quarantine, checker: checker}
}

func (w *Worker) Process(ctx context.Context, id string) (Review, error) {
	item, err := w.reviews.Begin(ctx, id)
	if err != nil {
		return Review{}, err
	}
	quarantine, err := w.quarantine.GetQuarantine(ctx, item.QuarantineID)
	if err != nil {
		_, _ = w.reviews.Complete(ctx, id, false, "quarantine unavailable")
		return Review{}, fmt.Errorf("load quarantine: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_, _ = w.reviews.Complete(context.WithoutCancel(ctx), id, false, "review canceled")
		return Review{}, err
	}
	findings := w.checker.Check(ctx, quarantine.Variant)
	if len(findings) > 0 {
		return w.reviews.Complete(ctx, id, false, "quality findings remain")
	}
	review, err := w.reviews.Complete(ctx, id, true, "")
	if err != nil {
		return Review{}, err
	}
	if review.Status != StatusReleased {
		return Review{}, fmt.Errorf("review did not reach released: %w", platform.ErrInvalid)
	}
	return review, nil
}
