package stepman

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/stepman/models"
)

// QueryStepInfoFromGit returns step info from a git source.
func QueryStepInfoFromGit(gitURL, tagOrBranch string) (models.StepInfoModel, error) {
	tmpStepDir, err := pathutil.NormalizedOSTempDirPath("__step__")
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query git step info: create tmp dir: %s", err)
	}

	if tagOrBranch == "" {
		tagOrBranch = "master"
	}

	if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
		repo, err := git.New(tmpStepDir)
		if err != nil {
			return err
		}
		return repo.CloneTagOrBranch(gitURL, tagOrBranch).Run()
	}); err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query git step info: clone %s: %s", gitURL, err)
	}

	stepDefinitionPth := filepath.Join(tmpStepDir, "step.yml")
	if exist, err := pathutil.IsPathExists(stepDefinitionPth); err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query git step info: check if step.yml exist: %s", err)
	} else if !exist {
		return models.StepInfoModel{}, fmt.Errorf("query git step info: step.yml does not exist at %s", stepDefinitionPth)
	}

	step, err := ParseStepDefinition(stepDefinitionPth, false)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query git step info: parse step.yml (%s): %s", stepDefinitionPth, err)
	}

	return models.StepInfoModel{
		Library:       "git",
		ID:            gitURL,
		Version:       tagOrBranch,
		Step:          step,
		DefinitionPth: stepDefinitionPth,
	}, nil
}

// QueryStepInfoFromPath returns step info from a local path source
func QueryStepInfoFromPath(dir string) (models.StepInfoModel, error) {
	stepDefinitionPth := filepath.Join(dir, "step.yml")
	if exist, err := pathutil.IsPathExists(stepDefinitionPth); err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query local step info: check if step.yml exist: %s", err)
	} else if !exist {
		return models.StepInfoModel{}, fmt.Errorf("query local step info: step.yml does not exist at %s", stepDefinitionPth)
	}

	step, err := ParseStepDefinition(stepDefinitionPth, false)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query local step info: parse step.yml (%s): %s", stepDefinitionPth, err)
	}

	return models.StepInfoModel{
		Library:       "path",
		ID:            dir,
		Version:       "",
		Step:          step,
		DefinitionPth: stepDefinitionPth,
	}, nil
}

// QueryStepInfoFromLibrary returns a step version based on the version string, which can be latest or locked to major or minor versions
func QueryStepInfoFromLibrary(library, id, version string, log Logger) (models.StepInfoModel, error) {
	// Check if setup was done for collection
	if exist, err := RootExistForLibrary(library); err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: check if setup was done for %s: %s", library, err)
	} else if !exist {
		if err := SetupLibrary(library, log); err != nil {
			return models.StepInfoModel{}, fmt.Errorf("query steplib step info: setup %s: %s", library, err)
		}
	}

	collection, err := ReadStepSpec(library)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("Failed to read steps spec (spec.json), err: %s", err)
	}

	return QueryStepInfoFromCollection(collection, library, id, version)
}

func QueryStepInfoFromCollection(collection models.StepCollectionModel, library, id, version string) (models.StepInfoModel, error) {
	stepVersion, err := ReadStepVersionInfo(collection, id, version)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: read step information: %s", err)
	}

	route, found := ReadRoute(library)
	if !found {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: no route found for %s", library)
	}

	stepDir := GetStepCollectionDirPath(route, id, stepVersion.Version)
	stepDefinitionPth := filepath.Join(stepDir, "step.yml")

	infoPath := GetStepGlobalInfoPath(route, id)

	groupInfo, _, err := ParseStepGroupInfoModel(infoPath)
	if err != nil {
		return models.StepInfoModel{}, err
	}

	return models.StepInfoModel{
		Library:       library,
		ID:            id,
		Version:       stepVersion.Version,
		LatestVersion: stepVersion.LatestAvailableVersion,
		Step:          stepVersion.Step,
		DefinitionPth: stepDefinitionPth,
		GroupInfo:     groupInfo,
	}, nil
}
