package analytics

import log "github.com/bitrise-io/bitrise/advancedlog"

type legacyLogger struct {
	debug bool
	log.Logger
}

func (l *legacyLogger) TInfof(format string, v ...interface{}) {
	log.Infof(format, v...)
}
func (l *legacyLogger) TWarnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}
func (l *legacyLogger) TPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
func (l *legacyLogger) TDonef(format string, v ...interface{}) {
	log.Donef(format, v...)
}
func (l *legacyLogger) TDebugf(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	log.Debugf(format, v...)
}
func (l *legacyLogger) TErrorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
func (l *legacyLogger) Println() {
	log.Print()
}
func (l *legacyLogger) EnableDebugLog(enable bool) {
	l.debug = enable
}
