package normalization

import (
	"context"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"strings"
)

type Reference interface {
	Base(context.Context, string, int64) (byte, error)
}
type Change struct {
	Kind          string `json:"kind"`
	Before        string `json:"before"`
	After         string `json:"after"`
	PositionDelta int64  `json:"position_delta"`
}
type Result struct {
	Original   variant.Variant `json:"original"`
	Normalized variant.Variant `json:"normalized"`
	Changes    []Change        `json:"changes"`
}
type Service struct{ reference Reference }

func NewService(reference Reference) *Service { return &Service{reference: reference} }
func (s *Service) Normalize(ctx context.Context, v variant.Variant) (Result, error) {
	if err := variant.EnsureValid(v); err != nil {
		return Result{}, err
	}
	original := v
	v.Chromosome = variant.NormalizeChromosome(v.Chromosome)
	v.Reference = strings.ToUpper(v.Reference)
	v.Alternate = strings.ToUpper(v.Alternate)
	changes := []Change{}
	if original.Chromosome != v.Chromosome {
		changes = append(changes, Change{Kind: "chromosome", Before: original.Chromosome, After: v.Chromosome})
	}
	beforeRef, beforeAlt := v.Reference, v.Alternate
	v = trimSuffix(v)
	v = trimPrefix(v)
	if beforeRef != v.Reference || beforeAlt != v.Alternate {
		changes = append(changes, Change{Kind: "minimal_representation", Before: beforeRef + ">" + beforeAlt, After: v.Reference + ">" + v.Alternate, PositionDelta: v.Position - original.Position})
	}
	if len(v.Reference) != len(v.Alternate) && s.reference != nil {
		shifted, delta, err := s.leftAlign(ctx, v, 100)
		if err != nil {
			return Result{}, fmt.Errorf("left align: %w", err)
		}
		if delta > 0 {
			changes = append(changes, Change{Kind: "left_align", Before: v.Key(), After: shifted.Key(), PositionDelta: -delta})
			v = shifted
		}
	}
	return Result{Original: original, Normalized: v, Changes: changes}, nil
}
func trimSuffix(v variant.Variant) variant.Variant {
	for len(v.Reference) > 1 && len(v.Alternate) > 1 && v.Reference[len(v.Reference)-1] == v.Alternate[len(v.Alternate)-1] {
		v.Reference = v.Reference[:len(v.Reference)-1]
		v.Alternate = v.Alternate[:len(v.Alternate)-1]
	}
	return v
}
func trimPrefix(v variant.Variant) variant.Variant {
	for len(v.Reference) > 1 && len(v.Alternate) > 1 && v.Reference[0] == v.Alternate[0] {
		v.Reference = v.Reference[1:]
		v.Alternate = v.Alternate[1:]
		v.Position++
	}
	return v
}
func (s *Service) leftAlign(ctx context.Context, v variant.Variant, max int) (variant.Variant, int64, error) {
	shift := int64(0)
	for v.Position > 1 && shift < int64(max) {
		base, err := s.reference.Base(ctx, v.Chromosome, v.Position-1)
		if err != nil {
			return v, shift, err
		}
		last := v.Reference[len(v.Reference)-1]
		if len(v.Alternate) > len(v.Reference) {
			last = v.Alternate[len(v.Alternate)-1]
		}
		if byte(strings.ToUpper(string(base))[0]) != last {
			break
		}
		v.Position--
		v.Reference = string(base) + v.Reference[:len(v.Reference)-1]
		v.Alternate = string(base) + v.Alternate[:len(v.Alternate)-1]
		shift++
	}
	return v, shift, nil
}
func SplitMultiAllelic(v variant.Variant) ([]variant.Variant, error) {
	alts := strings.Split(v.Alternate, ",")
	if len(alts) == 0 {
		return nil, platform.ErrInvalid
	}
	out := make([]variant.Variant, 0, len(alts))
	for i, alt := range alts {
		copyVariant := v.Clone()
		copyVariant.Alternate = alt
		copyVariant.Info = copyInfo(v.Info, i, len(alts))
		if err := variant.EnsureValid(copyVariant); err != nil {
			return nil, fmt.Errorf("alternate %d: %w", i+1, err)
		}
		out = append(out, copyVariant)
	}
	return out, nil
}
func copyInfo(info map[string][]string, index, total int) map[string][]string {
	out := map[string][]string{}
	for key, values := range info {
		switch len(values) {
		case total:
			out[key] = []string{values[index]}
		default:
			out[key] = append([]string(nil), values...)
		}
	}
	return out
}
