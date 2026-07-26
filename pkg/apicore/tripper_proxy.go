package apicore

import (
	"context"
)

func Proxy() Tripper {
	return func(next RoundTripper) RoundTripper {
		return RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
			if req != nil && req.Proxy == "" {
				req.Proxy = ProxyFromContext(ctx)
			}
			return next.RoundTrip(ctx, req)
		})
	}
}
