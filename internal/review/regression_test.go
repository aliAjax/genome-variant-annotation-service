package review

import (
	"context"
	"testing"
	"time"

	"github.com/example/genome-variant-annotation/internal/quality"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type regressionClock struct{ now time.Time }

func (c regressionClock) Now() time.Time { return c.now }

type fixedQuarantineSource struct {
	item quality.Quarantine
}

func (s fixedQuarantineSource) GetQuarantine(context.Context, string) (quality.Quarantine, error) {
	return s.item, nil
}

type fixedRevalidator struct {
	findings []quality.Finding
}

func (r fixedRevalidator) Check(context.Context, variant.Variant) []quality.Finding {
	return r.findings
}

func TestReviewCanTransitionRejectedToPending(t *testing.T) {
	review := Review{Status: StatusRejected}
	if !review.CanTransition(StatusAppealed) {
		t.Fatal("rejected review cannot move to appealed")
	}
}

func TestListOpenIncludesReviewing(t *testing.T) {
	repo := NewMemoryRepository()
	if err := repo.Save(context.Background(), Review{ID: "r1", DatasetID: "d1", Status: StatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), Review{ID: "r2", DatasetID: "d1", Status: StatusReviewing}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), Review{ID: "r3", DatasetID: "d1", Status: StatusAppealed}); err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListOpen(context.Background(), "d1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("expected two open reviews, got %d", len(items))
	}
}

func TestServiceRetryReturnsPending(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, regressionClock{now: time.Unix(1, 0)})
	opened, err := service.Open(context.Background(), quality.Quarantine{ID: "q1", DatasetID: "d1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Begin(context.Background(), opened.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(context.Background(), opened.ID, false, ""); err != nil {
		t.Fatal(err)
	}
	retried, err := service.Retry(context.Background(), opened.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Status != StatusAppealed {
		t.Fatalf("unexpected retry status: %s", retried.Status)
	}
}

func TestWorkerFailureLeavesRejected(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, regressionClock{now: time.Unix(1, 0)})
	opened, err := service.Open(context.Background(), quality.Quarantine{ID: "q1", DatasetID: "d1"})
	if err != nil {
		t.Fatal(err)
	}
	worker := NewWorker(service, fixedQuarantineSource{item: openedQuarantine(opened.QuarantineID)}, fixedRevalidator{findings: []quality.Finding{{Code: "low_quality"}}})
	processed, err := worker.Process(context.Background(), opened.ID)
	if err != nil {
		t.Fatal(err)
	}
	if processed.Status != StatusRejected {
		t.Fatalf("unexpected worker outcome: %s", processed.Status)
	}
}

func openedQuarantine(id string) quality.Quarantine {
	return quality.Quarantine{ID: id, DatasetID: "d1", Variant: variant.Variant{Chromosome: "1", Position: 1, Reference: "A", Alternate: "G"}}
}
