package qqpush

import (
	"context"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
)

type qqPushSender interface {
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

// QQPush 实现 QQPush 通知通道。
type QQPush struct {
	QQ     string
	Token  string
	sender qqPushSender
}

func New(cfg Config, sender qqPushSender) *QQPush {
	return &QQPush{QQ: cfg.QQ, Token: cfg.Token, sender: sender}
}

func (q *QQPush) Name() string { return "qqpush" }

func (q *QQPush) Send(ctx *log.Context, title, body string) error {
	postURL := fmt.Sprintf("https://wx.scjtqs.com/qq/push/pushMsg?token=%s", q.Token)
	postData, err := json.Marshal(map[string]any{
		"qq": q.QQ,
		"content": []map[string]any{
			{
				"msgtype": "text",
				"text":    body,
			},
		},
		"token": q.Token,
	})
	if err != nil {
		return err
	}

	header := make(http.Header)
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:32.0) Gecko/20100101 Firefox/32.0")
	status, data, err := q.sender.PostJSON(ctx, q.Name(), postURL, postData, header)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("qqpush return status=%d: %s", status, string(data))
	}
	return nil
}
