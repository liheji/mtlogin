package domain

// Error code convention: BBBEEE, where BBB is the business line and EEE starts
// from 000 within that line.
var (
	ConfigReadFailed           = Error{Code: 100000, Message: "读取配置失败"}
	ConfigParseFailed          = Error{Code: 100001, Message: "解析配置失败"}
	ConfigInvalid              = Error{Code: 100002, Message: "配置无效"}
	MTAuthFailed               = Error{Code: 200000, Message: "M-Team 认证失败"}
	MTHTTPStatus               = Error{Code: 200001, Message: "M-Team HTTP 状态异常"}
	MTBusinessFailed           = Error{Code: 200002, Message: "M-Team 业务失败"}
	MTRequestBuildFailed       = Error{Code: 200003, Message: "M-Team 请求构建失败"}
	MTTOTPRequired             = Error{Code: 200004, Message: "M-Team 需要二次验证码"}
	RefreshIdentityChanged     = Error{Code: 400000, Message: "刷新身份已变更"}
	RefreshIdentityResetFailed = Error{Code: 400001, Message: "刷新身份清理失败"}
	SchedulerBusy              = Error{Code: 500000, Message: "刷新任务正在运行"}
	RuntimeServiceNotPublished = Error{Code: 500001, Message: "刷新服务未发布"}
)
