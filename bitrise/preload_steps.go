package bitrise

import (
	"fmt"
	"sync"
	"time"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"

	// parallel download parameters
	poolSize = 10
	timeout  = 30 * time.Second
)

// preloadBitriseSteps preloads the cache with Bitrise owned steps
func preloadBitriseSteps(log stepman.Logger) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(bitriseStepLibURL); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(bitriseStepLibURL, log); err != nil {
			return fmt.Errorf("Failed to setup steplib: %w", err)
		}
	}

	stepLib, err := stepman.ReadStepSpec(bitriseStepLibURL)
	if err != nil {
		return err
	}

	waitGroup := &sync.WaitGroup{}

	// download parallel 10 steps in goroutines
	for stepID, step := range stepLib.Steps {
		waitGroup.Add(1)

		go func(stepID string, step models.StepGroupModel) {
			if step.Info.Maintainer != bitriseMaintainer {
				log.Warnf("Skipping step %s as it is not maintained by Bitrise", stepID)
			}

			latestVersion, found := step.LatestVersion()
			if !found {
				log.Warnf("Failed to find latest version for step %s", stepID)
			}

			_, _, err := preloadStep(stepLib, bitriseStepLibURL, stepID, step.LatestVersionNumber, latestVersion, log)
			if err != nil {
				log.Warnf("Failed to download step %s: %w", stepID, err)
			}

			waitGroup.Done()
		}(stepID, step)
	}

	waitGroup.Wait()

	return nil
}

func preloadStep(stepLib models.StepCollectionModel, stepLibURI, id, version string, step models.StepModel, log stepman.Logger) (string, string, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return "", "", fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	// is precompiled uncompressed step version in cache?
	targetExecutablePath := stepman.GetStepCacheExecutablePathForVersion(route, id, version)
	// checkSumPath := stepman.GetStepCacheExecutableChecksumPathForVersion(route, id, version)
	if exist, err := pathutil.IsPathExists(targetExecutablePath); err != nil {
		return "", "", fmt.Errorf("failed to check if %s path exist: %s", targetExecutablePath, err)
	} else if exist {
		// check checksum
		return "", targetExecutablePath, nil
	}

	// Compile Step, calclulate checksum
	stepSourceDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepSourceDir); err != nil {
		return "", "", fmt.Errorf("failed to check if %s path exist: %s", stepSourceDir, err)
	} else if exist { // version specific source cache exists
		return stepSourceDir, "", nil
	}

	// version specific source cache not exists
	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return "", "", fmt.Errorf("download failed: %s", err)
	}

	// err := toolkits.BuildGoStep(targetExecutablePath, step, stepSourceDir)
	// if err != nil {
	// 	return "", "", fmt.Errorf("failed to build step: %s", err)
	// }

	return stepSourceDir, "", nil
}
