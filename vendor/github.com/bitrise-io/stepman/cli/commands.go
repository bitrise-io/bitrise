package cli

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:   "version",
			Usage:  "Prints the version",
			Action: printVersionCmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format",
					Usage: "Output format. Accepted: json, yml",
				},
				cli.BoolFlag{
					Name:  "full",
					Usage: "Prints the build number and commit as well.",
				},
			},
		},
		{
			Name:   "setup",
			Usage:  "Initialize the specified collection, it's required before using a collection.",
			Action: setup,
			Flags: []cli.Flag{
				flCollection,
				flLocalCollection,
				flCopySpecJSON,
			},
		},
		{
			Name:   "update",
			Usage:  "Update the collection, if no --collection flag provided, all collections will updated.",
			Action: update,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:   "collections",
			Usage:  "List of localy available collections.",
			Action: collections,
			Flags: []cli.Flag{
				flFormat,
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
		stepInfoCommand,
		{
			Name:   "download",
			Usage:  "Download the step with provided --id and --version, from specified --collection, into local step downloads cache. If no --version defined, the latest version of the step (latest found in the collection) will be downloaded into the cache.",
			Action: download,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flUpdate,
			},
		},
		{
			Name:   "activate",
			Usage:  "Copy the step with specified --id, and --version, into provided path. If --version flag is not set, the latest version of the step will be used. If --copyyml flag is set, step.yml will be copied to the given path.",
			Action: activate,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flPath,
				flCopyYML,
				flUpdate,
			},
		},
		{
			Name:   "audit",
			Usage:  "Validates Step or Step Collection.",
			Action: audit,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   CollectionKey + ", " + collectionKeyShort,
					Usage:  "For validating Step Collection before share.",
					EnvVar: CollectionPathEnvKey,
				},
				cli.StringFlag{
					Name:  "step-yml",
					Usage: "For validating Step before share or before share Pull Request.",
				},
				cli.BoolFlag{
					Name:  "before-pr",
					Usage: "If flag is set, Step Pull Request required fields will be checked to. Note: only for Step audit.",
				},
			},
		},
		{
			Name:   "share",
			Usage:  "Publish your step.",
			Action: share,
			Flags: []cli.Flag{
				flToolMode,
			},
			Subcommands: []cli.Command{
				{
					Name:   "start",
					Usage:  "Preparations for publishing.",
					Action: start,
					Flags: []cli.Flag{
						flCollection,
						flToolMode,
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
						flToolMode,
					},
				},
				{
					Name:   "audit",
					Usage:  "Validates the step collection.",
					Action: shareAudit,
					Flags: []cli.Flag{
						flToolMode,
					},
				},
				{
					Name:   "finish",
					Usage:  "Finish up.",
					Action: finish,
					Flags: []cli.Flag{
						flToolMode,
					},
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "Delete the specified collection from local caches.",
			Action: deleteStepLib,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:   "export-spec",
			Usage:  "Export the generated StepLib spec.",
			Action: export,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "steplib",
					Usage: "StepLib URI",
				},
				cli.StringFlag{
					Name:  "output",
					Usage: "Output path",
				},
				cli.StringFlag{
					Name:  "export-type",
					Value: "full",
					Usage: "Export type, options: [full, latest, minimal]",
				},
			},
		},
	}
)
