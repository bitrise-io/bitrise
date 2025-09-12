package mise

import (
	"os"
	"path/filepath"
)

func Dirs(miseVersion string) (installDir, dataDir string) {
	dataHomeDir := os.Getenv("XDG_DATA_HOME")
	if dataHomeDir == "" {
		dataHomeDir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	cacheHomeDir := os.Getenv("XDG_CACHE_HOME")
	if cacheHomeDir == "" {
		cacheHomeDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}

	// We want to use an isolated mise instance with a pinned version, even in
	// local executions, so we install our version to the cache dir:
	installDir = filepath.Join(cacheHomeDir, "bitrise", "toolprovider", "mise", miseVersion)

	// ...but reuse the installed tool versions (if present in local runs) because those
	// are expensive to install and take up a lot of space.
	// Mirror default location: https://mise.jdx.dev/directories.html#local-share-mise
	dataDir = filepath.Join(dataHomeDir, "mise")
	if os.Getenv("MISE_DATA_DIR") != "" {
		dataDir = os.Getenv("MISE_DATA_DIR")
	}

	return installDir, dataDir
}
