package routing

import (
	"fmt"
	"strings"

	"github.com/example/genome-variant-annotation/internal/platform"
)

func LoadPolicy(defaultTarget string, targets []Target, rules []Rule) (Policy, error) {
	defaultTarget = strings.TrimSpace(defaultTarget)
	if defaultTarget == "" {
		return Policy{}, fmt.Errorf("load routing policy: %w", platform.ErrInvalid)
	}
	policy := Policy{
		DefaultTarget: defaultTarget,
		Targets:       make(map[string]Target, len(targets)),
		PanelTargets:  make(map[string]string),
		Rules:         append([]Rule(nil), rules...),
	}
	for _, target := range targets {
		name := strings.TrimSpace(target.Name)
		if name == "" || strings.TrimSpace(target.Endpoint) == "" {
			return Policy{}, fmt.Errorf("load routing target: %w", platform.ErrInvalid)
		}
		target.Name = name
		target.Labels = cloneLabels(target.Labels)
		policy.Targets[name] = target
	}
	if _, ok := policy.Targets[defaultTarget]; !ok {
		return Policy{}, fmt.Errorf("default target %q: %w", defaultTarget, ErrUnknownTarget)
	}
	for _, rule := range policy.Rules {
		if _, ok := policy.Targets[rule.Target]; !ok {
			return Policy{}, fmt.Errorf("rule %s target %q: %w", rule.ID, rule.Target, ErrUnknownTarget)
		}
	}
	return policy, nil
}

func (p *Policy) BindPanel(panel, target string) error {
	panel = strings.TrimSpace(panel)
	if panel == "" {
		return platform.ErrInvalid
	}
	if _, ok := p.Targets[target]; !ok {
		return ErrUnknownTarget
	}
	p.PanelTargets[panel] = target
	return nil
}

func cloneLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}
	copyLabels := make(map[string]string, len(labels))
	for key, value := range labels {
		copyLabels[key] = value
	}
	return copyLabels
}
