package apimtbrowse

import (
	"errors"
	"net/url"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/internal/mttransport"
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

func (c *Client) NewSession(state AuthState) *Session {
	return &Session{client: c, state: mttransport.NewState(mttransport.Identity{
		Token: state.Token, DID: state.DID, VisitorID: state.VisitorID,
	})}
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
