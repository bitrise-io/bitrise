package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	stepman "github.com/bitrise-io/stepman/cli"
	"github.com/bitrise-io/stepman/preload"
	"github.com/urfave/cli"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
)

var stepsCommand = cli.Command{
	Name:  "steps",
	Usage: "Manage Steps cache.",
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
	},
}

func listCachedSteps(c *cli.Context) error {
	steplibURL := c.String("steplib-url")
	maintaner := c.String("maintainer")

	log.Infof("Listing cached steps...")
	log.Infof("Steplib: %s", steplibURL)
	if maintaner != "" {
		log.Infof("Filtering Steps by maintaner: %s", maintaner)
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	if err := stepman.ListCachedSteps(steplibURL, maintaner, logger); err != nil {
		return err
	}
	return nil
}

func preloadSteps(c *cli.Context) error {
	steplibURL := c.String("steplib-url")
	maintaner := c.String("maintainer")

	opts := preload.CacheOpts{}
	numMajor := c.Uint("majors")
	if numMajor != 0 {
		opts.NumMajor = numMajor
	}
	numMinor := c.Uint("minors")
	if numMinor != 0 {
		opts.NumMinor = numMinor
	}
	minorsSince := c.Int("minors-since")
	if minorsSince != 0 {
		opts.LatestMinorsSinceMonths = minorsSince
	}
	patchesSince := c.Int("patches-since")
	if patchesSince != 0 {
		opts.PatchesSinceMonths = patchesSince
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	log.Info("Preloading...")
	log.Info("Steplib: %s", steplibURL)
	if maintaner != "" {
		log.Infof("Filtering Steps by maintaner: %s", maintaner)
	}
	log.Printf("Options: %#v\n", opts)

	if err := preload.CacheSteps(logger, bitriseStepLibURL, bitriseMaintainer, opts); err != nil {
		return err
	}

	log.Print()
	log.Donef("Preloading completed.")

	return nil
}
