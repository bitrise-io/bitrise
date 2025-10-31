package log

// DummyLogger ...
type DummyLogger struct{}

// NewDummyLogger ...
func NewDummyLogger() DummyLogger {
	return DummyLogger{}
}

// Donef ...
func (dl DummyLogger) Donef(format string, v ...interface{}) {}

// Successf ...
func (dl DummyLogger) Successf(format string, v ...interface{}) {}

// Infof ...
func (dl DummyLogger) Infof(format string, v ...interface{}) {}

// Printf ...
func (dl DummyLogger) Printf(format string, v ...interface{}) {}

// Debugf ...
func (dl DummyLogger) Debugf(format string, v ...interface{}) {}

// Warnf ...
func (dl DummyLogger) Warnf(format string, v ...interface{}) {}

// Errorf ...
func (dl DummyLogger) Errorf(format string, v ...interface{}) {}
