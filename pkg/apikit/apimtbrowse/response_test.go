package apimtbrowse

import (
	"errors"
	"testing"

	"mtlogin/pkg/apicore"
)

func TestParseResponseMapsUnauthorized(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 401, Body: "unauthorized"}, nil)
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("expected MTAuthFailed, got %v", err)
	}
}

func TestParseResponseMapsNonOKStatus(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 500, Body: "server error"}, nil)
	if !errors.Is(err, ErrHTTPStatus) {
		t.Fatalf("expected MTHTTPStatus, got %v", err)
	}
}

func TestParseResponseMapsMalformedJSON(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 200, Body: "<html>"}, nil)
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected MTBusinessFailed, got %v", err)
	}
}

func TestParseResponseDoesNotDuplicateBusinessErrorPrefix(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 200, Body: `{"code":"1","message":"FAIL"}`}, nil)
	if err == nil {
		t.Fatal("expected business error")
	}
	if got, want := err.Error(), "mteam business error: code=1"; got != want {
		t.Fatalf("error=%q want=%q", got, want)
	}
}

func TestParseResponseMapsTOTPBusinessCode(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 200, Body: `{"code":10001,"message":"need totp"}`}, nil)
	if !errors.Is(err, ErrTOTPRequired) {
		t.Fatalf("expected MTTOTPRequired, got %v", err)
	}
}

func TestParseResponseAcceptsSuccess(t *testing.T) {
	_, err := parseResponse(&apicore.Response{Status: 200, Body: `{"code":"0","message":"SUCCESS"}`}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
