package mtproto

import (
	"strings"
	"testing"

	"mtlogin/pkg/apicore"
)

func TestParseResponseDoesNotExposeResponseSecrets(t *testing.T) {
	_, err := ParseResponse(&apicore.Response{
		Status: 200,
		Body:   `{"code":"FAILED","message":"token=top-secret"}`,
	}, nil)
	if err == nil {
		t.Fatal("expected business error")
	}
	if strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("protocol error exposed response body: %v", err)
	}
}

func TestParseResponseDoesNotExposeMalformedBody(t *testing.T) {
	_, err := ParseResponse(&apicore.Response{Status: 200, Body: `<token>top-secret</token>`}, nil)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("decode error exposed response body: %v", err)
	}
}
