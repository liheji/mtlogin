package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type AppHook struct {
	Writer io.Writer
}

func (hook *AppHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *AppHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

type WarnHook struct {
	Writer io.Writer
}

func (hook *WarnHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (hook *WarnHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}
