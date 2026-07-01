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

	// --debug, --ci and --pr are bound to their env vars: analytics report them as
	// set when sourced from the env, and the flag/env value is resolved by the mode
	// resolvers — for --debug additionally in Run() before cobra parses, so the
	// logger can be configured up front.
	rootCmd.PersistentFlags().Bool(DebugModeKey, false, "If true it enables DEBUG mode.")
	rootCmd.PersistentFlags().Bool(CIKey, false, "If true it indicates that we're used by another tool so don't require any user input!")
	rootCmd.PersistentFlags().Bool(PRKey, false, "If true bitrise runs in pull request mode.")
	setFlagEnvVar(rootCmd.PersistentFlags(), DebugModeKey, configs.DebugModeEnvKey)
	setFlagEnvVar(rootCmd.PersistentFlags(), CIKey, configs.CIModeEnvKey)

	rootCmd.AddCommand(
		newLocalCommand(),
		newYMLCommand(),
		newStepCommand(),

		versionCommand,
		updateCommand,
		pluginCommand,
		envmanCommand,

		// Deprecated, kept for backward compatibility but hidden (see trigger.go).
		triggerCommand,
	)

	// Backward-compatible hidden aliases: the old top-level command names keep
	// working while the canonical versions live under the local/yml/step groups.
	rootCmd.AddCommand(
		asHidden(newRunCommand()),
		asHidden(newInitCommand()),
		asHidden(newSetupCommand()),
		asHidden(newToolsCommand()),
		asHidden(newWorkflowListCommand()),
		asHidden(newValidateCommand()),
		asHidden(newMergeCommand()),
		asHidden(newShareCommand()),
		newLegacyStepsCommand(),
	)

	// Register the help command eagerly so it shows up in the command list
	// regardless of how help is reached.
	rootCmd.InitDefaultHelpCmd()

	// Render cobra's native help for every command, and append the installed
	// plugin list to the root help (cobra has no notion of bitrise plugins).
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		if cmd == rootCmd {
			printInstalledPlugins(cmd.OutOrStdout())
		}
	})

	return rootCmd
}

func asHidden(cmd *cobra.Command) *cobra.Command {
	cmd.Hidden = true
	return cmd
}
