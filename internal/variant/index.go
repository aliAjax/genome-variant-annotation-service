package variant

import (
	"context"
	"sort"
	"sync"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type Index struct {
	mu         sync.RWMutex
	values     map[string]Variant
	chromosome map[string][]string
}

func NewIndex() *Index {
	return &Index{values: map[string]Variant{}, chromosome: map[string][]string{}}
}
func (i *Index) Put(_ context.Context, v Variant) error {
	if err := EnsureValid(v); err != nil {
		return err
	}
	key := v.Key()
	i.mu.Lock()
	defer i.mu.Unlock()
	if old, exists := i.values[key]; exists && old.Digest() != v.Digest() {
		return platform.ErrConflict
	}
	if _, exists := i.values[key]; !exists {
		i.chromosome[NormalizeChromosome(v.Chromosome)] = append(i.chromosome[NormalizeChromosome(v.Chromosome)], key)
	}
	i.values[key] = v
	return nil
}
func (i *Index) Get(_ context.Context, key string) (Variant, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	v, ok := i.values[key]
	if !ok {
		return Variant{}, platform.ErrNotFound
	}
	return v, nil
}
func (i *Index) Region(_ context.Context, chromosome string, start, end int64) []Variant {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := []Variant{}
	for _, key := range i.chromosome[NormalizeChromosome(chromosome)] {
		v := i.values[key]
		if v.Position <= end && v.End() >= start {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Position < out[b].Position })
	return out
}
func (i *Index) Delete(_ context.Context, key string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.values[key]
	if !ok {
		return platform.ErrNotFound
	}
	delete(i.values, key)
	keys := i.chromosome[NormalizeChromosome(v.Chromosome)]
	for n, candidate := range keys {
		if candidate == key {
			i.chromosome[NormalizeChromosome(v.Chromosome)] = append(keys[:n], keys[n+1:]...)
			break
		}
	}
	return nil
}
