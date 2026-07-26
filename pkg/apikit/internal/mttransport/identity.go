package mttransport

import (
	"context"
	"strings"

	"mtlogin/pkg/apicore"
)

func IdentityHeaders(state *State) apicore.Tripper {
	return func(next apicore.RoundTripper) apicore.RoundTripper {
		return apicore.RoundTripFunc(func(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
			if req == nil {
				return next.RoundTrip(ctx, req)
			}
			if req.Headers == nil {
				req.Headers = map[string]string{}
			}
			snapshot := state.Snapshot()
			if snapshot.Token != "" {
				req.Headers["Authorization"] = snapshot.Token
			}
			if snapshot.DID != "" {
				req.Headers["Did"] = snapshot.DID
			}
			if snapshot.VisitorID != "" {
				req.Headers["visitorid"] = snapshot.VisitorID
			}
			resp, err := next.RoundTrip(ctx, req)
			if resp != nil {
				for key, value := range resp.Headers {
					if strings.EqualFold(key, "Did") {
						state.UpdateDID(value)
						break
					}
				}
			}
			return resp, err
		})
	}
}
