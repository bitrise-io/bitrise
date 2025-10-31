package log

import (
	"fmt"
)

func printf(severity Severity, withTime bool, format string, v ...interface{}) {
	message := createLogMsg(severity, withTime, format, v...)
	if _, err := fmt.Fprintln(outWriter, message); err != nil {
		fmt.Printf("failed to print message: %s, error: %s\n", message, err)
	}
}

func createLogMsg(severity Severity, withTime bool, format string, v ...interface{}) string {
	colorFunc := severityColorFuncMap[severity]
	message := colorFunc(format, v...)
	if withTime {
		message = prefixCurrentTime(message)
	}

	return message
}

func prefixCurrentTime(message string) string {
	return fmt.Sprintf("%s %s", timestampField(), message)
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

// RInfof ...
func RInfof(stepID string, tag string, data map[string]interface{}, format string, v ...interface{}) {
	rprintf("info", stepID, tag, data, format, v...)
}

// RWarnf ...
func RWarnf(stepID string, tag string, data map[string]interface{}, format string, v ...interface{}) {
	rprintf("warn", stepID, tag, data, format, v...)
}

// RErrorf ...
func RErrorf(stepID string, tag string, data map[string]interface{}, format string, v ...interface{}) {
	rprintf("error", stepID, tag, data, format, v...)
}
