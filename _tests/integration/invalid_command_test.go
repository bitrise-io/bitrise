package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_InvalidCommand(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	t.Log("Invalid command")
	{
		_, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath(), "invalidcmd")
		require.EqualError(t, err, "exit status 1")
	}
}
