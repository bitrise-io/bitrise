package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/cli/local"
	"github.com/bitrise-io/bitrise/v2/cli/plugin"
	"github.com/bitrise-io/bitrise/v2/cli/step"
	"github.com/bitrise-io/bitrise/v2/cli/yml"
	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

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
	rootCmd.PersistentFlags().Bool(cmdutil.DebugModeKey, false, "If true it enables DEBUG mode.")
	rootCmd.PersistentFlags().Bool(cmdutil.CIKey, false, "If true it indicates that we're used by another tool so don't require any user input!")
	rootCmd.PersistentFlags().Bool(cmdutil.PRKey, false, "If true bitrise runs in pull request mode.")
	cmdutil.SetFlagEnvVar(rootCmd.PersistentFlags(), cmdutil.DebugModeKey, configs.DebugModeEnvKey)
	cmdutil.SetFlagEnvVar(rootCmd.PersistentFlags(), cmdutil.CIKey, configs.CIModeEnvKey)

	rootCmd.AddCommand(
		local.NewCmd(),
		yml.NewCmd(),
		step.NewCmd(),

		versionCommand,
		updateCommand,
		plugin.NewCmd(),
		envmanCommand,

		// Deprecated, kept for backward compatibility but hidden (see local/trigger.go).
		cmdutil.AsHidden(local.NewTriggerCommand()),
	)

	// Backward-compatible hidden aliases: the old top-level command names keep
	// working while the canonical versions live under the local/yml/step groups.
	rootCmd.AddCommand(
		cmdutil.AsHidden(local.NewRunCommand()),
		cmdutil.AsHidden(local.NewInitCommand()),
		cmdutil.AsHidden(local.NewSetupCommand()),
		cmdutil.AsHidden(local.NewToolsCommand()),
		cmdutil.AsHidden(local.NewWorkflowListCommand()),
		cmdutil.AsHidden(yml.NewValidateCommand()),
		cmdutil.AsHidden(yml.NewMergeCommand()),
		cmdutil.AsHidden(step.NewShareCommand()),
		step.NewLegacyStepsCommand(),
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
