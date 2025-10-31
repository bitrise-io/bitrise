package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var outWriter io.Writer = os.Stdout

// SetOutWriter ...
func SetOutWriter(writer io.Writer) {
	outWriter = writer
}

var enableDebugLog = false

// SetEnableDebugLog ...
func SetEnableDebugLog(enable bool) {
	enableDebugLog = enable
}

var timestampLayout = "15:04:05"

// SetTimestampLayout ...
func SetTimestampLayout(layout string) {
	timestampLayout = layout
}

func timestampField() string {
	currentTime := time.Now()
	return fmt.Sprintf("[%s]", currentTime.Format(timestampLayout))
}
