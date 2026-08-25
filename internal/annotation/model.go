package annotation

import "github.com/example/genome-variant-annotation/internal/variant"

type Evidence struct {
	Source   string            `json:"source"`
	Version  string            `json:"version"`
	RecordID string            `json:"record_id"`
	Fields   map[string]string `json:"fields,omitempty"`
}

type Consequence struct {
	Gene       string `json:"gene"`
	Transcript string `json:"transcript,omitempty"`
	Term       string `json:"term"`
	Impact     string `json:"impact"`
	Rank       int    `json:"rank"`
}

type Result struct {
	Variant        variant.Variant    `json:"variant"`
	VariantKey     string             `json:"variant_key"`
	DatasetID      string             `json:"dataset_id"`
	Genes          []string           `json:"genes"`
	Consequences   []Consequence      `json:"consequences"`
	Frequencies    map[string]float64 `json:"frequencies,omitempty"`
	ClinicalLabels []string           `json:"clinical_labels,omitempty"`
	Evidence       []Evidence         `json:"evidence"`
	Warnings       []string           `json:"warnings,omitempty"`
}

// Clone returns a deep copy of r so that stored results cannot be mutated
// through aliases shared with the caller's input or with later batches. Every
// reference-typed field (the embedded Variant, slice/map fields) is copied.
func (r Result) Clone() Result {
	out := r
	out.Variant = r.Variant.Clone()
	out.Genes = cloneStrings(r.Genes)
	out.ClinicalLabels = cloneStrings(r.ClinicalLabels)
	out.Warnings = cloneStrings(r.Warnings)
	if r.Consequences != nil {
		out.Consequences = append([]Consequence(nil), r.Consequences...)
	}
	if r.Frequencies != nil {
		freqs := make(map[string]float64, len(r.Frequencies))
		for key, value := range r.Frequencies {
			freqs[key] = value
		}
		out.Frequencies = freqs
	}
	if r.Evidence != nil {
		out.Evidence = make([]Evidence, len(r.Evidence))
		for i, ev := range r.Evidence {
			out.Evidence[i] = ev
			if ev.Fields != nil {
				fields := make(map[string]string, len(ev.Fields))
				for key, value := range ev.Fields {
					fields[key] = value
				}
				out.Evidence[i].Fields = fields
			}
		}
	}
	return out
}

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
