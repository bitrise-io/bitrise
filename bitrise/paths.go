package bitrise

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
)

var (
	// InputEnvstorePath ...
	InputEnvstorePath string
	// OutputEnvstorePath ...
	OutputEnvstorePath string
	// FormattedOutputPath ...
	FormattedOutputPath string
	// BitriseWorkDirPath ...
	BitriseWorkDirPath string
	// BitriseWorkStepsDirPath ...
	BitriseWorkStepsDirPath string
	// CurrentDir ...
	CurrentDir string
)

const (
	// EnvstorePathEnvKey ...
	EnvstorePathEnvKey = "ENVMAN_ENVSTORE_PATH"
	// FormattedOutputPathEnvKey ...
	FormattedOutputPathEnvKey = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
	// BitriseSourceDirEnvKey ...
	BitriseSourceDirEnvKey = "BITRISE_SOURCE_DIR"
	// BitriseDeployDirEnvKey ...
	BitriseDeployDirEnvKey = "BITRISE_DEPLOY_DIR"
)

func initBitriseWorkPaths() error {
	bitriseWorkDirPath, err := pathutil.NormalizedOSTempDirPath("bitrise")
	if err != nil {
		return err
	}
	if exist, err := pathutil.IsPathExists(bitriseWorkDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(bitriseWorkDirPath, 0777); err != nil {
			return err
		}
	}
	BitriseWorkDirPath = bitriseWorkDirPath

	bitriseWorkStepsDirPath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "step_src"))
	if err != nil {
		return err
	}
	if exist, err := pathutil.IsPathExists(bitriseWorkStepsDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(bitriseWorkStepsDirPath, 0777); err != nil {
			return err
		}
	}
	BitriseWorkStepsDirPath = bitriseWorkStepsDirPath

	return nil
}

// InitPaths ...
func InitPaths() error {
	if err := initBitriseWorkPaths(); err != nil {
		return fmt.Errorf("Failed to init bitrise paths: %s", err)
	}

	inputEnvstorePath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "input_envstore.yml"))
	if err != nil {
		return fmt.Errorf("Failed to set envstore path: %s", err)
	}
	InputEnvstorePath = inputEnvstorePath

	outputEnvstorePath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "output_envstore.yml"))
	if err != nil {
		return fmt.Errorf("Failed to set envstore path: %s", err)
	}
	OutputEnvstorePath = outputEnvstorePath

	formoutPath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "formatted_output.md"))
	if err != nil {
		return fmt.Errorf("Failed to set formatted output path: %s", err)
	}
	FormattedOutputPath = formoutPath

	currentDir, err := filepath.Abs("./")
	if err != nil {
		return fmt.Errorf("Failed to set current dir: %s", err)
	}
	CurrentDir = currentDir

	// BITRISE_SOURCE_DIR
	if os.Getenv(BitriseSourceDirEnvKey) == "" {
		if err := os.Setenv(BitriseSourceDirEnvKey, currentDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_SOURCE_DIR: %s", err)
		}
	}

	// BITRISE_DEPLOY_DIR
	if os.Getenv(BitriseDeployDirEnvKey) == "" {
		deployDir, err := pathutil.NormalizedOSTempDirPath("deploy")
		if err != nil {
			return fmt.Errorf("Failed to set deploy dir: %s", err)
		}

		if err := os.Setenv(BitriseDeployDirEnvKey, deployDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_DEPLOY_DIR: %s", err)
		}
	}

	return nil
}
