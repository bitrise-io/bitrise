package steplib

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

var errStepNotAvailableOfflineMode error = fmt.Errorf("step not available in offline mode")

func ActivateStep(stepLibURI, id, version, destination, destinationStepYML string, log stepman.Logger, isOfflineMode bool) error {
	stepCollection, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return fmt.Errorf("failed to read %s steplib: %s", stepLibURI, err)
	}

	step, version, err := queryStep(stepCollection, stepLibURI, id, version)
	if err != nil {
		return fmt.Errorf("failed to find step: %s", err)
	}

	srcFolder, err := activateStep(stepCollection, stepLibURI, id, version, step, log, isOfflineMode)
	if err != nil {
		if err == errStepNotAvailableOfflineMode {
			availableVersions := ListCachedStepVersions(log, stepCollection, stepLibURI, id)
			versionList := "Other versions available in the local cache:"
			for _, version := range availableVersions {
				versionList = versionList + fmt.Sprintf("\n- %s", version)
			}

			errMsg := fmt.Sprintf("version is not available in the local cache and $BITRISE_OFFLINE_MODE is set. %s", versionList)
			return fmt.Errorf("failed to download step: %s", errMsg)
		}

		return fmt.Errorf("failed to download step: %s", err)
	}

	if err := copyStep(srcFolder, destination); err != nil {
		return fmt.Errorf("copy step failed: %s", err)
	}

	if destinationStepYML != "" {
		if err := copyStepYML(stepLibURI, id, version, destinationStepYML); err != nil {
			return fmt.Errorf("copy step.yml failed: %s", err)
		}
	}

	return nil
}

func queryStep(stepLib models.StepCollectionModel, stepLibURI string, id, version string) (models.StepModel, string, error) {
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

func activateStep(stepLib models.StepCollectionModel, stepLibURI, id, version string, step models.StepModel, log stepman.Logger, isOfflineMode bool) (string, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", stepCacheDir, err)
	} else if exist {
		return stepCacheDir, nil
	}

	// version specific source cache not exists
	if isOfflineMode {
		return "", errStepNotAvailableOfflineMode
	}

	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return "", fmt.Errorf("download failed: %s", err)
	}

	return stepCacheDir, nil
}

func copyStep(src, dst string) error {
	if exist, err := pathutil.IsPathExists(dst); err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", dst, err)
	} else if !exist {
		if err := os.MkdirAll(dst, 0777); err != nil {
			return fmt.Errorf("failed to create dir for %s path: %s", dst, err)
		}
	}

	if err := command.CopyDir(src+"/", dst, true); err != nil {
		return fmt.Errorf("copy command failed: %s", err)
	}
	return nil
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

	for version, step := range stepLib.Steps[stepID].Versions {
		_, err := activateStep(stepLib, stepLibURI, stepID, version, step, log, true)
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
