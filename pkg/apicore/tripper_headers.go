package apicore

import (
	"context"
)

func Headers() Tripper {
	return func(next RoundTripper) RoundTripper {
		return RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
			ensureHeaders(req)
			for key, value := range HeadersFromContext(ctx) {
				if _, exists := req.Headers[key]; !exists {
					req.Headers[key] = value
				}
			}
			return next.RoundTrip(ctx, req)
		})
	}
}

func Header(key, value string) Tripper {
	return func(next RoundTripper) RoundTripper {
		return RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
			ensureHeaders(req)
			if _, exists := req.Headers[key]; !exists {
				req.Headers[key] = value
			}
			return next.RoundTrip(ctx, req)
		})
	}
}

func ensureHeaders(req *Request) {
	if req == nil {
		return
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
}
