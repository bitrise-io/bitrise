package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "Initialize the specified collection, it's required before using a collection.",
			Action:  setup,
			Flags: []cli.Flag{
				flCollection,
				flLocalCollection,
				flCopySpecJSON,
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Update the collection, if no --collection flag provided, all collections will updated.",
			Action:  update,
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
			Action: list,
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
				flID,
				flVersion,
				flFormat,
				flShort,
				flStepYML,
			},
		},
		{
			Name:    "download",
			Aliases: []string{"d"},
			Usage:   "Download the step with provided --id and --version, from specified --collection, into local step downloads cache. If no --version defined, the latest version of the step (latest found in the collection) will be downloaded into the cache.",
			Action:  download,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flUpdate,
			},
		},
		{
			Name:    "activate",
			Aliases: []string{"a"},
			Usage:   "Copy the step with specified --id, and --version, into provided path. If --version flag is not set, the latest version of the step will be used. If --copyyml flag is set, step.yml will be copied to the given path.",
			Action:  activate,
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
			Name:    "share",
			Aliases: []string{"s"},
			Usage:   "Publish your step.",
			Action:  share,
			Flags: []cli.Flag{
				flToolMode,
			},
			Subcommands: []cli.Command{
				{
					Name:    "start",
					Aliases: []string{"s"},
					Usage:   "Preparations for publishing.",
					Action:  start,
					Flags: []cli.Flag{
						flCollection,
						flToolMode,
					},
				},
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create your change - add it to your own copy of the collection.",
					Action:  create,
					Flags: []cli.Flag{
						flTag,
						flGit,
						flStepID,
						flToolMode,
					},
				},
				{
					Name:    "audit",
					Aliases: []string{"a"},
					Usage:   "Validates the step collection.",
					Action:  shareAudit,
					Flags: []cli.Flag{
						flToolMode,
					},
				},
				{
					Name:    "finish",
					Aliases: []string{"f"},
					Usage:   "Finish up.",
					Action:  finish,
					Flags: []cli.Flag{
						flToolMode,
					},
				},
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete the specified collection from local caches.",
			Action:  deleteCollection,
			Flags: []cli.Flag{
				flCollection,
			},
		},
	}
)
