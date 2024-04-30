package tools

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestEnvmanJSONPrint(t *testing.T) {
	// Initialized envstore -- Err should empty, output filled
	testDirPth, err := pathutil.NormalizedOSTempDirPath("test_env_store")
	require.NoError(t, err)

	envstorePth := filepath.Join(testDirPth, "envstore.yml")

	require.Equal(t, nil, EnvmanInit(envstorePth, true))

	out, err := EnvmanReadEnvList(envstorePth)
	require.NoError(t, err)
	require.Equal(t, models.EnvsJSONListModel{}, out)

	// Not initialized envstore -- Err should filled, output empty
	testDirPth, err = pathutil.NormalizedOSTempDirPath("test_env_store")
	require.NoError(t, err)

	envstorePth = filepath.Join(testDirPth, "envstore.yml")

	out, err = EnvmanReadEnvList(envstorePth)
	require.NotEqual(t, nil, err)
	require.Nil(t, out)
}
