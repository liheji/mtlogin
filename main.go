package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"mtlogin/internal/app"
	"mtlogin/pkg/log"
)

type applicationRunner interface {
	Start(context.Context) error
	RunOnce(context.Context) error
	Close() error
}

type applicationFactory func(context.Context) (applicationRunner, error)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return execute(ctx, os.Getenv("LOCAL_TEST_RUN") == "true", func(ctx context.Context) (applicationRunner, error) {
		return app.New(ctx)
	})
}

func execute(ctx context.Context, localRun bool, factory applicationFactory) (code int) {
	application, err := factory(ctx)
	if err != nil {
		log.Errorf("init app failed: %v", err)
		return 1
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Errorf("close app failed: %v", err)
			code = 1
		}
	}()

	if localRun {
		if err := application.RunOnce(ctx); err != nil {
			log.Errorf("local test run failed: %v", err)
			return 1
		}
		return 0
	}
	if err := application.Start(ctx); err != nil {
		log.Errorf("app stopped with error: %v", err)
		return 1
	}
	return 0
}
