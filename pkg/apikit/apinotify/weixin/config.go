package weixin

import "errors"

// Config 是企业微信通知通道自己的 TOML 配置。
type Config struct {
	Enabled        bool    `toml:"enabled"`
	UseSystemProxy bool    `toml:"use_system_proxy"`
	Timeout        float64 `toml:"timeout"`
	CorpID         string  `toml:"corp_id"`
	AgentSecret    string  `toml:"agent_secret"`
	AgentID        int     `toml:"agent_id"`
	UserID         string  `toml:"user_id"`
}

// Validate 校验企业微信启用时必须提供企业 ID 和应用密钥；timeout 单位为秒，0 表示继承全局配置。
func (c Config) Validate() error {
	if c.Timeout < 0 {
		return errors.New("weixin timeout must be >= 0")
	}
	if c.Enabled && (c.CorpID == "" || c.AgentSecret == "") {
		return errors.New("weixin corp_id and agent_secret are required")
	}
	return nil
}
