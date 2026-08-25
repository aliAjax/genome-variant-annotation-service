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
