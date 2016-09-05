package toolkits

import (
	"path/filepath"

	"github.com/bitrise-io/bitrise/configs"
)

// Toolkit ...
type Toolkit interface {
	// Install the toolkit
	Install() error
	// Bootstrap : initialize the toolkit for use,
	// e.g. setting Env Vars
	Bootstrap() error
}

//
// === Utils ===

func getBitriseToolkitsTmpDirPath() string {
	bitriseToolkitsDirPath := configs.GetBitriseToolkitsDirPath()
	return filepath.Join(bitriseToolkitsDirPath, "tmp")
}
