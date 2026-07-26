package apimtbrowse

import (
	"context"

	"mtlogin/pkg/apicore"
)

type captureHandler struct {
	request  apicore.Request
	response apicore.Response
}

func (e *captureHandler) RoundTrip(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
	e.request = *req
	return &e.response, nil
}

func testState() AuthState {
	return AuthState{Token: "token", DID: "did", VisitorID: "visitor"}
}
