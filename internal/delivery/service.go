package delivery

import "context"

type Batch struct {
	ID       string
	Payloads [][]byte
}

type BatchSender interface {
	Send(context.Context, Attempt) error
}

type Service struct {
	sender BatchSender
}

func NewService(sender BatchSender) *Service {
	return &Service{sender: sender}
}

func (s *Service) Deliver(ctx context.Context, batch Batch) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, payload := range batch.Payloads {
		if err := s.sender.Send(ctx, Attempt{BatchID: batch.ID, Payload: append([]byte(nil), payload...)}); err != nil {
			return err
		}
	}
	return nil
}
