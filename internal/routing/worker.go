package routing

import (
	"context"

	"github.com/example/genome-variant-annotation/internal/annotation"
)

type AuditSink interface {
	Record(context.Context, Decision) error
}

type Dispatcher struct {
	service *Service
	audit   AuditSink
}

func NewDispatcher(service *Service, audit AuditSink) *Dispatcher {
	return &Dispatcher{service: service, audit: audit}
}

func (d *Dispatcher) Dispatch(ctx context.Context, panel string, result annotation.Result, matcher Matcher) (Decision, error) {
	decision, err := d.service.Route(ctx, panel, result, matcher)
	if err != nil {
		return Decision{}, err
	}
	if !isNilInterface(d.audit) {
		if err := d.audit.Record(ctx, decision); err != nil {
			return Decision{}, err
		}
	}
	return decision, nil
}
