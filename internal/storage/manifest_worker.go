package storage

import "context"

type ManifestPartSource interface {
	NextPart(context.Context, string) (*ManifestPart, error)
}

type ManifestAppender interface {
	Append(context.Context, string, *ManifestPart) (*Manifest, error)
}

type ManifestWorker struct {
	source   ManifestPartSource
	appender ManifestAppender
}

func NewManifestWorker(source ManifestPartSource, appender ManifestAppender) *ManifestWorker {
	return &ManifestWorker{source: source, appender: appender}
}

func (w *ManifestWorker) Process(ctx context.Context, manifestID, objectID string) (*Manifest, error) {
	part, err := w.source.NextPart(ctx, objectID)
	if err != nil {
		return nil, err
	}
	return w.appender.Append(ctx, manifestID, part)
}
