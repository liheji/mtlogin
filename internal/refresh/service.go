package refresh

import (
	"errors"
	"time"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

// Service 实现保活刷新的核心用例：认证 → 预热 → Profile → UpdateLastBrowse/BrowseForum → 持久化 → 通知。
type Service struct {
	timeout  time.Duration
	store    Store
	client   Client
	auth     AuthManager
	notifier Notifier
}

type Config struct {
	Timeout time.Duration
}

func NewService(cfg Config, client Client, auth AuthManager, store Store, notifier Notifier) *Service {
	return &Service{
		timeout:  cfg.Timeout,
		store:    store,
		client:   client,
		auth:     auth,
		notifier: notifier,
	}
}

// Run 执行一次完整的刷新周期。
//
// 认证模式分支：
//   - Token Auth（配置了 mt.token_auth.auth + mt.token_auth.did + mt.token_auth.visitor_id）：
//     完全使用配置文件中的三项身份信息，跳过登录，不读取或回写 LevelDB。失败不 fallback 到账号密码，不重试。
//   - 账号密码模式：认证失败时立即清空本地 token/DID/visitorid，并通知用户登录已过期；
//     是否在本次 Run 立即重新登录由 refresh.retry_on_auth_failure 控制。
func (s *Service) Run(appCtx *log.Context) error {
	// 业务请求阶段使用独立超时 context；通知 context 由调用方单独传入。
	cancel := func() {}
	runCtx := appCtx
	if s.timeout > 0 {
		runCtx, cancel = appCtx.WithTimeout(s.timeout)
	}
	defer cancel()

	state, epoch, tokenMode, err := s.auth.Resolve(runCtx)
	if err != nil {
		s.notify(appCtx, buildMessage(messageRefreshFailed, err, nil))
		return err
	}
	if tokenMode {
		appCtx.Info("refresh: Run started (token auth mode, identity from config)")
	} else {
		appCtx.Info("refresh: Run started (password mode)")
	}
	err = s.runWithSession(runCtx, appCtx, state, epoch, tokenMode)
	if err == nil {
		return nil
	}
	if tokenMode {
		return err
	}
	if errors.Is(err, domain.MTAuthFailed) {
		return s.handlePasswordAuthExpired(runCtx, appCtx, epoch, err)
	}
	return err
}

// runWithSession 执行 Warmup → Profile → Warmup → UpdateLastBrowse/BrowseForum 链路，
// 按账号密码 / Token Auth 两种身份来源处理状态持久化和通知。
//
// 账号密码模式的 MTAuthFailed 不在此方法内清理身份，
// 而是返回给 Run() 统一调用 ResetAuthState。
//
// runCtx 用于业务请求（有 refresh.timeout 超时），notifyCtx 用于通知（从进程根 context 派生，不受业务超时影响）。
func (s *Service) runWithSession(runCtx, notifyCtx *log.Context, state domain.AuthState, epoch domain.Epoch, tokenMode bool) error {
	session := s.client.NewSession(state)
	profile, err := runChecks(runCtx, session)
	snapshot := session.Snapshot()

	// 账号密码模式会持久化 Session 里更新后的 token/DID/visitorID；Token Auth 模式禁止回写配置身份。
	// 认证失败必须跳过写回，交由 Run 后续清理本地身份；普通请求失败仍需保留服务端轮换后的 DID。
	if !tokenMode && !errors.Is(err, domain.MTAuthFailed) && snapshot != state {
		saved, saveErr := s.store.SaveAuthState(snapshot, epoch)
		if saveErr != nil {
			return saveErr
		}
		if !saved {
			return s.identityChanged(notifyCtx)
		}
	}

	if err != nil {
		// Token Auth 认证失败：每次都通知用户更新配置，不做节流。
		if errors.Is(err, domain.MTAuthFailed) && tokenMode {
			s.notify(notifyCtx, buildMessage(messageTokenAuthExpired, err, nil))
			return err
		}
		// 账号密码认证失败：返回错误给 Run() 统一清理本地身份。
		if errors.Is(err, domain.MTAuthFailed) {
			return err
		}
		// 非认证错误（网络等）：仅通知，不清理本地身份。
		s.notify(notifyCtx, buildMessage(messageRefreshFailed, err, profile))
		return err
	}

	s.notify(notifyCtx, buildMessage(messageRefreshSucceeded, nil, profile))
	return nil
}

func (s *Service) notify(ctx *log.Context, msg Message) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.Notify(ctx, msg); err != nil {
		ctx.Warnf("refresh: notification failed: %s", log.SanitizeText(err.Error()))
	}
}
