package variant

import (
	"context"
	"fmt"
)

type BatchResult struct {
	Accepted   []Variant      `json:"accepted"`
	Rejected   []BatchFailure `json:"rejected"`
	Duplicates int            `json:"duplicates"`
}
type BatchFailure struct {
	Index int    `json:"index"`
	Key   string `json:"key"`
	Error string `json:"error"`
}
type Writer interface {
	Put(context.Context, Variant) error
}

func WriteBatch(ctx context.Context, writer Writer, variants []Variant) BatchResult {
	// Use a fresh backing array for Accepted so the result never aliases the
	// caller's input slice; consecutive batches must not share storage.
	result := BatchResult{Accepted: make([]Variant, 0, len(variants)), Rejected: []BatchFailure{}}
	seen := map[string]struct{}{}
	for index, v := range variants {
		select {
		case <-ctx.Done():
			result.Rejected = append(result.Rejected, BatchFailure{Index: index, Key: v.Key(), Error: ctx.Err().Error()})
			continue
		default:
		}
		key := v.Key()
		if _, exists := seen[key]; exists {
			result.Duplicates++
			continue
		}
		seen[key] = struct{}{}
		// Clone before handing the variant to the writer and before storing it in
		// the result: the caller may keep mutating the input (e.g. reusing the
		// Info map across parsed batches), which would otherwise leak back into
		// the accepted records of this and earlier batches.
		clone := v.Clone()
		if err := writer.Put(ctx, clone); err != nil {
			result.Rejected = append(result.Rejected, BatchFailure{Index: index, Key: key, Error: fmt.Sprintf("%v", err)})
			continue
		}
		result.Accepted = append(result.Accepted, clone)
	}
	return result
}
