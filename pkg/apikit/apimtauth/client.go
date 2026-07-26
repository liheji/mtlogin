package apimtauth

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/internal/mttransport"
	"mtlogin/pkg/log"
	"mtlogin/pkg/util"
)

type Client struct {
	cfg      Config
	proxyURL *url.URL
	rt       apicore.RoundTripper
}

func NewClient(cfg Config, base apicore.RoundTripper) (*Client, error) {
	var proxyURL *url.URL
	var err error
	if cfg.Proxy != "" {
		proxyURL, err = url.Parse(cfg.Proxy)
		if err != nil {
			return nil, err
		}
	}
	if base == nil {
		return nil, errors.New("mteam round tripper is nil")
	}
	rt := apicore.Chain(
		base,
		apicore.Logging(),
		apicore.Proxy(),
		mttransport.CookieScope(),
		mttransport.Headers(mttransport.HeadersConfig{
			Referer:    cfg.Referer,
			UserAgent:  cfg.UserAgent,
			Version:    cfg.Version,
			WebVersion: cfg.WebVersion,
		}),
		mttransport.UnixNoAuth(),
	)
	return &Client{cfg: cfg, proxyURL: proxyURL, rt: rt}, nil
}

func (c *Client) GenerateVisitorID() string {
	id, err := util.String(32)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return id
}

func (c *Client) baseRequest() apicore.Request {
	req := apicore.Request{
		Server:          "mteam",
		Headers:         map[string]string{},
		Timeout:         c.cfg.Timeout,
		DisableRedirect: true,
	}
	if c.proxyURL != nil {
		req.Proxy = c.proxyURL.String()
	}
	return req
}

func (c *Client) Login(ctx *log.Context, req LoginRequest) (*AuthState, error) {
	if c == nil || c.rt == nil {
		return nil, errors.New("mteam client is nil")
	}
	if req.VisitorID == "" {
		return nil, errors.New("visitorid is required")
	}
	return c.login(ctx, req, false)
}
