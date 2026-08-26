package routing

import (
	"context"
	"fmt"

	"github.com/example/genome-variant-annotation/internal/annotation"
)

type Service struct {
	policy Policy
}

func NewService(policy Policy) *Service {
	return &Service{policy: policy}
}

func (s *Service) Route(ctx context.Context, panel string, result annotation.Result, matcher Matcher) (Decision, error) {
	if err := ctx.Err(); err != nil {
		return Decision{}, err
	}
	if isNilInterface(matcher) {
		return Decision{}, ErrNoMatcher
	}
	targetName := s.policy.DefaultTarget
	if panelTarget, ok := s.policy.PanelTargets[panel]; ok {
		targetName = panelTarget
	}
	if matched := matcher.Match(result); matched {
		for _, rule := range s.policy.Rules {
			if rule.ID == matcher.RuleID() {
				targetName = rule.Target
				break
			}
		}
	}
	target, ok := s.policy.Targets[targetName]
	if !ok {
		return Decision{}, fmt.Errorf("route target %q: %w", targetName, ErrUnknownTarget)
	}
	decision := Decision{
		VariantKey: result.VariantKey,
		Target:     target,
		RuleID:     matcher.RuleID(),
		Metadata: map[string]string{
			"panel":  panel,
			"target": target.Name,
			"rule":   matcher.RuleID(),
		},
		Result: result,
	}
	return decision, nil
}
