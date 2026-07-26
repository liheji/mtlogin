package apicore

import (
	"os"
	"strings"
	"testing"
)

func TestLoggingPackageDoesNotExposeSinkAPI(t *testing.T) {
	body, err := os.ReadFile("logging.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(body)
	for _, symbol := range []string{"type LogSink", "func SetLogSink", "func WriteAPIEvent"} {
		if strings.Contains(source, symbol) {
			t.Fatalf("apicore logging must not expose %s", symbol)
		}
	}
}
