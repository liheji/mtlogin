package refresh

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mtlogin/internal/domain"
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

func TestResolveDoesNotReuseCachedIdentityAfterUsernameChange(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "old-token", DID: "old-did", VisitorID: "old-visitor"}
	store.account = "old-user"
	client := &regressionLoginClient{}
	manager := NewAuthManager(AuthConfig{
		PasswordAuth: PasswordAuthConfig{Username: "new-user", Password: "password"},
	}, client, store)

	state, _, tokenMode, err := manager.Resolve(log.NewContext(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if tokenMode {
		t.Fatal("password mode must not be reported as token mode")
	}
	if client.loginCalls != 1 {
		t.Fatalf("expected a fresh login after username change, calls=%d", client.loginCalls)
	}
	if store.resetCalls != 1 {
		t.Fatalf("expected cached identity reset after username change, calls=%d", store.resetCalls)
	}
	if state.Token != "new-token" || state.DID != "new-did" || state.VisitorID != "new-visitor" {
		t.Fatalf("unexpected refreshed state: %+v", state)
	}
	if store.account != "new-user" {
		t.Fatalf("cached account was not rebound: %q", store.account)
	}
}

func TestResolveDoesNotReuseLegacyIdentityWithoutAccountBinding(t *testing.T) {
	store := newFakeStore()
	store.account = ""
	store.state = domain.AuthState{Token: "legacy-token", DID: "legacy-did", VisitorID: "legacy-visitor"}
	client := &regressionLoginClient{}
	manager := NewAuthManager(AuthConfig{
		PasswordAuth: PasswordAuthConfig{Username: "u", Password: "password"},
	}, client, store)

	state, _, _, err := manager.Resolve(log.NewContext(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if client.loginCalls != 1 || store.resetCalls != 1 {
		t.Fatalf("legacy identity should be reset and replaced: loginCalls=%d resetCalls=%d", client.loginCalls, store.resetCalls)
	}
	if state.Token != "new-token" {
		t.Fatalf("unexpected replacement state: %+v", state)
	}
}

func TestHandleAuthExpiredReportsResetFailureWithoutClaimingCleared(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "token", DID: "did", VisitorID: "visitor"}
	store.resetErr = errors.New("disk full")
	notifier := &fakeNotifier{}
	handler := newScriptedHandler()
	handler.defaultResponse = apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}}
	service := newTestService(t, store, handler, passwordMTConfig(), false, notifier)

	err := service.Run(log.NewContext(context.Background()))
	if err == nil || !errors.Is(err, domain.RefreshIdentityResetFailed) || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("expected reset failure, got %v", err)
	}
	if notifier.calls == 0 {
		t.Fatal("reset failure must notify the user")
	}
	if strings.Contains(notifier.messages[0].Body, "已清除") {
		t.Fatalf("reset failure notification falsely claims identity was cleared: %q", notifier.messages[0].Body)
	}
}

type regressionLoginClient struct {
	loginCalls int
}

func (c *regressionLoginClient) GenerateVisitorID() string { return "new-visitor" }

func (c *regressionLoginClient) Login(_ *log.Context, req domain.LoginRequest) (domain.AuthState, error) {
	c.loginCalls++
	return domain.AuthState{Token: "new-token", DID: "new-did", VisitorID: req.VisitorID}, nil
}
