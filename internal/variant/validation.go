package variant

import (
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"regexp"
	"strings"
)

var allelePattern = regexp.MustCompile(`^[ACGTNacgtn*<>.\[\]:_-]+$`)

type Issue struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Validate(v Variant) []Issue {
	issues := []Issue{}
	if NormalizeChromosome(v.Chromosome) == "" {
		issues = append(issues, Issue{"chromosome", "required", "chromosome is required"})
	}
	if v.Position < 1 {
		issues = append(issues, Issue{"position", "range", "position must be positive"})
	}
	if strings.TrimSpace(v.Reference) == "" || !allelePattern.MatchString(v.Reference) {
		issues = append(issues, Issue{"reference", "allele", "reference allele is invalid"})
	}
	if strings.TrimSpace(v.Alternate) == "" || !allelePattern.MatchString(v.Alternate) {
		issues = append(issues, Issue{"alternate", "allele", "alternate allele is invalid"})
	}
	if strings.EqualFold(v.Reference, v.Alternate) {
		issues = append(issues, Issue{"alternate", "unchanged", "alternate must differ from reference"})
	}
	if len(v.Reference) > 100000 || len(v.Alternate) > 100000 {
		issues = append(issues, Issue{"allele", "length", "allele exceeds maximum length"})
	}
	return issues
}
func EnsureValid(v Variant) error {
	issues := Validate(v)
	if len(issues) > 0 {
		return fmt.Errorf("%s: %w", issues[0].Message, platform.ErrInvalid)
	}
	return nil
}
