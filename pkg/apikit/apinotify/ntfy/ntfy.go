package ntfy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
)

type ntfySender interface {
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

// Ntfy 实现 Ntfy 通知通道。
type Ntfy struct {
	URL    string
	Topic  string
	Auth   string
	sender ntfySender
}

type ntfyTextMessage struct {
	Topic    string       `json:"topic"`
	Title    string       `json:"title"`
	Message  string       `json:"message"`
	Priority int          `json:"priority"`
	Tags     []string     `json:"tags,omitempty"`
	Actions  []ntfyAction `json:"actions,omitempty"`
	Attach   string       `json:"attach,omitempty"`
	Filename string       `json:"filename,omitempty"`
	Click    string       `json:"click,omitempty"`
}

type ntfyAction struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
}

func New(cfg Config, sender ntfySender) *Ntfy {
	return &Ntfy{URL: cfg.URL, Topic: cfg.Topic, Auth: auth(cfg.Username, cfg.Password, cfg.Token), sender: sender}
}

func (n *Ntfy) Name() string { return "ntfy" }

func (n *Ntfy) Send(ctx *log.Context, title, body string) error {
	message := ntfyTextMessage{
		Topic:    n.Topic,
		Title:    "M-Team 登录保活通知",
		Message:  body,
		Priority: 1,
		Tags:     []string{"bread"},
		Actions: []ntfyAction{
			{
				Action: "view",
				Label:  "手动登录",
				URL:    "https://kp.m-team.cc/",
			},
		},
		Click: "https://kp.m-team.cc/",
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}
	return n.sendRequest(ctx, data)
}

func (n *Ntfy) sendRequest(ctx context.Context, data []byte) error {
	header := make(http.Header)
	if n.Auth != "" {
		header.Set("Authorization", n.Auth)
	}

	status, body, err := n.sender.PostJSON(ctx, n.Name(), n.URL, data, header)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("API返回错误: %s", string(body))
	}
	return nil
}

func auth(user, password, token string) string {
	if token != "" {
		return "Bearer " + token
	}
	if user != "" && password != "" {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
	}
	return ""
}
