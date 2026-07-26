package apimtauth

import (
	"testing"
	"time"
)

func TestGenerateAtMatchesRFC6238Vector(t *testing.T) {
	code, err := generateAt("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", time.Unix(59, 0))
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" {
		t.Fatalf("code=%q", code)
	}
}

func TestGenerateAtRejectsInvalidSecret(t *testing.T) {
	if _, err := generateAt("invalid!", time.Unix(59, 0)); err == nil {
		t.Fatal("expected invalid secret error")
	}
}
