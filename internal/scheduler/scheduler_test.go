package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSampleDelaySeconds(t *testing.T) {
	for i := 0; i < 100; i++ {
		d := SampleDelay(time.Now().UnixNano(), Config{MinDelay: time.Second, MaxDelay: 2 * time.Second})
		if d < time.Second || d > 2*time.Second {
			t.Fatalf("delay out of range: %s", d)
		}
	}
	if d := SampleDelay(1, Config{}); d != 0 {
		t.Fatalf("zero delay should execute immediately, got %s", d)
	}
	if d := SampleDelay(1, Config{MinDelay: 500 * time.Millisecond, MaxDelay: 500 * time.Millisecond}); d != 500*time.Millisecond {
		t.Fatalf("expected sub-second delay, got %s", d)
	}
}

func TestNewUsesSchedulerConfig(t *testing.T) {
	r := &runnerStub{}
	s, err := New(Config{Crontab: "2 */2 * * *", MinDelay: time.Second, MaxDelay: 2 * time.Second}, r)
	if err != nil {
		t.Fatal(err)
	}
	if s.cfg.Crontab != "2 */2 * * *" || s.cfg.MinDelay != time.Second || s.cfg.MaxDelay != 2*time.Second {
		t.Fatalf("unexpected scheduler config: %+v", s.cfg)
	}
}

func TestRunWithDelaySkipsRunnerWhenContextCanceledBeforeDelayExpires(t *testing.T) {
	r := &countingRunner{}
	s := &Scheduler{
		cfg:    Config{MinDelay: time.Second, MaxDelay: time.Second},
		runner: r,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.runWithDelay(ctx)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("runWithDelay should return promptly when context is canceled")
	}
	if r.calls.Load() != 0 {
		t.Fatalf("runner should not be called after delay context is canceled")
	}
}

func TestStopWaitsForRunningCronJob(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	runner := runnerFunc(func(ctx context.Context) error {
		close(started)
		<-release
		return nil
	})
	s, err := New(Config{Crontab: "@every 1ms"}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("cron job did not start")
	}
	stopped := make(chan struct{})
	go func() {
		s.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("Stop returned before running cron job completed")
	case <-time.After(10 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return after cron job completed")
	}
}

type runnerStub struct{}

func (r *runnerStub) TryRun(ctx context.Context) error {
	return nil
}

type runnerFunc func(context.Context) error

func (f runnerFunc) TryRun(ctx context.Context) error {
	return f(ctx)
}

type countingRunner struct {
	calls atomic.Int32
}

func (r *countingRunner) TryRun(ctx context.Context) error {
	r.calls.Add(1)
	return nil
}
