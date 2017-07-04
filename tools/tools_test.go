package tools

import (
	"path/filepath"
	"testing"
	"os"
	"os/exec"
	"runtime"
	"io/ioutil"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
	"bytes"
	"strings"
)

func TestMoveFile(t *testing.T) {
	srcPath := os.TempDir() + "/src.tmp"
	_, err := os.Create(srcPath)
	require.Equal(t, nil, err)

	dstPath := os.TempDir() + "/dst.tmp"

	require.NoError(t, MoveFile(srcPath, dstPath))

	info, err:= os.Stat(dstPath)
	require.Equal(t, nil, err)
	require.Equal(t, false, info.IsDir())

	require.NoError(t, os.Remove(dstPath))
}

func TestMoveFileDifferentDevices(t *testing.T) {
	require.True(t, runtime.GOOS == "linux" || runtime.GOOS == "darwin")

	ramdiskPath := ""
	ramdiskName := "RAMDISK"
	volumeName := ""
	if runtime.GOOS == "linux" {
		dir, err := ioutil.TempDir("", ramdiskName)
		require.Equal(t, nil, err)
		require.NoError(t, exec.Command("mount", "-t", "tmpfs", "-o", "size=12m", "tmpfs", dir).Run())
	} else if runtime.GOOS == "darwin" {
		var stdout bytes.Buffer
		cmd := exec.Command("hdiutil", "attach", "-nomount", "ram://64")
		cmd.Stdout = &stdout
		require.NoError(t, cmd.Run())
		volumeName =  strings.TrimSpace(stdout.String())

		require.NoError(t, exec.Command("diskutil", "erasevolume", "MS-DOS", ramdiskName, volumeName).Run())
		ramdiskPath = "/Volumes/" + ramdiskName
	}

	filename := "test.tmp"
	srcPath := os.TempDir() + "/" + filename
	srcFile, err := os.Create(srcPath)
	require.NotEqual(t, nil, srcFile)
	require.Equal(t, nil, err)

	dstPath := ramdiskPath + "/" + filename
	require.NoError(t, MoveFile(srcPath, dstPath))
	info, err:= os.Stat(dstPath)
	require.Equal(t, nil, err)
	require.Equal(t, false, info.IsDir())

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

	outStr, err := StepmanJSONStepLibStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "0.9.0")
	require.Equal(t, nil, err)
	require.NotEqual(t, "", outStr)

	// Invalid params -- Err should empty, output filled
	outStr, err = StepmanJSONStepLibStepInfo("https://github.com/bitrise-io/bitrise-steplib", "script", "2")
	require.NotEqual(t, nil, err)
	require.Equal(t, "", outStr)
}

func TestEnvmanJSONPrint(t *testing.T) {
	// Initialized envstore -- Err should empty, output filled
	testDirPth, err := pathutil.NormalizedOSTempDirPath("test_env_store")
	require.Equal(t, nil, err)

	envstorePth := filepath.Join(testDirPth, "envstore.yml")

	require.Equal(t, nil, EnvmanInitAtPath(envstorePth))

	outStr, err := EnvmanJSONPrint(envstorePth)
	require.Equal(t, nil, err)
	require.NotEqual(t, "", outStr)

	// Not initialized envstore -- Err should filled, output empty
	testDirPth, err = pathutil.NormalizedOSTempDirPath("test_env_store")
	require.Equal(t, nil, err)

	envstorePth = filepath.Join("test_env_store", "envstore.yml")

	outStr, err = EnvmanJSONPrint(envstorePth)
	require.NotEqual(t, nil, err)
	require.Equal(t, "", outStr)
}