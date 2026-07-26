package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const (
	defaultLogDir        = "logs"
	defaultAppLogName    = "app.log"
	defaultWarnLogName   = "app.logging.wf"
	defaultHTTPLogName   = "http.log"
	defaultNotifyLogName = "notify.log"
)

var appLog = newDiscardLogger()
var httpLog = newDiscardLogger()
var notifyLog = newDiscardLogger()

func init() {
	if err := os.MkdirAll(defaultLogDir, 0755); err != nil {
		appLog.Errorf("create log dir failed: %v", err)
		return
	}

	appLog.AddHook(&AppHook{Writer: rotatingLog(filepath.Join(defaultLogDir, defaultAppLogName))})
	appLog.AddHook(&WarnHook{Writer: rotatingLog(filepath.Join(defaultLogDir, defaultWarnLogName))})
	httpLog.AddHook(&AppHook{Writer: rotatingLog(filepath.Join(defaultLogDir, defaultHTTPLogName))})
	notifyLog.AddHook(&AppHook{Writer: rotatingLog(filepath.Join(defaultLogDir, defaultNotifyLogName))})
}

func GetLog() *logrus.Logger {
	return appLog
}

func GetHttpLog() *logrus.Logger {
	return httpLog
}

func GetNotifyLog() *logrus.Logger {
	return notifyLog
}

func Debug(msg string) { appLog.Debug(msg) }
func Info(msg string)  { appLog.Info(msg) }
func Warn(msg string)  { appLog.Warn(msg) }
func Error(msg string) { appLog.Error(msg) }

func Debugf(format string, args ...any) { appLog.Debugf(format, args...) }
func Infof(format string, args ...any)  { appLog.Infof(format, args...) }
func Warnf(format string, args ...any)  { appLog.Warnf(format, args...) }
func Errorf(format string, args ...any) { appLog.Errorf(format, args...) }

func HttpInfo(msg string) {
	httpLog.Info(msg)
}

func NotifyInfo(msg string) {
	notifyLog.Info(msg)
}
