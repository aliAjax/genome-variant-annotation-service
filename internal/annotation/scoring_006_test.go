package annotation

import "testing"

func TestFrequencySummaryCopiesValues(t *testing.T) {
	input := map[string]float64{"global": 0.01}
	summary := SummarizeFrequencies(input, 0.05)
	input["global"] = 0.99
	if summary.Values["global"] != 0.01 {
		t.Fatalf("summary aliases input values: %+v", summary.Values)
	}
}
