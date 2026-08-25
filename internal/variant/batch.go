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
	result := BatchResult{Accepted: []Variant{}, Rejected: []BatchFailure{}}
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
		if err := writer.Put(ctx, v); err != nil {
			result.Rejected = append(result.Rejected, BatchFailure{Index: index, Key: key, Error: fmt.Sprintf("%v", err)})
			continue
		}
		result.Accepted = append(result.Accepted, v)
	}
	return result
}
