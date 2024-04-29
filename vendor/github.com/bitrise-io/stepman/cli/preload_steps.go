package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
	workers           = 10
)

type GoBuilder func(stepSourceAbsPath, packageName, targetExecutablePath string) error

type preloadResult struct {
	stepID  string
	version string
	err     error
}

// PreloadBitriseSteps preloads the cache with Bitrise owned steps
func PreloadBitriseSteps(goBuilder GoBuilder, log stepman.Logger) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(bitriseStepLibURL); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(bitriseStepLibURL, log); err != nil {
			return fmt.Errorf("failed to setup steplib: %w", err)
		}
	}

	stepLib, err := stepman.ReadStepSpec(bitriseStepLibURL)
	if err != nil {
		return err
	}

	// channel to collect preload results
	// preloadResults := make(chan preloadResult, len(stepLib.Steps))
	// channel to kick off 10 steps preloading at a time
	type stepWorkInfo struct {
		stepID string
		step   models.StepGroupModel
	}
	preloadQueue := make(chan stepWorkInfo, workers)

	// start preload workers
	waitGroup := &sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		waitGroup.Add(1)
		go func() {
			for s := range preloadQueue {
				if err := preloadStepVersions(log, goBuilder, stepLib, s.stepID, s.step); err != nil {
					log.Warnf("%w", err)
				}
			}

			waitGroup.Done()
		}()
	}

	for stepID, step := range stepLib.Steps {
		if step.Info.Maintainer != bitriseMaintainer {
			log.Infof("Skipping step %s as it is not maintained by Bitrise", stepID)
			continue
		}
		if step.Info.DeprecateNotes != "" {
			log.Infof("Skipping deprecated step %s", stepID)
			continue
		}

		preloadQueue <- stepWorkInfo{
			stepID: stepID,
			step:   step,
		}
	}
	close(preloadQueue)

	waitGroup.Wait()
	return nil
}

func preloadStepVersions(log stepman.Logger, goBuilder GoBuilder, stepLib models.StepCollectionModel, stepID string, step models.StepGroupModel) error {
	route, found := stepman.ReadRoute(bitriseStepLibURL)
	if !found {
		return fmt.Errorf("no route found for %s steplib", bitriseStepLibURL)
	}

	latestVersionNumber := step.LatestVersionNumber
	latestVersion, found := step.LatestVersion()
	if !found {
		return fmt.Errorf("failed to find latest version for step %s", stepID)
	}

	log.Infof("Preloading step %s@%s", stepID, latestVersionNumber)
	targetExecutablePathLatest, err := preloadStepExecutable(stepLib, bitriseStepLibURL, goBuilder, stepID, step.LatestVersionNumber, latestVersion, log, false)
	if err != nil {
		return fmt.Errorf("failed to preload step %s@%s: %w", stepID, latestVersionNumber, err)
	}

	filteredSteps, err := filterPreloadedStepVersions(stepID, step.Versions)
	if err != nil {
		return fmt.Errorf("failed to filter preloaded step versions: %w", err)
	}

	// Iterate over all versions and compress them if golang step
	for version, step := range filteredSteps {
		if version == latestVersionNumber {
			continue
		}

		log.Infof("Preloading step %s@%s", stepID, version)
		targetExecutablePath, err := preloadStepExecutable(stepLib, bitriseStepLibURL, goBuilder, stepID, version, step, log, true)
		if err != nil {
			log.Warnf("failed to preload step %s@%s: %w", stepID, version, err)
		}

		if targetExecutablePath != "" && targetExecutablePathLatest != "" {
			log.Infof("Compressing step %s@%s", stepID, version)
			patchFilePath := stepman.GetStepCompressedExecutablePathForVersion(latestVersionNumber, route, stepID, version)
			if err := compressStep(patchFilePath, targetExecutablePathLatest, targetExecutablePath); err != nil {
				return fmt.Errorf("failed to compress step %s@%s: %w", stepID, version, err)
			}
		}
	}

	return nil
}

func preloadStepExecutable(stepLib models.StepCollectionModel, stepLibURI string, goBuilder GoBuilder, id, version string, step models.StepModel, log stepman.Logger, cleanupSrc bool) (string, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	// Clean precompiled uncompressed step version
	targetExecutablePath := stepman.GetStepExecutablePathForVersion(route, id, version)
	exists, err := pathutil.IsPathExists(targetExecutablePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", targetExecutablePath, err)
	}
	if exists {
		if err := os.Remove(targetExecutablePath); err != nil {
			return "", fmt.Errorf("failed to remove %s: %s", targetExecutablePath, err)
		}
	}

	// Clean existing step source
	stepSourceDir := stepman.GetStepCacheDirPath(route, id, version)
	sourceExist, err := pathutil.IsPathExists(stepSourceDir)
	if err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", stepSourceDir, err)
	}
	if sourceExist {
		if err := os.RemoveAll(stepSourceDir); err != nil {
			return "", fmt.Errorf("failed to remove step source dir: %s", err)
		}
	}

	// Fetch source, compile step (if golang), calclulate checksum
	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return "", fmt.Errorf("download failed: %s", err)
	}

	if step.Toolkit == nil || step.Toolkit.Go == nil {
		return "", nil
	}

	if err := goBuilder(stepSourceDir, step.Toolkit.Go.PackageName, targetExecutablePath); err != nil {
		return "", fmt.Errorf("failed to build step: %s", err)
	}

	checkSumPath := stepman.GetStepExecutableChecksumPathForVersion(route, id, version)
	if err := writeChecksum(targetExecutablePath, checkSumPath); err != nil {
		return "", fmt.Errorf("failed to write checksum: %s", err)
	}

	if cleanupSrc {
		// remove step source as build is successful
		// also remove if not successful, as propably old step source does not work anymore
		if err := os.RemoveAll(stepSourceDir); err != nil {
			return "", fmt.Errorf("failed to remove step source dir: %s", err)
		}
	}

	return targetExecutablePath, nil
}
