package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	// BitriseCacheDirEnvKey ...
	BitriseCacheDirEnvKey = "BITRISE_CACHE_DIR"
)

// GetBitriseHomeDirPath ...
func GetBitriseHomeDirPath() string {
	return filepath.Join(pathutil.UserHomeDir(), ".bitrise")
}

func getBitriseConfigFilePath() string {
	return filepath.Join(GetBitriseHomeDirPath(), bitriseConfigFileName)
}

// GetBitriseToolsDirPath ...
func GetBitriseToolsDirPath() string {
	return filepath.Join(GetBitriseHomeDirPath(), "tools")
}

// GetBitriseToolkitsDirPath ...
func GetBitriseToolkitsDirPath() string {
	return filepath.Join(GetBitriseHomeDirPath(), "toolkits")
}

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

	bitriseWorkStepsDirPath, err := filepath.Abs(filepath.Join(BitriseWorkDirPath, "step_src"))
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

// GeneratePATHEnvString ...
func GeneratePATHEnvString(currentPATHEnv, pathToInclude string) string {
	if currentPATHEnv == "" {
		return pathToInclude
	}
	if pathToInclude == "" {
		return currentPATHEnv
	}
	if pathToInclude == currentPATHEnv {
		return currentPATHEnv
	}

	pthWithPathIncluded := currentPATHEnv
	if !strings.HasSuffix(pthWithPathIncluded, pathToInclude) &&
		!strings.Contains(pthWithPathIncluded, pathToInclude+":") {
		pthWithPathIncluded = pathToInclude + ":" + pthWithPathIncluded
	}
	return pthWithPathIncluded
}

// InitPaths ...
func InitPaths() error {
	if err := initBitriseWorkPaths(); err != nil {
		return fmt.Errorf("Failed to init bitrise paths, error: %s", err)
	}

	// --- Bitrise TOOLS
	{
		bitriseToolsDirPth := GetBitriseToolsDirPath()
		if err := pathutil.EnsureDirExist(bitriseToolsDirPth); err != nil {
			return err
		}
		pthWithBitriseTools := GeneratePATHEnvString(os.Getenv("PATH"), bitriseToolsDirPth)

		if IsDebugUseSystemTools() {
			log.Warn("[BitriseDebug] Using system tools, instead of the ones in BITRISE_HOME")
		} else {
			if err := os.Setenv("PATH", pthWithBitriseTools); err != nil {
				return fmt.Errorf("Failed to set PATH to include BITRISE_HOME/tools! Error: %s", err)
			}
		}
	}

	inputEnvstorePath, err := filepath.Abs(filepath.Join(BitriseWorkDirPath, "input_envstore.yml"))
	if err != nil {
		return fmt.Errorf("Failed to set input envstore path, error: %s", err)
	}
	InputEnvstorePath = inputEnvstorePath

	outputEnvstorePath, err := filepath.Abs(filepath.Join(BitriseWorkDirPath, "output_envstore.yml"))
	if err != nil {
		return fmt.Errorf("Failed to set output envstore path, error: %s", err)
	}
	OutputEnvstorePath = outputEnvstorePath

	formoutPath, err := filepath.Abs(filepath.Join(BitriseWorkDirPath, "formatted_output.md"))
	if err != nil {
		return fmt.Errorf("Failed to set formatted output path, error: %s", err)
	}
	FormattedOutputPath = formoutPath

	currentDir, err := filepath.Abs("./")
	if err != nil {
		return fmt.Errorf("Failed to set current dir, error: %s", err)
	}
	CurrentDir = currentDir

	// BITRISE_SOURCE_DIR
	if os.Getenv(BitriseSourceDirEnvKey) == "" {
		if err := os.Setenv(BitriseSourceDirEnvKey, currentDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_SOURCE_DIR, error: %s", err)
		}
	}

	// BITRISE_DEPLOY_DIR
	if os.Getenv(BitriseDeployDirEnvKey) == "" {
		deployDir, err := pathutil.NormalizedOSTempDirPath("deploy")
		if err != nil {
			return fmt.Errorf("Failed to set deploy dir, error: %s", err)
		}

		if err := os.Setenv(BitriseDeployDirEnvKey, deployDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_DEPLOY_DIR, error: %s", err)
		}
	}

	// BITRISE_CACHE_DIR
	if os.Getenv(BitriseCacheDirEnvKey) == "" {
		cacheDir, err := pathutil.NormalizedOSTempDirPath("cache")
		if err != nil {
			return fmt.Errorf("Failed to set cache dir, error: %s", err)
		}

		if err := os.Setenv(BitriseCacheDirEnvKey, cacheDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_CACHE_DIR, error: %s", err)
		}
	}

	return nil
}
