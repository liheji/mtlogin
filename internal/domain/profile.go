package domain

// Profile 是刷新和通知层需要的 M-Team 用户资料快照。
//
// 字段保持业务含义，不暴露 M-Team 原始 JSON 的嵌套结构。
type Profile struct {
	Username        string
	UploadedBytes   int64
	DownloadedBytes int64
	Bonus           float64
	LastLogin       string
	LastBrowse      string
}
