package log

// DefaultLogger ...
type DefaultLogger struct {
	ts bool
}

// NewDefaultLogger ...
func NewDefaultLogger(withTimestamp bool) DefaultLogger {
	return DefaultLogger{withTimestamp}
}

// Donef ...
func (dl DefaultLogger) Donef(format string, v ...interface{}) {
	fSelect(dl.ts, TDonef, Donef)(format, v...)
}

// Successf ...
func (dl DefaultLogger) Successf(format string, v ...interface{}) {
	fSelect(dl.ts, TSuccessf, Successf)(format, v...)
}

// Infof ...
func (dl DefaultLogger) Infof(format string, v ...interface{}) {
	fSelect(dl.ts, TInfof, Infof)(format, v...)
}

// Printf ...
func (dl DefaultLogger) Printf(format string, v ...interface{}) {
	fSelect(dl.ts, TPrintf, Printf)(format, v...)
}

// Warnf ...
func (dl DefaultLogger) Warnf(format string, v ...interface{}) {
	fSelect(dl.ts, TWarnf, Warnf)(format, v...)
}

// Errorf ...
func (dl DefaultLogger) Errorf(format string, v ...interface{}) {
	fSelect(dl.ts, TErrorf, Errorf)(format, v...)
}

// Debugf ...
func (dl DefaultLogger) Debugf(format string, v ...interface{}) {
	if enableDebugLog {
		fSelect(dl.ts, TDebugf, Debugf)(format, v...)
	}
}

type logfunc func(string, ...interface{})

func fSelect(t bool, tf logfunc, f logfunc) logfunc {
	if t {
		return tf
	}
	return f
}
