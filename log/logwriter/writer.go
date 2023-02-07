package logwriter

import (
	"strings"
	"unicode"

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

type LogLevelWriter struct {
	logger log.Logger

	currentColor corelog.ANSIColorCode
	currentLevel corelog.Level
	currentChunk string
}

// NewLogLevelWriter ...
func NewLogLevelWriter(logger log.Logger) *LogLevelWriter {
	return &LogLevelWriter{
		logger: logger,
	}
}

// TODO: handle if currentChunk is too big
// TODO: handle frequent Writes (mux)
func (w *LogLevelWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	chunk := string(p)

	if string(w.currentColor) == "" {
		// Start of a new message
		color := startColorCode(chunk)
		level, ok := ansiEscapeCodeToLevel[color]

		isMessageWithLevel := ok
		if ok {
			raw := removeColor(chunk, color)
			if hasAnyColor(raw) {
				isMessageWithLevel = false
			}
		}

		if isMessageWithLevel {
			// New message with log level
			if hasColorResetSuffix(chunk) {
				// End of a message with log level
				raw := removeColor(chunk, color)
				w.logger.LogMessage(raw, level)
				return len(p), nil
			} else {
				// Message with log level might be written in multiple chunks
				w.currentColor = color
				w.currentLevel = level
				w.currentChunk = chunk
				return len(p), nil
			}
		} else {
			// New message without a log level
			w.logger.LogMessage(chunk, corelog.NormalLevel)
			return len(p), nil
		}
	} else {
		// Continuation of a message with potential log level
		if hasAnyColor(chunk) {
			chunk = w.currentChunk + chunk
			w.logger.LogMessage(chunk, corelog.NormalLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.currentChunk = ""

			return len(p), nil
		}

		if hasColorResetSuffix(chunk) {
			// End of a message with log level
			chunk = w.currentChunk + chunk

			raw := removeColor(chunk, w.currentColor)
			w.logger.LogMessage(raw, w.currentLevel)

			w.currentColor = ""
			w.currentLevel = ""
			w.currentChunk = ""

			return len(p), nil
		} else {
			// Message with log level might be written in multiple chunks
			w.currentChunk = w.currentChunk + chunk
			return len(p), nil
		}
	}
}

func (w *LogLevelWriter) Flush() (int, error) {
	if len(w.currentChunk) > 0 {
		return w.Write([]byte(w.currentChunk))
	}
	return 0, nil
}

func startColorCode(s string) corelog.ANSIColorCode {
	var colorCode corelog.ANSIColorCode
	for code := range ansiEscapeCodeToLevel {
		if strings.HasPrefix(s, string(code)) {
			colorCode = code
			break
		}
	}
	return colorCode
}

func hasAnyColor(s string) bool {
	for code := range ansiEscapeCodeToLevel {
		if strings.Contains(s, string(code)) {
			return true
		}
	}
	return false
}

func hasColorResetSuffix(s string) bool {
	trimmed := strings.TrimRightFunc(s, unicode.IsSpace)
	return strings.HasSuffix(trimmed, string(corelog.ResetCode))
}

func removeColor(s string, color corelog.ANSIColorCode) string {
	cleaned := strings.Replace(s, string(color), "", 1)
	return strings.Replace(cleaned, string(corelog.ResetCode), "", 1)
}
