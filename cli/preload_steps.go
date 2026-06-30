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

func newListCachedStepsCommand() *cobra.Command {
	listCachedStepsCommand := &cobra.Command{
		Use:   "list-cached",
		Short: "List all the cached steps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			logCommandParameters(cmd)

			return listCachedSteps(cmd)
		},
	}

	listCachedStepsCommand.Flags().String("steplib-url", bitriseStepLibURL, "URL of the steplib to list or preload steps from")
	listCachedStepsCommand.Flags().String("maintainer", bitriseMaintainer, "Maintainer of the steps to list or preload")

	return listCachedStepsCommand
}

func newPreloadStepsCommand() *cobra.Command {
	preloadStepsCommand := &cobra.Command{
		Use:   "preload",
		Short: "Makes sure that Bitrise CLI can be used in offline mode by preloading Bitrise maintaned Steps.",
		Long:  fmt.Sprintf("Use the %s env var to test after preloading steps.", configs.IsSteplibOfflineModeEnvKey),
		RunE: func(cmd *cobra.Command, _ []string) error {
			logCommandParameters(cmd)

			if err := preloadSteps(cmd); err != nil {
				log.Errorf("Preload failed: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}

	pf := preloadStepsCommand.Flags()
	pf.String("steplib-url", bitriseStepLibURL, "URL of the steplib to list or preload steps from")
	pf.String("maintainer", bitriseMaintainer, "Maintainer of the steps to list or preload")
	pf.Uint("majors", 2, "Include X latest major versions")
	pf.Uint("minors", 1, "Include X latest minor versions for each major version")
	pf.Uint("minors-since", 2, "Include latest patch version of minors that were released in the last X months")
	pf.Uint("patches-since", 1, "Include all patch version that were released in the last X months")

	return preloadStepsCommand
}

func listCachedSteps(cmd *cobra.Command) error {
	steplibURL, _ := cmd.Flags().GetString("steplib-url")
	maintaner, _ := cmd.Flags().GetString("maintainer")

	log.Infof("Listing cached steps...")
	log.Infof("Steplib: %s", steplibURL)
	if maintaner != "" {
		log.Infof("Filtering Steps by maintaner: %s", maintaner)
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	return stepman.ListCachedSteps(steplibURL, maintaner, logger)
}

func preloadSteps(cmd *cobra.Command) error {
	steplibURL, _ := cmd.Flags().GetString("steplib-url")
	if steplibURL == "" {
		steplibURL = bitriseStepLibURL
	}
	maintaner, _ := cmd.Flags().GetString("maintainer")
	if maintaner == "" {
		maintaner = bitriseMaintainer
	}

	opts := preload.CacheOpts{}
	numMajor, _ := cmd.Flags().GetUint("majors")
	if numMajor != 0 {
		opts.NumMajor = numMajor
	}
	numMinor, _ := cmd.Flags().GetUint("minors")
	if numMinor != 0 {
		opts.NumMinor = numMinor
	}
	minorsSince, _ := cmd.Flags().GetUint("minors-since")
	if minorsSince != 0 {
		opts.LatestMinorsSinceMonths = int(minorsSince)
	}
	patchesSince, _ := cmd.Flags().GetUint("patches-since")
	if patchesSince != 0 {
		opts.PatchesSinceMonths = int(patchesSince)
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	log.Info("Preloading...")
	log.Infof("Steplib: %s", steplibURL)
	if maintaner != "" {
		log.Infof("Filtering Steps by maintaner: %s", maintaner)
	}
	log.Printf("Options: %#v\n", opts)

	if err := preload.CacheSteps(logger, steplibURL, maintaner, opts); err != nil {
		return err
	}

	log.Print()
	log.Donef("Preloading completed.")

	return nil
}
