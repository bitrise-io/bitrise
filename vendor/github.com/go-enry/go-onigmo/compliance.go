package onigmo

import "io"

// golang regexp.Regexp signatures package
type compliance interface {
	Copy() *Regexp
	Expand(dst []byte, template []byte, src []byte, match []int) []byte
	ExpandString(dst []byte, template string, src string, match []int) []byte
	Find(b []byte) []byte
	FindAll(b []byte, n int) [][]byte
	FindAllIndex(b []byte, n int) [][]int
	FindAllString(s string, n int) []string
	FindAllStringIndex(s string, n int) [][]int
	FindAllStringSubmatch(s string, n int) [][]string
	FindAllStringSubmatchIndex(s string, n int) [][]int
	FindAllSubmatch(b []byte, n int) [][][]byte
	FindAllSubmatchIndex(b []byte, n int) [][]int
	FindIndex(b []byte) (loc []int)
	FindReaderIndex(r io.RuneReader) (loc []int)
	FindReaderSubmatchIndex(r io.RuneReader) []int
	FindString(s string) string
	FindStringIndex(s string) (loc []int)
	FindStringSubmatch(s string) []string
	FindStringSubmatchIndex(s string) []int
	FindSubmatch(b []byte) [][]byte
	FindSubmatchIndex(b []byte) []int
	LiteralPrefix() (prefix string, complete bool)
	Longest()
	Match(b []byte) bool
	MatchReader(r io.RuneReader) bool
	MatchString(s string) bool
	NumSubexp() int
	ReplaceAll(src, repl []byte) []byte
	ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte
	ReplaceAllLiteral(src, repl []byte) []byte
	ReplaceAllLiteralString(src, repl string) string
	ReplaceAllString(src, repl string) string
	ReplaceAllStringFunc(src string, repl func(string) string) string
	Split(s string, n int) []string
	String() string
	SubexpNames() []string
}

// func CompilePOSIX ¶
// func MustCompilePOSIX(str string) *Regexp
