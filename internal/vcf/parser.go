package vcf

import (
	"bufio"
	"fmt"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/variant"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	MaximumLineBytes int
	MaximumRecords   int
}
type Result struct {
	Header   Header
	Variants []variant.Variant
	Warnings []string
}

func NewParser() *Parser { return &Parser{MaximumLineBytes: 1 << 20, MaximumRecords: 100000} }
func (p *Parser) Parse(r io.Reader) (Result, error) {
	result := Result{Header: NewHeader(), Variants: []variant.Variant{}}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64<<10), p.MaximumLineBytes)
	lineNo := 0
	sawColumns := false
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if strings.HasPrefix(line, "##") {
			if err := result.Header.AddMetadata(line); err != nil {
				return Result{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			continue
		}
		if strings.HasPrefix(line, "#CHROM") {
			if err := result.Header.SetColumns(line); err != nil {
				return Result{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			sawColumns = true
			continue
		}
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		if !sawColumns {
			return Result{}, fmt.Errorf("line %d: missing column header: %w", lineNo, platform.ErrInvalid)
		}
		v, err := parseRecord(line, lineNo)
		if err != nil {
			return Result{}, err
		}
		result.Variants = append(result.Variants, v)
		if len(result.Variants) > p.MaximumRecords {
			return Result{}, fmt.Errorf("record limit exceeded: %w", platform.ErrInvalid)
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("scan vcf: %v: %w", err, platform.ErrInvalid)
	}
	if result.Header.FileFormat == "" {
		result.Warnings = append(result.Warnings, "fileformat metadata is missing")
	}
	return result, nil
}
func parseRecord(line string, lineNo int) (variant.Variant, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 8 {
		return variant.Variant{}, fmt.Errorf("line %d: expected 8 fields: %w", lineNo, platform.ErrInvalid)
	}
	position, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return variant.Variant{}, fmt.Errorf("line %d position: %v: %w", lineNo, err, platform.ErrInvalid)
	}
	var quality *float64
	if fields[5] != "." {
		v, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			return variant.Variant{}, fmt.Errorf("line %d quality: %v: %w", lineNo, err, platform.ErrInvalid)
		}
		quality = &v
	}
	filters := []string{}
	if fields[6] != "." {
		filters = strings.Split(fields[6], ";")
	}
	v := variant.Variant{Chromosome: fields[0], Position: position, ID: fields[2], Reference: fields[3], Alternate: fields[4], Quality: quality, Filters: filters, Info: parseInfo(fields[7]), SourceLine: lineNo}
	return v, nil
}
func parseInfo(value string) map[string][]string {
	result := map[string][]string{}
	if value == "." || value == "" {
		return result
	}
	for _, part := range strings.Split(value, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 1 {
			result[kv[0]] = nil
		} else {
			result[kv[0]] = strings.Split(kv[1], ",")
		}
	}
	return result
}
