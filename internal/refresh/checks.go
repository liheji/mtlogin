package refresh

import (
	"context"
	"errors"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

// runChecks 串行执行 Warmup → Profile → Warmup → UpdateLastBrowse → BrowseForum。
// Warmup 失败短路后续；Profile 失败短路第二轮 Warmup 和后续步骤；
// 只要首次 UpdateLastBrowse 或 BrowseForum 内部任一 UpdateLastBrowse 成功，本次保活即成功。
func runChecks(ctx *log.Context, session Session) (*domain.Profile, error) {
	if err := session.Warmup(ctx); err != nil {
		return nil, err
	}
	profile, err := session.Profile(ctx)
	if err != nil {
		return nil, err
	}
	// 旧版在 Profile 成功后、UpdateLastBrowse 前会再次执行 funcState 预热。
	// 这里保留同样的请求节奏，避免 lastBrowse 更新前的浏览器态缺失。
	if err := session.Warmup(ctx); err != nil {
		return profile, err
	}
	updateErr := session.UpdateLastBrowse(ctx)
	// 页面加载完成后继续模拟论坛浏览；内部会按页面刷新节奏再次尝试 UpdateLastBrowse。
	forumUpdated, forumErr := session.BrowseForum(ctx)
	if forumErr != nil && shouldPropagateForumError(forumErr) {
		return profile, forumErr
	}
	if updateErr != nil && !forumUpdated {
		return profile, updateErr
	}
	return profile, nil
}

func shouldPropagateForumError(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, domain.MTAuthFailed)
}
