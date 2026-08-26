package routing

import (
	"errors"

	"github.com/example/genome-variant-annotation/internal/annotation"
)

var (
	ErrUnknownTarget = errors.New("unknown routing target")
	ErrNoMatcher     = errors.New("routing matcher is unavailable")
)

type Target struct {
	Name     string            `json:"name"`
	Endpoint string            `json:"endpoint"`
	Labels   map[string]string `json:"labels,omitempty"`
}

type Rule struct {
	ID            string `json:"id"`
	ClinicalLabel string `json:"clinical_label,omitempty"`
	MinimumImpact string `json:"minimum_impact,omitempty"`
	Target        string `json:"target"`
}

type Policy struct {
	DefaultTarget string            `json:"default_target"`
	Targets       map[string]Target `json:"targets"`
	PanelTargets  map[string]string `json:"panel_targets"`
	Rules         []Rule            `json:"rules"`
}

type Decision struct {
	VariantKey string            `json:"variant_key"`
	Target     Target            `json:"target"`
	RuleID     string            `json:"rule_id,omitempty"`
	Metadata   map[string]string `json:"metadata"`
	Result     annotation.Result `json:"result"`
}
