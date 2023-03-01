package logwriter

import (
	"strings"
	"sync"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/corelog"
)

var ansiEscapeCodeToLevel = map[corelog.ANSIColorCode]corelog.Level{
	corelog.RedCode:     corelog.ErrorLevel,
	corelog.YellowCode:  corelog.WarnLevel,
	corelog.BlueCode:    corelog.InfoLevel,
	corelog.GreenCode:   corelog.DoneLevel,
	corelog.MagentaCode: corelog.DebugLevel,
}

type LogWriter struct {
	mux    sync.Mutex
	logger log.Logger

	currentColor corelog.ANSIColorCode
	currentLevel corelog.Level
	currentChunk string
}

// NewLogWriter ...
func NewLogWriter(logger log.Logger) *LogWriter {
	return &LogWriter{
		logger: logger,
	}
}

// TODO: handle if currentChunk is too big

func (w *LogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	w.mux.Lock()
	defer w.mux.Unlock()

	chunk := string(p)
	w.processLog(chunk)
	return len(p), nil
}

/*
	A message might start with the color code and end with the reset code:
	[34;1m[MSG_START_1]Login to the service[MSG_END_1][0m`

	or might end with a newline and reset code (because our log package adds a newline and then the color reset code):
	[34;1m[MSG_START_1]Login to the service[MSG_END_1]
	[0m

	this results in subsequent messages starting with a reset code:
	[34;1m[MSG_START_1]Login to the service[MSG_END_1]
	[0m[35;1m[MSG_START_2]detected login method:
	- API key
	- username (bitrise-bot@email.com)[MSG_END_2]
	[0m
*/

func (w *LogWriter) processLog(chunk string) {
	if string(w.currentColor) == "" {
		// Start of a new message
		color := startColorCode(chunk)
		level, isMessageWithLevel := ansiEscapeCodeToLevel[color]

		if isMessageWithLevel {
			// New message with log level
			if hasColorResetSuffix(chunk) {
				// End of a message with log level
				raw := removeColor(chunk, color)
				w.logger.LogMessage(raw, level)
			} else {
				// Message with log level might be written in multiple chunks
				w.currentColor = color
				w.currentLevel = level
				w.currentChunk = chunk
			}
		} else {
			// New message without a log level
			w.logger.LogMessage(chunk, corelog.NormalLevel)
		}
	} else {
		// Continuation of a message with potential log level
		if hasColorResetPrefix(chunk) {
			// Our log package adds a newline and then the color reset code for colored messages
			chunk = w.currentChunk
			raw := removeColor(chunk, w.currentColor)
			w.logger.LogMessage(raw, w.currentLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.currentChunk = ""

			w.processLog(chunk)
		} else if hasColorResetSuffix(chunk) {
			// End of a message with log level
			chunk = w.currentChunk + chunk
			raw := removeColor(chunk, w.currentColor)
			w.logger.LogMessage(raw, w.currentLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.currentChunk = ""
		} else {
			// Message with log level might be written in multiple chunks
			w.currentChunk = w.currentChunk + chunk
		}
	}
}

func (w *LogWriter) Close() error {
	if len(w.currentChunk) > 0 {
		_, err := w.Write([]byte(w.currentChunk))
		return err
	}
	return nil
}

func startColorCode(s string) corelog.ANSIColorCode {
	s = strings.TrimPrefix(s, string(corelog.ResetCode))

	var colorCode corelog.ANSIColorCode
	for code := range ansiEscapeCodeToLevel {
		if strings.HasPrefix(s, string(code)) {
			colorCode = code
			break
		}
	}
	return colorCode
}

func hasColorResetPrefix(s string) bool {
	return strings.HasPrefix(s, string(corelog.ResetCode))
}

func hasColorResetSuffix(s string) bool {
	return strings.HasSuffix(s, string(corelog.ResetCode))
}

func removeColor(s string, color corelog.ANSIColorCode) string {
	s = strings.TrimPrefix(s, string(corelog.ResetCode))
	s = strings.TrimPrefix(s, string(color))
	s = strings.TrimSuffix(s, string(corelog.ResetCode))
	return s
}
