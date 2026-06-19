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
	// --debug is declared here for help, analytics and plugin/envman flag
	// recognition, but its effective value is resolved by isDebugMode before
	// cobra parses, so the logger can be configured up front.
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

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			printRootHelp(rootCmd)
			return
		}
		defaultHelp(cmd, args)
	})

	// urfave/cli left an unrecognised flag that followed a positional argument in
	// the argument list and ignored it (e.g. `bitrise run wf --unknown` still ran
	// the workflow); pflag rejects unknown flags outright. Reproduce the old
	// leniency so the migration stays behaviour-preserving — the next major
	// version, which reworks the command surface, can tighten this.
	enableUnknownFlagPassthrough(rootCmd)

	return rootCmd
}

// enableUnknownFlagPassthrough sets FParseErrWhitelist.UnknownFlags on the whole
// command tree. The flag is per-command (cobra does not inherit it), and the
// command that ultimately runs is the one that parses, so it must be set on every
// command, not just the root.
func enableUnknownFlagPassthrough(cmd *cobra.Command) {
	cmd.FParseErrWhitelist.UnknownFlags = true
	for _, sub := range cmd.Commands() {
		enableUnknownFlagPassthrough(sub)
	}
}
