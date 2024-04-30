package toolkits

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/go-utils/command"
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
