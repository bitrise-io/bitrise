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
	require.NoError(t, InitPaths())
	require.Equal(t, CurrentDir, os.Getenv(BitriseSourceDirEnvKey))

	// BITRISE_DEPLOY_DIR
	testTempDirectoryWithEnvVar(t, BitriseDeployDirEnvKey)

	// BITRISE_TEST_DEPLOY_DIR
	testTempDirectoryWithEnvVar(t, BitriseTestDeployDirEnvKey)

	// BITRISE_TMP_DIR
	testTempDirectoryWithEnvVar(t, BitriseTmpDirEnvKey)

	// BITRISE_HTML_REPORT_DIR
	testTempDirectoryWithEnvVar(t, BitriseHtmlReportDirEnvKey)
}

func testTempDirectoryWithEnvVar(t *testing.T, dirEnvKey string) {
	// Unset env var -> after InitPaths env var should be a temp dir
	if os.Getenv(dirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(dirEnvKey))
	}
	require.NoError(t, InitPaths())
	require.NotEqual(t, "", os.Getenv(dirEnvKey))

	// Set dirEnvKey -> after InitPaths dirEnvKey should keep content
	t.Setenv(dirEnvKey, "$HOME/test")
	require.NoError(t, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(dirEnvKey))
}
