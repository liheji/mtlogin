package feishu

import "errors"

// Config 是飞书通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool    `toml:"enabled"`
	UseSystemProxy bool    `toml:"use_system_proxy"`
	Timeout        float64 `toml:"timeout"`
	WebhookURL     string  `toml:"webhook_url"`
	Secret         string  `toml:"secret"`
}

// Validate 校验飞书启用时必须提供 webhook；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("feishu timeout must be >= 0")
	}
	if c.Enabled && c.WebhookURL == "" {
		return errors.New("feishu webhook_url is required")
	}
	return nil
}
