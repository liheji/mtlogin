package apimtauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mtlogin/pkg/apikit/internal/mtproto"
	"mtlogin/pkg/log"
	"mtlogin/pkg/util"
)

func (c *Client) login(ctx *log.Context, req LoginRequest, otp bool) (*AuthState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := "/api/login"
	t := time.Now().UnixMilli()
	body := url.Values{}
	if otp {
		tk, err := Generate(req.TotpSecret)
		if err != nil {
			return nil, err
		}
		body.Add("otpCode", tk)
	}
	body.Add("username", req.Username)
	body.Add("password", req.Password)
	body.Add("turnstile", "")
	body.Add("_timestamp", strconv.FormatInt(t, 10))
	body.Add("_sgin", mtproto.Signature(http.MethodPost, path, t))

	did, err := util.String(32)
	if err != nil {
		return nil, err
	}
	apiReq := c.baseRequest()
	apiReq.URL = fmt.Sprintf("https://%s%s", c.cfg.APIHost, path)
	apiReq.Method = http.MethodPost
	apiReq.Body = body.Encode()
	apiReq.Headers["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	apiReq.Headers["Did"] = did
	apiReq.Headers["visitorid"] = req.VisitorID

	res, err := c.rt.RoundTrip(ctx, &apiReq)
	_, err = mtproto.ParseResponse(res, err)
	if errors.Is(err, mtproto.ErrTOTPRequired) && !otp {
		return c.login(ctx, req, true)
	}
	if err != nil {
		return nil, translateProtocolError(err)
	}
	responseDID := res.Headers["Did"]
	if responseDID == "" {
		responseDID = res.Headers["did"]
	}
	token := res.Headers["Authorization"]
	if token == "" {
		token = res.Headers["authorization"]
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("%w: successful login response missing Authorization header", ErrAuthFailed)
	}
	state := AuthState{Token: token, DID: responseDID, VisitorID: req.VisitorID}
	if state.DID == "" {
		state.DID = did
	}
	return &state, nil
}
