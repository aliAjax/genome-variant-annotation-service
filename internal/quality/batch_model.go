package quality

import "github.com/example/genome-variant-annotation/internal/variant"

type BatchStage struct {
	Name       string
	MinQuality float64
}

type BatchPlan struct {
	Size   int
	Stages []BatchStage
}

type BatchFinding struct {
	VariantKey string
	Stage      string
	Findings   []Finding
}

func (p BatchPlan) Split(values []variant.Variant) ([][]variant.Variant, error) {
	if p.Size <= 0 {
		return nil, nil
	}
	chunks := make([][]variant.Variant, 0, (len(values)+p.Size-1)/p.Size)
	for start := 0; start < len(values); start += p.Size {
		end := start + p.Size
		chunks = append(chunks, values[start:end])
	}
	return chunks, nil
}
