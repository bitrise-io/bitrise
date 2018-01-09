package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_printf_with_time(t *testing.T) {
	SetTimestampLayout("15.04.05")
	var b bytes.Buffer
	SetOutWriter(&b)
	printf(normalSeverity, true, "test %s", "log")
	re := regexp.MustCompile(`\[.+\..+\..+\] test log`)
	require.True(t, re.MatchString(b.String()), b.String())
}

func Test_printf_severity(t *testing.T) {
	t.Log("error")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(errorSeverity, false, "test %s", "log")
		require.Equal(t, "\x1b[31;1mtest log\x1b[0m\n", b.String())
	}

	t.Log("warn")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(warnSeverity, false, "test %s", "log")
		require.Equal(t, "\x1b[33;1mtest log\x1b[0m\n", b.String())
	}

	t.Log("debug")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(debugSeverity, false, "test %s", "log")
		require.Equal(t, "test log\n", b.String())
	}

	t.Log("normal")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(normalSeverity, false, "test %s", "log")
		require.Equal(t, "test log\n", b.String())
	}

	t.Log("info")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(infoSeverity, false, "test %s", "log")
		require.Equal(t, "\x1b[34;1mtest log\x1b[0m\n", b.String())
	}

	t.Log("success")
	{
		var b bytes.Buffer
		SetOutWriter(&b)
		printf(successSeverity, false, "test %s", "log")
		require.Equal(t, "\x1b[32;1mtest log\x1b[0m\n", b.String())
	}
}
