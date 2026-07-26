package telegram

import "errors"

// Config 是 Telegram 通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool    `toml:"enabled"`
	UseSystemProxy bool    `toml:"use_system_proxy"`
	Timeout        float64 `toml:"timeout"`
	BotToken       string  `toml:"bot_token"`
	ChatID         int64   `toml:"chat_id"`
	Proxy          string  `toml:"proxy"`
}

// Validate 校验 Telegram 启用时必须提供 bot token 和 chat id；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("telegram timeout must be >= 0")
	}
	if c.Enabled && (c.BotToken == "" || c.ChatID == 0) {
		return errors.New("telegram bot_token and chat_id are required")
	}
	return nil
}
