package steplib

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

const precompiledStepsEnv = "BITRISE_EXPERIMENT_PRECOMPILED_STEPS"

func ActivateStep(stepLibURI, id, version, destination, destinationStepYML string, log stepman.Logger, isOfflineMode bool) (string, error) {
	stepCollection, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return "", fmt.Errorf("failed to read %s steplib: %s", stepLibURI, err)
	}

	step, version, err := queryStepMetadata(stepCollection, stepLibURI, id, version)
	if err != nil {
		return "", fmt.Errorf("failed to find step: %s", err)
	}

	if (os.Getenv(precompiledStepsEnv) == "true" || os.Getenv(precompiledStepsEnv) == "1") && step.Executables != nil {
		platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		executableForPlatform, ok := (*step.Executables)[platform]
		if ok {
			log.Debugf("Downloading executable for %s", platform)
			downloadStart := time.Now()
			execPath, err := activateStepExecutable(stepLibURI, id, version, executableForPlatform, destination, destinationStepYML)
			if err == nil {
				log.Debugf("Downloaded executable in %s", time.Since(downloadStart).Round(time.Millisecond))
				return execPath, nil
			}
			log.Warnf("Failed to download step executable, fallback to step source activation: %s", err)
		}
		log.Infof("No prebuilt executable found for %s, fallback to step source activation", platform)
	}
	err = activateStepSource(stepCollection, stepLibURI, id, version, step, destination, destinationStepYML, log, isOfflineMode)
	return "", err
}

func queryStepMetadata(stepLib models.StepCollectionModel, stepLibURI string, id, version string) (models.StepModel, string, error) {
	step, stepFound, versionFound := stepLib.GetStep(id, version)

	if !stepFound {
		return models.StepModel{}, "", fmt.Errorf("%s steplib does not contain %s step", stepLibURI, id)
	}
	if !versionFound {
		return models.StepModel{}, "", fmt.Errorf("%s steplib does not contain %s step %s version", stepLibURI, id, version)
	}

	if version == "" {
		latest, err := stepLib.GetLatestStepVersion(id)
		if err != nil {
			return models.StepModel{}, "", fmt.Errorf("failed to find latest version of %s step", id)
		}
		version = latest
	}

	return step, version, nil
}

func copyStepYML(libraryURL, id, version, dest string) error {
	route, found := stepman.ReadRoute(libraryURL)
	if !found {
		return fmt.Errorf("no route found for %s steplib", libraryURL)
	}

	if exist, err := pathutil.IsPathExists(dest); err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", dest, err)
	} else if exist {
		return fmt.Errorf("%s already exist", dest)
	}

	stepCollectionDir := stepman.GetStepCollectionDirPath(route, id, version)
	stepYMLSrc := filepath.Join(stepCollectionDir, "step.yml")
	if err := command.CopyFile(stepYMLSrc, dest); err != nil {
		return fmt.Errorf("copy command failed: %s", err)
	}
	return nil
}

func ListCachedStepVersions(log stepman.Logger, stepLib models.StepCollectionModel, stepLibURI, stepID string) []string {
	versions := []models.Semver{}

	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return nil
	}

	for version := range stepLib.Steps[stepID].Versions {
		stepCacheDir := stepman.GetStepCacheDirPath(route, stepID, version)
		_, err := os.Stat(stepCacheDir)
		if err != nil {
			continue
		}

		v, err := models.ParseSemver(version)
		if err != nil {
			log.Warnf("failed to parse version (%s): %s", version, err)
		}

		versions = append(versions, v)
	}

	slices.SortFunc(versions, models.CmpSemver)

	versionsStr := make([]string, len(versions))
	for i, v := range versions {
		versionsStr[i] = v.String()
	}

	return versionsStr
}
