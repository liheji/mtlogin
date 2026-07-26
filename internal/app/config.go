package app

import (
	"errors"
	"fmt"

	"mtlogin/internal/domain"
	"mtlogin/pkg/apikit/apimtauth"
	"mtlogin/pkg/apikit/apimtbrowse"
	"mtlogin/pkg/conf"
	"mtlogin/pkg/util"
)

func mapConfigError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, conf.ErrRead):
		return domain.ConfigReadFailed.WithCause(err)
	case errors.Is(err, conf.ErrParse):
		return domain.ConfigParseFailed.WithCause(err)
	case errors.Is(err, conf.ErrInvalid):
		return domain.ConfigInvalid.WithCause(err)
	default:
		return err
	}
}

func buildMTeamConfigs(raw conf.MTeamConfig) (apimtauth.Config, apimtbrowse.Config, error) {
	auth := apimtauth.Config{
		APIHost:    raw.APIHost,
		Referer:    raw.Referer,
		UserAgent:  raw.Header.UserAgent,
		Version:    raw.Header.Version,
		WebVersion: raw.Header.WebVersion,
		Proxy:      raw.Proxy,
		Timeout:    util.SecondsDuration(raw.Timeout),
		PasswordAuth: apimtauth.PasswordAuthConfig{
			Username:   raw.PasswordAuth.Username,
			Password:   raw.PasswordAuth.Password,
			TotpSecret: raw.PasswordAuth.TotpSecret,
		},
		TokenAuth: apimtauth.TokenAuthConfig{
			Auth:      raw.TokenAuth.Auth,
			DID:       raw.TokenAuth.DID,
			VisitorID: raw.TokenAuth.VisitorID,
		},
	}
	browse := apimtbrowse.Config{
		APIHost:    raw.APIHost,
		Referer:    raw.Referer,
		UserAgent:  raw.Header.UserAgent,
		Version:    raw.Header.Version,
		WebVersion: raw.Header.WebVersion,
		Proxy:      raw.Proxy,
		Timeout:    util.SecondsDuration(raw.Timeout),
	}
	if err := auth.Validate(); err != nil {
		return apimtauth.Config{}, apimtbrowse.Config{}, domain.ConfigInvalid.WithCause(fmt.Errorf("validate mt auth config: %w", err))
	}
	if err := browse.Validate(); err != nil {
		return apimtauth.Config{}, apimtbrowse.Config{}, domain.ConfigInvalid.WithCause(fmt.Errorf("validate mt browse config: %w", err))
	}
	return auth, browse, nil
}
