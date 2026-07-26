package domain

// AuthState 是 M-Team 运行态身份三件套。
//
// Token 表示 M-Team Authorization header value；DID 和 VisitorID 是会话协议字段。
// 账号密码模式下该状态可以持久化到 LevelDB，Token Auth 模式下只能来自配置文件。
type AuthState struct {
	Token     string
	DID       string
	VisitorID string
}

type Epoch uint64

func (s AuthState) Empty() bool {
	return s.Token == "" && s.DID == "" && s.VisitorID == ""
}

func (s AuthState) Complete() bool {
	return s.Token != "" && s.DID != "" && s.VisitorID != ""
}
