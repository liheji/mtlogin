package domain

// LoginRequest 是账号密码模式发起登录所需的领域输入。
//
// VisitorID 必须由调用方先生成并传入，客户端不会隐式补齐。
type LoginRequest struct {
	Username   string
	Password   string
	TotpSecret string
	VisitorID  string
}
