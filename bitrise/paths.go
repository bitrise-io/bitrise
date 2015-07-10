package bitrise

import (
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

var (
	// EnvstorePath ...
	EnvstorePath string
	// FormattedOutputPath ...
	FormattedOutputPath string
)

const (
	// EnvstorePathEnvKey ...
	EnvstorePathEnvKey string = "ENVMAN_ENVSTORE_PATH"
	// FormattedOutputPathEnvKey ...
	FormattedOutputPathEnvKey string = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
)

func init() {
	envstorePath, err := filepath.Abs(path.Join("./", "envstore.yml"))
	if err != nil {
		log.Fatal("Failed to set envstore path:", err)
	}
	EnvstorePath = envstorePath

	formoutPath, e := filepath.Abs(path.Join("./", "formout.md"))
	if e != nil {
		log.Fatal("Failed to set formatted output path:", e)
	}
	FormattedOutputPath = formoutPath
}
