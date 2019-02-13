package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetOutWriter(t *testing.T) {
	var b bytes.Buffer
	SetOutWriter(&b)
	Printf("test %s", "log")
	require.Equal(t, "test log\n", b.String())
}

func TestSetEnableDebugLog(t *testing.T) {
	t.Log("enable debug log")
	{
		SetEnableDebugLog(true)
		var b bytes.Buffer
		SetOutWriter(&b)
		Debugf("test %s", "log")
		require.Equal(t, "test log\n", b.String())
	}

	t.Log("disable debug log")
	{
		SetEnableDebugLog(false)
		var b bytes.Buffer
		SetOutWriter(&b)
		Debugf("test %s", "log")
		require.Equal(t, "", b.String())
	}
}

func TestSetTimestampLayout(t *testing.T) {
	var b bytes.Buffer
	SetOutWriter(&b)
	SetTimestampLayout("15-04-05")
	TPrintf("test %s", "log")
	re := regexp.MustCompile(`\[.+-.+-.+\] test log`)
	require.True(t, re.MatchString(b.String()), b.String())
}
