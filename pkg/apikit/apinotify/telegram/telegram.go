package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"mtlogin/pkg/log"
	"net/http"
)

type telegramSender interface {
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}

// TelegramBot 实现 Telegram 通知通道。
type TelegramBot struct {
	BotToken string
	ChatID   int64
	sender   telegramSender
}

func New(cfg Config, sender telegramSender) *TelegramBot {
	return &TelegramBot{BotToken: cfg.BotToken, ChatID: cfg.ChatID, sender: sender}
}

func (t *TelegramBot) Name() string { return "telegram" }

func (t *TelegramBot) Send(ctx *log.Context, title, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{"chat_id": t.ChatID, "text": body})
	if err != nil {
		return err
	}
	status, response, err := t.sender.PostJSON(ctx, t.Name(), fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken), payload, nil)
	if err != nil {
		ctx.Errorf("telegram send failed: %s", log.SanitizeText(err.Error()))
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("telegram return status=%d: %s", status, log.SanitizeText(string(response)))
	}
	return nil
}
