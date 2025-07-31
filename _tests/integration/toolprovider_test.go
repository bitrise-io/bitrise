//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestToolProvider(t *testing.T) {
	cmd := command.New(binPath(), "run", "toolprovider_test", "--config", "toolprovider_test_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
