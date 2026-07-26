package mtproto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"mtlogin/pkg/apicore"
)

var (
	ErrAuthFailed   = errors.New("mteam authentication failed")
	ErrHTTPStatus   = errors.New("mteam http status error")
	ErrBusiness     = errors.New("mteam business error")
	ErrTOTPRequired = errors.New("mteam totp required")
)

type APIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (r *APIResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Code    json.RawMessage `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Code) > 0 && string(raw.Code) != "null" {
		if raw.Code[0] == '"' {
			if err := json.Unmarshal(raw.Code, &r.Code); err != nil {
				return err
			}
		} else {
			decoder := json.NewDecoder(bytes.NewReader(raw.Code))
			decoder.UseNumber()
			var code json.Number
			if err := decoder.Decode(&code); err != nil {
				return err
			}
			r.Code = code.String()
		}
	}
	r.Message = raw.Message
	r.Data = raw.Data
	return nil
}

func ParseResponse(res *apicore.Response, err error) (APIResponse, error) {
	if err := ValidateHTTPResponse(res, err); err != nil {
		return APIResponse{}, err
	}
	var body APIResponse
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		return APIResponse{}, fmt.Errorf("%w: decode response: %v", ErrBusiness, err)
	}
	if body.Code == "0" {
		return body, nil
	}
	detail := fmt.Sprintf("code=%s", body.Code)
	switch body.Code {
	case "1001", "10001":
		return APIResponse{}, fmt.Errorf("%w: %s", ErrTOTPRequired, detail)
	case "401", "NOT_LOGIN", "NEED_LOGIN", "LOGIN_REQUIRED", "AUTH_EXPIRED":
		return APIResponse{}, fmt.Errorf("%w: %s", ErrAuthFailed, detail)
	default:
		return APIResponse{}, fmt.Errorf("%w: %s", ErrBusiness, detail)
	}
}

func ValidateHTTPResponse(res *apicore.Response, err error) error {
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("%w: empty response", ErrBusiness)
	}
	if res.Status == http.StatusUnauthorized {
		return fmt.Errorf("%w: status=%d", ErrAuthFailed, res.Status)
	}
	if res.Status != http.StatusOK {
		return fmt.Errorf("%w: status=%d", ErrHTTPStatus, res.Status)
	}
	return nil
}
