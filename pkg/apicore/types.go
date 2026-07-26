package apicore

import (
	"context"
	"net/http"
	"time"
)

type Request struct {
	Server          string
	URL             string
	Method          string
	Headers         map[string]string
	Body            string
	BodyBytes       []byte
	Timeout         time.Duration
	Proxy           string
	DisableRedirect bool
	UserAgent       string
	CookieScope     string
}

type Response struct {
	Status  int
	Body    string
	Headers map[string]string
}

// RoundTripper is apicore's own request executor abstraction. It deliberately
// does not implement net/http.RoundTripper because the production executor is
// tls-client, whose public API is Do(req), not an exported transport builder.
type RoundTripper interface {
	RoundTrip(ctx context.Context, req *Request) (*Response, error)
}

type RoundTripFunc func(ctx context.Context, req *Request) (*Response, error)

func (f RoundTripFunc) RoundTrip(ctx context.Context, req *Request) (*Response, error) {
	return f(ctx, req)
}

type Tripper func(RoundTripper) RoundTripper

func Chain(base RoundTripper, trippers ...Tripper) RoundTripper {
	if base == nil {
		base = RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
			return nil, http.ErrNotSupported
		})
	}
	rt := base
	for i := len(trippers) - 1; i >= 0; i-- {
		rt = trippers[i](rt)
	}
	return rt
}
