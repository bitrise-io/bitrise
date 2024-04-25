package onigmo

import (
	"bytes"
	"io"
	"unicode/utf8"
)

// MatchString reports whether the string s contains any match of the regular expression re.
func MatchString(pattern string, s string) (matched bool, error error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(s), nil
}

// Match reports whether the byte slice b contains any match of the regular
// expression re.
func (re *Regexp) Match(b []byte) bool {
	return re.match(b, len(b), 0)
}

// MatchString reports whether the string s contains any match of the regular
// expression re.
func (re *Regexp) MatchString(s string) bool {
	return re.Match([]byte(s))
}

// MatchReader reports whether the text returned by the RuneReader contains any
// match of the regular expression re.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) MatchReader(r io.RuneReader) bool {
	b, _ := readAll(r)
	return re.Match(b)
}

func readAll(r io.RuneReader) ([]byte, error) {
	var buf bytes.Buffer
	for {
		rune, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if _, err := buf.WriteRune(rune); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// allMatches calls deliver at most n times
// with the location of successive matches in the input text.
// The input text is b if non-nil, otherwise s.
func (re *Regexp) allMatches(b []byte, n int, deliver func([]int)) {
	end := len(b)

	for pos, i, prevMatchEnd := 0, 0, -1; i < n && pos <= end; {
		matches := re.find(b, end, pos)
		if len(matches) == 0 {
			break
		}

		accept := true
		if matches[1] == pos {
			// We've found an empty match.
			if matches[0] == prevMatchEnd {
				// We don't allow an empty match right
				// after a previous match, so ignore it.
				accept = false
			}

			// TODO: use step()
			_, width := utf8.DecodeRune(b[pos:end])
			if width > 0 {
				pos += width
			} else {
				pos = end + 1
			}
		} else {
			pos = matches[1]
		}
		prevMatchEnd = matches[1]

		if accept {
			deliver(re.pad(matches))
			i++
		}
	}
}

// The number of capture values in the program may correspond
// to fewer capturing expressions than are in the regexp.
// For example, "(a){0}" turns into an empty program, so the
// maximum capture in the program is 0 but we need to return
// an expression for \1.  Pad appends -1s to the slice a as needed.
func (re *Regexp) pad(a []int) []int {
	if a == nil {
		// No match.
		return nil
	}
	n := (1 + re.numSubexp) * 2
	for len(a) < n {
		a = append(a, -1)
	}
	return a
}
