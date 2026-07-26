package refresh

import (
	"context"
	"errors"
	"testing"

	"mtlogin/internal/domain"
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

// H1: retry_on_auth_failure=true 时，本地身份已被清空但重新登录失败，
// 必须通知用户，否则保活彻底坏掉却无人知晓。
func TestPasswordAuthNotifiesWhenImmediateReloginFails(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	handler.defaultResponse = apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}}
	// 重新登录本身也失败（密码/TOTP 错误）。
	handler.set("/api/login", apicore.Response{Status: 401, Body: `{"message":"bad credentials"}`, Headers: map[string]string{}})
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, passwordMTConfig(), true, notifier)

	err := svc.Run(log.NewContext(context.Background()))
	if err == nil {
		t.Fatal("expected error when re-login fails")
	}
	if store.resetCalls != 1 {
		t.Fatalf("expected identity reset once, got %d", store.resetCalls)
	}
	if handler.countPath("/api/login") != 1 {
		t.Fatalf("expected one immediate login attempt, got %d", handler.countPath("/api/login"))
	}
	if notifier.calls == 0 {
		t.Fatal("re-login failure must notify the user; got zero notifications")
	}
}

// M6: 重新登录成功、但新会话再次认证失败时，用户被告知“正在重新登录”后
// 必须收到最终失败结果，且不得因重试形成登录循环。
func TestPasswordAuthNotifiesWhenRetriedRunStillFailsAuth(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	handler.defaultResponse = apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}}
	handler.set("/api/login", apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{"Authorization": "login-token", "Did": "login-did"},
	})
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, passwordMTConfig(), true, notifier)

	err := svc.Run(log.NewContext(context.Background()))
	if !errors.Is(err, domain.MTAuthFailed) {
		t.Fatalf("expected auth error from retried run, got %v", err)
	}
	if handler.countPath("/api/login") != 1 {
		t.Fatalf("expected exactly one re-login (no loop), got %d", handler.countPath("/api/login"))
	}
	if notifier.calls < 2 {
		t.Fatalf("retried-run auth failure must notify a terminal result, calls=%d messages=%+v", notifier.calls, notifier.messages)
	}
}
