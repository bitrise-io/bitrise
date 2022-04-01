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

type errorFindingWriter interface {
	io.Writer
	findString(s string)
	getErrorMessage() *ErrorMessage
}

type defaultErrorFindingWriter struct {
	writer       io.Writer
	mux          sync.Mutex
	chunk        string
	collecting   bool
	errorMessage *ErrorMessage
}

func newWriter(writer io.Writer) errorFindingWriter {
	return &defaultErrorFindingWriter{writer: writer}
}

func (e *defaultErrorFindingWriter) Write(p []byte) (n int, err error) {
	e.mux.Lock()
	e.findString(string(p))
	e.mux.Unlock()
	return e.writer.Write(p)
}

func (e *defaultErrorFindingWriter) findString(s string) {
	haystack := e.chunk + s
	if e.collecting {
		if endIndex := getEndColorIndex(haystack); len(endIndex) > 0 {
			if endIndex[0] != 0 {
				e.errorMessage = &ErrorMessage{
					Timestamp: time.Now().UnixNano(),
					Message:   redRegexp.ReplaceAllString(haystack[0:endIndex[0]], ""),
				}
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

func (e *defaultErrorFindingWriter) getErrorMessage() *ErrorMessage {
	if e.collecting && e.chunk != "" {
		return &ErrorMessage{
			Timestamp: time.Now().UnixNano(),
			Message:   redRegexp.ReplaceAllString(e.chunk, ""),
		}
	}
	return e.errorMessage
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
