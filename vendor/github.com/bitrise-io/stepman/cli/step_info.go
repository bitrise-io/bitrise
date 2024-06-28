package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
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
			failf("Command failed: %s", err)
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
		return stepman.QueryStepInfoFromGit(id, version)
	case "path":
		return stepman.QueryStepInfoFromPath(id)
	default: // library step
		return stepman.QueryStepInfoFromLibrary(library, id, version, log)
	}
}
