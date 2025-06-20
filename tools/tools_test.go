package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/configs"
	envmanCLI "github.com/bitrise-io/envman/v2/cli"
	envmanEnv "github.com/bitrise-io/envman/v2/env"
	"github.com/bitrise-io/envman/v2/envman"
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

func TestEnvmanAddEnvs(t *testing.T) {
	defaultConfig, err := envman.GetConfigs()
	require.NoError(t, err)

	tests := []struct {
		name        string
		envstorePth string
		envsList    []models.EnvironmentItemModel
		wantErr     string
	}{
		{
			name:        "add valid envs",
			envstorePth: filepath.Join(os.TempDir(), "envstore.yml"),
			envsList:    []models.EnvironmentItemModel{{"key_1": "value_1"}, {"key_2": "value_2"}},
		},
		{
			name:        "add invalid envs",
			envstorePth: filepath.Join(os.TempDir(), "envstore.yml"),
			envsList:    []models.EnvironmentItemModel{{"key_1": "value_1", "key_2": "value_2"}},
			wantErr:     "more than 1 environment key specified: [key_1 key_2]",
		},
		{
			name:        "add too large env",
			envstorePth: filepath.Join(os.TempDir(), "envstore.yml"),
			envsList:    []models.EnvironmentItemModel{{"key": strings.Repeat("a", defaultConfig.EnvBytesLimitInKB*1024+1)}},
			wantErr: `env var (key) value is too large (256.0009765625 KB), max allowed size: 256 KB.
To increase env var limits please visit: https://support.bitrise.io/en/articles/9676692-env-var-value-too-large-env-var-list-too-large`,
		},
		{
			name:        "add env to a too large env list",
			envstorePth: filepath.Join(os.TempDir(), "envstore.yml"),
			envsList:    []models.EnvironmentItemModel{{"key_1": strings.Repeat("a", defaultConfig.EnvBytesLimitInKB*1024)}, {"key_2": "a"}},
			wantErr: `env var list is too large (256.0009765625 KB), max allowed size: 256 KB.
To increase env var limits please visit: https://support.bitrise.io/en/articles/9676692-env-var-value-too-large-env-var-list-too-large`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnvmanInit(tt.envstorePth, true)
			require.NoError(t, err)

			err = EnvmanAddEnvs(tt.envstorePth, tt.envsList)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)

				gotEnvs, err := EnvmanReadEnvList(tt.envstorePth)
				require.NoError(t, err)

				wantEnvs, err := envmanCLI.ConvertToEnvsJSONModel(tt.envsList, true, false, &envmanEnv.DefaultEnvironmentSource{})
				require.Equal(t, wantEnvs, gotEnvs)
			}

		})
	}
}

func Test_createGitHubBinDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		githubUser  string
		toolName    string
		toolVersion string
		unameGOOS   string
		unameGOARCH string
		want        string
	}{
		{
			name:        "envman pre 2.5.2 version",
			githubUser:  "bitrise-io",
			toolName:    "envman",
			toolVersion: "2.5.1",
			unameGOOS:   "Darwin",
			unameGOARCH: "arm64",
			want:        "https://github.com/bitrise-io/envman/releases/download/2.5.1/envman-Darwin-arm64",
		},
		{
			name:        "envman post 2.5.2 version",
			githubUser:  "bitrise-io",
			toolName:    "envman",
			toolVersion: "2.5.2",
			unameGOOS:   "Darwin",
			unameGOARCH: "arm64",
			want:        "https://github.com/bitrise-io/envman/releases/download/v2.5.2/envman-Darwin-arm64",
		},
		{
			name:        "stepman pre 0.17.2 version",
			githubUser:  "bitrise-io",
			toolName:    "stepman",
			toolVersion: "0.17.1",
			unameGOOS:   "Darwin",
			unameGOARCH: "arm64",
			want:        "https://github.com/bitrise-io/stepman/releases/download/0.17.1/stepman-Darwin-arm64",
		},
		{
			name:        "stepman post 0.17.2 version",
			githubUser:  "bitrise-io",
			toolName:    "stepman",
			toolVersion: "0.17.2",
			unameGOOS:   "Darwin",
			unameGOARCH: "arm64",
			want:        "https://github.com/bitrise-io/stepman/releases/download/v0.17.2/stepman-Darwin-arm64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createGitHubBinDownloadURL(tt.githubUser, tt.toolName, tt.toolVersion, tt.unameGOOS, tt.unameGOARCH); got != tt.want {
				t.Errorf("createGitHubBinDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
