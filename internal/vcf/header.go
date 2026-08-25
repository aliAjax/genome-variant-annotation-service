package vcf

import (
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"strings"
)

type Definition struct {
	ID          string `json:"id"`
	Number      string `json:"number"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
type Header struct {
	FileFormat string                `json:"file_format"`
	Metadata   map[string][]string   `json:"metadata"`
	Info       map[string]Definition `json:"info"`
	Format     map[string]Definition `json:"format"`
	Samples    []string              `json:"samples"`
}

func NewHeader() Header {
	return Header{Metadata: map[string][]string{}, Info: map[string]Definition{}, Format: map[string]Definition{}}
}
func (h *Header) AddMetadata(line string) error {
	if !strings.HasPrefix(line, "##") {
		return fmt.Errorf("metadata prefix: %w", platform.ErrInvalid)
	}
	parts := strings.SplitN(strings.TrimPrefix(line, "##"), "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("metadata delimiter: %w", platform.ErrInvalid)
	}
	if parts[0] == "fileformat" {
		h.FileFormat = parts[1]
	}
	h.Metadata[parts[0]] = append(h.Metadata[parts[0]], parts[1])
	return nil
}
func (h *Header) SetColumns(line string) error {
	fields := strings.Split(strings.TrimPrefix(line, "#"), "\t")
	if len(fields) < 8 {
		return fmt.Errorf("vcf columns: expected at least 8: %w", platform.ErrInvalid)
	}
	expected := []string{"CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO"}
	for i, v := range expected {
		if fields[i] != v {
			return fmt.Errorf("column %d: expected %s: %w", i+1, v, platform.ErrInvalid)
		}
	}
	if len(fields) > 9 {
		h.Samples = append([]string(nil), fields[9:]...)
	}
	return nil
}
