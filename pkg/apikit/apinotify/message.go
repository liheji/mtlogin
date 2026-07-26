package apinotify

// Level 表示通知语义级别，通道适配器可以据此映射颜色、优先级或标签。
type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// Message 是业务层和通知库之间唯一共享的消息结构。
// 新通道接入时应优先消费这些通用字段，避免把通道私有参数泄露到业务层。
type Message struct {
	Title    string
	Body     string
	Level    Level
	Tags     []string
	Metadata map[string]string
}
