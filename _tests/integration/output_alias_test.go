package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_OutputAlias(t *testing.T) {
	configPth := "output_alias_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "test", "--config", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
