package weixin

import (
	"context"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
)

type weixinSender interface {
	GetJSON(ctx context.Context, server, rawURL string, header http.Header) (int, []byte, error)
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

const (
	tokenURL   = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	sendMsgURL = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s"
)

// Weixin 实现企业微信通知通道。
type Weixin struct {
	CorpID      string
	AgentSecret string
	AgentID     int
	UserID      string
	sender      weixinSender
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func New(cfg Config, sender weixinSender) *Weixin {
	return &Weixin{CorpID: cfg.CorpID, AgentSecret: cfg.AgentSecret, AgentID: cfg.AgentID, UserID: cfg.UserID, sender: sender}
}

func (w *Weixin) Name() string { return "weixin" }

func (w *Weixin) Send(ctx *log.Context, title, body string) error {
	token, err := w.accessToken(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"touser":  w.UserID,
		"msgtype": "text",
		"text":    map[string]string{"content": body},
		"agentid": w.AgentID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	status, data, err := w.sender.PostJSON(ctx, w.Name(), fmt.Sprintf(sendMsgURL, token), data, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("weixin send status=%d: %s", status, string(data))
	}

	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	if response.ErrCode != 0 {
		return fmt.Errorf("发送消息失败: %s", response.ErrMsg)
	}
	return nil
}

func (w *Weixin) accessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", tokenURL, w.CorpID, w.AgentSecret)
	status, data, err := w.sender.GetJSON(ctx, w.Name(), url, nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("weixin token status=%d: %s", status, string(data))
	}

	var token tokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return "", err
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("获取 Access Token 失败")
	}
	return token.AccessToken, nil
}
