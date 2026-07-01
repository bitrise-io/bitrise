package local

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

// NewSetupCommand ...
func NewSetupCommand() *cobra.Command {
	setupCommand := &cobra.Command{
		Use:   "setup",
		Short: "Setup the current host. Install every required tool to run Workflows.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := setup(cmd); err != nil {
				log.Errorf("Setup failed, error: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}

	setupCommand.Flags().Bool("clean", false, "Removes bitrise's workdir before setup.")
	setupCommand.Flags().Bool("minimal", false, "Only installs the required tools for running in CI mode.")
	setupCommand.Flags().Bool("no-update", false, "Skip updating core tools (stepman/envman) and plugins if they are already installed, even if outdated.")

	return setupCommand
}

func setup(cmd *cobra.Command) error {
	clean, _ := cmd.Flags().GetBool("clean")
	minimal, _ := cmd.Flags().GetBool("minimal")
	noUpdate, _ := cmd.Flags().GetBool("no-update")
	noUpdate = noUpdate || os.Getenv(configs.SetupNoUpdateEnvKey) == "true"

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
