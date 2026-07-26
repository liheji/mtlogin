package apimtbrowse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mtlogin/pkg/log"
)

func (s *Session) Profile(ctx *log.Context) (*Profile, error) {
	resp, err := s.do(ctx, http.MethodPost, "/api/member/profile", nil)
	if err != nil {
		return nil, err
	}
	var data ProfileData
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return nil, fmt.Errorf("%w: profile data decode: %v", ErrBusiness, err)
		}
	}
	return &Profile{
		Username:        data.Username,
		UploadedBytes:   data.MemberCount.Uploaded,
		DownloadedBytes: data.MemberCount.Downloaded,
		Bonus:           data.MemberCount.Bonus,
		LastLogin:       data.MemberStatus.LastLogin,
		LastBrowse:      data.MemberStatus.LastBrowse,
	}, nil
}
