package worker

import (
	"context"
	"log/slog"
	"time"
)

type Cleaner interface {
	Cleanup(context.Context, time.Time) (int, error)
}
type Retention struct {
	cleaner  Cleaner
	interval time.Duration
	ttl      time.Duration
	log      *slog.Logger
}

func NewRetention(cleaner Cleaner, interval, ttl time.Duration, log *slog.Logger) *Retention {
	return &Retention{cleaner: cleaner, interval: interval, ttl: ttl, log: log}
}
func (w *Retention) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			count, err := w.cleaner.Cleanup(ctx, now.Add(-w.ttl))
			if err != nil {
				w.log.Error("retention cleanup failed", "error", err)
				continue
			}
			w.log.Info("retention cleanup completed", "removed", count)
		}
	}
}
