package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	stepman "github.com/bitrise-io/stepman/cli"
	"github.com/bitrise-io/stepman/preload"
	"github.com/spf13/cobra"
)

const (
	bitriseStepLibURL = "https://github.com/bitrise-io/bitrise-steplib.git"
	bitriseMaintainer = "bitrise"
)

var stepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Manage Steps cache.",
}

var listCachedStepsCmd = &cobra.Command{
	Use:   "list-cached",
	Short: "List all the cached steps",
	RunE:  runListCachedSteps,
}

var listCachedStepsOpts struct {
	steplibURL string
	maintainer string
}

var preloadStepsCmd = &cobra.Command{
	Use:   "preload",
	Short: fmt.Sprintf("Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps. Use the %s env var to test after preloading steps.", configs.IsSteplibOfflineModeEnvKey),
	RunE:  runPreloadSteps,
}

var preloadStepsOpts struct {
	steplibURL  string
	maintainer  string
	majors      uint
	minors      uint
	minorsSince uint
	patchesSince uint
}

func init() {
	stepsCmd.AddCommand(listCachedStepsCmd, preloadStepsCmd)

	listCachedStepsCmd.Flags().StringVar(&listCachedStepsOpts.steplibURL, "steplib-url", bitriseStepLibURL, "URL of the steplib to list or preload steps from")
	listCachedStepsCmd.Flags().StringVar(&listCachedStepsOpts.maintainer, "maintainer", bitriseMaintainer, "Maintainer of the steps to list or preload")

	preloadStepsCmd.Flags().StringVar(&preloadStepsOpts.steplibURL, "steplib-url", bitriseStepLibURL, "URL of the steplib to list or preload steps from")
	preloadStepsCmd.Flags().StringVar(&preloadStepsOpts.maintainer, "maintainer", bitriseMaintainer, "Maintainer of the steps to list or preload")
	preloadStepsCmd.Flags().UintVar(&preloadStepsOpts.majors, "majors", 2, "Include X latest major versions")
	preloadStepsCmd.Flags().UintVar(&preloadStepsOpts.minors, "minors", 1, "Include X latest minor versions for each major version")
	preloadStepsCmd.Flags().UintVar(&preloadStepsOpts.minorsSince, "minors-since", 2, "Include latest patch version of minors that were released in the last X months")
	preloadStepsCmd.Flags().UintVar(&preloadStepsOpts.patchesSince, "patches-since", 1, "Include all patch version that were released in the last X months")
}

func runListCachedSteps(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	return listCachedSteps(cmd, args)
}

func runPreloadSteps(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	if err := preloadSteps(cmd, args); err != nil {
		log.Errorf("Preload failed: %s", err)
		os.Exit(1)
	}
	return nil
}

func listCachedSteps(_ *cobra.Command, _ []string) error {
	steplibURL := listCachedStepsOpts.steplibURL
	maintaner := listCachedStepsOpts.maintainer

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

func preloadSteps(_ *cobra.Command, _ []string) error {
	steplibURL := preloadStepsOpts.steplibURL
	maintaner := preloadStepsOpts.maintainer

	opts := preload.CacheOpts{}
	numMajor := preloadStepsOpts.majors
	if numMajor != 0 {
		opts.NumMajor = numMajor
	}
	numMinor := preloadStepsOpts.minors
	if numMinor != 0 {
		opts.NumMinor = numMinor
	}
	minorsSince := preloadStepsOpts.minorsSince
	if minorsSince != 0 {
		opts.LatestMinorsSinceMonths = int(minorsSince)
	}
	patchesSince := preloadStepsOpts.patchesSince
	if patchesSince != 0 {
		opts.PatchesSinceMonths = int(patchesSince)
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
