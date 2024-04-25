// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package onigmo

// Split slices s into substrings separated by the expression and returns a slice of
// the substrings between those expression matches.
//
// The slice returned by this method consists of all the substrings of s
// not contained in the slice returned by FindAllString. When called on an expression
// that contains no metacharacters, it is equivalent to strings.SplitN.
//
// Example:
//   s := regexp.MustCompile("a*").Split("abaabaccadaaae", 5)
//   // s: ["", "b", "b", "c", "cadaaae"]
//
// The count determines the number of substrings to return:
//   n > 0: at most n substrings; the last substring will be the unsplit remainder.
//   n == 0: the result is nil (zero substrings)
//   n < 0: all substrings
func (re *Regexp) Split(s string, n int) []string {
	// Code copy and pasted from the golang implementation.

	if n == 0 {
		return nil
	}

	if len(re.pattern) > 0 && len(s) == 0 {
		return []string{""}
	}

	fsin := -1
	if re.hasMetacharacters {
		fsin = n
	}

	matches := re.FindAllStringIndex(s, fsin)
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]

	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}
