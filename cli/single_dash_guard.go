package cli

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

// rejectSingleDashLongFlags errors out before cobra parses args if any
// argument spells a registered long flag name with a single dash instead of
// "--" (e.g. "-config" instead of "--config"). This has to run before
// Execute(): by the time any hook on the resolved command sees the args,
// pflag has already parsed them, and the misparse it produces (or the
// cryptic "unknown shorthand flag" error, when the leading character isn't a
// registered shorthand) is indistinguishable after the fact from a
// legitimate shorthand invocation.
func rejectSingleDashLongFlags(root *cobra.Command, rawArgs []string) {
	target, _, err := root.Find(rawArgs)
	if err != nil {
		return // let cobra's own error handling take over
	}

	if arg, flagName, found := cmdutil.DetectSingleDashLongFlag(target, rawArgs); found {
		cmdutil.Failf("unknown flag: %s (did you mean --%s?)", arg, flagName)
	}
}
