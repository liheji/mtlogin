package store

import "mtlogin/internal/domain"

type Store interface {
	LoadAuthState() (domain.AuthState, domain.Epoch, bool, error)
	LoadAuthAccount() (string, error)
	SaveAuthState(domain.AuthState, domain.Epoch) (bool, error)
	SaveAuthAccount(string, domain.Epoch) (bool, error)
	ResetAuthState() (domain.Epoch, error)
	Close() error
}
