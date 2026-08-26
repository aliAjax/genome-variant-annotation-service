package artifact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Encoder interface {
	Encode(io.Writer, variant.Variant) error
}

type Service struct {
	repo    Repository
	store   ObjectWriter
	encoder Encoder
	clock   platform.Clock
}

func NewService(repo Repository, store ObjectWriter, encoder Encoder, clock platform.Clock) *Service {
	return &Service{repo: repo, store: store, encoder: encoder, clock: clock}
}

func (s *Service) Create(ctx context.Context, datasetID string) (Bundle, error) {
	if strings.TrimSpace(datasetID) == "" {
		return Bundle{}, platform.ErrInvalid
	}
	now := s.clock.Now()
	bundle := Bundle{ID: platform.NewID("artifact"), DatasetID: datasetID, Status: StatusQueued, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Save(ctx, bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func (s *Service) Export(ctx context.Context, id string, batches [][]variant.Variant) (bundle Bundle, err error) {
	bundle, err = s.repo.Get(ctx, id)
	if err != nil {
		return Bundle{}, err
	}
	bundle.Status = StatusBuilding
	bundle.Error = ""
	bundle.UpdatedAt = s.clock.Now()
	if err = s.repo.Save(ctx, bundle); err != nil {
		return Bundle{}, err
	}
	tx := NewObjectTransaction(s.store)
	for index, batch := range batches {
		name := fmt.Sprintf("%s/part-%05d.vcf", bundle.ID, index+1)
		tx.Stage(name)
		if err = s.writeObject(ctx, name, batch); err != nil {
			rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
			bundle.Status = StatusFailed
			bundle.Error = err.Error()
			bundle.UpdatedAt = s.clock.Now()
			_ = s.repo.Save(context.WithoutCancel(ctx), bundle)
			return bundle, errors.Join(err, rollbackErr)
		}
		bundle.Objects = append(bundle.Objects, name)
	}
	if err = tx.Commit(ctx); err != nil {
		rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
		bundle.Status = StatusFailed
		bundle.Error = err.Error()
		_ = s.repo.Save(context.WithoutCancel(ctx), bundle)
		return bundle, errors.Join(err, rollbackErr)
	}
	bundle.Status = StatusReady
	bundle.UpdatedAt = s.clock.Now()
	if err = s.repo.Save(ctx, bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func (s *Service) writeObject(ctx context.Context, name string, batch []variant.Variant) (err error) {
	writer, err := s.store.Create(ctx, name)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, writer.Close()) }()
	for _, item := range batch {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.encoder.Encode(writer, item); err != nil {
			return err
		}
	}
	return nil
}
