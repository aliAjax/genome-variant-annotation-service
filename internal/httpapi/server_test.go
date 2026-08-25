package httpapi

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/example/genome-variant-annotation/internal/platform"
)

func TestServerReturnsBadRequestForVCFParse(t *testing.T) {
	if got := statusFor(fmt.Errorf("parse: %w", platform.ErrInvalid)); got != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", got)
	}
}
