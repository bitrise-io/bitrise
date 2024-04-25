package onigmo

// ReplaceAll returns a copy of src, replacing matches of the Regexp with the
// replacement text repl. Inside repl, $ signs are interpreted as in Expand, so
// for instance $1 represents the text of the first submatch.
func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
	srepl := ""
	return re.replaceAll(src, func(dst []byte, match []int) []byte {
		if len(srepl) != len(repl) {
			srepl = string(repl)
		}

		return re.expand(dst, srepl, src, "", match)
	})
}

// ReplaceAllFunc returns a copy of src in which all matches of the Regexp have
// been replaced by the return value of function repl applied to the matched
// byte slice. The replacement returned by repl is substituted directly, without
// using Expand.
func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	return re.replaceAll(src, func(dst []byte, match []int) []byte {
		return append(dst, repl(src[match[0]:match[1]])...)
	})
}

// ReplaceAllString returns a copy of src, replacing matches of the Regexp with
// the replacement string repl. Inside repl, $ signs are interpreted as in
// Expand, so for instance $1 represents the text of the first submatch.
func (re *Regexp) ReplaceAllString(src, repl string) string {
	return string(re.ReplaceAll([]byte(src), []byte(repl)))
}

// ReplaceAllStringFunc returns a copy of src in which all matches of the Regexp
// have been replaced by the return value of function repl applied to the
// matched substring. The replacement returned by repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	b := re.replaceAll([]byte(src), func(dst []byte, match []int) []byte {
		return append(dst, repl(src[match[0]:match[1]])...)
	})

	return string(b)
}

// ReplaceAllLiteralString returns a copy of src, replacing matches of the Regexp
// with the replacement string repl. The replacement repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return string(re.replaceAll([]byte(src), func(dst []byte, match []int) []byte {
		return append(dst, repl...)
	}))
}

// ReplaceAllLiteral returns a copy of src, replacing matches of the Regexp
// with the replacement bytes repl. The replacement repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	return re.replaceAll(src, func(dst []byte, match []int) []byte {
		return append(dst, repl...)
	})
}

func (re *Regexp) replaceAll(src []byte, repl func(dst []byte, m []int) []byte) []byte {
	matches := re.findAll(src, len(src))
	if len(matches) == 0 {
		return src
	}

	lastMatchEnd := 0 // end position of the most recent match
	var buf []byte

	for _, a := range matches {
		// Copy the unmatched characters before this match.
		buf = append(buf, src[lastMatchEnd:a[0]]...)

		// Now insert a copy of the replacement string, but not for a
		// match of the empty string immediately after another match.
		// (Otherwise, we get double replacement for patterns that
		// match both empty and nonempty strings.)
		if a[1] > lastMatchEnd || a[0] == 0 {
			buf = repl(buf, a)
		}
		lastMatchEnd = a[1]
	}

	// Copy the unmatched characters after the last match.
	buf = append(buf, src[lastMatchEnd:]...)
	return buf
}
