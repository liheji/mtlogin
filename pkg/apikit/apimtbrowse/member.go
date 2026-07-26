package apimtbrowse

import (
	"net/http"

	"mtlogin/pkg/log"
)

func (s *Session) UpdateLastBrowse(ctx *log.Context) error {
	_, err := s.do(ctx, http.MethodPost, "/api/member/updateLastBrowse", nil)
	return err
}
