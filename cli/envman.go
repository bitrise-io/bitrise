package cli

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/spf13/cobra"
)

// envmanCommand is dispatched before cobra (see envmanPassthrough), so cobra
// only needs it registered to show up in help and completion. DisableFlagParsing
// keeps cobra from rejecting envman's own flags if it is ever reached directly.
var envmanCommand = &cobra.Command{
	Use:                "envman",
	Short:              "Runs an envman command.",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logCommandParameters(cmd)

		if err := runCommandWith("envman", args); err != nil {
			failf("Command failed, error: %s", err)
		}
		return nil
	},
}

func runCommandWith(toolName string, args []string) error {
	cmd := command.NewWithStandardOuts(toolName, args...).SetStdin(os.Stdin)
	return cmd.Run()
}
