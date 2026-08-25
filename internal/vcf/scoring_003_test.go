package vcf

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/example/genome-variant-annotation/internal/platform"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestParseRecordWrapsInvalid(t *testing.T) {
	_, err := parseRecord("1\tbad\t.\tA\tG\t.\tPASS\t.", 1)
	if !errors.Is(err, platform.ErrInvalid) {
		t.Fatalf("expected invalid sentinel, got %v", err)
	}
}

func TestParseWrapsInvalid(t *testing.T) {
	_, err := NewParser().Parse(failingReader{})
	if !errors.Is(err, platform.ErrInvalid) {
		t.Fatalf("expected parser error to preserve sentinel, got %v", err)
	}
}

func TestParseQualityWrapsInvalid(t *testing.T) {
	_, err := parseRecord("1\t1\t.\tA\tG\tbad\tPASS\t.", 1)
	if !errors.Is(err, platform.ErrInvalid) {
		t.Fatalf("expected quality invalid sentinel, got %v", err)
	}
}

func TestStreamWrapsInvalid(t *testing.T) {
	_, errCh := Stream(context.Background(), failingReader{}, 1024)
	err := <-errCh
	if !errors.Is(err, platform.ErrInvalid) {
		t.Fatalf("expected stream error to preserve sentinel, got %v", err)
	}
}
