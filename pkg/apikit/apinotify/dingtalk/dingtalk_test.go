package dingtalk

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"mtlogin/pkg/log"
)

func TestSignedWebhookURLEscapesSignQueryValue(t *testing.T) {
	raw := signedWebhookURL("https://oapi.dingtalk.com/robot/send?access_token=token", "123", "+/=")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.Query().Get("sign"); got != "+/=" {
		t.Fatalf("sign query value was not preserved after escaping: %q", got)
	}
	if raw == "https://oapi.dingtalk.com/robot/send?access_token=token&timestamp=123&sign=+/=" {
		t.Fatalf("sign query value must be escaped in raw URL: %s", raw)
	}
}

type responseSender struct {
	status int
	body   []byte
}

func (s responseSender) PostJSON(context.Context, string, string, []byte, http.Header) (int, []byte, error) {
	return s.status, s.body, nil
}

func TestDingTalkBusinessErrorOnHTTP200(t *testing.T) {
	bot := New(Config{WebhookToken: "token"}, responseSender{
		status: http.StatusOK,
		body:   []byte(`{"errcode":310000,"errmsg":"token is invalid"}`),
	})

	err := bot.Send(log.NewContext(context.Background()), "title", "body")
	if err == nil {
		t.Fatal("expected DingTalk business error for non-zero errcode")
	}
	if got := err.Error(); !containsAll(got, "errcode=310000", "token is invalid") {
		t.Fatalf("unexpected DingTalk error: %v", err)
	}
}

func TestDingTalkBusinessSuccessOnHTTP200(t *testing.T) {
	bot := New(Config{WebhookToken: "token"}, responseSender{
		status: http.StatusOK,
		body:   []byte(`{"errcode":0,"errmsg":"ok"}`),
	})

	if err := bot.Send(log.NewContext(context.Background()), "title", "body"); err != nil {
		t.Fatalf("expected DingTalk success, got %v", err)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
