package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

var errStepNotAvailableOfflineMode error = fmt.Errorf("step not available in offline mode")

var activateCommand = cli.Command{
	Name:  "activate",
	Usage: "Copy the step with specified --id, and --version, into provided path. If --version flag is not set, the latest version of the step will be used. If --copyyml flag is set, step.yml will be copied to the given path.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   CollectionKey + ", " + collectionKeyShort,
			Usage:  "Collection of step.",
			EnvVar: CollectionPathEnvKey,
		},
		cli.StringFlag{
			Name:  IDKey + ", " + idKeyShort,
			Usage: "Step id.",
		},
		cli.StringFlag{
			Name:  VersionKey + ", " + versionKeyShort,
			Usage: "Step version.",
		},
		cli.StringFlag{
			Name:  PathKey + ", " + pathKeyShort,
			Usage: "Path where the step will copied.",
		},
		cli.StringFlag{
			Name:  CopyYMLKey + ", " + copyYMLKeyShort,
			Usage: "Path where the activated step's step.yml will be copied.",
		},
		cli.BoolFlag{
			Name:  UpdateKey + ", " + updateKeyShort,
			Usage: "If flag is set, and collection doesn't contains the specified step, the collection will updated.",
		},
	},
	Action: func(c *cli.Context) error {
		if err := activate(c); err != nil {
			failf("Command failed: %s", err)
		}
		return nil
	},
}

func activate(c *cli.Context) error {
	stepLibURI := c.String(CollectionKey)
	if stepLibURI == "" {
		return fmt.Errorf("no steplib specified")
	}

	id := c.String(IDKey)
	if id == "" {
		return fmt.Errorf("no step ID specified")
	}

	path := c.String(PathKey)
	if path == "" {
		return fmt.Errorf("no destination path specified")
	}

	version := c.String(VersionKey)
	copyYML := c.String(CopyYMLKey)
	update := c.Bool(UpdateKey)
	logger := log.NewDefaultLogger(false)
	isOfflineMode := false

	output, err := Activate(stepLibURI, id, version, path, copyYML, update, logger, isOfflineMode)
	if err != nil {
		return err
	}

	logger.Printf("%v+", output)

	return nil
}

// Activate ...
func Activate(stepLibURI, id, version, destination, destinationStepYML string, updateLibrary bool, log stepman.Logger, isOfflineMode bool) (models.ActivatedStep, error) {
	output := models.ActivatedStep{}

	stepLib, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return output, fmt.Errorf("failed to read %s steplib: %s", stepLibURI, err)
	}

	step, version, err := queryStep(stepLib, stepLibURI, id, version, updateLibrary, log)
	if err != nil {
		return output, fmt.Errorf("failed to find step: %s", err)
	}

	activatedStep, err := activateStep(stepLib, stepLibURI, id, version, step, log, isOfflineMode)
	if err != nil {
		if err == errStepNotAvailableOfflineMode {
			availableVersions := listCachedStepVersion(log, stepLib, stepLibURI, id)
			versionList := "Other versions available in the local cache:"
			for _, version := range availableVersions {
				versionList = versionList + fmt.Sprintf("\n- %s", version)
			}

			errMsg := fmt.Sprintf("version is not available in the local cache and $BITRISE_BETA_OFFLINE_MODE is set. %s", versionList)
			return models.ActivatedStep{}, fmt.Errorf("failed to download step: %s", errMsg)
		}

		return output, fmt.Errorf("failed to download step: %s", err)
	}

	if activatedStep.SourceAbsDirPath != "" {
		if err := copyStep(activatedStep.SourceAbsDirPath, destination); err != nil {
			return output, fmt.Errorf("copy step failed: %s", err)
		}
	}

	if destinationStepYML != "" {
		if err := copyStepYML(stepLibURI, id, version, destinationStepYML); err != nil {
			return output, fmt.Errorf("copy step.yml failed: %s", err)
		}

		activatedStep.StepYMLPath = destinationStepYML
	}

	return activatedStep, nil
}

func queryStep(stepLib models.StepCollectionModel, stepLibURI string, id, version string, updateLibrary bool, log stepman.Logger) (models.StepModel, string, error) {
	step, stepFound, versionFound := stepLib.GetStep(id, version)
	if (!stepFound || !versionFound) && updateLibrary {
		var err error
		stepLib, err = stepman.UpdateLibrary(stepLibURI, log)
		if err != nil {
			return models.StepModel{}, "", fmt.Errorf("failed to update %s steplib: %s", stepLibURI, err)
		}

		step, stepFound, versionFound = stepLib.GetStep(id, version)
	}
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

func getExecutableFromCache(executablePath, checkSumPath string) error {
	for _, path := range []string{executablePath, checkSumPath} {
		exist, err := pathutil.IsPathExists(path)
		if err != nil {
			return fmt.Errorf("failed to check if path (%s) exist: %w", path, err)
		}

		if !exist {
			return fmt.Errorf("executable not found in cache, path (%s) missing", path)
		}
	}

	return checkChecksum(executablePath, checkSumPath)
}

func activateStep(stepLib models.StepCollectionModel, stepLibURI, id, version string, step models.StepModel, log stepman.Logger, isOfflineMode bool) (models.ActivatedStep, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return models.ActivatedStep{}, fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	stepBinDir := stepman.GetStepBinDirPathForVersion(route, id, version)
	stepBinDirExist, err := pathutil.IsPathExists(stepBinDir)
	if err != nil {
		return models.ActivatedStep{}, fmt.Errorf("failed to check if %s path exist: %s", stepBinDir, err)
	}
	if stepBinDirExist {
		// is precompiled uncompressed step version in cache?
		executablePath := stepman.GetStepExecutablePathForVersion(route, id, version)
		checkSumPath := stepman.GetStepExecutableChecksumPathForVersion(route, id, version)

		err := getExecutableFromCache(executablePath, checkSumPath)
		if err == nil {
			return models.NewActivatedStepFromExecutable(executablePath), nil
		}
		log.Warnf("[Stepman] %s", err)

		// is precompiled binary patch in cache?
		fromPatchVersion := stepLib.Steps[id].LatestVersionNumber
		fromPatchExecutablePath := stepman.GetStepExecutablePathForVersion(route, id, fromPatchVersion)
		binaryPatchPath := stepman.GetStepCompressedExecutablePathForVersion(fromPatchVersion, route, id, version)

		err = uncompressStepFromCache(fromPatchExecutablePath, binaryPatchPath, executablePath, checkSumPath)
		if err == nil {
			return models.NewActivatedStepFromExecutable(executablePath), nil
		}
		log.Warnf("[Stepman] %s", err)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		return models.ActivatedStep{}, fmt.Errorf("failed to check if %s path exist: %s", stepCacheDir, err)
	} else if exist { // version specific source cache exists
		return models.NewActivatedStepFromSourceDir(stepCacheDir), nil
	}

	// version specific source cache not exists
	if isOfflineMode {
		return models.ActivatedStep{}, errStepNotAvailableOfflineMode
	}

	if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log); err != nil {
		return models.ActivatedStep{}, fmt.Errorf("download failed: %s", err)
	}

	return models.NewActivatedStepFromSourceDir(stepCacheDir), nil
}

func listCachedStepVersion(log stepman.Logger, stepLib models.StepCollectionModel, stepLibURI, stepID string) []string {
	versions := []string{}
	for version, step := range stepLib.Steps[stepID].Versions {
		_, err := activateStep(stepLib, stepLibURI, stepID, version, step, log, true)
		if err == nil {
			versions = append(versions, version)
		}
	}

	return versions
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
