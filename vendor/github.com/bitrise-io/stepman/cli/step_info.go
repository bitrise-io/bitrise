package cli

import (
	"fmt"
	"log"
	"time"

	"path/filepath"

	"github.com/bitrise-io/go-utils/command/git"
	flog "github.com/bitrise-io/go-utils/log"
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
			log.Fatalf("Command failed, error: %s", err)
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
		return fmt.Errorf("Missing required input: library")
	}
	if id == "" {
		return fmt.Errorf("Missing required input: id")
	}

	version := c.String(VersionKey)

	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	}
	if format != OutputFormatRaw && format != OutputFormatJSON {
		return fmt.Errorf("Invalid input value: format = %s", format)
	}

	var log flog.Logger
	log = flog.NewDefaultRawLogger()
	if format == OutputFormatJSON {
		log = flog.NewDefaultJSONLoger()
	}
	// ---

	stepInfo := models.StepInfoModel{
		Library: library,
		ID:      id,
		Version: version,
	}

	switch library {
	case "git":
		stepGitSourceURI := id
		tmpStepDir, err := pathutil.NormalizedOSTempDirPath("__step__")
		if err != nil {
			return fmt.Errorf("failed to create tmp dir, error: %s", err)
		}

		tagOrBranch := version
		if tagOrBranch == "" {
			tagOrBranch = "master"
		}

		if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
			repo, err := git.New(tmpStepDir)
			if err != nil {
				return err
			}
			return repo.CloneTagOrBranch(stepGitSourceURI, tagOrBranch).Run()
		}); err != nil {
			return fmt.Errorf("failed to clone step from: %s, error: %s", stepGitSourceURI, err)
		}

		stepDefinitionPth := filepath.Join(tmpStepDir, "step.yml")
		if exist, err := pathutil.IsPathExists(stepDefinitionPth); err != nil {
			return fmt.Errorf("failed to check if step definition (step.yml) exist at: %s, error: %s", stepDefinitionPth, err)
		} else if !exist {
			return fmt.Errorf("step definition (step.yml) does not exist at: %s", stepDefinitionPth)
		}

		step, err := stepman.ParseStepDefinition(stepDefinitionPth, false)
		if err != nil {
			return fmt.Errorf("failed to parse step definition at: %s, error: %s", stepDefinitionPth, err)
		}

		stepInfo.Version = tagOrBranch
		stepInfo.Step = step
		stepInfo.DefinitionPth = stepDefinitionPth
	case "path":
		stepDir := id
		stepDefinitionPth := filepath.Join(stepDir, "step.yml")
		if exist, err := pathutil.IsPathExists(stepDefinitionPth); err != nil {
			return fmt.Errorf("failed to check if step definition (step.yml) exist at: %s, error: %s", stepDefinitionPth, err)
		} else if !exist {
			return fmt.Errorf("step definition (step.yml) does not exist at: %s", stepDefinitionPth)
		}

		step, err := stepman.ParseStepDefinition(stepDefinitionPth, false)
		if err != nil {
			return fmt.Errorf("failed to parse step definition at: %s, error: %s", stepDefinitionPth, err)
		}

		stepInfo.Step = step
		stepInfo.DefinitionPth = stepDefinitionPth
	default: // library step
		// Check if setup was done for collection
		if exist, err := stepman.RootExistForLibrary(library); err != nil {
			return fmt.Errorf("Failed to check if setup was done for steplib (%s), error: %s", library, err)
		} else if !exist {
			if err := stepman.SetupLibrary(library); err != nil {
				return fmt.Errorf("Failed to setup steplib (%s), error: %s", library, err)
			}
		}

		stepVersion, err := stepman.ReadStepVersionInfo(library, id, version)
		if err != nil {
			return fmt.Errorf("Failed to read Step information, error: %s", err)
		}

		route, found := stepman.ReadRoute(library)
		if !found {
			return fmt.Errorf("No route found for library: %s", library)
		}

		stepDir := stepman.GetStepCollectionDirPath(route, id, stepVersion.Version)
		stepDefinitionPth := filepath.Join(stepDir, "step.yml")

		stepInfo.Step = stepVersion.Step
		stepInfo.Version = stepVersion.Version
		stepInfo.LatestVersion = stepVersion.LatestAvailableVersion
		stepInfo.DefinitionPth = stepDefinitionPth
	}

	log.Print(stepInfo)
	return nil
}
