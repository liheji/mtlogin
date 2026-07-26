package ntfy

import "errors"

// Config 是 Ntfy 通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool    `toml:"enabled"`
	UseSystemProxy bool    `toml:"use_system_proxy"`
	Timeout        float64 `toml:"timeout"`
	URL            string  `toml:"url"`
	Topic          string  `toml:"topic"`
	Username       string  `toml:"username"`
	Password       string  `toml:"password"`
	Token          string  `toml:"token"`
}

// Validate 校验 Ntfy 启用时必须提供服务地址和主题；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("ntfy timeout must be >= 0")
	}
	if c.Enabled && (c.URL == "" || c.Topic == "") {
		return errors.New("ntfy url and topic are required")
	}
	return nil
}
