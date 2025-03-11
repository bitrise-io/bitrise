package tools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	envman "github.com/bitrise-io/envman/v2/cli"
	envmanEnv "github.com/bitrise-io/envman/v2/env"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	stepman "github.com/bitrise-io/stepman/cli"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/mod/semver"
	"golang.org/x/sys/unix"
)

const envVarLimitErrorKnowledgeBaseURL = "https://support.bitrise.io/en/articles/9676692-env-var-value-too-large-env-var-list-too-large"

const (
	EnvmanToolName  = "envman"
	StepmanToolName = "stepman"
)

func UnameGOOS() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin", nil
	case "linux":
		return "Linux", nil
	}
	return "", fmt.Errorf("unsupported platform (%s)", runtime.GOOS)
}

func UnameGOARCH() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", nil
	case "arm64":
		return "arm64", nil
	}
	return "", fmt.Errorf("unsupported architecture (%s)", runtime.GOARCH)
}

func InstallToolFromGitHub(toolName, githubUser, toolVersion string) error {
	unameGOOS, err := UnameGOOS()
	if err != nil {
		return fmt.Errorf("failed to determine OS: %w", err)
	}
	unameGOARCH, err := UnameGOARCH()
	if err != nil {
		return fmt.Errorf("failed to determine ARCH: %w", err)
	}

	downloadURL := createGitHubBinDownloadURL(githubUser, toolName, toolVersion, unameGOOS, unameGOARCH)
	return InstallFromURL(toolName, downloadURL)
}

func createGitHubBinDownloadURL(githubUser, toolName, toolVersion, unameGOOS, unameGOARCH string) string {
	shouldAddVPrefix := false
	if toolName == EnvmanToolName {
		if semver.Compare("v"+toolVersion, "v2.5.2") >= 0 {
			shouldAddVPrefix = true
		}
	} else if toolName == StepmanToolName {
		if semver.Compare("v"+toolVersion, "v0.17.2") >= 0 {
			shouldAddVPrefix = true
		}
	}
	if shouldAddVPrefix {
		toolVersion = "v" + toolVersion
	}

	return "https://github.com/" + githubUser + "/" + toolName + "/releases/download/" + toolVersion + "/" + toolName + "-" + unameGOOS + "-" + unameGOARCH
}

func DownloadFile(downloadURL, targetDirPath string) error {
	outFile, err := os.Create(targetDirPath)
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Warnf("Failed to close (%s)", targetDirPath)
		}
	}()
	if err != nil {
		return fmt.Errorf("create %s: %s", targetDirPath, err)
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	client := retryablehttp.NewClient()
	client.Logger = &httpLogAdaptor{logger: logger}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download %s: %s", downloadURL, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("download %s: %s", downloadURL, resp.Status)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body", downloadURL)
		}
	}()
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}

	return nil
}

func InstallFromURL(toolBinName, downloadURL string) error {
	if len(toolBinName) < 1 {
		return fmt.Errorf("no Tool (bin) Name provided! URL was: %s", downloadURL)
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("__tmp_download_dest__")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir for download destination")
	}
	tmpDestinationPth := filepath.Join(tmpDir, toolBinName)

	if err := DownloadFile(downloadURL, tmpDestinationPth); err != nil {
		return fmt.Errorf("failed to download, error: %s", err)
	}

	bitriseToolsDirPath := configs.GetBitriseToolsDirPath()
	destinationPth := filepath.Join(bitriseToolsDirPath, toolBinName)

	if exist, err := pathutil.IsPathExists(destinationPth); err != nil {
		return fmt.Errorf("failed to check if file exist (%s), error: %s", destinationPth, err)
	} else if exist {
		if err := os.Remove(destinationPth); err != nil {
			return fmt.Errorf("failed to remove file (%s), error: %s", destinationPth, err)
		}
	}

	if err := MoveFile(tmpDestinationPth, destinationPth); err != nil {
		return fmt.Errorf("failed to copy (%s) to (%s), error: %s", tmpDestinationPth, destinationPth, err)
	}

	if err := os.Chmod(destinationPth, 0755); err != nil {
		return fmt.Errorf("failed to make file (%s) executable, error: %s", destinationPth, err)
	}

	return nil
}

// ------------------
// --- Stepman

func StepmanSetup(collection string) error {
	log := log.NewLogger(log.GetGlobalLoggerOpts())
	return stepman.Setup(collection, "", log)
}

func StepmanStepInfo(collection, stepID, stepVersion string) (stepmanModels.StepInfoModel, error) {
	log := log.NewLogger(log.GetGlobalLoggerOpts())
	return stepman.QueryStepInfo(collection, stepID, stepVersion, log)
}

//
// Share

func StepmanShare() error {
	args := []string{"share", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

func StepmanShareAudit() error {
	args := []string{"share", "audit", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

func StepmanShareCreate(tag, git, stepID string) error {
	args := []string{"share", "create", "--tag", tag, "--git", git, "--stepid", stepID, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

func StepmanShareFinish() error {
	args := []string{"share", "finish", "--toolmode"}
	return command.RunCommand("stepman", args...)
}

func StepmanShareStart(collection string) error {
	args := []string{"share", "start", "--collection", collection, "--toolmode"}
	return command.RunCommand("stepman", args...)
}

// ------------------
// --- Envman

func EnvmanInit(envStorePth string, clear bool) error {
	return envman.InitEnvStore(envStorePth, clear)
}

func EnvmanAddEnvs(envstorePth string, envsList []envmanModels.EnvironmentItemModel) error {
	for _, env := range envsList {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return err
		}

		isExpand := envmanModels.DefaultIsExpand
		if opts.IsExpand != nil {
			isExpand = *opts.IsExpand
		}

		skipIfEmpty := envmanModels.DefaultSkipIfEmpty
		if opts.SkipIfEmpty != nil {
			skipIfEmpty = *opts.SkipIfEmpty
		}

		sensitive := envmanModels.DefaultIsSensitive
		if opts.IsSensitive != nil {
			sensitive = *opts.IsSensitive
		}

		if err := envman.AddEnv(envstorePth, key, value, isExpand, false, skipIfEmpty, sensitive); err != nil {
			var envVarValueTooLargeErr envman.EnvVarValueTooLargeError
			var envVarListTooLargeErr envman.EnvVarListTooLargeError
			if errors.As(err, &envVarValueTooLargeErr) || errors.As(err, &envVarListTooLargeErr) {
				return fmt.Errorf("%w.\nTo increase env var limits please visit: %s", err, envVarLimitErrorKnowledgeBaseURL)
			}
			return err
		}
	}
	return nil
}

func EnvmanReadEnvList(envStorePth string) (envmanModels.EnvsJSONListModel, error) {
	return envman.ReadEnvsJSONList(envStorePth, true, false, &envmanEnv.DefaultEnvironmentSource{})
}

func EnvmanClear(envStorePth string) error {
	return envman.ClearEnvs(envStorePth)
}

// ------------------
// --- Utility

// GetSecretKeysAndValues filters out built in configuration parameters from the secret envs
func GetSecretKeysAndValues(secrets []envmanModels.EnvironmentItemModel) ([]string, []string) {
	var secretKeys []string
	var secretValues []string
	for _, secret := range secrets {
		key, value, err := secret.GetKeyValuePair()
		if err != nil || len(value) < 1 || IsBuiltInFlagTypeKey(key) {
			if err != nil {
				log.Warnf("Error getting key-value pair from secret (%v): %s", secret, err)
			}
			continue
		}
		secretKeys = append(secretKeys, key)
		secretValues = append(secretValues, value)
	}

	return secretKeys, secretValues
}

func MoveFile(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err == nil {
		return nil
	}

	if linkErr, ok := err.(*os.LinkError); ok {
		if linkErr.Err == unix.EXDEV {
			info, err := os.Stat(oldpath)
			if err != nil {
				return err
			}

			data, err := os.ReadFile(oldpath)
			if err != nil {
				return err
			}

			err = os.WriteFile(newpath, data, info.Mode())
			if err != nil {
				return err
			}

			return os.Remove(oldpath)
		}
	}

	return err
}

// IsBuiltInFlagTypeKey returns true if the env key is a built-in flag type env key
func IsBuiltInFlagTypeKey(env string) bool {
	switch env {
	case configs.IsSecretFilteringKey,
		configs.IsSecretEnvsFilteringKey,
		configs.CIModeEnvKey,
		configs.PRModeEnvKey,
		configs.DebugModeEnvKey,
		configs.PullRequestIDEnvKey:
		return true
	default:
		return false
	}
}

// httpLogAdaptor adapts the retryablehttp.Logger interface to the log.Logger.
type httpLogAdaptor struct {
	logger log.Logger
}

// Printf implements the retryablehttp.Logger interface
func (a *httpLogAdaptor) Printf(fmtStr string, vars ...interface{}) {
	a.logger.Debugf(fmtStr, vars...)
}
