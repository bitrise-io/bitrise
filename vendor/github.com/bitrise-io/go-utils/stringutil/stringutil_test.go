package stringutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFirstLine(t *testing.T) {
	t.Log("Empty input")
	{
		require.Equal(t, "", ReadFirstLine("", false))
		require.Equal(t, "", ReadFirstLine("", true))
	}

	t.Log("Multiline empty input - ignore-empty-lines:false")
	{
		firstLine := ReadFirstLine(`


`, false)
		require.Equal(t, "", firstLine)
	}

	t.Log("Multiline empty input - ignore-empty-lines:true")
	{
		firstLine := ReadFirstLine(`


`, true)
		require.Equal(t, "", firstLine)
	}

	t.Log("Multiline non empty input - ignore-empty-lines:false")
	{
		firstLine := ReadFirstLine(`first line

second line`, false)
		require.Equal(t, "first line", firstLine)
	}

	t.Log("Multiline empty input - ignore-empty-lines:true")
	{
		firstLine := ReadFirstLine(`first line

second line`, true)
		require.Equal(t, "first line", firstLine)
	}

	t.Log("Multiline non empty input, with leading empty line - ignore-empty-lines:false")
	{
		firstLine := ReadFirstLine(`

first line

second line`, false)
		require.Equal(t, "", firstLine)
	}

	t.Log("Multiline non empty input, with leading empty line - ignore-empty-lines:true")
	{
		firstLine := ReadFirstLine(`

first line

second line`, true)
		require.Equal(t, "first line", firstLine)
	}
}

func TestCaseInsensitiveEquals(t *testing.T) {
	var emptyStr string // ""

	require.Equal(t, true, CaseInsensitiveEquals(emptyStr, emptyStr))
	require.Equal(t, true, CaseInsensitiveEquals("", ""))

	require.Equal(t, true, CaseInsensitiveEquals(emptyStr, ""))
	require.Equal(t, true, CaseInsensitiveEquals("", emptyStr))

	require.Equal(t, false, CaseInsensitiveEquals(emptyStr, "a"))
	require.Equal(t, false, CaseInsensitiveEquals("a", emptyStr))

	require.Equal(t, true, CaseInsensitiveEquals("a", "a"))
	require.Equal(t, true, CaseInsensitiveEquals("a", "A"))

	require.Equal(t, true, CaseInsensitiveEquals("ab", "Ab"))

	require.Equal(t, false, CaseInsensitiveEquals("ab", "ba"))
	require.Equal(t, false, CaseInsensitiveEquals("ab", "Ba"))
}

func TestCaseInsensitiveContains(t *testing.T) {
	var emptyStr string // ""

	require.Equal(t, true, CaseInsensitiveContains(emptyStr, emptyStr))
	require.Equal(t, true, CaseInsensitiveContains("", ""))

	require.Equal(t, true, CaseInsensitiveContains(emptyStr, ""))
	require.Equal(t, true, CaseInsensitiveContains("", emptyStr))

	require.Equal(t, false, CaseInsensitiveContains(emptyStr, "a"))
	require.Equal(t, true, CaseInsensitiveContains("a", emptyStr))

	require.Equal(t, true, CaseInsensitiveContains("a", "a"))
	require.Equal(t, true, CaseInsensitiveContains("A", "a"))

	require.Equal(t, true, CaseInsensitiveContains("abc", "a"))
	require.Equal(t, true, CaseInsensitiveContains("abc", "B"))
	require.Equal(t, true, CaseInsensitiveContains("abc", "BC"))
	require.Equal(t, true, CaseInsensitiveContains("abc", "ABC"))

	require.Equal(t, false, CaseInsensitiveContains("abc", "d"))
	require.Equal(t, false, CaseInsensitiveContains("abc", "ba"))
	require.Equal(t, false, CaseInsensitiveContains("abc", "BAC"))
}

func TestGenericTrim(t *testing.T) {
	require.Equal(t, "", genericTrim("", 4, false, false))
	require.Equal(t, "", genericTrim("", 4, false, true))
	require.Equal(t, "", genericTrim("", 4, true, false))
	require.Equal(t, "", genericTrim("", 4, true, true))

	require.Equal(t, "1234", genericTrim("123456789", 4, false, false))
	require.Equal(t, "1...", genericTrim("123456789", 4, false, true))
	require.Equal(t, "6789", genericTrim("123456789", 4, true, false))
	require.Equal(t, "...9", genericTrim("123456789", 4, true, true))
}

func TestMaxLastChars(t *testing.T) {
	require.Equal(t, "", MaxLastChars("", 10))
	require.Equal(t, "a", MaxLastChars("a", 1))
	require.Equal(t, "a", MaxLastChars("ba", 1))
	require.Equal(t, "ba", MaxLastChars("ba", 10))
	require.Equal(t, "a", MaxLastChars("cba", 1))
	require.Equal(t, "cba", MaxLastChars("cba", 10))

	require.Equal(t, "llo world!", MaxLastChars("hello world!", 10))
}

func TestMaxLastCharsWithDots(t *testing.T) {
	require.Equal(t, "", MaxLastCharsWithDots("", 10))
	require.Equal(t, "", MaxLastCharsWithDots("1234", 1))
	require.Equal(t, "...56", MaxLastCharsWithDots("123456", 5))
	require.Equal(t, "123456", MaxFirstCharsWithDots("123456", 6))
	require.Equal(t, "123456", MaxLastCharsWithDots("123456", 10))

	require.Equal(t, "... world!", MaxLastCharsWithDots("hello world!", 10))
}

func TestMaxFirstChars(t *testing.T) {
	require.Equal(t, "", MaxFirstChars("", 10))
	require.Equal(t, "a", MaxFirstChars("a", 1))
	require.Equal(t, "b", MaxFirstChars("ba", 1))
	require.Equal(t, "ba", MaxFirstChars("ba", 10))
	require.Equal(t, "c", MaxFirstChars("cba", 1))
	require.Equal(t, "cba", MaxFirstChars("cba", 10))

	require.Equal(t, "hello worl", MaxFirstChars("hello world!", 10))
}

func TestMaxFirstCharsWithDots(t *testing.T) {
	require.Equal(t, "", MaxFirstCharsWithDots("", 10))
	require.Equal(t, "", MaxFirstCharsWithDots("1234", 1))
	require.Equal(t, "12...", MaxFirstCharsWithDots("123456", 5))
	require.Equal(t, "123456", MaxFirstCharsWithDots("123456", 6))
	require.Equal(t, "123456", MaxFirstCharsWithDots("123456", 10))

	require.Equal(t, "hello w...", MaxFirstCharsWithDots("hello world!", 10))
}
