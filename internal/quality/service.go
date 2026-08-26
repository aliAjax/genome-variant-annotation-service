package quality

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"sync"
	"time"
)

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Finding struct {
	Code       string    `json:"code"`
	Severity   Severity  `json:"severity"`
	VariantKey string    `json:"variant_key"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
type Quarantine struct {
	ID         string          `json:"id"`
	DatasetID  string          `json:"dataset_id"`
	Variant    variant.Variant `json:"variant"`
	Findings   []Finding       `json:"findings"`
	RetryCount int             `json:"retry_count"`
	Resolved   bool            `json:"resolved"`
}
type Service struct {
	mu    sync.RWMutex
	items map[string]Quarantine
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: map[string]Quarantine{}, clock: clock}
}
func (s *Service) Check(_ context.Context, v variant.Variant) []Finding {
	result := []Finding{}
	for _, issue := range variant.Validate(v) {
		result = append(result, Finding{Code: issue.Code, Severity: SeverityError, VariantKey: v.Key(), Message: issue.Message, CreatedAt: s.clock.Now()})
	}
	if v.Quality != nil && *v.Quality < 20 {
		result = append(result, Finding{Code: "low_quality", Severity: SeverityWarning, VariantKey: v.Key(), Message: "quality below 20", CreatedAt: s.clock.Now()})
	}
	if len(v.Filters) > 0 && v.Filters[0] != "PASS" && v.Filters[0] != "." {
		result = append(result, Finding{Code: "filtered", Severity: SeverityWarning, VariantKey: v.Key(), Message: "input filter is not PASS", CreatedAt: s.clock.Now()})
	}
	return result
}
func (s *Service) Quarantine(ctx context.Context, dataset string, v variant.Variant) (Quarantine, error) {
	findings := s.Check(ctx, v)
	if len(findings) == 0 {
		return Quarantine{}, fmt.Errorf("variant has no quality findings: %w", platform.ErrInvalid)
	}
	item := Quarantine{ID: platform.NewID("quarantine"), DatasetID: dataset, Variant: v, Findings: findings}
	s.mu.Lock()
	s.items[item.ID] = item
	s.mu.Unlock()
	return item, nil
}
func (s *Service) List(_ context.Context, dataset string) []Quarantine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Quarantine{}
	for _, item := range s.items {
		if dataset == "" || item.DatasetID == dataset {
			out = append(out, item)
		}
	}
	return out
}
func (s *Service) Resolve(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return platform.ErrNotFound
	}
	item.Resolved = true
	s.items[id] = item
	return nil
}
