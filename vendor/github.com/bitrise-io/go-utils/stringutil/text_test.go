package stringutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndentTextWithMaxLength(t *testing.T) {
	t.Log("Empty")
	{
		input := ""
		output := IndentTextWithMaxLength(input, "", 80, true)
		require.Equal(t, "", output)
	}

	t.Log("One liner")
	{
		input := "one liner"
		output := IndentTextWithMaxLength(input, "", 80, true)
		require.Equal(t, "one liner", output)
	}

	t.Log("One liner - with indent")
	{
		input := "one liner"
		output := IndentTextWithMaxLength(input, " => ", 76, true)
		require.Equal(t, " => one liner", output)
	}

	t.Log("One liner - max width")
	{
		input := "one"
		output := IndentTextWithMaxLength(input, "", 3, true)
		require.Equal(t, "one", output)
	}

	t.Log("One liner - longer than max width")
	{
		input := "onetwo"
		output := IndentTextWithMaxLength(input, "", 3, true)
		require.Equal(t, "one\ntwo", output)
	}

	t.Log("One liner - max width - with indent")
	{
		input := "one"
		require.Equal(t, " on\n e", IndentTextWithMaxLength(input, " ", 2, true))
		require.Equal(t, "on\n e", IndentTextWithMaxLength(input, " ", 2, false))
	}

	t.Log("One liner - max width - with first-line indent false")
	{
		input := "one"
		output := IndentTextWithMaxLength(input, " ", 2, false)
		require.Equal(t, "on\n e", output)
	}

	t.Log("One liner - longer than max width - with indent")
	{
		input := "onetwo"
		require.Equal(t, " on\n et\n wo", IndentTextWithMaxLength(input, " ", 2, true))
		require.Equal(t, "on\n et\n wo", IndentTextWithMaxLength(input, " ", 2, false))
	}

	t.Log("Two lines, shorter than max")
	{
		input := `first line
second line`
		output := IndentTextWithMaxLength(input, "", 80, true)
		require.Equal(t, `first line
second line`, output)
	}

	t.Log("Two lines, shorter than max - with indent")
	{
		input := `first line
second line`

		require.Equal(t, `  first line
  second line`, IndentTextWithMaxLength(input, "  ", 78, true))
		require.Equal(t, `first line
  second line`, IndentTextWithMaxLength(input, "  ", 78, false))
	}

	t.Log("Two lines, longer than max")
	{
		input := `firstline
secondline`
		output := IndentTextWithMaxLength(input, "", 5, true)
		require.Equal(t, `first
line
secon
dline`, output)
	}

	t.Log("Max length = 0")
	{
		input := "Indent is longer than max length"
		require.Equal(t, "", IndentTextWithMaxLength(input, "...", 0, true))
		require.Equal(t, "", IndentTextWithMaxLength(input, "...", 0, false))
		require.Equal(t, "", IndentTextWithMaxLength(input, "", 0, true))
		require.Equal(t, "", IndentTextWithMaxLength(input, "", 0, false))
	}

	t.Log("Max length = 0 - multi-line")
	{
		input := `Indent is longer
than max
length`
		require.Equal(t, "", IndentTextWithMaxLength(input, "...", 0, true))
		require.Equal(t, "", IndentTextWithMaxLength(input, "...", 0, false))
		require.Equal(t, "", IndentTextWithMaxLength(input, "", 0, true))
		require.Equal(t, "", IndentTextWithMaxLength(input, "", 0, false))
	}
}
