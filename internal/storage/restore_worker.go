package storage

import "context"

type RestoreRunner interface {
	Resume(context.Context, string) (*RestoreSession, error)
}

type RestoreWorker struct {
	runner   RestoreRunner
	previous context.Context
}

func NewRestoreWorker(runner RestoreRunner) *RestoreWorker {
	return &RestoreWorker{runner: runner}
}

func (w *RestoreWorker) Process(ctx context.Context, sessionID string) (*RestoreSession, error) {
	if w.previous != nil {
		ctx = w.previous
	}
	w.previous = ctx
	return w.runner.Resume(ctx, sessionID)
}
