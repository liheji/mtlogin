package apimtbrowse

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

func TestProfileBuildsExpectedRequestAndMapsData(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS","data":{"username":"u","memberCount":{"uploaded":"10","downloaded":"2","bonus":"3.5"},"memberStatus":{"lastLogin":"l","lastBrowse":"b"}}}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io", UserAgent: "browse-ua"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	profile, err := session.Profile(log.NewContext(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if handler.request.Method != http.MethodPost || !strings.Contains(handler.request.URL, "/api/member/profile") {
		t.Fatalf("unexpected request: %+v", handler.request)
	}
	if handler.request.CookieScope != "mteam" || handler.request.UserAgent != "browse-ua" {
		t.Fatalf("browse client did not assemble mteam transport chain: %+v", handler.request)
	}
	if profile.Username != "u" || profile.UploadedBytes != 10 || profile.DownloadedBytes != 2 || profile.Bonus != 3.5 {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestProfileMapsHTTPStatusToErrno(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  500,
		Body:    `server error`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	_, err = session.Profile(log.NewContext(context.Background()))
	if !errors.Is(err, ErrHTTPStatus) {
		t.Fatalf("expected MTHTTPStatus, got %v", err)
	}
}
