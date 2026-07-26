package dingtalk

import "errors"

// Config 是钉钉机器人通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool     `toml:"enabled"`
	UseSystemProxy bool     `toml:"use_system_proxy"`
	Timeout        float64  `toml:"timeout"`
	WebhookToken   string   `toml:"webhook_token"`
	Secret         string   `toml:"secret"`
	AtMobiles      []string `toml:"at_mobiles"`
}

// Validate 校验钉钉启用时必须提供机器人 token；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("dingtalk timeout must be >= 0")
	}
	if c.Enabled && c.WebhookToken == "" {
		return errors.New("dingtalk webhook_token is required")
	}
	return nil
}
