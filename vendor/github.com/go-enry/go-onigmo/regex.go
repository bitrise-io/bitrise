package onigmo

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lonigmo
#include <stdlib.h>
#include "chelper.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"
)

const numMatchStartSize = 4

var mutex sync.Mutex

var _ compliance = &Regexp{}

// Regexp is the representation of a compiled regular expression. A Regexp is
// safe for concurrent use by multiple goroutines.
type Regexp struct {
	pattern  string
	options  Option
	encoding Encoding
	syntax   Syntax

	regex     C.OnigRegex
	errorInfo *C.OnigErrorInfo
	errorBuf  *C.char

	numSubexp         int
	subexpNames       []string
	idxSubexpNames    map[string]int
	hasMetacharacters bool
}

// NewRegexp creates and initializes a new Regexp with the given pattern and option.
func NewRegexp(pattern string, encoding Encoding, options Option, syntax Syntax) (*Regexp, error) {
	re := &Regexp{
		pattern:  pattern,
		encoding: encoding,
		options:  options,
		syntax:   syntax,
	}

	runtime.SetFinalizer(re, (*Regexp).Free)
	return re, re.initRegexp()
}

func (re *Regexp) initRegexp() error {
	patternCharPtr := C.CString(re.pattern)
	defer C.free(unsafe.Pointer(patternCharPtr))

	mutex.Lock()
	defer mutex.Unlock()

	errorCode := C.NewOnigRegex(patternCharPtr, C.int(len(re.pattern)), C.int(re.options), &re.regex, &re.encoding, re.syntax, &re.errorInfo, &re.errorBuf)
	if errorCode != 0 {
		return errors.New(C.GoString(re.errorBuf))
	}

	re.numSubexp = int(C.onig_number_of_captures(re.regex))
	re.hasMetacharacters = QuoteMeta(re.pattern) != re.pattern

	return re.loadSubexpNames()
}

func (re *Regexp) loadSubexpNames() error {
	count := int(C.onig_number_of_names(re.regex))
	if count == 0 {
		return nil
	}

	bufferSize := len(re.pattern) * 2
	nameBuffer := make([]byte, bufferSize)
	groupNumbers := make([]int32, count)
	bufferPtr := unsafe.Pointer(&nameBuffer[0])
	numbersPtr := unsafe.Pointer(&groupNumbers[0])

	length := int(C.GetCaptureNames(re.regex, bufferPtr, (C.int)(bufferSize), (*C.int)(numbersPtr)))
	if length == 0 {
		return fmt.Errorf("could not get the capture group names")
	}

	re.subexpNames = strings.Split(string(nameBuffer[:length]), ";")
	if len(re.subexpNames) != count {
		return fmt.Errorf(
			"unexpected number of capture group names, got %d, expected %d,",
			len(re.subexpNames), count,
		)
	}

	re.idxSubexpNames = make(map[string]int, len(groupNumbers))
	for i, idx := range groupNumbers {
		re.idxSubexpNames[re.subexpNames[i]] = int(idx)
	}

	return nil
}

// Compile parses a regular expression and returns, if successful, a Regexp
// object that can be used to match against text. The encoding is set to UTF8
// and the systax is set to Perl 5.10+ which is the most compatible with Go.
func Compile(str string) (*Regexp, error) {
	return NewRegexp(str, EncodingUTF8, OptionNone, SyntaxPerl)
}

// MustCompile is like Compile but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled
// regular expressions.
func MustCompile(str string) *Regexp {
	regexp, error := Compile(str)
	if error != nil {
		panic("regexp: compiling " + str + ": " + error.Error())
	}

	return regexp
}

// Free release all the cgo resource used by the regexp. This function it's
// used as finalizer the Regexp.
func (re *Regexp) Free() {
	mutex.Lock()
	if re.regex != nil {
		C.onig_free(re.regex)
		re.regex = nil
	}
	mutex.Unlock()
	if re.errorInfo != nil {
		C.free(unsafe.Pointer(re.errorInfo))
		re.errorInfo = nil
	}
	if re.errorBuf != nil {
		C.free(unsafe.Pointer(re.errorBuf))
		re.errorBuf = nil
	}
}

func (re *Regexp) find(b []byte, n int, offset int) []int {
	if len(re.pattern) == 0 && len(b) == 0 {
		return make([]int, (re.numSubexp+1)*2)
	}

	match := make([]int, (re.numSubexp+1)*2)

	if n == 0 {
		b = []byte{0}
	}

	bytesPtr := unsafe.Pointer(&b[0])

	// captures contains two pairs of ints, start and end, so we need list
	// twice the size of the capture groups.
	captures := make([]C.int, (re.numSubexp+1)*2)
	capturesPtr := unsafe.Pointer(&captures[0])

	var numCaptures int32
	numCapturesPtr := unsafe.Pointer(&numCaptures)

	pos := int(C.SearchOnigRegex(
		bytesPtr, C.int(n), C.int(offset), C.int(OptionNone),
		re.regex, re.errorInfo, (*C.char)(nil), (*C.int)(capturesPtr), (*C.int)(numCapturesPtr),
	))

	if pos < 0 {
		return nil
	}

	if numCaptures <= 0 {
		panic("cannot have 0 captures when processing a match")
	}

	for i := range captures {
		match[i] = int(captures[i])
	}

	return match
}

func (re *Regexp) match(b []byte, n int, offset int) bool {
	if n == 0 {
		b = []byte{0}
	}

	bytesPtr := unsafe.Pointer(&b[0])
	pos := int(C.SearchOnigRegex(
		bytesPtr, C.int(n), C.int(offset), C.int(OptionNone),
		re.regex, re.errorInfo, nil, nil, nil,
	))

	return pos >= 0
}

func (re *Regexp) findAll(b []byte, n int) [][]int {
	if n < 0 {
		n = len(b)
	}

	capture := make([][]int, 0, numMatchStartSize)
	var offset int
	for offset <= n {
		match := re.find(b, n, offset)
		if match == nil {
			break
		}

		capture = append(capture, match)

		// move offset to the ending index of the current match and prepare to
		// find the next non-overlapping match.
		offset = match[1]

		// if match[0] == match[1], it means the current match does not advance
		// the search. we need to exit the loop to avoid getting stuck here.
		if match[0] == match[1] {
			if offset < n && offset >= 0 {
				//there are more bytes, so move offset by a word
				_, width := utf8.DecodeRune(b[offset:])
				offset += width
			} else {
				//search is over, exit loop
				break
			}
		}
	}

	return capture
}

// NumSubexp returns the number of parenthesized subexpressions in this Regexp.
func (re *Regexp) NumSubexp() int {
	return int(re.numSubexp)
}

// SubexpNames returns the names of the parenthesized subexpressions
// in this Regexp. The name for the first sub-expression is names[1],
// so that if m is a match slice, the name for m[i] is SubexpNames()[i].
// Since the Regexp as a whole cannot be named, names[0] is always
// the empty string. The slice should not be modified.
func (re *Regexp) SubexpNames() []string {
	return re.subexpNames
}

func (re *Regexp) String() string {
	return re.pattern
}

// LiteralPrefix it's not implemented. Panics when it called.
func (re *Regexp) LiteralPrefix() (prefix string, complete bool) {
	panic("function not implemented")
	return "", false
}

// Copy returns a new Regexp object copied from re.
func (re *Regexp) Copy() *Regexp {
	copy, _ := NewRegexp(re.pattern, re.encoding, re.options, re.syntax)
	return copy
}

// Longest makes future searches prefer the leftmost-longest match.
// That is, when matching against text, the regexp returns a match that
// begins as early as possible in the input (leftmost), and among those
// it chooses a match that is as long as possible.
// This method modifies the Regexp and may not be called concurrently
// with any other methods.
func (re *Regexp) Longest() {
	re.options = re.options | OptionFindLongest
	re.initRegexp()
}
