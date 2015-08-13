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
			Name:    "share",
			Aliases: []string{"s"},
			Usage:   "Publish your step.",
			Action:  share,
			Subcommands: []cli.Command{
				{
					Name:    "start",
					Aliases: []string{"s"},
					Usage:   "Preparations for publishing.",
					Action:  start,
					Flags: []cli.Flag{
						flCollection,
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
					},
				},
				{
					Name:    "finish",
					Aliases: []string{"f"},
					Usage:   "Finish up.",
					Action:  finish,
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
