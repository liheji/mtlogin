package log

import (
	"bytes"
	"encoding/json"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

type Message struct {
	kv  [][2]string
	msg string
	mu  sync.Mutex
}

func (msg *Message) AddField(key string, value any) {
	msg.mu.Lock()
	defer msg.mu.Unlock()
	for i, pair := range msg.kv {
		if pair[0] == key {
			msg.kv[i][1] = field2String(value)
			return
		}
	}
	msg.kv = append(msg.kv, [2]string{key, field2String(value)})
}

func (msg *Message) AddMsg(value any) {
	msg.mu.Lock()
	defer msg.mu.Unlock()
	msg.msg = field2String(value)
}

func (msg *Message) String() string {
	msg.mu.Lock()
	defer msg.mu.Unlock()

	out := ""
	for _, pair := range msg.kv {
		out += pair[0] + "[" + pair[1] + "] "
	}
	return out + msg.msg
}

func newDiscardLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(&LogFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(io.Discard)
	return logger
}

func rotatingLog(filename string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		LocalTime:  true,
	}
}

func field2String(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	default:
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(value); err != nil {
			return ""
		}
		return strings.TrimSpace(buf.String())
	}
}
