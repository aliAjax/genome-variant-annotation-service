package job

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/genome-variant-annotation/internal/annotation"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type okAnnotator struct{}

func (okAnnotator) Annotate(_ context.Context, _ string, v variant.Variant) (annotation.Result, error) {
	return annotation.Result{Variant: v, VariantKey: v.Key()}, nil
}

func TestJobProcessingCountsAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	repo := NewMemoryRepository()
	service := NewService(repo, okAnnotator{}, platform.SystemClock{}, 10)
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		j, err := service.Submit(ctx, "d1", []variant.Variant{{Chromosome: "1", Position: int64(i + 1), Reference: "A", Alternate: "G"}})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, j.ID)
	}
	service.Run(ctx, 0)
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		done := true
		for _, id := range ids {
			job, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if job.Status != StatusCompleted {
				done = false
			}
		}
		if done {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("jobs were not all completed")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestJobProcessingConcurrentWorkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	repo := NewMemoryRepository()
	service := NewService(repo, okAnnotator{}, platform.SystemClock{}, 10)
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		j, err := service.Submit(ctx, "d1", []variant.Variant{{Chromosome: "1", Position: int64(i + 1), Reference: "A", Alternate: "G"}})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, j.ID)
	}
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			service.worker(ctx)
		}()
	}
	close(start)
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		done := true
		for _, id := range ids {
			job, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if job.Status != StatusCompleted {
				done = false
			}
		}
		if done {
			workers.Wait()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("concurrent workers did not complete all jobs")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestJobRepositoryGetReturnsCopy(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	j := Job{ID: "j1", Input: []variant.Variant{{Chromosome: "1", Position: 10, Reference: "A", Alternate: "G"}}, Results: []annotation.Result{{VariantKey: "1:10:A:G"}}}
	if err := repo.Save(ctx, j); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "j1")
	if err != nil {
		t.Fatal(err)
	}
	got.Input[0].Position = 999
	got.Results[0].VariantKey = "changed"
	again, err := repo.Get(ctx, "j1")
	if err != nil {
		t.Fatal(err)
	}
	if again.Input[0].Position != 10 || again.Results[0].VariantKey != "1:10:A:G" {
		t.Fatalf("repository returned internal slices: %+v", again)
	}
}
