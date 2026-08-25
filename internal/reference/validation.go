package reference

import (
	"context"
	"sort"
)

type Conflict struct {
	LeftID     string `json:"left_id"`
	RightID    string `json:"right_id"`
	Chromosome string `json:"chromosome"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
	Reason     string `json:"reason"`
}

func DetectConflicts(features []Feature) []Conflict {
	items := append([]Feature(nil), features...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Chromosome == items[j].Chromosome {
			return items[i].Start < items[j].Start
		}
		return items[i].Chromosome < items[j].Chromosome
	})
	conflicts := make([]Conflict, 0)
	for i := 1; i < len(items); i++ {
		left, right := items[i-1], items[i]
		if left.Chromosome != right.Chromosome || !left.Overlaps(right.Start, right.End) {
			continue
		}
		if left.Kind == right.Kind && left.ID != right.ID {
			start := left.Start
			if right.Start > start {
				start = right.Start
			}
			end := left.End
			if right.End < end {
				end = right.End
			}
			conflicts = append(conflicts, Conflict{LeftID: left.ID, RightID: right.ID, Chromosome: left.Chromosome, Start: start, End: end, Reason: "overlapping records of the same kind"})
		}
	}
	return conflicts
}

func (s *Service) ValidateDataset(ctx context.Context, id string) ([]Conflict, error) {
	features, err := s.repo.Query(ctx, id, "1", 1, 1<<62)
	if err != nil {
		return nil, err
	}
	return DetectConflicts(features), nil
}
