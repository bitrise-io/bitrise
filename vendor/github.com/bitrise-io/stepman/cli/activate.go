package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

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
			log.Fatalf("Command failed: %s", err)
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

	return Activate(stepLibURI, id, version, path, copyYML, update)
}

// Activate ...
func Activate(stepLibURI, id, version, destination, destinationStepYML string, updateLibrary bool) error {
	stepLib, err := stepman.ReadStepSpec(stepLibURI)
	if err != nil {
		return fmt.Errorf("failed to read %s steplib: %s", stepLibURI, err)
	}

	step, version, err := queryStep(stepLib, stepLibURI, id, version, updateLibrary)
	if err != nil {
		return fmt.Errorf("failed to find step: %s", err)
	}

	srcFolder, err := downloadStep(stepLib, stepLibURI, id, version, step)
	if err != nil {
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

func queryStep(stepLib models.StepCollectionModel, stepLibURI string, id, version string, updateLibrary bool) (models.StepModel, string, error) {
	step, stepFound, versionFound := stepLib.GetStep(id, version)
	if (!stepFound || !versionFound) && updateLibrary {
		log.Infof("StepLib doesn't contain step (%s) with version: %s -- Updating StepLib", id, version)

		var err error
		stepLib, err = stepman.UpdateLibrary(stepLibURI)
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

func downloadStep(stepLib models.StepCollectionModel, stepLibURI, id, version string, step models.StepModel) (string, error) {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return "", fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	if exist, err := pathutil.IsPathExists(stepCacheDir); err != nil {
		return "", fmt.Errorf("failed to check if %s path exist: %s", stepCacheDir, err)
	} else if !exist {
		if err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit); err != nil {
			return "", fmt.Errorf("download failed: %s", err)
		}
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
