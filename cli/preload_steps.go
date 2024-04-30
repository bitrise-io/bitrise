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
	Flags: []cli.Flag{},
}

func preloadSteps(c *cli.Context) error {
	log.Info("Preloading Bitrise maintained Steps...")
	log.Printf("")

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	if err := stepman.PreloadBitriseSteps(toolkits.GoBuildStep, logger); err != nil {
		return err
	}

	log.Print()
	log.Donef("Preloading completed.")

	return nil
}
