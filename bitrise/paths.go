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
	bitriseWorkDirPath, err := filepath.Abs(path.Join("./", ".bitrise"))
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

	bitriseWorkStepsDirPath, err1 := filepath.Abs(path.Join(BitriseWorkDirPath, "steps"))
	if err1 != nil {
		return err1
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

	envstorePath, err2 := filepath.Abs(path.Join(BitriseWorkDirPath, "envstore.yml"))
	if err2 != nil {
		log.Fatal("Failed to set envstore path:", err2)
	}
	EnvstorePath = envstorePath

	formoutPath, err3 := filepath.Abs(path.Join(BitriseWorkDirPath, "formout.md"))
	if err3 != nil {
		log.Fatal("Failed to set formatted output path:", err3)
	}
	FormattedOutputPath = formoutPath
}
