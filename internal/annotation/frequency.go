package annotation

import (
	"fmt"
	"sort"
)

type FrequencySummary struct {
	Maximum    float64            `json:"maximum"`
	Population string             `json:"population"`
	Rare       bool               `json:"rare"`
	Values     map[string]float64 `json:"values"`
}

func SummarizeFrequencies(values map[string]float64, rareThreshold float64) FrequencySummary {
	result := FrequencySummary{Rare: true, Values: values}
	keys := make([]string, 0, len(values))
	for key, value := range values {
		keys = append(keys, key)
		result.Values[key] = value
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := values[key]
		if value > result.Maximum {
			result.Maximum = value
			result.Population = key
		}
		if value >= rareThreshold {
			result.Rare = false
		}
	}
	return result
}

func ValidateFrequency(value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("frequency must be between zero and one")
	}
	return nil
}

func MergeFrequencies(sets ...map[string]float64) map[string]float64 {
	result := map[string]float64{}
	for _, set := range sets {
		for population, value := range set {
			if current, exists := result[population]; !exists || value > current {
				result[population] = value
			}
		}
	}
	return result
}
