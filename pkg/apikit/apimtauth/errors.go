package apimtauth

import (
	"errors"
	"fmt"

	"mtlogin/pkg/apikit/internal/mtproto"
)

var (
	ErrAuthFailed   = errors.New("mteam authentication failed")
	ErrHTTPStatus   = errors.New("mteam http status error")
	ErrBusiness     = errors.New("mteam business error")
	ErrTOTPRequired = errors.New("mteam totp required")
)

func translateProtocolError(err error) error {
	if err == nil {
		return nil
	}
	var target error
	switch {
	case errors.Is(err, mtproto.ErrAuthFailed):
		target = ErrAuthFailed
	case errors.Is(err, mtproto.ErrHTTPStatus):
		target = ErrHTTPStatus
	case errors.Is(err, mtproto.ErrTOTPRequired):
		target = ErrTOTPRequired
	case errors.Is(err, mtproto.ErrBusiness):
		target = ErrBusiness
	default:
		return err
	}
	return fmt.Errorf("%w: %v", target, err)
}
