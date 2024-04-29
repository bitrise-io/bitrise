package cli

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
)

type GoBuilder func(stepSourceAbsPath, packageName, targetExecutablePath string) error

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

	route, found := stepman.ReadRoute(bitriseStepLibURL)
	if !found {
		return fmt.Errorf("no route found for %s steplib", bitriseStepLibURL)
	}

	waitGroup := &sync.WaitGroup{}

	for stepID, step := range stepLib.Steps {
		if step.Info.Maintainer != bitriseMaintainer {
			log.Warnf("Skipping step %s as it is not maintained by Bitrise", stepID)
			continue
		}
		if step.Info.DeprecateNotes != "" {
			log.Warnf("Skipping deprecated step %s", stepID)
			continue
		}

		waitGroup.Add(1)
		go func(stepID string, step models.StepGroupModel) {
			latestVersionNumber := step.LatestVersionNumber
			latestVersion, found := step.LatestVersion()
			if !found {
				log.Warnf("Failed to find latest version for step %s", stepID)
			}

			log.Warnf("Preloading step %s@%s", stepID, latestVersionNumber)
			targetExecutablePathLatest, err := preloadStepExecutable(stepLib, bitriseStepLibURL, goBuilder, stepID, step.LatestVersionNumber, latestVersion, log)
			if err != nil {
				log.Warnf("Failed to download step %s@%s: %w", stepID, latestVersionNumber, err)
			}

			filteredSteps, err := filterPreloadedStepVersions(stepID, step.Versions)
			if err != nil {
				log.Warnf("Failed to filter preloaded step versions: %w", err)
			}

			// Iterate over all versions and compress them if golang step
			for version, step := range filteredSteps {
				if version == latestVersionNumber {
					continue
				}

				log.Warnf("Preloading step %s@%s", stepID, version)
				targetExecutablePath, err := preloadStepExecutable(stepLib, bitriseStepLibURL, goBuilder, stepID, version, step, log)
				if err != nil {
					log.Warnf("Failed to preload step %s@%s: %w", stepID, version, err)
				}

				if targetExecutablePath != "" && targetExecutablePathLatest != "" {
					log.Warnf("Compressing step %s@%s", stepID, version)

					patchFilePath := stepman.GetStepCompressedExecutablePathForVersion(latestVersionNumber, route, stepID, version)
					if err := compressStep(patchFilePath, targetExecutablePathLatest, targetExecutablePath); err != nil {
						log.Warnf("Failed to compress step  %s@%s: %w", stepID, version, err)
					}
				}
			}

			waitGroup.Done()
		}(stepID, step)
	}

	waitGroup.Wait()

	return nil
}

func writeChecksum(patchFilePath, checksumPath string) error {
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(patchFilePath)))
	if err := os.WriteFile(checksumPath, []byte(checksum), 0400); err != nil {
		return fmt.Errorf("Failed to write checksum (%s) to file %s: %w", checksum, checksumPath, err)
	}

	return nil
}

func checkChecksum(executablePath, checksumPath string) error {
	checksum, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("Failed to read checksum from file %s: %w", checksumPath, err)
	}

	calculatedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(executablePath)))
	if string(checksum) != calculatedChecksum {
		return fmt.Errorf("Checksum mismatch %s expected %s, got %s", executablePath, checksum, calculatedChecksum)
	}

	return nil
}

func filterPreloadedStepVersions(stepID string, steps map[string]models.StepModel) (map[string]models.StepModel, error) {
	filteredSteps := map[string]models.StepModel{}
	allVersions := map[uint64]map[uint64]models.Semver{}

	// keep no version, as it is only used internally and takes up the most space
	if stepID == "project-scanner" {
		return filteredSteps, nil
	}

	for stepVersion, step := range steps {
		// Releases in the last year
		if time.Since(*step.PublishedAt) < 24*time.Hour*365 {
			filteredSteps[stepVersion] = step
		}

		// All minor versions
		/*
			version, err := models.ParseSemver(stepVersion)
			if err != nil {
				return filteredSteps, fmt.Errorf("failed to parse version %s: %w", stepVersion, err)
			}

			if _, found := allVersions[version.Major]; !found {
				allVersions[version.Major] = map[uint64]models.Semver{}
				allVersions[version.Major][version.Minor] = version

				continue
			}

			curVersion, found := allVersions[version.Major][version.Minor]
			if !found {
				allVersions[version.Major][version.Minor] = version

				continue
			} else if version.Patch > curVersion.Patch {
				allVersions[version.Major][version.Minor] = version
			}*/
	}

	for _, minor := range allVersions {
		for _, version := range minor {
			filteredSteps[version.String()] = steps[version.String()]
		}
	}

	return filteredSteps, nil
}

func compressStep(patchFilePath, targetExecutablePathLatest, targetExecutablePath string) error {
	if targetExecutablePath == "" || targetExecutablePathLatest == "" {
		return nil
	}

	compressCmd := command.New("zstd", "-f", "--patch-from="+targetExecutablePathLatest, targetExecutablePath, "-o", patchFilePath)
	log.Warnf("$ %s", compressCmd.PrintableCommandArgs())
	out, err := compressCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compress with command (%s), output: %s", compressCmd.PrintableCommandArgs(), out)
	}

	if err := os.Remove(targetExecutablePath); err != nil {
		return fmt.Errorf("failed to remove uncompressed step executable: %s", err)
	}

	return nil
}

func uncompressStep(patchFromPath, targetVersionPatchPath, targetExecutablePath, checkSumPath string) error {
	decompressCmd := command.New("zstd", "-d", "--patch-from", patchFromPath, targetVersionPatchPath, "-o", targetExecutablePath)
	decompressCmd.SetStdout(nil).SetStderr(nil)

	exit, err := decompressCmd.RunAndReturnExitCode()
	if err != nil {
		return fmt.Errorf("failed to apply patch with command (%s), exit code: %d: %s", decompressCmd.PrintableCommandArgs(), exit, err)
	}

	checksumExist, err := pathutil.IsPathExists(checkSumPath)
	if err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", checkSumPath, err)
	}
	if !checksumExist {
		return fmt.Errorf("checksum file not found for %s", targetExecutablePath)
	}

	return checkChecksum(targetExecutablePath, checkSumPath)
}

func preloadStepExecutable(stepLib models.StepCollectionModel, stepLibURI string, goBuilder GoBuilder, id, version string, step models.StepModel, log stepman.Logger) (string, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	// is precompiled uncompressed step version in cache?
	targetExecutablePath := stepman.GetStepCacheExecutablePathForVersion(route, id, version)
	exists, err := pathutil.IsPathExists(targetExecutablePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", targetExecutablePath, err)
	}
	if exists {
		if err := os.Remove(targetExecutablePath); err != nil {
			return "", fmt.Errorf("failed to remove %s: %s", targetExecutablePath, err)
		}
	}

	// Fetch source, compile step (if golang), calclulate checksum
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

	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return "", fmt.Errorf("download failed: %s", err)
	}

	if step.Toolkit == nil || step.Toolkit.Go == nil {
		return "", nil
	}

	if err := goBuilder(stepSourceDir, step.Toolkit.Go.PackageName, targetExecutablePath); err != nil {
		return "", fmt.Errorf("failed to build step: %s", err)
	}

	checkSumPath := stepman.GetStepCacheExecutableChecksumPathForVersion(route, id, version)
	checksumExist, err := pathutil.IsPathExists(checkSumPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", checkSumPath, err)
	}
	if checksumExist {
		if err := os.Remove(checkSumPath); err != nil {
			return "", fmt.Errorf("failed to remove checksum file: %s", err)
		}
	}
	if err := writeChecksum(targetExecutablePath, checkSumPath); err != nil {
		return "", fmt.Errorf("failed to write checksum: %s", err)
	}

	// remove stepSourceDir as build is successful
	// also remove if not successful, as propably old step source does not work anymore
	if err := os.RemoveAll(stepSourceDir); err != nil {
		return "", fmt.Errorf("failed to remove step source dir: %s", err)
	}

	return targetExecutablePath, nil
}
