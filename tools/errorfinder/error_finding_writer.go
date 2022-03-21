package errorfinder

import (
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	redControl = "\x1b[31;1m"
)

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
					Message:   strings.ReplaceAll(haystack[0:endIndex[0]], redControl, ""),
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
		index := strings.Index(haystack, redControl)
		if index != -1 {
			e.chunk = ""
			e.collecting = true
			if len(haystack) > index+len(redControl) {
				e.findString(haystack[index+len(redControl):])
			}
		} else {
			if len(haystack) <= len(redControl) {
				e.chunk = haystack
			} else {
				e.chunk = haystack[len(haystack)-len(redControl):]
			}
		}
	}
}

func (e *defaultErrorFindingWriter) getErrorMessage() *ErrorMessage {
	if e.collecting && e.chunk != "" {
		return &ErrorMessage{
			Timestamp: time.Now().UnixNano(),
			Message:   strings.ReplaceAll(e.chunk, redControl, ""),
		}
	}
	return e.errorMessage
}

func getEndColorIndex(haystack string) []int {
	colorIndex := controlRegexp.FindStringIndex(haystack)
	if len(colorIndex) == 0 {
		return colorIndex
	}
	redIndex := strings.Index(haystack, redControl)
	if redIndex == -1 || redIndex > colorIndex[0] {
		return colorIndex
	}
	offset := redIndex + len(redControl)
	index := getEndColorIndex(haystack[offset:])
	if len(index) > 0 {
		index[0] += offset
		index[1] += offset
	}
	return index
}
