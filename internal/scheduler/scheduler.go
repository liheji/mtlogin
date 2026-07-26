package scheduler

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

// Runner 由 runtime.Runtime 实现，内部持有全局 Run 锁。
type Runner interface {
	TryRun(delayCtx context.Context) error
}

type Config struct {
	Crontab  string
	MinDelay time.Duration
	MaxDelay time.Duration
}

type Scheduler struct {
	cfg      Config
	runner   Runner
	cron     *cron.Cron
	cancel   context.CancelFunc
	mu       sync.Mutex
	started  bool
	stopping bool
	stopDone chan struct{}
}

func New(cfg Config, runner Runner) (*Scheduler, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(cfg.Crontab); err != nil {
		return nil, err
	}
	return &Scheduler{cfg: cfg, runner: runner, cron: cron.New(cron.WithParser(parser)), stopDone: make(chan struct{})}, nil
}

// SampleDelay 返回 [min_delay, max_delay] 秒范围内的均匀随机延迟。
// min_delay 和 max_delay 均为 0 时返回 0（立即执行）。
func SampleDelay(seed int64, cfg Config) time.Duration {
	minDelay := cfg.MinDelay
	maxDelay := cfg.MaxDelay
	if minDelay == 0 && maxDelay == 0 {
		return 0
	}
	if maxDelay <= minDelay {
		return minDelay
	}
	r := rand.New(rand.NewSource(seed))
	return minDelay + time.Duration(r.Float64()*float64(maxDelay-minDelay))
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if s.stopping {
		return errors.New("scheduler is stopped")
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	if _, err := s.cron.AddFunc(s.cfg.Crontab, func() {
		s.runWithDelay(runCtx)
	}); err != nil {
		cancel()
		return err
	}
	s.cron.Start()
	s.started = true
	return nil
}

func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.stopping {
		done := s.stopDone
		s.mu.Unlock()
		<-done
		return
	}
	s.stopping = true
	if s.cancel != nil {
		s.cancel()
	}
	stopCtx := s.cron.Stop()
	done := s.stopDone
	s.mu.Unlock()

	<-stopCtx.Done()
	s.mu.Lock()
	s.started = false
	close(done)
	s.mu.Unlock()
}

// runWithDelay 先等待随机延迟（如有），再调用 Runner.TryRun。
// 延迟期间 ctx 被取消（进程退出或 scheduler 停止）则跳过本次执行。
func (s *Scheduler) runWithDelay(ctx context.Context) {
	delay := SampleDelay(time.Now().UnixNano(), s.cfg)
	if delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
	log.Info("scheduler: delay expired, starting refresh run")
	if err := s.runner.TryRun(ctx); err != nil {
		switch {
		case errors.Is(err, domain.SchedulerBusy):
			log.Info("scheduler: previous run still in progress, skipping")
		case errors.Is(err, context.Canceled):
			log.Info("scheduler: run canceled during shutdown")
		default:
			log.Warnf("scheduler: refresh run failed: %v", err)
		}
	}
}
