package refresh

import (
	"errors"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

// identityChanged 发送“身份状态已重置导致 Run 终止”的通知。
func (s *Service) identityChanged(ctx *log.Context) error {
	err := domain.RefreshIdentityChanged.WithCause(errors.New("identity config changed; current run aborted"))
	s.notify(ctx, buildMessage(messageIdentityChanged, err, nil))
	return err
}

// handlePasswordAuthExpired 只在账号密码模式下由 Run 调用（Token Auth 已在
// runWithSession 内单独通知并返回）。
func (s *Service) handlePasswordAuthExpired(runCtx, notifyCtx *log.Context, epoch domain.Epoch, authErr error) error {
	state, newEpoch, retried, resetErr := s.auth.HandleAuthExpired(runCtx, epoch)
	if resetErr != nil {
		// 身份清理失败时不能通知“本地身份已清除”，避免给出错误状态。
		// 必须通知用户，否则保活会静默失效。
		if errors.Is(resetErr, domain.RefreshIdentityResetFailed) {
			s.notify(notifyCtx, buildMessage(messageIdentityResetFailed, resetErr, nil))
		} else if errors.Is(resetErr, domain.RefreshIdentityChanged) {
			s.notify(notifyCtx, buildMessage(messageIdentityChanged, resetErr, nil))
		} else {
			s.notify(notifyCtx, buildMessage(messageReloginFailed, resetErr, nil))
		}
		return resetErr
	}
	notifyCtx.Warnf("refresh: auth expired, identity reset old_epoch=%d new_epoch=%d", epoch, newEpoch)

	if retried {
		notifyCtx.Info("refresh: retry_on_auth_failure=true, re-logging in this run")
		s.notify(notifyCtx, buildMessage(messagePasswordAuthExpiredRetry, authErr, nil))
		retryErr := s.runWithSession(runCtx, notifyCtx, state, newEpoch, false)
		// 重新登录成功但新会话再次认证失败：补发终态失败通知。
		// 此处不再 reset/重试，避免登录循环。
		if retryErr != nil && errors.Is(retryErr, domain.MTAuthFailed) {
			s.notify(notifyCtx, buildMessage(messageReloginFailed, retryErr, nil))
		}
		return retryErr
	}
	notifyCtx.Info("refresh: retry_on_auth_failure=false, will re-login on next scheduled run")
	s.notify(notifyCtx, buildMessage(messagePasswordAuthExpiredNextRun, authErr, nil))
	return authErr
}
