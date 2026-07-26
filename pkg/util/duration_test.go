package util

import (
	"testing"
	"time"
)

func TestSecondsDurationPreservesSubMillisecondPrecision(t *testing.T) {
	if got, want := SecondsDuration(0.0005), 500*time.Microsecond; got != want {
		t.Fatalf("SecondsDuration(0.0005) = %s, want %s", got, want)
	}
}
