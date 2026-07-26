package feishu

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
	"strconv"
	"time"
)

type feishuSender interface {
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

// FeishuBot 实现飞书机器人通知通道。
type FeishuBot struct {
	WebhookURL string
	Secret     string
	sender     feishuSender
}

type feishuTextMessage struct {
	// 加签开启时，timestamp 与 sign 必须随 JSON body 发送（飞书自定义机器人规范）。
	Timestamp string `json:"timestamp,omitempty"`
	Sign      string `json:"sign,omitempty"`
	MsgType   string `json:"msg_type"`
	Content   struct {
		Text string `json:"text"`
	} `json:"content"`
}

func New(cfg Config, sender feishuSender) *FeishuBot {
	return &FeishuBot{WebhookURL: cfg.WebhookURL, Secret: cfg.Secret, sender: sender}
}

func (bot *FeishuBot) Name() string { return "feishu" }

func (bot *FeishuBot) Send(ctx *log.Context, title, body string) error {
	message := feishuTextMessage{MsgType: "text"}
	message.Content.Text = body
	if bot.Secret != "" {
		sign, timestamp := bot.generateSign()
		message.Sign = sign
		message.Timestamp = strconv.FormatInt(timestamp, 10)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}
	return bot.sendRequest(ctx, data)
}

func (bot *FeishuBot) sendRequest(ctx *log.Context, data []byte) error {
	status, body, err := bot.sender.PostJSON(ctx, bot.Name(), bot.WebhookURL, data, make(http.Header))
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("API返回错误: %s", log.SanitizeText(string(body)))
	}
	if len(body) > 0 {
		var response struct {
			Code int64  `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return fmt.Errorf("飞书响应解析失败: %w", err)
		}
		if response.Code != 0 {
			return fmt.Errorf("飞书业务返回错误 code=%d: %s", response.Code, log.SanitizeText(response.Msg))
		}
	}
	ctx.Debugf("feishu message sent response=%s", log.SanitizeText(string(body)))
	return nil
}

func (bot *FeishuBot) generateSign() (string, int64) {
	if bot.Secret == "" {
		return "", 0
	}
	timestamp := time.Now().Unix()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, bot.Secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write([]byte{})
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), timestamp
}
