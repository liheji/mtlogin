package feishu

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"mtlogin/pkg/log"
)

type captureSender struct {
	body           []byte
	header         http.Header
	responseStatus int
	responseBody   []byte
}

func (c *captureSender) PostJSON(_ context.Context, _, _ string, body []byte, header http.Header) (int, []byte, error) {
	c.body = append([]byte(nil), body...)
	c.header = header
	status := c.responseStatus
	if status == 0 {
		status = http.StatusOK
	}
	response := c.responseBody
	if response == nil {
		response = []byte(`{"code":0}`)
	}
	return status, response, nil
}

// 飞书自定义机器人加签要求 timestamp 与 sign 放在 JSON body，而非 HTTP 头。
func TestFeishuSignatureGoesIntoBody(t *testing.T) {
	sender := &captureSender{}
	bot := New(Config{WebhookURL: "https://open.feishu.cn/hook/x", Secret: "s3cr3t"}, sender)

	if err := bot.Send(log.NewContext(context.Background()), "t", "hello"); err != nil {
		t.Fatal(err)
	}

	var payload map[string]any
	if err := json.Unmarshal(sender.body, &payload); err != nil {
		t.Fatalf("body not valid JSON: %v", err)
	}
	if payload["sign"] == nil || payload["sign"] == "" {
		t.Errorf("sign must be in body, body=%s", sender.body)
	}
	if payload["timestamp"] == nil || payload["timestamp"] == "" {
		t.Errorf("timestamp must be in body, body=%s", sender.body)
	}
	if payload["msg_type"] != "text" {
		t.Errorf("msg_type should remain text, body=%s", sender.body)
	}
	// 不得再放到 HTTP 头（那是入站事件订阅回调用的）。
	if sender.header.Get("X-Lark-Signature") != "" || sender.header.Get("X-Lark-Request-Timestamp") != "" {
		t.Errorf("signature must not be sent as headers, got %v", sender.header)
	}
}

// 未配置 secret 时不应带 sign/timestamp 字段。
func TestFeishuNoSignatureWithoutSecret(t *testing.T) {
	sender := &captureSender{}
	bot := New(Config{WebhookURL: "https://open.feishu.cn/hook/x"}, sender)

	if err := bot.Send(log.NewContext(context.Background()), "t", "hello"); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(sender.body, &payload); err != nil {
		t.Fatalf("body not valid JSON: %v", err)
	}
	if _, ok := payload["sign"]; ok {
		t.Errorf("sign must be absent without secret, body=%s", sender.body)
	}
	if _, ok := payload["timestamp"]; ok {
		t.Errorf("timestamp must be absent without secret, body=%s", sender.body)
	}
}

func TestFeishuBusinessErrorOnHTTP200(t *testing.T) {
	sender := &captureSender{responseBody: []byte(`{"code":19001,"msg":"token invalid"}`)}
	bot := New(Config{WebhookURL: "https://open.feishu.cn/hook/x"}, sender)

	err := bot.Send(log.NewContext(context.Background()), "t", "hello")
	if err == nil {
		t.Fatal("expected Feishu business error for non-zero code")
	}
	if !strings.Contains(err.Error(), "code=19001") || !strings.Contains(err.Error(), "token invalid") {
		t.Fatalf("unexpected Feishu error: %v", err)
	}
}
