package configs

import (
	"os"
	"path/filepath"
	"testing"

	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestSetupForVersionChecks(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)

	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()

	t.Setenv("HOME", fakeHomePth)
	t.Setenv("XDG_CONFIG_HOME", "")

	versionMatch, _ := CheckIsSetupWasDoneForVersion("0.9.7")
	require.Equal(t, false, versionMatch)

	require.Equal(t, nil, SaveSetupSuccessForVersion("0.9.7"))

	versionMatch, _ = CheckIsSetupWasDoneForVersion("0.9.7")
	require.Equal(t, true, versionMatch)

	versionMatch, _ = CheckIsSetupWasDoneForVersion("0.9.8")
	require.Equal(t, false, versionMatch)
}

func TestLoadLegacyConfig(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)

	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()

	t.Setenv("HOME", fakeHomePth)
	t.Setenv("XDG_CONFIG_HOME", "")

	// Missing file returns the zero ConfigModel, false, not an error.
	got, exists, err := LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, ConfigModel{}, got)
	require.Equal(t, false, exists)

	// A brand-new user (no pre-existing legacy file) only gets the new
	// config.yml written -- LoadLegacyConfig still sees zero.
	require.Equal(t, nil, SaveSetupSuccessForVersion("1.2.3"))
	got, exists, err = LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, "", got.SetupVersion)
	require.Equal(t, false, exists)

	// Once a legacy file exists, LoadLegacyConfig is a pure passthrough
	// again: it sees whatever SaveSetupSuccessForVersion (and friends) write.
	require.Equal(t, nil, saveLegacyConfig(ConfigModel{SetupVersion: "0.0.1"}))
	require.Equal(t, nil, SaveSetupSuccessForVersion("1.2.3"))
	got, exists, err = LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, "1.2.3", got.SetupVersion)
	require.Equal(t, true, exists)
}

// TestSaveSetupSuccessForVersion_NewUser_OnlySavesToGlobalConfig asserts a
// brand-new user (no pre-existing ~/.bitrise/config.json) never gets that
// legacy file created; the value lands only in the new global config.yml.
func TestSaveSetupSuccessForVersion_NewUser_OnlySavesToGlobalConfig(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()
	t.Setenv("HOME", fakeHomePth)
	t.Setenv("XDG_CONFIG_HOME", "")

	require.Equal(t, nil, SaveSetupSuccessForVersion("3.0.0"))

	_, exists, err := LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, false, exists)

	globalCfg, err := internalconfig.Load()
	require.Equal(t, nil, err)
	require.Equal(t, "3.0.0", globalCfg.SetupVersion)

	// CheckIsSetupWasDoneForVersion falls back to the global config so a
	// brand-new user's saved state is actually observed on the next check.
	versionMatch, _ := CheckIsSetupWasDoneForVersion("3.0.0")
	require.Equal(t, true, versionMatch)
}

// TestSaveSetupSuccessForVersion_ExistingUser_SyncsBothFiles asserts an
// existing user's legacy file keeps being updated (it's still the
// highest-precedence read layer), and the new global config.yml is kept in
// sync alongside it.
func TestSaveSetupSuccessForVersion_ExistingUser_SyncsBothFiles(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()
	t.Setenv("HOME", fakeHomePth)
	t.Setenv("XDG_CONFIG_HOME", "")

	require.Equal(t, nil, EnsureBitriseConfigDirExists())
	require.Equal(t, nil, saveLegacyConfig(ConfigModel{SetupVersion: "2.9.0"}))

	require.Equal(t, nil, SaveSetupSuccessForVersion("3.0.0"))

	legacy, _, err := LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, "3.0.0", legacy.SetupVersion)

	globalCfg, err := internalconfig.Load()
	require.Equal(t, nil, err)
	require.Equal(t, "3.0.0", globalCfg.SetupVersion)
}

// TestSaveSetupSuccessForVersion_ExistingUser_GlobalSyncFailureDoesNotFailSave
// asserts the global-config mirror is best-effort *when legacy exists*: a
// broken new-location write (here, a plain file standing in for the
// XDG_CONFIG_HOME directory) must not fail the overall Save, since the legacy
// write already succeeded and is the load-bearing one.
func TestSaveSetupSuccessForVersion_ExistingUser_GlobalSyncFailureDoesNotFailSave(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()
	t.Setenv("HOME", fakeHomePth)

	require.Equal(t, nil, EnsureBitriseConfigDirExists())
	require.Equal(t, nil, saveLegacyConfig(ConfigModel{SetupVersion: "2.9.0"}))

	blockingFile := fakeHomePth + "-blocking-file"
	require.Equal(t, nil, os.WriteFile(blockingFile, []byte("x"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", blockingFile)

	require.Equal(t, nil, SaveSetupSuccessForVersion("3.0.0"))

	legacy, _, err := LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, "3.0.0", legacy.SetupVersion)
}

// TestSaveSetupSuccessForVersion_NewUser_GlobalSyncFailureFailsSave is the
// other half: with no legacy file, config.yml is the *only* place the value
// gets persisted, so a broken new-location write must surface as an error
// instead of being silently swallowed and reported as a successful save.
func TestSaveSetupSuccessForVersion_NewUser_GlobalSyncFailureFailsSave(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()
	t.Setenv("HOME", fakeHomePth)

	blockingFile := fakeHomePth + "-blocking-file"
	require.Equal(t, nil, os.WriteFile(blockingFile, []byte("x"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", blockingFile)

	require.Error(t, SaveSetupSuccessForVersion("3.0.0"))

	_, exists, err := LoadLegacyConfig()
	require.Equal(t, nil, err)
	require.Equal(t, false, exists)
}

// TestCheckIsSetupWasDoneForVersion_FallsBackToPerDirConfig asserts a value
// living only in the per-directory .bitrise-cli.yml (no legacy, no global
// file) is still observed on read.
func TestCheckIsSetupWasDoneForVersion_FallsBackToPerDirConfig(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	defer func() {
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()
	t.Setenv("HOME", fakeHomePth)
	t.Setenv("XDG_CONFIG_HOME", "")

	projectDir := t.TempDir()
	require.Equal(t, nil, os.WriteFile(
		filepath.Join(projectDir, internalconfig.DirFileName),
		[]byte("setup_version: 4.5.6\n"),
		0o644,
	))
	t.Chdir(projectDir)

	versionMatch, version := CheckIsSetupWasDoneForVersion("4.5.6")
	require.Equal(t, true, versionMatch)
	require.Equal(t, "4.5.6", version)
}
