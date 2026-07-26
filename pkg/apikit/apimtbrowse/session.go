package apimtbrowse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/internal/mtproto"
	"mtlogin/pkg/apikit/internal/mttransport"
	"mtlogin/pkg/log"
)

type Session struct {
	client *Client
	state  *mttransport.State
}

func (s *Session) Snapshot() AuthState {
	state := s.state.Snapshot()
	return AuthState{Token: state.Token, DID: state.DID, VisitorID: state.VisitorID}
}

func (s *Session) do(ctx *log.Context, method, path string, extraBody map[string]string) (mtproto.APIResponse, error) {
	if err := ctx.Err(); err != nil {
		return mtproto.APIResponse{}, err
	}
	reqTransport := s.buildRequest()
	params := s.buildCommonParams(method, path, extraBody)
	reqURL := url.URL{Scheme: "https", Host: s.client.cfg.APIHost, Path: path}
	if method == http.MethodPost {
		body, contentType := MultipartBody(params)
		reqTransport.Body = body
		reqTransport.Headers["Content-Type"] = contentType
	} else {
		reqURL.RawQuery = QueryBody(params)
		reqTransport.Body = ""
		delete(reqTransport.Headers, "Content-Type")
	}
	reqTransport.URL = reqURL.String()
	reqTransport.Method = method
	rt := apicore.Chain(s.client.rt, mttransport.IdentityHeaders(s.state))
	res, err := rt.RoundTrip(ctx, &reqTransport)
	if path == "/ping" {
		return mtproto.APIResponse{}, translateProtocolError(mtproto.ValidateHTTPResponse(res, err))
	}
	return parseResponse(res, err)
}

func (s *Session) doJSON(ctx *log.Context, method, path string, body map[string]any) (mtproto.APIResponse, error) {
	if err := ctx.Err(); err != nil {
		return mtproto.APIResponse{}, err
	}
	if body == nil {
		body = map[string]any{}
	}
	reqTransport := s.buildRequest()
	params := s.buildCommonParams(method, path, nil)
	reqURL := url.URL{Scheme: "https", Host: s.client.cfg.APIHost, Path: path}
	for k, v := range params {
		body[k] = v
	}
	data, err := json.Marshal(body)
	if err != nil {
		return mtproto.APIResponse{}, fmt.Errorf("%w: json marshal: %v", ErrRequestBuild, err)
	}
	reqTransport.BodyBytes = data
	reqTransport.Body = string(data)
	reqTransport.Headers["Content-Type"] = "application/json; charset=UTF-8"
	reqTransport.URL = reqURL.String()
	reqTransport.Method = method
	rt := apicore.Chain(s.client.rt, mttransport.IdentityHeaders(s.state))
	res, err := rt.RoundTrip(ctx, &reqTransport)
	return parseResponse(res, err)
}

func (s *Session) buildRequest() apicore.Request {
	req := s.client.baseRequest()
	req.Headers["origin"] = s.client.cfg.Referer
	return req
}

func (s *Session) buildCommonParams(method, path string, extra map[string]string) map[string]string {
	t := time.Now().UnixMilli()
	params := map[string]string{
		"_timestamp": strconv.FormatInt(t, 10),
		"_sgin":      mtproto.Signature(method, path, t),
	}
	for k, v := range extra {
		params[k] = v
	}
	return params
}
