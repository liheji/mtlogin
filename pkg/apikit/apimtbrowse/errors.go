package apimtbrowse

import (
	"errors"
	"fmt"
	"strings"

	"mtlogin/pkg/apikit/internal/mtproto"
)

var (
	ErrAuthFailed   = errors.New("mteam authentication failed")
	ErrHTTPStatus   = errors.New("mteam http status error")
	ErrBusiness     = errors.New("mteam business error")
	ErrTOTPRequired = errors.New("mteam totp required")
	ErrRequestBuild = errors.New("mteam request build failed")
)

func translateProtocolError(err error) error {
	if err == nil {
		return nil
	}
	var target, source error
	switch {
	case errors.Is(err, mtproto.ErrAuthFailed):
		target = ErrAuthFailed
		source = mtproto.ErrAuthFailed
	case errors.Is(err, mtproto.ErrHTTPStatus):
		target = ErrHTTPStatus
		source = mtproto.ErrHTTPStatus
	case errors.Is(err, mtproto.ErrTOTPRequired):
		target = ErrTOTPRequired
		source = mtproto.ErrTOTPRequired
	case errors.Is(err, mtproto.ErrBusiness):
		target = ErrBusiness
		source = mtproto.ErrBusiness
	default:
		return err
	}
	detail := strings.TrimPrefix(err.Error(), source.Error())
	detail = strings.TrimPrefix(detail, ": ")
	if detail == "" {
		return target
	}
	return fmt.Errorf("%w: %s", target, detail)
}
