package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/urfave/cli"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
)

var stepsCommand = cli.Command{
	Name:  "steps",
	Usage: "Step related sub commands to list, preload and manage steps and the step library.",
	Subcommands: []cli.Command{
		{
			Name:  "list-cached",
			Usage: "List all the cached steps",
			Action: func(c *cli.Context) error {
				logCommandParameters(c)

				return listCachedSteps(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "steplib-url",
					Usage: "URL of the steplib to list or preload steps from",
					Value: bitriseStepLibURL,
				},
				cli.StringFlag{
					Name:  "maintainer",
					Usage: "Maintainer of the steps to list or preload",
					Value: bitriseMaintainer,
				},
			},
		},
		{
			Name:      "preload",
			Usage:     "Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps.",
			UsageText: fmt.Sprintf("Use the %s env var to test after preloading steps.", configs.IsSteplibOfflineModeEnvKey),
			Action: func(c *cli.Context) error {
				logCommandParameters(c)

				if err := preloadSteps(c); err != nil {
					log.Errorf("Preload failed: %s", err)
					os.Exit(1)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "steplib-url",
					Usage: "URL of the steplib to list or preload steps from",
					Value: bitriseStepLibURL,
				},
				cli.StringFlag{
					Name:  "maintainer",
					Usage: "Maintainer of the steps to list or preload",
					Value: bitriseMaintainer,
				},
				cli.UintFlag{
					Name:  "majors",
					Usage: "Include X latest major versions",
					Value: 2,
				},
				cli.UintFlag{
					Name:  "minors",
					Usage: "Include X latest minor versions for each major version",
					Value: 1,
				},
				cli.UintFlag{
					Name:  "minors-since",
					Usage: "Include latest patch version of minors that were released in the last X months",
					Value: 2,
				},
				cli.UintFlag{
					Name:  "patches-since",
					Usage: "Include all patch version that were released in the last X months",
					Value: 1,
				},
			},
		},
		{
			Name:      "generate-steplib",
			Usage:     "Generate the V2 step library inventory tree from a steplib source.",
			UsageText: "bitrise steps generate-steplib --output DIR [--steplib-url URL] [--commit-sha SHA]",
			Action: func(c *cli.Context) error {
				logCommandParameters(c)

				if err := generateSteplib(c); err != nil {
					log.Errorf("Generate steplib failed: %s", err)
					os.Exit(1)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "steplib-url",
					Usage: "URL of the steplib to generate the V2 inventory from. " +
						"Optional — defaults to the official Bitrise steplib. " +
						"Typically a git repo URL (the default branch is cloned); " +
						"a file:// URI is treated as a local directory instead.",
					Value:    bitriseStepLibURL,
					Required: false,
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Output directory for the generated V2 inventory tree (required)",
					Required: true,
				},
				cli.StringFlag{
					Name:  "commit-sha",
					Usage: "Override the commit SHA recorded in meta.json (auto-detected by underlying otherwise; mostly useful for testing).",
					Required: false,
				},
				cli.StringFlag{
					Name:  "timestamp",
					Usage: "RFC3339 timestamp to record as generated_at (for reproducible output). Defaults to current UTC time.",
					Required: false,
				},
			},
		},
	},
}
