package qqpush

import "errors"

// Config 是 QQPush 通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool    `toml:"enabled"`
	UseSystemProxy bool    `toml:"use_system_proxy"`
	Timeout        float64 `toml:"timeout"`
	QQ             string  `toml:"qq"`
	Token          string  `toml:"token"`
}

// Validate 校验 QQPush 启用时必须提供 QQ 和 token；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("qqpush timeout must be >= 0")
	}
	if c.Enabled && (c.QQ == "" || c.Token == "") {
		return errors.New("qqpush qq and token are required")
	}
	return nil
}
