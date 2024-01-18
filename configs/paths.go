package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

var (
	InputEnvstorePath       string
	OutputEnvstorePath      string
	FormattedOutputPath     string
	BitriseWorkDirPath      string
	BitriseWorkStepsDirPath string
	CurrentDir              string
)

const (
	EnvstorePathEnvKey         = "ENVMAN_ENVSTORE_PATH"
	FormattedOutputPathEnvKey  = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
	BitriseDataHomeDirEnvKey   = "BITRISE_DATA_HOME_DIR"
	BitriseSourceDirEnvKey     = "BITRISE_SOURCE_DIR"
	BitriseDeployDirEnvKey     = "BITRISE_DEPLOY_DIR"
	BitriseTestDeployDirEnvKey = "BITRISE_TEST_DEPLOY_DIR"
	// BitrisePerStepTestResultDirEnvKey is a unique subdirectory in BITRISE_TEST_DEPLOY_DIR for each step run, steps should place test reports and attachments into this directory
	BitrisePerStepTestResultDirEnvKey = "BITRISE_TEST_RESULT_DIR"
	BitriseTmpDirEnvKey               = "BITRISE_TMP_DIR"
	BitriseHtmlReportDirEnvKey        = "BITRISE_HTML_REPORT_DIR"
)

func GetBitriseHomeDirPath() string {
	return filepath.Join(pathutil.UserHomeDir(), ".bitrise")
}

func getBitriseConfigFilePath() string {
	return filepath.Join(GetBitriseHomeDirPath(), bitriseConfigFileName)
}

func GetBitriseToolsDirPath() string {
	return filepath.Join(GetBitriseHomeDirPath(), "tools")
}

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
		if err := os.MkdirAll(bitriseWorkDirPath, 0755); err != nil {
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
		if err := os.MkdirAll(bitriseWorkStepsDirPath, 0755); err != nil {
			return err
		}
	}
	BitriseWorkStepsDirPath = bitriseWorkStepsDirPath

	return nil
}

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

func InitPaths() error {
	if err := initBitriseWorkPaths(); err != nil {
		return fmt.Errorf("init bitrise paths: %s", err)
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

	// BITRISE_HTML_REPORT_DIR
	if os.Getenv(BitriseHtmlReportDirEnvKey) == "" {
		reportDir, err := pathutil.NormalizedOSTempDirPath("html-reports")
		if err != nil {
			return fmt.Errorf("Failed to create html-reports dir, error: %s", err)
		}

		if err := os.Setenv(BitriseHtmlReportDirEnvKey, reportDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_HTML_REPORT_DIR, error: %s", err)
		}
	}

	// BITRISE_TEST_RESULTS_DIR
	if os.Getenv(BitriseTestDeployDirEnvKey) == "" {
		testsDir, err := pathutil.NormalizedOSTempDirPath("test_results")
		if err != nil {
			return fmt.Errorf("Failed to set deploy dir, error: %s", err)
		}

		if err := os.Setenv(BitriseTestDeployDirEnvKey, testsDir); err != nil {
			return fmt.Errorf("Failed to set %s, error: %s", BitriseTestDeployDirEnvKey, err)
		}
	}

	//BITRISE_TMP_DIR
	if os.Getenv(BitriseTmpDirEnvKey) == "" {
		tmpDir, err := pathutil.NormalizedOSTempDirPath("tmp")
		if err != nil {
			return fmt.Errorf("Failed to set tmp dir, error: %s", err)
		}

		if err := os.Setenv(BitriseTmpDirEnvKey, tmpDir); err != nil {
			return fmt.Errorf("Failed to set BITRISE_TMP_DIR, error: %s", err)
		}
	}

	return nil
}
