package review

import (
	"context"
	"testing"
	"time"

	"github.com/example/genome-variant-annotation/internal/quality"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestReviewLifecycle(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, fixedClock{now: time.Unix(1, 0)})
	opened, err := service.Open(context.Background(), quality.Quarantine{ID: "q1", DatasetID: "d1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Begin(context.Background(), opened.ID); err != nil {
		t.Fatal(err)
	}
	released, err := service.Complete(context.Background(), opened.ID, true, "")
	if err != nil {
		t.Fatal(err)
	}
	if released.Status != StatusReleased || !released.Terminal() {
		t.Fatalf("unexpected release state: %+v", released)
	}
}
