package storage

import "context"

type RestoreState string

const (
	RestorePending RestoreState = "pending"
	RestoreRunning RestoreState = "running"
	RestoreDone    RestoreState = "done"
)

type RestoreSession struct {
	ID         string       `json:"id"`
	DatasetID  string       `json:"dataset_id"`
	Cursor     int          `json:"cursor"`
	State      RestoreState `json:"state"`
	requestCtx context.Context
}

func NewRestoreSession(ctx context.Context, id, datasetID string) *RestoreSession {
	return &RestoreSession{ID: id, DatasetID: datasetID, State: RestorePending, requestCtx: ctx}
}

func (s *RestoreSession) BindContext(ctx context.Context) {
	s.requestCtx = ctx
}

func (s *RestoreSession) EffectiveContext(current context.Context) context.Context {
	if s.requestCtx != nil {
		return s.requestCtx
	}
	return current
}

func cloneRestoreSession(session *RestoreSession) *RestoreSession {
	if session == nil {
		return nil
	}
	cloned := *session
	return &cloned
}
