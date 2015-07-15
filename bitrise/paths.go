package bitrise

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
)

var (
	// EnvstorePath ...
	EnvstorePath string
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
	EnvstorePathEnvKey string = "ENVMAN_ENVSTORE_PATH"
	// FormattedOutputPathEnvKey ...
	FormattedOutputPathEnvKey string = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
)

// CleanupBitriseWorkPath ...
func CleanupBitriseWorkPath() error {
	if exist, err := pathutil.IsPathExists(BitriseWorkDirPath); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(BitriseWorkDirPath); err != nil {
			return err
		}
	}
	return initBitriseWorkPaths()
}

func initBitriseWorkPaths() error {
	bitriseWorkDirPath, err := filepath.Abs("./.bitrise")
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

	bitriseWorkStepsDirPath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "steps"))
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

func init() {
	if err := initBitriseWorkPaths(); err != nil {
		log.Fatal("Failed to init bitrise paths:", err)
	}

	envstorePath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "envstore.yml"))
	if err != nil {
		log.Fatal("Failed to set envstore path:", err)
	}
	EnvstorePath = envstorePath

	formoutPath, err := filepath.Abs(path.Join(BitriseWorkDirPath, "formout.md"))
	if err != nil {
		log.Fatal("Failed to set formatted output path:", err)
	}
	FormattedOutputPath = formoutPath

	currentDir, err := filepath.Abs("./")
	if err != nil {
		log.Fatal("Failed to set current dir:", err)
	}
	CurrentDir = currentDir
}
