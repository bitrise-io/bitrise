package configs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestSetupForVersionChecks(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	originalHome := os.Getenv("HOME")

	defer func() {
		require.Equal(t, nil, os.Setenv("HOME", originalHome))
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()

	require.Equal(t, nil, os.Setenv("HOME", fakeHomePth))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, nil, SaveSetupSuccessForVersion("0.9.7"))

	require.Equal(t, true, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.8"))
}
