package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePATHEnvString(t *testing.T) {
	t.Log("Empty starting PATH")
	require.Equal(t, "/MY/PATH",
		GeneratePATHEnvString("", "/MY/PATH"))

	t.Log("Empty PathToInclude")
	require.Equal(t, "/usr/bin:/usr/local/bin:/bin",
		GeneratePATHEnvString("/usr/bin:/usr/local/bin:/bin", ""))

	t.Log("Both Empty")
	require.Equal(t, "",
		GeneratePATHEnvString("", ""))

	t.Log("PATH = the path to include")
	require.Equal(t, "/MY/PATH",
		GeneratePATHEnvString("/MY/PATH", "/MY/PATH"))

	t.Log("PathToInclude is not in the PATH yet")
	require.Equal(t, "/MY/PATH:/usr/bin:/usr/local/bin:/bin",
		GeneratePATHEnvString("/usr/bin:/usr/local/bin:/bin", "/MY/PATH"))

	t.Log("PathToInclude is at the START of the PATH")
	require.Equal(t, "/MY/PATH:/usr/bin:/usr/local/bin:/bin",
		GeneratePATHEnvString("/MY/PATH:/usr/bin:/usr/local/bin:/bin", "/MY/PATH"))

	t.Log("PathToInclude is at the END of the PATH")
	require.Equal(t, "/usr/bin:/usr/local/bin:/bin:/MY/PATH",
		GeneratePATHEnvString("/usr/bin:/usr/local/bin:/bin:/MY/PATH", "/MY/PATH"))

	t.Log("PathToInclude is in the MIDDLE of the PATH")
	require.Equal(t, "/usr/bin:/MY/PATH:/usr/local/bin:/bin",
		GeneratePATHEnvString("/usr/bin:/MY/PATH:/usr/local/bin:/bin", "/MY/PATH"))
}

func TestInitPaths(t *testing.T) {
	//
	// BITRISE_SOURCE_DIR

	// Unset BITRISE_SOURCE_DIR -> after InitPaths BITRISE_SOURCE_DIR should be CurrentDir
	if os.Getenv(BitriseSourceDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseSourceDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.Equal(t, CurrentDir, os.Getenv(BitriseSourceDirEnvKey))

	// Set BITRISE_SOURCE_DIR -> after InitPaths BITRISE_SOURCE_DIR should keep content
	require.Equal(t, nil, os.Setenv(BitriseSourceDirEnvKey, "$HOME/test"))
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseSourceDirEnvKey))

	//
	// BITRISE_DEPLOY_DIR

	// Unset BITRISE_DEPLOY_DIR -> after InitPaths BITRISE_DEPLOY_DIR should be temp dir
	if os.Getenv(BitriseDeployDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseDeployDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseDeployDirEnvKey))

	// Set BITRISE_DEPLOY_DIR -> after InitPaths BITRISE_DEPLOY_DIR should keep content
	require.Equal(t, nil, os.Setenv(BitriseDeployDirEnvKey, "$HOME/test"))
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseDeployDirEnvKey))

	//
	// BITRISE_TEST_RESULTS_DIR

	// Unset BITRISE_TEST_RESULTS_DIR -> after InitPaths BITRISE_TEST_RESULTS_DIR should be temp dir
	if os.Getenv(BitriseTestResultsDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseTestResultsDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseTestResultsDirEnvKey))

	// Set BITRISE_TEST_RESULTS_DIR -> after InitPaths BITRISE_TEST_RESULTS_DIR should keep content
	require.Equal(t, nil, os.Setenv(BitriseTestResultsDirEnvKey, "$HOME/test"))
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseTestResultsDirEnvKey))

	//
	// BITRISE_CACHE_DIR

	// Unset BITRISE_CACHE_DIR -> after InitPaths BITRISE_CACHE_DIR should be temp dir
	if os.Getenv(BitriseCacheDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseCacheDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseCacheDirEnvKey))

	// Set BITRISE_CACHE_DIR -> after InitPaths BITRISE_CACHE_DIR should keep content
	require.Equal(t, nil, os.Setenv(BitriseCacheDirEnvKey, "$HOME/test"))
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseCacheDirEnvKey))

	//
	// BITRISE_TMP_DIR

	// Unset BITRISE_TMP_DIR -> after InitPaths BITRISE_TMP_DIR should be temp dir
	if os.Getenv(BitriseTmpDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseTmpDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseTmpDirEnvKey))

	// Set BITRISE_TMP_DIR -> after InitPaths BITRISE_TMP_DIR should keep content
	require.Equal(t, nil, os.Setenv(BitriseTmpDirEnvKey, "$HOME/test"))
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseTmpDirEnvKey))
}
