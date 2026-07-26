package apimtauth

import (
	"errors"
	"time"
)

type Config struct {
	APIHost      string
	Referer      string
	UserAgent    string
	Version      string
	WebVersion   string
	Proxy        string
	Timeout      time.Duration
	PasswordAuth PasswordAuthConfig
	TokenAuth    TokenAuthConfig
}

type PasswordAuthConfig struct {
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	TotpSecret string `toml:"totp_secret"`
}

type TokenAuthConfig struct {
	Auth      string `toml:"auth"`
	DID       string `toml:"did"`
	VisitorID string `toml:"visitor_id"`
}

func (c Config) Validate() error {
	if c.APIHost == "" {
		return errors.New("mt.api_host is required")
	}
	if c.Timeout <= 0 {
		return errors.New("mt.timeout must be > 0")
	}
	hasPassword := c.PasswordAuth.Username != "" && c.PasswordAuth.Password != ""
	hasPasswordPartial := c.PasswordAuth.Username != "" || c.PasswordAuth.Password != "" || c.PasswordAuth.TotpSecret != ""
	hasToken := c.TokenAuth.Auth != "" && c.TokenAuth.DID != "" && c.TokenAuth.VisitorID != ""
	hasTokenPartial := c.TokenAuth.Auth != "" || c.TokenAuth.DID != "" || c.TokenAuth.VisitorID != ""
	if hasPasswordPartial && !hasPassword {
		return errors.New("mt.password_auth.username and password are required when password_auth is configured")
	}
	if hasTokenPartial && !hasToken {
		return errors.New("mt.token_auth.auth, did and visitor_id are required when token_auth is configured")
	}
	if !hasPassword && !hasToken {
		return errors.New("either mt.password_auth.username/password or mt.token_auth.auth/did/visitor_id is required")
	}
	return nil
}
