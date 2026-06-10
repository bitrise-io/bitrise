package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the current host. Install every required tool to run Workflows.",
	RunE:  runSetup,
}

var setupOpts struct {
	clean    bool
	minimal  bool
	noUpdate bool
}

func init() {
	setupCmd.Flags().BoolVar(&setupOpts.clean, "clean", false, "Removes bitrise's workdir before setup.")
	setupCmd.Flags().BoolVar(&setupOpts.minimal, "minimal", false, "Only installs the required tools for running in CI mode.")
	setupCmd.Flags().BoolVar(&setupOpts.noUpdate, "no-update", false, "Skip updating core tools (stepman/envman) and plugins if they are already installed, even if outdated.")
}

func runSetup(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	if err := setup(cmd, args); err != nil {
		log.Errorf("Setup failed, error: %s", err)
		os.Exit(1)
	}
	return nil
}

func setup(_ *cobra.Command, _ []string) error {
	clean := setupOpts.clean
	minimal := setupOpts.minimal
	noUpdate := setupOpts.noUpdate || os.Getenv(configs.SetupNoUpdateEnvKey) == "true"

	setupMode := bitrise.SetupModeDefault
	if minimal {
		setupMode = bitrise.SetupModeMinimal
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	if err := bitrise.RunSetup(logger, version.VERSION, setupMode, clean, noUpdate); err != nil {
		return err
	}

	logger.Print()
	logger.Infof("To start using bitrise:")
	logger.Printf("* cd into your project's directory (if you're not there already)")
	logger.Printf("* call: bitrise init")
	logger.Printf("* follow the guide")
	logger.Print()
	logger.Donef("That's all :)")

	return nil
}
