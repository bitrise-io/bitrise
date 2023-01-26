package errorfinder

import (
	"io"
	"regexp"
	"sync"
	"time"
)

const maxLength = 20

var redRegexp = regexp.MustCompile(`\x1b\[[^m]*31[^m]*m`)
var controlRegexp = regexp.MustCompile(`\x1b\[[^m]*m`)

// ErrorMessage ...
type ErrorMessage struct {
	Timestamp int64
	Message   string
}

// ErrorFinder parses the data coming via the `Write` method and keeps the latest "red" block (that matches \x1b[31;1m control sequence)
// and hands over tha data to the wrapped `io.Writer` instance.
type ErrorFinder struct {
	mux               sync.Mutex
	writer            io.Writer
	timestampProvider TimestampProvider

	chunk         string
	collecting    bool
	errorMessages []ErrorMessage
}

type TimestampProvider interface {
	CurrentTimestamp() int64
}

type DefaultTimestampProvider struct {
}

func NewDefaultTimestampProvider() TimestampProvider {
	return DefaultTimestampProvider{}
}

func (p DefaultTimestampProvider) CurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

// NewErrorFinder ...
func NewErrorFinder(writer io.Writer, timestampProvider TimestampProvider) *ErrorFinder {
	return &ErrorFinder{
		writer:            writer,
		timestampProvider: timestampProvider,
	}
}

func (e *ErrorFinder) Write(p []byte) (n int, err error) {
	e.mux.Lock()
	e.findString(string(p))
	e.mux.Unlock()

	if e.writer != nil {
		return e.writer.Write(p)
	}
	return n, nil
}

func (e *ErrorFinder) GetErrorMessage() []ErrorMessage {
	if e.collecting && e.chunk != "" {
		e.errorMessages = append(e.errorMessages, ErrorMessage{
			Timestamp: e.timestampProvider.CurrentTimestamp(),
			Message:   redRegexp.ReplaceAllString(e.chunk, ""),
		})
		e.chunk = ""
		e.collecting = false
	}

	return e.errorMessages
}

func (e *ErrorFinder) findString(s string) {
	haystack := e.chunk + s
	if e.collecting {
		if endIndex := getEndColorIndex(haystack); len(endIndex) > 0 {
			if endIndex[0] != 0 {
				e.errorMessages = append(e.errorMessages, ErrorMessage{
					Timestamp: e.timestampProvider.CurrentTimestamp(),
					Message:   redRegexp.ReplaceAllString(haystack[0:endIndex[0]], ""),
				})
			}
			e.chunk = ""
			e.collecting = false
			if len(haystack) > endIndex[1] {
				e.findString(haystack[endIndex[1]:])
			}
		} else {
			e.chunk = haystack
		}
	} else {
		indices := redRegexp.FindStringIndex(haystack)
		if len(indices) > 0 {
			e.chunk = ""
			e.collecting = true
			if len(haystack) > indices[1] {
				e.findString(haystack[indices[1]:])
			}
		} else {
			if len(haystack) <= maxLength {
				e.chunk = haystack
			} else {
				e.chunk = haystack[len(haystack)-maxLength:]
			}
		}
	}
}

func getEndColorIndex(haystack string) []int {
	colorIndex := controlRegexp.FindStringIndex(haystack)
	if len(colorIndex) == 0 {
		return colorIndex
	}
	redIndices := redRegexp.FindStringIndex(haystack)
	if len(redIndices) == 0 || redIndices[0] > colorIndex[0] {
		return colorIndex
	}
	offset := redIndices[1]
	index := getEndColorIndex(haystack[offset:])
	if len(index) > 0 {
		index[0] += offset
		index[1] += offset
	}
	return index
}
