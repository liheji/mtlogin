package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"mtlogin/internal/domain"
	"mtlogin/internal/refresh"
	"mtlogin/pkg/log"
)

func TestTryRunWithoutPublishedServiceReturnsErrno(t *testing.T) {
	rt := New(context.Background())

	err := rt.TryRun(context.Background())
	if !errors.Is(err, domain.RuntimeServiceNotPublished) {
		t.Fatalf("expected RuntimeServiceNotPublished, got %v", err)
	}
}

func TestTryRunReturnsSchedulerBusyWhenRunIsInProgress(t *testing.T) {
	rt := New(context.Background())
	release := make(chan struct{})
	started := make(chan struct{})
	rt.Publish(&refresh.Service{})
	rt.run = func(ctx *log.Context, svc *refresh.Service) error {
		close(started)
		<-release
		return nil
	}
	done := make(chan error, 1)
	go func() {
		done <- rt.TryRun(context.Background())
	}()
	<-started

	err := rt.TryRun(context.Background())
	if !errors.Is(err, domain.SchedulerBusy) {
		t.Fatalf("expected SchedulerBusy, got %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first run should complete without error, got %v", err)
	}
}

func TestTryRunChecksDelayContextBeforeTakingRunLock(t *testing.T) {
	rt := New(context.Background())
	rt.Publish(&refresh.Service{})
	var calls atomic.Int32
	rt.run = func(ctx *log.Context, svc *refresh.Service) error {
		calls.Add(1)
		return nil
	}
	delayCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rt.TryRun(delayCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("run should not be called when delay context is already canceled")
	}
}

func TestTryRunUsesRootContextAfterRunStarts(t *testing.T) {
	rootCtx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()
	rt := New(rootCtx)
	rt.Publish(&refresh.Service{})
	delayCtx, cancelDelay := context.WithCancel(context.Background())
	var runErr error
	rt.run = func(ctx *log.Context, svc *refresh.Service) error {
		cancelDelay()
		select {
		case <-ctx.Done():
			runErr = ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
		return nil
	}

	if err := rt.TryRun(delayCtx); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("started run should not be canceled by delay context: %v", runErr)
	}
}

func TestStopCancelsAndWaitsForRunningTask(t *testing.T) {
	rt := New(context.Background())
	rt.Publish(&refresh.Service{})
	started := make(chan struct{})
	finished := make(chan struct{})
	rt.run = func(ctx *log.Context, svc *refresh.Service) error {
		close(started)
		<-ctx.Done()
		close(finished)
		return ctx.Err()
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- rt.TryRun(context.Background())
	}()
	<-started
	rt.Stop()
	select {
	case <-finished:
	default:
		t.Fatal("Stop returned before the running task finished")
	}
	if !errors.Is(<-runDone, context.Canceled) {
		t.Fatal("running task was not canceled")
	}
	if err := rt.TryRun(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("stopped runtime accepted a new run: %v", err)
	}
}
