package apimtbrowse

import (
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/internal/mtproto"
)

func parseResponse(res *apicore.Response, err error) (mtproto.APIResponse, error) {
	response, err := mtproto.ParseResponse(res, err)
	return response, translateProtocolError(err)
}
