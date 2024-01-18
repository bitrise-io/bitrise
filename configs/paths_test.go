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
	t.Setenv(BitriseSourceDirEnvKey, "$HOME/test")
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
	t.Setenv(BitriseDeployDirEnvKey, "$HOME/test")
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseDeployDirEnvKey))

	//
	// BITRISE_TEST_DEPLOY_DIR

	// Unset BITRISE_TEST_DEPLOY_DIR -> after InitPaths BITRISE_TEST_DEPLOY_DIR should be temp dir
	if os.Getenv(BitriseTestDeployDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseTestDeployDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseTestDeployDirEnvKey))

	// Set BITRISE_TEST_DEPLOY_DIR -> after InitPaths BITRISE_TEST_DEPLOY_DIR should keep content
	t.Setenv(BitriseTestDeployDirEnvKey, "$HOME/test")
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseTestDeployDirEnvKey))

	//
	// BITRISE_TMP_DIR

	// Unset BITRISE_TMP_DIR -> after InitPaths BITRISE_TMP_DIR should be temp dir
	if os.Getenv(BitriseTmpDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseTmpDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseTmpDirEnvKey))

	// Set BITRISE_TMP_DIR -> after InitPaths BITRISE_TMP_DIR should keep content
	t.Setenv(BitriseTmpDirEnvKey, "$HOME/test")
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseTmpDirEnvKey))

	//
	// BITRISE_HTML_REPORT_DIR

	// Unset BITRISE_HTML_REPORT_DIR -> after InitPaths BITRISE_HTML_REPORT_DIR should be temp dir
	if os.Getenv(BitriseHtmlReportDirEnvKey) != "" {
		require.Equal(t, nil, os.Unsetenv(BitriseHtmlReportDirEnvKey))
	}
	require.Equal(t, nil, InitPaths())
	require.NotEqual(t, "", os.Getenv(BitriseHtmlReportDirEnvKey))

	// Set BITRISE_HTML_REPORT_DIR -> after InitPaths BITRISE_TEST_DEPLOY_DIR should keep content
	t.Setenv(BitriseHtmlReportDirEnvKey, "$HOME/test")
	require.Equal(t, nil, InitPaths())
	require.Equal(t, "$HOME/test", os.Getenv(BitriseHtmlReportDirEnvKey))
}
