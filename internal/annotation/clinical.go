package annotation

import "sort"

type ClinicalAssertion struct {
	RecordID       string `json:"record_id"`
	Classification string `json:"classification"`
	ReviewStatus   string `json:"review_status"`
	Condition      string `json:"condition"`
	Updated        string `json:"updated"`
}
type ClinicalConsensus struct {
	Classification string              `json:"classification"`
	Conflicting    bool                `json:"conflicting"`
	Assertions     []ClinicalAssertion `json:"assertions"`
}

var clinicalRank = map[string]int{"pathogenic": 5, "likely_pathogenic": 4, "uncertain_significance": 3, "likely_benign": 2, "benign": 1}

func BuildClinicalConsensus(assertions []ClinicalAssertion) ClinicalConsensus {
	result := ClinicalConsensus{Assertions: append([]ClinicalAssertion(nil), assertions...)}
	seen := map[string]struct{}{}
	highest := ""
	rank := 0
	for _, assertion := range assertions {
		seen[assertion.Classification] = struct{}{}
		if clinicalRank[assertion.Classification] > rank {
			rank = clinicalRank[assertion.Classification]
			highest = assertion.Classification
		}
	}
	result.Classification = highest
	result.Conflicting = len(seen) > 1
	sort.Slice(result.Assertions, func(i, j int) bool {
		return clinicalRank[result.Assertions[i].Classification] > clinicalRank[result.Assertions[j].Classification]
	})
	return result
}
