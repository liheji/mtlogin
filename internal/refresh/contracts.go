package refresh

import (
	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

type Session interface {
	Warmup(ctx *log.Context) error
	Profile(ctx *log.Context) (*domain.Profile, error)
	BrowseForum(ctx *log.Context) (bool, error)
	UpdateLastBrowse(ctx *log.Context) error
	Snapshot() domain.AuthState
}

type Store interface {
	LoadAuthState() (state domain.AuthState, epoch domain.Epoch, ok bool, err error)
	LoadAuthAccount() (account string, err error)
	SaveAuthState(state domain.AuthState, epoch domain.Epoch) (saved bool, err error)
	SaveAuthAccount(account string, epoch domain.Epoch) (saved bool, err error)
	ResetAuthState() (newEpoch domain.Epoch, err error)
	Close() error
}

type Notifier interface {
	Notify(ctx *log.Context, msg Message) error
}

type Client interface {
	NewSession(state domain.AuthState) Session
}

type AuthManager interface {
	Resolve(ctx *log.Context) (state domain.AuthState, epoch domain.Epoch, tokenMode bool, err error)
	HandleAuthExpired(ctx *log.Context, epoch domain.Epoch) (state domain.AuthState, newEpoch domain.Epoch, retried bool, err error)
}
