package readerutil

import (
	"bufio"
	"io"
	"strings"
)

// ReadLongLine - an alternative to bufio.Scanner.Scan,
// which can't handle long lines. This function is slower than
// bufio.Scanner.Scan, but can handle arbitrary long lines.
func ReadLongLine(r *bufio.Reader) (string, error) {
	// Do NOT create a `bufio.Reader` inside thise function,
	// get it as an input! (just in case you'd thing about doing a "revision" on this)
	// Creating the `bufio.Reader` here would reset/alter the reader,
	// if it would be created for every line! Not a good idea!

	isPrefix := true
	var err error
	var line, ln []byte

	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}

	return string(ln), err
}

// WalkLineFn - gets a line as its input
// if returns an error it stops the walk/reading
// to break the walk early, without an error, just return io.EOF
type WalkLineFn func(string) error

// WalkLines ...
func WalkLines(inpReader io.Reader, walkFn WalkLineFn) error {
	reader := bufio.NewReader(inpReader)

	var walkErr error
	line, readErr := ReadLongLine(reader)
	for ; walkErr == nil && readErr == nil; line, readErr = ReadLongLine(reader) {
		walkErr = walkFn(line)
	}

	// if walk returned an error (other than io.EOF)
	// return that error
	if walkErr != nil && walkErr != io.EOF {
		return walkErr
	}

	// otherwise, if there was a read error (except io.EOF), return that
	if readErr != nil && readErr != io.EOF {
		return readErr
	}

	return nil
}

// WalkLinesString ...
func WalkLinesString(inputStr string, walkFn WalkLineFn) error {
	return WalkLines(strings.NewReader(inputStr), walkFn)
}
