package onigmo

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lonigmo
#include <onigmo.h>
*/
import "C"

// Option represents the compile time options.
type Option int

const (
	// OptionNone is the default value of option.
	OptionNone Option = 0
	// OptionIgnoreCase ambiguity match on.
	OptionIgnoreCase Option = 1
	// OptionExtend extended pattern form.
	OptionExtend Option = (OptionIgnoreCase << 1)
	// OptionMultiline '.' match with newline
	OptionMultiline Option = (OptionExtend << 1)
	// OptionSingleLine transforms '^' -into '\A', '$' -> '\Z'
	OptionSingleLine Option = (OptionMultiline << 1)
	// OptionFindLongest find longest match.
	OptionFindLongest Option = (OptionSingleLine << 1)
	// OptionFindNotEmpty ignore empty match.
	OptionFindNotEmpty Option = (OptionFindLongest << 1)
	// OptionNegateSingleLine disables OptionSingleLine witch is enable on
	// SyntaxPosixBasic, SyntaxPosixExtended, SyntaxPerl, SyntaxPerl58,
	// SyntaxPerl58NG, SyntaxPython and SyntaxJava
	OptionNegateSingleLine Option = (OptionFindNotEmpty << 1)
	// OptionDontCaptureGroup only named group captured.
	OptionDontCaptureGroup Option = (OptionNegateSingleLine << 1)
	// OptionCaptureGroup named and no-named group captured.
	OptionCaptureGroup Option = (OptionDontCaptureGroup << 1)
)

// Encoding defines the regular expression character encoding.
type Encoding = C.OnigEncoding

// Onigmo supported encoding types.
var (
	EncodingASCII       Encoding = &C.OnigEncodingASCII
	EncodingISO88591    Encoding = &C.OnigEncodingISO_8859_1
	EncodingISO88592    Encoding = &C.OnigEncodingISO_8859_2
	EncodingISO88593    Encoding = &C.OnigEncodingISO_8859_3
	EncodingISO88594    Encoding = &C.OnigEncodingISO_8859_4
	EncodingISO88595    Encoding = &C.OnigEncodingISO_8859_5
	EncodingISO88596    Encoding = &C.OnigEncodingISO_8859_6
	EncodingISO88597    Encoding = &C.OnigEncodingISO_8859_7
	EncodingISO88598    Encoding = &C.OnigEncodingISO_8859_8
	EncodingISO88599    Encoding = &C.OnigEncodingISO_8859_9
	EncodingISO885910   Encoding = &C.OnigEncodingISO_8859_10
	EncodingISO885911   Encoding = &C.OnigEncodingISO_8859_11
	EncodingISO885913   Encoding = &C.OnigEncodingISO_8859_13
	EncodingISO885914   Encoding = &C.OnigEncodingISO_8859_14
	EncodingISO885915   Encoding = &C.OnigEncodingISO_8859_15
	EncodingISO885916   Encoding = &C.OnigEncodingISO_8859_16
	EncodingUTF8        Encoding = &C.OnigEncodingUTF_8
	EncodingUTF16BE     Encoding = &C.OnigEncodingUTF_16BE
	EncodingUTF16LE     Encoding = &C.OnigEncodingUTF_16LE
	EncodingUTF32BE     Encoding = &C.OnigEncodingUTF_32BE
	EncodingUTF32LE     Encoding = &C.OnigEncodingUTF_32LE
	EncodingEUCJP       Encoding = &C.OnigEncodingEUC_JP
	EncodingEUCTW       Encoding = &C.OnigEncodingEUC_TW
	EncodingEUCKR       Encoding = &C.OnigEncodingEUC_KR
	EncodingEUCCN       Encoding = &C.OnigEncodingEUC_CN
	EncodingShiftJIS    Encoding = &C.OnigEncodingShift_JIS
	EncodingWindows31J  Encoding = &C.OnigEncodingWindows_31J
	EncodingKOI8R       Encoding = &C.OnigEncodingKOI8_R
	EncodingKOI8U       Encoding = &C.OnigEncodingKOI8_U
	EncodingWindows1250 Encoding = &C.OnigEncodingWindows_1250
	EncodingWindows1251 Encoding = &C.OnigEncodingWindows_1251
	EncodingWindows1252 Encoding = &C.OnigEncodingWindows_1252
	EncodingWindows1253 Encoding = &C.OnigEncodingWindows_1253
	EncodingWindows1254 Encoding = &C.OnigEncodingWindows_1254
	EncodingWindows1257 Encoding = &C.OnigEncodingWindows_1257
	EncodingBIG5        Encoding = &C.OnigEncodingBIG5
	EncodingGB18030     Encoding = &C.OnigEncodingGB18030
)

type Syntax = *C.OnigSyntaxType

// Onigmo supported syntaxes
var (
	// plain text
	SyntaxASIS Syntax = &C.OnigSyntaxASIS
	// POSIX Basic RE
	SyntaxPosixBasic Syntax = &C.OnigSyntaxPosixBasic
	// POSIX Extended RE
	SyntaxPosixExtended Syntax = &C.OnigSyntaxPosixExtended
	// Emacs
	SyntaxEmacs Syntax = &C.OnigSyntaxEmacs
	// grep
	SyntaxGrep Syntax = &C.OnigSyntaxGrep
	// GNU regex
	SyntaxGnuRegex Syntax = &C.OnigSyntaxGnuRegex
	// Java (Sun java.util.regex)
	SyntaxJava Syntax = &C.OnigSyntaxJava
	// Perl 5.8
	SyntaxPerl58 Syntax = &C.OnigSyntaxPerl58
	// Perl 5.8 + named group
	SyntaxPerl58NG Syntax = &C.OnigSyntaxPerl58_NG
	// Perl 5.10+
	SyntaxPerl Syntax = &C.OnigSyntaxPerl
	// Python
	SyntaxRuby Syntax = &C.OnigSyntaxRuby
	// Ruby
	SyntaxPython Syntax = &C.OnigSyntaxPython
)
