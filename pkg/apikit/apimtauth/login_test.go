package apimtauth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

type captureHandler struct {
	request  apicore.Request
	response apicore.Response
}

func (e *captureHandler) RoundTrip(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
	e.request = *req
	return &e.response, nil
}

type sequenceHandler struct {
	requests  []apicore.Request
	responses []apicore.Response
}

func (e *sequenceHandler) RoundTrip(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
	e.requests = append(e.requests, *req)
	idx := len(e.requests) - 1
	if idx >= len(e.responses) {
		return &e.responses[len(e.responses)-1], nil
	}
	return &e.responses[idx], nil
}

func TestLoginDoesNotSendOriginHeader(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{"Authorization": "token", "Did": "did"},
	}}
	client, err := NewClient(Config{
		APIHost:    "api.m-team.io",
		Referer:    "https://kp.m-team.cc/",
		UserAgent:  "ua",
		Version:    "1.1.4",
		WebVersion: "1140",
	}, handler)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Login(log.NewContext(context.Background()), LoginRequest{
		Username:  "u",
		Password:  "p",
		VisitorID: "visitor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := handler.request.Headers["origin"]; ok {
		t.Fatalf("login request should not send origin header: %+v", handler.request.Headers)
	}
	if handler.request.Server != "mteam" {
		t.Fatalf("mteam requests must set Server for unified API logging, got %q", handler.request.Server)
	}
	if handler.request.CookieScope != "mteam" || handler.request.UserAgent != "ua" {
		t.Fatalf("auth client did not assemble mteam transport chain: %+v", handler.request)
	}
}

func TestLoginRetriesOnTOTPBusinessCode(t *testing.T) {
	handler := &sequenceHandler{responses: []apicore.Response{
		{Status: 200, Body: `{"code":10001,"message":"need totp"}`, Headers: map[string]string{}},
		{Status: 200, Body: `{"code":"0","message":"SUCCESS"}`, Headers: map[string]string{"Authorization": "token", "Did": "did"}},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	state, err := client.Login(log.NewContext(context.Background()), LoginRequest{
		Username:   "u",
		Password:   "p",
		TotpSecret: "2SH3V3GDW7ZNMGYE",
		VisitorID:  "visitor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Token != "token" || state.DID != "did" || state.VisitorID != "visitor" {
		t.Fatalf("unexpected login state: %+v", state)
	}
	if len(handler.requests) != 2 {
		t.Fatalf("expected two login attempts, got %d", len(handler.requests))
	}
	if !strings.Contains(handler.requests[1].Body, "otpCode=") {
		t.Fatalf("second login attempt should include otpCode, body=%q", handler.requests[1].Body)
	}
}

func TestLoginReturnsTOTPRequiredAfterRetryAlsoNeedsTOTP(t *testing.T) {
	handler := &sequenceHandler{responses: []apicore.Response{
		{Status: 200, Body: `{"code":10001,"message":"need totp"}`, Headers: map[string]string{}},
		{Status: 200, Body: `{"code":"10001","message":"need totp again"}`, Headers: map[string]string{}},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Login(log.NewContext(context.Background()), LoginRequest{
		Username:   "u",
		Password:   "p",
		TotpSecret: "2SH3V3GDW7ZNMGYE",
		VisitorID:  "visitor",
	})
	if !errors.Is(err, ErrTOTPRequired) {
		t.Fatalf("expected MTTOTPRequired, got %v", err)
	}
}

func TestLoginRejectsSuccessfulResponseWithoutAuthorization(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{"Did": "did"},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	state, err := client.Login(log.NewContext(context.Background()), LoginRequest{
		Username:  "u",
		Password:  "p",
		VisitorID: "visitor",
	})
	if err == nil {
		t.Fatalf("expected missing Authorization to fail, state=%+v", state)
	}
	if state != nil {
		t.Fatalf("missing Authorization must not return auth state: %+v", state)
	}
}
