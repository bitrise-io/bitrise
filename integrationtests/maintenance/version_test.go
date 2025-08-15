//go:build linux_and_mac
// +build linux_and_mac

package maintenance

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/bitrise/v2/models"

	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_VersionOutput(t *testing.T) {
	t.Log("Version --full")
	{
		out, err := command.RunCommandAndReturnCombinedStdoutAndStderr(testhelpers.BinPath(), "version", "--full")
		require.NoError(t, err)

		expectedOSVersion := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		expectedVersionOut := fmt.Sprintf(`version: %s
format version: %s
os: %s
go: %s
build number: 
commit: (devel)`, version.VERSION, models.FormatVersion, expectedOSVersion, runtime.Version())

		require.Equal(t, expectedVersionOut, out)
	}
}