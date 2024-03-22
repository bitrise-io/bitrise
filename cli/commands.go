package cli

import "github.com/urfave/cli"

// Flags ...
const (
	JSONParamsKey       = "json-params"
	JSONParamsBase64Key = "json-params-base64"

	WorkflowKey = "workflow"

	PatternKey        = "pattern"
	PushBranchKey     = "push-branch"
	PRSourceBranchKey = "pr-source-branch"
	PRTargetBranchKey = "pr-target-branch"
	PRReadyStateKey   = "pr-ready-state"

	ConfigKey      = "config"
	InventoryKey   = "inventory"
	OuputFormatKey = "format"
)

var (
	commands = []cli.Command{
		initCmd,
		setupCommand,
		{
			Name:   "version",
			Usage:  "Prints the version",
			Action: printVersionCmd,
			Flags: []cli.Flag{
				flOutputFormat,
				cli.BoolFlag{Name: "full", Usage: "Prints the build number as well."},
			},
		},
		{
			Name:   "validate",
			Usage:  "Validates a specified bitrise config.",
			Action: validate,
			Flags: []cli.Flag{
				flConfig,
				flConfigBase64,
				flInventory,
				flInventoryBase64,
				flFormat,
			},
		},
		updateCommand,
		runCommand,
		{
			Name:   "export",
			Usage:  "Export the bitrise configuration.",
			Action: export,
			Flags: []cli.Flag{
				flConfig,
				flConfigBase64,
				flFormat,
				flOutputPath,
				flPretty,
			},
		},
		{
			Name:   "normalize",
			Usage:  "Normalize the bitrise configuration.",
			Action: normalize,
			Flags: []cli.Flag{
				flConfig,
				flConfigBase64,
			},
		},
		{
			Name:   "step-list",
			Usage:  "List of available steps.",
			Action: stepList,
			Flags: []cli.Flag{
				flCollection,
				flFormat,
			},
		},
		{
			Name:    "step-info",
			Aliases: []string{"i"},
			Usage:   "Provides information (step ID, last version, given version) about specified step.",
			Action:  stepInfo,
			Flags: []cli.Flag{
				flCollection,
				flVersion,
				flFormat,
				flShort,
				flStepYML,
			},
		},
		workflowListCommand,
		{
			Name:   "share",
			Usage:  "Publish your step.",
			Action: share,
			Subcommands: []cli.Command{
				{
					Name:   "start",
					Usage:  "Preparations for publishing.",
					Action: start,
					Flags: []cli.Flag{
						flCollection,
					},
				},
				{
					Name:   "create",
					Usage:  "Create your change - add it to your own copy of the collection.",
					Action: create,
					Flags: []cli.Flag{
						flTag,
						flGit,
						flStepID,
					},
				},
				{
					Name:   "audit",
					Usage:  "Validates the step collection.",
					Action: shareAudit,
				},
				{
					Name:   "finish",
					Usage:  "Finish up.",
					Action: finish,
				},
			},
		},
		pluginCommand,
		stepmanCommand,
		envmanCommand,
	}
)
