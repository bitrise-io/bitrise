package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

var stepInfoCommand = cli.Command{
	Name:  "step-info",
	Usage: "Prints the step definition (step.yml content).",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "library",
			Usage:  "Library of the step (options: LIBRARY_URI, git, path).",
			EnvVar: "STEPMAN_LIBRARY_URI",
		},
		cli.StringFlag{
			Name:  "id",
			Usage: "ID of the step (options: ID_IN_LIBRARY, GIT_URI, LOCAL_STEP_DIRECTORY_PATH).",
		},
		cli.StringFlag{
			Name:  "version",
			Usage: "Version of the step (options: VERSION_IN_LIBRARY, GIT_BRANCH_OR_TAG).",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "Output format (options: raw, json).",
		},
		cli.StringFlag{
			Name:   "collection, c",
			Usage:  "[DEPRECATED] Collection of step.",
			EnvVar: CollectionPathEnvKey,
		},
		cli.BoolFlag{
			Name:  "short",
			Usage: "[DEPRECATED] Show short version of infos.",
		},
		cli.StringFlag{
			Name:  "step-yml",
			Usage: "[DEPRECATED] Path of step.yml",
		},
	},
	Action: func(c *cli.Context) error {
		if err := stepInfo(c); err != nil {
			log.Errorf("Command failed: %s", err)
			os.Exit(1)
		}
		return nil
	},
}

func stepInfo(c *cli.Context) error {
	// Input parsing
	library := c.String("library")
	if library == "" {
		collection := c.String(CollectionKey)
		library = collection
	}

	id := c.String(IDKey)
	if id == "" {
		stepYMLPath := c.String(StepYMLKey)
		if stepYMLPath != "" {
			id = stepYMLPath
			library = "path"
		}
	}

	if library == "" {
		return fmt.Errorf("step info: missing required input: library")
	}
	if id == "" {
		return fmt.Errorf("step info: missing required input: id")
	}

	version := c.String(VersionKey)

	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	}
	if format != OutputFormatRaw && format != OutputFormatJSON {
		return fmt.Errorf("step info: invalid format value: %s, valid values: [%s, %s]", format, OutputFormatRaw, OutputFormatJSON)
	}

	var logger log.Logger
	logger = log.NewDefaultRawLogger()
	if format == OutputFormatJSON {
		logger = log.NewDefaultJSONLoger()
	}

	stepInfo, err := QueryStepInfo(library, id, version, log.NewDefaultLogger(false))
	if err != nil {
		return err
	}

	logger.Print(stepInfo)
	return nil
}

// QueryStepInfo returns a matching step info.
// In cases of git and path sources the step.yml is read, otherwise the step is looked up in a step library.
func QueryStepInfo(library, id, version string, log stepman.Logger) (models.StepInfoModel, error) {
	switch library {
	case "git":
		return QueryStepInfoFromGit(id, version)
	case "path":
		return QueryStepInfoFromPath(id)
	default: // library step
		return QueryStepInfoFromLibrary(library, id, version, log)
	}
}

// QueryStepInfoFromGit returns step info from git source.
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

	step, err := stepman.ParseStepDefinition(stepDefinitionPth, false)
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

	step, err := stepman.ParseStepDefinition(stepDefinitionPth, false)
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
func QueryStepInfoFromLibrary(library, id, version string, log stepman.Logger) (models.StepInfoModel, error) {
	// Check if setup was done for collection
	if exist, err := stepman.RootExistForLibrary(library); err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: check if setup was done for %s: %s", library, err)
	} else if !exist {
		if err := stepman.SetupLibrary(library, log); err != nil {
			return models.StepInfoModel{}, fmt.Errorf("query steplib step info: setup %s: %s", library, err)
		}
	}

	stepVersion, err := stepman.ReadStepVersionInfo(library, id, version)
	if err != nil {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: read step information: %s", err)
	}

	route, found := stepman.ReadRoute(library)
	if !found {
		return models.StepInfoModel{}, fmt.Errorf("query steplib step info: no route found for %s", library)
	}

	stepDir := stepman.GetStepCollectionDirPath(route, id, stepVersion.Version)
	stepDefinitionPth := filepath.Join(stepDir, "step.yml")

	infoPath := stepman.GetStepGlobalInfoPath(route, id)

	groupInfo, _, err := stepman.ParseStepGroupInfoModel(infoPath)
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
