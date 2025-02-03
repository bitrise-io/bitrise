package configmerge

import (
	"fmt"
	"time"

	"github.com/bitrise-io/bitrise/log"
)

const timestampLayout = "15:04:05"

type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type debugLogger struct {
	logger log.Logger
}

func newDebugLogger(logger log.Logger) Logger {
	return debugLogger{logger: logger}
}

func (l debugLogger) Debugf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logger.Debugf("[%s] %s", l.timestamp(), message)
}

func (l debugLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l debugLogger) timestamp() string {
	return time.Now().Format(timestampLayout)
}
