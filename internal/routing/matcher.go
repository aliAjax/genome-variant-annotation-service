package routing

import (
	"reflect"
	"strings"

	"github.com/example/genome-variant-annotation/internal/annotation"
)

type Matcher interface {
	Match(annotation.Result) bool
	RuleID() string
}

type ruleMatcher struct {
	rule Rule
}

func Compile(rule *Rule) Matcher {
	if rule == nil {
		return nil
	}
	return &ruleMatcher{rule: *rule}
}

func (m *ruleMatcher) Match(result annotation.Result) bool {
	if m.rule.ClinicalLabel != "" && !containsFold(result.ClinicalLabels, m.rule.ClinicalLabel) {
		return false
	}
	if m.rule.MinimumImpact != "" && !hasImpact(result.Consequences, m.rule.MinimumImpact) {
		return false
	}
	return true
}

func (m *ruleMatcher) RuleID() string {
	return m.rule.ID
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func containsFold(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(value, expected) {
			return true
		}
	}
	return false
}

func hasImpact(values []annotation.Consequence, minimum string) bool {
	minimumRank := impactRank(minimum)
	for _, value := range values {
		if impactRank(value.Impact) >= minimumRank {
			return true
		}
	}
	return false
}

func impactRank(impact string) int {
	switch strings.ToLower(impact) {
	case "high":
		return 4
	case "moderate":
		return 3
	case "low":
		return 2
	case "modifier":
		return 1
	default:
		return 0
	}
}
