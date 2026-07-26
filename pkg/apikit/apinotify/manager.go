package apinotify

import (
	"errors"
	"fmt"
	"io"
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/apinotify/base"
	"mtlogin/pkg/apikit/apinotify/dingtalk"
	"mtlogin/pkg/apikit/apinotify/feishu"
	"mtlogin/pkg/apikit/apinotify/ntfy"
	"mtlogin/pkg/apikit/apinotify/qqpush"
	"mtlogin/pkg/apikit/apinotify/telegram"
	"mtlogin/pkg/apikit/apinotify/weixin"
	"mtlogin/pkg/log"
	"sync"
	"time"

	"mtlogin/pkg/util"
)

// Notifier 是单个通知通道的统一发送接口，由各原始通知包直接实现。
type Notifier interface {
	Name() string
	Send(ctx *log.Context, title, body string) error
}

type managedNotifier struct {
	notifier Notifier
	timeout  time.Duration
	proxy    string
}

// Manager 持有已启用的通知通道，并负责并发发送、超时控制和关闭协调。
type Manager struct {
	notifiers []managedNotifier
	timeout   time.Duration
	sender    Sender
	proxy     string
	mu        sync.Mutex
	wg        sync.WaitGroup
	closing   bool
}

type ManagerOption func(*Manager)

func SystemProxy(proxy string) ManagerOption {
	return func(m *Manager) {
		m.proxy = proxy
	}
}

// NewManager 根据启动期加载的通知配置构建通道，始终添加 base 兜底通知。
func NewManager(cfg Config, base apicore.RoundTripper, options ...ManagerOption) *Manager {
	rt := apicore.Chain(base, apicore.Logging(), apicore.Proxy())
	return newManager(cfg, apicore.NewClient(rt), options...)
}

func newManager(cfg Config, sender Sender, options ...ManagerOption) *Manager {
	timeout := 300 * time.Second
	if cfg.Timeout > 0 {
		timeout = util.SecondsDuration(cfg.Timeout)
	}
	m := &Manager{timeout: timeout, sender: sender}
	for _, option := range options {
		option(m)
	}
	m.add(base.New(), 0, "")
	if cfg.QQPush.Enabled {
		m.add(qqpush.New(cfg.QQPush, m.sender), cfg.QQPush.Timeout, m.systemProxyFor(cfg.QQPush.UseSystemProxy))
	}
	if cfg.Weixin.Enabled {
		m.add(weixin.New(cfg.Weixin, m.sender), cfg.Weixin.Timeout, m.systemProxyFor(cfg.Weixin.UseSystemProxy))
	}
	if cfg.DingTalk.Enabled {
		m.add(dingtalk.New(cfg.DingTalk, m.sender), cfg.DingTalk.Timeout, m.systemProxyFor(cfg.DingTalk.UseSystemProxy))
	}
	if cfg.Telegram.Enabled {
		m.add(telegram.New(cfg.Telegram, m.sender), cfg.Telegram.Timeout, m.telegramProxy(cfg.Telegram))
	}
	if cfg.Feishu.Enabled {
		m.add(feishu.New(cfg.Feishu, m.sender), cfg.Feishu.Timeout, m.systemProxyFor(cfg.Feishu.UseSystemProxy))
	}
	if cfg.Ntfy.Enabled {
		m.add(ntfy.New(cfg.Ntfy, m.sender), cfg.Ntfy.Timeout, m.systemProxyFor(cfg.Ntfy.UseSystemProxy))
	}
	return m
}

func (m *Manager) systemProxyFor(useSystemProxy bool) string {
	if useSystemProxy {
		return m.proxy
	}
	return ""
}

func (m *Manager) telegramProxy(cfg telegram.Config) string {
	if cfg.Proxy != "" {
		return cfg.Proxy
	}
	return m.systemProxyFor(cfg.UseSystemProxy)
}

func (m *Manager) add(n Notifier, timeoutSeconds float64, proxy string) {
	timeout := m.timeout
	if timeoutSeconds > 0 {
		timeout = util.SecondsDuration(timeoutSeconds)
	}
	m.notifiers = append(m.notifiers, managedNotifier{notifier: n, timeout: timeout, proxy: proxy})
}

// Notify 并发向所有通道发送消息。单通道失败会聚合返回，但不会阻断其他通道。
func (m *Manager) Notify(ctx *log.Context, msg Message) error {
	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return ErrManagerClosed
	}
	m.wg.Add(1)
	m.mu.Unlock()
	defer m.wg.Done()

	errCh := make(chan error, len(m.notifiers))
	var wg sync.WaitGroup
	for _, item := range m.notifiers {
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errCh <- SendError{Channel: item.notifier.Name(), Err: fmt.Errorf("panic: %v", r)}
				}
			}()
			sendCtx, cancel := ctx.WithTimeout(item.timeout)
			defer cancel()
			if item.proxy != "" {
				sendCtx = log.WrapContext(apicore.WithProxy(sendCtx, item.proxy), sendCtx)
			}
			if err := item.notifier.Send(sendCtx, msg.Title, msg.Body); err != nil {
				errCh <- SendError{Channel: item.notifier.Name(), Err: err}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Close 等待正在发送的通知结束，再关闭支持 io.Closer 的通道资源。
func (m *Manager) Close() error {
	m.mu.Lock()
	m.closing = true
	m.mu.Unlock()

	m.wg.Wait()

	var errs []error
	for _, item := range m.notifiers {
		closer, ok := item.notifier.(io.Closer)
		if !ok {
			continue
		}
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", item.notifier.Name(), err))
		}
	}
	return errors.Join(errs...)
}
