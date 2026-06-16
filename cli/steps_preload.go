package cli

import (
	"github.com/bitrise-io/bitrise/v2/log"
	stepman "github.com/bitrise-io/stepman/cli"
	"github.com/bitrise-io/stepman/preload"
	"github.com/urfave/cli"
)

func listCachedSteps(c *cli.Context, logger log.Logger) error {
	steplibURL := c.String("steplib-url")
	maintaner := c.String("maintainer")

	logger.Infof("Listing cached steps...")
	logger.Infof("Steplib: %s", steplibURL)
	if maintaner != "" {
		logger.Infof("Filtering Steps by maintaner: %s", maintaner)
	}

	if err := stepman.ListCachedSteps(steplibURL, maintaner, logger); err != nil {
		return err
	}
	return nil
}

func preloadSteps(c *cli.Context, logger log.Logger) error {
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

	logger.Infof("Preloading...")
	logger.Infof("Steplib: %s", steplibURL)
	if maintaner != "" {
		logger.Infof("Filtering Steps by maintaner: %s", maintaner)
	}
	logger.Printf("Options: %#v\n", opts)

	if err := preload.CacheSteps(logger, bitriseStepLibURL, bitriseMaintainer, opts); err != nil {
		return err
	}

	logger.Print()
	logger.Donef("Preloading completed.")

	return nil
}
