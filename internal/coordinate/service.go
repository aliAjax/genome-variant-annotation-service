package coordinate

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"sort"
	"sync"
)

type Segment struct {
	SourceBuild string `json:"source_build"`
	TargetBuild string `json:"target_build"`
	Chromosome  string `json:"chromosome"`
	SourceStart int64  `json:"source_start"`
	SourceEnd   int64  `json:"source_end"`
	TargetStart int64  `json:"target_start"`
	Reverse     bool   `json:"reverse"`
}
type Result struct {
	Original  variant.Variant `json:"original"`
	Converted variant.Variant `json:"converted"`
	Segment   Segment         `json:"segment"`
}
type Service struct {
	mu       sync.RWMutex
	segments []Segment
}

func NewService() *Service { return &Service{segments: []Segment{}} }
func (s *Service) AddSegments(segments []Segment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, seg := range segments {
		if seg.SourceBuild == "" || seg.TargetBuild == "" || seg.SourceStart < 1 || seg.SourceEnd < seg.SourceStart || seg.TargetStart < 1 {
			return platform.ErrInvalid
		}
		s.segments = append(s.segments, seg)
	}
	sort.Slice(s.segments, func(i, j int) bool { return s.segments[i].SourceStart < s.segments[j].SourceStart })
	return nil
}
func (s *Service) Convert(ctx context.Context, source, target string, v variant.Variant) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, seg := range s.segments {
		if seg.SourceBuild != source || seg.TargetBuild != target || seg.Chromosome != v.Chromosome || v.Position < seg.SourceStart || v.End() > seg.SourceEnd {
			continue
		}
		converted := v
		offset := v.Position - seg.SourceStart
		if seg.Reverse {
			converted.Position = seg.TargetStart + (seg.SourceEnd - v.End())
		} else {
			converted.Position = seg.TargetStart + offset
		}
		return Result{Original: v, Converted: converted, Segment: seg}, nil
	}
	return Result{}, fmt.Errorf("coordinate mapping: %w", platform.ErrNotFound)
}
