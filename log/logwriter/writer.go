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

	currentColor     corelog.ANSIColorCode
	currentLevel     corelog.Level
	bufferedMessages []string
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

func (w *LogWriter) logMessage(chunk string, color corelog.ANSIColorCode, level corelog.Level) {
	if level != corelog.NormalLevel {
		chunk = removeColor(chunk, color)
	}
	w.logger.LogMessage(chunk, level)
}

// processLog identifies messages with log level in the incoming data stream.
func (w *LogWriter) processLog(chunk string) {
	if len(w.bufferedMessages) == 0 {
		// Start of a new message
		color := startColorCode(chunk)
		level, isMessageWithLevel := ansiEscapeCodeToLevel[color]

		if isMessageWithLevel {
			// New message with log level
			if hasColorResetSuffix(chunk) {
				// End of a message with log level
				w.logMessage(chunk, color, level)
			} else {
				// Start buffering the message
				w.currentColor = color
				w.currentLevel = level
				w.bufferedMessages = []string{chunk}
			}
		} else {
			// New message without a log level
			w.logMessage(chunk, "", corelog.NormalLevel)
		}
	} else {
		// Continuation of a message with potential log level
		if hasColorResetPrefix(chunk) {
			// End of message with newline and color reset at the end.

			w.logMessage(strings.Join(w.bufferedMessages, "\n"), w.currentColor, w.currentLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.bufferedMessages = nil

			// Chunk might contain a new message
			if trimColorResetPrefix(chunk) != "" {
				w.processLog(chunk)
			}
		} else if hasColorResetSuffix(chunk) {
			// End of a message with color reset at the end.
			w.bufferedMessages = append(w.bufferedMessages, chunk)
			w.logMessage(strings.Join(w.bufferedMessages, "\n"), w.currentColor, w.currentLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.bufferedMessages = nil
		} else {
			// Continue buffering the message
			w.bufferedMessages = append(w.bufferedMessages, chunk)
		}
	}
}

func (w *LogWriter) Close() error {
	if len(w.bufferedMessages) > 0 {
		w.logMessage(strings.Join(w.bufferedMessages, "\n"), w.currentColor, w.currentLevel)

		w.currentColor = ""
		w.currentLevel = ""
		w.bufferedMessages = nil
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

func trimColorResetPrefix(s string) string {
	return strings.TrimPrefix(s, string(corelog.ResetCode))
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
