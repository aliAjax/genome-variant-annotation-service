package delivery

import (
	"context"
	"time"
)

type RetryPolicy struct {
	MaxAttempts    int
	AttemptTimeout time.Duration
}

func (p RetryPolicy) attempts() int {
	if p.MaxAttempts < 1 {
		return 1
	}
	return p.MaxAttempts
}

func (p RetryPolicy) AttemptContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if p.AttemptTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, p.AttemptTimeout)
}
