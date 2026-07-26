package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type dingTalkSender interface {
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

// DingTalk 实现钉钉机器人通知通道。
type DingTalk struct {
	WebhookURL string
	Secret     string
	AtMobiles  []string
	sender     dingTalkSender
}

type dingTalkTextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	} `json:"at"`
}

func New(cfg Config, sender dingTalkSender) *DingTalk {
	return &DingTalk{
		WebhookURL: fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", cfg.WebhookToken),
		Secret:     cfg.Secret,
		AtMobiles:  cfg.AtMobiles,
		sender:     sender,
	}
}

func (d *DingTalk) Name() string { return "dingtalk" }

func (d *DingTalk) Send(ctx *log.Context, title, body string) error {
	timestamp, sign := d.generateSign()
	msg := dingTalkTextMessage{MsgType: "text"}
	msg.Text.Content = body
	msg.At.AtMobiles = d.AtMobiles
	msg.At.IsAtAll = len(d.AtMobiles) == 0

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message error: %w", err)
	}

	reqURL := signedWebhookURL(d.WebhookURL, timestamp, sign)
	status, responseBody, err := d.sender.PostJSON(ctx, d.Name(), reqURL, data, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("dingtalk robot return status=%d: %s", status, log.SanitizeText(string(responseBody)))
	}
	if len(responseBody) == 0 {
		return nil
	}
	var response struct {
		ErrCode int64  `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("dingtalk robot response decode failed: %w", err)
	}
	if response.ErrCode != 0 {
		return fmt.Errorf("dingtalk robot business error errcode=%d: %s", response.ErrCode, log.SanitizeText(response.ErrMsg))
	}
	return nil
}

func (d *DingTalk) generateSign() (string, string) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	stringToSign := timestamp + "\n" + d.Secret
	h := hmac.New(sha256.New, []byte(d.Secret))
	h.Write([]byte(stringToSign))
	return timestamp, base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func signedWebhookURL(webhookURL, timestamp, sign string) string {
	values := url.Values{}
	values.Set("timestamp", timestamp)
	values.Set("sign", sign)
	separator := "?"
	if strings.Contains(webhookURL, "?") {
		separator = "&"
	}
	return webhookURL + separator + values.Encode()
}
