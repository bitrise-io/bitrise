package log

// UtilsLogAdapter extends the bitrise/log.Logger to meet the go-utils/v2/log.Logger interface.
type UtilsLogAdapter struct {
	debug bool
	Logger
}

func NewUtilsLogAdapter() UtilsLogAdapter {
	opts := GetGlobalLoggerOpts()
	return UtilsLogAdapter{
		Logger: NewLogger(opts),
		debug:  opts.DebugLogEnabled,
	}
}

func (l *UtilsLogAdapter) TInfof(format string, v ...interface{}) {
	Infof(format, v...)
}
func (l *UtilsLogAdapter) TWarnf(format string, v ...interface{}) {
	Warnf(format, v...)
}
func (l *UtilsLogAdapter) TPrintf(format string, v ...interface{}) {
	Printf(format, v...)
}
func (l *UtilsLogAdapter) TDonef(format string, v ...interface{}) {
	Donef(format, v...)
}
func (l *UtilsLogAdapter) TDebugf(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	Debugf(format, v...)
}
func (l *UtilsLogAdapter) TErrorf(format string, v ...interface{}) {
	Errorf(format, v...)
}
func (l *UtilsLogAdapter) Println() {
	Print()
}
func (l *UtilsLogAdapter) EnableDebugLog(enable bool) {
	l.debug = enable
}
