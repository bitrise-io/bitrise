package corelog

import (
	"encoding/json"
	"fmt"
	"io"
)

type messageType string

const (
	logMessageType   messageType = "log"
	eventMessageType messageType = "event"
)

type messageLog struct {
	Timestamp   string      `json:"timestamp"`
	MessageType messageType `json:"type"`
	Producer    Producer    `json:"producer"`
	ProducerID  string      `json:"producer_id,omitempty"`
	Level       Level       `json:"level"`
	Message     string      `json:"message"`
}

type eventLog struct {
	Timestamp   string      `json:"timestamp"`
	MessageType messageType `json:"type"`
	EventType   string      `json:"event_type"`
	Content     interface{} `json:"content"`
}

type jsonLogger struct {
	encoder *json.Encoder
}

func newJSONLogger(output io.Writer) *jsonLogger {
	logger := jsonLogger{
		encoder: json.NewEncoder(output),
	}

	return &logger
}

// LogMessage ...
func (l *jsonLogger) LogMessage(message string, fields MessageLogFields) {
	msg := messageLog{
		MessageType: logMessageType,
		Message:     message,
		Timestamp:   fields.Timestamp,
		Producer:    fields.Producer,
		ProducerID:  fields.ProducerID,
		Level:       fields.Level,
	}
	err := l.encoder.Encode(msg)
	if err != nil {
		// Encountered an error during writing the json message to the output. Manually construct a json message for
		// the error and print it to the output
		fmt.Println(l.logMessageForError(err, fields.Timestamp, fmt.Sprintf("%#v", msg)))
	}
}

// LogEvent ...
func (l *jsonLogger) LogEvent(content interface{}, fields EventLogFields) {
	msg := eventLog{
		MessageType: eventMessageType,
		Content:     content,
		Timestamp:   fields.Timestamp,
		EventType:   fields.EventType,
	}
	err := l.encoder.Encode(msg)
	if err != nil {
		// Encountered an error during writing the json message to the output. Manually construct a json message for
		// the error and print it to the output
		fmt.Println(l.logMessageForError(err, fields.Timestamp, fmt.Sprintf("%#v", msg)))
	}
}

func (l *jsonLogger) logMessageForError(err error, timestamps, msg string) string {
	message := "{"
	message += fmt.Sprintf(`"timestamp":"%s",`, timestamps)
	message += fmt.Sprintf(`"type":"%s",`, string(logMessageType))
	message += fmt.Sprintf(`"producer":"%s",`, BitriseCLI)
	message += fmt.Sprintf(`"level":"%s",`, string(ErrorLevel))
	message += fmt.Sprintf(`"message":"log message (%s) serialization failed: %s"`, msg, err)
	message += "}"

	return message
}
