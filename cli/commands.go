package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/cli/legacy"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

// requireKnownSubcommand is the RunE for parent commands that only dispatch to
// subcommands (they have no action of their own): show help when invoked bare,
// and error on an unknown subcommand instead of silently succeeding.
func requireKnownSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               filepath.Base(os.Args[0]),
		Short:             "Bitrise Automations Workflow Runner",
		Version:           version.VERSION,
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: before,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			return errors.New("")
		},
	}

	// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
	// Reproduce the previous framework's command surface: declaration-order (not
	// alphabetical) command and flag listing, a bare version string (not cobra's
	// "bitrise version X"), and no auto-generated `completion` command. The next
	// major version can drop these and adopt cobra's defaults.
	cobra.EnableCommandSorting = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// END MIGRATION PERIOD COMPATIBILITY

	// --debug is declared here for help, analytics and plugin/envman flag
	// recognition, but its effective value is resolved by legacy.IsDebugMode
	// before cobra parses, so the logger can be configured up front.
	rootCmd.PersistentFlags().Bool(DebugModeKey, false, "If true it enables DEBUG mode.")
	rootCmd.PersistentFlags().Bool(CIKey, false, "If true it indicates that we're used by another tool so don't require any user input!")
	rootCmd.PersistentFlags().Bool(PRKey, false, "If true bitrise runs in pull request mode.")
	markEnvVar(rootCmd.PersistentFlags(), DebugModeKey, configs.DebugModeEnvKey)
	markEnvVar(rootCmd.PersistentFlags(), CIKey, configs.CIModeEnvKey)

	rootCmd.AddCommand(
		initCommand,
		setupCommand,
		stepsCommand,
		toolsCommand,
		versionCommand,
		validateCommand,
		updateCommand,
		runCommand,
		triggerCheckCommand,
		triggerCommand,
		workflowListCommand,
		shareCommand,
		pluginCommand,
		envmanCommand,
		mergeConfigCommand,
	)

	// Register the help command eagerly so it shows up in the command list
	// regardless of how help is reached.
	rootCmd.InitDefaultHelpCmd()

	// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
	// Route the root command's --help to the hand-rendered printRootHelp; other
	// commands fall through to cobra's native help. The next major can drop this.
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			printRootHelp(rootCmd)
			return
		}
		defaultHelp(cmd, args)
	})

	// TODO: MIGRATION PERIOD - NEEDED TO KEEP COMPATIBILITY
	// Reproduce urfave/cli's leniency towards an unrecognised flag that follows a
	// positional argument (see legacy.EnableUnknownFlagPassthrough) so the
	// migration stays behaviour-preserving.
	legacy.EnableUnknownFlagPassthrough(rootCmd)

	return rootCmd
}
