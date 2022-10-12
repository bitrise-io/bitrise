package analytics

import (
	"github.com/bitrise-io/bitrise/log"
)

// utilsLogAdapter extends the bitrise/log.Logger to meet the go-utils/v2/log.Logger interface.
type utilsLogAdapter struct {
	debug bool
	log.Logger
}

func newUtilsLogAdapter() utilsLogAdapter {
	opts := log.GetGlobalLoggerOpts()
	return utilsLogAdapter{
		Logger: log.NewLogger(opts),
		debug:  opts.DebugLogEnabled,
	}
}

func (l *utilsLogAdapter) TInfof(format string, v ...interface{}) {
	log.Infof(format, v...)
}
func (l *utilsLogAdapter) TWarnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}
func (l *utilsLogAdapter) TPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
func (l *utilsLogAdapter) TDonef(format string, v ...interface{}) {
	log.Donef(format, v...)
}
func (l *utilsLogAdapter) TDebugf(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	log.Debugf(format, v...)
}
func (l *utilsLogAdapter) TErrorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
func (l *utilsLogAdapter) Println() {
	log.Print()
}
func (l *utilsLogAdapter) EnableDebugLog(enable bool) {
	l.debug = enable
}
