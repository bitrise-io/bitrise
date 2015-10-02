package bitrise

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestSetupForVersionChecks(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)

	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()

	require.Equal(t, nil, os.Setenv("HOME", fakeHomePth))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, nil, SaveSetupSuccessForVersion("0.9.7"))

	require.Equal(t, true, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.8"))
}
