package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	// List commands in declaration order (like the previous framework) rather
	// than cobra's default alphabetical sort.
	cobra.EnableCommandSorting = false

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

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().Bool(DebugModeKey, false, "If true it enables DEBUG mode.")
	rootCmd.PersistentFlags().Bool(CIKey, false, "If true it indicates that we're used by another tool so don't require any user input!")
	rootCmd.PersistentFlags().Bool(PRKey, false, "If true bitrise runs in pull request mode.")
	markEnvVar(rootCmd.PersistentFlags(), DebugModeKey, configs.DebugModeEnvKey)
	markEnvVar(rootCmd.PersistentFlags(), CIKey, configs.CIModeEnvKey)

	rootCmd.AddCommand(
		initCmd,
		setupCommand,
		stepsCommand,
		toolsCommand,
		versionCmd,
		validateCmd,
		updateCommand,
		runCommand,
		triggerCheckCmd,
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

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			printRootHelp(rootCmd)
			return
		}
		defaultHelp(cmd, args)
	})

	return rootCmd
}
