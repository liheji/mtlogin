package mttransport

import (
	"context"

	"mtlogin/pkg/apicore"
)

const cookieScope = "mteam"

func CookieScope() apicore.Tripper {
	return func(next apicore.RoundTripper) apicore.RoundTripper {
		return apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
			if req != nil {
				req.CookieScope = cookieScope
			}
			return next.RoundTrip(ctx, req)
		})
	}
}
