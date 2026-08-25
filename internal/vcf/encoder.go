package vcf

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/example/genome-variant-annotation/internal/variant"
)

type Encoder struct {
	writer  *bufio.Writer
	pending []encodedVariant
}

// encodedVariant snapshots the INFO field at WriteVariant time. Encoding lazily
// at Flush would otherwise observe later mutations to v.Info made by the
// caller between WriteVariant and Flush.
type encodedVariant struct {
	variant     variant.Variant
	encodedInfo string
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{writer: bufio.NewWriter(w), pending: []encodedVariant{}}
}

func (e *Encoder) WriteHeader(h Header) error {
	format := h.FileFormat
	if format == "" {
		format = "VCFv4.3"
	}
	if _, err := fmt.Fprintf(e.writer, "##fileformat=%s\n", format); err != nil {
		return err
	}
	keys := make([]string, 0, len(h.Metadata))
	for key := range h.Metadata {
		if key != "fileformat" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range h.Metadata[key] {
			if _, err := fmt.Fprintf(e.writer, "##%s=%s\n", key, value); err != nil {
				return err
			}
		}
	}
	columns := "#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO"
	if len(h.Samples) > 0 {
		columns += "\tFORMAT\t" + strings.Join(h.Samples, "\t")
	}
	_, err := fmt.Fprintln(e.writer, columns)
	return err
}

func (e *Encoder) WriteVariant(v variant.Variant) error {
	// Snapshot the encoded INFO now so later mutations to v.Info (e.g. reusing
	// the Info map for the next record) cannot change what Flush emits.
	e.pending = append(e.pending, encodedVariant{variant: v, encodedInfo: encodeInfo(v.Info)})
	return nil
}

func (e *Encoder) writeVariant(ev encodedVariant) error {
	v := ev.variant
	quality := "."
	if v.Quality != nil {
		quality = strconv.FormatFloat(*v.Quality, 'f', -1, 64)
	}
	filters := "."
	if len(v.Filters) > 0 {
		filters = strings.Join(v.Filters, ";")
	}
	info := ev.encodedInfo
	if info == "" {
		info = "."
	}
	id := v.ID
	if id == "" {
		id = "."
	}
	_, err := fmt.Fprintf(e.writer, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n", v.Chromosome, v.Position, id, v.Reference, v.Alternate, quality, filters, info)
	return err
}

func (e *Encoder) Flush() error {
	for _, ev := range e.pending {
		if err := e.writeVariant(ev); err != nil {
			return err
		}
	}
	e.pending = e.pending[:0]
	return e.writer.Flush()
}

func encodeInfo(info map[string][]string) string {
	if len(info) == 0 {
		return "."
	}
	keys := make([]string, 0, len(info))
	for key := range info {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		if len(info[key]) == 0 {
			values = append(values, key)
		} else {
			values = append(values, key+"="+strings.Join(info[key], ","))
		}
	}
	return strings.Join(values, ";")
}
