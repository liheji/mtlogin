package mttransport

import (
	"context"
	"strconv"
	"time"

	"mtlogin/pkg/apicore"
)

type HeadersConfig struct {
	Referer    string
	UserAgent  string
	Version    string
	WebVersion string
}

func Headers(cfg HeadersConfig) apicore.Tripper {
	return func(next apicore.RoundTripper) apicore.RoundTripper {
		return apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
			if req == nil {
				return next.RoundTrip(ctx, req)
			}
			if req.Headers == nil {
				req.Headers = map[string]string{}
			}
			if req.UserAgent == "" {
				req.UserAgent = cfg.UserAgent
			}
			if req.Headers["referer"] == "" {
				req.Headers["referer"] = cfg.Referer
			}
			if req.Headers["Accept"] == "" {
				req.Headers["Accept"] = "application/json;charset=UTF-8"
			}
			if req.Headers["Ts"] == "" {
				req.Headers["Ts"] = strconv.FormatInt(time.Now().Unix(), 10)
			}
			if req.Headers["version"] == "" {
				req.Headers["version"] = cfg.Version
			}
			if req.Headers["webversion"] == "" {
				req.Headers["webversion"] = cfg.WebVersion
			}
			return next.RoundTrip(ctx, req)
		})
	}
}
