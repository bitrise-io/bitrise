package corelog

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// RFC3339Micro ...
const RFC3339Micro = "2006-01-02T15:04:05.999999Z07:00"

type jsonLogger struct {
	encoder      *json.Encoder
	timeProvider func() time.Time
}

func newJSONLogger(output io.Writer, timeProvider func() time.Time) *jsonLogger {
	logger := jsonLogger{
		encoder:      json.NewEncoder(output),
		timeProvider: timeProvider,
	}

	return &logger
}

// LogMessage ...
func (j *jsonLogger) LogMessage(producer Producer, level Level, message string) {
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
	message += fmt.Sprintf("\"producer\":\"%s\",", string(BitriseCLI))
	message += fmt.Sprintf("\"level\":\"%s\",", string(ErrorLevel))
	message += fmt.Sprintf("\"message\":\"log message serialization failed: %s\"", err)
	message += "}"

	return message
}
