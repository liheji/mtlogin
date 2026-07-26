package log

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogFormatter struct{}

func (m *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	var line string
	if entry.HasCaller() {
		line = fmt.Sprintf("%s %s %s:%d func[%s] %s\n",
			strings.ToUpper(entry.Level.String()),
			timestamp,
			entry.Caller.File,
			entry.Caller.Line,
			entry.Caller.Function,
			entry.Message,
		)
	} else {
		line = fmt.Sprintf("%s %s %s\n",
			strings.ToUpper(entry.Level.String()),
			timestamp,
			entry.Message,
		)
	}
	b.WriteString(line)
	return b.Bytes(), nil
}
