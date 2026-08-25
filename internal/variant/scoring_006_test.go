package variant

import (
	"context"
	"testing"
)

type captureWriter struct {
	variants []Variant
}

func (w *captureWriter) Put(_ context.Context, v Variant) error {
	w.variants = append(w.variants, v)
	return nil
}

func TestWriteBatchDoesNotAliasInput(t *testing.T) {
	input := []Variant{{Chromosome: "1", Position: 10, Reference: "A", Alternate: "G", Info: map[string][]string{"AF": {"0.1"}}}}
	writer := &captureWriter{}
	result := WriteBatch(context.Background(), writer, input)
	input[0].Info["AF"][0] = "0.9"
	if result.Accepted[0].Info["AF"][0] != "0.1" {
		t.Fatalf("accepted variant aliases input: %+v", result.Accepted)
	}
}
