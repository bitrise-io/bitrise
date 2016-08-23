package toolkits

import (
	"path/filepath"

	"github.com/bitrise-io/bitrise/configs"
)

// Toolkit ...
type Toolkit interface {
	Install() error
}

//
// === Utils ===

func getBitriseToolkitsTmpDirPath() string {
	bitriseToolkitsDirPath := configs.GetBitriseToolkitsDirPath()
	return filepath.Join(bitriseToolkitsDirPath, "tmp")
}
