package apimtbrowse

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

func TestWarmupEndpointsAreStableAndUseSginKey(t *testing.T) {
	endpoints := WarmupEndpoints()
	want := []string{
		"/api/system/unix",
		"/ping",
		"/api/laboratory/funcState",
		"/api/fun/first",
		"/api/system/state",
		"/api/links/view",
		"/api/msg/statistic",
		"/api/member/profile",
		"/api/tracker/myPeerStatus",
		"/api/system/sysConf",
		"/api/news/list",
		"/api/torrent/teamList",
		"/api/system/promotion/rules",
		"/api/system/getConf",
		"/api/examine/getMyActiveList",
	}
	if len(endpoints) != len(want) {
		t.Fatalf("endpoint count: got %d want %d", len(endpoints), len(want))
	}
	for i := range want {
		if endpoints[i].Path != want[i] {
			t.Fatalf("endpoint[%d]: got %s want %s", i, endpoints[i].Path, want[i])
		}
	}
	body, _ := MultipartBody(map[string]string{"_timestamp": "1", "_sgin": "x"})
	if !strings.Contains(body, `name="_sgin"`) {
		t.Fatalf("multipart body must keep _sgin field, got %s", body)
	}
}

func TestWarmupAcceptsPlainTextPingResponse(t *testing.T) {
	handler := &stubHandler{responses: map[string]apicore.Response{
		"/ping": {
			Status:  http.StatusOK,
			Body:    "pong",
			Headers: map[string]string{},
		},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	err = client.NewSession(testState()).Warmup(log.NewContext(context.Background()))
	if err != nil {
		t.Fatalf("Warmup should accept the plain-text ping response, got %v", err)
	}
}
