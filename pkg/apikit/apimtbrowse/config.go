package apimtbrowse

import (
	"errors"
	"time"
)

type Config struct {
	APIHost    string
	Referer    string
	UserAgent  string
	Version    string
	WebVersion string
	Proxy      string
	Timeout    time.Duration
}

func (c Config) Validate() error {
	if c.APIHost == "" {
		return errors.New("mt.api_host is required")
	}
	if c.Timeout <= 0 {
		return errors.New("mt.timeout must be > 0")
	}
	return nil
}
