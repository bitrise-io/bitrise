package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Logger ...
type Logger interface {
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Printf(format string, v ...interface{})
	Donef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	TInfof(format string, v ...interface{})
	TWarnf(format string, v ...interface{})
	TPrintf(format string, v ...interface{})
	TDonef(format string, v ...interface{})
	TDebugf(format string, v ...interface{})
	TErrorf(format string, v ...interface{})
	Println()
	EnableDebugLog(enable bool)
}

const defaultTimeStampLayout = "15:04:05"

type logger struct {
	enableDebugLog  bool
	timestampLayout string
	stdout          io.Writer
}

// NewLogger ...
func NewLogger() Logger {
	return &logger{enableDebugLog: false, timestampLayout: defaultTimeStampLayout, stdout: os.Stdout}
}

// EnableDebugLog ...
func (l *logger) EnableDebugLog(enable bool) {
	l.enableDebugLog = enable
}

// Infof ...
func (l *logger) Infof(format string, v ...interface{}) {
	l.printf(infoSeverity, false, format, v...)
}

// Warnf ...
func (l *logger) Warnf(format string, v ...interface{}) {
	l.printf(warnSeverity, false, format, v...)
}

// Printf ...
func (l *logger) Printf(format string, v ...interface{}) {
	l.printf(normalSeverity, false, format, v...)
}

// Donef ...
func (l *logger) Donef(format string, v ...interface{}) {
	l.printf(doneSeverity, false, format, v...)
}

// Debugf ...
func (l *logger) Debugf(format string, v ...interface{}) {
	if l.enableDebugLog {
		l.printf(debugSeverity, false, format, v...)
	}
}

// Errorf ...
func (l *logger) Errorf(format string, v ...interface{}) {
	l.printf(errorSeverity, false, format, v...)
}

// TInfof ...
func (l *logger) TInfof(format string, v ...interface{}) {
	l.printf(infoSeverity, true, format, v...)
}

// TWarnf ...
func (l *logger) TWarnf(format string, v ...interface{}) {
	l.printf(warnSeverity, true, format, v...)
}

// TPrintf ...
func (l *logger) TPrintf(format string, v ...interface{}) {
	l.printf(normalSeverity, true, format, v...)
}

// TDonef ...
func (l *logger) TDonef(format string, v ...interface{}) {
	l.printf(doneSeverity, true, format, v...)
}

// TDebugf ...
func (l *logger) TDebugf(format string, v ...interface{}) {
	if l.enableDebugLog {
		l.printf(debugSeverity, true, format, v...)
	}
}

// TErrorf ...
func (l *logger) TErrorf(format string, v ...interface{}) {
	l.printf(errorSeverity, true, format, v...)
}

// Println ...
func (l *logger) Println() {
	fmt.Println()
}

func (l *logger) timestampField() string {
	currentTime := time.Now()
	return fmt.Sprintf("[%s]", currentTime.Format(l.timestampLayout))
}

func (l *logger) prefixCurrentTime(message string) string {
	return fmt.Sprintf("%s %s", l.timestampField(), message)
}

func (l *logger) createLogMsg(severity Severity, withTime bool, format string, v ...interface{}) string {
	colorFunc := severityColorFuncMap[severity]
	message := colorFunc(format, v...)
	if withTime {
		message = l.prefixCurrentTime(message)
	}

	return message
}

func (l *logger) printf(severity Severity, withTime bool, format string, v ...interface{}) {
	message := l.createLogMsg(severity, withTime, format, v...)
	if _, err := fmt.Fprintln(l.stdout, message); err != nil {
		fmt.Printf("failed to print message: %s, error: %s\n", message, err)
	}
}
