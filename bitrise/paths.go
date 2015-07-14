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

func init() {
	bitriseWorkDirPath, err := filepath.Abs(path.Join("./", ".bitrise"))
	if err != nil {
		log.Fatal("Failed to set bitrise work dir path:", err)
	}
	if exist, err := pathutil.IsPathExists(bitriseWorkDirPath); err != nil {
		log.Fatal("Failed to check bitrise work dir path:", err)
	} else if !exist {
		if err := os.MkdirAll(bitriseWorkDirPath, 0777); err != nil {
			log.Fatal("Failed to create bitrise work dir path:", err)
		}
	}
	BitriseWorkDirPath = bitriseWorkDirPath

	bitriseWorkStepsDirPath, err1 := filepath.Abs(path.Join(BitriseWorkDirPath, "steps"))
	if err1 != nil {
		log.Fatal("Failed to set bitrise steps work dir path:", err1)
	}
	if exist, err := pathutil.IsPathExists(bitriseWorkStepsDirPath); err != nil {
		log.Fatal("Failed to check bitrise work dir path:", err)
	} else if !exist {
		if err := os.MkdirAll(bitriseWorkStepsDirPath, 0777); err != nil {
			log.Fatal("Failed to create bitrise steps work dir path:", err)
		}
	}
	BitriseWorkStepsDirPath = bitriseWorkStepsDirPath

	envstorePath, err2 := filepath.Abs(path.Join(bitriseWorkDirPath, "envstore.yml"))
	if err2 != nil {
		log.Fatal("Failed to set envstore path:", err2)
	}
	EnvstorePath = envstorePath

	formoutPath, err3 := filepath.Abs(path.Join(bitriseWorkDirPath, "formout.md"))
	if err3 != nil {
		log.Fatal("Failed to set formatted output path:", err3)
	}
	FormattedOutputPath = formoutPath
}
