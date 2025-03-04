package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestMoveFile(t *testing.T) {
	srcPath := filepath.Join(os.TempDir(), "src.tmp")
	_, err := os.Create(srcPath)
	require.NoError(t, err)

	dstPath := filepath.Join(os.TempDir(), "dst.tmp")
	require.NoError(t, MoveFile(srcPath, dstPath))

	info, err := os.Stat(dstPath)
	require.NoError(t, err)
	require.False(t, info.IsDir())

	require.NoError(t, os.Remove(dstPath))
}

func TestMoveFileDifferentDevices(t *testing.T) {
	require.True(t, runtime.GOOS == "linux" || runtime.GOOS == "darwin")

	ramdiskPath := ""
	ramdiskName := "RAMDISK"
	volumeName := ""
	if runtime.GOOS == "linux" {
		tmpDir := t.TempDir()

		ramdiskPath = tmpDir
		require.NoError(t, exec.Command("mount", "-t", "tmpfs", "-o", "size=12m", "tmpfs", ramdiskPath).Run())
	} else if runtime.GOOS == "darwin" {
		out, err := command.New("hdiutil", "attach", "-nomount", "ram://64").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err)

		volumeName = out
		require.NoError(t, exec.Command("diskutil", "erasevolume", "MS-DOS", ramdiskName, volumeName).Run())

		ramdiskPath = "/Volumes/" + ramdiskName
	}

	filename := "test.tmp"
	srcPath := filepath.Join(os.TempDir(), filename)
	_, err := os.Create(srcPath)
	require.NoError(t, err)

	dstPath := filepath.Join(ramdiskPath, filename)
	require.NoError(t, MoveFile(srcPath, dstPath))

	info, err := os.Stat(dstPath)
	require.NoError(t, err)
	require.False(t, info.IsDir())

	if runtime.GOOS == "linux" {
		require.NoError(t, exec.Command("umount", ramdiskPath).Run())
		require.NoError(t, os.RemoveAll(ramdiskPath))
	} else if runtime.GOOS == "darwin" {
		require.NoError(t, exec.Command("hdiutil", "detach", volumeName).Run())
	}
}

func TestStepmanJSONStepLibStepInfo(t *testing.T) {
	// setup
	require.NoError(t, configs.InitPaths())

	// Valid params -- Err should empty, output filled
	require.Equal(t, nil, StepmanSetup("https://github.com/bitrise-io/bitrise-steplib"))

	info, err := StepmanStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "0.9.0")
	require.NoError(t, err)
	require.NotEqual(t, "", info.JSON())

	// Invalid params -- Err returned, output is invalid
	info, err = StepmanStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "2.x")
	require.Error(t, err)
}

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
