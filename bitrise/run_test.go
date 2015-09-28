package bitrise

import (
	"path"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestStepmanStepLibStepInfo(t *testing.T) {
	// Valid params -- Err should empty, output filled
	require.Equal(t, nil, StepmanSetup("https://github.com/bitrise-io/bitrise-steplib"))

	outStr, err := StepmanStepLibStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "0.9.0")
	require.Equal(t, nil, err)
	require.NotEqual(t, "", outStr)

	// Invalid params -- Err should empty, output filled
	outStr, err = StepmanStepLibStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "2")
	require.NotEqual(t, nil, err)
	require.Equal(t, "", outStr)
}

func TestEnvmanJSONPrint(t *testing.T) {
	// Initialized envstore -- Err should empty, output filled
	testDirPth, err := pathutil.NormalizedOSTempDirPath("test_env_store")
	require.Equal(t, nil, err)

	envstorePth := path.Join(testDirPth, "envstore.yml")

	require.Equal(t, nil, EnvmanInitAtPath(envstorePth))

	outStr, err := EnvmanJSONPrint(envstorePth)
	require.Equal(t, nil, err)
	require.NotEqual(t, "", outStr)

	// Not initialized envstore -- Err should filled, output empty
	testDirPth, err = pathutil.NormalizedOSTempDirPath("test_env_store")
	require.Equal(t, nil, err)

	envstorePth = path.Join("test_env_store", "envstore.yml")

	outStr, err = EnvmanJSONPrint(envstorePth)
	require.NotEqual(t, nil, err)
	require.Equal(t, "", outStr)
}
