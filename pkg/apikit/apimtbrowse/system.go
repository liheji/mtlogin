package apimtbrowse

import (
	"net/http"

	"mtlogin/pkg/log"
)

type WarmupEndpoint struct {
	Method string
	Path   string
	ExtMap map[string]string
}

func WarmupEndpoints() []WarmupEndpoint {
	return []WarmupEndpoint{
		{Method: http.MethodGet, Path: "/api/system/unix"},
		{Method: http.MethodGet, Path: "/ping"},
		{Method: http.MethodPost, Path: "/api/laboratory/funcState"},
		{Method: http.MethodPost, Path: "/api/fun/first"},
		{Method: http.MethodPost, Path: "/api/system/state"},
		{Method: http.MethodPost, Path: "/api/links/view"},
		{Method: http.MethodPost, Path: "/api/msg/statistic"},
		{Method: http.MethodPost, Path: "/api/member/profile"},
		{Method: http.MethodPost, Path: "/api/tracker/myPeerStatus"},
		{Method: http.MethodPost, Path: "/api/system/sysConf"},
		{Method: http.MethodPost, Path: "/api/news/list"},
		{Method: http.MethodPost, Path: "/api/torrent/teamList"},
		{Method: http.MethodPost, Path: "/api/system/promotion/rules"},
		{Method: http.MethodPost, Path: "/api/system/getConf", ExtMap: map[string]string{"items": "LeechWarn"}},
		{Method: http.MethodPost, Path: "/api/examine/getMyActiveList"},
	}
}

func (s *Session) Warmup(ctx *log.Context) error {
	for _, endpoint := range WarmupEndpoints() {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := s.do(ctx, endpoint.Method, endpoint.Path, endpoint.ExtMap)
		if err != nil {
			return err
		}
	}
	return nil
}
