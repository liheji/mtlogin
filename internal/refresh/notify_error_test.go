package refresh

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"mtlogin/pkg/log"
)

type errorNotifier struct {
	err error
}

func (n errorNotifier) Notify(*log.Context, Message) error {
	return n.err
}

type refreshLogHook struct {
	entries []string
}

func (h *refreshLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *refreshLogHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry.Message)
	return nil
}

func TestServiceNotificationFailureIsBestEffortAndLogged(t *testing.T) {
	hook := &refreshLogHook{}
	logger := log.GetLog()
	oldHooks := logger.ReplaceHooks(logrus.LevelHooks{})
	defer logger.ReplaceHooks(oldHooks)
	logger.AddHook(hook)

	service := newTestService(t, newFakeStore(), newScriptedHandler(), tokenMTConfig(), false, &fakeNotifier{})
	service.notifier = errorNotifier{err: errors.New("request failed: token=top-secret")}

	if err := service.Run(log.NewContext(context.Background())); err != nil {
		t.Fatalf("notification failure must not fail refresh run: %v", err)
	}
	joined := strings.Join(hook.entries, "\n")
	if !strings.Contains(joined, "notification failed") {
		t.Fatalf("expected notification failure to be logged: %s", joined)
	}
	if strings.Contains(joined, "top-secret") {
		t.Fatalf("notification error leaked sensitive value: %s", joined)
	}
}
