package annotation

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/example/genome-variant-annotation/internal/normalization"
	"github.com/example/genome-variant-annotation/internal/reference"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type ReferenceQuery interface {
	Query(context.Context, string, string, int64, int64) ([]reference.Feature, error)
}

type Service struct {
	normalizer *normalization.Service
	reference  ReferenceQuery
	version    string
}

func NewService(n *normalization.Service, r ReferenceQuery) *Service {
	return &Service{normalizer: n, reference: r, version: "annotation-v1"}
}

func (s *Service) Annotate(ctx context.Context, datasetID string, input variant.Variant) (Result, error) {
	normalized, err := s.normalizer.Normalize(ctx, input)
	if err != nil {
		return Result{}, fmt.Errorf("normalize: %w", err)
	}
	// Clone the normalized variant so the Result owns its own Info/Filters/
	// Quality, independent of the input the caller passed (the normalizer
	// returns a struct copy that still aliases the input's Info map).
	v := normalized.Normalized.Clone()
	features, err := s.reference.Query(ctx, datasetID, v.Chromosome, v.Position, v.End())
	if err != nil {
		return Result{}, fmt.Errorf("query reference: %w", err)
	}
	result := Result{Variant: v, VariantKey: v.Key(), DatasetID: datasetID, Genes: []string{}, Consequences: []Consequence{}, Frequencies: map[string]float64{}, Evidence: []Evidence{}}
	genes := map[string]struct{}{}
	for _, feature := range features {
		if feature.Gene != "" {
			genes[feature.Gene] = struct{}{}
		}
		result.Evidence = append(result.Evidence, Evidence{Source: feature.Kind, Version: s.version, RecordID: feature.ID, Fields: cloneMap(feature.Attributes)})
		switch feature.Kind {
		case "exon":
			result.Consequences = append(result.Consequences, consequenceForCoding(v, feature))
		case "splice":
			result.Consequences = append(result.Consequences, Consequence{Gene: feature.Gene, Transcript: feature.Transcript, Term: "splice_region_variant", Impact: "MODERATE", Rank: 2})
		case "frequency":
			if raw := feature.Attributes["frequency"]; raw != "" {
				if value, parseErr := strconv.ParseFloat(raw, 64); parseErr == nil {
					result.Frequencies[feature.Attributes["population"]] = value
				}
			}
		case "clinical":
			if label := feature.Attributes["label"]; label != "" {
				result.ClinicalLabels = append(result.ClinicalLabels, label)
			}
		}
	}
	for gene := range genes {
		result.Genes = append(result.Genes, gene)
	}
	sort.Strings(result.Genes)
	sort.Strings(result.ClinicalLabels)
	if len(result.Consequences) == 0 {
		result.Consequences = append(result.Consequences, Consequence{Term: "intergenic_variant", Impact: "MODIFIER", Rank: 5})
	}
	sort.Slice(result.Consequences, func(i, j int) bool { return result.Consequences[i].Rank < result.Consequences[j].Rank })
	if len(normalized.Changes) > 0 {
		result.Warnings = append(result.Warnings, "variant was normalized before annotation")
	}
	return result, nil
}

func consequenceForCoding(v variant.Variant, f reference.Feature) Consequence {
	term, impact, rank := "coding_sequence_variant", "MODERATE", 3
	if v.IsSNV() {
		term, impact, rank = "missense_variant", "MODERATE", 2
	}
	if v.IsInsertion() || v.IsDeletion() {
		delta := len(v.Alternate) - len(v.Reference)
		if delta%3 != 0 {
			term, impact, rank = "frameshift_variant", "HIGH", 1
		} else {
			term, impact, rank = "inframe_indel", "MODERATE", 2
		}
	}
	return Consequence{Gene: f.Gene, Transcript: f.Transcript, Term: term, Impact: impact, Rank: rank}
}

func cloneMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[strings.TrimSpace(key)] = value
	}
	return out
}
