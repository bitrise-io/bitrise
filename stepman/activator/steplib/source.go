package steplib

import (
	"context"
	"errors"
	"fmt"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
)

// activateStepSourceWithAPI materializes id@version's source into destDir
// without cloning a git steplib.
func activateStepSourceWithAPI(libraryAPI *steplibrary.Client, id, version string, source *models.StepSourceModel, destDir string, log stepman.Logger, isOfflineMode bool) error {
	if isOfflineMode {
		return errors.New("offline mode is not supported with Steplib API")
	}

	if source == nil || source.Git == "" {
		return fmt.Errorf("step %s@%s has no source git URL to download from", id, version)
	}

	locations, err := libraryAPI.StepSourceDownloadLocations(context.Background(), id, version, source.Git)
	if err != nil {
		return fmt.Errorf("resolve download locations for %s@%s: %s", id, version, err)
	}
	if len(locations) == 0 {
		return fmt.Errorf("step %s@%s has no download location", id, version)
	}

	if err := stepman.DownloadStepSourceArchive(destDir, locations, id, version, source.Commit, log); err != nil {
		return fmt.Errorf("download step source %s@%s: %s", id, version, err)
	}
	return nil
}
