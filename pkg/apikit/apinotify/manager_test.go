package apinotify

import "testing"

func TestNewManagerAlwaysAddsBaseNotifier(t *testing.T) {
	m := newManager(Config{Timeout: 10}, nil)
	if len(m.notifiers) != 1 {
		t.Fatalf("expected base notifier, got %d", len(m.notifiers))
	}
	if got := m.notifiers[0].notifier.Name(); got != "base" {
		t.Fatalf("base notifier name=%q", got)
	}
}
