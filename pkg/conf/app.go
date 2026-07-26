package conf

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron/v3"
)

const AppConfigPath = "conf/app.toml"

// AppConfig 对应固定文件 conf/app.toml 的进程级配置。
type AppConfig struct {
	System  SystemConfig    `toml:"system" json:"system"`
	App     SchedulerConfig `toml:"app" json:"app"`
	Refresh RefreshConfig   `toml:"refresh" json:"refresh"`
	MT      MTeamConfig     `toml:"mt" json:"mt"`
}

type SystemConfig struct {
	Proxy string `toml:"proxy" json:"proxy"`
}

type SchedulerConfig struct {
	Crontab  string  `toml:"crontab" json:"crontab"`
	MinDelay float64 `toml:"min_delay" json:"min_delay"`
	MaxDelay float64 `toml:"max_delay" json:"max_delay"`
}

type RefreshConfig struct {
	Timeout            float64 `toml:"timeout" json:"timeout"`
	RetryOnAuthFailure bool    `toml:"retry_on_auth_failure" json:"retry_on_auth_failure"`
}

type MTeamConfig struct {
	Proxy        string             `toml:"proxy" json:"proxy"`
	APIHost      string             `toml:"api_host" json:"api_host"`
	Referer      string             `toml:"referer" json:"referer"`
	Timeout      float64            `toml:"timeout" json:"timeout"`
	PasswordAuth PasswordAuthConfig `toml:"password_auth" json:"password_auth"`
	TokenAuth    TokenAuthConfig    `toml:"token_auth" json:"token_auth"`
	Header       MTeamHeaderConfig  `toml:"header" json:"header"`
}

type PasswordAuthConfig struct {
	Username   string `toml:"username" json:"username"`
	Password   string `toml:"password" json:"-"`
	TotpSecret string `toml:"totp_secret" json:"-"`
}

type TokenAuthConfig struct {
	Auth      string `toml:"auth" json:"-"`
	DID       string `toml:"did" json:"did"`
	VisitorID string `toml:"visitor_id" json:"visitor_id"`
}

type MTeamHeaderConfig struct {
	Version    string `toml:"version" json:"version"`
	WebVersion string `toml:"web_version" json:"web_version"`
	UserAgent  string `toml:"user_agent" json:"user_agent"`
}

func LoadApp(path string) (*AppConfig, error) {
	cfg := &AppConfig{}
	meta, err := LoadTOMLWithMeta(path, cfg)
	if err != nil {
		return nil, err
	}
	if err := rejectRemovedKeys(meta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := rejectUnknownKeys(meta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return cfg, nil
}

func (c AppConfig) Validate() error {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(c.App.Crontab); err != nil {
		return fmt.Errorf("invalid app.crontab: %w", err)
	}
	if c.App.MinDelay < 0 || c.App.MaxDelay < 0 {
		return errors.New("app.min_delay and app.max_delay must be >= 0")
	}
	if c.App.MinDelay > c.App.MaxDelay {
		return errors.New("app.min_delay must be <= app.max_delay")
	}
	if c.Refresh.Timeout <= 0 {
		return errors.New("refresh.timeout must be > 0")
	}
	return nil
}

func rejectRemovedKeys(meta toml.MetaData) error {
	for _, key := range meta.Undecoded() {
		switch key.String() {
		case "refresh.normal_clear_token_threshold", "refresh.external_auth_notify_initial", "refresh.external_auth_notify_interval", "refresh.cookie_mode":
			return fmt.Errorf("%s has been removed from refresh config", key.String())
		case "mt.external_auth", "mt.external_auth.auth", "mt.external_auth.did", "mt.external_auth.visitor_id":
			return errors.New("mt.external_auth has been renamed to mt.token_auth")
		}
	}
	return nil
}
