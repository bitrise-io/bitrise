package logger

// Logger ...
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Done(args ...interface{})
	Donef(format string, args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}
