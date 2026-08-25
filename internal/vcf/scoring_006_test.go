package vcf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/genome-variant-annotation/internal/variant"
)

func TestEncoderStableOutput(t *testing.T) {
	var buffer bytes.Buffer
	encoder := NewEncoder(&buffer)
	v := variant.Variant{Chromosome: "1", Position: 10, Reference: "A", Alternate: "G", Info: map[string][]string{"AF": {"0.1"}}}
	if err := encoder.WriteVariant(v); err != nil {
		t.Fatal(err)
	}
	v.Info["AF"][0] = "0.9"
	if err := encoder.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buffer.String(), "AF=0.1") {
		t.Fatalf("encoded output was mutated after write: %q", buffer.String())
	}
}
