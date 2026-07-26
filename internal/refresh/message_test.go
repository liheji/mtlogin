package refresh

import (
	"errors"
	"strings"
	"testing"

	"mtlogin/internal/domain"
)

func TestBuildMessageDoesNotExposeErrorCause(t *testing.T) {
	err := domain.MTBusinessFailed.WithCause(errors.New("token=top-secret"))
	msg := buildMessage(messageRefreshFailed, err, nil)
	if strings.Contains(msg.Body, "top-secret") {
		t.Fatalf("notification exposed error cause: %q", msg.Body)
	}
	if !strings.Contains(msg.Body, domain.MTBusinessFailed.Message) {
		t.Fatalf("notification lost stable error message: %q", msg.Body)
	}
}

func TestBuildMessageRefreshSucceededIncludesProfile(t *testing.T) {
	msg := buildMessage(messageRefreshSucceeded, nil, testProfile())

	if msg.Title != "m-team 账号刷新成功" {
		t.Fatalf("title=%q", msg.Title)
	}
	if msg.Level != LevelSuccess {
		t.Fatalf("level=%q", msg.Level)
	}
	assertContainsAll(t, msg.Body, "m-team 账号 alice 刷新成功", "上传量: 2.00 Gb", "下载量: 3.00 Gb", "魔力值: 9.50", "上次登录时间: 2026-06-13", "刷新前上次访问时间: 2026-06-14")
	if strings.Contains(msg.Body, "上次刷新时间") {
		t.Fatalf("success message kept the misleading last refresh label: %q", msg.Body)
	}
}

func TestBuildMessageTokenAuthExpiredIncludesConfigKeys(t *testing.T) {
	err := domain.MTAuthFailed.WithCause(errors.New("need login"))
	msg := buildMessage(messageTokenAuthExpired, err, nil)

	if msg.Title != "m-team 登录保活失败" {
		t.Fatalf("title=%q", msg.Title)
	}
	if msg.Level != LevelError {
		t.Fatalf("level=%q", msg.Level)
	}
	assertContainsAll(t, msg.Body, "mt.token_auth.auth", "mt.token_auth.did", "mt.token_auth.visitor_id", domain.PublicMessage(err))
}

func TestBuildMessagePasswordAuthExpiredRetryAndNoRetry(t *testing.T) {
	err := domain.MTAuthFailed.WithCause(errors.New("need login"))

	retry := buildMessage(messagePasswordAuthExpiredRetry, err, nil)
	if retry.Title != "m-team 登录已过期" {
		t.Fatalf("retry title=%q", retry.Title)
	}
	assertContainsAll(t, retry.Body, domain.PublicMessage(err), "本地 token、DID、visitorid 已清除", "本次将立即重新登录")

	noRetry := buildMessage(messagePasswordAuthExpiredNextRun, err, nil)
	if noRetry.Title != "m-team 登录已过期" {
		t.Fatalf("no retry title=%q", noRetry.Title)
	}
	assertContainsAll(t, noRetry.Body, domain.PublicMessage(err), "本地 token、DID、visitorid 已清除", "下次刷新将重新登录")
}

func TestBuildMessageRefreshFailedIncludesProfileWhenAvailable(t *testing.T) {
	err := domain.MTHTTPStatus.WithCause(errors.New("http status=500"))
	profile := testProfile()

	msg := buildMessage(messageRefreshFailed, err, profile)

	if msg.Title != "m-team 登录保活失败" {
		t.Fatalf("title=%q", msg.Title)
	}
	if msg.Level != LevelError {
		t.Fatalf("level=%q", msg.Level)
	}
	assertContainsAll(t, msg.Body, "账号: alice", "上传量: 2.00 Gb", domain.PublicMessage(err))
}

func TestBuildMessageRefreshFailedHandlesNilProfile(t *testing.T) {
	err := errors.New("warmup failed")

	msg := buildMessage(messageRefreshFailed, err, nil)

	if msg.Body != domain.PublicMessage(err) {
		t.Fatalf("body=%q", msg.Body)
	}
}

func testProfile() *domain.Profile {
	return &domain.Profile{
		Username:        "alice",
		UploadedBytes:   2 * 1073741824,
		DownloadedBytes: 3 * 1073741824,
		Bonus:           9.5,
		LastLogin:       "2026-06-13",
		LastBrowse:      "2026-06-14",
	}
}

func assertContainsAll(t *testing.T, text string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(text, part) {
			t.Fatalf("expected %q to contain %q", text, part)
		}
	}
}
