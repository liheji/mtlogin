package apinotify

import (
	"errors"
	"fmt"

	"mtlogin/pkg/apikit/apinotify/dingtalk"
	"mtlogin/pkg/apikit/apinotify/feishu"
	"mtlogin/pkg/apikit/apinotify/ntfy"
	"mtlogin/pkg/apikit/apinotify/qqpush"
	"mtlogin/pkg/apikit/apinotify/telegram"
	"mtlogin/pkg/apikit/apinotify/weixin"
	"mtlogin/pkg/conf"
)

const NotifyConfigPath = "conf/notify.toml"

// Config 对应固定文件 conf/notify.toml 的完整配置。
type Config struct {
	Timeout  float64         `toml:"timeout"`
	QQPush   qqpush.Config   `toml:"qqpush"`
	Weixin   weixin.Config   `toml:"weixin"`
	DingTalk dingtalk.Config `toml:"dingtalk"`
	Telegram telegram.Config `toml:"telegram"`
	Feishu   feishu.Config   `toml:"feishu"`
	Ntfy     ntfy.Config     `toml:"ntfy"`
}

// Validate 做通道最小必填项校验，避免运行期才暴露 token/webhook 缺失。
func (c Config) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("notify timeout must be > 0")
	}
	validators := []func() error{
		c.QQPush.Validate,
		c.Weixin.Validate,
		c.DingTalk.Validate,
		c.Telegram.Validate,
		c.Feishu.Validate,
		c.Ntfy.Validate,
	}
	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) Clone() *Config {
	next := c
	if len(c.DingTalk.AtMobiles) > 0 {
		next.DingTalk.AtMobiles = append([]string(nil), c.DingTalk.AtMobiles...)
	}
	return &next
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{Timeout: 10}
	if err := conf.LoadTOML(path, cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate notify config: %w", err)
	}
	return cfg.Clone(), nil
}
