package steplib

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/bitrise-io/stepman/stepman"
	"gopkg.in/yaml.v2"
)

const precompiledStepsEnv = "BITRISE_EXPERIMENT_PRECOMPILED_STEPS"
const precompiledStepsStorageURLsEnv = "BITRISE_PRECOMPILED_STEPS_STORAGE_URLS"

var precompiledStepsDefaultStorageURLs = []string{
	"https://steplib.bitrise.io",
	"https://storage.googleapis.com/bitrise-steplib-storage",
}

type ResolvedStep struct {
	// ExecPath is optional: it holds the activated step executable only for
	// precompiled executable activation, and is empty for source activation.
	ExecPath string
	StepInfo models.StepInfoModel
}

func ActivateStep(id stepid.CanonicalID, destination, destinationStepYML string, log stepman.Logger, isOfflineMode bool, libraryAPI *steplibrary.Client) (ResolvedStep, error) {
	var stepInfo models.StepInfoModel
	var resolveErr error
	if libraryAPI != nil {
		stepInfo, resolveErr = libraryAPI.FetchStepMetadata(context.Background(), id)
	} else {
		// Legacy path: resolve the step from the local steplib spec (resolving the
		// version constraint to a concrete version). This repeats the resolution
		// already done by prepareStepLibForActivation, but keeps the legacy path
		// self-contained instead of threading resolved info in.
		stepInfo, resolveErr = stepman.QueryStepInfoFromLibrary(id.SteplibSource, id.IDorURI, id.Version, log)
	}
	if resolveErr != nil {
		return ResolvedStep{ExecPath: "", StepInfo: models.StepInfoModel{}}, resolveErr
	}
	stepModel := stepInfo.Step
	version := stepInfo.Version

	// Place the step.yml at destinationStepYML once, up front.
	if libraryAPI == nil {
		if err := copyStepYML(id.SteplibSource, id.IDorURI, version, destinationStepYML); err != nil {
			return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, fmt.Errorf("copy step.yml: %s", err)
		}
	} else {
		if err := writeStepYML(stepInfo.Step, destinationStepYML); err != nil {
			return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, err
		}
	}

	execPath, err := downloadPrecompiled(log, stepModel, id, destination)
	if execPath != "" {
		return ResolvedStep{ExecPath: execPath, StepInfo: stepInfo}, err
	}

	// Fall back to step source activation.
	if libraryAPI != nil {
		// activate the source over the API, without git clone
		if err := activateStepSourceWithAPI(libraryAPI, id.SteplibSource, version, stepModel.Source, destination, log, isOfflineMode); err != nil {
			return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, err
		}
		return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, nil
	}

	// using git cloned steplib
	stepCollection, err := stepman.ReadStepSpec(id.SteplibSource)
	if err != nil {
		return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, fmt.Errorf("failed to read %s steplib: %s", id.SteplibSource, err)
	}
	if err := activateStepSource(stepCollection, id.SteplibSource, id.IDorURI, version, stepModel, destination, log, isOfflineMode); err != nil {
		return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, err
	}

	return ResolvedStep{ExecPath: "", StepInfo: stepInfo}, nil
}

func downloadPrecompiled(log stepman.Logger, step models.StepModel, id stepid.CanonicalID, destination string) (string, error) {
	if (os.Getenv(precompiledStepsEnv) == "true" || os.Getenv(precompiledStepsEnv) == "1") && step.Executables != nil {
		platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		executableForPlatform, ok := (*step.Executables)[platform]
		if ok && executableForPlatform.Hash != "" && executableForPlatform.StorageURI != "" {
			log.Debugf("Downloading executable for %s", platform)
			downloadStart := time.Now()
			execPath, err := activateStepExecutable(id.IDorURI, executableForPlatform, destination, log)
			if err == nil {
				log.Debugf("Downloaded executable in %s", time.Since(downloadStart).Round(time.Millisecond))

				return execPath, nil
			}
			log.Warnf("Failed to download step executable, fallback to step source activation: %s", err)
		}
		log.Infof("No prebuilt executable found for %s, fallback to step source activation", platform)
	}
	return "", nil
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

func writeStepYML(step models.StepModel, outputPath string) error {
	stepYML, err := yaml.Marshal(step)
	if err != nil {
		return fmt.Errorf("marshal step model to YAML: %w", err)
	}

	fileManager := fileutil.NewFileManager()
	if err := fileManager.WriteBytes(outputPath, stepYML); err != nil {
		return fmt.Errorf("write step.yml: %w", err)
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
