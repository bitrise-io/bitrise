package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// RFC3339Micro ...
const RFC3339Micro = "2006-01-02T15:04:05.999999Z07:00"

func defaultTimeProvider() time.Time {
	return time.Now()
}

type jsonLogger struct {
	debugLogEnabled bool
	encoder         *json.Encoder
	timeProvider    func() time.Time
}

func newJSONLogger(output io.Writer, provider func() time.Time) SimplifiedLogger {
	logger := jsonLogger{
		debugLogEnabled: false,
		encoder:         json.NewEncoder(output),
		timeProvider:    provider,
	}

	return &logger
}

// EnableDebugLog ...
func (j *jsonLogger) EnableDebugLog(enabled bool) {
	j.debugLogEnabled = enabled
}

// IsDebugLogEnabled ...
func (j *jsonLogger) IsDebugLogEnabled() bool {
	return j.debugLogEnabled
}

// LogMessage ...
func (j *jsonLogger) LogMessage(producer Producer, level Level, message string) {
	if !j.debugLogEnabled && level == DebugLevel {
		return
	}

	logMessage := logMessage{
		Timestamp:   j.timeProvider().Format(RFC3339Micro),
		MessageType: "log",
		Producer:    string(producer),
		Level:       string(level),
		Message:     message,
	}

	err := j.encoder.Encode(logMessage)
	if err != nil {
		// Encountered an error during writing the json message to the output. Manually construct a json message for
		// the error and print it to the output
		fmt.Println(j.logMessageForError(err))
	}
}

func (j *jsonLogger) logMessageForError(err error) string {
	message := "{"
	message += fmt.Sprintf("\"timestamp\":\"%s\",", j.timeProvider().Format(RFC3339Micro))
	message += "\"type\":\"log\","
	message += fmt.Sprintf("\"producer\":\"%s\",", string(CLI))
	message += fmt.Sprintf("\"level\":\"%s\",", string(ErrorLevel))
	message += fmt.Sprintf("\"message\":\"log message serialization failed: %s\"", err)
	message += "}"

	return message
}
