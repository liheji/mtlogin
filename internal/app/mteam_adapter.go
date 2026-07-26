package app

import (
	"errors"

	"mtlogin/internal/domain"
	"mtlogin/internal/refresh"
	"mtlogin/pkg/apikit/apimtauth"
	"mtlogin/pkg/apikit/apimtbrowse"
	"mtlogin/pkg/log"
)

type authClientAdapter struct {
	client *apimtauth.Client
}

func (a authClientAdapter) GenerateVisitorID() string {
	return a.client.GenerateVisitorID()
}

func (a authClientAdapter) Login(ctx *log.Context, req domain.LoginRequest) (domain.AuthState, error) {
	state, err := a.client.Login(ctx, apimtauth.LoginRequest{
		Username: req.Username, Password: req.Password, TotpSecret: req.TotpSecret, VisitorID: req.VisitorID,
	})
	if err != nil {
		return domain.AuthState{}, mapAuthError(err)
	}
	return domain.AuthState{Token: state.Token, DID: state.DID, VisitorID: state.VisitorID}, nil
}

type browseClientAdapter struct {
	client *apimtbrowse.Client
}

func (a browseClientAdapter) NewSession(state domain.AuthState) refresh.Session {
	return browseSessionAdapter{session: a.client.NewSession(apimtbrowse.AuthState{
		Token: state.Token, DID: state.DID, VisitorID: state.VisitorID,
	})}
}

type browseSessionAdapter struct {
	session *apimtbrowse.Session
}

func (a browseSessionAdapter) Warmup(ctx *log.Context) error {
	return mapBrowseError(a.session.Warmup(ctx))
}

func (a browseSessionAdapter) Profile(ctx *log.Context) (*domain.Profile, error) {
	profile, err := a.session.Profile(ctx)
	if err != nil {
		return nil, mapBrowseError(err)
	}
	return &domain.Profile{
		Username: profile.Username, UploadedBytes: profile.UploadedBytes, DownloadedBytes: profile.DownloadedBytes,
		Bonus: profile.Bonus, LastLogin: profile.LastLogin, LastBrowse: profile.LastBrowse,
	}, nil
}

func (a browseSessionAdapter) BrowseForum(ctx *log.Context) (bool, error) {
	ok, err := a.session.BrowseForum(ctx)
	return ok, mapBrowseError(err)
}

func (a browseSessionAdapter) UpdateLastBrowse(ctx *log.Context) error {
	return mapBrowseError(a.session.UpdateLastBrowse(ctx))
}

func (a browseSessionAdapter) Snapshot() domain.AuthState {
	state := a.session.Snapshot()
	return domain.AuthState{Token: state.Token, DID: state.DID, VisitorID: state.VisitorID}
}

func mapAuthError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, apimtauth.ErrAuthFailed):
		return domain.MTAuthFailed.WithCause(err)
	case errors.Is(err, apimtauth.ErrHTTPStatus):
		return domain.MTHTTPStatus.WithCause(err)
	case errors.Is(err, apimtauth.ErrTOTPRequired):
		return domain.MTTOTPRequired.WithCause(err)
	case errors.Is(err, apimtauth.ErrBusiness):
		return domain.MTBusinessFailed.WithCause(err)
	default:
		return err
	}
}

func mapBrowseError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, apimtbrowse.ErrAuthFailed):
		return domain.MTAuthFailed.WithCause(err)
	case errors.Is(err, apimtbrowse.ErrHTTPStatus):
		return domain.MTHTTPStatus.WithCause(err)
	case errors.Is(err, apimtbrowse.ErrTOTPRequired):
		return domain.MTTOTPRequired.WithCause(err)
	case errors.Is(err, apimtbrowse.ErrRequestBuild):
		return domain.MTRequestBuildFailed.WithCause(err)
	case errors.Is(err, apimtbrowse.ErrBusiness):
		return domain.MTBusinessFailed.WithCause(err)
	default:
		return err
	}
}
