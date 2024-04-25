package onigmo

import (
	"io"
)

// FindIndex returns a two-element slice of integers defining the location of
// the leftmost match in b of the regular expression. The match itself is at
// b[loc[0]:loc[1]]. A return value of nil indicates no match.
func (re *Regexp) FindIndex(b []byte) []int {
	match := re.find(b, len(b), 0)
	if len(match) == 0 {
		return nil
	}

	return match[:2]
}

// Find returns a slice holding the text of the leftmost match in b of the
// regular expression. A return value of nil indicates no match.
func (re *Regexp) Find(b []byte) []byte {
	loc := re.FindIndex(b)
	if loc == nil {
		return nil
	}

	return b[loc[0]:loc[1]:loc[1]]
}

// FindString returns a string holding the text of the leftmost match in s of
// the regular expression. If there is no match, the return value is an empty
// string, but it will also be empty if the regular expression successfully
// matches an empty string. Use FindStringIndex or FindStringSubmatch if it is
// necessary to distinguish these cases.
func (re *Regexp) FindString(s string) string {
	mb := re.Find([]byte(s))
	if mb == nil {
		return ""
	}

	return string(mb)
}

// FindStringIndex returns a two-element slice of integers defining the location
// of the leftmost match in s of the regular expression. The match itself is at
// s[loc[0]:loc[1]]. A return value of nil indicates no match.
func (re *Regexp) FindStringIndex(s string) []int {
	return re.FindIndex([]byte(s))
}

// FindAllIndex is the 'All' version of FindIndex; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllIndex(b []byte, n int) [][]int {
	if n < 0 {
		n = len(b) + 1
	}
	var result [][]int
	re.allMatches(b, n, func(match []int) {
		if result == nil {
			result = make([][]int, 0, startSize)
		}
		result = append(result, match[0:2])
	})
	return result
}

const startSize = 10 // The size at which to start a slice in the 'All' routines.

// FindAll is the 'All' version of Find; it returns a slice of all successive
// matches of the expression, as defined by the 'All' description in the package
// comment. A return value of nil indicates no match.
func (re *Regexp) FindAll(b []byte, n int) [][]byte {
	if n < 0 {
		n = len(b) + 1
	}
	var result [][]byte
	re.allMatches(b, n, func(match []int) {
		if result == nil {
			result = make([][]byte, 0, startSize)
		}
		result = append(result, b[match[0]:match[1]:match[1]])
	})
	return result
}

// FindAllString is the 'All' version of FindString; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllString(s string, n int) []string {
	if n < 0 {
		n = len(s) + 1
	}
	var result []string

	b := []byte(s)
	re.allMatches(b, n, func(match []int) {
		if result == nil {
			result = make([]string, 0, startSize)
		}
		f := string(b[match[0]:match[1]])
		result = append(result, f)
	})
	return result
}

// FindAllStringIndex is the 'All' version of FindStringIndex; it returns a
// slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllStringIndex(s string, n int) [][]int {
	if n < 0 {
		n = len(s) + 1
	}
	var result [][]int
	re.allMatches([]byte(s), n, func(match []int) {
		if result == nil {
			result = make([][]int, 0, startSize)
		}
		result = append(result, match[0:2])
	})
	return result
}

// FindSubmatchIndex returns a slice holding the index pairs identifying the
// leftmost match of the regular expression in b and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' and 'Index' descriptions in the
// package comment. A return value of nil indicates no match.
func (re *Regexp) FindSubmatchIndex(b []byte) []int {
	match := re.find(b, len(b), 0)
	if len(match) == 0 {
		return nil
	}

	return match
}

// FindSubmatch returns a slice of slices holding the text of the leftmost match
// of the regular expression in b and the matches, if any, of its subexpressions,
// as defined by the 'Submatch' descriptions in the package comment. A return
// value of nil indicates no match.
func (re *Regexp) FindSubmatch(b []byte) [][]byte {
	a := re.FindSubmatchIndex(b)
	if a == nil {
		return nil
	}

	ret := make([][]byte, 1+re.numSubexp)
	for i := range ret {
		if 2*i < len(a) && a[2*i] >= 0 {
			ret[i] = b[a[2*i]:a[2*i+1]:a[2*i+1]]
		}
	}

	return ret
}

// FindStringSubmatch returns a slice of strings holding the text of the
// leftmost match of the regular expression in s and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' description in the package
// comment. A return value of nil indicates no match.
func (re *Regexp) FindStringSubmatch(s string) []string {
	b := []byte(s)
	match := re.FindSubmatch(b)
	if match == nil {
		return nil
	}

	results := make([]string, 0, len(match))
	for _, match := range match {
		results = append(results, string(match))
	}

	return results
}

// FindStringSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regular expression in s and the matches, if any, of
// its subexpressions, as defined by the 'Submatch' and 'Index' descriptions in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindStringSubmatchIndex(s string) []int {
	return re.FindSubmatchIndex([]byte(s))
}

// FindAllSubmatchIndex is the 'All' version of FindSubmatchIndex; it returns a
// slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	if n < 0 {
		n = len(b) + 1
	}
	var result [][]int
	re.allMatches(b, n, func(match []int) {
		if result == nil {
			result = make([][]int, 0, startSize)
		}
		result = append(result, match)
	})
	return result
}

// FindAllSubmatch is the 'All' version of FindSubmatch; it returns a slice of
// all successive matches of the expression, as defined by the 'All' description
// in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	if n < 0 {
		n = len(b) + 1
	}
	var result [][][]byte
	re.allMatches(b, n, func(match []int) {
		if result == nil {
			result = make([][][]byte, 0, startSize)
		}
		slice := make([][]byte, len(match)/2)
		for j := range slice {
			if match[2*j] >= 0 {
				slice[j] = b[match[2*j]:match[2*j+1]:match[2*j+1]]
			}
		}
		result = append(result, slice)
	})
	return result
}

// FindAllStringSubmatch is the 'All' version of FindStringSubmatch; it returns
// a slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	if n < 0 {
		n = len(s) + 1
	}
	var result [][]string
	re.allMatches([]byte(s), n, func(match []int) {
		if result == nil {
			result = make([][]string, 0, startSize)
		}
		slice := make([]string, len(match)/2)
		for j := range slice {
			if match[2*j] >= 0 {
				slice[j] = s[match[2*j]:match[2*j+1]]
			}
		}
		result = append(result, slice)
	})
	return result
}

// FindAllStringSubmatchIndex is the 'All' version of FindStringSubmatchIndex;
// it returns a slice of all successive matches of the expression, as defined
// by the 'All' description in the package comment. A return value of nil
// indicates no match.
func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	if n < 0 {
		n = len(s) + 1
	}
	var result [][]int
	re.allMatches([]byte(s), n, func(match []int) {
		if result == nil {
			result = make([][]int, 0, startSize)
		}
		result = append(result, match)
	})
	return result
}

// FindReaderIndex returns a two-element slice of integers defining the location
// of the leftmost match of the regular expression in text read from the
// RuneReader. The match text was found in the input stream at byte offset
// loc[0] through loc[1]-1. A return value of nil indicates no match.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) FindReaderIndex(r io.RuneReader) []int {
	b, _ := readAll(r)
	return re.FindIndex(b)
}

// FindReaderSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regular expression of text read by the RuneReader,
// and the matches, if any, of its subexpressions, as defined by the 'Submatch'
// and 'Index' descriptions in the package comment. A return value of nil
// indicates no match.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) FindReaderSubmatchIndex(r io.RuneReader) []int {
	b, _ := readAll(r)
	return re.FindSubmatchIndex(b)
}
