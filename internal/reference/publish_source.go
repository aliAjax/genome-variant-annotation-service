package reference

import (
	"context"
	"fmt"
	"io"
)

type PublicationObjectStore interface {
	OpenPublication(context.Context, string) (io.ReadCloser, error)
}

type PublicationSource struct {
	store PublicationObjectStore
}

func NewPublicationSource(store PublicationObjectStore) *PublicationSource {
	return &PublicationSource{store: store}
}

func (s *PublicationSource) Load(ctx context.Context, key string) (data []byte, err error) {
	reader, err := s.store.OpenPublication(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("open publication %s: %v", key, err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			err = fmt.Errorf("close publication %s: %w", key, closeErr)
		}
	}()
	data, err = io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read publication %s: %v", key, err)
	}
	return data, nil
}
