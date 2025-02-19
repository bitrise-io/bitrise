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
		stepsCommand,
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
			Name:   "trigger-check",
			Usage:  "Prints out which workflow will triggered by specified pattern.",
			Action: triggerCheck,
			Flags: []cli.Flag{
				// cli params
				cli.StringFlag{Name: PatternKey, Usage: "trigger pattern."},
				cli.StringFlag{Name: ConfigKey + ", " + configShortKey, Usage: "Path where the workflow config file is located."},
				cli.StringFlag{Name: InventoryKey + ", " + inventoryShortKey, Usage: "Path of the inventory file."},

				cli.StringFlag{Name: PushBranchKey, Usage: "Git push branch name."},
				cli.StringFlag{Name: PRSourceBranchKey, Usage: "Git pull request source branch name."},
				cli.StringFlag{Name: PRTargetBranchKey, Usage: "Git pull request target branch name."},
				cli.StringFlag{Name: PRReadyStateKey, Usage: "Git pull request ready state. Options: ready_for_review draft converted_to_ready_for_review"},
				cli.StringFlag{Name: TagKey, Usage: "Git tag name."},

				cli.StringFlag{Name: OuputFormatKey, Usage: "Output format. Accepted: json, yml."},

				// cli params used in CI mode
				cli.StringFlag{Name: JSONParamsKey, Usage: "Specify command flags with json string-string hash."},
				cli.StringFlag{Name: JSONParamsBase64Key, Usage: "Specify command flags with base64 encoded json string-string hash."},

				// should deprecate
				cli.StringFlag{Name: ConfigBase64Key, Usage: "base64 encoded config data."},
				cli.StringFlag{Name: InventoryBase64Key, Usage: "base64 encoded inventory data."},
			},
		},
		triggerCommand,
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
		envmanCommand,
		mergeConfigCommand,
	}
)
