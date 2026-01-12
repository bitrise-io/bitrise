package preload

import (
	"fmt"
	"os"
	"sync"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	workers = 10
)

type CacheOpts struct {
	NumMajor                uint
	NumMinor                uint
	LatestMinorsSinceMonths int
	PatchesSinceMonths      int
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

// CacheSteps preloads the step cache for offline access
func CacheSteps(log stepman.Logger, steplibURL, maintaner string, opts CacheOpts) error {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(steplibURL); err != nil {
		return err
	} else if !exist {
		if err := stepman.SetupLibrary(steplibURL, log); err != nil {
			return fmt.Errorf("failed to setup steplib: %w", err)
		}
	}

	stepLib, err := stepman.ReadStepSpec(steplibURL)
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
				results, err := preloadStepVersions(log, steplibURL, stepLib, s.stepID, s.step, opts)
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
			if maintaner != "" && step.Info.Maintainer != maintaner {
				log.Infof("Skipping step %s as maintaner is not '%s'", stepID, maintaner)
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

func preloadStepVersions(log stepman.Logger, steplibURL string, stepLib models.StepCollectionModel, stepID string, step models.StepGroupModel, opts CacheOpts) ([]preloadResult, error) {
	results := []preloadResult{}

	_, found := stepman.ReadRoute(steplibURL)
	if !found {
		return results, fmt.Errorf("no route found for %s steplib", steplibURL)
	}

	// Preload latest version
	latestVersionNumber := step.LatestVersionNumber
	latestVersion, found := step.LatestVersion()
	if !found {
		return results, fmt.Errorf("failed to find latest version for step %s", stepID)
	}

	log.Infof("Preloading step %s@latest", stepID)
	err := preloadStep(log, stepLib, steplibURL, stepID, step.LatestVersionNumber, latestVersion)
	if err != nil {
		return results, fmt.Errorf("failed to preload step %s@%s: %w", stepID, latestVersionNumber, err)
	}

	results = append(results, preloadResult{
		stepID:  stepID,
		version: latestVersionNumber,
		status:  "OK",
		err:     nil,
	})

	filteredSteps, err := filterPreloadedStepVersions(stepID, step.Versions, opts)
	if err != nil {
		return results, fmt.Errorf("failed to filter preloaded step versions: %w", err)
	}

	// Iterate over all versions other then latest
	for version, step := range filteredSteps {
		if version == latestVersionNumber {
			continue
		}

		log.Debugf("Preloading step %s@%s", stepID, version)
		err := preloadStep(log, stepLib, steplibURL, stepID, version, step)
		if err != nil {
			results = append(results, preloadResult{
				stepID:  stepID,
				version: version,
				err:     fmt.Errorf("failed to preload step %s@%s: %w", stepID, version, err),
				status:  "FAILED",
			})

			continue
		}

		results = append(results, preloadResult{
			stepID:  stepID,
			version: version,
			status:  "OK (source)",
			err:     nil,
		})

		continue
	}

	return results, nil
}

func cleanStepSourceDir(route stepman.SteplibRoute, id, version string) (string, error) {
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

	return stepSourceDir, nil
}

func preloadStep(log stepman.Logger, stepLib models.StepCollectionModel, stepLibURI string, id, version string, step models.StepModel) error {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	_, err := cleanStepSourceDir(route, id, version)
	if err != nil {
		return err
	}

	log.Debugf("Downloading step %s@%s", id, version)
	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return fmt.Errorf("download failed: %s", err)
	}

	return nil
}
