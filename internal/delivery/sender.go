package delivery

import (
	"context"
	"errors"
)

type Attempt struct {
	BatchID  string
	Sequence int
	Payload  []byte
}

type Transport interface {
	Send(context.Context, Attempt) error
}

type Sender struct {
	transport Transport
	policy    RetryPolicy
}

func NewSender(transport Transport, policy RetryPolicy) *Sender {
	return &Sender{transport: transport, policy: policy}
}

func (s *Sender) Send(ctx context.Context, attempt Attempt) error {
	var lastErr error
	for number := 1; number <= s.policy.attempts(); number++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		attempt.Sequence = number
		attemptCtx, cancel := s.policy.AttemptContext(ctx)
		err := s.transport.Send(attemptCtx, attempt)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if !shouldRetry(ctx, err, number, s.policy.attempts()) {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			return err
		}
	}
	return lastErr
}

func shouldRetry(ctx context.Context, err error, attempt, limit int) bool {
	if ctx.Err() != nil || attempt >= limit {
		return false
	}
	return !errors.Is(err, context.Canceled)
}
