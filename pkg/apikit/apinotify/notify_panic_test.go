package apinotify

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mtlogin/pkg/log"
)

type panicNotifier struct{ name string }

func (p panicNotifier) Name() string { return p.name }
func (p panicNotifier) Send(*log.Context, string, string) error {
	panic("boom from channel")
}

type okNotifier struct {
	name   string
	called bool
}

func (o *okNotifier) Name() string { return o.name }
func (o *okNotifier) Send(*log.Context, string, string) error {
	o.called = true
	return nil
}

func TestNotifyRecoversChannelPanic(t *testing.T) {
	ok := &okNotifier{name: "ok"}
	m := &Manager{}
	m.add(panicNotifier{name: "boom"}, 0, "")
	m.add(ok, 0, "")

	ctx := log.NewContext(context.Background())

	// A panicking channel must not crash the process; Notify must return the
	// panic as an aggregated error and still let sibling channels run.
	err := m.Notify(ctx, Message{Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("expected error from panicking channel, got nil")
	}
	if !strings.Contains(err.Error(), "boom") || !strings.Contains(err.Error(), "panic") {
		t.Errorf("error should mention channel and panic, got: %v", err)
	}
	var sendErr SendError
	if !errors.As(err, &sendErr) {
		t.Errorf("panic should surface as SendError, got: %v", err)
	}
	if !ok.called {
		t.Error("sibling channel must still be invoked despite the panic")
	}
}
