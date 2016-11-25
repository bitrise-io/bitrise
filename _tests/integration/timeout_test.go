package integration

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_TimeoutTest(t *testing.T) {
	configPth := "timeout_test_bitrise.yml"

	t.Log("")
	{
		start := time.Now()
		cmd := cmdex.NewCommand(binPath(), "run", "timeout", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		elapsed := time.Since(start)
		require.EqualError(t, err, "exit status 1", out)
		require.Equal(t, true, elapsed < 10*time.Second)
	}
}
