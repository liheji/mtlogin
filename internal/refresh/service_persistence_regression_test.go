package refresh

import (
	"context"
	"errors"
	"testing"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

func TestRunWithSessionPersistsRotatedDIDOnNonAuthFailure(t *testing.T) {
	store := newFakeStore()
	initial := domain.AuthState{Token: "token", DID: "old-did", VisitorID: "visitor"}
	rotated := domain.AuthState{Token: "token", DID: "new-did", VisitorID: "visitor"}
	store.state = initial
	session := &orderedSession{
		updateErr: errors.New("temporary update failure"),
		snapshot:  rotated,
	}
	service := NewService(Config{Timeout: 0}, testClientFunc{
		NewSessionFunc: func(domain.AuthState) Session { return session },
	}, nil, store, &fakeNotifier{})
	ctx := log.NewContext(context.Background())

	err := service.runWithSession(ctx, ctx, initial, store.epoch, false)
	if err == nil {
		t.Fatal("expected non-auth request failure")
	}
	if store.saveAuthCalls != 1 {
		t.Fatalf("expected rotated session state to be persisted once, calls=%d", store.saveAuthCalls)
	}
	if store.state.DID != "new-did" {
		t.Fatalf("expected rotated DID to be persisted, state=%+v", store.state)
	}
}

func TestRunWithSessionSkipsRotatedDIDOnAuthFailure(t *testing.T) {
	store := newFakeStore()
	initial := domain.AuthState{Token: "token", DID: "old-did", VisitorID: "visitor"}
	store.state = initial
	session := &orderedSession{
		updateErr: domain.MTAuthFailed.WithCause(errors.New("session expired")),
		snapshot:  domain.AuthState{Token: "token", DID: "new-did", VisitorID: "visitor"},
	}
	service := NewService(Config{Timeout: 0}, testClientFunc{
		NewSessionFunc: func(domain.AuthState) Session { return session },
	}, nil, store, &fakeNotifier{})
	ctx := log.NewContext(context.Background())

	err := service.runWithSession(ctx, ctx, initial, store.epoch, false)
	if !errors.Is(err, domain.MTAuthFailed) {
		t.Fatalf("expected authentication failure, got %v", err)
	}
	if store.saveAuthCalls != 0 {
		t.Fatalf("authentication failure must not persist session state, calls=%d", store.saveAuthCalls)
	}
}
