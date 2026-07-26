package apicore

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestErrorStringSanitizesSensitiveValues(t *testing.T) {
	text := errorString(errors.New("request failed: https://api.telegram.org/bot123456/sendMessage?token=top-secret"))
	if strings.Contains(text, "123456") || strings.Contains(text, "top-secret") {
		t.Fatalf("error text was not sanitized: %q", text)
	}
}

func TestLoggingCallsNext(t *testing.T) {
	called := false
	rt := Logging()(RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
		called = true
		return &Response{Status: 204}, nil
	}))

	resp, err := rt.RoundTrip(context.Background(), &Request{Server: "test", URL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || resp.Status != 204 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if !called {
		t.Fatal("expected logging tripper to call next")
	}
}

func TestLoggingReturnsNextError(t *testing.T) {
	want := errors.New("boom")
	rt := Logging()(RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
		return nil, want
	}))

	_, err := rt.RoundTrip(context.Background(), &Request{Server: "test", URL: "https://example.com"})
	if !errors.Is(err, want) {
		t.Fatalf("expected next error %v, got %v", want, err)
	}
}
