package runtime

import (
	"context"
	"sync"
	"sync/atomic"

	"mtlogin/internal/domain"
	"mtlogin/internal/refresh"

	"mtlogin/pkg/log"
)

// Runtime 是应用层长生命周期协调点，持有：
//   - 全局 Run 锁（同一时间最多一个刷新任务）
//   - 当前 refresh.Service 原子指针
type Runtime struct {
	rootCtx context.Context
	cancel  context.CancelFunc
	service atomic.Pointer[refresh.Service]
	runMu   sync.Mutex // 全局串行锁（TryLock）
	stopped atomic.Bool
	run     func(ctx *log.Context, svc *refresh.Service) error
}

func New(rootCtx context.Context) *Runtime {
	runCtx, cancel := context.WithCancel(rootCtx)
	return &Runtime{
		rootCtx: runCtx,
		cancel:  cancel,
		run: func(ctx *log.Context, svc *refresh.Service) error {
			return svc.Run(ctx)
		},
	}
}

// Publish 发布启动期构建的 refresh.Service，后续 TryRun 使用该 service。
func (r *Runtime) Publish(svc *refresh.Service) {
	r.service.Store(svc)
}

// TryRun 实现 scheduler.Runner 接口。全局串行：同时只允许一个刷新任务执行。
//
// delayCtx 仅在获取锁之前检查；一旦获取锁，使用 rootCtx 派生 runCtx，
// 确保 scheduler delayCtx 取消不会打断正在执行的 Run。
// 只有 SIGTERM/SIGINT 等进程根 context 取消才能取消正在执行的 Run。
func (r *Runtime) TryRun(delayCtx context.Context) error {
	if r.stopped.Load() {
		return context.Canceled
	}
	if err := delayCtx.Err(); err != nil {
		return err
	}
	if err := r.rootCtx.Err(); err != nil {
		return err
	}
	if !r.runMu.TryLock() {
		return domain.SchedulerBusy
	}
	defer func() {
		r.runMu.Unlock()
	}()
	if r.stopped.Load() {
		return context.Canceled
	}

	svc := r.service.Load()
	if svc == nil {
		return domain.RuntimeServiceNotPublished
	}

	// 从 rootCtx 派生 runCtx，不从 delayCtx 派生。
	runCtx := log.NewContext(r.rootCtx)
	runCtx.Debug("runtime: starting TryRun")
	childCtx, cancel := runCtx.WithCancel()
	defer cancel()

	return r.run(childCtx, svc)
}

func (r *Runtime) Stop() {
	if r == nil {
		return
	}
	if r.stopped.CompareAndSwap(false, true) {
		r.cancel()
	}
	r.runMu.Lock()
	r.runMu.Unlock()
}
