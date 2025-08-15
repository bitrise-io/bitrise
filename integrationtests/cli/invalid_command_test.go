//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_InvalidCommand(t *testing.T) {
	t.Log("Invalid command")
	{
		_, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "invalidcmd")
		require.EqualError(t, err, "exit status 1")
	}
}