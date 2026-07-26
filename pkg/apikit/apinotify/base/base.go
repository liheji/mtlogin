package base

import "mtlogin/pkg/log"

// Base 是常开的兜底通知通道，将所有通知写入本地通知日志。
type Base struct{}

func New() *Base {
	return &Base{}
}

func (b *Base) Name() string { return "base" }

func (b *Base) Send(ctx *log.Context, title, body string) error {
	msg := &log.Message{}
	if ctx != nil && ctx.LogID() != "" {
		logID := ctx.LogID()
		msg.AddField("logid", logID)
	}
	msg.AddField("notify", b.Name())
	msg.AddField("title", title)
	msg.AddMsg(body)
	log.NotifyInfo(msg.String())
	return nil
}
