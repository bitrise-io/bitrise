package cmdutil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RequireArgs returns an Args validator that names exactly which positional
// argument(s) are missing, instead of cobra's generic "accepts N arg(s), received M".
func RequireArgs(names ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) >= len(names) {
			return nil
		}
		missing := names[len(args):]
		var msg string
		if len(missing) == 1 {
			msg = fmt.Sprintf("missing argument: %s", missing[0])
		} else {
			msg = fmt.Sprintf("missing arguments: %s", strings.Join(missing, " "))
		}
		return fmt.Errorf("%s\nRun '%s --help' for usage", msg, cmd.CommandPath())
	}
}

// DelegateToList forwards a bare parent invocation to its "list" subcommand,
// propagating the parent's context so resolved config is available.
//
// Invoking RunE bypasses cobra's execute(), so the two things it would do for
// the subcommand are done here instead: InheritedFlags is called for its side
// effect of merging the parent chain's persistent flags (e.g. --workspace) into
// the subcommand's flagset, and required flags are validated explicitly.
func DelegateToList(cmd *cobra.Command, args []string) error {
	for _, sub := range cmd.Commands() {
		if sub.Name() != "list" {
			continue
		}
		sub.SetContext(cmd.Context())
		_ = sub.InheritedFlags()
		if err := sub.ValidateRequiredFlags(); err != nil {
			return err
		}
		return sub.RunE(sub, args)
	}
	return cmd.Help()
}

// SilenceRootErrors prevents cobra from printing a returned error to stderr by
// setting SilenceErrors on both the command and its root. Use this when the
// command has already printed its own error summary and the automatic
// "Error: ..." line would be redundant.
func SilenceRootErrors(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	if root := cmd.Root(); root != nil {
		root.SilenceErrors = true
	}
}
