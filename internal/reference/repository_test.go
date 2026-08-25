package reference

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConcurrentReferenceQueryStable(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	if err := repo.Save(ctx, Dataset{ID: "d1", Status: StatusPublished}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddFeatures(ctx, "d1", []Feature{
		{Chromosome: "1", Start: 10, End: 20, Kind: "exon", ID: "a", Gene: "G1"},
		{Chromosome: "1", Start: 100, End: 120, Kind: "exon", ID: "b", Gene: "G2"},
	}); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			got, err := repo.Query(ctx, "d1", "1", 10, 20)
			if err != nil {
				t.Errorf("query %d: %v", n, err)
				return
			}
			if len(got) != 1 || got[0].ID != "a" {
				t.Errorf("query %d: unexpected result %+v", n, got)
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestDraftDatasetQueryRejected(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	service := NewService(repo, systemClock{})
	dataset, err := service.Create(ctx, "draft", "GRCh38")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Query(ctx, dataset.ID, "1", 1, 10); err == nil {
		t.Fatal("expected draft query to be rejected")
	}
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }
