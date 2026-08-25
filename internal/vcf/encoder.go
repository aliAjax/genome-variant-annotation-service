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

type Encoder struct{ writer *bufio.Writer }

func NewEncoder(w io.Writer) *Encoder { return &Encoder{writer: bufio.NewWriter(w)} }

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
	quality := "."
	if v.Quality != nil {
		quality = strconv.FormatFloat(*v.Quality, 'f', -1, 64)
	}
	filters := "."
	if len(v.Filters) > 0 {
		filters = strings.Join(v.Filters, ";")
	}
	info := encodeInfo(v.Info)
	id := v.ID
	if id == "" {
		id = "."
	}
	_, err := fmt.Fprintf(e.writer, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n", v.Chromosome, v.Position, id, v.Reference, v.Alternate, quality, filters, info)
	return err
}

func (e *Encoder) Flush() error { return e.writer.Flush() }

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
