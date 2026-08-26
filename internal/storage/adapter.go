package storage

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Write(context.Context, string, io.Reader) (string, error)
	Read(context.Context, string) (io.ReadCloser, error)
	Remove(context.Context, string) error
}
type Compressor interface {
	Compress(io.Writer, io.Reader) error
	Decompress(io.Writer, io.Reader) error
}
