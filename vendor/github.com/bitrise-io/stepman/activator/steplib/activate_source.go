package steplib

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/internal/httpfetch"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
)

func activateStepSource(
	stepLib models.StepCollectionModel,
	stepLibURI, id, version string,
	step models.StepModel,
	destination string,
	log stepman.Logger,
	fetcher httpfetch.Client,
) error {
	route, found := stepman.ReadRoute(stepLibURI)
	if !found {
		return fmt.Errorf("no route found for %s steplib", stepLibURI)
	}

	stepCacheDir := stepman.GetStepCacheDirPath(route, id, version)
	stepCacheDirExists, err := pathutil.IsPathExists(stepCacheDir)
	if err != nil {
		return fmt.Errorf("failed to check if %s path exist: %s", stepCacheDir, err)
	}

	if !stepCacheDirExists {
		err := stepman.DownloadStep(stepLibURI, stepLib, id, version, step.Source.Commit, log, fetcher)
		if err != nil {
			return fmt.Errorf("download failed: %s", err)
		}
	}

	if err := copyStep(stepCacheDir, destination); err != nil {
		return fmt.Errorf("copy step failed: %s", err)
	}

	return nil
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
