package telegram

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"mtlogin/pkg/log"
)

type errorSender struct {
	err error
}

func (s errorSender) PostJSON(context.Context, string, string, []byte, http.Header) (int, []byte, error) {
	return 0, nil, s.err
}

type captureHook struct {
	entries []string
}

func (h *captureHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *captureHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry.Message)
	return nil
}

func TestTelegramSendErrorDoesNotLogBotToken(t *testing.T) {
	logger := log.GetLog()
	hook := &captureHook{}
	oldHooks := logger.ReplaceHooks(logrus.LevelHooks{})
	defer logger.ReplaceHooks(oldHooks)
	logger.AddHook(hook)

	const token = "123456:ABCDEF-secret"
	bot := New(Config{BotToken: token, ChatID: 1}, errorSender{
		err: errors.New(`Get "https://api.telegram.org/bot123456:ABCDEF-secret/sendMessage": connection refused`),
	})
	if err := bot.Send(log.NewContext(context.Background()), "title", "body"); err == nil {
		t.Fatal("expected Telegram send error")
	}

	joined := strings.Join(hook.entries, "\n")
	if strings.Contains(joined, token) {
		t.Fatalf("Telegram bot token leaked in application log: %s", joined)
	}
	if !strings.Contains(joined, "[REDACTED]") {
		t.Fatalf("expected redacted Telegram error in application log: %s", joined)
	}
}
