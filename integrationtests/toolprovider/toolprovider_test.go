//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestToolProvider(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func TestMiseToolProvider(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "toolprovider_test", "--config", "toolprovider_test_mise_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
