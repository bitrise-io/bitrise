package cli

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
	workers           = 10
	MonthDuration     = 24 * time.Hour * 31
)

type PreloadOpts struct {
	NumMajor            uint
	NumMinor            uint
	LatestMinorsSince   time.Duration
	PatchesSince        time.Duration
	UseBinaryExecutable bool
}

type GoBuilder func(stepSourceAbsPath, packageName, targetExecutablePath string) error

func DefaultPreloadOpts() PreloadOpts {
	return PreloadOpts{
		NumMajor:            2,
		NumMinor:            1,
		LatestMinorsSince:   2 * MonthDuration,
		PatchesSince:        1 * MonthDuration,
		UseBinaryExecutable: false,
	}
}

type stepWorkInfo struct {
	stepID string
	step   models.StepGroupModel
}

type preloadResult struct {
	stepID  string
	version string
	status  string
	err     error
}

// PreloadBitriseSteps preloads the cache with Bitrise owned steps
func PreloadBitriseSteps(log stepman.Logger, goBuilder GoBuilder, opts PreloadOpts) error {
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

	preloadQueue := make(chan stepWorkInfo)
	preloadResults := make(chan preloadResult)
	errC := make(chan error)

	workersWaitGroup := &sync.WaitGroup{}
	resultsWaitGroup := &sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		workersWaitGroup.Add(1)
		go func() {
			for s := range preloadQueue {
				results, err := preloadStepVersions(log, goBuilder, stepLib, s.stepID, s.step, opts)
				if err != nil {
					log.Debugf("Failed to preload step %s: %s", s.stepID, err)
					errC <- err
				}

				for _, result := range results {
					log.Debugf("Preloading step %s@%s finished, status: %s %s", result.stepID, result.version, result.status, result.err)
					preloadResults <- result
				}
			}

			workersWaitGroup.Done()
		}()
	}

	go func() {
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
	}()

	results := map[string][]preloadResult{}
	resultsWaitGroup.Add(1)
	go func() {
		for r := range preloadResults {
			if r.err != nil {
				log.Warnf("Failed to preload step %s@%s: %s", r.stepID, r.version, r.err)
			}

			results[r.stepID] = append(results[r.stepID], r)
		}

		resultsWaitGroup.Done()
	}()

	workersWaitGroup.Wait()
	close(preloadResults)
	resultsWaitGroup.Wait()

	close(errC)
	for err := range errC {
		return err
	}

	log.Infof("\n=== Results ===\n")
	for _, stepResults := range results {
		for _, result := range stepResults {
			status := colorstring.Green(result.status)
			if result.err != nil {
				status = colorstring.Red(fmt.Sprintf("Failed: %s", result.err))
			}
			log.Infof("Preloading step %s@%s finished: %s", result.stepID, result.version, status)
			if result.err != nil {
				return result.err
			}
		}
	}

	return nil
}

func preloadStepVersions(log stepman.Logger, goBuilder GoBuilder, stepLib models.StepCollectionModel, stepID string, step models.StepGroupModel, opts PreloadOpts) ([]preloadResult, error) {
	results := []preloadResult{}

	route, found := stepman.ReadRoute(bitriseStepLibURL)
	if !found {
		return results, fmt.Errorf("no route found for %s steplib", bitriseStepLibURL)
	}

	latestVersionNumber := step.LatestVersionNumber
	latestVersion, found := step.LatestVersion()
	if !found {
		return results, fmt.Errorf("failed to find latest version for step %s", stepID)
	}

	log.Infof("Preloading step %s", stepID)
	targetExecutablePathLatest, err := preloadStepExecutable(log, stepLib, bitriseStepLibURL, goBuilder, stepID, step.LatestVersionNumber, latestVersion, false)
	if err != nil {
		return results, fmt.Errorf("failed to preload step %s@%s: %w", stepID, latestVersionNumber, err)
	}

	filteredSteps, err := filterPreloadedStepVersions(stepID, step.Versions, opts)
	if err != nil {
		return results, fmt.Errorf("failed to filter preloaded step versions: %w", err)
	}

	// Iterate over all versions and compress them if golang step
	for version, step := range filteredSteps {
		if version == latestVersionNumber {
			continue
		}

		log.Debugf("Preloading step %s@%s", stepID, version)
		targetExecutablePath, err := preloadStepExecutable(log, stepLib, bitriseStepLibURL, goBuilder, stepID, version, step, true)
		if err != nil {
			results = append(results, preloadResult{
				stepID:  stepID,
				version: version,
				err:     fmt.Errorf("failed to preload step %s@%s: %w", stepID, version, err),
			})

			continue
		}

		if targetExecutablePath == "" || targetExecutablePathLatest == "" {
			results = append(results, preloadResult{
				stepID:  stepID,
				version: version,
				status:  "OK (no compression)",
			})

			continue
		}

		patchFilePath := stepman.GetStepCompressedExecutablePathForVersion(latestVersionNumber, route, stepID, version)
		if err := compressStep(log, patchFilePath, targetExecutablePathLatest, targetExecutablePath); err != nil {
			results = append(results, preloadResult{
				stepID:  stepID,
				version: version,
				err:     fmt.Errorf("failed to compress step %s@%s: %w", stepID, version, err),
			})

			continue
		}

		results = append(results, preloadResult{
			stepID:  stepID,
			version: version,
			status:  "OK (compressed)",
		})
	}

	return results, nil
}

func preloadStepExecutable(log stepman.Logger, stepLib models.StepCollectionModel, stepLibURI string, goBuilder GoBuilder, id, version string, step models.StepModel, cleanupSrc bool) (string, error) {
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
	log.Debugf("Downloading step %s@%s", id, version)
	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return "", fmt.Errorf("download failed: %s", err)
	}

	if step.Toolkit == nil || step.Toolkit.Go == nil {
		return "", nil
	}

	log.Debugf("Building step %s@%s", id, version)
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
