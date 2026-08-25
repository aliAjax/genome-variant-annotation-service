package variant

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

type Variant struct {
	Chromosome string              `json:"chromosome"`
	Position   int64               `json:"position"`
	ID         string              `json:"id,omitempty"`
	Reference  string              `json:"reference"`
	Alternate  string              `json:"alternate"`
	Quality    *float64            `json:"quality,omitempty"`
	Filters    []string            `json:"filters,omitempty"`
	Info       map[string][]string `json:"info,omitempty"`
	SourceLine int                 `json:"source_line,omitempty"`
}

func (v Variant) Key() string {
	return fmt.Sprintf("%s:%d:%s:%s", NormalizeChromosome(v.Chromosome), v.Position, strings.ToUpper(v.Reference), strings.ToUpper(v.Alternate))
}
func (v Variant) Digest() string {
	h := sha256.Sum256([]byte(v.Key()))
	return fmt.Sprintf("sha256:%x", h[:])
}
func (v Variant) IsSNV() bool       { return len(v.Reference) == 1 && len(v.Alternate) == 1 }
func (v Variant) IsInsertion() bool { return len(v.Alternate) > len(v.Reference) }
func (v Variant) IsDeletion() bool  { return len(v.Reference) > len(v.Alternate) }
func (v Variant) End() int64        { return v.Position + int64(len(v.Reference)) - 1 }
func NormalizeChromosome(v string) string {
	v = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(v), "chr"))
	switch v {
	case "m":
		return "MT"
	case "x", "y", "mt":
		return strings.ToUpper(v)
	default:
		return v
	}
}
func (v Variant) SortedInfoKeys() []string {
	keys := make([]string, 0, len(v.Info))
	for k := range v.Info {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
