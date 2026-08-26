package storage

import "context"

type ManifestService struct {
	repository ManifestRepository
}

func NewManifestService(repository ManifestRepository) *ManifestService {
	return &ManifestService{repository: repository}
}

func (s *ManifestService) Append(ctx context.Context, id string, part *ManifestPart) (*Manifest, error) {
	manifest, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := manifest.AddPart(part); err != nil {
		return nil, err
	}
	if err := s.repository.Save(ctx, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}
