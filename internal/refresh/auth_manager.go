package refresh

import (
	"errors"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

type PasswordAuthConfig struct {
	Username   string
	Password   string
	TotpSecret string
}

type AuthConfig struct {
	PasswordAuth       PasswordAuthConfig
	TokenAuth          domain.AuthState
	RetryOnAuthFailure bool
}

type LoginClient interface {
	GenerateVisitorID() string
	Login(ctx *log.Context, req domain.LoginRequest) (domain.AuthState, error)
}

type authManager struct {
	cfg    AuthConfig
	client LoginClient
	store  Store
}

func NewAuthManager(cfg AuthConfig, client LoginClient, store Store) AuthManager {
	return &authManager{cfg: cfg, client: client, store: store}
}

func (m *authManager) Resolve(ctx *log.Context) (domain.AuthState, domain.Epoch, bool, error) {
	if !m.cfg.TokenAuth.Empty() {
		if !m.cfg.TokenAuth.Complete() {
			err := domain.ConfigInvalid.WithCause(errors.New("mt.token_auth.auth, mt.token_auth.did and mt.token_auth.visitor_id are required"))
			return domain.AuthState{}, 0, true, err
		}
		return m.cfg.TokenAuth, 0, true, nil
	}

	state, epoch, _, err := m.store.LoadAuthState()
	if err != nil {
		return domain.AuthState{}, 0, false, err
	}
	account, err := m.store.LoadAuthAccount()
	if err != nil {
		return domain.AuthState{}, 0, false, err
	}
	if account != m.cfg.PasswordAuth.Username && !state.Empty() {
		epoch, err = m.store.ResetAuthState()
		if err != nil {
			return domain.AuthState{}, 0, false, domain.RefreshIdentityResetFailed.WithCause(err)
		}
		state = domain.AuthState{}
	}
	if account != m.cfg.PasswordAuth.Username {
		saved, err := m.store.SaveAuthAccount(m.cfg.PasswordAuth.Username, epoch)
		if err != nil {
			return domain.AuthState{}, 0, false, err
		}
		if !saved {
			return domain.AuthState{}, 0, false, identityChangedError()
		}
	}
	if state.VisitorID == "" {
		state.VisitorID = m.client.GenerateVisitorID()
		saved, err := m.store.SaveAuthState(state, epoch)
		if err != nil {
			return domain.AuthState{}, 0, false, err
		}
		if !saved {
			return domain.AuthState{}, 0, false, identityChangedError()
		}
	}
	if state.Token == "" || state.DID == "" {
		state, err = m.loginAndSave(ctx, state.VisitorID, epoch)
		if err != nil {
			return domain.AuthState{}, 0, false, err
		}
	}
	return state, epoch, false, nil
}

func (m *authManager) HandleAuthExpired(ctx *log.Context, epoch domain.Epoch) (domain.AuthState, domain.Epoch, bool, error) {
	newEpoch, err := m.store.ResetAuthState()
	if err != nil {
		return domain.AuthState{}, 0, false, domain.RefreshIdentityResetFailed.WithCause(err)
	}
	if !m.cfg.RetryOnAuthFailure {
		return domain.AuthState{}, newEpoch, false, nil
	}

	state := domain.AuthState{VisitorID: m.client.GenerateVisitorID()}
	saved, err := m.store.SaveAuthAccount(m.cfg.PasswordAuth.Username, newEpoch)
	if err != nil {
		return domain.AuthState{}, 0, false, err
	}
	if !saved {
		return domain.AuthState{}, 0, false, identityChangedError()
	}
	saved, err = m.store.SaveAuthState(state, newEpoch)
	if err != nil {
		return domain.AuthState{}, 0, false, err
	}
	if !saved {
		return domain.AuthState{}, 0, false, identityChangedError()
	}
	state, err = m.loginAndSave(ctx, state.VisitorID, newEpoch)
	if err != nil {
		return domain.AuthState{}, 0, false, err
	}
	return state, newEpoch, true, nil
}

func (m *authManager) loginAndSave(ctx *log.Context, visitorID string, epoch domain.Epoch) (domain.AuthState, error) {
	state, err := m.client.Login(ctx, domain.LoginRequest{
		Username:   m.cfg.PasswordAuth.Username,
		Password:   m.cfg.PasswordAuth.Password,
		TotpSecret: m.cfg.PasswordAuth.TotpSecret,
		VisitorID:  visitorID,
	})
	if err != nil {
		return domain.AuthState{}, err
	}
	saved, err := m.store.SaveAuthState(state, epoch)
	if err != nil {
		return domain.AuthState{}, err
	}
	if !saved {
		return domain.AuthState{}, identityChangedError()
	}
	return state, nil
}

func identityChangedError() error {
	return domain.RefreshIdentityChanged.WithCause(errors.New("identity config changed; current run aborted"))
}
