package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/envman/models"

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
