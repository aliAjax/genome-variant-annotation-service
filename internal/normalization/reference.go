package normalization

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"strings"
	"sync"
)

type MemoryReference struct {
	mu        sync.RWMutex
	sequences map[string]string
}

func NewMemoryReference() *MemoryReference { return &MemoryReference{sequences: map[string]string{}} }
func (r *MemoryReference) Set(chromosome, sequence string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sequences[strings.ToUpper(chromosome)] = strings.ToUpper(sequence)
}
func (r *MemoryReference) Base(_ context.Context, chromosome string, position int64) (byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sequence, ok := r.sequences[strings.ToUpper(chromosome)]
	if !ok {
		return 0, fmt.Errorf("chromosome %s: %w", chromosome, platform.ErrNotFound)
	}
	if position < 1 || position > int64(len(sequence)) {
		return 0, fmt.Errorf("position %d: %w", position, platform.ErrInvalid)
	}
	return sequence[position-1], nil
}
