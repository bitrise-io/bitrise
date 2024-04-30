package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/toolkits"
	stepman "github.com/bitrise-io/stepman/cli"
	"github.com/urfave/cli"
)

var prelaodStepsCommand = cli.Command{
	Name:  "beta-preload-steps",
	Usage: "Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps.",
	Action: func(c *cli.Context) error {
		if err := preloadSteps(c); err != nil {
			log.Errorf("Preload failed: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "binary",
			Usage: "Compile and compress steps executables to take up less space",
		},
		cli.UintFlag{
			Name:  "majors",
			Usage: "Include X latest major versions",
		},
		cli.UintFlag{
			Name:  "minors",
			Usage: "Include X latest minor versions for each major version",
		},
		cli.UintFlag{
			Name:  "minors-since",
			Usage: "Include latest patch version of minors that were released in the last X months",
		},
		cli.UintFlag{
			Name:  "pathces-since",
			Usage: "Include all patch version that were released in the last X months",
		},
	},
}

func preloadSteps(c *cli.Context) error {
	opts := stepman.DefaultPreloadOpts()
	shouldCompile := c.Bool("binary")
	opts.UseBinaryExecutable = shouldCompile
	numMajor := c.Uint("majors")
	if numMajor != 0 {
		opts.NumMajor = numMajor
	}
	numMinor := c.Uint("minors")
	if numMinor != 0 {
		opts.NumMinor = numMinor
	}
	minorsSince := c.Duration("minors-since")
	if minorsSince != 0 {
		opts.LatestMinorsSince = minorsSince
	}
	patchesSince := c.Duration("patches-since")
	if patchesSince != 0 {
		opts.PatchesSince = patchesSince
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	log.Info("Preloading Bitrise maintained Steps...")
	log.Printf("Options: %#v", opts)

	if err := stepman.PreloadBitriseSteps(logger, toolkits.GoBuildStep, opts); err != nil {
		return err
	}

	log.Print()
	log.Donef("Preloading completed.")

	return nil
}
