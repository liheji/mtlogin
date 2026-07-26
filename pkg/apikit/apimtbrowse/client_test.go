package apimtbrowse

import "testing"

func TestNewClientRequiresRoundTripper(t *testing.T) {
	_, err := NewClient(Config{APIHost: "api.m-team.io"}, nil)
	if err == nil {
		t.Fatal("expected NewClient to reject nil round tripper")
	}
}
