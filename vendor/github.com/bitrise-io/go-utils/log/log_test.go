package log

import (
	"bytes"
	"testing"

	"regexp"

	"github.com/stretchr/testify/require"
)

func TestPrintf(t *testing.T) {
	t.Log("string")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printf("test")
		require.Equal(t, "test\n", b.String())
	}

	t.Log("format")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printf("%s", "test")
		require.Equal(t, "test\n", b.String())
	}

	t.Log("complex format")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printf("%s %s", "log", "test")
		require.Equal(t, "log test\n", b.String())
	}
}

func TestPrintft(t *testing.T) {
	t.Log("string")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printft("test")

		pattern := `\[.*\] test`
		re := regexp.MustCompile(pattern)

		require.Equal(t, true, re.MatchString(b.String()))
	}

	t.Log("format")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printft("%s", "test")

		pattern := `\[.*\] test`
		re := regexp.MustCompile(pattern)

		require.Equal(t, true, re.MatchString(b.String()))
	}

	t.Log("complex format")
	{
		var b bytes.Buffer
		SetOutWriter(&b)

		Printft("%s %s", "log", "test")

		pattern := `\[.*\] log test`
		re := regexp.MustCompile(pattern)

		require.Equal(t, true, re.MatchString(b.String()))
	}

}
