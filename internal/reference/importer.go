package reference

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type ImportSummary struct {
	DatasetID string        `json:"dataset_id"`
	Accepted  int           `json:"accepted"`
	Rejected  int           `json:"rejected"`
	Errors    []ImportError `json:"errors,omitempty"`
}
type ImportError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
}

func (s *Service) ImportJSONL(ctx context.Context, datasetID string, r io.Reader, maximum int) (ImportSummary, error) {
	if maximum <= 0 {
		maximum = 100000
	}
	summary := ImportSummary{DatasetID: datasetID, Errors: []ImportError{}}
	batch := make([]Feature, 0, 1000)
	scanner := bufio.NewScanner(io.LimitReader(r, 128<<20))
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	line := 0
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := s.AddFeatures(ctx, datasetID, batch); err != nil {
			return err
		}
		summary.Accepted += len(batch)
		batch = batch[:0]
		return nil
	}
	for scanner.Scan() {
		line++
		if summary.Accepted+summary.Rejected+len(batch) >= maximum {
			return summary, platform.ErrBudget
		}
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		default:
		}
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var feature Feature
		if err := json.Unmarshal(scanner.Bytes(), &feature); err != nil {
			summary.Rejected++
			summary.Errors = append(summary.Errors, ImportError{Line: line, Message: err.Error()})
			continue
		}
		if feature.Chromosome == "" || feature.Start < 1 || feature.End < feature.Start {
			summary.Rejected++
			summary.Errors = append(summary.Errors, ImportError{Line: line, Message: "invalid interval"})
			continue
		}
		batch = append(batch, feature)
		if len(batch) == cap(batch) {
			if err := flush(); err != nil {
				return summary, fmt.Errorf("flush feature batch: %w", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return summary, fmt.Errorf("scan features: %w", err)
	}
	if err := flush(); err != nil {
		return summary, fmt.Errorf("flush final batch: %w", err)
	}
	return summary, nil
}
