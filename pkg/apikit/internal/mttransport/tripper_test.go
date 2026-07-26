package mttransport

import (
	"context"
	"testing"

	"mtlogin/pkg/apicore"
)

func TestIdentityHeadersInjectsAndUpdatesState(t *testing.T) {
	state := NewState(Identity{Token: "token", DID: "old", VisitorID: "visitor"})
	base := apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
		if req.Headers["Authorization"] != "token" || req.Headers["Did"] != "old" || req.Headers["visitorid"] != "visitor" {
			t.Fatalf("identity headers missing: %+v", req.Headers)
		}
		return &apicore.Response{Headers: map[string]string{"Did": "new"}}, nil
	})
	_, err := apicore.Chain(base, IdentityHeaders(state)).RoundTrip(context.Background(), &apicore.Request{})
	if err != nil {
		t.Fatal(err)
	}
	if state.Snapshot().DID != "new" {
		t.Fatalf("did was not updated: %+v", state.Snapshot())
	}
}

func TestHeadersSetsTLSUserAgentAndCookieScope(t *testing.T) {
	base := apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
		if req.UserAgent != "ua" || req.CookieScope != "mteam" {
			t.Fatalf("request metadata missing: %+v", req)
		}
		return &apicore.Response{}, nil
	})
	chain := apicore.Chain(base, CookieScope(), Headers(HeadersConfig{UserAgent: "ua"}))
	if _, err := chain.RoundTrip(context.Background(), &apicore.Request{}); err != nil {
		t.Fatal(err)
	}
}

func TestUnixNoAuthRemovesIdentityHeaders(t *testing.T) {
	base := apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
		if req.Headers["Authorization"] != "" || req.Headers["Did"] != "" {
			t.Fatalf("unix request kept auth headers: %+v", req.Headers)
		}
		return &apicore.Response{}, nil
	})
	req := &apicore.Request{URL: "https://api.m-team.io/api/system/unix", Headers: map[string]string{"Authorization": "token", "Did": "did"}}
	if _, err := apicore.Chain(base, UnixNoAuth()).RoundTrip(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}
