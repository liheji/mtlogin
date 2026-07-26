package log

import (
	"context"
	"time"

	"mtlogin/pkg/util"
)

type Context struct {
	context.Context
	logID string
}

func NewContext(parent context.Context) *Context {
	if parent == nil {
		parent = context.Background()
	}
	return &Context{Context: parent, logID: util.NewLogID()}
}

func WrapContext(parent context.Context, current *Context) *Context {
	if parent == nil {
		parent = context.Background()
	}
	if current == nil {
		return NewContext(parent)
	}
	return &Context{Context: parent, logID: current.logID}
}

func (c *Context) LogID() string {
	if c == nil {
		return ""
	}
	return c.logID
}

func (c *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return &Context{Context: ctx, logID: c.logID}, cancel
}

func (c *Context) WithTimeout(timeout time.Duration) (*Context, context.CancelFunc) {
	if timeout <= 0 {
		return c.WithCancel()
	}
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	return &Context{Context: ctx, logID: c.logID}, cancel
}

func (c *Context) WithValue(key, value any) *Context {
	if c == nil {
		return NewContext(context.WithValue(context.Background(), key, value))
	}
	ctx := context.WithValue(c.Context, key, value)
	return &Context{Context: ctx, logID: c.logID}
}

func (c *Context) formatPrefix() string {
	msg := &Message{}
	if logID := c.LogID(); logID != "" {
		msg.AddField("logid", logID)
	}
	return msg.String()
}

func (c *Context) Debug(msg string) { Debug(c.formatPrefix() + msg) }
func (c *Context) Info(msg string)  { Info(c.formatPrefix() + msg) }
func (c *Context) Warn(msg string)  { Warn(c.formatPrefix() + msg) }
func (c *Context) Error(msg string) { Error(c.formatPrefix() + msg) }

func (c *Context) Debugf(format string, args ...any) {
	Debugf(c.formatPrefix()+format, args...)
}

func (c *Context) Infof(format string, args ...any) {
	Infof(c.formatPrefix()+format, args...)
}

func (c *Context) Warnf(format string, args ...any) {
	Warnf(c.formatPrefix()+format, args...)
}

func (c *Context) Errorf(format string, args ...any) {
	Errorf(c.formatPrefix()+format, args...)
}
