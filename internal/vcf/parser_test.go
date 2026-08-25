package vcf

import (
	"strings"
	"testing"
)

func TestParseVCF(t *testing.T) {
	input := "##fileformat=VCFv4.3\n#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO\n1\t100\t.\tA\tG\t60\tPASS\tAF=0.01\n"
	result, err := NewParser().Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if result.Header.FileFormat != "VCFv4.3" || len(result.Variants) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Variants[0].Key() != "1:100:A:G" {
		t.Fatalf("unexpected key: %s", result.Variants[0].Key())
	}
}

func TestRejectMissingHeader(t *testing.T) {
	_, err := NewParser().Parse(strings.NewReader("1\t100\t.\tA\tG\t60\tPASS\t.\n"))
	if err == nil {
		t.Fatal("expected missing header error")
	}
}
