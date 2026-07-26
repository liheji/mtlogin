package mttransport

import (
	"context"
	"strings"

	"mtlogin/pkg/apicore"
)

func UnixNoAuth() apicore.Tripper {
	return func(next apicore.RoundTripper) apicore.RoundTripper {
		return apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
			if req != nil && strings.Contains(req.URL, "/api/system/unix") {
				deleteHeader(req.Headers, "Authorization")
				deleteHeader(req.Headers, "Did")
			}
			return next.RoundTrip(ctx, req)
		})
	}
}

func deleteHeader(headers map[string]string, key string) {
	for candidate := range headers {
		if strings.EqualFold(candidate, key) {
			delete(headers, candidate)
		}
	}
}
