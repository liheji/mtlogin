package main

import (
	"context"
	"errors"
	"testing"
)

type fakeApplication struct {
	startErr error
	closeErr error
	closed   bool
}

func (f *fakeApplication) Start(ctx context.Context) error {
	return f.startErr
}

func (f *fakeApplication) RunOnce(ctx context.Context) error {
	return f.startErr
}

func (f *fakeApplication) Close() error {
	f.closed = true
	return f.closeErr
}

func TestExecuteClosesApplicationOnRunFailure(t *testing.T) {
	application := &fakeApplication{startErr: errors.New("run failed")}
	code := execute(context.Background(), false, func(context.Context) (applicationRunner, error) {
		return application, nil
	})
	if code != 1 {
		t.Fatalf("exit code=%d", code)
	}
	if !application.closed {
		t.Fatal("application was not closed")
	}
}

func TestExecuteReportsCloseFailure(t *testing.T) {
	application := &fakeApplication{closeErr: errors.New("close failed")}
	code := execute(context.Background(), true, func(context.Context) (applicationRunner, error) {
		return application, nil
	})
	if code != 1 || !application.closed {
		t.Fatalf("code=%d closed=%v", code, application.closed)
	}
}
