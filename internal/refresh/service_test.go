package refresh

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"mtlogin/internal/domain"
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/apimtauth"
	"mtlogin/pkg/apikit/apimtbrowse"
	"mtlogin/pkg/log"
)

func TestTokenAuthUsesConfiguredIdentityWithoutSaving(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "stored-visitor"}
	handler := newScriptedHandler()
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, tokenMTConfig(), false, notifier)

	if err := svc.Run(log.NewContext(context.Background())); err != nil {
		t.Fatal(err)
	}
	if store.saveAuthCalls != 0 {
		t.Fatalf("token auth must not call SaveAuthState, calls=%d", store.saveAuthCalls)
	}
	if got := handler.firstHeader("Authorization"); got != "cfg-token" {
		t.Fatalf("token auth should use configured token, got %q", got)
	}
	if got := handler.firstHeader("Did"); got != "cfg-did" {
		t.Fatalf("token auth should use configured did, got %q", got)
	}
	if got := handler.firstHeader("visitorid"); got != "cfg-visitor" {
		t.Fatalf("token auth should use configured visitorid, got %q", got)
	}
	if store.state.Token != "stored-token" || store.state.DID != "stored-did" || store.state.VisitorID != "stored-visitor" {
		t.Fatalf("token auth must not mutate stored state, got %+v", store.state)
	}
}

func TestTokenAuthDoesNotPersistSessionIdentityChanges(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	handler.defaultResponse.Headers = map[string]string{"Did": "session-did"}
	svc := newTestService(t, store, handler, tokenMTConfig(), false, &fakeNotifier{})

	if err := svc.Run(log.NewContext(context.Background())); err != nil {
		t.Fatal(err)
	}
	if store.saveAuthCalls != 0 {
		t.Fatalf("token auth must not call SaveAuthState, calls=%d", store.saveAuthCalls)
	}
	if store.state.Token != "" || store.state.DID != "stored-did" || store.state.VisitorID != "visitor" {
		t.Fatalf("token auth must not mutate stored state, got %+v", store.state)
	}
}

func TestTokenAuthFailureAlwaysNotifiesUserToUpdateConfig(t *testing.T) {
	store := newFakeStore()
	handler := newScriptedHandler()
	handler.defaultResponse = apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}}
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, tokenMTConfig(), false, notifier)

	for i := 0; i < 2; i++ {
		err := svc.Run(log.NewContext(context.Background()))
		if !errors.Is(err, domain.MTAuthFailed) {
			t.Fatalf("expected auth error, got %v", err)
		}
	}
	if notifier.calls != 2 {
		t.Fatalf("token auth failure should notify every time, calls=%d", notifier.calls)
	}
	if !strings.Contains(notifier.messages[0].Body, "mt.token_auth.auth") || !strings.Contains(notifier.messages[0].Body, "mt.token_auth.did") || !strings.Contains(notifier.messages[0].Body, "mt.token_auth.visitor_id") {
		t.Fatalf("token auth failure should ask user to update config fields, got %q", notifier.messages[0].Body)
	}
}

func TestPasswordAuthRetriesOnWarmupAuthFailure(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	handler.defaultResponse = apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}}
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, passwordMTConfig(), false, notifier)

	err := svc.Run(log.NewContext(context.Background()))
	if !errors.Is(err, domain.MTAuthFailed) {
		t.Fatalf("expected auth error, got %v", err)
	}
	if handler.countPath("/api/login") != 0 {
		t.Fatalf("expected no immediate login retry, got %d", handler.countPath("/api/login"))
	}
	if store.resetCalls != 1 {
		t.Fatalf("expected identity reset once, got %d", store.resetCalls)
	}
	if store.state != (domain.AuthState{}) {
		t.Fatalf("expected local auth state to be cleared, got %+v", store.state)
	}
	if notifier.calls != 1 || !strings.Contains(notifier.messages[0].Body, "登录已过期") {
		t.Fatalf("expected login expired notification, calls=%d messages=%+v", notifier.calls, notifier.messages)
	}
}

func TestPasswordAuthCanRetryLoginImmediatelyAfterAuthFailure(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	handler.set("/api/system/unix",
		apicore.Response{Status: 401, Body: `{"message":"need login"}`, Headers: map[string]string{}},
		successResponse(),
	)
	handler.set("/api/login", apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{"Authorization": "login-token", "Did": "login-did"},
	})
	notifier := &fakeNotifier{}
	svc := newTestService(t, store, handler, passwordMTConfig(), true, notifier)

	if err := svc.Run(log.NewContext(context.Background())); err != nil {
		t.Fatal(err)
	}
	if store.resetCalls != 1 {
		t.Fatalf("expected identity reset once, got %d", store.resetCalls)
	}
	if handler.countPath("/api/login") != 1 {
		t.Fatalf("expected immediate login retry once, got %d", handler.countPath("/api/login"))
	}
	if store.state.Token != "login-token" || store.state.DID != "login-did" || store.state.VisitorID == "" {
		t.Fatalf("expected new login state to be saved, got %+v", store.state)
	}
}

func TestPasswordAuthNoRetryOnNonAuthError(t *testing.T) {
	store := newFakeStore()
	handler := newScriptedHandler()
	handler.set("/api/login", apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS"}`,
		Headers: map[string]string{"Authorization": "login-token", "Did": "login-did"},
	})
	handler.set("/api/system/unix", apicore.Response{Status: 500, Body: `server error`, Headers: map[string]string{}})
	svc := newTestService(t, store, handler, passwordMTConfig(), false, &fakeNotifier{})

	err := svc.Run(log.NewContext(context.Background()))
	if err == nil || errors.Is(err, domain.MTAuthFailed) {
		t.Fatalf("expected non-auth error, got %v", err)
	}
	if handler.countPath("/api/login") != 1 {
		t.Fatalf("expected 1 login call (no retry), got %d", handler.countPath("/api/login"))
	}
	if store.resetCalls != 0 {
		t.Fatalf("non-auth error must not reset auth state, got %d", store.resetCalls)
	}
}

func TestPasswordAuthWithTokenSkipsLogin(t *testing.T) {
	store := newFakeStore()
	store.state = domain.AuthState{Token: "stored-token", DID: "stored-did", VisitorID: "visitor"}
	handler := newScriptedHandler()
	svc := newTestService(t, store, handler, passwordMTConfig(), false, &fakeNotifier{})

	if err := svc.Run(log.NewContext(context.Background())); err != nil {
		t.Fatal(err)
	}
	if handler.countPath("/api/login") != 0 {
		t.Fatalf("expected 0 login calls, got %d", handler.countPath("/api/login"))
	}
}

func TestRunChecksWarmsUpBeforeProfileAndBeforeUpdate(t *testing.T) {
	session := &orderedSession{browseForumUpdated: true}
	profile, err := runChecks(log.NewContext(context.Background()), session)
	if err != nil {
		t.Fatal(err)
	}
	if profile == nil {
		t.Fatal("expected profile")
	}
	want := []string{"warmup", "profile", "warmup", "update", "browseforum"}
	assertCalls(t, session.calls, want)
}

func TestRunChecksUsesForumUpdateWhenInitialUpdateFails(t *testing.T) {
	updateErr := errors.New("update failed")
	session := &orderedSession{updateErr: updateErr, browseForumUpdated: true}
	profile, err := runChecks(log.NewContext(context.Background()), session)
	if err != nil {
		t.Fatalf("expected forum update success to satisfy refresh, got %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile")
	}
	want := []string{"warmup", "profile", "warmup", "update", "browseforum"}
	assertCalls(t, session.calls, want)
}

func TestRunChecksReturnsInitialUpdateErrorWhenNoUpdateSucceeds(t *testing.T) {
	updateErr := errors.New("update failed")
	session := &orderedSession{updateErr: updateErr}
	profile, err := runChecks(log.NewContext(context.Background()), session)
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected initial update error, got %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile")
	}
	want := []string{"warmup", "profile", "warmup", "update", "browseforum"}
	assertCalls(t, session.calls, want)
}

type fakeStore struct {
	state         domain.AuthState
	account       string
	epoch         domain.Epoch
	saveAuthCalls int
	resetCalls    int
	resetErr      error
}

func newFakeStore() *fakeStore { return &fakeStore{account: "u", epoch: 1} }

func (f *fakeStore) LoadAuthState() (domain.AuthState, domain.Epoch, bool, error) {
	ok := f.state.Token != "" || f.state.DID != "" || f.state.VisitorID != ""
	return f.state, f.epoch, ok, nil
}

func (f *fakeStore) LoadAuthAccount() (string, error) {
	return f.account, nil
}

func (f *fakeStore) SaveAuthAccount(account string, epoch domain.Epoch) (bool, error) {
	if epoch != f.epoch {
		return false, nil
	}
	f.account = account
	return true, nil
}

func (f *fakeStore) SaveAuthState(state domain.AuthState, epoch domain.Epoch) (bool, error) {
	f.saveAuthCalls++
	if epoch != f.epoch {
		return false, nil
	}
	f.state = state
	return true, nil
}
func (f *fakeStore) ResetAuthState() (domain.Epoch, error) {
	f.resetCalls++
	if f.resetErr != nil {
		return f.epoch, f.resetErr
	}
	f.epoch++
	f.state = domain.AuthState{}
	f.account = ""
	return f.epoch, nil
}
func (f *fakeStore) Close() error { return nil }

type fakeNotifier struct {
	calls    int
	messages []Message
}

func (f *fakeNotifier) Notify(ctx *log.Context, msg Message) error {
	f.calls++
	f.messages = append(f.messages, msg)
	return nil
}

type orderedSession struct {
	calls              []string
	updateErr          error
	browseForumUpdated bool
	browseForumErr     error
	snapshot           domain.AuthState
}

func (s *orderedSession) Warmup(ctx *log.Context) error {
	s.calls = append(s.calls, "warmup")
	return nil
}

func (s *orderedSession) Profile(ctx *log.Context) (*domain.Profile, error) {
	s.calls = append(s.calls, "profile")
	return &domain.Profile{Username: "u"}, nil
}

func (s *orderedSession) BrowseForum(ctx *log.Context) (bool, error) {
	s.calls = append(s.calls, "browseforum")
	return s.browseForumUpdated, s.browseForumErr
}

func (s *orderedSession) UpdateLastBrowse(ctx *log.Context) error {
	s.calls = append(s.calls, "update")
	return s.updateErr
}

func (s *orderedSession) Snapshot() domain.AuthState {
	return s.snapshot
}

type requestRecord struct {
	path    string
	headers map[string]string
}

type scriptedHandler struct {
	mu              sync.Mutex
	responses       map[string][]apicore.Response
	calls           map[string]int
	requests        []requestRecord
	defaultResponse apicore.Response
}

func newScriptedHandler() *scriptedHandler {
	return &scriptedHandler{
		responses:       map[string][]apicore.Response{},
		calls:           map[string]int{},
		defaultResponse: successResponse(),
	}
}

func (e *scriptedHandler) set(path string, responses ...apicore.Response) {
	e.responses[path] = responses
}

func (e *scriptedHandler) RoundTrip(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	parsed, err := url.Parse(req.URL)
	path := req.URL
	if err == nil {
		path = parsed.Path
	}
	e.requests = append(e.requests, requestRecord{path: path, headers: cloneHeaders(req.Headers)})

	idx := e.calls[path]
	e.calls[path] = idx + 1
	responses := e.responses[path]
	if len(responses) == 0 {
		return &e.defaultResponse, nil
	}
	if idx >= len(responses) {
		return &responses[len(responses)-1], nil
	}
	return &responses[idx], nil
}

func (e *scriptedHandler) countPath(path string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.calls[path]
}

func (e *scriptedHandler) firstHeader(key string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, req := range e.requests {
		if req.headers[key] != "" {
			return req.headers[key]
		}
	}
	return ""
}

func successResponse() apicore.Response {
	return apicore.Response{Status: 200, Body: `{"code":"0","message":"SUCCESS","data":{}}`, Headers: map[string]string{}}
}

func cloneHeaders(headers map[string]string) map[string]string {
	next := make(map[string]string, len(headers))
	for k, v := range headers {
		next[k] = v
	}
	return next
}

func newTestService(t *testing.T, store *fakeStore, handler *scriptedHandler, authCfg apimtauth.Config, retry bool, notifier *fakeNotifier) *Service {
	t.Helper()
	browseClient, err := apimtbrowse.NewClient(apimtbrowse.Config{
		APIHost:    "api.m-team.io",
		Referer:    "https://kp.m-team.cc/",
		UserAgent:  "ua",
		Version:    "1.1.4",
		WebVersion: "1140",
	}, handler)
	if err != nil {
		t.Fatal(err)
	}
	authClient, err := apimtauth.NewClient(authCfg, handler)
	if err != nil {
		t.Fatal(err)
	}
	authManager := NewAuthManager(AuthConfig{
		PasswordAuth: PasswordAuthConfig{
			Username: authCfg.PasswordAuth.Username, Password: authCfg.PasswordAuth.Password, TotpSecret: authCfg.PasswordAuth.TotpSecret,
		},
		TokenAuth: domain.AuthState{
			Token: authCfg.TokenAuth.Auth, DID: authCfg.TokenAuth.DID, VisitorID: authCfg.TokenAuth.VisitorID,
		},
		RetryOnAuthFailure: retry,
	}, testAuthClientAdapter{client: authClient}, store)
	return NewService(Config{
		Timeout: 30 * time.Second,
	}, testClientFunc{
		NewSessionFunc: func(state domain.AuthState) Session {
			return testBrowseSessionAdapter{session: browseClient.NewSession(apimtbrowse.AuthState{
				Token: state.Token, DID: state.DID, VisitorID: state.VisitorID,
			})}
		},
	}, authManager, store, notifier)
}

type testClientFunc struct {
	NewSessionFunc func(state domain.AuthState) Session
}

func (a testClientFunc) NewSession(state domain.AuthState) Session {
	return a.NewSessionFunc(state)
}

type testAuthClientAdapter struct {
	client *apimtauth.Client
}

func (a testAuthClientAdapter) GenerateVisitorID() string {
	return a.client.GenerateVisitorID()
}

func (a testAuthClientAdapter) Login(ctx *log.Context, req domain.LoginRequest) (domain.AuthState, error) {
	state, err := a.client.Login(ctx, apimtauth.LoginRequest{
		Username: req.Username, Password: req.Password, TotpSecret: req.TotpSecret, VisitorID: req.VisitorID,
	})
	if err != nil {
		if errors.Is(err, apimtauth.ErrAuthFailed) {
			return domain.AuthState{}, domain.MTAuthFailed.WithCause(err)
		}
		return domain.AuthState{}, err
	}
	return domain.AuthState{Token: state.Token, DID: state.DID, VisitorID: state.VisitorID}, nil
}

type testBrowseSessionAdapter struct {
	session *apimtbrowse.Session
}

func (a testBrowseSessionAdapter) Warmup(ctx *log.Context) error {
	return testBrowseError(a.session.Warmup(ctx))
}

func (a testBrowseSessionAdapter) Profile(ctx *log.Context) (*domain.Profile, error) {
	profile, err := a.session.Profile(ctx)
	if err != nil {
		return nil, testBrowseError(err)
	}
	return &domain.Profile{
		Username: profile.Username, UploadedBytes: profile.UploadedBytes, DownloadedBytes: profile.DownloadedBytes,
		Bonus: profile.Bonus, LastLogin: profile.LastLogin, LastBrowse: profile.LastBrowse,
	}, nil
}

func (a testBrowseSessionAdapter) BrowseForum(ctx *log.Context) (bool, error) {
	ok, err := a.session.BrowseForum(ctx)
	return ok, testBrowseError(err)
}

func (a testBrowseSessionAdapter) UpdateLastBrowse(ctx *log.Context) error {
	return testBrowseError(a.session.UpdateLastBrowse(ctx))
}

func (a testBrowseSessionAdapter) Snapshot() domain.AuthState {
	state := a.session.Snapshot()
	return domain.AuthState{Token: state.Token, DID: state.DID, VisitorID: state.VisitorID}
}

func testBrowseError(err error) error {
	if errors.Is(err, apimtbrowse.ErrAuthFailed) {
		return domain.MTAuthFailed.WithCause(err)
	}
	return err
}

func passwordMTConfig() apimtauth.Config {
	return apimtauth.Config{
		APIHost:      "api.m-team.io",
		PasswordAuth: apimtauth.PasswordAuthConfig{Username: "u", Password: "p"},
	}
}

func tokenMTConfig() apimtauth.Config {
	return apimtauth.Config{
		APIHost:   "api.m-team.io",
		TokenAuth: apimtauth.TokenAuthConfig{Auth: "cfg-token", DID: "cfg-did", VisitorID: "cfg-visitor"},
	}
}

func assertCalls(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("calls: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("calls: got %v want %v", got, want)
		}
	}
}
