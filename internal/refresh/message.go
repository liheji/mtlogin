package refresh

import (
	"fmt"

	"mtlogin/internal/domain"
)

type MessageLevel string

const (
	LevelSuccess MessageLevel = "success"
	LevelWarning MessageLevel = "warning"
	LevelError   MessageLevel = "error"
)

type Message struct {
	Title string
	Body  string
	Level MessageLevel
}

type messageKind int

const (
	messageRefreshFailed messageKind = iota
	messageRefreshSucceeded
	messageIdentityChanged
	messageTokenAuthExpired
	messagePasswordAuthExpiredRetry
	messagePasswordAuthExpiredNextRun
	messageReloginFailed
	messageIdentityResetFailed
)

func buildMessage(kind messageKind, err error, profile *domain.Profile) Message {
	errText := ""
	if err != nil {
		errText = domain.PublicMessage(err)
	}
	withPrefix := func(prefix, suffix string) string {
		if prefix == "" {
			return suffix
		}
		if suffix == "" {
			return prefix
		}
		return prefix + "\n" + suffix
	}

	switch kind {
	case messageRefreshSucceeded:
		return Message{
			Title: "m-team 账号刷新成功",
			Body:  formatProfile(profile, true),
			Level: LevelSuccess,
		}
	case messageIdentityChanged:
		return Message{
			Title: "m-team 刷新已终止",
			Body:  errText,
			Level: LevelWarning,
		}
	case messageTokenAuthExpired:
		return Message{
			Title: "m-team 登录保活失败",
			Body:  withPrefix(errText, "Token 认证信息可能已过期，请更新 conf/app.toml 中的 mt.token_auth.auth、mt.token_auth.did、mt.token_auth.visitor_id 后重启。"),
			Level: LevelError,
		}
	case messagePasswordAuthExpiredRetry:
		return Message{
			Title: "m-team 登录已过期",
			Body:  withPrefix(errText, "账号密码模式登录已过期，本地 token、DID、visitorid 已清除。\n本次将立即重新登录。"),
			Level: LevelWarning,
		}
	case messagePasswordAuthExpiredNextRun:
		return Message{
			Title: "m-team 登录已过期",
			Body:  withPrefix(errText, "账号密码模式登录已过期，本地 token、DID、visitorid 已清除。\n下次刷新将重新登录。"),
			Level: LevelWarning,
		}
	case messageReloginFailed:
		return Message{
			Title: "m-team 重新登录失败",
			Body:  withPrefix(errText, "账号密码模式登录已过期且重新登录失败，请检查账号、密码或 TOTP 配置。\n本地 token、DID、visitorid 已清除，下次刷新将再次尝试登录。"),
			Level: LevelError,
		}
	case messageIdentityResetFailed:
		return Message{
			Title: "m-team 登录状态清理失败",
			Body:  withPrefix(errText, "账号密码模式登录已过期，但本地 token、DID、visitorid 清理失败，无法确认当前身份状态。请检查本地存储权限后重试。"),
			Level: LevelError,
		}
	default:
		body := formatProfile(profile, false)
		if body != "" {
			body = withPrefix(body, errText)
		} else {
			body = errText
		}
		return Message{
			Title: "m-team 登录保活失败",
			Body:  body,
			Level: LevelError,
		}
	}
}

func formatProfile(profile *domain.Profile, success bool) string {
	if profile == nil {
		return ""
	}
	if success {
		return fmt.Sprintf(
			"m-team 账号 %s 刷新成功\n上传量: %.2f Gb\n下载量: %.2f Gb\n魔力值: %.2f\n上次登录时间: %s\n刷新前上次访问时间: %s",
			profile.Username, float64(profile.UploadedBytes)/1073741824, float64(profile.DownloadedBytes)/1073741824,
			profile.Bonus, profile.LastLogin, profile.LastBrowse,
		)
	}
	return fmt.Sprintf(
		"账号: %s\n上传量: %.2f Gb\n下载量: %.2f Gb\n魔力值: %.2f",
		profile.Username, float64(profile.UploadedBytes)/1073741824, float64(profile.DownloadedBytes)/1073741824,
		profile.Bonus,
	)
}
