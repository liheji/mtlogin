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

func TestUpdateLastBrowseBuildsExpectedRequest(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	if err := session.UpdateLastBrowse(log.NewContext(context.Background())); err != nil {
		t.Fatal(err)
	}
	if handler.request.Method != http.MethodPost || !strings.Contains(handler.request.URL, "/api/member/updateLastBrowse") {
		t.Fatalf("unexpected request: %+v", handler.request)
	}
}

func TestUpdateLastBrowseMapsBusinessErrors(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"NOT_LOGIN","message":"need login"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	err = session.UpdateLastBrowse(log.NewContext(context.Background()))
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("expected auth errno, got %v", err)
	}
}
