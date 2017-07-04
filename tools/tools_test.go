package tools

import (
	"path/filepath"
	"testing"
	"log"
	"os"
	"syscall"
	"os/exec"
	"runtime"
	"io/ioutil"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
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
	if runtime.GOOS != "linux" {
		log.Println("Test requires linux")
		return
	}

	dir, err := ioutil.TempDir("", "ramdisk")
	require.Equal(t, nil, err)

	require.NoError(t, syscall.Mount("tmpfs", dir, "tmpfs", 0, "size=16m"))

	filename := "test.tmp"
	srcPath := os.TempDir() + "/" + filename
	srcFile, err := os.Create(srcPath)
	require.NotEqual(t, nil, srcFile)
	require.Equal(t, nil, err)

	dstPath := dir + "/" + filename
	require.NoError(t, MoveFile(srcPath, dstPath))
	info, err:= os.Stat(dstPath)
	require.Equal(t, nil, err)
	require.Equal(t, false, info.IsDir())

	require.NoError(t, exec.Command("umount", dir).Run())
	require.NoError(t, os.RemoveAll(dir))
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