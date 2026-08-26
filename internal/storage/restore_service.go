package storage

import "context"

type RestoreLoader interface {
	Restore(context.Context, *RestoreSession) error
}

type RestoreService struct {
	repository RestoreRepository
	loader     RestoreLoader
}

func NewRestoreService(repository RestoreRepository, loader RestoreLoader) *RestoreService {
	return &RestoreService{repository: repository, loader: loader}
}

func (s *RestoreService) Resume(ctx context.Context, id string) (*RestoreSession, error) {
	session, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	runCtx := session.EffectiveContext(ctx)
	if err := runCtx.Err(); err != nil {
		return nil, err
	}
	session.BindContext(runCtx)
	session.State = RestoreRunning
	if err := s.loader.Restore(runCtx, session); err != nil {
		return nil, err
	}
	session.Cursor++
	session.State = RestoreDone
	if err := s.repository.Save(runCtx, session); err != nil {
		return nil, err
	}
	return cloneRestoreSession(session), nil
}
