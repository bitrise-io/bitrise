package log

import (
	"fmt"
)

func printf(severity Severity, withTime bool, format string, v ...interface{}) {
	colorFunc := severityColorFuncMap[severity]
	message := colorFunc(format, v...)
	if withTime {
		message = fmt.Sprintf("%s %s", timestampField(), message)
	}

	fmt.Fprintln(outWriter, message)
}

// Successf ...
func Successf(format string, v ...interface{}) {
	printf(successSeverity, false, format, v...)
}

// Donef ...
func Donef(format string, v ...interface{}) {
	Successf(format, v...)
}

// Infof ...
func Infof(format string, v ...interface{}) {
	printf(infoSeverity, false, format, v...)
}

// Printf ...
func Printf(format string, v ...interface{}) {
	printf(normalSeverity, false, format, v...)
}

// Debugf ...
func Debugf(format string, v ...interface{}) {
	if enableDebugLog {
		printf(debugSeverity, false, format, v...)
	}
}

// Warnf ...
func Warnf(format string, v ...interface{}) {
	printf(warnSeverity, false, format, v...)
}

// Errorf ...
func Errorf(format string, v ...interface{}) {
	printf(errorSeverity, false, format, v...)
}

// TSuccessf ...
func TSuccessf(format string, v ...interface{}) {
	printf(successSeverity, true, format, v...)
}

// TDonef ...
func TDonef(format string, v ...interface{}) {
	TSuccessf(format, v...)
}

// TInfof ...
func TInfof(format string, v ...interface{}) {
	printf(infoSeverity, true, format, v...)
}

// TPrintf ...
func TPrintf(format string, v ...interface{}) {
	printf(normalSeverity, true, format, v...)
}

// TDebugf ...
func TDebugf(format string, v ...interface{}) {
	if enableDebugLog {
		printf(debugSeverity, true, format, v...)
	}
}

// TWarnf ...
func TWarnf(format string, v ...interface{}) {
	printf(warnSeverity, true, format, v...)
}

// TErrorf ...
func TErrorf(format string, v ...interface{}) {
	printf(errorSeverity, true, format, v...)
}
